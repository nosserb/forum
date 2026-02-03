package db

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

type Session struct {
	SessionID string    `db:"session_id"`
	UserID    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Post struct {
	ID        int       `db:"id"`
	AuthorID  int       `db:"author_id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	Likes     int       `db:"likes"`
	Dislikes  int       `db:"dislikes"`
	CreatedAt time.Time `db:"created_at"`
}

type Comment struct {
	ID        int       `db:"id"`
	PostID    int       `db:"post_id"`
	AuthorID  int       `db:"author_id"`
	Content   string    `db:"content"`
	Likes     int       `db:"likes"`
	Dislikes  int       `db:"dislikes"`
	CreatedAt time.Time `db:"created_at"`
}

type Category struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type PostCategory struct {
	PostID     int `db:"post_id"`
	CategoryID int `db:"category_id"`
}

type Reaction struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	PostID    *int      `db:"post_id"`    // pointeur pour permettre NULL
	CommentID *int      `db:"comment_id"` // pointeur pour permettre NULL
	Type      string    `db:"type"`       // "like" ou "dislike"
	CreatedAt time.Time `db:"created_at"`
}

func Initialisation(database *sql.DB) {
	var err error

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		log.Fatalf("Error users table : %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		log.Fatalf("Error sessions table : %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (author_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		log.Fatalf("Error posts table : %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		author_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (author_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		log.Fatalf("Error comments table : %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`)
	if err != nil {
		log.Fatalf("Error categories table : %v", err)
	}

	// Implémente les catégories
	_, err = database.Exec(`
	INSERT OR IGNORE INTO categories (name) VALUES
	('Gaming'), ('Cook'), ('Anime'), ('Movie'), ('Others');
	`)
	if err != nil {
		log.Fatalf("Error seeding categories : %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS post_categories (
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, category_id),
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);
	`)
	if err != nil {
		log.Fatalf("Error post_categories table : %v", err)
	}

	_, err = database.Exec(`
CREATE TABLE IF NOT EXISTS reactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	post_id INTEGER,
	comment_id INTEGER,
	type TEXT NOT NULL CHECK (type IN ('like', 'dislike')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY (post_id) REFERENCES posts(id),
	FOREIGN KEY (comment_id) REFERENCES comments(id),
	CHECK (
		(post_id IS NOT NULL AND comment_id IS NULL) OR
		(post_id IS NULL AND comment_id IS NOT NULL)
	),
	UNIQUE (user_id, post_id, comment_id)
);
`)
	if err != nil {
		log.Fatalf("Error reactions table : %v", err)
	}

}
