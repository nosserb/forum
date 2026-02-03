package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	"forum/controller/cookies"
	"forum/controller/logging"
	forumDB "forum/model/functions"
)

// Gère l'inscription
func SignupHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, templates *template.Template) {
	if r.Method != http.MethodPost {
		if templates != nil {
			_ = templates.ExecuteTemplate(w, "signup.html", nil)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusMethodNotAllowed)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	log.Printf("[SIGNUP] Attempt: Email=%s, Username=%s", email, username)

	// should not happen but anyway
	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Email and password required")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	// email or username is already used
	userID, err := forumDB.InsertUser(db, email, username, password)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		logging.Logger.Printf("[SIGNUP] InsertUser error: %v", err)
		ErrorHandler(w, r, http.StatusConflict, "Cannot create account (email or username already used)")
		return
	}

	log.Printf("[SIGNUP] Account created: Username=%s", username)

	// Init session ID cookie
	err = cookies.WriteSessionCookie(w, r, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("Error writing session cookie : %v", err)
		return
	}

	// Redirection vers le forum
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Gère la connexion
func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, templates *template.Template) {
	if r.Method != http.MethodPost {
		if templates != nil {
			_ = templates.ExecuteTemplate(w, "login.html", nil)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusMethodNotAllowed)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	log.Printf("[LOGIN] Attempt: Email=%s", email)

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Email and password required")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	user, err := forumDB.FindUser(db, email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logging.Logger.Printf("[LOGIN] FindUser error: %v", err)
		ErrorHandler(w, r, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if user.Password != password {
		w.WriteHeader(http.StatusUnauthorized)
		logging.Logger.Printf("[LOGIN] Incorrect password for Email=%s", email)
		ErrorHandler(w, r, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	username := user.Username
	if username == "" {
		if at := strings.Index(email, "@"); at > 0 {
			username = email[:at]
		} else {
			username = email
		}
	}

	session, _ := forumDB.FetchSessionByUser(db, int64(user.ID))
	if session != (forumDB.Session{}) {
		w.WriteHeader(http.StatusUnauthorized)
		logging.Logger.Printf("[LOGIN] User %s is already logged in", user.Username)
		ErrorHandler(w, r, http.StatusUnauthorized, "User is already logged in")
		return
	}

	log.Printf("[LOGIN] Successful login: Username=%s", username)

	// Init session ID cookie
	err = cookies.WriteSessionCookie(w, r, int64(user.ID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("Error writing session cookie : %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
