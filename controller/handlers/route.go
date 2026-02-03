package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
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

func RegisterRoutes(mux *http.ServeMux, tmpl *template.Template, dbConn *sql.DB) {
	templates = tmpl

	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/post", PostRouteHandler)
	mux.HandleFunc("/reply", ReplyHandler)
	mux.HandleFunc("/logout", LogoutHandler)
	mux.HandleFunc("/like", LikeHandler)
	mux.HandleFunc("/dislike", DislikeHandler)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, dbConn, templates)
	})

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		SignupHandler(w, r, dbConn, templates)
	})
}
