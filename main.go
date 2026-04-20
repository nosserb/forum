package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/controller/handlers"
	"forum/controller/logging"
	"forum/controller/server"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"forum/controller/cookies"
	"forum/model/data"
	forumDB "forum/model/functions"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Ouverture de la connexion à la base SQLite
	db, err := sql.Open("sqlite3", "./model/forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	forumDB.Initialisation(db)

	// Delete all previous sessions
	_, err = forumDB.DeleteAllSessions(db)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	// Parse templates
	templates := server.ParseTemplates("./view/assets/templates/*.html")

	sseClients := make(map[int][]chan handlers.Notification, 1000)
	sseMu := &sync.RWMutex{}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./view/assets/static"))
	mux.Handle("/statics/", http.StripPrefix("/statics/", fs))

	// tmp := fetch all posts
	allPosts, err := forumDB.FetchPosts(db)
	if err != nil {
		log.Fatal(err)
	}
	data.CombinedData = data.AllData{
		ToDisplay: data.ToDisplay{
			Posts: allPosts,
		},
		Username: "",
	}

	handlers.RegisterRoutes(mux, templates, db, sseClients, sseMu)

	// dev/testing route
	mux.HandleFunc("/dev", func(w http.ResponseWriter, r *http.Request) {
		devHandler(w, r, db, sseClients, sseMu)
	})

	mux.HandleFunc("/dev/sendnotif", func(w http.ResponseWriter, r *http.Request) {
		DevSendNotifHandler(w, r, db, sseClients, sseMu)
	})

	// cookie db init - temporary
	cookies.SetDB(db)

	// handlers db init - temporary
	handlers.SetDB(db)

	// Init custom logger
	logging.Init()

	// websocket //
	manager := handlers.NewManager()
	handler := &handlers.Handler{
		DB:      db,
		Manager: manager,
	}
	mux.Handle("/ws", handler)
	// --- //

	logging.Logger.Println("Server starting : http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logging.Logger.Fatal(err)
	}
}

func devHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sseClients map[int][]chan handlers.Notification, sseMu *sync.RWMutex) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `
<html>
<head>
<title>Dev Debug</title>
</head>
<body>

<h2>Envoyer une notification SSE</h2>

<label>Receiver username: <input type="text" id="receiverUsername" value="testuser"></label><br>
<label>Sender username: <input type="text" id="senderUsername" value="devuser"></label><br>

<label>Type:
<select id="type">
<option value="debug">debug</option>
<option value="like">like</option>
<option value="dislike">dislike</option>
<option value="comment">comment</option>
<option value="message">message</option>
</select>
</label><br>

<label>Subject Type:
<select id="subjectType">
<option value="user">user</option>
<option value="post">post</option>
<option value="comment">comment</option>
</select>
</label><br>

<label>Subject ID: <input type="number" id="subjectID" value="1"></label><br>
<label>Subject Label: <input type="text" id="subjectLabel" value="Test message"></label><br>

<button onclick="sendNotif()">Envoyer</button>
<pre id="resp"></pre>

<script>
function sendNotif() {
    const payload = {
        receiverUsername: document.getElementById("receiverUsername").value,
        senderUsername: document.getElementById("senderUsername").value,
        type: document.getElementById("type").value,
        subjectType: document.getElementById("subjectType").value,
        subjectID: parseInt(document.getElementById("subjectID").value),
        subjectLabel: document.getElementById("subjectLabel").value
    };

    fetch("/dev/sendnotif", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(payload)
    })
    .then(r => r.text())
    .then(data => {
        document.getElementById("resp").textContent = data;
    });
}
</script>

<hr>
</body>
</html>
`

	_, _ = w.Write([]byte(html))

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;")
	if err != nil {
		w.Write([]byte("Erreur: " + err.Error()))
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			w.Write([]byte("Erreur: " + err.Error()))
			return
		}
		tables = append(tables, name)
	}

	for _, table := range tables {
		w.Write([]byte("<h3>Table: " + table + "</h3>"))

		query := "SELECT * FROM " + table + ";"
		tblRows, err := db.Query(query)
		if err != nil {
			w.Write([]byte("(erreur lecture table: " + err.Error() + ")<br>"))
			continue
		}

		cols, err := tblRows.Columns()
		if err != nil {
			w.Write([]byte("(erreur colonnes: " + err.Error() + ")<br>"))
			tblRows.Close()
			continue
		}
		w.Write([]byte("<b>Colonnes:</b> " + strings.Join(cols, ", ") + "<br>"))

		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}

		for tblRows.Next() {
			if err := tblRows.Scan(ptrs...); err != nil {
				w.Write([]byte("(erreur scan: " + err.Error() + ")<br>"))
				continue
			}

			var out []string
			for _, v := range vals {
				if v == nil {
					out = append(out, "NULL")
				} else {
					out = append(out, fmt.Sprintf("%v", v))
				}
			}
			w.Write([]byte(strings.Join(out, " | ") + "<br>"))
		}
		tblRows.Close()
	}

	w.Write([]byte("</body></html>"))
	fmt.Printf("DevHandler accessed: %v %v %v\n", r.RemoteAddr, r.Method, r.URL.Path)
}

func DevSendNotifHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sseClients map[int][]chan handlers.Notification, sseMu *sync.RWMutex) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ReceiverUsername string `json:"receiverUsername"`
		SenderUsername   string `json:"senderUsername"`
		Type             string `json:"type"`
		SubjectType      string `json:"subjectType"`
		SubjectID        int    `json:"subjectID"`
		SubjectLabel     string `json:"subjectLabel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	users, err := forumDB.FetchUsersBy(db, "username", payload.ReceiverUsername)
	if err != nil || len(users) == 0 {
		http.Error(w, "Receiver not found", http.StatusNotFound)
		return
	}

	receiver := users[0]

	notif := handlers.Notification{
		ReceiverID:   receiver.ID,
		SenderName:   payload.SenderUsername,
		Type:         payload.Type,
		SubjectType:  payload.SubjectType,
		SubjectID:    payload.SubjectID,
		SubjectLabel: payload.SubjectLabel,
		CreatedAt:    time.Now(),
	}

	handlers.SendNotification(notif, sseClients, sseMu)
	w.Write([]byte("ok"))
}
