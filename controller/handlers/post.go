package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"forum/controller/logging"

	forumDB "forum/model/functions"
)

type Reply struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	ID        int    `json:"id"`
	Likes     int    `json:"likes"`
	Dislikes  int    `json:"dislikes"`
}

type Post struct {
	ID         int      `json:"id"`
	Username   string   `json:"username"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Replies    []Reply  `json:"replies"`
	CreatedAt  string   `json:"createdAt"`
	Likes      int      `json:"likes"`
	Dislikes   int      `json:"dislikes"`
	Categories []string `json:"categories"`
	ImageID    int      `json:"imageid"`
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
	trimmedTitle := strings.TrimSpace(title)
	trimmedContent := strings.TrimSpace(content)

	// return error if empty title/content
	if trimmedContent == "" || trimmedTitle == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Titre ou contenu du post vide")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(title) > 100 {
		ErrorHandler(w, r, http.StatusBadRequest, "Le titre doit faire au maximum 100 caractères")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(content) > 7500 {
		ErrorHandler(w, r, http.StatusBadRequest, "Le contenu doit faire au maximum 7500 caractères")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	// Insert en DB
	postID, err := forumDB.InsertPost(db, int64(user.ID), title, content)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la création du post")
		logging.Logger.Printf("InsertPost error: %v", err)
		return
	}

	// Insert image if present
	readfile, fileHeader, err := r.FormFile("image")
	if err == nil {
		imageData, err := io.ReadAll(readfile)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la lecture de l'image")
			logging.Logger.Printf("Read image error : %v", err)
			return
		}

		acceptedImageTypes := []string{"image/png", "image/jpeg", "image/gif", "image/svg", "image/webp"}

		imgType := http.DetectContentType(imageData)

		if !slices.Contains(acceptedImageTypes, imgType) {
			ErrorHandler(w, r, 500, "Unsupported image type")
			return
		}

		imageID, err := forumDB.InsertImage(db, postID, fileHeader.Filename, imgType, imageData)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de l'insertion de l'image dans la db")
			logging.Logger.Printf("InsertImage error : %v", err)
			return
		}

		_, err = db.Exec(`
		UPDATE posts SET image_id = ? WHERE id = ?;
		`, imageID, postID)
		if err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors du lien post-image")
			logging.Logger.Printf("Link image-post error : %v", err)
			return
		}
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

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusSeeOther)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Affiche un post spécifique et ses commentaires depuis la DB
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'ID du post
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post ID")
		return
	}

	// Récupérer le post
	posts, err := forumDB.FetchPostsBy(db, "id", postID)
	if err != nil || len(posts) == 0 {
		log.Println(err)
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("Post not found: %d", postID)
		return
	}
	post := posts[0]

	// Récupérer l'auteur du post
	users, err := forumDB.FetchUsersBy(db, "id", post.AuthorID)
	postUsername := "Unknown"
	if err == nil && len(users) > 0 {
		postUsername = users[0].Username
	}

	// Récupérer les catégories du post
	postCategories, err := forumDB.FetchPostCategoriesBy(db, "post_id", postID)
	categories := []string{}
	if err == nil {
		allCategories, err := forumDB.FetchCategories(db)
		if err == nil {
			catMap := make(map[int]string)
			for _, cat := range allCategories {
				catMap[cat.ID] = cat.Name
			}
			for _, pc := range postCategories {
				if name, ok := catMap[pc.CategoryID]; ok {
					categories = append(categories, name)
				}
			}
		}
	}

	// Récupérer les commentaires
	comments, err := forumDB.FetchCommentsBy(db, "post_id", postID)
	if err != nil {
		logging.Logger.Printf("Error fetching comments: %v", err)
		comments = []forumDB.Comment{}
	}

	// Enrichir les commentaires avec les noms d'utilisateur
	replies := []Reply{}
	for _, comment := range comments {
		users, err := forumDB.FetchUsersBy(db, "id", comment.AuthorID)
		username := "Unknown"
		if err == nil && len(users) > 0 {
			username = users[0].Username
		}

		replies = append(replies, Reply{
			ID:        comment.ID,
			Username:  username,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04"),
			Likes:     comment.Likes,
			Dislikes:  comment.Dislikes,
		})
	}

	// Récupérer l'utilisateur connecté
	user := getUserFromCookie(r)

	// Récupérer les réactions de l'utilisateur
	likedPosts := make(map[int]bool)
	dislikedPosts := make(map[int]bool)
	likedComments := make(map[int]bool)
	dislikedComments := make(map[int]bool)

	if user.ID != 0 {
		reactions, err := forumDB.FetchReactionsBy(db, "user_id", user.ID)
		if err == nil {
			for _, reaction := range reactions {
				if reaction.PostID != nil {
					if reaction.Type == "like" {
						likedPosts[*reaction.PostID] = true
					} else if reaction.Type == "dislike" {
						dislikedPosts[*reaction.PostID] = true
					}
				}
				if reaction.CommentID != nil {
					if reaction.Type == "like" {
						likedComments[*reaction.CommentID] = true
					} else if reaction.Type == "dislike" {
						dislikedComments[*reaction.CommentID] = true
					}
				}
			}
		}
	}

	// Récupérer les utilisateurs online
	onlineUsers, err := forumDB.FetchOnlineUsers(db)
	if err != nil {
		logging.Logger.Printf("Error fetching online users: %v", err)
		onlineUsers = []forumDB.User{}
	}

	// Construire le view model
	viewData := struct {
		Username         string         `json:"username"`
		Post             Post           `json:"post"`
		LikedPosts       map[int]bool   `json:"likedPosts"`
		DislikedPosts    map[int]bool   `json:"dislikedPosts"`
		LikedComments    map[int]bool   `json:"likedComments"`
		DislikedComments map[int]bool   `json:"dislikedComments"`
		OnlineUsers      []forumDB.User `json:"onlineUsers"`
	}{
		Username: user.Username,
		Post: Post{
			ID:         post.ID,
			Username:   postUsername,
			Title:      post.Title,
			Content:    post.Content,
			CreatedAt:  post.CreatedAt.Format("2006-01-02 15:04"),
			Likes:      post.Likes,
			Dislikes:   post.Dislikes,
			Categories: categories,
			Replies:    replies,
			ImageID:    post.ImageID,
		},
		LikedPosts:       likedPosts,
		DislikedPosts:    dislikedPosts,
		LikedComments:    likedComments,
		DislikedComments: dislikedComments,
		OnlineUsers:      onlineUsers,
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(viewData); err != nil {
			ErrorHandler(w, r, http.StatusInternalServerError, "Error rendering JSON")
			logging.Logger.Printf("JSON encode error: %v", err)
		}
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusOK)
		return
	}

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusOK)
	http.Redirect(w, r, "/?post="+strconv.Itoa(postID), http.StatusSeeOther)
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
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post id")
		logging.Logger.Printf("Invalid post ID : %v", postID)
		return
	}

	content := r.FormValue("content")
	if strings.TrimSpace(content) == "" {
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
		ErrorHandler(w, r, http.StatusInternalServerError, "Erreur lors de la publication du commentaire")
		logging.Logger.Printf("InsertComment error: %v", err)
		return
	}

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusSeeOther)
	http.Redirect(w, r, "/?post="+strconv.Itoa(postID), http.StatusSeeOther)
}
