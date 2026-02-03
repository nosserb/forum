package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"forum/controller/logging"

	forumDB "forum/model/functions"
)

type Reply struct {
	Username  string
	Content   string
	CreatedAt string
	ID        int
	Likes     int
	Dislikes  int
}

type Post struct {
	ID         int
	Username   string
	Title      string
	Content    string
	Replies    []Reply
	CreatedAt  string
	Likes      int
	Dislikes   int
	Categories []string
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	// return error if empty title/content
	if strings.TrimSpace(content) == "" || strings.TrimSpace(title) == "" {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Titre ou contenu du post vide")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	// Insert en DB
	postID, err := forumDB.InsertPost(db, int64(user.ID), title, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la création du post")
		logging.Logger.Printf("InsertPost error: %v", err)
		return
	}

	// Récupère les catégories sélectionnées (plusieurs valeurs possibles)
	if err := r.ParseForm(); err == nil {
		cats := r.Form["category"]
		for _, cs := range cats {
			if cs == "" {
				continue
			}
			cid, err := strconv.ParseInt(cs, 10, 64)
			if err != nil {
				logging.Logger.Printf("Invalid category id: %v", err)
				continue
			}
			if err := forumDB.InsertPostCategory(db, postID, cid); err != nil {
				logging.Logger.Printf("InsertPostCategory error: %v", err)
			}
		}
	}

	logging.Logger.Printf("[POST] user=%s title=%q", user.Username, title)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Affiche un post spécifique et ses commentaires depuis la DB
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusNotFound)
		ErrorHandler(w, r, http.StatusNotFound, "ID not found")
		logging.Logger.Printf("ID \"%v\" not found", idStr)
		return
	}

	// TO DO : fix to another code
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post ID")
		logging.Logger.Printf("Invalid post ID : %v", id)
		return
	}

	// Récupère le post depuis la DB
	posts, err := forumDB.FetchPostsBy(db, "id", id)
	if err != nil || len(posts) == 0 {
		w.WriteHeader(http.StatusNotFound)
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("Post %v not found", id)
		return
	}
	dbPost := posts[0]

	// Récupère les réactions de l'utilisateur (si connecté) pour marquer les boutons actifs
	likedPosts := map[int]bool{}
	dislikedPosts := map[int]bool{}
	likedComments := map[int]bool{}
	dislikedComments := map[int]bool{}
	user := getUserFromCookie(r)
	if user.Username != "" {
		if reacts, err := forumDB.FetchReactionsBy(db, "user_id", user.ID); err == nil {
			for _, rct := range reacts {
				if rct.PostID != nil {
					pid := *rct.PostID
					if rct.Type == "like" {
						likedPosts[pid] = true
					} else if rct.Type == "dislike" {
						dislikedPosts[pid] = true
					}
				}
				if rct.CommentID != nil {
					cid := *rct.CommentID
					if rct.Type == "like" {
						likedComments[cid] = true
					} else if rct.Type == "dislike" {
						dislikedComments[cid] = true
					}
				}
			}
		}
	}

	// Prend les données de la db pour les rendre prêt a l'affichage
	viewPost := Post{
		ID:        dbPost.ID,
		Title:     dbPost.Title,
		Content:   dbPost.Content,
		Replies:   []Reply{},
		CreatedAt: dbPost.CreatedAt.Format("2006-01-02 15:04"),
		Likes:     dbPost.Likes,
		Dislikes:  dbPost.Dislikes,
	}

	// Récupère les catégories liées au post (noms)
	if pcs, err := forumDB.FetchPostCategoriesBy(db, "post_id", dbPost.ID); err == nil {
		// build map id->name
		catMap := map[int]string{}
		if cats, err := forumDB.FetchCategories(db); err == nil {
			for _, c := range cats {
				catMap[c.ID] = c.Name
			}
		}

		// if there are no categories in DB, provide default names for ids 1..5
		if len(catMap) == 0 {
			for i := 1; i <= 5; i++ {
				catMap[i] = fmt.Sprintf("Cat %d", i)
			}
		}
		for _, rel := range pcs {
			if name, ok := catMap[rel.CategoryID]; ok {
				viewPost.Categories = append(viewPost.Categories, name)
			} else {
				// fallback to id string
				viewPost.Categories = append(viewPost.Categories, strconv.Itoa(rel.CategoryID))
			}
		}
	}

	// Récupérer le username de l'auteur du post
	users, err := forumDB.FetchUsersBy(db, "id", dbPost.AuthorID)
	if err == nil && len(users) > 0 {
		viewPost.Username = users[0].Username
	} else {
		viewPost.Username = fmt.Sprintf("user#%d", dbPost.AuthorID)
	}

	// Récupérer les commentaires du post
	comments, err := forumDB.FetchCommentsBy(db, "post_id", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors du chargement des commentaires")
		logging.Logger.Printf("Erreur lors du chargement des commentaires sur le post %v", id)
		return
	}

	for _, c := range comments {
		// Trouver le username de l'auteur du commentaire
		cusers, err := forumDB.FetchUsersBy(db, "id", c.AuthorID)
		uname := fmt.Sprintf("user#%d", c.AuthorID)
		if err == nil && len(cusers) > 0 {
			uname = cusers[0].Username
		}
		rep := Reply{
			ID:        c.ID,
			Username:  uname,
			Content:   c.Content,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04"),
			Likes:     c.Likes,
			Dislikes:  c.Dislikes,
		}
		viewPost.Replies = append(viewPost.Replies, rep)
	}

	data := struct {
		Username         string
		Post             *Post
		LikedPosts       map[int]bool
		DislikedPosts    map[int]bool
		LikedComments    map[int]bool
		DislikedComments map[int]bool
	}{
		Username:         getUserFromCookie(r).Username,
		Post:             &viewPost,
		LikedPosts:       likedPosts,
		DislikedPosts:    dislikedPosts,
		LikedComments:    likedComments,
		DislikedComments: dislikedComments,
	}

	if err := templates.ExecuteTemplate(w, "post.html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logging.Logger.Printf("Error executing post template: %v", err)
		ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering template")
		return
	}
}

// Permet de répondre a un post uniquement si on est connecté
// Insère le commentaire dans la DB
func ReplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// TO DO : fix to another code
	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post id")
		logging.Logger.Printf("Invalid post ID : %v", postID)
		return
	}

	content := r.FormValue("content")
	if strings.TrimSpace(content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		ErrorHandler(w, r, http.StatusBadRequest, "Contenu du commentaire vide")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}
	// keeping it for later
	/* 	if content == "" {
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
		return
	} */

	_, err = forumDB.InsertComment(db, int64(postID), int64(user.ID), content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la publication du commentaire")
		logging.Logger.Printf("InsertComment error: %v", err)
		return
	}

	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}
