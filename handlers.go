package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"time"

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
	if user.Role != "guest" {
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
	if user.Role == "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	db, _ := sql.Open("sqlite3", "./db/database.db")
	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)

	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func signup(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	if user.Role != "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case "GET":
		tmpls.ExecuteTemplate(w, "signup", user)
	case "POST":
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
	data.PostLikes = getPostsUserLiked(ID)
	data.CommentLikes = getCommentsUserLiked(ID)

	tmpls.ExecuteTemplate(w, "user", data)
}

func newPost(w http.ResponseWriter, r *http.Request) {
	user := getUserByCookie(w, r)
	if user.Role == "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		var data newPostPage
		data.User = user
		data.Categories = getCategoriesList()
		tmpls.ExecuteTemplate(w, "newpost", data)
	case "POST":
		path, err := saveImage(r)
		if err != nil && err.Error() != "http: no such file" {
			w.Write([]byte(err.Error()))
			return
		}

		post := Post{
			AuthorID: user.ID,
			Title:    r.FormValue("title"),
			Data:     r.FormValue("data"),
			Date:     time.Now(),
			Image:    path,
		}

		categories, err := getNewPostCategories(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		id := strconv.Itoa(int(addPostToDB(post, categories)))
		http.Redirect(w, r, "/post/"+id, http.StatusSeeOther)
	}
}
