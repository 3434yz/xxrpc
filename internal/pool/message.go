package pool

import (
	"sync"

	"xxrpc/protocol"
)

var (
	requestPool = sync.Pool{
		New: func() any {
			return &protocol.Request{}
		},
	}

	responsePool = sync.Pool{
		New: func() any {
			return &protocol.Response{}
		},
	}

	GetRequest = func() *protocol.Request {
		return requestPool.Get().(*protocol.Request)
	}

	GetResponse = func() *protocol.Response {
		return responsePool.Get().(*protocol.Response)
	}

	PutRequest = func(req *protocol.Request) {
		req.Reset()
		requestPool.Put(req)
	}

	PutResponse = func(resp *protocol.Response) {
		resp.Reset()
		responsePool.Put(resp)
	}
)
