package registry

import (
	"fmt"
	"xxrpc/internal/codec"
)

// 调用侧是从Registry中找到对应的服务和方法
// 服务端注册侧是将服务和方法注册到Registry中
// 根据传入的MethodName找到对应的处理函数
// 处理函数的签名是 func([]byte) ([]byte, error)
type Service interface {
	Register(*Registry, codec.Codec) // 注册服务和方法到注册表
	Name() string                    // 返回服务名称
}

type HandlerFunc func([]byte) ([]byte, error)

type ServiceMethod struct {
	Handler HandlerFunc
}

type Registry struct {
	ServiceMethods map[string]*ServiceMethod
}

func NewRegister() *Registry {
	return &Registry{
		ServiceMethods: make(map[string]*ServiceMethod),
	}
}

// Register 注册一个服务
func (r *Registry) Register(svc Service, c codec.Codec) error {
	svc.Register(r, c)
	return nil
}

// Find 找到某一个服务的方法
func (r *Registry) Find(serviceMethodName string) (HandlerFunc, error) {
	serviceMethod, ok := r.ServiceMethods[serviceMethodName]
	if !ok {
		return nil, fmt.Errorf("serviceMethodName %s not found ", serviceMethodName)
	}
	return serviceMethod.Handler, nil
}
