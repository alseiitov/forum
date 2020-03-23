package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	Password  string
	Email     string
}

func main() {
	createDB()

	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/signup", signup)
	fmt.Println("Running...")
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	template, _ := template.ParseFiles("./tmpls/index.html")
	template.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		template, _ := template.ParseFiles("./tmpls/login.html")
		template.Execute(w, nil)
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		template, _ := template.ParseFiles("./tmpls/signup.html")
		template.Execute(w, nil)
	case "POST":
		r.ParseForm()
		user := User{
			Username:  r.FormValue("username"),
			FirstName: r.FormValue("fname"),
			LastName:  r.FormValue("lname"),
			Password:  r.FormValue("password2"),
			Email:     r.FormValue("email"),
		}
		if checkNewUser(user) {
			addUser(user)
		}
	}
}

func checkNewUser(user User) bool {
	return true
}

func addUser(user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	add, _ := db.Prepare("INSERT INTO users (username, password, firstname, lastname, email) VALUES (?, ?, ?, ?, ?)")
	add.Exec(user.Username, user.Password, user.FirstName, user.LastName, user.Email)
	fmt.Println("added new user")
}

func createDB() {
	db, _ := sql.Open("sqlite3", "./db/database.db")

	users, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS users (
		id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		username	TEXT UNIQUE, 
		password	TEXT, 
		firstname	TEXT, 
		lastname	TEXT, 
		email		TEXT UNIQUE
	)`)
	if err != nil {
		fmt.Println(err.Error())
	}
	users.Exec()
	// users, _ = db.Prepare("INSERT INTO users (firstname, lastname) VALUES (?, ?)")
	// users.Exec("John", "Doe")

	posts, err := db.Prepare(`
			CREATE TABLE IF NOT EXISTS posts (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			title		TEXT,
			author		INTEGER,
			data		TEXT,
			categorie	TEXT,
			date		TEXT,
			likes		INTEGER
	)`)
	if err != nil {
		fmt.Println(err.Error())
	}
	posts.Exec()
	// posts, _ = db.Prepare("INSERT INTO posts (author, categorie) VALUES (?, ?)")
	// posts.Exec(5, "Cars")
}
