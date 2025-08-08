package protocol

type Request struct {
	Method string  // e.g., "UserService.GetUser"
	Params *[]byte // 参数的序列化数据
}

type Response struct {
	Data  *[]byte // 序列化返回值
	Error string  // 错误信息
}
