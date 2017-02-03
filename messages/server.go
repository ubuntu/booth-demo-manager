package messages

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// we just kill the webserver on shutdown, pending requests will just be dropped

// Server maintaining the web socket server
type Server struct {
	// Messages in queue
	Messages   <-chan *Action
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Action
	err        chan error
	quit       chan struct{}
}

// NewServer create a new ws server
func NewServer(url string) *Server {

	s := &Server{
		Messages:   make(chan *Action),
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Action),
		err:        make(chan error),
		quit:       make(chan struct{}),
	}
	http.Handle(url, websocket.Handler(s.onNewClient))

	return s
}

// Send a new message to all clients
func (s *Server) Send(msg *Action) {
	s.broadcast <- msg
}

// Quit the ws server
func (s *Server) Quit() {
	close(s.quit)
}

func (s *Server) onNewClient(conn *websocket.Conn) {

	client, err := NewClient(conn, s)
	if err != nil {
		s.err <- fmt.Errorf("Couldn't accept new connection: %v", err)
		conn.Close()
		return
	}
	s.register <- client

	// Main loop for client
	client.Listen()
}

// Listen to new ws client conn
func (s *Server) Listen() {
	log.Println("Start ws listener...")

	for {
		select {

		// broadcast message to all clients
		case msg := <-s.broadcast:
			log.Println("Send to all clients:", msg)
			for c := range s.clients {
				c.Send(msg)
			}

		// new client connected
		case c := <-s.register:
			log.Println("New client connected")
			s.clients[c] = struct{}{}
			log.Println("Now", len(s.clients), "clients connected.")
			// TODO: send current page msg
			//c.Send()

		// client disconnected
		case c := <-s.unregister:
			log.Println("Disconnected client")
			delete(s.clients, c)

		// error reported
		case err := <-s.err:
			log.Printf("Error: %v\n", err)

		// server shutdown
		case <-s.quit:
			for c := range s.clients {
				c.Quit()
			}
			return
		}
	}
}
