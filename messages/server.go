/*
Copyright 2017 Canonical Ltd.
This file is part of booth-demo-manager.

booth-demo-manager is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, version 3 of the License.

Foobar is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with booth-demo-manager.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	Messages      chan *Action
	clients       map[*Client]struct{}
	register      chan *Client
	unregister    chan *Client
	broadcast     chan *Action
	err           chan error
	quit          chan struct{}
	newClientFunc SendNewClient
	res           string
}

// SendNewClient handler for init message on connected new client
type SendNewClient func(c *Client)

// NewServer create a new ws server
func NewServer(url string, newClientFunc SendNewClient) *Server {

	s := &Server{
		Messages:      make(chan *Action),
		clients:       make(map[*Client]struct{}),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan *Action),
		err:           make(chan error),
		quit:          make(chan struct{}),
		newClientFunc: newClientFunc,
		res:           url,
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
		s.err <- fmt.Errorf("Couldn't accept new connection on %s: %v", s.res, err)
		conn.Close()
		return
	}
	s.register <- client

	// Main loop for client
	client.Listen()
}

// Listen to new ws client conn
func (s *Server) Listen() {
	log.Println("Start ws listener on", s.res)

	for {
		select {

		// broadcast message to all clients
		case msg := <-s.broadcast:
			log.Println("Send to all", s.res, "clients:", msg)
			for c := range s.clients {
				c.Send(msg)
			}

		// new client connected
		case c := <-s.register:
			log.Println("New client connected on", s.res)
			s.clients[c] = struct{}{}
			log.Println(len(s.clients), "clients connected on", s.res)
			// Send init message(s)
			s.newClientFunc(c)

		// client disconnected
		case c := <-s.unregister:
			log.Println("Disconnected client from", s.res)
			delete(s.clients, c)

		// error reported
		case err := <-s.err:
			log.Printf("Error: on %s %v\n", s.res, err)

		// server shutdown
		case <-s.quit:
			log.Printf("Quit server")
			for c := range s.clients {
				c.Quit()
			}
			// TODO: proper shutdown: wait for all client to be unregistered before shutting down goroutine
			// (and add goroutine lock)
			// We'll only do it once we upgrade to 1.8 to have easy shutdownable main webserver
			return
		}
	}
}
