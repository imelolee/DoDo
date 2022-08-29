package controller

import (
	"context"
	"fmt"
	"github.com/genleel/DoDo/model"
	"github.com/genleel/DoDo/proto/followService"
	"github.com/genleel/DoDo/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
	"net/http"
	"strconv"
)

// RelationActionResp 关注和取消关注需要返回结构。
type RelationActionResp struct {
	Response
}

// FollowingResp 获取关注列表需要返回的结构。
type FollowingResp struct {
	Response
	UserList []model.FeedUser `json:"user_list,omitempty"`
}

// FollowersResp 获取粉丝列表需要返回的结构。
type FollowersResp struct {
	Response
	// 必须大写，才能序列化
	UserList []model.FeedUser `json:"user_list,omitempty"`
}

// RelationAction 处理关注和取消关注请求。
func RelationAction(c *gin.Context) {
	userId, err := strconv.ParseInt(c.GetString("userId"), 10, 64)
	toUserId, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 64)
	// fmt.Println(userId, toUserId, actionType)
	// 传入参数格式有问题。
	if nil != err || nil != err || nil != err || actionType < 1 || actionType > 2 {
		fmt.Printf("fail")
		c.JSON(http.StatusOK, RelationActionResp{
			Response{
				StatusCode: -1,
				StatusMsg:  "用户id格式错误",
			},
		})
		return
	}

	followMicro := utils.InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())

	switch {
	// 关注
	case 1 == actionType:
		go followClient.AddFollowRelation(context.TODO(), &followService.UserTargetReq{
			UserId:   userId,
			TargetId: toUserId,
		})
	// 取关
	case 2 == actionType:
		go followClient.DeleteFollowRelation(context.TODO(), &followService.UserTargetReq{
			UserId:   userId,
			TargetId: toUserId,
		})
	}
	fmt.Println("关注、取关成功。")
	c.JSON(http.StatusOK, RelationActionResp{
		Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
	})
}

// GetFollowing 处理获取关注列表请求。
func GetFollowing(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	// 用户id解析出错。
	if err != nil {
		c.JSON(http.StatusOK, FollowingResp{
			Response: Response{
				StatusCode: -1,
				StatusMsg:  "用户id格式错误。",
			},
			UserList: nil,
		})
		return
	}
	// 正常获取关注列表
	followMicro := utils.InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())
	followingRsp, err := followClient.GetFollowing(context.TODO(), &followService.UserIdReq{
		UserId: userId,
	})
	// 获取关注列表时出错。
	if err != nil {
		c.JSON(http.StatusOK, FollowingResp{
			Response: Response{
				StatusCode: -1,
				StatusMsg:  "获取关注列表时出错.",
			},
			UserList: nil,
		})
		return
	}
	// 成功获取到关注列表。
	fmt.Println("获取关注列表成功.")
	var tmpList []model.FeedUser
	gconv.Struct(followingRsp.User, &tmpList)
	c.JSON(http.StatusOK, FollowingResp{
		UserList: tmpList,
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
	})
}

// GetFollowers 处理获取关注列表请求
func GetFollowers(c *gin.Context) {
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	// 用户id解析出错。
	if nil != err {
		c.JSON(http.StatusOK, FollowersResp{
			Response: Response{
				StatusCode: -1,
				StatusMsg:  "用户id格式错误。",
			},
			UserList: nil,
		})
		return
	}
	// 正常获取粉丝列表
	followMicro := utils.InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())
	followerRsp, err := followClient.GetFollowers(context.TODO(), &followService.UserIdReq{
		UserId: userId,
	})
	// 获取关注列表时出错。
	if err != nil {
		c.JSON(http.StatusOK, FollowersResp{
			Response: Response{
				StatusCode: -1,
				StatusMsg:  "获取粉丝列表时出错。",
			},
			UserList: nil,
		})
		return
	}
	// 成功获取到粉丝列表。
	var tmpList []model.FeedUser
	gconv.Struct(followerRsp.User, &tmpList)
	c.JSON(http.StatusOK, FollowersResp{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "OK",
		},
		UserList: tmpList,
	})
}
