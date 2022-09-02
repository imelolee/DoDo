package model

import (
	log "go-micro.dev/v4/logger"
	"strconv"
	"time"
	"userService/config"
)

// Follow 用户关系结构，对应用户关系表。
type Follow struct {
	Id         int64
	UserId     int64
	FollowerId int64
	Cancel     int8
}

// FindRelation 给定当前用户和目标用户id，查询follow表中相应的记录。
func FindRelation(userId int64, targetId int64) (*Follow, error) {
	// follow变量用于后续存储数据库查出来的用户关系。
	follow := Follow{}
	//当查询出现错误时，日志打印err msg，并return err.
	if err := Db.
		Where("user_id = ?", targetId).
		Where("follower_id = ?", userId).
		Where("cancel = ?", 0).
		Take(&follow).Error; nil != err {
		// 当没查到数据时，gorm也会报错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		log.Infof(err.Error())
		return nil, err
	}
	//正常情况，返回取到的值和空err.
	return &follow, nil
}

// GetFollowersIds 给定用户id，查询他关注了哪些人的id。
func GetFollowersIds(userId int64) ([]int64, error) {
	var ids []int64
	if err := Db.Model(Follow{}).
		Where("user_id = ?", userId).
		Where("cancel = ?", 0).
		Pluck("follower_id", &ids).Error; nil != err {
		// 没有粉丝，但是不能算错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		// 查询出错。
		log.Infof(err.Error())
		return nil, err
	}
	// 查询成功。
	return ids, nil
}

// GetFollowingIds 给定用户id，查询他关注了哪些人的id。
func GetFollowingIds(userId int64) ([]int64, error) {
	var ids []int64
	if err := Db.Model(Follow{}).
		Where("follower_id = ?", userId).
		Pluck("user_id", &ids).Error; nil != err {
		// 没有关注任何人，但是不能算错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		// 查询出错。
		log.Infof(err.Error())
		return nil, err
	}
	// 查询成功。
	return ids, nil
}

func IsFollowing(userId int64, targetId int64) (bool, error) {
	// 先查Redis里面是否有此关系。
	if flag, _ := RdbFollowingPart.SIsMember(Ctx, strconv.Itoa(int(userId)), targetId).Result(); flag {
		// 重现设置过期时间。
		RdbFollowingPart.Expire(Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
		return true, nil
	}
	// SQL 查询。
	relation, err := FindRelation(userId, targetId)

	if nil != err {
		return false, err
	}
	if nil == relation {
		return false, nil
	}
	// 存在此关系，将其注入Redis中。
	go addRelationToRedis(int(userId), int(targetId))
	return true, nil
}

// Redis中添加用户关注关系
func addRelationToRedis(userId int, targetId int) {
	// 第一次存入时，给该key添加一个-1为key，防止脏数据的写入。当然set可以去重，直接加，便于CPU。
	RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(userId)), -1)
	// 将查询到的关注关系注入Redis.
	RdbFollowingPart.SAdd(Ctx, strconv.Itoa(int(userId)), targetId)
	// 更新过期时间。
	RdbFollowingPart.Expire(Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
}

func GetFollowerCnt(userId int64) (int64, error) {
	// 查Redis中是否已经存在。
	if cnt, _ := RdbFollowers.SCard(Ctx, strconv.Itoa(int(userId))).Result(); cnt > 0 {
		// 更新过期时间。
		RdbFollowers.Expire(Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
		return cnt - 1, nil
	}
	// SQL中查询。
	ids, err := GetFollowersIds(userId)
	if nil != err {
		return 0, err
	}
	// 将数据存入Redis.
	// 更新followers 和 followingPart
	go addFollowersToRedis(int(userId), ids)

	return int64(len(ids)), nil
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

func GetFollowingCnt(userId int64) (int64, error) {
	// 查看Redis中是否有关注数。
	if RdbFollowing == nil {
		InitRedis()
	}
	if cnt, _ := RdbFollowing.SCard(Ctx, strconv.Itoa(int(userId))).Result(); cnt > 0 {
		// 更新过期时间。
		RdbFollowing.Expire(Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
		return cnt - 1, nil

	}
	// 用SQL查询。
	ids, err := GetFollowingIds(userId)

	if nil != err {
		return 0, err
	}
	// 更新Redis中的followers和followPart
	go addFollowingToRedis(int(userId), ids)

	return int64(len(ids)), nil
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
