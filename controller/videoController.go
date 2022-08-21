package controller

import (
	"context"
	"github.com/genleel/DoDo/proto/videoService"
	util "github.com/genleel/DoDo/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// Feed 视频流
func Feed(c *gin.Context) {
	latestTime := c.Query("latest_time")
	log.Printf("传入的时间: " + latestTime)
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)

	// 初始化客户端
	microService := util.InitMicro()
	microClient := videoService.NewVideoService("videoService", microService.Client())
	rsp, err := microClient.Feed(context.TODO(), &videoService.FeedReq{
		LatestTime: latestTime,
		UserId:     userId,
	})

	if err != nil {
		log.Printf("用户注册找不到远程服务:", err)
		return
	}
	c.JSON(http.StatusOK, rsp)

}
