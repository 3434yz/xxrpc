package client

import (
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
		Params: &payload,
	}

	c.seq++

	data, _ := c.codec.Marshal(req)
	if err := protocol.WriteFrame(c.conn, data); err != nil {
		return nil, err
	}

	buf, err := protocol.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}

	var resp protocol.Response
	if err := c.codec.Unmarshal(buf, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
