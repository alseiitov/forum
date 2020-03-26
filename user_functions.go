package main

import (
	"database/sql"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func checkNewUser(user User) string {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

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
	defer db.Close()

	data := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", uuid)
	var session Session
	data.Scan(&session.UserID, &session.UUID, &session.Date)
	return session
}

func getUserByID(id int) User {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	data := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	var user User
	data.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Avatar)

	return user
}

func getUserByName(username string) User {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

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
