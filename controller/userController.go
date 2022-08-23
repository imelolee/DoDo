package controller

import (
	"context"
	"github.com/genleel/DoDo/proto/userService"
	"github.com/genleel/DoDo/utils"
	"log"

	"github.com/gin-gonic/gin"
	"net/http"
)

// Login POST douyin/user/login/ 用户登录
func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	// 初始化客户端
	microService := util.InitMicro()
	microClient := userService.NewUserService("userservice", GetMicroClient)

	rsp, err := microClient.Login(context.TODO(), &userService.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Printf("找不到远程服务:", err)
		return
	}
	c.JSON(http.StatusOK, rsp)
}

func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	// 初始化客户端
	microService := util.InitMicro()
	microClient := userService.NewUserService("userservice", microService.Client())
	rsp, err := microClient.Register(context.TODO(), &userService.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Printf("用户注册找不到远程服务:", err)
		return
	}
	c.JSON(http.StatusOK, rsp)

}
