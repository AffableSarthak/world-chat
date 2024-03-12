package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// World chat is public chat forum, go to (worldchat.com/any-forum-name), anyone can join and start chatting.
func main() {

	room := newRoom()
	// Serve the frontend for the application
	http.HandleFunc("GET /", serveFrontend)
	http.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		newWsConn(room, w, r)
	})

	// Listen and serve
	log.Println("Server started, go to link http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Error listening and serving: ", err)
	}
}

func newRoom() *room {
	return &room{
		Clients:    make(map[*client]bool),
		Unregister: make(chan *client),
		Register:   make(chan *client),
		Broadcast:  make(chan []byte),
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
		RoomID       string `json:"roomId"`
		RoomName     string `json:"roomName"`
		Message      string `json:"message"`
		UserIdentity string `json:"userIdentification"`
	}

	client struct {
		Conn *websocket.Conn
		Room *room
		Send chan []byte // channel for each client to write onto the websocket write buffer
	}

	room struct {
		Clients    map[*client]bool
		Register   chan *client
		Unregister chan *client
		Broadcast  chan []byte
	}
)

func newWsConn(room *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)

	_ = room
	// Error upgrading the connection to a web socket
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()
	// conn is now ws connection

	// Create client
	client := &client{
		Room: room,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// Register the client to a room
	client.Room.Register <- client

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
