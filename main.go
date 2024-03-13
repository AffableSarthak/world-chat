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

	// Web socket connection
	room := newRoom()
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

func (room *room) run() {
	for {
		select {
		case client := <-room.Register:
			room.Clients[client] = true
			fmt.Println(room.Clients)
		case client := <-room.Unregister:
			if _, ok := room.Clients[client]; ok {
				delete(room.Clients, client)
				close(client.Send)
			}
			fmt.Println(room.Clients)
		case message := <-room.Broadcast:
			for client := range room.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(room.Clients, client)
				}
			}
		}
	}
}

func newWsConn(room *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	room.run()
	// Error upgrading the connection to a web socket
	if err != nil {
		log.Println(err)
		return
	}

	// Close the conn when
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

	// Create go func, concurrent functions for both reading and writing into the ws conn

	// Reading
	// go func() {
	// 	for {
	// 		select {
	// 			case
	// 		}
	// 	}
	// }
}

// // Single Client Architecture, read and write to the client
// for {
// 	// Read message from the client (JSON)
// 	var wsJson worldChatJson
// 	err = conn.ReadJSON(&wsJson)

// 	if err != nil {
// 		// log.Println("Some error in reading the message from the frontend", err)

// 		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 			log.Printf("error: %v", err)
// 		}
// 		return
// 	}

// 	// DO something with wsJson object
// 	fmt.Println(wsJson)

// 	// Wrtie Message to the client
// 	err = conn.WriteJSON(wsJson)
// 	if err != nil {
// 		// Error writing to the client
// 		log.Println(err)
// 		return
// 	}
// }
