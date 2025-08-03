package server

import (
	"encoding/binary"
	"io"
	"net"
	"reflect"

	"xxrpc/codec"
	"xxrpc/protocol"
	"xxrpc/registry"
)

type Server struct {
	addr     string
	codec    codec.JSONCodec
	registry *registry.Registry
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

func NewServer(addr string, registry *registry.Registry) *Server {
	return &Server{
		addr:     addr,
		codec:    codec.JSONCodec{},
		registry: registry,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
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
