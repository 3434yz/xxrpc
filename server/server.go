package server

import (
	"encoding/binary"
	"io"
	"net"
	"reflect"

	"xxrpc/codec"
	"xxrpc/protocol"
)

type Server struct {
	addr     string
	codec    codec.JSONCodec
	registry *Registry
}

func (s *Server) Register(name string, service any) {
	s.registry.Register(name, service)
}

func (s *Server) Invoke(req *protocol.Request) (*protocol.Response, error) {
	method, err := s.registry.Find(req.Service, req.Method)
	if err != nil {
		return nil, err
	}

	methodType := method.method.Type
	paramType := methodType.In(1) // 第一个参数是接收者，第二个参数是实际的参数

	paramPtr := reflect.New(paramType)
	if err := s.codec.Decode(req.Params, paramPtr.Interface()); err != nil {
		return nil, err
	}

	results := method.method.Func.Call([]reflect.Value{method.receiver, paramPtr.Elem()})

	retValue := results[0].Interface()
	var retErr error
	if !results[1].IsNil() {
		retErr = results[1].Interface().(error)
	}

	// 如果有错误
	if retErr != nil {
		return &protocol.Response{
			Error: retErr.Error(),
		}, nil
	}

	// 编码返回值
	respData, err := s.codec.Encode(retValue)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		Data:  respData,
		Error: "",
	}, nil
}

func New(addr string) *Server {
	return &Server{
		addr:     addr,
		codec:    codec.JSONCodec{},
		registry: NewRegister(),
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
