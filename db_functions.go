package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

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
			categorie	TEXT NOT NULL,
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

	add, _ := db.Prepare("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)")
	defer add.Close()

	user.Password = encryptPass(user)
	add.Exec(user.Username, user.Password, user.Email, "user")
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
	row.Scan(&ID)

	return name
}

func getPostsByCategorieID(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM posts WHERE categorie = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Image, &p.Author, &p.Data, &p.Categorie, &p.Date, &p.Likes)
		author := getUserByID(p.Author)
		p.AuthorUsername = author.Username
		posts = append(posts, p)
	}
	return posts
}
