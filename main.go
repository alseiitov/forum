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

	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/categorie/", categorie)
	http.HandleFunc("/post/", post)
	http.HandleFunc("/user/", user)
	http.HandleFunc("/newpost", newPost)
	http.HandleFunc("/likes/", likes)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Running...")
	fmt.Println(http.ListenAndServe(":"+port, nil))
}
