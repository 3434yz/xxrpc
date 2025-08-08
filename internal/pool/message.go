package pool

import (
	"sync"

	"xxrpc/internal/buffer"
	"xxrpc/protocol"
)

var (
	requestPool = sync.Pool{
		New: func() any {
			return &protocol.Request{Params: buffer.GetBuffer()}
		},
	}

	responsePool = sync.Pool{
		New: func() any {
			return &protocol.Response{Data: buffer.GetBuffer()}
		},
	}

	GetRequest = func() *protocol.Request {
		return requestPool.Get().(*protocol.Request)
	}

	GetResponse = func() *protocol.Response {
		return responsePool.Get().(*protocol.Response)
	}

	PutRequest = func(req *protocol.Request) {
		req.Method = ""
		if req.Params != nil {
			buffer.PutBuffer(req.Params)
		}
		requestPool.Put(req)
	}

	PutResponse = func(resp *protocol.Response) {
		if resp.Data != nil {
			buffer.PutBuffer(resp.Data)
		}
		responsePool.Put(resp)
	}
)
