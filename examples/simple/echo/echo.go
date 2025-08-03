package echo

import "fmt"

type EchoService struct{}

type SayHelloReq struct {
	Message string
}

type SayHelloResp struct {
	Message string
}

func (s *EchoService) SayHello(msg *SayHelloReq) (*SayHelloResp, error) {
	return &SayHelloResp{Message: fmt.Sprintf("Echo:%s", msg.Message)}, nil
}
