package main

import (
	"context"
	"fmt"
	"net"

	"github.com/chhsu0222/gRPC-example/02-server/echo"
	"google.golang.org/grpc"
)

type EchoServer struct{}

func (e *EchoServer) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{
		Response: "My Echo: " + req.Message,
	}, nil
}

func main() {
	// create a listener
	lst, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	// setup gRPC server
	s := grpc.NewServer()

	srv := &EchoServer{}

	// register echo server with the gRPC server
	echo.RegisterEchoServerServer(s, srv)
	fmt.Println("Now serving at port 8080")

	// tell gRPC to serve on our listener
	err = s.Serve(lst)
	if err != nil {
		panic(err)
	}
}
