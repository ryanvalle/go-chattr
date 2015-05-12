package main

import (
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

// Temp1 Represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	temp1    *template.Template
}

// ServeHTTP handles the HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.temp1 = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.temp1.Execute(w, nil)
}

func main() {
	r := newRoom() // new room is established in room.go
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// start the room as a separate thread so main
	// can be used to run the webserver. Chatting
	// functions are done on a separate thread
	go r.run()
	
	// Start the Web Server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
