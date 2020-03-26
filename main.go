package main

import (
	"fmt"
	"net/http"
)

func main() {
	initDB()
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

	fmt.Println("Running...")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
