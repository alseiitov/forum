package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	initDB()
	// fillWithSomeData()
	go cleanExpiredSessions()

	images := http.FileServer(http.Dir("./db/images"))
	http.Handle("/images/", http.StripPrefix("/images/", images))
	static := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", static))

	http.HandleFunc("/", all(index))
	http.HandleFunc("/user/", all(user))
	http.HandleFunc("/post/", all(post))
	http.HandleFunc("/categorie/", all(categorie))
	http.HandleFunc("/login", unauthorized(login))
	http.HandleFunc("/signup", unauthorized(signup))
	http.HandleFunc("/newpost", authorized(newPost))
	http.HandleFunc("/logout", authorized(logout))
	http.HandleFunc("/likes/", authorized(likes))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Running...")
	fmt.Println(http.ListenAndServe(":"+port, nil))
}
