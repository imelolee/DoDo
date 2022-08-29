package handler

import (
	"context"
	"likeService/config"
	"likeService/model"
	pb "likeService/proto"
	"strconv"
	"strings"
	"sync"
	"time"
	videoService "videoService/proto"

	log "go-micro.dev/v4/logger"
)

type LikeService struct{}

// IsFavorite 根据userId,videoId查询点赞状态 这边可以快一点,通过查询两个Redis DB;
func (e *LikeService) IsFavorite(ctx context.Context, req *pb.VideoUserReq, rsp *pb.BoolRsp) error {
	log.Infof("Received LikeService.IsFavorite request: %v", req)
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(req.UserId, 10)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(req.VideoId, 10)
	//step1:查询Redis LikeUserId,key：strUserId中是否存在value:videoId,key中存在value 返回true，不存在返回false
	if n, err := model.RdbLikeUserId.Exists(model.Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			rsp.Flag = false
			return err
		}
		exist, err := model.RdbLikeUserId.SIsMember(model.Ctx, strUserId, req.VideoId).Result()
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			rsp.Flag = false
			return err
		}
		rsp.Flag = exist

		return nil

	} else { //step2:LikeUserId不存在key,查询Redis LikeVideoId,key中存在value 返回true，不存在返回false
		if n, err := model.RdbLikeVideoId.Exists(model.Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回默认false,返回错误信息
			if err != nil {
				rsp.Flag = false
				return err
			}
			exist, err := model.RdbLikeVideoId.SIsMember(model.Ctx, strVideoId, req.UserId).Result()
			//如果有问题，说明查询redis失败,返回默认false,返回错误信息
			if err != nil {
				rsp.Flag = false
				return err
			}
			rsp.Flag = exist
			return nil
		} else {
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				rsp.Flag = false
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := model.RdbLikeUserId.Expire(model.Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				rsp.Flag = false
				return err
			}
			//step3:LikeUserId LikeVideoId中都没有对应key，通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
			videoIdList, err := model.GetLikeVideoIdList(req.UserId)
			//如果有问题，说明查询数据库失败，返回默认false,返回错误信息："get likeVideoIdList failed"
			if err != nil {
				rsp.Flag = false
				return err
			}
			//维护Redis LikeUserId(key:strUserId)，遍历videoIdList加入
			for _, likeVideoId := range videoIdList {
				model.RdbLikeUserId.SAdd(model.Ctx, strUserId, likeVideoId)
			}
			//查询Redis LikeUserId,key：strUserId中是否存在value:videoId,存在返回true,不存在返回false
			exist, err := model.RdbLikeUserId.SIsMember(model.Ctx, strUserId, req.VideoId).Result()
			//如果有问题，说明操作redis失败,返回默认false,返回错误信息
			if err != nil {
				rsp.Flag = false
				return err
			}

			rsp.Flag = exist
			return nil
		}
	}
}

//FavouriteCount 根据videoId获取对应点赞数量;
func (e *LikeService) FavouriteCount(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.FavouriteCount request: %v", req)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(req.Id, 10)
	//step1 如果key:strVideoId存在 则计算集合中userId个数
	if n, err := model.RdbLikeVideoId.Exists(model.Ctx, strVideoId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认false,返回错误信息
		if err != nil {
			rsp.Count = 0
			return err
		}
		//获取集合中userId个数
		count, err := model.RdbLikeVideoId.SCard(model.Ctx, strVideoId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			rsp.Count = 0
			return err
		}
		rsp.Count = count - 1
	} else {
		//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
			model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
			rsp.Count = 0
			return err
		}
		//给键值设置有效期，类似于gc机制
		_, err := model.RdbLikeVideoId.Expire(model.Ctx, strVideoId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
			rsp.Count = 0
			return err
		}
		//如果Redis LikeVideoId不存在此key,通过videoId查询likes表,返回所有点赞userId，并维护到Redis LikeVideoId(key:strVideoId)
		//再通过set集合中userId个数,获取点赞数量
		userIdList, err := model.GetLikeUserIdList(req.Id)
		//如果有问题，说明查询数据库失败，返回默认0,返回错误信息："get likeUserIdList failed"
		if err != nil {
			rsp.Count = 0
			return err
		}
		//维护Redis LikeVideoId(key:strVideoId)，遍历userIdList加入
		for _, likeUserId := range userIdList {
			model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, likeUserId)
		}
		//再通过set集合中userId个数,获取点赞数量
		count, err := model.RdbLikeVideoId.SCard(model.Ctx, strVideoId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			rsp.Count = 0
			return err
		}

		rsp.Count = count - 1
		return nil
	}
	return nil
}

