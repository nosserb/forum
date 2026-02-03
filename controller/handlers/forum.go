package handlers

import (
	"net/http"

	"forum/controller/cookies"
	"forum/controller/logging"

	forumDB "forum/model/functions"
)

// ForumData contient le username et les posts
type ForumData struct {
	Username string
	Posts    []forumDB.Post
}

// Récupère l'utilisateur depuis le cookie
func getUserFromCookie(r *http.Request) forumDB.User {
	cookie, err := r.Cookie("sessionCookie")
	if err != nil {
		return forumDB.User{}
	}
	user, err := forumDB.FetchUserBySession(db, cookie.Value)
	if err != nil {
		return forumDB.User{}
	}
	return user
}

// Déconnexion (supprime le cookie)
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	logging.Logger.Println("[LOGOUT] called")
	err := cookies.EndSession(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
