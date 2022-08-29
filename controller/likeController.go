package controller

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/genleel/DoDo/model"
	"github.com/genleel/DoDo/proto/likeService"
	"github.com/genleel/DoDo/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
	"net/http"
	"strconv"
)

type likeResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

type GetFavouriteListResponse struct {
	StatusCode int32             `json:"status_code"`
	StatusMsg  string            `json:"status_msg,omitempty"`
	VideoList  []model.FeedVideo `json:"video_list,omitempty"`
}

// FavoriteAction 点赞或者取消赞操作;
func FavoriteAction(c *gin.Context) {

	user, _ := c.Get("userId")
	strUserId := user.(*jwt.StandardClaims).Id
	userId, _ := strconv.ParseInt(strUserId, 10, 64)
	strVideoId := c.Query("video_id")
	videoId, _ := strconv.ParseInt(strVideoId, 10, 64)
	strActionType := c.Query("action_type")
	actionType, _ := strconv.ParseInt(strActionType, 10, 64)

	likeMicro := utils.InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())
	//获取点赞或者取消赞操作的错误信息
	_, err := likeClient.FavouriteAction(context.TODO(), &likeService.ActionReq{
		UserId:     userId,
		VideoId:    videoId,
		ActionType: actionType,
	})
	if err == nil {
		fmt.Printf("方法like.FavouriteAction(userid, videoId, int32(actiontype) 成功")
		c.JSON(http.StatusOK, likeResponse{
			StatusCode: 0,
			StatusMsg:  "favourite action success",
		})
	} else {
		fmt.Printf("方法like.FavouriteAction(userid, videoId, int32(actiontype) 失败：%v", err)
		c.JSON(http.StatusOK, likeResponse{
			StatusCode: 1,
			StatusMsg:  "favourite action fail",
		})
	}
}

// GetFavouriteList 获取点赞列表;
func GetFavouriteList(c *gin.Context) {
	strUserId := c.Query("user_id")
	strCurId := c.GetString("userId")
	userId, _ := strconv.ParseInt(strUserId, 10, 64)
	curId, _ := strconv.ParseInt(strCurId, 10, 64)
	likeMicro := utils.InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())
	videoRsp, err := likeClient.GetFavouriteList(context.TODO(), &likeService.UserCurReq{
		UserId: userId,
		CurId:  curId,
	})

	var tmpList []model.FeedVideo
	gconv.Struct(videoRsp.Video, &tmpList)

	if err == nil {
		fmt.Printf("方法like.GetFavouriteList(userid) 成功")
		c.JSON(http.StatusOK, GetFavouriteListResponse{
			StatusCode: 0,
			StatusMsg:  "get favouriteList success",
			VideoList:  tmpList,
		})
	} else {
		fmt.Printf("方法like.GetFavouriteList(userid) 失败：%v", err)
		c.JSON(http.StatusOK, GetFavouriteListResponse{
			StatusCode: 1,
			StatusMsg:  "get favouriteList fail ",
		})
	}
}
