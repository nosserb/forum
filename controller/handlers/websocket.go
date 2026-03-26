package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	forumDB "forum/model/functions"

	"github.com/gorilla/websocket"
)

// Structs added to replace the main func //

type Manager struct {
	instanceIn   map[int64]chan string // instanceID → in
	instanceOut  map[int64]chan string // instanceID → out
	userIn       map[int64]chan string // userID → in
	instanceUser map[int64]int64

	outEvents    chan string
	newInstances chan Instance

	instanceCounter int64

	userInMu sync.RWMutex
}

type Instance struct {
	ID  int64
	In  chan string
	Out chan string
}

func NewManager() *Manager {
	m := &Manager{
		instanceIn:   make(map[int64]chan string),
		instanceOut:  make(map[int64]chan string),
		userIn:       make(map[int64]chan string),
		instanceUser: make(map[int64]int64),
		outEvents:    make(chan string, 100),
		newInstances: make(chan Instance, 10),
	}

	go m.run()

	return m
}

func (m *Manager) run() {
	for {
		select {

		// --- Événements venant des handlers ---
		case e := <-m.outEvents:
			parts := strings.Split(e, "|")

			switch parts[0] {

			// Format attendu :
			// "__register__|instanceID|userID"
			case "__register__":
				if len(parts) < 3 {
					continue
				}

				instanceID, err1 := strconv.ParseInt(parts[1], 10, 64)
				userID, err2 := strconv.ParseInt(parts[2], 10, 64)
				if err1 != nil || err2 != nil {
					continue
				}

				// Associer instance → user
				m.instanceUser[instanceID] = userID

				// Associer user → in avec mutex
				if in, ok := m.instanceIn[instanceID]; ok {
					m.userInMu.Lock()
					m.userIn[userID] = in
					m.userInMu.Unlock()
				}

			// Format attendu :
			// "__unregister__|instanceID"
			case "__unregister__":
				if len(parts) < 2 {
					continue
				}

				instanceID, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					continue
				}

				// Retrouver le user associé
				if userID, ok := m.instanceUser[instanceID]; ok {
					m.userInMu.Lock()
					delete(m.userIn, userID)
					m.userInMu.Unlock()
					delete(m.instanceUser, instanceID)
				}

				delete(m.instanceIn, instanceID)
				delete(m.instanceOut, instanceID)

			default:
				// log éventuel
			}

		// --- Nouveau handler ---
		case newInst := <-m.newInstances:
			m.instanceIn[newInst.ID] = newInst.In
			m.instanceOut[newInst.ID] = newInst.Out
		}
	}
}

