package handler

import (
	"commentService/config"
	"commentService/model"
	"context"
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/gogf/gf/util/gconv"
	"go-micro.dev/v4"
	"log"
	"sync"
	userModel "userService/model"
	userService "userService/proto"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	consulReg := consul.NewRegistry()
	return micro.NewService(micro.Registry(consulReg))
}

// 在redis中存储video_id对应的comment_id
func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := model.RdbVCid.SAdd(model.Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-有err，则直接删除key
		log.Println("redis save send: vId - cId failed, key deleted")
		model.RdbVCid.Del(model.Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = model.RdbCVid.Set(model.Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Println("redis save one cId - vId failed")
	}
}

// 此函数用于给一个评论赋值：评论信息+用户信息 填充
func oneComment(comment *model.CommentInfo, com *model.Comment, userId int64) {
	var wg sync.WaitGroup
	wg.Add(1)
	//根据评论用户id和当前用户id，查询评论用户信息
	userMicro := InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())

	userData, err := userClient.GetFeedUserByIdWithCurId(context.TODO(), &userService.CurIdReq{
		Id:    com.VideoId,
		CurId: com.UserId,
	})

	var userInfo userModel.FeedUser
	err = gconv.Struct(userData.User, &userInfo)
	if err != nil {
		log.Printf("类型转换失败:", err)
	}

	comment.Id = com.Id
	comment.Content = com.CommentText
	comment.CreateDate = com.CreateDate.Format(config.DateTime)
	comment.UserInfo = userInfo
	if err != nil {
		log.Println("CommentService-GetList: GetUserByIdWithCurId return err: " + err.Error()) //函数返回提示错误信息
	}
	wg.Done()
	wg.Wait()
}
