package model

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	"likeService/config"
	pb "likeService/proto"
	"strconv"
	"strings"
	"sync"
	"time"
	videoModel "videoService/model"
)

// 根据videoId,登录用户curId，添加视频对象到点赞列表空间
func addFavouriteVideoList(videoId int64, curId int64, favoriteVideoList *[]*pb.Video, wg *sync.WaitGroup) {

	defer wg.Done()
	//调用videoService接口，GetVideo：根据videoId，当前用户id:curId，返回Video类型对象
	video, err := videoModel.GetVideo(videoId, curId)
	if err != nil {
		log.Infof("videoModel.GetVideo err:", err)
	}
	var tmpVideo *pb.Video
	gconv.Struct(video, &tmpVideo)

	*favoriteVideoList = append(*favoriteVideoList, tmpVideo)
}

// 根据videoId，将该视频点赞数加入对应提前开辟好的空间内
func addVideoLikeCount(videoId int64, videoLikeCountList *[]int64, wg *sync.WaitGroup) {
	defer wg.Done()
	//调用FavouriteCount：根据videoId,获取点赞数

	count, err := FavouriteCount(videoId)
	if err != nil {
		fmt.Println("likeModel.FavouriteCount err:", err)
		return
	}
	*videoLikeCountList = append(*videoLikeCountList, count)
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

// GetLikeInfo 根据userId,videoId查询点赞信息
func GetLikeInfo(userId int64, videoId int64) (Like, error) {
	//创建一条空like结构体，用来存储查询到的信息
	var likeInfo Like
	//根据userid,videoId查询是否有该条信息，如果有，存储在likeInfo,返回查询结果
	err := Db.Model(Like{}).Where(map[string]interface{}{"user_id": userId, "video_id": videoId}).
		First(&likeInfo).Error
	if err != nil {
		//查询数据为0，打印"can't find data"，返回空结构体，这时候就应该要考虑是否插入这条数据了
		if "record not found" == err.Error() {
			log.Infof("can't find data")
			return Like{}, nil
		} else {
			//如果查询数据库失败，返回获取likeInfo信息失败
			log.Infof(err.Error())
			return likeInfo, errors.New("get likeInfo failed")
		}
	}
	return likeInfo, nil
}

// InsertLike 插入点赞数据
func InsertLike(likeData Like) error {
	//创建点赞数据，默认为点赞，cancel为0，返回错误结果
	err := Db.Model(Like{}).Create(&likeData).Error
	//如果有错误结果，返回插入失败
	if err != nil {
		log.Infof(err.Error())
		return errors.New("insert data fail")
	}
	return nil
}

// UpdateLike 根据userId，videoId,actionType点赞或者取消赞
func UpdateLike(userId int64, videoId int64, actionType int32) error {
	//更新当前用户观看视频的点赞状态“cancel”，返回错误结果
	err := Db.Model(Like{}).Where(map[string]interface{}{"user_id": userId, "video_id": videoId}).
		Update("cancel", actionType).Error
	//如果出现错误，返回更新数据库失败
	if err != nil {
		log.Infof(err.Error())
		return errors.New("update data fail")
	}
	//更新操作成功
	return nil
}

// IsFavorite 根据userId,videoId查询点赞状态 这边可以快一点,通过查询两个Redis DB;
func IsFavorite(videoId int64, userId int64) (bool, error) {
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(userId, 10)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(videoId, 10)
	//step1:查询Redis LikeUserId,key：strUserId中是否存在value:videoId,key中存在value 返回true，不存在返回false
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

//FavouriteCount 根据videoId获取对应点赞数量;
func FavouriteCount(id int64) (int64, error) {
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(id, 10)
	//step1 如果key:strVideoId存在 则计算集合中userId个数
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

// FavouriteAction 根据userId，videoId,actionType对视频进行点赞或者取消赞操作;
func FavouriteAction(videoId int64, userId int64, action int64) error {
	//将int64 videoId转换为 string strVideoId
	strUserId := strconv.FormatInt(userId, 10)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(videoId, 10)
	//将要操作数据库likes表的信息打入消息队列RmqLikeAdd或者RmqLikeDel
	//拼接打入信息
	sb := strings.Builder{}
	sb.WriteString(strUserId)
	sb.WriteString(" ")
	sb.WriteString(strVideoId)

	//step1:维护Redis LikeUserId、LikeVideoId;
	//执行点赞操作维护
	if action == config.LikeAction {
		//查询Redis LikeUserId(key:strUserId)是否已经加载过此信息
		if n, err := RdbLikeUserId.Exists(Ctx, strUserId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			}
			//如果加载过此信息key:strUserId，则加入value:videoId
			//如果redis LikeUserId 添加失败，数据库操作成功，会有脏数据，所以只有redis操作成功才执行数据库likes表操作
			if _, err1 := RdbLikeUserId.SAdd(Ctx, strUserId, videoId).Result(); err1 != nil {
				return err
			} else {
				//如果数据库操作失败了，redis是正确数据，客户端显示的是点赞成功，不会影响后续结果
				//只有当该用户取消所有点赞视频的时候redis才会重新加载数据库信息，这时候因为取消赞了必然和数据库信息一致
				//同样这条信息消费成功与否也不重要，因为redis是正确信息,理由如上
				RmqLikeAdd.Publish(sb.String())
			}
		} else {
			//如果不存在，则维护Redis LikeUserId 新建key:strUserId,设置过期时间，加入DefaultRedisValue，
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := RdbLikeUserId.Expire(Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return err
			}
			videoIdList, err := GetLikeVideoIdList(userId)
			//如果有问题，说明查询失败，返回错误信息："get likeVideoIdList failed"
			if err != nil {
				return err
			}
			//遍历videoIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeVideoId := range videoIdList {
				if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, likeVideoId).Result(); err != nil {
					RdbLikeUserId.Del(Ctx, strUserId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, videoId).Result(); err != nil {
				return err
			} else {
				RmqLikeAdd.Publish(sb.String())
			}
		}
		//查询Redis LikeVideoId(key:strVideoId)是否已经加载过此信息
		if n, err := RdbLikeVideoId.Exists(Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //如果加载过此信息key:strVideoId，则加入value:userId
			//如果redis LikeVideoId 添加失败，返回错误信息
			if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, userId).Result(); err != nil {
				return err
			}
		} else { //如果不存在，则维护Redis LikeVideoId 新建key:strVideoId，设置有效期，加入DefaultRedisValue
			//通过videoId查询likes表,返回所有点赞userId，加入key:strVideoId集合中,
			//再加入当前userId,再更新likes表此条数据
			//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
				RdbLikeVideoId.Del(Ctx, strVideoId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := RdbLikeVideoId.Expire(Ctx, strVideoId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				RdbLikeVideoId.Del(Ctx, strVideoId)
			}
			userIdList, err := GetLikeUserIdList(videoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err != nil {
				return err
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, likeUserId).Result(); err != nil {
					RdbLikeVideoId.Del(Ctx, strVideoId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, userId).Result(); err != nil {
				return err
			}
		}

		return nil

	} else { //执行取消赞操作维护
		//查询Redis LikeUserId(key:strUserId)是否已经加载过此信息
		if n, err := RdbLikeUserId.Exists(Ctx, strUserId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //防止出现redis数据不一致情况，当redis删除操作成功，才执行数据库更新操作
			if _, err := RdbLikeUserId.SRem(Ctx, strUserId, videoId).Result(); err != nil {
				return err
			} else {
				//后续数据库的操作，可以在mq里设置若执行数据库更新操作失败，重新消费该信息
				RmqLikeDel.Publish(sb.String())
			}
		} else { //如果不存在，则维护Redis LikeUserId 新建key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库
			// 还没更新完出现脏读，或者数据库操作失败造成的脏读
			//通过userId查询likes表,返回所有点赞videoId，加入key:strUserId集合中,
			//再删除当前videoId,再更新likes表此条数据
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := RdbLikeUserId.Expire(Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return err
			}
			videoIdList, err := GetLikeVideoIdList(userId)
			//如果有问题，说明查询失败，返回错误信息："get likeVideoIdList failed"
			if err != nil {
				return err
			}
			//遍历videoIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql 数据原子性
			for _, likeVideoId := range videoIdList {
				if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, likeVideoId).Result(); err != nil {
					RdbLikeUserId.Del(Ctx, strUserId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := RdbLikeUserId.SRem(Ctx, strUserId, videoId).Result(); err != nil {
				return err
			} else {
				RmqLikeDel.Publish(sb.String())
			}
		}

		//查询Redis LikeVideoId(key:strVideoId)是否已经加载过此信息
		if n, err := RdbLikeVideoId.Exists(Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //如果加载过此信息key:strVideoId，则删除value:userId
			//如果redis LikeVideoId 删除失败，返回错误信息
			if _, err := RdbLikeVideoId.SRem(Ctx, strVideoId, userId).Result(); err != nil {
				return err
			}
		} else { //如果不存在，则维护Redis LikeVideoId 新建key:strVideoId,加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库
			// 还没更新完出现脏读，或者数据库操作失败造成的脏读
			//通过videoId查询likes表,返回所有点赞userId，加入key:strVideoId集合中,
			//再删除当前userId,再更新likes表此条数据
			//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
				RdbLikeVideoId.Del(Ctx, strVideoId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := RdbLikeVideoId.Expire(Ctx, strVideoId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				RdbLikeVideoId.Del(Ctx, strVideoId)
				return err
			}

			userIdList, err := GetLikeUserIdList(videoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err != nil {
				return err
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err := RdbLikeVideoId.SAdd(Ctx, strVideoId, likeUserId).Result(); err != nil {
					RdbLikeVideoId.Del(Ctx, strVideoId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := RdbLikeVideoId.SRem(Ctx, strVideoId, userId).Result(); err != nil {
				return err
			}
		}
		return nil
	}
}

//GetFavouriteList 根据userId，curId(当前用户Id),返回userId的点赞列表;
func GetFavouriteList(userId int64, curId int64) ([]*pb.Video, error) {
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(userId, 10)
	//step1:查询Redis LikeUserId,如果key：strUserId存在,则获取集合中全部videoId
	if n, err := RdbLikeUserId.Exists(Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认nil,返回错误信息
		if err != nil {
			return nil, err
		}
		//获取集合中全部videoId
		videoIdList, err := RdbLikeUserId.SMembers(Ctx, strUserId).Result()
		//如果有问题，说明查询redis失败,返回默认nil,返回错误信息
		if err != nil {
			return nil, err
		}
		//提前开辟点赞列表空间
		favoriteVideoList := new([]*pb.Video)
		//采用协程并发将Video类型对象添加到集合中去
		i := len(videoIdList) - 1 //去掉DefaultRedisValue
		if i == 0 {
			return *favoriteVideoList, nil
		}
		var wg sync.WaitGroup
		wg.Add(i)
		for j := 0; j <= i; j++ {
			//将string videoId转换为 int64 VideoId
			videoId, _ := strconv.ParseInt(videoIdList[j], 10, 64)
			if videoId == config.DefaultRedisValue {
				continue
			}
			go addFavouriteVideoList(videoId, curId, favoriteVideoList, &wg)
		}
		wg.Wait()

		return *favoriteVideoList, nil
	} else { //如果Redis LikeUserId不存在此key,通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
		//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
			RdbLikeUserId.Del(Ctx, strUserId)

			return nil, err
		}
		//给键值设置有效期，类似于gc机制
		_, err := RdbLikeUserId.Expire(Ctx, strUserId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			RdbLikeUserId.Del(Ctx, strUserId)
			return nil, err
		}
		videoIdList, err := GetLikeVideoIdList(userId)
		//如果有问题，说明查询数据库失败，返回nil和错误信息:"get likeVideoIdList failed"
		if err != nil {
			return nil, err
		}
		//提前开辟点赞列表空间
		favoriteVideoList := new([]*pb.Video)
		//采用协程并发将Video类型对象添加到集合中去
		i := len(videoIdList) - 1 //去掉DefaultRedisValue
		if i == 0 {
			return *favoriteVideoList, nil
		}
		var wg sync.WaitGroup
		wg.Add(i)
		for j := 0; j <= i; j++ {
			if videoIdList[j] == config.DefaultRedisValue {
				continue
			}
			go addFavouriteVideoList(videoIdList[j], curId, favoriteVideoList, &wg)
		}
		wg.Wait()

		return *favoriteVideoList, nil
	}

}

//TotalFavourite 根据userId获取这个用户总共被点赞数量
func TotalFavourite(id int64) (int64, error) {
	//根据userId获取这个用户的发布视频列表信息
	videoList, err := videoModel.GetVideoIdList(id)
	if err != nil {
		return 0, err
	}
	var sum int64 //该用户的总被点赞数
	//提前开辟空间,存取每个视频的点赞数
	videoLikeCountList := new([]int64)
	//采用协程并发将对应videoId的点赞数添加到集合中去
	i := len(videoList)
	var wg sync.WaitGroup
	wg.Add(i)
	for j := 0; j < i; j++ {
		go addVideoLikeCount(videoList[j], videoLikeCountList, &wg)
	}
	wg.Wait()
	//遍历累加，求总被点赞数
	for _, count := range *videoLikeCountList {
		sum += count
	}
	return sum, nil
}

//FavouriteVideoCount 根据userId获取这个用户点赞视频数量
func FavouriteVideoCount(id int64) (int64, error) {
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(id, 10)
	//step1:查询Redis LikeUserId,如果key：strUserId存在,则获取集合中元素个数
	if n, err := RdbLikeUserId.Exists(Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认0,返回错误信息
		if err != nil {
			return 0, err
		} else {
			count, err := RdbLikeUserId.SCard(Ctx, strUserId).Result()
			//如果有问题，说明操作redis失败,返回默认0,返回错误信息
			if err != nil {
				return 0, err
			}
			return count - 1, nil //去掉DefaultRedisValue

		}
	} else { //如果Redis LikeUserId不存在此key,通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
		//再通过set集合中userId个数,获取点赞数量
		//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
			RdbLikeUserId.Del(Ctx, strUserId)
			return 0, err
		}
		//给键值设置有效期，类似于gc机制
		_, err := RdbLikeUserId.Expire(Ctx, strUserId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			RdbLikeUserId.Del(Ctx, strUserId)
			return 0, err
		}
		videoIdList, err := GetLikeVideoIdList(id)
		//如果有问题，说明查询数据库失败，返回默认0,返回错误信息："get likeVideoIdList failed"
		if err != nil {
			return 0, err
		}
		//维护Redis LikeUserId(key:strUserId)，遍历videoIdList加入
		for _, likeVideoId := range videoIdList {
			if _, err := RdbLikeUserId.SAdd(Ctx, strUserId, likeVideoId).Result(); err != nil {
				RdbLikeUserId.Del(Ctx, strUserId)
				return 0, err
			}
		}
		//再通过set集合中videoId个数,获取点赞数量
		count, err := RdbLikeUserId.SCard(Ctx, strUserId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			return 0, err
		}
		return count - 1, nil //去掉DefaultRedisValue
	}
}
