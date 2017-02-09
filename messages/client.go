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
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/net/websocket"
)

// Client represents a ws connection
type Client struct {
	server *Server
	conn   *websocket.Conn
	send   chan *Action
	quit   chan struct{}
}

// NewClient creates a new ws client connection
func NewClient(conn *websocket.Conn, server *Server) (*Client, error) {

	if conn == nil {
		return nil, errors.New("ws cannot be nil")
	}

	if server == nil {
		return nil, errors.New("server cannot be nil")
	}

	return &Client{
		server: server,
		conn:   conn,
		send:   make(chan *Action),
		quit:   make(chan struct{}),
	}, nil
}

// Send a message to this client
func (c *Client) Send(msg *Action) {
	select {
	case c.send <- msg:
	// If we couldn't send the message in a second, consider the client died
	case <-time.After(time.Second):
		err := fmt.Errorf("client disconnected abruptely")
		c.server.err <- err
		c.Quit()
	}

}

// Quit close down client connection
func (c *Client) Quit() {
	if c.quit != nil {
		close(c.quit)
	}
}

// Listen Write and Read request via channel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via channel
func (c *Client) listenWrite() {
	for {
		select {

		// send message to the client
		case msg := <-c.send:
			log.Println("Send:", msg)
			websocket.JSON.Send(c.conn, msg)

		// receive quit request
		case <-c.quit:
			c.quit = nil
			c.conn.Close()
			close(c.send)
			c.server.unregister <- c
			return
		}
	}
}

// Listen read request via channel
func (c *Client) listenRead() {
	for {
		var action Action
		err := websocket.JSON.Receive(c.conn, &action)
		if err != nil {
			if err == io.EOF {
				c.Quit()
				return
			}
			c.server.err <- err
		}
		c.server.Messages <- &action
	}
}
