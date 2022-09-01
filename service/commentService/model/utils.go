package model

import (
	"commentService/config"
	"fmt"
	"github.com/gogf/gf/util/gconv"
	"log"
	"sync"
	userModel "userService/model"
)

// 在redis中存储video_id对应的comment_id
func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := RdbVCid.SAdd(Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-有err，则直接删除key
		log.Println("redis save send: vId - cId failed, key deleted")
		RdbVCid.Del(Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = RdbCVid.Set(Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Println("Redis save cId - vId failed.")
	}
}

// 此函数用于给一个评论赋值：评论信息+用户信息 填充
func oneComment(comment *CommentInfo, com *Comment, userId int64) {
	var wg sync.WaitGroup
	wg.Add(1)
	//根据评论用户id和当前用户id，查询评论用户信息

	user, err := userModel.GetFeedUserByIdWithCurId(com.VideoId, com.UserId)
	if err != nil {
		fmt.Println("userModel.GetFeedUserByIdWithCurId err:", err)
		return
	}
	var userInfo userModel.FeedUser
	err = gconv.Struct(user, &userInfo)
	if err != nil {
		log.Printf("类型转换失败:", err)
	}

	comment.Id = com.Id
	comment.Content = com.CommentText
	comment.CreateDate = com.CreateDate.Format(config.DateTime)
	comment.UserInfo = userInfo
	if err != nil {
		fmt.Println("CommentService-GetList: GetUserByIdWithCurId return err:", err) //函数返回提示错误信息
	}
	wg.Done()
	wg.Wait()
}
