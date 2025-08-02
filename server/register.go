package server

import (
	"fmt"
	"reflect"
)

type serviceMethod struct {
	method reflect.Method
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
func (r *Registry) Register(serviceName string, svc interface{}) error {
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
func (r *Registry) Find(serviceName, methodName string) (*serviceMethod, error) {
	if methods, ok := r.services[serviceName]; ok {
		if method, ok := methods[methodName]; ok {
			return method, nil
		}
	}
	return nil, fmt.Errorf("service %s method %s not found", serviceName, methodName)
}
