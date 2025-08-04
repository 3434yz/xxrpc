package protocol

type Request struct {
	Service string // e.g., "UserService"
	Method  string // e.g., "GetUser"
	Params  []byte // 参数的序列化数据
}

func (r *Request) Reset() {
	r.Service = ""
	r.Method = ""
	r.Params = nil
}

type Response struct {
	Data  []byte // 序列化返回值
	Error string // 错误信息
}

func (r *Response) Reset() {
	r.Data = nil
	r.Error = ""
}