// FavouriteAction 根据userId，videoId,actionType对视频进行点赞或者取消赞操作;
func (e *LikeService) FavouriteAction(ctx context.Context, req *pb.ActionReq, rsp *pb.ActionRsp) error {
	log.Infof("Received LikeService.FavouriteAction request: %v", req)
	//将int64 videoId转换为 string strVideoId
	strUserId := strconv.FormatInt(req.UserId, 10)
	//将int64 videoId转换为 string strVideoId
	strVideoId := strconv.FormatInt(req.VideoId, 10)
	//将要操作数据库likes表的信息打入消息队列RmqLikeAdd或者RmqLikeDel
	//拼接打入信息
	sb := strings.Builder{}
	sb.WriteString(strUserId)
	sb.WriteString(" ")
	sb.WriteString(strVideoId)

	//step1:维护Redis LikeUserId、LikeVideoId;
	//执行点赞操作维护
	if req.ActionType == config.LikeAction {
		//查询Redis LikeUserId(key:strUserId)是否已经加载过此信息
		if n, err := model.RdbLikeUserId.Exists(model.Ctx, strUserId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {

				return err
			}
			//如果加载过此信息key:strUserId，则加入value:videoId
			//如果redis LikeUserId 添加失败，数据库操作成功，会有脏数据，所以只有redis操作成功才执行数据库likes表操作
			if _, err1 := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, req.VideoId).Result(); err1 != nil {
				return err
			} else {
				//如果数据库操作失败了，redis是正确数据，客户端显示的是点赞成功，不会影响后续结果
				//只有当该用户取消所有点赞视频的时候redis才会重新加载数据库信息，这时候因为取消赞了必然和数据库信息一致
				//同样这条信息消费成功与否也不重要，因为redis是正确信息,理由如上
				model.RmqLikeAdd.Publish(sb.String())
			}
		} else {
			//如果不存在，则维护Redis LikeUserId 新建key:strUserId,设置过期时间，加入DefaultRedisValue，
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := model.RdbLikeUserId.Expire(model.Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				return err
			}
			videoIdList, err := model.GetLikeVideoIdList(req.UserId)
			//如果有问题，说明查询失败，返回错误信息："get likeVideoIdList failed"
			if err != nil {
				return err
			}
			//遍历videoIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeVideoId := range videoIdList {
				if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, likeVideoId).Result(); err != nil {
					model.RdbLikeUserId.Del(model.Ctx, strUserId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, req.VideoId).Result(); err != nil {
				return err
			} else {
				model.RmqLikeAdd.Publish(sb.String())
			}
		}
		//查询Redis LikeVideoId(key:strVideoId)是否已经加载过此信息
		if n, err := model.RdbLikeVideoId.Exists(model.Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //如果加载过此信息key:strVideoId，则加入value:userId
			//如果redis LikeVideoId 添加失败，返回错误信息
			if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, req.UserId).Result(); err != nil {
				return err
			}
		} else { //如果不存在，则维护Redis LikeVideoId 新建key:strVideoId，设置有效期，加入DefaultRedisValue
			//通过videoId查询likes表,返回所有点赞userId，加入key:strVideoId集合中,
			//再加入当前userId,再更新likes表此条数据
			//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
				model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := model.RdbLikeVideoId.Expire(model.Ctx, strVideoId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
			}
			userIdList, err := model.GetLikeUserIdList(req.VideoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err != nil {
				return err
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, likeUserId).Result(); err != nil {
					model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, req.UserId).Result(); err != nil {
				return err
			}
		}

		return nil

	} else { //执行取消赞操作维护
		//查询Redis LikeUserId(key:strUserId)是否已经加载过此信息
		if n, err := model.RdbLikeUserId.Exists(model.Ctx, strUserId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //防止出现redis数据不一致情况，当redis删除操作成功，才执行数据库更新操作
			if _, err := model.RdbLikeUserId.SRem(model.Ctx, strUserId, req.VideoId).Result(); err != nil {
				return err
			} else {
				//后续数据库的操作，可以在mq里设置若执行数据库更新操作失败，重新消费该信息
				model.RmqLikeDel.Publish(sb.String())
			}
		} else { //如果不存在，则维护Redis LikeUserId 新建key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库
			// 还没更新完出现脏读，或者数据库操作失败造成的脏读
			//通过userId查询likes表,返回所有点赞videoId，加入key:strUserId集合中,
			//再删除当前videoId,再更新likes表此条数据
			//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := model.RdbLikeUserId.Expire(model.Ctx, strUserId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				return err
			}
			videoIdList, err := model.GetLikeVideoIdList(req.UserId)
			//如果有问题，说明查询失败，返回错误信息："get likeVideoIdList failed"
			if err != nil {
				return err
			}
			//遍历videoIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql 数据原子性
			for _, likeVideoId := range videoIdList {
				if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, likeVideoId).Result(); err != nil {
					model.RdbLikeUserId.Del(model.Ctx, strUserId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := model.RdbLikeUserId.SRem(model.Ctx, strUserId, req.VideoId).Result(); err != nil {
				return err
			} else {
				model.RmqLikeDel.Publish(sb.String())
			}
		}

		//查询Redis LikeVideoId(key:strVideoId)是否已经加载过此信息
		if n, err := model.RdbLikeVideoId.Exists(model.Ctx, strVideoId).Result(); n > 0 {
			//如果有问题，说明查询redis失败,返回错误信息
			if err != nil {
				return err
			} //如果加载过此信息key:strVideoId，则删除value:userId
			//如果redis LikeVideoId 删除失败，返回错误信息
			if _, err := model.RdbLikeVideoId.SRem(model.Ctx, strVideoId, req.UserId).Result(); err != nil {
				return err
			}
		} else { //如果不存在，则维护Redis LikeVideoId 新建key:strVideoId,加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库
			// 还没更新完出现脏读，或者数据库操作失败造成的脏读
			//通过videoId查询likes表,返回所有点赞userId，加入key:strVideoId集合中,
			//再删除当前userId,再更新likes表此条数据
			//key:strVideoId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
			if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, config.DefaultRedisValue).Result(); err != nil {
				model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
				return err
			}
			//给键值设置有效期，类似于gc机制
			_, err := model.RdbLikeVideoId.Expire(model.Ctx, strVideoId,
				time.Duration(config.OneMonth)*time.Second).Result()
			if err != nil {
				model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
				return err
			}

			userIdList, err := model.GetLikeUserIdList(req.VideoId)
			//如果有问题，说明查询失败，返回错误信息："get likeUserIdList failed"
			if err != nil {
				return err
			}
			//遍历userIdList,添加进key的集合中，若失败，删除key，并返回错误信息，这么做的原因是防止脏读，
			//保证redis与mysql数据一致性
			for _, likeUserId := range userIdList {
				if _, err := model.RdbLikeVideoId.SAdd(model.Ctx, strVideoId, likeUserId).Result(); err != nil {
					model.RdbLikeVideoId.Del(model.Ctx, strVideoId)
					return err
				}
			}
			//这样操作理由同上
			if _, err := model.RdbLikeVideoId.SRem(model.Ctx, strVideoId, req.UserId).Result(); err != nil {
				return err
			}
		}
		return nil
	}
}

