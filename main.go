package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, _ := sql.Open("sqlite3", "./db/database.db")

	users, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
		id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		username	TEXT UNIQUE, 
		firstname	TEXT, 
		lastname	TEXT, 
		email		TEXT UNIQUE
	)`)
	if err != nil {
		fmt.Println(err.Error())
	}
	users.Exec()
	users, _ = db.Prepare("INSERT INTO users (firstname, lastname) VALUES (?, ?)")
	users.Exec("John", "Doe")

	posts, err := db.Prepare(`
			CREATE TABLE IF NOT EXISTS posts (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			title		TEXT,
			author		INTEGER,
			data		TEXT,
			categorie	TEXT,
			date		TEXT,
			likes		INTEGER,
	)`)
	if err != nil {
		fmt.Println(err.Error())
	}
	posts.Exec()
	posts, _ = db.Prepare("INSERT INTO posts (author, categorie) VALUES (?, ?)")
	posts.Exec(5, "Cars")
}
