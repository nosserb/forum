package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"sync"
)

var (
	db        *sql.DB
	templates *template.Template
)

// Pour que les handlers ait accès a la db
func SetDB(database *sql.DB) {
	db = database
}

// La route modifier pour chaque post
func PostRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ViewPostHandler(w, r)
		return
	}
	PostHandler(w, r)
}

func RegisterRoutes(
	mux *http.ServeMux, tmpl *template.Template, dbConn *sql.DB, sseClients map[int][]chan Notification, sseMu *sync.RWMutex) {
	templates = tmpl

	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/post", PostRouteHandler)
	mux.HandleFunc("/reply", func(w http.ResponseWriter, r *http.Request) {
		ReplyHandler(w, r, sseClients, sseMu)
	})
	mux.HandleFunc("/logout", LogoutHandler)
	mux.HandleFunc("/like", func(w http.ResponseWriter, r *http.Request) {
		LikeHandler(w, r, sseClients, sseMu)
	})
	mux.HandleFunc("/dislike", func(w http.ResponseWriter, r *http.Request) {
		DislikeHandler(w, r, sseClients, sseMu)
	})
	mux.HandleFunc("/filter", FilterHandler)
	mux.HandleFunc("/images", ImageHandler)

	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		SSEHandler(w, r, dbConn, sseClients, sseMu)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, dbConn, templates)
	})

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		SignupHandler(w, r, dbConn, templates)
	})
}
