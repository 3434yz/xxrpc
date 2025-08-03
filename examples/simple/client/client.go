package main

import (
	"fmt"
	"log"

	"xxrpc/client"
	"xxrpc/examples/simple/echo"
)

func main() {
	cli, err := client.Dial(":8888")
	if err != nil {
		log.Fatalf("client dial error: %v", err)
	}

	resp, err := cli.Call("EchoService", "SayHello", echo.SayHelloReq{
		Message: "Hello, World!",
	})
	if err != nil {
		log.Fatalf("rpc call error: %v", err)
	}

	fmt.Printf("client got: %v\n", string(resp.Data))
}
