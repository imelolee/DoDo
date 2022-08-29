package handler

import (
	"context"
	"followService/config"
	"followService/model"
	"sync"

	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"

	"strconv"
	"strings"

	pb "followService/proto"
	userModel "userService/model"
	userService "userService/proto"
)

type FollowService struct{}

func (e *FollowService) IsFollowing(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.IsFollowing request: %v", req)
	// 先查Redis里面是否有此关系。
	if flag, _ := model.RdbFollowingPart.SIsMember(model.Ctx, strconv.Itoa(int(req.UserId)), req.TargetId).Result(); flag {
		// 重现设置过期时间。
		model.RdbFollowingPart.Expire(model.Ctx, strconv.Itoa(int(req.UserId)), config.ExpireTime)
		rsp.Flag = true
		return nil
	}
	// SQL 查询。
	relation, err := model.FindRelation(req.UserId, req.TargetId)

	if nil != err {
		rsp.Flag = false
		return err
	}
	if nil == relation {
		rsp.Flag = false
		return nil
	}
	// 存在此关系，将其注入Redis中。
	go addRelationToRedis(int(req.UserId), int(req.TargetId))
	rsp.Flag = true
	return nil
}

func (e *FollowService) GetFollowerCnt(ctx context.Context, req *pb.UserIdReq, rsp *pb.CountRsp) error {
	log.Infof("Received FollowService.GetFollowerCnt request: %v", req)
	// 查Redis中是否已经存在。
	if cnt, _ := model.RdbFollowers.SCard(model.Ctx, strconv.Itoa(int(req.UserId))).Result(); cnt > 0 {
		// 更新过期时间。
		model.RdbFollowers.Expire(model.Ctx, strconv.Itoa(int(req.UserId)), config.ExpireTime)
		rsp.Count = cnt - 1
		return nil
	}
	// SQL中查询。
	ids, err := model.GetFollowersIds(req.UserId)
	if nil != err {
		rsp.Count = 0
		return err
	}
	// 将数据存入Redis.
	// 更新followers 和 followingPart
	go addFollowersToRedis(int(req.UserId), ids)
	rsp.Count = int64(len(ids))

	return nil
}

func (e *FollowService) GetFollowingCnt(ctx context.Context, req *pb.UserIdReq, rsp *pb.CountRsp) error {
	log.Infof("Received FollowService.GetFollowingCnt request: %v", req)
	// 查看Redis中是否有关注数。
	if cnt, _ := model.RdbFollowing.SCard(model.Ctx, strconv.Itoa(int(req.UserId))).Result(); cnt > 0 {
		// 更新过期时间。
		model.RdbFollowing.Expire(model.Ctx, strconv.Itoa(int(req.UserId)), config.ExpireTime)
		rsp.Count = cnt - 1
		return nil

	}
	// 用SQL查询。
	ids, err := model.GetFollowingIds(req.UserId)

	if nil != err {
		rsp.Count = 0
		return err
	}
	// 更新Redis中的followers和followPart
	go addFollowingToRedis(int(req.UserId), ids)
	rsp.Count = int64(len(ids))

	return nil
}

func (e *FollowService) AddFollowRelation(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.AddFollowRelation request: %v", req)
	// 加信息打入消息队列。
	sb := strings.Builder{}
	sb.WriteString(strconv.Itoa(int(req.UserId)))
	sb.WriteString(" ")
	sb.WriteString(strconv.Itoa(int(req.TargetId)))
	model.RmqFollowAdd.Publish(sb.String())
	// 记录日志
	// 更新redis信息。
	updateRedisWithAdd(req.UserId, req.TargetId)

	rsp.Flag = true

	return nil
}

func (e *FollowService) DeleteFollowRelation(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.DeleteFollowRelation request: %v", req)
	// 加信息打入消息队列。
	sb := strings.Builder{}
	sb.WriteString(strconv.Itoa(int(req.UserId)))
	sb.WriteString(" ")
	sb.WriteString(strconv.Itoa(int(req.TargetId)))
	model.RmqFollowDel.Publish(sb.String())
	// 记录日志
	// 更新redis信息。
	updateRedisWithDel(req.UserId, req.TargetId)

	rsp.Flag = true

	return nil
}

func (e *FollowService) GetFollowing(ctx context.Context, req *pb.UserIdReq, rsp *pb.UserListRsp) error {
	log.Infof("Received FollowService.GetFollowing request: %v", req)
	// 获取关注对象的id数组。
	ids, err := model.GetFollowingIds(req.UserId)
	// 查询出错
	if nil != err {
		rsp.User = nil
		return err
	}
	// 没得关注者
	if nil == ids {
		rsp.User = nil
		return nil
	}
	// 根据每个id来查询用户信息。
	followingNum := len(ids)
	userMicro := InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())
	users := make([]pb.FeedUser, followingNum)
	for i := 0; i < followingNum; i++ {

		userRsp, _ := userClient.GetFeedUserByIdWithCurId(context.TODO(), &userService.CurIdReq{
			Id:    ids[i],
			CurId: req.UserId,
		})

		var tmpUser *pb.FeedUser
		gconv.Struct(userRsp.User, &tmpUser)

		users[i] = *tmpUser

	}
	// 返回关注对象列表
	var followUser []*pb.FeedUser
	_ = gconv.Struct(users, &followUser)
	rsp.User = followUser

	return nil
}

func (e *FollowService) GetFollowers(ctx context.Context, req *pb.UserIdReq, rsp *pb.UserListRsp) error {
	log.Infof("Received FollowService.GetFollowers request: %v", req)
	// 获取粉丝的id数组。
	ids, err := model.GetFollowersIds(req.UserId)
	// 查询出错
	if nil != err {
		rsp.User = nil
		return err
	}
	// 没得粉丝
	if nil == ids {
		rsp.User = nil
		return nil
	}
	// 根据每个id来查询用户信息。
	len := len(ids)
	if len > 0 {
		len -= 1
	}
	users := make([]userModel.FeedUser, len)
	var wg sync.WaitGroup
	wg.Add(len)
	i, j := 0, 0
	for ; i < len; j++ {
		// 越过-1
		if ids[j] == -1 {
			continue
		}
		//开启协程来查。
		go func(i int, idx int64) {
			defer wg.Done()
			userMicro := InitMicro()
			userClient := userService.NewUserService("userService", userMicro.Client())
			// 调用微服务的方法
			userRsp, _ := userClient.GetFeedUserByIdWithCurId(context.TODO(), &userService.CurIdReq{
				Id:    idx,
				CurId: req.UserId,
			})
			var user userModel.FeedUser
			_ = gconv.Struct(userRsp.User, user)

			users[i] = user

		}(i, ids[i])
		i++
	}
	wg.Wait()
	// 返回粉丝列表。

	var followerUser []*pb.FeedUser
	_ = gconv.Struct(users, &followerUser)
	rsp.User = followerUser

	return nil
}
