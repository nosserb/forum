package handlers

import (
	"log"
	"net/http"
	"strconv"

	"forum/controller/logging"
	"forum/model/data"
	forumDB "forum/model/functions"
)

func FilterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed : Use GET", http.StatusMethodNotAllowed)
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	// Récupérer l'utilisateur
	user := getUserFromCookie(r)
	var sessionID string
	if cookie, err := r.Cookie("sessionCookie"); err == nil {
		if _, err := forumDB.FetchSession(db, cookie.Value); err == nil {
			sessionID = cookie.Value
		}
	}

	// init empty data
	filteredData := data.AllData{
		Username:  user.Username,
		UserID:    user.ID,
		SessionID: sessionID,
	}

	allPosts, err := forumDB.FetchPosts(db)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusMethodNotAllowed)
		return
	}

	// range over all posts, adding id to filtered if checking all filters
	for _, post := range allPosts {

		// Categories Filter
		categoriesValue := r.FormValue("Categories")
		if categoriesValue != "none" && categoriesValue != "" {
			categoryID, err := strconv.Atoi(categoriesValue)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
				return
			}
			relationTable, err := forumDB.FetchPostCategoriesBy(db, "category_id", categoryID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
				return
			}
			found := false
			for _, relation := range relationTable {
				if relation.PostID == post.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Created Filter
		if r.FormValue("created") == "on" {
			user := getUserFromCookie(r)
			if user == (forumDB.User{}) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			if post.AuthorID != user.ID {
				continue
			}
		}

		// Liked Filter
		if r.FormValue("liked") == "on" {
			user := getUserFromCookie(r)
			if user == (forumDB.User{}) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			reactions, err := forumDB.FetchReactionsBy(db, "post_id", post.ID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
				return
			}

			found := false
			for _, react := range reactions {
				if react.UserID == user.ID && react.Type == "like" {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Disliked Filter
		if r.FormValue("disliked") == "on" {
			user := getUserFromCookie(r)
			if user == (forumDB.User{}) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			reactions, err := forumDB.FetchReactionsBy(db, "post_id", post.ID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
				return
			}

			found := false
			for _, react := range reactions {
				if react.UserID == user.ID && react.Type == "dislike" {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Commented Filter
		if r.FormValue("commented") == "on" {
			user := getUserFromCookie(r)
			if user == (forumDB.User{}) {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			comments, err := forumDB.FetchCommentsBy(db, "post_id", post.ID)
			if err != nil {
				ErrorHandler(w, r, http.StatusInternalServerError, err.Error())
				logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
				return
			}

			found := false
			for _, comment := range comments {
				if comment.AuthorID == user.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filteredData.ToDisplay.Posts = append(filteredData.ToDisplay.Posts, post)
	}

	FormatPostDates(filteredData.ToDisplay.Posts)

	postCats, categories, err := PostsWithCategories(db, filteredData.ToDisplay.Posts)
	if err != nil {
		logging.Logger.Printf("Error enriching posts with categories: %v", err)
	}
	filteredData.PostCategories = postCats
	filteredData.Categories = categories

	filteredData.Liked, filteredData.Disliked = GetUserReactions(db, user)

	// Récupère les utilisateurs online
	onlineUsers, err := forumDB.FetchOnlineUsers(db)
	if err != nil {
		logging.Logger.Printf("Error fetching online users: %v", err)
		onlineUsers = []forumDB.User{}
	}
	filteredData.OnlineUsers = onlineUsers

	if err := templates.ExecuteTemplate(w, "forum.html", filteredData); err != nil {
		log.Printf("Error executing forum template: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
		return
	}

	//logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusInternalServerError)
	// alt logger
	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.String(), r.Proto, http.StatusOK)
}
