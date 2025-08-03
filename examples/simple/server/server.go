package main

import (
	"log"
	"xxrpc/registry"
	"xxrpc/server"

	"xxrpc/examples/simple/echo"
)

func main() {
	s := server.NewServer(":8888", registry.NewRegister())
	log.Println("RPC Server listening on :8888")

	s.Register("EchoService", &echo.EchoService{})

	if err := s.Start(); err != nil {
		log.Fatalf("server start error: %v", err)
	}
}
