package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Role     string
	Avatar   string
}

type Session struct {
	UserID int
	UUID   string
	Date   time.Time
}

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

var tmpls = template.Must(template.ParseGlob("./tmpls/*"))

func main() {
	for _, table := range tablesForDB {
		createDB(table)
	}
	go cleanExpiredSessions()

	images := http.FileServer(http.Dir("./db/images"))
	http.Handle("/images/", http.StripPrefix("/images/", images))
	static := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", static))

	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/secret", secret)

	fmt.Println("Running...")
	log.Println(http.ListenAndServe(":8080", nil))

}

func index(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	tmpls.ExecuteTemplate(w, "index", user)
}

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "login", nil)
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		user := getUserByName(username)
		salt := user.Email + user.Username
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(salt+password+salt))
		if err != nil {
			w.Write([]byte("Wrong Pass"))
			return
		} else {
			addSessionToDB(w, r, user)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	if user.ID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	db, _ := sql.Open("sqlite3", "./db/database.db")
	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "signup", nil)
	case "POST":
		r.ParseForm()
		user := User{
			Username: r.FormValue("username"),
			Password: r.FormValue("password2"),
			Email:    r.FormValue("email"),
		}
		err := checkNewUser(user)
		if err == "" {
			addUserToDB(user)
			addSessionToDB(w, r, user)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			w.Write([]byte(err))
		}
	}
}

func secret(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	if user.Role == "user" {
		w.Write([]byte("ok"))
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

func checkNewUser(user User) string {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	username := db.QueryRow("SELECT username FROM users WHERE username = $1", user.Username)
	email := db.QueryRow("SELECT email FROM users WHERE email = $1", user.Email)
	c := User{}
	username.Scan(&c.Username)
	email.Scan(&c.Email)

	if c.Username != "" {
		return "Username is already in use, please try again!"
	}
	if c.Email != "" {
		return "E-mail is already in use, please try again!"
	}

	return ""
}

func addUserToDB(user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	add, _ := db.Prepare("INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)")
	user.Password = encryptPass(user)
	add.Exec(user.Username, user.Password, user.Email, "user")
}

func getUserByCookie(w http.ResponseWriter, req *http.Request) User {
	userCookie, err := req.Cookie("session")
	if err != nil {
		sessionID, _ := uuid.NewV4()
		userCookie = &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
		}
	}
	userCookie.MaxAge = 60 * 60 * 24
	http.SetCookie(w, userCookie)

	session := getSessionByUUID(userCookie.Value)
	user := getUserByID(session.UserID)

	if user.Role == "" {
		user.Role = "guest"
	}

	return user
}

func getSessionByUUID(uuid string) Session {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	data := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", uuid)
	var session Session
	data.Scan(&session.UserID, &session.UUID, &session.Date)
	return session
}

func getUserByID(id int) User {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	data := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	var user User
	data.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Avatar)

	return user
}

func getUserByName(username string) User {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	data := db.QueryRow("SELECT * FROM users WHERE username = $1", username)
	var user User
	data.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Avatar)

	return user
}

func encryptPass(user User) string {
	salt := user.Email + user.Username
	encryptedPass, _ := bcrypt.GenerateFromPassword([]byte(salt+user.Password+salt), bcrypt.MinCost)

	return string(encryptedPass)
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
