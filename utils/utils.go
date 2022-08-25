package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	// 初始化客户端
	consulReg := consul.NewRegistry()
	return micro.NewService(micro.Registry(consulReg))

}

// Encoder 密码加密
func Encoder(password string) string {
	h := hmac.New(sha256.New, []byte(password))
	sha := hex.EncodeToString(h.Sum(nil))
	fmt.Println("Result: " + sha)
	return sha
}
