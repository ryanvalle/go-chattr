package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)
// Basic room struct
// * forward: channel that holds incoming messages 
// 			that should be sent to other clients
// * join: safely add clients to clients map
// * leave: safely remove clients from clients map
// * clients: hold all current clients in room
type room struct {
	forward	chan []byte
	join		chan *client
	leave		chan *client
	clients	map[*client]bool
}

// we setup a new room. This can be done inline for we
// want to establish a new room, but setting it as a 
// function of its own allows us to make the setup
// re-usable and more portable
func newRoom() *room {
	return &room{
		forward:	make(chan []byte),
		join:			make(chan *client),
		leave:		make(chan *client),
		clients:	make(map[*client]bool),
	}
}

// run the room
// utilizing the select statement to synchronize or modify
// shared memory
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// Client joins the room
			// We mark the client as true -- low memory
			r.clients[client] = true
		case client := <-r.leave:
			// client leaves the room
			// If we detect a leave event from the client,
			// we delete them from the clients map and close
			// their send channel
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// if a message is received, we loop through all the clients
			// and send the message to the individual clients.
			// if the client is detected as closed, the select statement
			// fallsback to the default and deletes the client from the
			// map and closes their known socket.
			for client := range r.clients {
				select {
				case client.send <- msg:
					//send the message to client
				default:
					// failure to send message
					delete(r.clients, client)
					close(client.send)
				}
			}	
		}
	}
}

// Serve the page
const (
	socketBufferSize 	= 1024
	messageBufferSize = 256
)

// to use websockets, we need to upgrade HTTP
// this upgrades HTTP to use websockets, but sets
// the upgrade as a variable to make it re-usable.
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// when a request comes in, we set the request as
// a socket using upgrade.Upgrader.
// if no errors are found, client is created and
// passed into join channel of current room
// we defer the leave function for when the client is done
// write method is called as go routine to run the method
// in a different thread.
// read method is called on the main thread
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	client := &client {
		socket: socket,
		send: 	make(chan []byte, messageBufferSize),
		room:		r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}