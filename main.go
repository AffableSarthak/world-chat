package main

import (
	"log"
	"net/http"
)

// World chat is public chat forum, go to (worldchat.com/any-forum-name), anyone can join and start chatting.
func main() {

	// Serve the frontend for the application
	http.HandleFunc("GET /", serveFrontend)

	// Listen and serve
	log.Println("Server started, go to link http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Error listening and serving: ", err)
	}
}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend")
}
