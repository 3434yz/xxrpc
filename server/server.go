package server

import (
	"encoding/binary"
	"io"
	"net"
	"reflect"

	"go.uber.org/zap"

	"xxrpc/codec"
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

func WithLogger(func(*Server) *Server) Option {
	return optionFunc(func(srv *Server) {
		srv.logger = zap.NewExample()
	})
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

func (s *Server) Invoke(req *protocol.Request) (*protocol.Response, error) {
	receiver, method, err := s.registry.Find(req.Service, req.Method)
	if err != nil {
		return nil, err
	}

	methodType := method.Type
	paramType := methodType.In(1)

	// 创建参数实例
	paramPtr := reflect.New(paramType) // 指针类型
	if err := s.codec.Decode(req.Params, paramPtr.Interface()); err != nil {
		return nil, err
	}

	// 调用目标方法：receiver.Method(args)
	results := method.Func.Call([]reflect.Value{receiver, paramPtr.Elem()})

	// 解析返回值
	retValue := results[0].Interface()
	var retErr error
	if !results[1].IsNil() {
		retErr = results[1].Interface().(error)
	}

	if retErr != nil {
		return &protocol.Response{
			Error: retErr.Error(),
		}, nil
	}

	// 编码返回数据
	respData, err := s.codec.Encode(retValue)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		Data:  respData,
		Error: "",
	}, nil
}

func NewServer(addr string, registry *registry.Registry, opts ...Option) *Server {
	s := &Server{
		addr:     addr,
		codec:    codec.JSONCodec{},
		registry: registry,
		logger:   zap.NewExample(),
	}

	for _, opt := range opts {
		opt.Apply(s)
	}

	return s
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

		var req protocol.Request
		if err := s.codec.Decode(data, &req); err != nil {
			return
		}

		resp, err := s.Invoke(&req)
		if err != nil {
			resp = &protocol.Response{
				Error: err.Error(),
			}
		}
		respData, _ := s.codec.Encode(resp)
		binary.Write(conn, binary.BigEndian, uint32(len(respData)))
		conn.Write(respData)
	}
}
