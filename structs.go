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
