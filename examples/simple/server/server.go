package main

import (
	"go.uber.org/zap"

	"xxrpc/examples/simple/echo"
	"xxrpc/internal/codec"
	"xxrpc/registry"
	"xxrpc/server"
)

func main() {
	s := server.NewServer(":8888", registry.NewRegister(),
		server.WithLogger(zap.L().Named("XXRPC")),
		server.WithCodec(&codec.JsoniterCodec{}),
	)
	s.Logger().Info("RPC Server listening on :8888")

	s.Register("EchoService", &echo.EchoService{})

	if err := s.Start(); err != nil {
		s.Logger().Fatal("failed to start server", zap.Error(err))
	}
}
