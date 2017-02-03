package messages

import (
	"errors"
	"fmt"
	"io"
	"log"

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
	default:
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
