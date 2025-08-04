package protocol

type Request struct {
	Service string // e.g., "UserService"
	Method  string // e.g., "GetUser"
	Params  []byte // 参数的序列化数据
}

func (r *Request) Reset() {
	r.Service = ""
	r.Method = ""
	r.Params = r.Params[:0] // 清空切片但不释放内存
}

type Response struct {
	Data  []byte // 序列化返回值
	Error string // 错误信息
}

func (r *Response) Reset() {
	r.Data = r.Data[:0] // 清空切片但不释放内存
	r.Error = ""
}
