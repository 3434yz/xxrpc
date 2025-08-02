package main

import (
	"fmt"
	"log"

	"xxrpc/client"
)

func main() {
	cli, err := client.Dial(":8888")
	if err != nil {
		log.Fatalf("client dial error: %v", err)
	}

	resp, err := cli.Call("EchoService", "SayHello", "hello world")
	if err != nil {
		log.Fatalf("rpc call error: %v", err)
	}

	fmt.Printf("client got: %v\n", string(resp.Data))
}
