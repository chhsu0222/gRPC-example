package main

import (
	"io"

	"github.com/chhsu0222/gRPC-example/04-chat/chat"
)

// Connection struct holds one connection
type Connection struct {
	conn chat.Chat_ChatServer
	send chan *chat.ChatMessage
	quit chan struct{}
}

// NewConnection function takes in a stream. It will call the
// start method in another goroutine and return the Connection.
func NewConnection(conn chat.Chat_ChatServer) *Connection {
	c := &Connection{
		conn: conn,
		send: make(chan *chat.ChatMessage),
		quit: make(chan struct{}),
	}
	go c.start()
	return c
}

// Close method closes the quit and send channel.
// It will cause a panic in Send method (sending on a closed channel).
func (c *Connection) Close() error {
	close(c.quit)
	close(c.send)
	return nil
}

// Send method will send the message through the send channel to the "start"
// area. The "start" area will actually send the message along that Connection
// in a thread safe way.
func (c *Connection) Send(msg *chat.ChatMessage) {
	defer func() {
		// Ignore any errors about sending on a closed channel
		recover()
	}()
	c.send <- msg
}

// The only thing that will actually send along the Connection is the start method
// in its own goroutine, which we can only send messages along it through the send
// channel.
func (c *Connection) start() {
	running := true
	for running {
		select {
		case msg := <-c.send:
			c.conn.Send(msg) // Ignoring the error, they just don't get this message.
		case <-c.quit:
			running = false
		}
	}
}

// GetMessages method receives a message and then it sends the message through
// the broadcast channel in a new goroutine.
func (c *Connection) GetMessages(broadcast chan<- *chat.ChatMessage) error {
	for {
		msg, err := c.conn.Recv()
		if err == io.EOF {
			c.Close()
			return nil
		} else if err != nil {
			c.Close()
			return err
		}
		go func(msg *chat.ChatMessage) {
			select {
			case broadcast <- msg:
			case <-c.quit:
			}
		}(msg)
	}
}
