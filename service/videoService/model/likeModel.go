package model

import (
	"errors"
	log "go-micro.dev/v4/logger"
	"likeService/config"

	"strconv"
	"time"
)

// Like 表的结构。
type Like struct {
	Id      int64 //自增主键
	UserId  int64 //点赞用户id
	VideoId int64 //视频id
	Cancel  int8  //是否点赞，0为点赞，1为取消赞
}

// GetLikeUserIdList 根据videoId获取点赞userId
func GetLikeUserIdList(videoId int64) ([]int64, error) {
	var likeUserIdList []int64 //存所有该视频点赞用户id；
	//查询likes表对应视频id点赞用户，返回查询结果
	err := Db.Model(Like{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.IsLike}).
		Pluck("user_id", &likeUserIdList).Error
	//查询过程出现错误，返回默认值0，并输出错误信息
	if err != nil {
		log.Infof(err.Error())
		return nil, errors.New("get likeUserIdList failed")
	} else {
		//没查询到或者查询到结果，返回数量以及无报错
		return likeUserIdList, nil
	}
}

//FavouriteCount 根据videoId获取对应点赞数量;
func FavouriteCount(id int64) (int64, error) {
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(id, 10)
	//step1 如果key:strVideoId存在 则计算集合中userId个数
	if RdbLikeVideoId == nil {
		InitRedis()
	}
	if n, err := RdbLikeVideoId.Exists(Ctx, strVideoId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			return 0, err
		}
		//获取集合中userId个数
		count, err := RdbLikeVideoId.SCard(Ctx, strVideoId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			return 0, err
		}
		return count - 1, nil
	} else {
		//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
			RdbLikeVideoId.Del(Ctx, strVideoId)
			return 0, err
		}
		//给键值设置有效期，类似于gc机制
		_, err := RdbLikeVideoId.Expire(Ctx, strVideoId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			RdbLikeVideoId.Del(Ctx, strVideoId)
			return 0, err
		}
		//如果Redis LikeVideoId不存在此key,通过videoId查询likes表,返回所有点赞userId，并维护到Redis LikeVideoId(key:strVideoId)
		//再通过set集合中userId个数,获取点赞数量
		userIdList, err := GetLikeUserIdList(id)
		//如果有问题，说明查询数据库失败，返回默认0,返回错误信息："get likeUserIdList failed"
		if err != nil {
			return 0, err
		}
		//维护Redis LikeVideoId(key:strVideoId)，遍历userIdList加入
		for _, likeUserId := range userIdList {
			RdbLikeVideoId.SAdd(Ctx, strVideoId, likeUserId)
		}
		//再通过set集合中userId个数,获取点赞数量
		count, err := RdbLikeVideoId.SCard(Ctx, strVideoId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			return 0, err
		}

		return count - 1, nil
	}

}

// IsFavorite 根据userId,videoId查询点赞状态 这边可以快一点,通过查询两个Redis DB;
func IsFavorite(videoId int64, userId int64) (bool, error) {
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(userId, 10)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(videoId, 10)
	//step1:查询Redis LikeUserId,key：strUserId中是否存在value:videoId,key中存在value 返回true，不存在返回false
	if RdbLikeUserId == nil {
		InitRedis()
	}
	if n, err := RdbLikeUserId.Exists(Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			return false, err
		}
		exist, err := RdbLikeUserId.SIsMember(Ctx, strUserId, videoId).Result()
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			return false, err
		}

		return exist, nil

	} else { //step2:LikeUserId不存在key,查询Redis LikeVideoId,key中存在value 返回true，不存在返回false
		if n, err := RdbLikeVideoId.Exists(Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回默认false,返回错误信息
			if err != nil {
				return false, err
			}
			exist, err := RdbLikeVideoId.SIsMember(Ctx, strVideoId, userId).Result()
			//如果有问题，说明查询redis失败,返回默认false,返回错误信息
			if err != nil {
				return false, err
			}
			return exist, nil
		} else {
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return false, err
			}
			//给键值设置有效期，类似于gc机制
			_, err := RdbLikeUserId.Expire(Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return false, err
			}
			//step3:LikeUserId LikeVideoId中都没有对应key，通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
			videoIdList, err := GetLikeVideoIdList(userId)
			//如果有问题，说明查询数据库失败，返回默认false,返回错误信息："get likeVideoIdList failed"
			if err != nil {
				return false, err
			}
			//维护Redis LikeUserId(key:strUserId)，遍历videoIdList加入
			for _, likeVideoId := range videoIdList {
				RdbLikeUserId.SAdd(Ctx, strUserId, likeVideoId)
			}
			//查询Redis LikeUserId,key：strUserId中是否存在value:videoId,存在返回true,不存在返回false
			exist, err := RdbLikeUserId.SIsMember(Ctx, strUserId, videoId).Result()
			//如果有问题，说明操作redis失败,返回默认false,返回错误信息
			if err != nil {
				return false, err
			}
			return exist, nil
		}
	}
}

// GetLikeVideoIdList 根据userId查询所属点赞全部videoId
func GetLikeVideoIdList(userId int64) ([]int64, error) {
	var likeVideoIdList []int64
	err := Db.Model(Like{}).Where(map[string]interface{}{"user_id": userId, "cancel": config.IsLike}).
		Pluck("video_id", &likeVideoIdList).Error
	if err != nil {
		//查询数据为0，返回空likeVideoIdList切片，以及返回无错误
		if "record not found" == err.Error() {
			log.Infof("there are no likeVideoId")
			return likeVideoIdList, nil
		} else {
			//如果查询数据库失败，返回获取likeVideoIdList失败
			log.Infof(err.Error())
			return likeVideoIdList, errors.New("get likeVideoIdList failed")
		}
	}
	return likeVideoIdList, nil
}
