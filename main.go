package main

import (
	"log"
	"net/http"
)

func main() {
	initDB()
	// fillWithSomeData()
	handlers()
	go cleanExpiredSessions()

	log.Println("Running...")
	log.Println(http.ListenAndServe(getPort(), nil))
}
