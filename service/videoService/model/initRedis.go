package model

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

var RdbLikeUserId *redis.Client  //key:userId,value:VideoId
var RdbLikeVideoId *redis.Client //key:VideoId,value:userId
var RdbFollowers *redis.Client
var RdbFollowing *redis.Client
var RdbFollowingPart *redis.Client
var RdbVCid *redis.Client //redis db11 -- video_id + comment_id
var RdbCVid *redis.Client //redis db12 -- comment_id + video_id

// InitRedis 初始化Redis连接。
func InitRedis() {

	RdbLikeUserId = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   5, //  选择将点赞视频id信息存入 DB5.
	})

	RdbLikeVideoId = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   6, //  选择将点赞用户id信息存入 DB6.
	})

	RdbFollowers = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0, // 粉丝列表信息存入 DB0.
	})
	RdbFollowing = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   1, // 关注列表信息信息存入 DB1.
	})
	RdbFollowingPart = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   3, // 当前用户是否关注了自己粉丝信息存入 DB1.
	})

	RdbVCid = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   11, // lsy 选择将video_id中的评论id s存入 DB11.
	})

	RdbCVid = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   12, // lsy 选择将comment_id对应video_id存入 DB12.
	})

}
