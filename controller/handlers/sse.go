package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	forumDB "forum/model/functions"
)

type Notification struct {
	ReceiverID   int
	SenderID     int
	SenderName   string
	Type         string
	SubjectType  string
	SubjectID    int
	SubjectLabel string
	CreatedAt    time.Time
}

func (n Notification) Format() string {
	return fmt.Sprintf("%s|%s|%s|%s|%d|%s",
		n.SenderName,
		n.CreatedAt.Format(time.RFC3339),
		n.Type,
		n.SubjectType,
		n.SubjectID,
		n.SubjectLabel,
	)
}

func SSEHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sseClients map[int][]chan Notification, sseMu *sync.RWMutex) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", 500)
		return
	}

	cookie, err := r.Cookie("sessionCookie")
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}

	user, err := forumDB.FetchUserBySession(db, cookie.Value)
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan Notification, 5)

	sseMu.Lock()
	sseClients[user.ID] = append(sseClients[user.ID], ch)
	sseMu.Unlock()

	defer func() {
		sseMu.Lock()
		list := sseClients[user.ID]
		for i, c := range list {
			if c == ch {
				sseClients[user.ID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(sseClients[user.ID]) == 0 {
			delete(sseClients, user.ID)
		}
		sseMu.Unlock()
	}()

	for {
		select {
		case notif := <-ch:
			msg := notif.Format()
			w.Write([]byte("data: " + msg + "\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func SendNotification(notif Notification, sseClients map[int][]chan Notification, sseMu *sync.RWMutex) {
	sseMu.RLock()
	conns, ok := sseClients[notif.ReceiverID]
	if !ok || len(conns) == 0 {
		sseMu.RUnlock()
		return
	}

	channels := make([]chan Notification, len(conns))
	copy(channels, conns)
	sseMu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- notif:
		default:
		}
	}
}
