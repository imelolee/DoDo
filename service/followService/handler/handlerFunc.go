package handler

import (
	"followService/config"
	"followService/model"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"strconv"
	"time"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	consulReg := consul.NewRegistry()
	return micro.NewService(micro.Registry(consulReg))
}

// Redis中添加用户关注关系
func addRelationToRedis(userId int, targetId int) {
	// 第一次存入时，给该key添加一个-1为key，防止脏数据的写入。当然set可以去重，直接加，便于CPU。
	model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(int(userId)), -1)
	// 将查询到的关注关系注入Redis.
	model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(int(userId)), targetId)
	// 更新过期时间。
	model.RdbFollowingPart.Expire(model.Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
}

// 在redis中加入关注者信息
func addFollowersToRedis(userId int, ids []int64) {
	model.RdbFollowers.SAdd(model.Ctx, strconv.Itoa(userId), -1)
	for i, id := range ids {
		model.RdbFollowers.SAdd(model.Ctx, strconv.Itoa(userId), id)
		model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(int(id)), userId)
		model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(int(id)), -1)
		// 更新部分关注者的时间
		model.RdbFollowingPart.Expire(model.Ctx, strconv.Itoa(int(id)),
			config.ExpireTime+time.Duration((i%10)<<8))
	}
	// 更新followers的过期时间。
	model.RdbFollowers.Expire(model.Ctx, strconv.Itoa(userId), config.ExpireTime)
}

// 在redis中加入关注信息
func addFollowingToRedis(userId int, ids []int64) {
	model.RdbFollowing.SAdd(model.Ctx, strconv.Itoa(userId), -1)
	for i, id := range ids {
		model.RdbFollowing.SAdd(model.Ctx, strconv.Itoa(userId), id)
		model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(userId), id)
		model.RdbFollowingPart.SAdd(model.Ctx, strconv.Itoa(userId), -1)
		// 更新过期时间
		model.RdbFollowingPart.Expire(model.Ctx, strconv.Itoa(userId),
			config.ExpireTime+time.Duration((i%10)<<8))
	}
	// 更新following的过期时间
	model.RdbFollowing.Expire(model.Ctx, strconv.Itoa(userId), config.ExpireTime)
}

// 添加关注时设置Redis
func updateRedisWithAdd(userId int64, targetId int64) {
	targetIdStr := strconv.Itoa(int(targetId))
	if cnt, _ := model.RdbFollowers.SCard(model.Ctx, targetIdStr).Result(); 0 != cnt {
		model.RdbFollowers.SAdd(model.Ctx, targetIdStr, userId)
		model.RdbFollowers.Expire(model.Ctx, targetIdStr, config.ExpireTime)
	}

	followingUserIdStr := strconv.Itoa(int(userId))
	if cnt, _ := model.RdbFollowing.SCard(model.Ctx, followingUserIdStr).Result(); 0 != cnt {
		model.RdbFollowing.SAdd(model.Ctx, followingUserIdStr, targetId)
		model.RdbFollowing.Expire(model.Ctx, followingUserIdStr, config.ExpireTime)
	}

	followingPartUserIdStr := followingUserIdStr
	model.RdbFollowingPart.SAdd(model.Ctx, followingPartUserIdStr, targetId)
	// 可能是第一次给改用户加followingPart的关注者，需要加上-1防止脏读。
	model.RdbFollowingPart.SAdd(model.Ctx, followingPartUserIdStr, -1)
	model.RdbFollowingPart.Expire(model.Ctx, followingPartUserIdStr, config.ExpireTime)

}

// 当取关时，更新redis里的信息
func updateRedisWithDel(userId int64, targetId int64) {
	/*
		1-Redis是否存在followers_targetId.
		2-Redis是否存在following_userId.
		2-Redis是否存在following_part_userId.
	*/
	// step1
	targetIdStr := strconv.Itoa(int(targetId))
	if cnt, _ := model.RdbFollowers.SCard(model.Ctx, targetIdStr).Result(); 0 != cnt {
		model.RdbFollowers.SRem(model.Ctx, targetIdStr, userId)
		model.RdbFollowers.Expire(model.Ctx, targetIdStr, config.ExpireTime)
	}
	// step2
	followingIdStr := strconv.Itoa(int(userId))
	if cnt, _ := model.RdbFollowing.SCard(model.Ctx, followingIdStr).Result(); 0 != cnt {
		model.RdbFollowing.SRem(model.Ctx, followingIdStr, targetId)
		model.RdbFollowing.Expire(model.Ctx, followingIdStr, config.ExpireTime)
	}
	// step3
	followingPartUserIdStr := followingIdStr
	if cnt, _ := model.RdbFollowingPart.Exists(model.Ctx, followingPartUserIdStr).Result(); 0 != cnt {
		model.RdbFollowingPart.SRem(model.Ctx, followingPartUserIdStr, targetId)
		model.RdbFollowingPart.Expire(model.Ctx, followingPartUserIdStr, config.ExpireTime)
	}
}
