package db

import (
	"database/sql"
	"time"
)

type InsertType int

const (
	InsertUserType InsertType = iota
	InsertSessionType
	InsertPostType
	InsertCommentType
	InsertCategoryType
	InsertPostCategoryType
	InsertReactionType
	DeleteUserType
	DeleteSessionType
	DeletePostType
	DeleteCommentType
	DeleteCategoryType
	DeletePostCategoryType
	DeleteReactionType
)

type InsertRequest struct {
	Type     InsertType
	Data     interface{}
	RespChan chan int
}

func InsertUser(db *sql.DB, email, username, password string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO users (email, username, password)
		VALUES (?, ?, ?)`, email, username, password)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func InsertSession(db *sql.DB, sessionID string, userID int64) error {
	_, err := db.Exec(`
		INSERT INTO sessions (session_id, user_id, created_at)
		VALUES (?, ?, ?)`, sessionID, userID, time.Now())
	return err
}

func DeleteSession(db *sql.DB, sessionID string) (int64, error) {
	res, err := db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func DeleteAllSessions(db *sql.DB) (int64, error) {
	res, err := db.Exec("DELETE FROM sessions")
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func DeleteUser(db *sql.DB, userID int64) (int64, error) {
	res, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertPost(db *sql.DB, authorID int64, title, content string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO posts (author_id, title, content)
		VALUES (?, ?, ?)`, authorID, title, content)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeletePost(db *sql.DB, postID int64) (int64, error) {
	res, err := db.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertComment(db *sql.DB, postID, authorID int64, content string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO comments (post_id, author_id, content)
		VALUES (?, ?, ?)`, postID, authorID, content)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteComment(db *sql.DB, commentID int64) (int64, error) {
	res, err := db.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertCategory(db *sql.DB, name string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO categories (name)
		VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteCategory(db *sql.DB, categoryID int64) (int64, error) {
	res, err := db.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertPostCategory(db *sql.DB, postID, categoryID int64) error {
	_, err := db.Exec(`
		INSERT INTO post_categories (post_id, category_id)
		VALUES (?, ?)`, postID, categoryID)
	return err
}

func DeletePostCategory(db *sql.DB, postID, categoryID int64) (int64, error) {
	res, err := db.Exec(`
		DELETE FROM post_categories 
		WHERE post_id = ? AND category_id = ?`, postID, categoryID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertReaction(db *sql.DB, userID int64, postID, commentID *int64, reactionType string) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO reactions (user_id, post_id, comment_id, type, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		userID, postID, commentID, reactionType, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteReaction(db *sql.DB, userID int64, postID, commentID *int64, reactionType string) (int64, error) {
	res, err := db.Exec(`
		DELETE FROM reactions
		WHERE user_id = ? 
		  AND post_id IS ? 
		  AND comment_id IS ? 
		  AND type = ?`,
		userID, postID, commentID, reactionType,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func InsertWorker(db *sql.DB, reqChan <-chan InsertRequest) {
	go func() {
		for req := range reqChan {
			var success int
			switch req.Type {
			case InsertUserType:
				if user, ok := req.Data.(User); ok {
					_, err := db.Exec(`INSERT INTO users (email, username, password) VALUES (?, ?, ?)`,
						user.Email, user.Username, user.Password)
					if err == nil {
						success = 1
					}
				}
			case InsertSessionType:
				if session, ok := req.Data.(Session); ok {
					_, err := db.Exec(`INSERT INTO sessions (session_id, user_id, created_at) VALUES (?, ?, ?)`,
						session.SessionID, session.UserID, session.CreatedAt)
					if err == nil {
						success = 1
					}
				}
			case InsertPostType:
				if post, ok := req.Data.(Post); ok {
					_, err := db.Exec(`INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)`,
						post.AuthorID, post.Title, post.Content)
					if err == nil {
						success = 1
					}
				}
			case InsertCommentType:
				if comment, ok := req.Data.(Comment); ok {
					_, err := db.Exec(`INSERT INTO comments (post_id, author_id, content) VALUES (?, ?, ?)`,
						comment.PostID, comment.AuthorID, comment.Content)
					if err == nil {
						success = 1
					}
				}
			case InsertCategoryType:
				if cat, ok := req.Data.(Category); ok {
					_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, cat.Name)
					if err == nil {
						success = 1
					}
				}
			case InsertPostCategoryType:
				if pc, ok := req.Data.(PostCategory); ok {
					_, err := db.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
						pc.PostID, pc.CategoryID)
					if err == nil {
						success = 1
					}
				}
			case InsertReactionType:
				if reaction, ok := req.Data.(Reaction); ok {
					_, err := db.Exec(`INSERT INTO reactions (user_id, post_id, comment_id, type, created_at) VALUES (?, ?, ?, ?, ?)`,
						reaction.UserID, reaction.PostID, reaction.CommentID, reaction.Type, reaction.CreatedAt)
					if err == nil {
						success = 1
					}
				}
			case DeleteUserType:
				if userID, ok := req.Data.(int64); ok {
					res, err := db.Exec(`DELETE FROM users WHERE id = ?`, userID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeleteSessionType:
				if sessionID, ok := req.Data.(string); ok {
					res, err := db.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeletePostType:
				if postID, ok := req.Data.(int64); ok {
					res, err := db.Exec(`DELETE FROM posts WHERE id = ?`, postID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeleteCommentType:
				if commentID, ok := req.Data.(int64); ok {
					res, err := db.Exec(`DELETE FROM comments WHERE id = ?`, commentID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeleteCategoryType:
				if categoryID, ok := req.Data.(int64); ok {
					res, err := db.Exec(`DELETE FROM categories WHERE id = ?`, categoryID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeletePostCategoryType:
				if pc, ok := req.Data.(PostCategory); ok {
					res, err := db.Exec(`DELETE FROM post_categories WHERE post_id = ? AND category_id = ?`,
						pc.PostID, pc.CategoryID)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			case DeleteReactionType:
				if reaction, ok := req.Data.(Reaction); ok {
					res, err := db.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id IS ? AND comment_id IS ? AND type = ?`,
						reaction.UserID, reaction.PostID, reaction.CommentID, reaction.Type)
					if err == nil {
						rows, _ := res.RowsAffected()
						success = int(rows)
					}
				}
			}
			req.RespChan <- success
		}
	}()
}