//GetFavouriteList 根据userId，curId(当前用户Id),返回userId的点赞列表;
func (e *LikeService) GetFavouriteList(ctx context.Context, req *pb.UserCurReq, rsp *pb.FavouriteListRsp) error {
	log.Infof("Received LikeService.GetFavouriteList request: %v", req)
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(req.UserId, 10)
	//step1:查询Redis LikeUserId,如果key：strUserId存在,则获取集合中全部videoId
	if n, err := model.RdbLikeUserId.Exists(model.Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认nil,返回错误信息
		if err != nil {
			rsp.Video = nil
			return err
		}
		//获取集合中全部videoId
		videoIdList, err := model.RdbLikeUserId.SMembers(model.Ctx, strUserId).Result()
		//如果有问题，说明查询redis失败,返回默认nil,返回错误信息
		if err != nil {
			rsp.Video = nil
			return err
		}
		//提前开辟点赞列表空间
		favoriteVideoList := new([]*pb.Video)
		//采用协程并发将Video类型对象添加到集合中去
		i := len(videoIdList) - 1 //去掉DefaultRedisValue
		if i == 0 {
			rsp.Video = *favoriteVideoList
		}
		var wg sync.WaitGroup
		wg.Add(i)
		for j := 0; j <= i; j++ {
			//将string videoId转换为 int64 VideoId
			videoId, _ := strconv.ParseInt(videoIdList[j], 10, 64)
			if videoId == config.DefaultRedisValue {
				continue
			}
			go addFavouriteVideoList(videoId, req.CurId, favoriteVideoList, &wg)
		}
		wg.Wait()
		rsp.Video = *favoriteVideoList
		return nil
	} else { //如果Redis LikeUserId不存在此key,通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
		//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
			model.RdbLikeUserId.Del(model.Ctx, strUserId)
			rsp.Video = nil
			return err
		}
		//给键值设置有效期，类似于gc机制
		_, err := model.RdbLikeUserId.Expire(model.Ctx, strUserId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			model.RdbLikeUserId.Del(model.Ctx, strUserId)
			rsp.Video = nil
			return err
		}
		videoIdList, err := model.GetLikeVideoIdList(req.UserId)
		//如果有问题，说明查询数据库失败，返回nil和错误信息:"get likeVideoIdList failed"
		if err != nil {
			rsp.Video = nil
			return err
		}
		//提前开辟点赞列表空间
		favoriteVideoList := new([]*pb.Video)
		//采用协程并发将Video类型对象添加到集合中去
		i := len(videoIdList) - 1 //去掉DefaultRedisValue
		if i == 0 {
			rsp.Video = *favoriteVideoList
		}
		var wg sync.WaitGroup
		wg.Add(i)
		for j := 0; j <= i; j++ {
			if videoIdList[j] == config.DefaultRedisValue {
				continue
			}
			go addFavouriteVideoList(videoIdList[j], req.CurId, favoriteVideoList, &wg)
		}
		wg.Wait()
		rsp.Video = *favoriteVideoList
		return nil
	}

}

