package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var tmpls = template.Must(template.ParseGlob("./tmpls/*"))

func index(w http.ResponseWriter, r *http.Request) {
	var data IndexPage
	data.User = getUserByCookie(w, r)
	data.Categories = getCategoriesList()
	tmpls.ExecuteTemplate(w, "index", data)
}

func login(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "login", user)
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		user := getUserByName(username)
		salt := user.Email + user.Username
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(salt+password+salt))
		if err != nil {
			w.Write([]byte("Wrong Pass"))
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
	user := getUserByCookie(w, r)
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "signup", user)
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

func forum(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path[7:], "/")
	fmt.Println(path)
	user := getUserByCookie(w, r)
	switch len(path) {
	case 1:
		var data CategoriePage
		ID, _ := strconv.Atoi(path[0])
		data.ID = ID
		data.Name = getCategorieName(ID)
		data.User = getUserByCookie(w, r)
		data.Posts = getPostsByCategorieID(ID)

		tmpls.ExecuteTemplate(w, "categorie", data)
	case 2:
		var data PostPage
		ID, _ := strconv.Atoi(path[1])
		data.User = user
		data.Post = getPostByID(ID)
		data.Comments = getCommentsByPostID(ID)

		tmpls.ExecuteTemplate(w, "post", data)
	default:
	}
}
