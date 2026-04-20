package db

import (
	"database/sql"
	"fmt"
)

func FetchUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, email, username, first_name, last_name, age, gender, password, created_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.Age, &u.Gender, &u.Password, &u.CreatedAt); err != nil {

			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func FetchUsersBy(db *sql.DB, field string, value any) ([]User, error) {
	allowedFields := map[string]bool{
		"id":         true,
		"email":      true,
		"username":   true,
		"first_name": true,
		"last_name":  true,
		"age":        true,
		"gender":     true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, email, username, first_name, last_name, age, gender, password, created_at
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
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.Age, &u.Gender, &u.Password, &u.CreatedAt); err != nil {
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
        SELECT u.id, u.email, u.username, u.first_name, u.last_name, u.age, u.gender, u.password, u.created_at
        FROM users u
        JOIN sessions s ON u.id = s.user_id
        WHERE s.session_id = ?`, sessionID).
		Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.Age, &u.Gender, &u.Password, &u.CreatedAt)
	if err != nil {
		return u, err
	}
	return u, nil
}

func FetchPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`
		SELECT p.id, p.author_id, p.title, p.content, p.likes, p.dislikes, p.created_at, u.username
		FROM posts p
		JOIN users u ON u.id = p.author_id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Likes, &p.Dislikes, &p.CreatedAt, &p.AuthorUsername)
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
		SELECT p.id, p.author_id, p.title, p.content, p.likes, p.dislikes, p.created_at, u.username, p.image_id
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.` + field + ` = ?
		ORDER BY p.created_at DESC`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post

		// security for null imageID
		var imageID sql.NullInt64

		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &p.Likes, &p.Dislikes, &p.CreatedAt, &p.AuthorUsername, &imageID); err != nil {
			return nil, err
		}

		if imageID.Valid {
			p.ImageID = int(imageID.Int64)
		} else {
			p.ImageID = 0
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

func FetchPrivateMessages(db *sql.DB) ([]PrivateMessage, error) {
	rows, err := db.Query(`
		SELECT id, sender_id, receiver_id, content, read_status, created_at
		FROM private_messages
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var pm PrivateMessage
		if err := rows.Scan(
			&pm.ID,
			&pm.SenderID,
			&pm.ReceiverID,
			&pm.Content,
			&pm.ReadStatus,
			&pm.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, pm)
	}
	return messages, nil
}

func FetchPrivateMessagesBy(db *sql.DB, field string, value any) ([]PrivateMessage, error) {
	allowedFields := map[string]bool{
		"id":          true,
		"sender_id":   true,
		"receiver_id": true,
	}

	if !allowedFields[field] {
		return nil, fmt.Errorf("invalid filter field: %s", field)
	}

	query := `
		SELECT id, sender_id, receiver_id, content, read_status, created_at
		FROM private_messages
		WHERE ` + field + ` = ?
		ORDER BY created_at DESC`

	rows, err := db.Query(query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var pm PrivateMessage
		if err := rows.Scan(
			&pm.ID,
			&pm.SenderID,
			&pm.ReceiverID,
			&pm.Content,
			&pm.ReadStatus,
			&pm.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, pm)
	}
	return messages, nil
}

func FetchPrivateMessagesBetween(db *sql.DB, userA, userB int64, limit, offset int) ([]PrivateMessage, error) {
	rows, err := db.Query(`
		SELECT id, sender_id, receiver_id, content, read_status, created_at
		FROM private_messages
		WHERE 
			(sender_id = ? AND receiver_id = ?)
			OR
			(sender_id = ? AND receiver_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userA, userB, userB, userA, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var pm PrivateMessage
		if err := rows.Scan(
			&pm.ID,
			&pm.SenderID,
			&pm.ReceiverID,
			&pm.Content,
			&pm.ReadStatus,
			&pm.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, pm)
	}
	return messages, nil
}

func FetchPrivateMessageCorrespondents(db *sql.DB, userID int64) ([]int64, error) {
	rows, err := db.Query(`
		SELECT DISTINCT
			CASE
				WHEN sender_id = ? THEN receiver_id
				ELSE sender_id
			END AS correspondent_id
		FROM private_messages
		WHERE sender_id = ? OR receiver_id = ?
	`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var correspondents []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		correspondents = append(correspondents, id)
	}

	return correspondents, nil
}

func FetchOnlineUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query(`
		SELECT DISTINCT u.id, u.email, u.username, u.first_name, u.last_name, u.age, u.gender, u.password, u.created_at
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		ORDER BY u.username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.Age, &u.Gender, &u.Password, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func FetchImage(db *sql.DB, imageID int64) (Image, error) {
	var img Image
	err := db.QueryRow(`
        SELECT id, post_id, name, type, data
        FROM images
        WHERE id = ?`, imageID).
		Scan(&img.ID, &img.PostID, &img.Name, &img.Type, &img.Data)
	if err != nil {
		return img, err
	}
	return img, nil
}
