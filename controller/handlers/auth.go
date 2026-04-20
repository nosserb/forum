package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"forum/controller/cookies"
	"forum/controller/logging"
	forumDB "forum/model/functions"
)

// Gère l'inscription
func SignupHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, templates *template.Template) {
	const (
		minAge = 13
		maxAge = 120
	)

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
	firstName := strings.TrimSpace(r.FormValue("firstName"))
	lastName := strings.TrimSpace(r.FormValue("lastName"))
	gender := strings.TrimSpace(r.FormValue("gender"))
	ageStr := strings.TrimSpace(r.FormValue("age"))
	password := r.FormValue("password")

	// should not happen but anyway
	if email == "" || password == "" || username == "" || gender == "" || ageStr == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Email, username, age, gender and password required")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(username) > 20 {
		ErrorHandler(w, r, http.StatusBadRequest, "Username must be 20 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(password) < 8 || len(password) > 25 {
		ErrorHandler(w, r, http.StatusBadRequest, "Password must be between 8 and 25 characters")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	age, err := strconv.Atoi(ageStr)
	if err != nil || age < minAge || age > maxAge {
		ErrorHandler(w, r, http.StatusBadRequest, "Age must be between 13 and 120")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	// email or username is already used
	userID, err := forumDB.InsertUser(db, email, username, firstName, lastName, gender, password, age)
	if err != nil {
		logging.Logger.Printf("[SIGNUP] InsertUser error: %v", err)
		ErrorHandler(w, r, http.StatusConflict, "Cannot create account (email or username already used)")
		return
	}

	// Init session ID cookie
	err = cookies.WriteSessionCookie(w, r, userID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("Error writing session cookie : %v", err)
		return
	}

	// Redirection vers le forum
	http.Redirect(w, r, "/", http.StatusSeeOther)
	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusSeeOther)
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

	identifier := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if identifier == "" || password == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Email or nickname and password required")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	user, err := forumDB.FindUser(db, identifier)
	if err != nil {
		users, uErr := forumDB.FetchUsersBy(db, "username", identifier)
		if uErr != nil || len(users) == 0 {
			logging.Logger.Printf("[LOGIN] FindUser error: %v", err)
			ErrorHandler(w, r, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		user = users[0]
	}

	if user.Password != password {
		logging.Logger.Printf("[LOGIN] Incorrect password for Identifier=%s", identifier)
		ErrorHandler(w, r, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	session, _ := forumDB.FetchSessionByUser(db, int64(user.ID))
	if session != (forumDB.Session{}) {
		logging.Logger.Printf("[LOGIN] User %s is already logged in", user.Username)
		ErrorHandler(w, r, http.StatusUnauthorized, "User is already logged in")
		return
	}

	// Init session ID cookie
	err = cookies.WriteSessionCookie(w, r, int64(user.ID))
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("Error writing session cookie : %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusSeeOther)
}
