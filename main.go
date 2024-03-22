package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	client struct {
		Conn *websocket.Conn
		Room *room
		Send chan []byte // channel for each client to write onto the websocket write buffer
	}

	roomInfo struct {
		isActive bool
		clientID string
		roomId   string
		roomName string
	}
	room struct {
		Clients    map[*client]roomInfo
		Register   chan *client
		Unregister chan *client
		Broadcast  chan []byte
	}
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
	log.Println("Server started, go to link http://localhost:8090")
	err := http.ListenAndServe(":8090", nil)

	if err != nil {
		log.Fatal("Error listening and serving: ", err)
	}
}

// Serve the web app
func serveFrontend(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend")
}

// Create a Room, basically each client is part of a room
func newRoom() *room {
	return &room{
		Clients:    make(map[*client]roomInfo),
		Unregister: make(chan *client),
		Register:   make(chan *client),
		Broadcast:  make(chan []byte),
	}
}

// WS connection properties
var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newWsConn(room *room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)

	lul := r.URL.Query()
	var roomID string = r.URL.Query().Get("roomID")
	var roomName string = r.URL.Query().Get("roomName")
	var clientID string = lul.Get("clientID")

	// Instantiate all room related fuctionality
	go room.run(roomID, roomName, clientID)

	// Error upgrading the connection to a web socket
	if err != nil {
		log.Println(err)
		return
	}

	// Close the conn when
	// defer func() {
	// 	fmt.Println("NO MUTEX, CLOSING ASAP lol")
	// 	conn.Close()
	// }()
	// conn is now ws connection
	// Create client
	client := &client{
		Room: room,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// Register the client to a room
	fmt.Println("TRYING TO REGISTER THE CLIENT")
	client.Room.Register <- client

	// Create go func, concurrent functions for both reading and writing into the ws conn
	// Reading
	go client.readData()
	// Writing
	go client.writeData()
}

func (room *room) run(roomId, roomName, clientId string) {
	fmt.Println("INSIDE THE ROOOM:", roomName)
	for {
		select {
		case client := <-room.Register:
			room.Clients[client] = roomInfo{
				isActive: true,
				clientID: clientId,
				roomId:   roomId,
				roomName: roomName,
			}
			fmt.Println("CLIENT REGISTERED", room.Clients[client])

		case client := <-room.Unregister:
			if _, ok := room.Clients[client]; ok {
				delete(room.Clients, client)
				close(client.Send)
			}
		case message := <-room.Broadcast:
			fmt.Println("MESSAGE", message)
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

func (c *client) readData() {
	// defer func() {
	// 	fmt.Println("DOES IT COME HERE <- readData()")
	// 	c.Room.Unregister <- c
	// 	c.Conn.Close()
	// }()
	for {
		_, message, err := c.Conn.ReadMessage()
		fmt.Println(message, "MESSAGE FROM CLIENT")
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}
		select {
		case msg := <-c.Room.Broadcast:
			// if len(msg) == 0 {
			// 	return
			// }
			fmt.Println("MSG:", msg)
			c.Room.Broadcast <- message
		}
	}
}

func (c *client) writeData() {
	// defer func() {
	// 	fmt.Println("DOES IT COME HERE <- writeData()")
	// }()
	for {
		// c.Conn.WriteMessage(websocket.TextMessage, <-c.Send)
		msg := <-c.Send
		fmt.Println("MSG:", msg)
		c.Conn.WriteMessage(websocket.TextMessage, msg)
		// select {
		// case message := <-c.Send:

		// 	c.Conn.WriteMessage(websocket.TextMessage, message)
		// }
	}
}

// // Single Client Architecture, read and write to the client
// for {
// 	// Read message from the client (JSON)

// var wsJson worldChatJson
// err = conn.ReadJSON(&wsJson)
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
