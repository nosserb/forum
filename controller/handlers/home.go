package handlers

import (
	"net/http"

	"forum/controller/logging"

	"forum/model/data"
	forumDB "forum/model/functions"
)

// Affiche la page forum avec les posts depuis la DB
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		ErrorHandler(w, r, http.StatusNotFound, "Page does not exist")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.String(), r.Proto, http.StatusNotFound)
		return
	}

	// Récupère l'utilisateur (si connecté)
	user := getUserFromCookie(r)

	// Récupère les posts depuis la DB
	dbPosts, err := forumDB.FetchPosts(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la récupération des posts")
		logging.Logger.Printf("FetchPosts error: %v", err)
		return
	}

	FormatPostDates(dbPosts)

	liked, disliked := GetUserReactions(db, user)

	postCats, categories, err := PostsWithCategories(db, dbPosts)
	if err != nil {
		logging.Logger.Printf("Error enriching posts with categories: %v", err)
	}

	// Construire un view model cohérent pour la template
	viewData := data.AllData{
		Username:       user.Username,
		ToDisplay:      data.ToDisplay{Posts: dbPosts},
		Liked:          liked,
		Disliked:       disliked,
		PostCategories: postCats,
		Categories:     categories,
	}

	if err := templates.ExecuteTemplate(w, "forum.html", viewData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
		logging.Logger.Printf("Template execute error: %v", err)
		return
	}
}
