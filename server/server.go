package server

import (
	"encoding/binary"
	"io"
	"net"
	"reflect"

	"go.uber.org/zap"

	"xxrpc/codec"
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

func NewServer(addr string, registry *registry.Registry, opts ...Option) *Server {
	s := &Server{
		addr:     addr,
		codec:    codec.JSONCodec{},
		registry: registry,
	}

	for _, opt := range opts {
		opt.Apply(s)
	}

	return s
}

type Server struct {
	addr     string
	codec    codec.JSONCodec
	registry *registry.Registry

	logger *zap.Logger
}

func (s *Server) Register(name string, service any) {
	s.registry.Register(name, service)
}

func (s *Server) Invoke(req *protocol.Request, resp *protocol.Response) error {
	receiver, method, err := s.registry.Find(req.Service, req.Method)
	if err != nil {
		resp.Error = err.Error()
		return err
	}

	methodType := method.Type
	paramType := methodType.In(1)

	// 创建参数实例
	paramPtr := reflect.New(paramType)
	if err := s.codec.Decode(req.Params, paramPtr.Interface()); err != nil {
		resp.Error = err.Error()
		return err
	}

	// 调用目标方法
	results := method.Func.Call([]reflect.Value{receiver, paramPtr.Elem()})
	retValue := results[0].Interface()

	var retErr error
	if !results[1].IsNil() {
		retErr = results[1].Interface().(error)
	}

	if retErr != nil {
		resp.Error = retErr.Error()
		return nil
	}

	// 编码返回值
	respData, err := s.codec.Encode(retValue)
	if err != nil {
		resp.Error = err.Error()
		return err
	}

	resp.Data = respData
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
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			return
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(conn, data); err != nil {
			return
		}

		var req = pool.GetRequest()
		if err := s.codec.Decode(data, req); err != nil {
			return
		}

		var resp = pool.GetResponse()
		if err := s.Invoke(req, resp); err != nil {
			resp.Error = err.Error()
		}

		respData, err := s.codec.Encode(resp)
		if err != nil {
			s.logger.Error("failed to encode response", zap.Error(err))
		}
		binary.Write(conn, binary.BigEndian, uint32(len(respData)))
		conn.Write(respData)
		pool.PutRequest(req)
		pool.PutResponse(resp)
	}
}
