package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"forum/controller/logging"
	forumDB "forum/model/functions"
)

// LikeHandler gère les likes en base et met à jour le compteur dans la table posts
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	// Accept either post_id or comment_id
	pidStr := r.FormValue("post_id")
	cidStr := r.FormValue("comment_id")

	var postID int64
	var commentID int64
	var err error
	hasPost := false
	hasComment := false
	if pidStr != "" {
		var pid int
		pid, err = strconv.Atoi(pidStr)
		if err == nil {
			postID = int64(pid)
			hasPost = true
		}
	}
	if !hasPost && cidStr != "" {
		var cid int
		cid, err = strconv.Atoi(cidStr)
		if err == nil {
			commentID = int64(cid)
			hasComment = true
		}
	}
	if !hasPost && !hasComment {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	// Cherche si l'utilisateur a déjà une réaction sur ce post/comment
	reactions, err := forumDB.FetchReactionsBy(db, "user_id", user.ID)
	if err != nil {
		logging.Logger.Printf("FetchReactionsBy error: %v", err)
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	var existing *forumDB.Reaction
	for _, rr := range reactions {
		if hasPost && rr.PostID != nil && *rr.PostID == int(postID) {
			tmp := rr
			existing = &tmp
			break
		}
		if hasComment && rr.CommentID != nil && *rr.CommentID == int(commentID) {
			tmp := rr
			existing = &tmp
			break
		}
	}

	// Helper to write JSON response when AJAX
	writeCounts := func() {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.Header.Get("Accept") == "application/json" {
			var likes, dislikes int
			if hasPost {
				row := db.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", postID)
				_ = row.Scan(&likes, &dislikes)
			} else if hasComment {
				row := db.QueryRow("SELECT likes, dislikes FROM comments WHERE id = ?", commentID)
				_ = row.Scan(&likes, &dislikes)
			}
			out := map[string]int{"likes": likes, "dislikes": dislikes}
			b, _ := json.Marshal(out)
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
			return
		}
		// fallback redirect
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	}

	// Si déjà like alors retirer
	if existing != nil && existing.Type == "like" {
		if _, err := forumDB.DeleteReaction(db, int64(user.ID), func() *int64 {
			if hasPost {
				return &postID
			}
			return nil
		}(), func() *int64 {
			if hasComment {
				return &commentID
			}
			return nil
		}(), "like"); err != nil {
			logging.Logger.Printf("DeleteReaction error: %v", err)
		} else {
			if hasPost {
				safeDecrement(db, "likes", int(postID))
			} else if hasComment {
				if _, err := db.Exec("UPDATE comments SET likes = CASE WHEN likes > 0 THEN likes - 1 ELSE 0 END WHERE id = ?", commentID); err != nil {
					logging.Logger.Printf("Update comments likes error: %v", err)
				}
			}
		}
		writeCounts()
		return
	}

	// Si dislike déjà existant alors on change (remove dislike)
	if existing != nil && existing.Type == "dislike" {
		if _, err := forumDB.DeleteReaction(db, int64(user.ID), func() *int64 {
			if hasPost {
				return &postID
			}
			return nil
		}(), func() *int64 {
			if hasComment {
				return &commentID
			}
			return nil
		}(), "dislike"); err != nil {
			logging.Logger.Printf("DeleteReaction error: %v", err)
		} else {
			if hasPost {
				safeDecrement(db, "dislikes", int(postID))
			} else if hasComment {
				if _, err := db.Exec("UPDATE comments SET dislikes = CASE WHEN dislikes > 0 THEN dislikes - 1 ELSE 0 END WHERE id = ?", commentID); err != nil {
					logging.Logger.Printf("Update comments dislikes error: %v", err)
				}
			}
		}
	}

	// Inserer le like
	if _, err := forumDB.InsertReaction(db, int64(user.ID), func() *int64 {
		if hasPost {
			return &postID
		}
		return nil
	}(), func() *int64 {
		if hasComment {
			return &commentID
		}
		return nil
	}(), "like"); err != nil {
		logging.Logger.Printf("InsertReaction error: %v", err)
	} else {
		if hasPost {
			if _, err := db.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", postID); err != nil {
				logging.Logger.Printf("Update posts likes error: %v", err)
			}
		} else if hasComment {
			if _, err := db.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", commentID); err != nil {
				logging.Logger.Printf("Update comments likes error: %v", err)
			}
		}
	}

	writeCounts()
}

