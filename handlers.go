package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var tmpls = template.Must(template.ParseGlob("./tmpls/*"))

func index(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	var data IndexPage
	data.User = user
	categories, err := getCategoriesList()
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
	data.Categories = categories

	err = tmpls.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	if user.Role != "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	switch r.Method {
	case "GET":
		err := tmpls.ExecuteTemplate(w, "login", user)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}
	case "POST":
		username := strings.ToLower(r.FormValue("username"))
		password := r.FormValue("password")

		if isEmpty(username) || isEmpty(password) {
			http.Error(w, "400 Bad Request, Can't add empty text", http.StatusBadRequest)
			return
		}

		user, err := getUserByName(username)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		if user.ID == 0 {
			w.Write([]byte("User not found"))
			return
		}

		salt := user.Email + user.Username
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(salt+password+salt))
		if err != nil {
			w.Write([]byte("Wrong Pass"))
			return
		}

		err = addSessionToDB(w, r, user)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	if user.Role == "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
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
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	if user.Role != "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		err := tmpls.ExecuteTemplate(w, "signup", user)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		if !isValidUsername(username) {
			w.Write([]byte("Invalid  username"))
			return
		}
		if !isValidPassword(password) {
			w.Write([]byte("Invalid  password"))
			return
		}
		if !isValidEmail(email) {
			w.Write([]byte("Invalid  email"))
			return
		}

		user := User{
			Username: strings.ToLower(username),
			Password: password,
			Email:    strings.ToLower(email),
		}

		err := checkNewUser(user)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		err = user.InsertIntoDB()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = addSessionToDB(w, r, user)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func categorie(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	ID, err := strconv.Atoi(r.URL.Path[11:])
	if err != nil || ID <= 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	var data CategoriePage
	data.ID = ID
	data.User = user
	c, err := getCategorieByID(ID)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	data.Name = c.Name
	if data.Name == "" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	data.Posts, err = getPostsByCategorieID(ID)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpls.ExecuteTemplate(w, "categorie", data)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	ID, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil || ID <= 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		var data PostPage
		data.User = user
		data.Post, err = getPostByID(user.ID, ID)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		if data.Post.ID == 0 {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}
		data.Comments, err = getCommentsByPostID(user.ID, ID)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpls.ExecuteTemplate(w, "post", data)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	case "POST":
		if user.Role == "guest" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		var comment Comment
		commentData := r.FormValue("comment")
		if isEmpty(commentData) {
			http.Error(w, "400 Bad Request, Can't add empty text", http.StatusBadRequest)
			return
		}

		comment.AuthorID = user.ID
		comment.PostID = ID
		comment.Data = commentData
		comment.Date = time.Now()

		err = comment.InsertIntoDB()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/post/"+strconv.Itoa(ID), http.StatusSeeOther)
	}
}

func user(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	ID, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil || ID <= 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	var data ProfilePage
	data.User = user
	var err1, err2, err3, err4, err5 error
	data.Profile, err1 = getUserByID(ID)
	data.Posts, err2 = getPostsByUserID(ID)
	data.Comments, err3 = getCommentsByUserID(ID)
	data.LikedPosts, err4 = getPostsUserLiked(ID)
	data.LikedComments, err5 = getCommentsUserLiked(ID)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	if data.Profile.ID == 0 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	err = tmpls.ExecuteTemplate(w, "user", data)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func newPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	if user.Role == "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case "GET":
		var data newPostPage
		data.User = user
		categories, err := getCategoriesList()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.Categories = categories

		err = tmpls.ExecuteTemplate(w, "newpost", data)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	case "POST":
		path, err := saveImage(r)
		if err != nil && err.Error() != "http: no such file" {
			http.Error(w, "400 Bad Request\n"+err.Error(), http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		data := r.FormValue("data")

		if isEmpty(title) || isEmpty(data) {
			http.Error(w, "400 Bad Request, Can't add empty text", http.StatusBadRequest)
			return
		}

		post := Post{
			AuthorID: user.ID,
			Title:    title,
			Data:     data,
			Date:     time.Now(),
			Image:    path,
		}

		categories, err := getNewPostCategories(r)
		if err != nil || len(categories) == 0 {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}

		id := strconv.Itoa(int(post.InsertIntoDB(categories)))
		http.Redirect(w, r, "/post/"+id, http.StatusSeeOther)
	}
}

func likes(w http.ResponseWriter, r *http.Request) {
	user, err := getUserByCookie(w, r)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	if user.Role == "guest" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pathArr := strings.Split(r.URL.String(), "/")
	if len(pathArr) != 5 {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	likeType := pathArr[2]
	if likeType != "like" && likeType != "dislike" {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	whatToLike := pathArr[3]
	if whatToLike != "post" && whatToLike != "comment" {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	ID, err := strconv.Atoi(pathArr[4])
	if err != nil || ID <= 0 {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	switch whatToLike {
	case "post":
		post, err := getPostByID(user.ID, ID)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		if post.ID == 0 {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		var like PostLike
		like.PostID = ID
		like.AuthorID = user.ID
		switch likeType {
		case "like":
			like.Type = "like"
			if post.Disliked {
				err := like.InsertIntoDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}

				like.Type = "dislike"
				err = like.DeleteFromDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}

			} else {
				if post.Liked {
					err := like.DeleteFromDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				} else {
					err := like.InsertIntoDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				}
			}
		case "dislike":
			like.Type = "dislike"
			if post.Liked {
				err := like.InsertIntoDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}

				like.Type = "like"
				err = like.DeleteFromDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				if post.Disliked {
					err := like.DeleteFromDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				} else {
					err := like.InsertIntoDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				}
			}
		}
		http.Redirect(w, r, "/post/"+strconv.Itoa(post.ID), http.StatusSeeOther)
	case "comment":
		comment := getCommentByID(user.ID, ID)
		if comment.ID == 0 {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		var like CommentLike
		like.CommentID = ID
		like.AuthorID = user.ID
		switch likeType {
		case "like":
			like.Type = "like"
			if comment.Disliked {
				err := like.InsertIntoDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}
				like.Type = "dislike"
				err = like.DeleteFromDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				if comment.Liked {
					err := like.DeleteFromDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				} else {
					err := like.InsertIntoDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				}
			}
		case "dislike":
			like.Type = "dislike"
			if comment.Liked {
				err := like.InsertIntoDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}
				like.Type = "like"
				err = like.DeleteFromDB()
				if err != nil {
					http.Error(w, "500 Internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				if comment.Disliked {
					err := like.DeleteFromDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				} else {
					err := like.InsertIntoDB()
					if err != nil {
						http.Error(w, "500 Internal server error", http.StatusInternalServerError)
						return
					}
				}
			}
		}
		http.Redirect(w, r, "/post/"+strconv.Itoa(comment.PostID), http.StatusSeeOther)
	}
}
