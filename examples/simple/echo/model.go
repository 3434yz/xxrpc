package echo

import "time"

// 基础请求/响应结构
type SayHelloReq struct {
	Message string
}

type SayHelloResp struct {
	Message string
}

type ComplexHelloReq struct {
	Message   string
	ID        int64
	Timestamp time.Time
	Metadata  map[string]string
	Tags      []string
	Nested    struct {
		Name  string
		Score float64
	}
	Data       []byte
	Attributes [5]int
	Enabled    bool
	Options    *struct {
		Retry   int
		Timeout time.Duration
	}
}

type ComplexHelloResp struct {
	Status  string
	ReplyID string
}

// 大数据量测试结构
type BigDataReq struct {
	Size  int    // 要求返回的字符串大小(字节)
	Count int    // 重复次数
	Data  string // 基础数据
}

type BigDataResp struct {
	TotalSize int    // 总数据大小
	Data      string // 生成的大数据
}

// 计算型测试结构
type MathOperationReq struct {
	A      int
	B      int
	Op     string // 操作类型: add, sub, mul, div
	Repeat int    // 重复计算次数(模拟复杂计算)
}

type MathOperationResp struct {
	Result int
	Steps  int // 实际计算步数
}

// 延迟测试结构
type DelayReq struct {
	DelayMs int // 延迟毫秒数
}

type DelayResp struct {
	ActualDelayMs int
	Timestamp     int64 // 服务器处理完成时间戳
}

// 随机数据测试结构
type RandomDataReq struct {
	Size int // 随机字节数
}

type RandomDataResp struct {
	Size int
	Data []byte
}

// 字符串处理测试结构
type StringProcessReq struct {
	Input    string
	Reversed bool // 是否反转
	Upper    bool // 是否转为大写
	Repeat   int  // 重复次数
}

type StringProcessResp struct {
	Processed string
	Length    int
}
