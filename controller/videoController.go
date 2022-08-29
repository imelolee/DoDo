package controller

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/genleel/DoDo/model"
	"github.com/genleel/DoDo/proto/videoService"
	"github.com/genleel/DoDo/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []model.FeedVideo `json:"video_list"`
	NextTime  int64             `json:"next_time"`
}

type VideoListResponse struct {
	Response
	VideoList []model.FeedVideo `json:"video_list"`
}

// Feed /feed/
func Feed(c *gin.Context) {
	inputTime := c.Query("latest_time")
	fmt.Println("传入的时间:" + inputTime)
	var lastTime time.Time
	if inputTime != "0" {
		me, _ := strconv.ParseInt(inputTime, 10, 64)
		// 毫秒时间戳转换
		lastTime = time.Unix(me/1000, 0)
	} else {
		lastTime = time.Now()
	}
	fmt.Printf("获取到时间戳:%v", lastTime)
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	fmt.Printf("获取到用户id:%v\n", userId)

	videoMicro := utils.InitMicro()
	videoClient := videoService.NewVideoService("videoService", videoMicro.Client())

	feedRsp, err := videoClient.Feed(context.TODO(), &videoService.FeedReq{
		LatestTime: lastTime.Unix(),
		UserId:     userId,
	})

	if err != nil {
		fmt.Printf("videoService.Feed err：%v", err)
		c.JSON(http.StatusOK, FeedResponse{
			Response: Response{StatusCode: 1, StatusMsg: "获取视频流失败"},
		})
		return
	}

	var tmpList []model.FeedVideo
	gconv.Struct(feedRsp.VideoList, &tmpList)

	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0},
		VideoList: tmpList,
		NextTime:  feedRsp.NextTime,
	})
}

// Publish /publish/action/
func Publish(c *gin.Context) {
	data, err := c.FormFile("data")
	userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
	fmt.Printf("获取到用户id:%v\n", userId)
	title := c.PostForm("title")
	fmt.Printf("获取到视频title:%v\n", title)
	if err != nil {
		fmt.Printf("获取视频流失败:%v", err)
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	dataFile, _ := data.Open()
	byteData, err := ioutil.ReadAll(dataFile)

	videoMicro := utils.InitMicro()
	videoClient := videoService.NewVideoService("videoService", videoMicro.Client())
	_, err = videoClient.Publish(context.TODO(), &videoService.PublishReq{
		Data:     byteData,
		UserId:   userId,
		Title:    title,
		FileSize: data.Size,
		FileExt:  path.Ext(data.Filename),
	})
	if err != nil {
		fmt.Printf("videoService.Publish err：%v", err)
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  "Upload success.",
	})
}

// PublishList /publish/list/
func PublishList(c *gin.Context) {
	user_Id := c.Query("user_id")
	userId, _ := strconv.ParseInt(user_Id, 10, 64)
	fmt.Printf("获取到用户id:%v\n", userId)
	curUser, _ := c.Get("userId")
	cur_id := curUser.(*jwt.StandardClaims).Id
	curId, _ := strconv.ParseInt(cur_id, 10, 64)
	fmt.Printf("获取到当前用户id:%v\n", curId)
	videoMicro := utils.InitMicro()
	videoClient := videoService.NewVideoService("videoService", videoMicro.Client())
	listRsp, err := videoClient.GetPublishList(context.TODO(), &videoService.PublishListReq{
		UserId: userId,
		CurId:  curId,
	})
	if err != nil {
		fmt.Printf("videoService.GetPublishList err：%v\n", err)
		c.JSON(http.StatusOK, VideoListResponse{
			Response: Response{StatusCode: 1, StatusMsg: "获取视频列表失败"},
		})
		return
	}
	var tmpList []model.FeedVideo
	gconv.Struct(listRsp.Video, &tmpList)
	c.JSON(http.StatusOK, VideoListResponse{
		Response:  Response{StatusCode: 0},
		VideoList: tmpList,
	})
}
