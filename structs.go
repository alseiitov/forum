package main

import "time"

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

type Categorie struct {
	ID   int
	Name string
}

type Post struct {
	ID             int
	Title          string
	Image          string
	AuthorID       int
	AuthorUsername string
	Data           string
	Date           time.Time
	Likes          int
}

type Comment struct {
	ID             int
	AuthorID       int
	AuthorUsername string
	PostID         int
	Data           string
	Date           time.Time
}

type IndexPage struct {
	User       User
	Categories []Categorie
}

type CategoriePage struct {
	ID    int
	Name  string
	User  User
	Posts []Post
}
type PostPage struct {
	User     User
	Post     Post
	Comments []Comment
}
type ProfilePage struct {
	User     User
	Profile  User
	Posts    []Post
	Comments []Comment
}
