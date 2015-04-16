package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"trace"
)


type room struct {
	forward chan []byte
	join chan *client
	leave chan *client
	clients map[*client]bool
	tracer trace.Tracer
}

func newRoom() *room {
	return &room {
		forward:	make(chan []byte),
		join: 		make(chan *client),
		leave:		make(chan *client),
		clients:	make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New Client Joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client Left")
		case msg := <- r.forward:
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send message to client
					r.tracer.Trace("-- sent to client")
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("-- failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize 	= 1024
	messageBufferSize = 246
)

var upgrader = &websocket.Upgrader { ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize }

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
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