//TotalFavourite 根据userId获取这个用户总共被点赞数量
func (e *LikeService) TotalFavourite(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.TotalFavourite request: %v", req)
	//根据userId获取这个用户的发布视频列表信息
	videoMicro := InitMicro()
	videoClient := videoService.NewVideoService("videoService", videoMicro.Client())

	videoRsp, err := videoClient.GetVideoIdList(context.TODO(), &videoService.VideoIdReq{
		UserId: req.Id,
	})
	if err != nil {
		rsp.Count = 0
		return err
	}
	var sum int64 //该用户的总被点赞数
	//提前开辟空间,存取每个视频的点赞数
	videoLikeCountList := new([]int64)
	//采用协程并发将对应videoId的点赞数添加到集合中去
	i := len(videoRsp.VideoId)
	var wg sync.WaitGroup
	wg.Add(i)
	for j := 0; j < i; j++ {
		go addVideoLikeCount(videoRsp.VideoId[j], videoLikeCountList, &wg)
	}
	wg.Wait()
	//遍历累加，求总被点赞数
	for _, count := range *videoLikeCountList {
		sum += count
	}
	rsp.Count = sum
	return nil
}

//FavouriteVideoCount 根据userId获取这个用户点赞视频数量
func (e *LikeService) FavouriteVideoCount(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.FavouriteVideoCount request: %v", req)
	//将int64 userId转换为 string strUserId
	strUserId := strconv.FormatInt(req.Id, 10)
	//step1:查询Redis LikeUserId,如果key：strUserId存在,则获取集合中元素个数
	if n, err := model.RdbLikeUserId.Exists(model.Ctx, strUserId).Result(); n > 0 {
		//如果有问题，说明查询redis失败,返回默认0,返回错误信息
		if err != nil {
			rsp.Count = 0
			return err
		} else {
			count, err := model.RdbLikeUserId.SCard(model.Ctx, strUserId).Result()
			//如果有问题，说明操作redis失败,返回默认0,返回错误信息
			if err != nil {
				rsp.Count = 0
				return err
			}
			rsp.Count = count - 1
			return nil //去掉DefaultRedisValue

		}
	} else { //如果Redis LikeUserId不存在此key,通过userId查询likes表,返回所有点赞videoId，并维护到Redis LikeUserId(key:strUserId)
		//再通过set集合中userId个数,获取点赞数量
		//key:strUserId，加入value:DefaultRedisValue,过期才会删，防止删最后一个数据的时候数据库还没更新完出现脏读，或者数据库操作失败造成的脏读
		if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, config.DefaultRedisValue).Result(); err != nil {
			model.RdbLikeUserId.Del(model.Ctx, strUserId)
			rsp.Count = 0
			return err
		}
		//给键值设置有效期，类似于gc机制
		_, err := model.RdbLikeUserId.Expire(model.Ctx, strUserId,
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			model.RdbLikeUserId.Del(model.Ctx, strUserId)
			rsp.Count = 0
			return err
		}
		videoIdList, err := model.GetLikeVideoIdList(req.Id)
		//如果有问题，说明查询数据库失败，返回默认0,返回错误信息："get likeVideoIdList failed"
		if err != nil {
			rsp.Count = 0
			return err
		}
		//维护Redis LikeUserId(key:strUserId)，遍历videoIdList加入
		for _, likeVideoId := range videoIdList {
			if _, err := model.RdbLikeUserId.SAdd(model.Ctx, strUserId, likeVideoId).Result(); err != nil {
				model.RdbLikeUserId.Del(model.Ctx, strUserId)
				rsp.Count = 0
				return err
			}
		}
		//再通过set集合中videoId个数,获取点赞数量
		count, err := model.RdbLikeUserId.SCard(model.Ctx, strUserId).Result()
		//如果有问题，说明操作redis失败,返回默认0,返回错误信息
		if err != nil {
			rsp.Count = 0
			return err
		}
		rsp.Count = count - 1
		return nil //去掉DefaultRedisValue
	}
}
