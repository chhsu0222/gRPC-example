package main

import (
	"context"
	"fmt"

	"github.com/chhsu0222/gRPC-example/03-client/echo"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	// make a connection to the server
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// whenever we got a connection, we need to defer closing that connection.
	defer conn.Close()

	// create a client
	e := echo.NewEchoServerClient(conn)
	// call the service method
	resp, err := e.Echo(ctx, &echo.EchoRequest{
		Message: "Hello world!",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Got from server:", resp.Response)
}
