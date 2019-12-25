package main

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/chhsu0222/gRPC-example/04-chat/chat"
	"google.golang.org/grpc"
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

// ChatServer has a slice of connections that are currently connected.
// Any time we access the list of connections, we will use the lock
// and do the changes and then unlock it afterwards. It will receive
// messages on the broadcast channel and send them to each of the
// connections.
type ChatServer struct {
	broadcast   chan *chat.ChatMessage
	quit        chan struct{}
	connections []*Connection
	connLock    sync.Mutex
}

// NewChatServer returns the ChatServer. Note that we don't need to bother
// with the slice of connections because the append function will create
// that if needed. We also don't ever have to actually create a sync.Mutex
// because the zero value is valid for it.
// We run the start method in a new goroutine.
func NewChatServer() *ChatServer {
	srv := &ChatServer{
		broadcast: make(chan *chat.ChatMessage),
		quit:      make(chan struct{}),
	}
	go srv.start()
	return srv
}

// Close closes the quit channel
func (c *ChatServer) Close() error {
	close(c.quit)
	return nil
}

// start method is the broadcast handler.
// For each connection, send the message in that goroutine safe way.
// Note that we do it in another goroutine so that it doesn't lock
// up the entire server if some connections being slow.
func (c *ChatServer) start() {
	running := true
	for running {
		select {
		case msg := <-c.broadcast:
			c.connLock.Lock()
			for _, v := range c.connections {
				go v.Send(msg)
			}
			c.connLock.Unlock()
		case <-c.quit:
			running = false
		}
	}
}

// Chat method implements the prototype of Chat service in chat.proto.
// Whenever we've got a stream, whether in or out on a particular method
// in gRPC, in Go it will change that into a single argument that's got
// the Send and Recv methods on it. That's fulfilling the interface.
func (c *ChatServer) Chat(stream chat.Chat_ChatServer) error {
	conn := NewConnection(stream)

	c.connLock.Lock()
	c.connections = append(c.connections, conn)
	c.connLock.Unlock()

	// start to chat
	err := conn.GetMessages(c.broadcast)

	// the connection is done, remove it from the slice
	c.connLock.Lock()
	for i, v := range c.connections {
		if v == conn {
			c.connections = append(c.connections[:i], c.connections[i+1:]...)
		}
	}
	c.connLock.Unlock()

	return err
}

func main() {
	lst, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	srv := NewChatServer()
	chat.RegisterChatServer(s, srv)

	fmt.Println("Now serving at port 8080")
	err = s.Serve(lst)
	if err != nil {
		panic(err)
	}
}
