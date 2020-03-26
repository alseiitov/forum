package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

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
	if user.ID != 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
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
	if user.ID != 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "signup", user)
	case "POST":
		r.ParseForm()
		user := User{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
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

func categorie(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	ID, err := strconv.Atoi(r.URL.Path[11:])
	if err != nil || ID <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var data CategoriePage
	data.ID = ID
	data.User = user
	data.Name = getCategorieName(ID)
	data.Posts = getPostsByCategorieID(ID)

	tmpls.ExecuteTemplate(w, "categorie", data)
}

func post(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	ID, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil || ID <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var data PostPage
	data.User = user
	data.Post = getPostByID(ID)
	data.Comments = getCommentsByPostID(ID)

	fmt.Println(data.Post.Data)
	tmpls.ExecuteTemplate(w, "post", data)
}

func user(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	ID, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil || ID <= 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var data ProfilePage
	data.User = user
	data.Profile = getUserByID(ID)
	data.Posts = getPostsByUserID(ID)
	data.Comments = getCommentsByUserID(ID)
	data.Likes = getLikesByUserID(ID)

	tmpls.ExecuteTemplate(w, "user", data)
}
