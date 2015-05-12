package main

import (
	"github.com/gorilla/websocket"
)

// representative of a single chatting user
// socket: websocket of the client
// send: the channel which messages are sent
// room: the room the client is chatting in
type client struct {
	socket	*websocket.Conn
	send		chan []byte
	room		*room
}

// read function
// Reads from the client using the ReadMessage method
// Sends any receive messages to the forward channel on the room type
// If socket gets error, loop will break and socket will close
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// write function
// Accept and write messages continually using WriteMessage method
// If any error, for loop breaks and socket is closed
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}