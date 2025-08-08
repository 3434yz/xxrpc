package server

import (
	"io"
	"net"

	"go.uber.org/zap"

	"xxrpc/internal/codec"
	"xxrpc/internal/pool"
	"xxrpc/protocol"
	"xxrpc/registry"
)

type Option interface {
	Apply(*Server)
}

// Option is a functional option for configuring the server.
type optionFunc func(*Server)

func (f optionFunc) Apply(srv *Server) {
	f(srv)
}

func WithLogger(logger *zap.Logger) Option {
	return optionFunc(func(srv *Server) {
		srv.logger = logger
	})
}

func WithCodec(c codec.Codec) Option {
	return optionFunc(func(srv *Server) {
		srv.codec = c
	})
}

func NewServer(addr string, registry *registry.Registry, opts ...Option) *Server {
	s := &Server{
		addr:     addr,
		registry: registry,
	}

	for _, opt := range opts {
		opt.Apply(s)
	}

	return s
}

type Server struct {
	addr     string
	codec    codec.Codec
	registry *registry.Registry

	logger *zap.Logger
}

func (s *Server) Register(service registry.Service) {
	s.registry.Register(service, s.codec)
}

func (s *Server) Invoke(req *protocol.Request, resp *protocol.Response) error {
	handler, err := s.registry.Find(req.Method)
	if err != nil {
		resp.Error = err.Error()
		return err
	}

	respData, err := handler(*req.Params)
	if err != nil {
		resp.Error = err.Error()
		return nil
	}

	resp.Data = &respData
	resp.Error = ""
	return nil
}

func (s *Server) Logger() *zap.Logger {
	return s.logger
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Error("failed to start server", zap.Error(err))
		return err
	}
	s.logger.Info("RPC Server listening", zap.String("address", s.addr))
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		data, err := protocol.ReadFrame(conn)
		if err != nil {
			if err != io.EOF {
				s.logger.Error("failed to read frame", zap.Error(err))
			}
			return
		}

		req := pool.GetRequest()
		if err := s.codec.Unmarshal(data, req); err != nil {
			s.logger.Error("failed to decode request", zap.Error(err))
			pool.PutRequest(req)
			return
		}

		resp := pool.GetResponse()
		if err := s.Invoke(req, resp); err != nil {
			resp.Error = err.Error()
		}

		respData, err := s.codec.Marshal(resp)
		if err != nil {
			s.logger.Error("failed to encode response", zap.Error(err))
		} else {
			if err := protocol.WriteFrame(conn, respData); err != nil {
				s.logger.Error("failed to write frame", zap.Error(err))
				return
			}
		}

		pool.PutRequest(req)
		pool.PutResponse(resp)
	}
}
