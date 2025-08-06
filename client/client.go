package client

import (
	"encoding/binary"
	"net"

	"xxrpc/internal/codec"
	"xxrpc/protocol"
)

type Client struct {
	conn  net.Conn
	codec codec.JSONCodec
	seq   uint64
}

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Call(serviceMethod string, args any) (*protocol.Response, error) {
	payload, _ := c.codec.Marshal(args)
	req := protocol.Request{
		Method: serviceMethod,
		Params: payload,
	}

	c.seq++

	data, _ := c.codec.Marshal(req)
	binary.Write(c.conn, binary.BigEndian, uint32(len(data)))
	c.conn.Write(data)

	var length uint32
	binary.Read(c.conn, binary.BigEndian, &length)
	buf := make([]byte, length)
	c.conn.Read(buf)

	var resp protocol.Response
	c.codec.Unmarshal(buf, &resp)
	return &resp, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
