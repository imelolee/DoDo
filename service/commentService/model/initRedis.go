package model

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

var RdbVCid *redis.Client //redis db11 -- video_id + comment_id
var RdbCVid *redis.Client //redis db12 -- comment_id + video_id

// InitRedis 初始化Redis连接。
func InitRedis() {
	RdbVCid = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   11, // lsy 选择将video_id中的评论id s存入 DB11.
	})

	RdbCVid = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   12, // lsy 选择将comment_id对应video_id存入 DB12.
	})
}
