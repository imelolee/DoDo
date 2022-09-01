package model

import (
	"followService/config"
	"strconv"
	"time"
)

// Redis中添加用户关注关系
func addRelationToRedis(userId int, targetId int) {
	// 第一次存入时，给该key添加一个-1为key，防止脏数据的写入。当然set可以去重，直接加，便于CPU。
	RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(userId)), -1)
	// 将查询到的关注关系注入Redis.
	RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(userId)), targetId)
	// 更新过期时间。
	RdbFollowingPart.Expire(Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
}

// 在redis中加入关注者信息
func addFollowersToRedis(userId int, ids []int64) {
	RdbFollowers.SAdd(Ctx, strconv.Itoa(userId), -1)
	for i, id := range ids {
		RdbFollowers.SAdd(Ctx, strconv.Itoa(userId), id)
		RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(id)), userId)
		RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(id)), -1)
		// 更新部分关注者的时间
		RdbFollowingPart.Expire(Ctx, strconv.Itoa(int(id)),
			config.ExpireTime+time.Duration((i%10)<<8))
	}
	// 更新followers的过期时间。
	RdbFollowers.Expire(Ctx, strconv.Itoa(userId), config.ExpireTime)
}

// 在redis中加入关注信息
func addFollowingToRedis(userId int, ids []int64) {
	RdbFollowing.SAdd(Ctx, strconv.Itoa(userId), -1)
	for i, id := range ids {
		RdbFollowing.SAdd(Ctx, strconv.Itoa(userId), id)
		RdbFollowingPart.SAdd(Ctx, strconv.Itoa(userId), id)
		RdbFollowingPart.SAdd(Ctx, strconv.Itoa(userId), -1)
		// 更新过期时间
		RdbFollowingPart.Expire(Ctx, strconv.Itoa(userId),
			config.ExpireTime+time.Duration((i%10)<<8))
	}
	// 更新following的过期时间
	RdbFollowing.Expire(Ctx, strconv.Itoa(userId), config.ExpireTime)
}

// 添加关注时设置Redis
func updateRedisWithAdd(userId int64, targetId int64) {
	targetIdStr := strconv.Itoa(int(targetId))
	if cnt, _ := RdbFollowers.SCard(Ctx, targetIdStr).Result(); 0 != cnt {
		RdbFollowers.SAdd(Ctx, targetIdStr, userId)
		RdbFollowers.Expire(Ctx, targetIdStr, config.ExpireTime)
	}

	followingUserIdStr := strconv.Itoa(int(userId))
	if cnt, _ := RdbFollowing.SCard(Ctx, followingUserIdStr).Result(); 0 != cnt {
		RdbFollowing.SAdd(Ctx, followingUserIdStr, targetId)
		RdbFollowing.Expire(Ctx, followingUserIdStr, config.ExpireTime)
	}

	followingPartUserIdStr := followingUserIdStr
	RdbFollowingPart.SAdd(Ctx, followingPartUserIdStr, targetId)
	// 可能是第一次给改用户加followingPart的关注者，需要加上-1防止脏读。
	RdbFollowingPart.SAdd(Ctx, followingPartUserIdStr, -1)
	RdbFollowingPart.Expire(Ctx, followingPartUserIdStr, config.ExpireTime)

}

// 当取关时，更新redis里的信息
func updateRedisWithDelete(userId int64, targetId int64) {
	/*
		1-Redis是否存在followers_targetId.
		2-Redis是否存在following_userId.
		2-Redis是否存在following_part_userId.
	*/
	// step1
	targetIdStr := strconv.Itoa(int(targetId))
	if cnt, _ := RdbFollowers.SCard(Ctx, targetIdStr).Result(); 0 != cnt {
		RdbFollowers.SRem(Ctx, targetIdStr, userId)
		RdbFollowers.Expire(Ctx, targetIdStr, config.ExpireTime)
	}
	// step2
	followingIdStr := strconv.Itoa(int(userId))
	if cnt, _ := RdbFollowing.SCard(Ctx, followingIdStr).Result(); 0 != cnt {
		RdbFollowing.SRem(Ctx, followingIdStr, targetId)
		RdbFollowing.Expire(Ctx, followingIdStr, config.ExpireTime)
	}
	// step3
	followingPartUserIdStr := followingIdStr
	if cnt, _ := RdbFollowingPart.Exists(Ctx, followingPartUserIdStr).Result(); 0 != cnt {
		RdbFollowingPart.SRem(Ctx, followingPartUserIdStr, targetId)
		RdbFollowingPart.Expire(Ctx, followingPartUserIdStr, config.ExpireTime)
	}
}
