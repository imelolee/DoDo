package model

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var RdbFollowers *redis.Client
var RdbFollowing *redis.Client
var RdbFollowingPart *redis.Client

// InitRedis 初始化Redis连接。
func InitRedis() {
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
}
