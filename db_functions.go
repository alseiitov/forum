package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

var defaultAvatar = "/images/avatars/avatar.jpg"

func initDB() {
	tablesForDB := []string{
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
			author		INTEGER NOT NULL,
			data		TEXT NOT NULL,
			date		DATETIME,
			image		TEXT
		)`,

		`	CREATE	TABLE IF NOT EXISTS comments (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			author_id	INTEGER,
			post_id		INTEGER,
			data		TEXT,
			date		DATETIME
		)`,

		`	CREATE TABLE IF NOT EXISTS sessions (
			user_id		INTEGER NOT NULL,
			uuid		TEXT NOT NULL,
			date		DATETIME
		)`,

		`	CREATE TABLE IF NOT EXISTS categories (
			id			INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
			name		TEXT NOT NULL
		)`,

		`	CREATE TABLE IF NOT EXISTS posts_categories (
			post_id			INTEGER,
			categorie_id	INTEGER
		)`,

		`	CREATE TABLE IF NOT EXISTS posts_likes (
			post_id			INTEGER NOT NULL,
			author_id		INTEGER NOT NULL,
			type			TEXT	NOT NULL
		)`,

		`	CREATE TABLE IF NOT EXISTS comments_likes (
			comment_id		INTEGER NOT NULL,
			author_id		INTEGER NOT NULL,
			type			TEXT	NOT NULL
		)`,
	}

	for _, table := range tablesForDB {
		createDB(table)
	}
}

func createDB(table string) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	stmt, err := db.Prepare(table)
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Exec()
}

func fillWithSomeData() {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	user, _ := db.Prepare("INSERT INTO users (username, password, email, role, avatar) VALUES (?, ?, ?, ?, ?)")
	defer user.Close()
	user.Exec("admin", "$2a$04$Jo85X2JGUOFmF9flUnPhpeTNv6X8AWPUcKtoF4kcHgJBAU3vm3sEi", "aaa@aaa.com", "admin", "/images/avatars/avatar.jpg")
	user.Exec("user", "$2a$04$f9zX9hgA8c3wEwcJJAMDIOwBr1L.tV97tdBuPc02Rq1xucbtVBA16", "user@user.com", "user", "/images/avatars/avatar.jpg")

	post, _ := db.Prepare("INSERT INTO posts (title, author, data, date, image) VALUES (?, ?, ?, ?, ?)")
	defer post.Close()
	post.Exec("Title", 1, "Lorem ipsum", time.Now(), nil)

	categorie, _ := db.Prepare("INSERT INTO categories (name) VALUES (?)")
	defer categorie.Close()
	categorie.Exec("Music")

	postscategories, _ := db.Prepare("INSERT INTO posts_categories (post_id, categorie_id) VALUES (?, ?)")
	defer postscategories.Close()
	postscategories.Exec(1, 1)

	comment, _ := db.Prepare("INSERT INTO comments (author_id, post_id, data, date) VALUES (?, ?, ?, ?)")
	defer comment.Close()
	comment.Exec(1, 1, "comment", time.Now())

	postLike, _ := db.Prepare("INSERT INTO posts_likes (post_id, author_id, type) VALUES (?, ?, ?)")
	defer postLike.Close()
	postLike.Exec(1, 1, "like")
	postLike.Exec(1, 1, "dislike")

	commentLike, _ := db.Prepare("INSERT INTO comments_likes (comment_id, author_id, type) VALUES (?, ?, ?)")
	defer commentLike.Close()
	commentLike.Exec(1, 1, "like")
	commentLike.Exec(1, 1, "dislike")
}

func cleanExpiredSessions() {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	for {
		db.Exec("DELETE FROM sessions WHERE date < $1", time.Now())
		time.Sleep(10 * time.Minute)
	}
}

func addSessionToDB(w http.ResponseWriter, r *http.Request, user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)

	sessionID, _ := uuid.NewV4()
	cookie := &http.Cookie{
		Name:  "session",
		Value: sessionID.String(),
	}
	cookie.MaxAge = 60 * 60 * 24 // 24 hours
	http.SetCookie(w, cookie)

	add, _ := db.Prepare("INSERT INTO sessions (user_id, uuid, date) VALUES (?, ?, ?)")
	defer add.Close()

	add.Exec(user.ID, sessionID, time.Now().Add(24*time.Hour))
}

func addUserToDB(user User) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	add, _ := db.Prepare("INSERT INTO users (username, password, email, role, avatar) VALUES (?, ?, ?, ?, ?)")
	defer add.Close()

	user.Password = encryptPass(user)
	user.Role = "user"
	user.Avatar = defaultAvatar

	add.Exec(user.Username, user.Password, user.Email, user.Role, defaultAvatar)
}

func getCategoriesList() []Categorie {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM categories")
	defer rows.Close()

	var categories []Categorie
	for rows.Next() {
		var c Categorie
		rows.Scan(&c.ID, &c.Name)
		categories = append(categories, c)
	}
	return categories
}

func getCategorieByID(ID int) Categorie {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT * FROM categories WHERE id = $1", ID)

	var c Categorie
	row.Scan(&c.ID, &c.Name)

	return c
}

func getPostsByCategorieID(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT post_id FROM posts_categories WHERE categorie_id = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var postID int
		rows.Scan(&postID)

		post := getPostByID(0, postID)
		post.AuthorUsername = getUserByID(post.AuthorID).Username

		posts = append(posts, post)
	}
	return posts
}

func getPostByID(requesterID int, ID int) Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT * FROM posts WHERE id = $1", ID)

	var p Post
	row.Scan(&p.ID, &p.Title, &p.AuthorID, &p.Data, &p.Date, &p.Image)
	p.AuthorUsername = getUserByID(p.AuthorID).Username
	p.Likes, p.Dislikes, p.Liked, p.Disliked = getLikesByPostID(requesterID, p.ID)

	return p
}

func getCommentByID(requesterID int, ID int) Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT * FROM comments WHERE id = $1", ID)

	var c Comment
	row.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
	c.AuthorUsername = getUserByID(c.AuthorID).Username
	c.PostTitle = getPostByID(0, c.PostID).Title
	c.Likes, c.Dislikes, c.Liked, c.Disliked = getLikesByCommentID(requesterID, c.ID)

	return c
}

func getCommentsByPostID(requesterID int, ID int) []Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT id FROM comments WHERE post_id = $1", ID)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID)
		c = getCommentByID(requesterID, c.ID)
		comments = append(comments, c)
	}
	return comments
}

func getPostsByUserID(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM posts WHERE author = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Image, &p.AuthorID, &p.Data, &p.Date)
		p.AuthorUsername = getUserByID(p.AuthorID).Username
		posts = append(posts, p)
	}
	return posts
}

func getCommentsByUserID(ID int) []Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM comments WHERE author_id = $1", ID)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
		c.AuthorUsername = getUserByID(c.AuthorID).Username
		c.PostTitle = getPostByID(0, ID).Title
		comments = append(comments, c)
	}
	return comments
}

func getPostsUserLiked(ID int) []Post {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT post_id FROM posts_likes WHERE author_id = $1", ID)
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		rows.Scan(&p.ID)
		p = getPostByID(ID, p.ID)
		posts = append(posts, p)
	}
	return posts
}

func getCommentsUserLiked(ID int) []Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT comment_id FROM comments_likes WHERE author_id = $1", ID)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID)
		c = getCommentByID(0, c.ID)
		comments = append(comments, c)
	}
	return comments
}

func getLikesByPostID(requesterID int, ID int) ([]PostLike, []PostLike, bool, bool) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM posts_likes WHERE post_id = $1", ID)
	defer rows.Close()

	var Likes, Dislikes []PostLike
	var Liked, Disliked bool

	for rows.Next() {
		var l PostLike
		rows.Scan(&l.PostID, &l.AuthorID, &l.Type)
		if l.Type == "like" {
			Likes = append(Likes, l)
			if l.AuthorID == requesterID {
				Liked = true
			}
		}
		if l.Type == "dislike" {
			Dislikes = append(Dislikes, l)
			if l.AuthorID == requesterID {
				Disliked = true
			}
		}
	}
	return Likes, Dislikes, Liked, Disliked
}

func getLikesByCommentID(requesterID int, ID int) ([]CommentLike, []CommentLike, bool, bool) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	rows, _ := db.Query("SELECT * FROM comments_likes WHERE comment_id = $1", ID)
	defer rows.Close()

	var Likes, Dislikes []CommentLike
	var Liked, Disliked bool

	for rows.Next() {
		var l CommentLike
		rows.Scan(&l.CommentID, &l.AuthorID, &l.Type)
		if l.Type == "like" {
			Likes = append(Likes, l)
			if l.AuthorID == requesterID {
				Liked = true
			}
		}
		if l.Type == "dislike" {
			Dislikes = append(Dislikes, l)
			if l.AuthorID == requesterID {
				Disliked = true
			}
		}
	}
	return Likes, Dislikes, Liked, Disliked
}

func addPostToDB(post Post, categories []int) int64 {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	add, _ := db.Prepare("INSERT INTO posts (title, author, data, date, image) VALUES (?, ?, ?, ?, ?)")
	defer add.Close()
	add.Exec(post.Title, post.AuthorID, post.Data, post.Date, post.Image)
	last, _ := db.Exec(`SELECT last_insert_rowid();`)
	id, _ := last.LastInsertId()

	for _, cat := range categories {
		addCat, _ := db.Prepare("INSERT INTO posts_categories (post_id, categorie_id) VALUES (?, ?)")
		defer addCat.Close()
		addCat.Exec(id, cat)
	}
	return id
}

func addCommentToDB(comment Comment) {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	add, _ := db.Prepare("INSERT INTO comments (author_id, post_id, data, date) VALUES (?, ?, ?, ?)")
	defer add.Close()
	add.Exec(comment.AuthorID, comment.PostID, comment.Data, comment.Date)
}
