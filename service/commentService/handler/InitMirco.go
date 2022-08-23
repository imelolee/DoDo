package handler

import (
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	// 初始化客户端
	consulReg := consul.NewRegistry()

	return micro.NewService(micro.Registry(consulReg))

}
