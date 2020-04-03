package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

var defaultAvatar = "/images/avatars/avatar.jpg"

func initDB() {
	dbSchemes := []string{
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

	for _, table := range dbSchemes {
		err := createDB(table)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func createDB(table string) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {

		return err
	}

	stmt, err := db.Prepare(table)
	defer stmt.Close()
	if err != nil {
		return err
	}

	stmt.Exec()
	return nil
}

func fillWithSomeData() {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	user, _ := db.Prepare("INSERT INTO users (username, password, email, role, avatar) VALUES (?, ?, ?, ?, ?)")
	defer user.Close()
	user.Exec("admin", "$2a$04$Jo85X2JGUOFmF9flUnPhpeTNv6X8AWPUcKtoF4kcHgJBAU3vm3sEi", "aaa@aaa.com", "admin", "/images/avatars/avatar.jpg")
	user.Exec("user", "$2a$04$f9zX9hgA8c3wEwcJJAMDIOwBr1L.tV97tdBuPc02Rq1xucbtVBA16", "user@user.com", "user", "/images/avatars/avatar.jpg")

	categorie, _ := db.Prepare("INSERT INTO categories (name) VALUES (?)")
	defer categorie.Close()
	categorie.Exec("Music")
	categorie.Exec("Games")
	categorie.Exec("Movies, Series")
	categorie.Exec("Books")
	categorie.Exec("News")
	categorie.Exec("IT, Programming")
	categorie.Exec("Other")

}

func cleanExpiredSessions() {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()

	for {
		db.Exec("DELETE FROM sessions WHERE date < $1", time.Now())
		time.Sleep(10 * time.Minute)
	}
}

func addSessionToDB(w http.ResponseWriter, r *http.Request, user User) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	db.Exec("DELETE FROM sessions WHERE user_id = $1", user.ID)

	sessionID, _ := uuid.NewV4()
	cookie := &http.Cookie{
		Name:  "session",
		Value: sessionID.String(),
	}
	cookie.MaxAge = 60 * 60 * 24 // 24 hours
	http.SetCookie(w, cookie)

	add, err := db.Prepare("INSERT INTO sessions (user_id, uuid, date) VALUES (?, ?, ?)")
	defer add.Close()
	if err != nil {
		return err
	}

	add.Exec(user.ID, sessionID, time.Now().Add(24*time.Hour))
	return nil
}

func (user User) InsertIntoDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	add, err := db.Prepare("INSERT INTO users (username, password, email, role, avatar) VALUES (?, ?, ?, ?, ?)")
	defer add.Close()
	if err != nil {
		return err
	}

	user.Password = encryptPass(user)
	user.Role = "user"
	user.Avatar = defaultAvatar

	add.Exec(user.Username, user.Password, user.Email, user.Role, defaultAvatar)
	return nil
}

func getCategoriesList() ([]Categorie, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT * FROM categories")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var categories []Categorie
	for rows.Next() {
		var c Categorie
		rows.Scan(&c.ID, &c.Name)
		categories = append(categories, c)
	}
	return categories, nil
}

func getCategorieByID(ID int) (Categorie, error) {
	var c Categorie

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return c, err
	}

	row := db.QueryRow("SELECT * FROM categories WHERE id = $1", ID)

	row.Scan(&c.ID, &c.Name)

	return c, nil
}

func getPostsByCategorieID(ID int) ([]Post, error) {
	var posts []Post

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return posts, err
	}

	rows, err := db.Query("SELECT post_id FROM posts_categories WHERE categorie_id = $1", ID)
	defer rows.Close()
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var postID int
		rows.Scan(&postID)

		post, _ := getPostByID(0, postID)

		u, _ := getUserByID(post.AuthorID)
		post.AuthorUsername = u.Username

		posts = append(posts, post)
	}
	return posts, nil
}

func getPostByID(requesterID int, ID int) (Post, error) {
	var p Post

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return p, err
	}

	row := db.QueryRow("SELECT * FROM posts WHERE id = $1", ID)

	row.Scan(&p.ID, &p.Title, &p.AuthorID, &p.Data, &p.Date, &p.Image)
	u, _ := getUserByID(p.AuthorID)
	p.AuthorUsername = u.Username
	p.Likes, p.Dislikes, p.Liked, p.Disliked = getLikesByPostID(requesterID, p.ID)

	return p, nil
}

