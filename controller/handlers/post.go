package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

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

	switch r.FormValue("action") {
	case "delete":
		DeletePostHandler(w, r)
		return
	case "edit":
		EditPostHandler(w, r)
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
		ErrorHandler(w, r, http.StatusBadRequest, "Post title or content cannot be empty")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(title) > 100 {
		ErrorHandler(w, r, http.StatusBadRequest, "The title must be 100 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(content) > 7500 {
		ErrorHandler(w, r, http.StatusBadRequest, "The content must be 7500 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	// Insert en DB
	postID, err := forumDB.InsertPost(db, int64(user.ID), title, content)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while creating the post")
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

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be logged in to delete a post")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil || postID <= 0 {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post ID")
		logging.Logger.Printf("invalid post id for deletion: %q", r.FormValue("post_id"))
		return
	}

	posts, err := forumDB.FetchPostsBy(db, "id", postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while retrieving the post")
		logging.Logger.Printf("FetchPostsBy error during deletion: %v", err)
		return
	}

	if len(posts) == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("post not found for deletion: %d", postID)
		return
	}

	if posts[0].AuthorID != user.ID {
		ErrorHandler(w, r, http.StatusForbidden, "You can only delete your own posts")
		logging.Logger.Printf("forbidden post deletion attempt user=%d post=%d author=%d", user.ID, postID, posts[0].AuthorID)
		return
	}

	deleted, err := forumDB.DeletePost(db, int64(postID))
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while deleting the post")
		logging.Logger.Printf("DeletePost error: %v", err)
		return
	}

	if deleted == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("no rows deleted for post %d", postID)
		return
	}

	logging.Logger.Printf("[DELETE POST] user=%s post=%d", user.Username, postID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be logged in to edit a post")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil || postID <= 0 {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid post ID")
		logging.Logger.Printf("invalid post id for edit: %q", r.FormValue("post_id"))
		return
	}

	posts, err := forumDB.FetchPostsBy(db, "id", postID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while retrieving the post")
		logging.Logger.Printf("FetchPostsBy error during edit: %v", err)
		return
	}

	if len(posts) == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("post not found for edit: %d", postID)
		return
	}

	if posts[0].AuthorID != user.ID {
		ErrorHandler(w, r, http.StatusForbidden, "You can only edit your own posts")
		logging.Logger.Printf("forbidden post edit attempt user=%d post=%d author=%d", user.ID, postID, posts[0].AuthorID)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	trimmedTitle := strings.TrimSpace(title)
	trimmedContent := strings.TrimSpace(content)

	if trimmedTitle == "" || trimmedContent == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Post title or content cannot be empty")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(title) > 100 {
		ErrorHandler(w, r, http.StatusBadRequest, "The title must be 100 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(content) > 7500 {
		ErrorHandler(w, r, http.StatusBadRequest, "The content must be 7500 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	updated, err := forumDB.UpdatePost(db, int64(postID), title, content)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while editing the post")
		logging.Logger.Printf("UpdatePost error: %v", err)
		return
	}

	if updated == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Post not found")
		logging.Logger.Printf("no rows updated for post %d", postID)
		return
	}

	logging.Logger.Printf("[EDIT POST] user=%s post=%d", user.Username, postID)
	http.Redirect(w, r, "/?post="+strconv.Itoa(postID), http.StatusSeeOther)
}

// Permet de répondre a un post uniquement si on est connecté
// Insère le commentaire dans la DB
func ReplyHandler(
	w http.ResponseWriter,
	r *http.Request,
	sseClients map[int][]chan Notification,
	sseMu *sync.RWMutex,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.FormValue("action") {
	case "edit":
		EditCommentHandler(w, r)
		return
	case "delete":
		DeleteCommentHandler(w, r)
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
		ErrorHandler(w, r, http.StatusBadRequest, "Comment content cannot be empty")
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
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while publishing the comment")
		logging.Logger.Printf("InsertComment error: %v", err)
		return
	}

	var receiverID int
	var subjectLabel string

	posts, err := forumDB.FetchPostsBy(db, "id", postID)
	if err == nil && len(posts) > 0 {
		receiverID = posts[0].AuthorID
		subjectLabel = posts[0].Title
	}

	if receiverID != 0 && receiverID != user.ID {
		notif := Notification{
			ReceiverID:   receiverID,
			SenderID:     user.ID,
			SenderName:   user.Username,
			Type:         "comment",
			SubjectType:  "post",
			SubjectID:    postID,
			SubjectLabel: subjectLabel,
			CreatedAt:    time.Now(),
		}

		SendNotification(notif, sseClients, sseMu)
	}

	logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusSeeOther)
	http.Redirect(w, r, "/?post="+strconv.Itoa(postID), http.StatusSeeOther)
}