// DislikeHandler gère les dislikes en base et met à jour le compteur
func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	user := getUserFromCookie(r)
	if user.Username == "" {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	// Accept either post_id or comment_id
	pidStr := r.FormValue("post_id")
	cidStr := r.FormValue("comment_id")

	var postID int64
	var commentID int64
	var err error
	hasPost := false
	hasComment := false
	if pidStr != "" {
		var pid int
		pid, err = strconv.Atoi(pidStr)
		if err == nil {
			postID = int64(pid)
			hasPost = true
		}
	}
	if !hasPost && cidStr != "" {
		var cid int
		cid, err = strconv.Atoi(cidStr)
		if err == nil {
			commentID = int64(cid)
			hasComment = true
		}
	}
	if !hasPost && !hasComment {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	reactions, err := forumDB.FetchReactionsBy(db, "user_id", user.ID)
	if err != nil {
		logging.Logger.Printf("FetchReactionsBy error: %v", err)
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	var existing *forumDB.Reaction
	for _, rr := range reactions {
		if hasPost && rr.PostID != nil && *rr.PostID == int(postID) {
			tmp := rr
			existing = &tmp
			break
		}
		if hasComment && rr.CommentID != nil && *rr.CommentID == int(commentID) {
			tmp := rr
			existing = &tmp
			break
		}
	}

	// Helper to write JSON response when AJAX
	writeCounts := func() {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" || r.Header.Get("Accept") == "application/json" {
			var likes, dislikes int
			if hasPost {
				row := db.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", postID)
				_ = row.Scan(&likes, &dislikes)
			} else if hasComment {
				row := db.QueryRow("SELECT likes, dislikes FROM comments WHERE id = ?", commentID)
				_ = row.Scan(&likes, &dislikes)
			}
			out := map[string]int{"likes": likes, "dislikes": dislikes}
			b, _ := json.Marshal(out)
			w.Header().Set("Content-Type", "application/json")
			w.Write(b)
			return
		}
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
	}

	// Si déjà dislike on retire
	if existing != nil && existing.Type == "dislike" {
		if _, err := forumDB.DeleteReaction(db, int64(user.ID), func() *int64 {
			if hasPost {
				return &postID
			}
			return nil
		}(), func() *int64 {
			if hasComment {
				return &commentID
			}
			return nil
		}(), "dislike"); err != nil {
			logging.Logger.Printf("DeleteReaction error: %v", err)
		} else {
			if hasPost {
				safeDecrement(db, "dislikes", int(postID))
			} else if hasComment {
				if _, err := db.Exec("UPDATE comments SET dislikes = CASE WHEN dislikes > 0 THEN dislikes - 1 ELSE 0 END WHERE id = ?", commentID); err != nil {
					logging.Logger.Printf("Update comments dislikes error: %v", err)
				}
			}
		}
		writeCounts()
		return
	}

	// Si like existe déjà on change
	if existing != nil && existing.Type == "like" {
		if _, err := forumDB.DeleteReaction(db, int64(user.ID), func() *int64 {
			if hasPost {
				return &postID
			}
			return nil
		}(), func() *int64 {
			if hasComment {
				return &commentID
			}
			return nil
		}(), "like"); err != nil {
			logging.Logger.Printf("DeleteReaction error: %v", err)
		} else {
			if hasPost {
				safeDecrement(db, "likes", int(postID))
			} else if hasComment {
				if _, err := db.Exec("UPDATE comments SET likes = CASE WHEN likes > 0 THEN likes - 1 ELSE 0 END WHERE id = ?", commentID); err != nil {
					logging.Logger.Printf("Update comments likes error: %v", err)
				}
			}
		}
	}

	// Inserer le dislike
	if _, err := forumDB.InsertReaction(db, int64(user.ID), func() *int64 {
		if hasPost {
			return &postID
		}
		return nil
	}(), func() *int64 {
		if hasComment {
			return &commentID
		}
		return nil
	}(), "dislike"); err != nil {
		logging.Logger.Printf("InsertReaction error: %v", err)
	} else {
		if hasPost {
			if _, err := db.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", postID); err != nil {
				logging.Logger.Printf("Update posts dislikes error: %v", err)
			}
		} else if hasComment {
			if _, err := db.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", commentID); err != nil {
				logging.Logger.Printf("Update comments dislikes error: %v", err)
			}
		}
	}

	writeCounts()
}

// Le Décement GRACIEUX qui retire un like ou dislike
func safeDecrement(database *sql.DB, field string, postID int) {
	// On empèche d'avoir des valeurs négatif
	query := "UPDATE posts SET " + field + " = CASE WHEN " + field + " > 0 THEN " + field + " - 1 ELSE 0 END WHERE id = ?"
	if _, err := database.Exec(query, postID); err != nil {
		logging.Logger.Printf("safeDecrement error: %v", err)
	}
}
