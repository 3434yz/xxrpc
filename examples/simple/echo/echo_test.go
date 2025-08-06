package echo

import (
	"testing"
	"xxrpc/internal/codec"
	"xxrpc/registry"
)

func BenchmarkEchoHandler(b *testing.B) {
	codec := &codec.JsoniterCodec{}
	svc := &EchoService{}
	r := registry.NewRegister()
	r.Register(svc, codec)

	handler, _ := r.Find("EchoService.SayHello")

	req := &SayHelloReq{Message: "hello"}
	reqBytes, _ := codec.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler(reqBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}
