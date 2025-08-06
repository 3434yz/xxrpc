package main

import (
	"net/http"
	_ "net/http/pprof"

	"go.uber.org/zap"

	"xxrpc/examples/simple/echo"
	"xxrpc/internal/codec"
	"xxrpc/registry"
	"xxrpc/server"
)

func init() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}

func main() {
	s := server.NewServer(":8888", registry.NewRegister(),
		server.WithLogger(zap.L().Named("XXRPC")),
		server.WithCodec(&codec.JsoniterCodec{}),
	)
	s.Logger().Info("RPC Server listening on :8888")

	s.Register(&echo.EchoService{})

	if err := s.Start(); err != nil {
		s.Logger().Fatal("failed to start server", zap.Error(err))
	}
}
