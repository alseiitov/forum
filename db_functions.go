package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

var defaultAvatar = "/images/avatars/avatar.jpg"

func initDB() {
	tablesForDB := []string{
		`	CREATE TABLE IF NOT EXISTS users (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
			username	TEXT UNIQUE NOT NULL, 
			password	TEXT NOT NULL, 
			email		TEXT UNIQUE NOT NULL,
			role		TEXT,
			avatar		TEXT
		)`,

		`	CREATE TABLE IF NOT EXISTS posts (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			title		TEXT NOT NULL,
			image		TEXT,
			author		INTEGER NOT NULL,
			data		TEXT,
			date		DATETIME,
			likes		INTEGER
		)`,

		`	CREATE	TABLE IF NOT EXISTS comments (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			author_id	INTEGER,
			post_id		INTEGER,
			data		TEXT,
			date		DATETIME
		)`,

		`	CREATE TABLE IF NOT EXISTS sessions (
			user_id		INTEGER NOT NULL,
			uuid		TEXT NOT NULL,
			date		DATETIME
		)`,

		`	CREATE TABLE IF NOT EXISTS categories (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			name		TEXT NOT NULL
		)`,

		`	CREATE TABLE IF NOT EXISTS posts_categories (
			post_id			INTEGER,
			categorie_id	INTEGER
		)`,
	}

	for _, table := range tablesForDB {
		createDB(table)
	}
}

func createDB(table string) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	stmt, err := db.Prepare(table)
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec()
}

func cleanExpiredSessions() {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	for {
		db.Exec("DELETE FROM sessions WHERE date < $1", time.Now())
		time.Sleep(10 * time.Minute)
	}
}

func addSessionToDB(w http.ResponseWriter, r *http.Request, user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)

	sessionID, _ := uuid.NewV4()
	cookie := &http.Cookie{
		Name:  "session",
		Value: sessionID.String(),
	}
	cookie.MaxAge = 60 * 60 * 24 // 24 hours
	http.SetCookie(w, cookie)

	add, _ := db.Prepare("INSERT INTO sessions (user_id, uuid, date) VALUES (?, ?, ?)")
	defer add.Close()

	add.Exec(user.ID, sessionID, time.Now().Add(24*time.Hour))
}

func addUserToDB(user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	add, _ := db.Prepare("INSERT INTO users (username, password, email, role, avatar) VALUES (?, ?, ?, ?, ?)")
	defer add.Close()

	user.Password = encryptPass(user)
	user.Role = "user"
	user.Avatar = defaultAvatar

	add.Exec(user.Username, user.Password, user.Email, user.Role, defaultAvatar)
}

func getCategoriesList() []Categorie {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM categories")
	defer rows.Close()

	var categories []Categorie
	for rows.Next() {
		var c Categorie
		rows.Scan(&c.ID, &c.Name)
		categories = append(categories, c)
	}
	return categories
}

func getCategorieName(ID int) string {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT name FROM categories WHERE id = $1", ID)

	var name string
	row.Scan(&name)

	return name
}

func getPostsByCategorieID(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT post_id FROM posts_categories WHERE categorie_id = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var postID int
		rows.Scan(&postID)

		post := getPostByID(postID)
		post.AuthorUsername = getUserByID(post.AuthorID).Username

		posts = append(posts, post)
	}
	return posts
}

func getPostByID(ID int) Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT * FROM posts WHERE id = $1", ID)

	var p Post
	row.Scan(&p.ID, &p.Title, &p.Image, &p.AuthorID, &p.Data, &p.Date, &p.Likes)
	author := getUserByID(p.AuthorID)
	p.AuthorUsername = author.Username

	return p
}

func getCommentsByPostID(ID int) []Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM comments WHERE post_id = $1", ID)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
		author := getUserByID(c.AuthorID)
		c.AuthorUsername = author.Username
		comments = append(comments, c)
	}
	return comments
}

func getPostsByUserID(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM posts WHERE author = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Image, &p.AuthorID, &p.Data, &p.Date, &p.Likes)
		p.AuthorUsername = getUserByID(p.AuthorID).Username
		posts = append(posts, p)
	}
	return posts
}

func getCommentsByUserID(ID int) []Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM comments WHERE author_id = $1", ID)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
		c.AuthorUsername = getUserByID(c.AuthorID).Username
		comments = append(comments, c)
	}
	return comments
}
