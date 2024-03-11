package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// World chat is public chat forum, go to (worldchat.com/any-forum-name), anyone can join and start chatting.
func main() {

	// Serve the frontend for the application
	http.HandleFunc("GET /", serveFrontend)
	http.HandleFunc("GET /ws", newWsConn)

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

// WS connection properties
var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type (
	worldChatJson struct {
		RoomID   string `json:"roomId"`
		RoomName string `json:"roomName"`
		Message  string `json:"message"`
	}
)

func newWsConn(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)

	// Error upgrading the connection to a web socket
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()
	// conn is now ws connection

	// Single Client Architecture, read and write to the client
	for {
		// Read message from the client (JSON)
		var wsJson worldChatJson
		err = conn.ReadJSON(&wsJson)

		if err != nil {
			// log.Println("Some error in reading the message from the frontend", err)

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}

		// DO something with wsJson object
		fmt.Println(wsJson)

		// Wrtie Message to the client
		err = conn.WriteJSON(wsJson)
		if err != nil {
			// Error writing to the client
			log.Println(err)
			return
		}
	}

}

// go func (){
// 	for {
// 		conn.ReadJSON()
// 	}
// }
