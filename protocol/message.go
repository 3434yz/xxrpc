package protocol

type Request struct {
	Method string // e.g., "UserService.GetUser"
	Params []byte // 参数的序列化数据
}

func (r *Request) Reset() {
	r.Method = ""
	if r.Params != nil {
		r.Params = r.Params[:0] // 清空切片但不释放内存
	}
}

type Response struct {
	Data  []byte // 序列化返回值
	Error string // 错误信息
}

func (r *Response) Reset() {
	r.Error = ""
	if r.Data != nil {
		r.Data = r.Data[:0] // 清空切片但不释放内存
	}
}
