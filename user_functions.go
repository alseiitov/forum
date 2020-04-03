package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func checkNewUser(user User) error {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	username := db.QueryRow("SELECT username FROM users WHERE username = $1", user.Username)
	email := db.QueryRow("SELECT email FROM users WHERE email = $1", user.Email)
	c := User{}
	username.Scan(&c.Username)
	email.Scan(&c.Email)

	if c.Username != "" {
		return errors.New("username or email is already in use, please try again")
	}

	if c.Email != "" {
		return errors.New("e-mail is already in use, please try again")
	}

	return nil
}

func getUserByCookie(w http.ResponseWriter, req *http.Request) (User, error) {
	userCookie, err := req.Cookie("session")
	if err != nil {
		sessionID, _ := uuid.NewV4()
		userCookie = &http.Cookie{
			Name:  "session",
			Value: sessionID.String(),
		}
		userCookie.MaxAge = 60 * 60 * 24
		http.SetCookie(w, userCookie)
	}

	session := getSessionByUUID(userCookie.Value)
	user, err := getUserByID(session.UserID)
	if err != nil {
		return user, err
	}

	if user.Role == "" {
		user.Role = "guest"
	}

	return user, nil
}

func getSessionByUUID(uuid string) Session {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()

	data := db.QueryRow("SELECT * FROM sessions WHERE uuid = $1", uuid)
	var session Session
	data.Scan(&session.UserID, &session.UUID, &session.Date)
	return session
}

func getUserByID(id int) (User, error) {
	var user User

	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return user, err
	}
	defer db.Close()

	data := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
	data.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Avatar)

	return user, nil
}

func getUserByNameOrEmail(usernameOrEmail string) (User, error) {
	var user User
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return user, err
	}
	defer db.Close()

	data := db.QueryRow("SELECT * FROM users WHERE username = $1 OR email = $1", usernameOrEmail)

	data.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Avatar)

	return user, nil
}

func encryptPass(user User) string {
	salt := user.Email + user.Username
	encryptedPass, _ := bcrypt.GenerateFromPassword([]byte(salt+user.Password+salt), bcrypt.MinCost)

	return string(encryptedPass)
}