// --- //

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request, params []any) {
	// --- Extraction des paramètres ---
	in := params[0].(chan string)
	out := params[1].(chan string)
	db := params[2].(*sql.DB)
	instances := params[3].(*map[int64]chan string)
	mu := params[4].(*sync.RWMutex)
	instanceID := params[5].(int64)

	// --- Upgrade HTTP → WebSocket ---
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		out <- "Upgrade error: " + err.Error()
		return
	}

	defer func() {
		conn.Close()
		out <- "__unregister__|" + strconv.FormatInt(instanceID, 10)
	}()

	// --- Demander le sessionID au client ---
	if err := conn.WriteMessage(
		websocket.TextMessage,
		[]byte("0|0|"+time.Now().Format("2006-01-02 15:04:05")+"|requestsessionid"),
	); err != nil {
		return
	}

	// --- Attendre la réponse sessionID ---
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return
	}

	parts := strings.SplitN(string(msg), "|", 4)
	if len(parts) < 4 || !strings.HasPrefix(parts[3], "sessionid:") {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("0|0|"+time.Now().Format("2006-01-02 15:04:05")+"|invalid session format"),
		)
		return
	}

	sessionID := strings.TrimPrefix(parts[3], "sessionid:")

	// --- Récupération session en base ---
	session, err := forumDB.FetchSession(db, sessionID)
	if err != nil {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("0|0|"+time.Now().Format("2006-01-02 15:04:05")+"|invalid session"),
		)
		return
	}

	userID := session.UserID

	// --- Envoi de la commande d'enregistrement avec conversion int64 → string ---
	// could break because of int64 conversion
	out <- "__register__|" + strconv.FormatInt(instanceID, 10) + "|" + strconv.FormatInt(int64(userID), 10)

	// --- Récupérer tous les messages de l'utilisateur et les envoyer au client ---
	// --- Récupérer les correspondants ---
	correspondents, err := forumDB.FetchPrivateMessageCorrespondents(db, int64(userID))
	if err != nil {
		out <- "DB fetch correspondents error: " + err.Error()
	} else {

		for _, otherID := range correspondents {

			// récupérer les 10 derniers messages de la conversation
			msgs, err := forumDB.FetchPrivateMessagesBetween(db, int64(userID), otherID, 10, 0)
			if err != nil {
				out <- "DB fetch messages error: " + err.Error()
				continue
			}

			// envoyer du plus ancien au plus récent
			for i := len(msgs) - 1; i >= 0; i-- {
				m := msgs[i]

				line := strconv.FormatInt(int64(m.SenderID), 10) + "|" +
					strconv.FormatInt(int64(m.ReceiverID), 10) + "|" +
					m.CreatedAt.Format("2006-01-02 15:04:05") + "|" +
					m.Content

				if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
					out <- "Write error: " + err.Error()
					return
				}
			}
		}
	}

	// --- Message initial du serveur ---
	conn.WriteMessage(
		websocket.TextMessage,
		[]byte("0|0|"+time.Now().Format("2006-01-02 15:04:05")+"|message initial du serveur"),
	)

	// --- Channel pour gérer la lecture WebSocket ---
	done := make(chan struct{})

	// --- Goroutine lecture WebSocket ---
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					out <- "Client disconnected"
				} else {
					out <- "Read error: " + err.Error()
				}
				return
			}

			// --- Décortiquer le message reçu ---
			parts := strings.SplitN(string(msg), "|", 4)
			if len(parts) < 4 {
				out <- "Invalid message format"
				continue
			}

			senderID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				out <- "Invalid sender ID"
				continue
			}

			receiverID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				out <- "Invalid receiver ID"
				continue
			}

			timestamp := parts[2]
			content := parts[3]

			// --- Traitement messages système ---
			if senderID == 0 && receiverID == 0 {

				switch {

				case content == "sendallactiveusers":

					mu.RLock()
					var activeIDs []string
					for id := range *instances {
						activeIDs = append(activeIDs, strconv.FormatInt(id, 10))
					}
					mu.RUnlock()

					in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
						"|activeusers:" + strings.Join(activeIDs, ",")

				case strings.HasPrefix(content, "linkidtouser|"):

					sub := strings.SplitN(content, "|", 2)
					if len(sub) < 2 {
						continue
					}

					targetID, err := strconv.ParseInt(sub[1], 10, 64)
					if err != nil {
						continue
					}

					users, err := forumDB.FetchUsersBy(db, "id", targetID)
					if err != nil || len(users) == 0 {
						in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
							"|usernotfound"
						continue
					}

					u := users[0]

					userInfo := strconv.Itoa(u.ID) + "," +
						u.Username + "," +
						u.FirstName + "," +
						u.LastName + "," +
						u.Email

					in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
						"|userinfo:" + userInfo

				case strings.HasPrefix(content, "fetchmessages|"):

					sub := strings.SplitN(content, "|", 3)
					if len(sub) < 3 {
						in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
							"|invalidfetchformat"
						continue
					}

					// --- Parse correspondant ID ---
					correspondentID, err := strconv.ParseInt(sub[1], 10, 64)
					if err != nil {
						in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
							"|invalidcorrespondentid"
						continue
					}

					// --- Parse offset ---
					offset, err := strconv.Atoi(sub[2])
					if err != nil || offset < 0 {
						in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
							"|invalidoffset"
						continue
					}

					// --- Récupérer 10 messages à partir de l'offset ---
					msgs, err := forumDB.FetchPrivateMessagesBetween(db, int64(userID), correspondentID, 10, offset)
					if err != nil {
						in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
							"|dbfetcherror:" + err.Error()
						continue
					}

					// --- Envoyer du plus ancien au plus récent ---
					for i := len(msgs) - 1; i >= 0; i-- {
						m := msgs[i]
						line := strconv.FormatInt(int64(m.SenderID), 10) + "|" +
							strconv.FormatInt(int64(m.ReceiverID), 10) + "|" +
							m.CreatedAt.Format("2006-01-02 15:04:05") + "|" +
							m.Content

						in <- line
					}

				default:
					in <- "0|0|" + time.Now().Format("2006-01-02 15:04:05") +
						"|unknownsystemcommand"
				}

				continue
			}

			// --- Sécurité sender ---
			if senderID != int64(userID) {
				out <- "Sender ID mismatch"
				continue
			}

			// --- Enregistrer le message en base ---
			_, err = forumDB.InsertPrivateMessage(db, int64(userID), receiverID, content)
			if err != nil {
				out <- "DB insert error: " + err.Error()
				continue
			}

			// --- Construire la ligne à envoyer ---
			// could break because of int64 conversion
			line := strconv.FormatInt(int64(userID), 10) + "|" +
				strconv.FormatInt(receiverID, 10) + "|" +
				timestamp + "|" +
				content

			// --- Envoyer au destinataire si connecté ---
			mu.RLock()
			recvChan, ok := (*instances)[receiverID]
			mu.RUnlock()
			if ok {
				select {
				case recvChan <- line: // non bloquant
				default:
				}
			}

		}
	}()

	// --- Goroutine écriture WebSocket ---
	go func() {
		for {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	<-done
}

// Reworwed Handler //

type Handler struct {
	DB      *sql.DB
	Manager *Manager
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	in := make(chan string, 10)
	out := make(chan string, 10)

	go func() {
		for msg := range out {
			h.Manager.outEvents <- msg
		}
	}()

	instanceID := atomic.AddInt64(&h.Manager.instanceCounter, 1)

	h.Manager.newInstances <- Instance{
		ID:  instanceID,
		In:  in,
		Out: out,
	}

	wsHandler(w, r, []any{
		in,
		out,
		h.DB,
		&h.Manager.userIn,
		&h.Manager.userInMu,
		instanceID,
	})
}

// --- //
