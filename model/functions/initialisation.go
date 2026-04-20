package db

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Age       int       `db:"age"`
	Gender    string    `db:"gender"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

type Session struct {
	SessionID string    `db:"session_id"`
	UserID    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type Post struct {
	ID             int       `db:"id"`
	AuthorID       int       `db:"author_id"`
	AuthorUsername string    `db:"username"`
	Title          string    `db:"title"`
	Content        string    `db:"content"`
	Likes          int       `db:"likes"`
	Dislikes       int       `db:"dislikes"`
	CreatedAt      time.Time `db:"created_at"`
	ImageID        int       `db:"image_id"`
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
	PostID    *int      `db:"post_id"`
	CommentID *int      `db:"comment_id"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}

type PrivateMessage struct {
	ID         int       `db:"id"`
	SenderID   int       `db:"sender_id"`
	ReceiverID int       `db:"receiver_id"`
	Content    string    `db:"content"`
	ReadStatus string    `db:"read_status"`
	CreatedAt  time.Time `db:"created_at"`
}

type Image struct {
	ID     int    `db:"id"`
	PostID *int   `db:"post_id"`
	Name   string `db:"name"`
	Type   string `db:"type"`
	Data   []byte `db:"data"`
}

func Initialisation(database *sql.DB) {
	var err error

	_, err = database.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	username TEXT NOT NULL UNIQUE,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	age INTEGER NOT NULL,
	gender TEXT NOT NULL,
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
		image_id INTEGER REFERENCES images(id),
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

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS private_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		read_status TEXT NOT NULL DEFAULT 'unread' CHECK (read_status IN ('read', 'unread')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users(id),
		FOREIGN KEY (receiver_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		log.Fatalf("Error private_messages table: %v", err)
	}

	_, err = database.Exec(`
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		data LONGBLOB NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id)				
	);
	`)
	if err != nil {
		log.Fatalf("Error private_messages table: %v", err)
	}
}