func getCommentByID(requesterID int, ID int) Comment {
	db, _ := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()

	row := db.QueryRow("SELECT * FROM comments WHERE id = $1", ID)

	var c Comment
	row.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
	u, _ := getUserByID(c.AuthorID)
	c.AuthorUsername = u.Username
	p, _ := getPostByID(0, c.PostID)
	c.PostTitle = p.Title
	c.Likes, c.Dislikes, c.Liked, c.Disliked = getLikesByCommentID(requesterID, c.ID)

	return c
}

func getCommentsByPostID(requesterID int, ID int) ([]Comment, error) {
	var comments []Comment

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return comments, err
	}

	rows, err := db.Query("SELECT id FROM comments WHERE post_id = $1", ID)
	defer rows.Close()
	if err != nil {
		return comments, err
	}

	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID)
		c = getCommentByID(requesterID, c.ID)
		comments = append(comments, c)
	}
	return comments, nil
}

func getPostsByUserID(ID int) ([]Post, error) {
	var posts []Post

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return posts, err
	}

	rows, err := db.Query("SELECT * FROM posts WHERE author = $1", ID)
	defer rows.Close()
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Image, &p.AuthorID, &p.Data, &p.Date)
		u, _ := getUserByID(p.AuthorID)
		p.AuthorUsername = u.Username
		posts = append(posts, p)
	}
	return posts, nil
}

func getCommentsByUserID(ID int) ([]Comment, error) {
	var comments []Comment

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return comments, err
	}

	rows, err := db.Query("SELECT * FROM comments WHERE author_id = $1", ID)
	defer rows.Close()
	if err != nil {
		return comments, err
	}

	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID, &c.AuthorID, &c.PostID, &c.Data, &c.Date)
		u, _ := getUserByID(c.AuthorID)
		c.AuthorUsername = u.Username
		p, _ := getPostByID(0, ID)
		c.PostTitle = p.Title
		comments = append(comments, c)
	}
	return comments, nil
}

func getPostsUserLiked(ID int) ([]Post, error) {
	var posts []Post

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return posts, err
	}

	rows, err := db.Query("SELECT post_id FROM posts_likes WHERE author_id = $1", ID)
	defer rows.Close()
	if err != nil {
		return posts, err
	}

	for rows.Next() {
		var p Post
		rows.Scan(&p.ID)
		p, _ = getPostByID(ID, p.ID)
		posts = append(posts, p)
	}
	return posts, nil
}

func getCommentsUserLiked(ID int) ([]Comment, error) {
	var comments []Comment

	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return comments, err
	}

	rows, err := db.Query("SELECT comment_id FROM comments_likes WHERE author_id = $1", ID)
	defer rows.Close()
	if err != nil {
		return comments, err
	}

	for rows.Next() {
		var c Comment
		rows.Scan(&c.ID)
		c = getCommentByID(ID, c.ID)
		comments = append(comments, c)
	}

	return comments, nil
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

func (post Post) InsertIntoDB(categories []int) int64 {
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

func (comment Comment) InsertIntoDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	add, err := db.Prepare("INSERT INTO comments (author_id, post_id, data, date) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	defer add.Close()
	add.Exec(comment.AuthorID, comment.PostID, comment.Data, comment.Date)
	return nil
}

func (like PostLike) InsertIntoDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	add, err := db.Prepare("INSERT INTO posts_likes (post_id, author_id, type) VALUES (?, ?, ?)")
	defer add.Close()
	if err != nil {
		return err
	}

	add.Exec(like.PostID, like.AuthorID, like.Type)
	return nil
}

func (like PostLike) DeleteFromDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM posts_likes WHERE post_id = $1 AND author_id = $2 AND type = $3", like.PostID, like.AuthorID, like.Type)
	if err != nil {
		return err
	}
	return nil
}

func (like CommentLike) InsertIntoDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	add, err := db.Prepare("INSERT INTO comments_likes (comment_id, author_id, type) VALUES (?, ?, ?)")
	defer add.Close()
	if err != nil {
		return err
	}

	add.Exec(like.CommentID, like.AuthorID, like.Type)
	return nil
}

func (like CommentLike) DeleteFromDB() error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM comments_likes WHERE comment_id = $1 AND author_id = $2 AND type = $3", like.CommentID, like.AuthorID, like.Type)
	if err != nil {
		return err
	}
	return nil
}
