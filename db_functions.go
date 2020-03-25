package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

var tablesForDB = []string{
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

	`	CREATE		TABLE IF NOT EXISTS sessions (
		user_id		INTEGER NOT NULL,
		uuid		TEXT NOT NULL,
		date		DATETIME
	)`,
}

func initDB() {
	for _, table := range tablesForDB {
		createDB(table)
	}
}

func createDB(table string) {
	db, _ := sql.Open("sqlite3", "./db/database.db")

	stmt, err := db.Prepare(table)
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec()
}

func cleanExpiredSessions() {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	for {
		db.Exec("DELETE FROM sessions WHERE date < $1", time.Now())
		time.Sleep(10 * time.Minute)
	}
}

func addSessionToDB(w http.ResponseWriter, r *http.Request, user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)

	sessionID, _ := uuid.NewV4()
	cookie := &http.Cookie{
		Name:  "session",
		Value: sessionID.String(),
	}
	cookie.MaxAge = 60 * 60 * 24 // 24 hours
	http.SetCookie(w, cookie)

	add, _ := db.Prepare("INSERT INTO sessions (user_id, uuid, date) VALUES (?, ?, ?)")
	add.Exec(user.ID, sessionID, time.Now().Add(24*time.Hour))
}

func addUserToDB(user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	add, _ := db.Prepare("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)")
	user.Password = encryptPass(user)
	add.Exec(user.Username, user.Password, user.Email, "user")
}
