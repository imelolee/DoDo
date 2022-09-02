package controller

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/genleel/DoDo/config"
	"github.com/genleel/DoDo/model"
	"github.com/genleel/DoDo/proto/userService"
	"github.com/genleel/DoDo/utils"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type UserfmtinResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	User model.FeedUser `json:"user"`
}

// Register POST douyin/user/register/ 用户注册
func Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	fmt.Println("注册用户：", username)

	userMicro := utils.InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())

	userRsp, err := userClient.GetTableUserByUsername(context.TODO(), &userService.UsernameReq{
		Name: username,
	})
	if err != nil {
		fmt.Println("userClient.GetTableUserByUsername err:", err)
	}

	if username == userRsp.User.Name {
		c.JSON(http.StatusOK, UserfmtinResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User already exist"},
		})
	} else {
		newUser := model.User{
			Name:     username,
			Password: utils.Encoder(password),
		}

		var tmpUser *userService.User
		gconv.Struct(newUser, &tmpUser)

		insertRsp, err := userClient.InsertTableUser(context.TODO(), &userService.UserReq{
			User: tmpUser,
		})
		if insertRsp.Flag != true || err != nil {
			fmt.Println("Insert User Failed.")
		}
		userRsp, err := userClient.GetTableUserByUsername(context.TODO(), &userService.UsernameReq{
			Name: username,
		})
		if err != nil {
			fmt.Println("userClient.GetTableUserByUsername err:", err)
		}

		token := GenerateToken(userRsp.User)
		fmt.Println("注册的用户id: ", userRsp.User.Id)
		c.JSON(http.StatusOK, UserfmtinResponse{
			Response: Response{StatusCode: 0},
			UserId:   userRsp.User.Id,
			Token:    token,
		})
	}
}

// GenerateToken 根据username生成一个token
func GenerateToken(user *userService.User) string {
	fmt.Printf("Generate token: %v\n", user)
	expiredTime := time.Now().Unix() + int64(config.OneDayOfHours)
	fmt.Printf("Expired time: %v\n", expiredTime)
	id64 := user.Id
	fmt.Printf("id: %v\n", strconv.FormatInt(id64, 10))
	claims := jwt.StandardClaims{
		Audience:  user.Name,
		ExpiresAt: expiredTime,
		Id:        strconv.FormatInt(id64, 10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "tiktok",
		NotBefore: time.Now().Unix(),
		Subject:   "token",
	}
	var jwtSecret = []byte(config.Secret)
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	if err == nil {
		token = "Bearer " + token
		fmt.Println("Generate token success!\n")
		return token
	} else {
		fmt.Println("Generate token failed.\n")
		return "fail"
	}

	return token
}

// fmtin POST douyin/user/fmtin/ 用户登录
func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	fmt.Println("登录用户：", username)
	encodePwd := utils.Encoder(password)
	userMicro := utils.InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())
	userRsp, err := userClient.GetTableUserByUsername(context.TODO(), &userService.UsernameReq{
		Name: username,
	})
	if err != nil {
		fmt.Println("userClient.GetTableUserByUsername err:", err)
	}

	if encodePwd == userRsp.User.Password {
		token := GenerateToken(userRsp.User)
		c.JSON(http.StatusOK, UserfmtinResponse{
			Response: Response{StatusCode: 0},
			UserId:   userRsp.User.Id,
			Token:    token,
		})
	} else {
		c.JSON(http.StatusOK, UserfmtinResponse{
			Response: Response{StatusCode: 1, StatusMsg: "Username or Password Error"},
		})
	}
}

// UserInfo GET douyin/user/ 用户信息
func UserInfo(c *gin.Context) {
	user_id := c.Query("user_id")
	id, _ := strconv.ParseInt(user_id, 10, 64)
	userMicro := utils.InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())

	userRsp, err := userClient.GetFeedUserById(context.TODO(), &userService.IdReq{
		Id: id,
	})

	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User Doesn't Exist"},
		})
	} else {
		var tmpUser model.FeedUser
		gconv.Struct(userRsp.User, &tmpUser)
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0},
			User:     tmpUser,
		})
	}
}
