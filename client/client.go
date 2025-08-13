package client

import (
	"net"

	"xxrpc/internal/codec"
	"xxrpc/protocol"
)

type Client struct {
	fc    *protocol.FrameConn
	codec codec.Codec
}

func Dial(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	fc := protocol.NewFrameConn(conn)
	return &Client{fc: fc, codec: &codec.JsoniterCodec{}}, nil
}

func (c *Client) Call(serviceMethod string, args any) (*protocol.Response, error) {
	payload, _ := c.codec.Marshal(args)
	req := protocol.Request{
		Method: serviceMethod,
		Params: &payload,
	}

	data, _ := c.codec.Marshal(req)
	if err := c.fc.WriteFrame(data); err != nil {
		return nil, err
	}

	buf, err := c.fc.ReadFrame()
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
	if c.fc != nil {
		return c.fc.Close()
	}
	return nil
}
