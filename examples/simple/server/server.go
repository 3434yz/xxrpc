package main

import (
	"log"
	"xxrpc/server"
)

type EchoService struct{}

func (s *EchoService) SayHello(msg string) (string, error) {
	return "Echo: " + msg, nil
}

func main() {
	s := server.New(":8888")
	log.Println("RPC Server listening on :8888")

	s.Register("EchoService", &EchoService{})

	if err := s.Start(); err != nil {
		log.Fatalf("server start error: %v", err)
	}
}
