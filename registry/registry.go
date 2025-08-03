package registry

import (
	"fmt"
	"reflect"
)

type serviceMethod struct {
	method   reflect.Method
	receiver reflect.Value
}

type Registry struct {
	services map[string]map[string]*serviceMethod
}

func NewRegister() *Registry {
	return &Registry{
		services: make(map[string]map[string]*serviceMethod),
	}
}

// Register 注册一个服务
func (r *Registry) Register(serviceName string, svc any) error {
	svcType := reflect.TypeOf(svc)
	svcValue := reflect.ValueOf(svc)

	methods := make(map[string]*serviceMethod)
	for i := 0; i < svcType.NumMethod(); i++ {
		m := svcType.Method(i)
		methods[m.Name] = &serviceMethod{
			method:   m,
			receiver: svcValue,
		}
	}
	r.services[serviceName] = methods
	return nil
}

// Find 找到某一个服务的方法
func (r *Registry) Find(serviceName, methodName string) (reflect.Value, reflect.Method, error) {
	methods, ok := r.services[serviceName]
	if !ok {
		return reflect.Value{}, reflect.Method{}, fmt.Errorf("service %s not found", serviceName)
	}
	method, ok := methods[methodName]
	if !ok {
		return reflect.Value{}, reflect.Method{}, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
	}
	return method.receiver, method.method, nil
}