func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be logged in to edit a comment")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusUnauthorized)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil || commentID <= 0 {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid comment ID")
		logging.Logger.Printf("invalid comment id for edit: %q", r.FormValue("comment_id"))
		return
	}

	comments, err := forumDB.FetchCommentsBy(db, "id", commentID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while retrieving the comment")
		logging.Logger.Printf("FetchCommentsBy error during edit: %v", err)
		return
	}

	if len(comments) == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Comment not found")
		logging.Logger.Printf("comment not found for edit: %d", commentID)
		return
	}

	comment := comments[0]
	if comment.AuthorID != user.ID {
		ErrorHandler(w, r, http.StatusForbidden, "You can only edit your own comments")
		logging.Logger.Printf("forbidden comment edit attempt user=%d comment=%d author=%d", user.ID, commentID, comment.AuthorID)
		return
	}

	content := r.FormValue("content")
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		ErrorHandler(w, r, http.StatusBadRequest, "Comment content cannot be empty")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	if len(content) > 7500 {
		ErrorHandler(w, r, http.StatusBadRequest, "The content must be 7500 characters or fewer")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusBadRequest)
		return
	}

	updated, err := forumDB.UpdateComment(db, int64(commentID), content)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while editing the comment")
		logging.Logger.Printf("UpdateComment error: %v", err)
		return
	}

	if updated == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Comment not found")
		logging.Logger.Printf("no rows updated for comment %d", commentID)
		return
	}

	logging.Logger.Printf("[EDIT COMMENT] user=%s comment=%d", user.Username, commentID)
	http.Redirect(w, r, "/?post="+strconv.Itoa(comment.PostID), http.StatusSeeOther)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		ErrorHandler(w, r, http.StatusUnauthorized, "You must be logged in to delete a comment")
		logging.Logger.Printf("%v \"%v %v %v\" %v", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, http.StatusUnauthorized)
		return
	}

	commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	if err != nil || commentID <= 0 {
		ErrorHandler(w, r, http.StatusBadRequest, "Invalid comment ID")
		logging.Logger.Printf("invalid comment id for deletion: %q", r.FormValue("comment_id"))
		return
	}

	comments, err := forumDB.FetchCommentsBy(db, "id", commentID)
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while retrieving the comment")
		logging.Logger.Printf("FetchCommentsBy error during deletion: %v", err)
		return
	}

	if len(comments) == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Comment not found")
		logging.Logger.Printf("comment not found for deletion: %d", commentID)
		return
	}

	comment := comments[0]
	if comment.AuthorID != user.ID {
		ErrorHandler(w, r, http.StatusForbidden, "You can only delete your own comments")
		logging.Logger.Printf("forbidden comment deletion attempt user=%d comment=%d author=%d", user.ID, commentID, comment.AuthorID)
		return
	}

	deleted, err := forumDB.DeleteComment(db, int64(commentID))
	if err != nil {
		ErrorHandler(w, r, http.StatusInternalServerError, "An error occurred while deleting the comment")
		logging.Logger.Printf("DeleteComment error: %v", err)
		return
	}

	if deleted == 0 {
		ErrorHandler(w, r, http.StatusNotFound, "Comment not found")
		logging.Logger.Printf("no rows deleted for comment %d", commentID)
		return
	}

	logging.Logger.Printf("[DELETE COMMENT] user=%s comment=%d", user.Username, commentID)
	http.Redirect(w, r, "/?post="+strconv.Itoa(comment.PostID), http.StatusSeeOther)
}
