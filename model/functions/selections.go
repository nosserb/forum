package db

import (
	"database/sql"
	"fmt"
)

func FetchUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, email, username, password, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func FetchUsersBy(db *sql.DB, field string, value any) ([]User, error) {
	allowedFields := map[string]bool{
		"id":       true,
		"email":    true,
		"username": true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, email, username, password, created_at
		FROM users
		WHERE ` + field + ` = ?`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func FetchSession(db *sql.DB, sessionID string) (Session, error) {
	var s Session
	err := db.QueryRow(`
        SELECT session_id, user_id, created_at
        FROM sessions
        WHERE session_id = ?`, sessionID).
		Scan(&s.SessionID, &s.UserID, &s.CreatedAt)
	if err != nil {
		return s, err
	}
	return s, nil
}

func FetchSessionByUser(db *sql.DB, userID int64) (Session, error) {
	var s Session
	err := db.QueryRow(`
		SELECT session_id, user_id, created_at
		FROM sessions
		WHERE user_id = ?`, userID).
		Scan(&s.SessionID, &s.UserID, &s.CreatedAt)
	if err != nil {
		return s, err
	}
	return s, nil
}

func FetchUserBySession(db *sql.DB, sessionID string) (User, error) {
	var u User
	err := db.QueryRow(`
        SELECT u.id, u.email, u.username, u.password, u.created_at
        FROM users u
        JOIN sessions s ON u.id = s.user_id
        WHERE s.session_id = ?`, sessionID).
		Scan(&u.ID, &u.Email, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		return u, err
	}
	return u, nil
}

func FetchPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`
		SELECT id, author_id, title, content, likes, dislikes, created_at
		FROM posts
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Likes, &p.Dislikes, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func FetchPostsBy(db *sql.DB, field string, value any) ([]Post, error) {
	allowedFields := map[string]bool{
		"id":         true,
		"author_id":  true,
		"title":      true,
		"likes":      true,
		"dislikes":   true,
		"created_at": true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, author_id, title, content, likes, dislikes, created_at
		FROM posts
		WHERE ` + field + ` = ?`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Likes, &p.Dislikes, &p.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func FetchComments(db *sql.DB) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT id, post_id, author_id, content, created_at, likes, dislikes
		FROM comments
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func FetchCommentsBy(db *sql.DB, field string, value any) ([]Comment, error) {
	allowedFields := map[string]bool{
		"post_id":   true,
		"author_id": true,
		"id":        true,
		"likes":     true,
		"dislikes":  true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, post_id, author_id, content, created_at, likes, dislikes
		FROM comments
		WHERE ` + field + ` = ?`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.Likes, &c.Dislikes); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func FetchCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func FetchPostCategoriesBy(db *sql.DB, field string, value any) ([]PostCategory, error) {
	allowedFields := map[string]bool{
		"post_id":     true,
		"category_id": true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT post_id, category_id
		FROM post_categories
		WHERE ` + field + ` = ?`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postCategories []PostCategory
	for rows.Next() {
		var pc PostCategory
		if err := rows.Scan(&pc.PostID, &pc.CategoryID); err != nil {
			return nil, err
		}
		postCategories = append(postCategories, pc)
	}

	return postCategories, nil
}

func FetchReactions(db *sql.DB) ([]Reaction, error) {
	rows, err := db.Query(`
		SELECT id, user_id, post_id, comment_id, type, created_at
		FROM reactions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.ID, &r.UserID, &r.PostID, &r.CommentID, &r.Type, &r.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	return reactions, nil
}

func FetchReactionsBy(db *sql.DB, field string, value any) ([]Reaction, error) {
	allowedFields := map[string]bool{
		"id":         true,
		"user_id":    true,
		"post_id":    true,
		"comment_id": true,
		"type":       true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, user_id, post_id, comment_id, type, created_at
		FROM reactions
		WHERE ` + field + ` = ?`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		if err := rows.Scan(&r.ID, &r.UserID, &r.PostID, &r.CommentID, &r.Type, &r.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	return reactions, nil
}
