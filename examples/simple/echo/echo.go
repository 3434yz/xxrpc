package echo

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xxrpc/internal/codec"
	"xxrpc/registry"
)

type EchoService struct{}

func (EchoService) Name() string {
	return "EchoService"
}

func (s *EchoService) Register(r *registry.Registry, codec codec.Codec) {
	r.ServiceMethods[s.Name()+".SayHello"] = &registry.ServiceMethod{
		Handler: func(data []byte) ([]byte, error) {
			var req SayHelloReq
			if err := codec.Unmarshal(data, &req); err != nil {
				return nil, err
			}
			resp, err := s.SayHello(&req)
			if err != nil {
				return nil, err
			}
			return codec.Marshal(resp)
		},
	}

	r.ServiceMethods[s.Name()+".ComplexHello"] = &registry.ServiceMethod{
		Handler: func(data []byte) ([]byte, error) {
			var req ComplexHelloReq
			if err := codec.Unmarshal(data, &req); err != nil {
				return nil, err
			}
			resp, err := s.ComplexHello(&req)
			if err != nil {
				return nil, err
			}
			return codec.Marshal(resp)
		},
	}
}

func (s *EchoService) ComplexHello(req *ComplexHelloReq) (*ComplexHelloResp, error) {
	return &ComplexHelloResp{
		Status:  "ok",
		ReplyID: fmt.Sprintf("ack-%d", time.Now().UnixNano()),
	}, nil
}

// 基础回声方法
func (s *EchoService) SayHello(msg *SayHelloReq) (*SayHelloResp, error) {
	return &SayHelloResp{Message: fmt.Sprintf("Echo:%s", msg.Message)}, nil
}

// 大数据传输测试：生成指定大小的数据
func (s *EchoService) BigData(req *BigDataReq) (*BigDataResp, error) {
	if req.Size <= 0 {
		req.Size = 1024 // 默认1KB
	}
	if req.Count <= 0 {
		req.Count = 1
	}

	// 生成基础数据
	base := req.Data
	if base == "" {
		base = "default-data-"
	}

	// 构建指定大小的字符串
	var builder strings.Builder
	for i := 0; i < req.Count; i++ {
		// 确保总长度接近目标size
		appendStr := base + strconv.Itoa(i)
		if builder.Len()+len(appendStr) > req.Size {
			break
		}
		builder.WriteString(appendStr)
	}

	data := builder.String()
	return &BigDataResp{
		TotalSize: len(data),
		Data:      data,
	}, nil
}

// 计算密集型测试：执行多次数学运算
func (s *EchoService) MathOperation(req *MathOperationReq) (*MathOperationResp, error) {
	if req.Repeat <= 0 {
		req.Repeat = 1000 // 默认重复1000次模拟计算压力
	}

	result := 0
	for i := 0; i < req.Repeat; i++ {
		switch req.Op {
		case "add":
			result = req.A + req.B + i // 每次计算加i增加随机性
		case "sub":
			result = req.A - req.B - i
		case "mul":
			result = (req.A * req.B) % (i + 1) // 避免溢出
		case "div":
			if req.B == 0 || i == 0 {
				result = 0
			} else {
				result = (req.A / req.B) % i
			}
		default:
			return nil, fmt.Errorf("unsupported operation: %s", req.Op)
		}
	}

	return &MathOperationResp{
		Result: result,
		Steps:  req.Repeat,
	}, nil
}

// 延迟测试：模拟网络延迟或处理耗时
func (s *EchoService) Delay(req *DelayReq) (*DelayResp, error) {
	if req.DelayMs < 0 {
		req.DelayMs = 0
	}
	if req.DelayMs > 5000 { // 限制最大延迟5秒
		req.DelayMs = 5000
	}

	start := time.Now()
	time.Sleep(time.Duration(req.DelayMs) * time.Millisecond)
	actual := int(time.Since(start).Milliseconds())

	return &DelayResp{
		ActualDelayMs: actual,
		Timestamp:     time.Now().UnixMilli(),
	}, nil
}

// 随机数据测试：生成随机字节流，测试二进制数据传输
func (s *EchoService) RandomData(req *RandomDataReq) (*RandomDataResp, error) {
	if req.Size <= 0 {
		req.Size = 1024
	}
	if req.Size > 1024*1024 { // 限制最大1MB
		req.Size = 1024 * 1024
	}

	data := make([]byte, req.Size)
	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}

	return &RandomDataResp{
		Size: req.Size,
		Data: data,
	}, nil
}

// 字符串处理测试：多步骤字符串操作，测试CPU处理能力
func (s *EchoService) StringProcess(req *StringProcessReq) (*StringProcessResp, error) {
	result := req.Input

	// 执行字符串操作
	if req.Reversed {
		runes := []rune(result)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	}

	if req.Upper {
		result = strings.ToUpper(result)
	}

	// 重复多次放大数据量
	if req.Repeat > 1 {
		result = strings.Repeat(result, req.Repeat)
	}

	return &StringProcessResp{
		Processed: result,
		Length:    len(result),
	}, nil
}
