package handler

import (
	"context"
	followModel "followService/model"
	followService "followService/proto"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	likeService "likeService/proto"
	"userService/model"
	pb "userService/proto"
)

type UserService struct{}

func (e *UserService) GetTableUserList(ctx context.Context, req *pb.Req, rsp *pb.UserListRsp) error {
	log.Infof("Received UserService.GetTableUserList request: %v\n", req)
	tableUsers, err := model.GetTableUserList()
	if err != nil {
		rsp.User = nil
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户查询失败"
		return err
	}
	rsp.User = tableUsers
	rsp.StatusCode = 0
	rsp.StatusMsg = "用户查询成功"
	return nil
}

func (e *UserService) GetTableUserByUsername(ctx context.Context, req *pb.UsernameReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserByUsername request: %v\n", req)
	tableUser, err := model.GetTableUserByUsername(req.Name)
	if err != nil {
		rsp.User = nil
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户查询失败"
		return err
	}
	rsp.User = tableUser
	rsp.StatusCode = 0
	rsp.StatusMsg = "用户查询成功"
	return nil
}

func (e *UserService) GetTableUserById(ctx context.Context, req *pb.IdReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		rsp.User = nil
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户查询失败"
	}
	rsp.User = tableUser
	rsp.StatusCode = 0
	rsp.StatusMsg = "用户查询成功"
	return nil
}

func (e *UserService) InsertTableUser(ctx context.Context, req *pb.UserReq, rsp *pb.BoolRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	success := model.InsertTableUser(req.User)
	if success == false {
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户插入失败"
		rsp.Flag = false
		return nil
	} else {
		rsp.Flag = success
		rsp.StatusCode = 0
		rsp.StatusMsg = "用户插入成功"
		return nil
	}

}

// GetFeedUserById 未登录情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserById(ctx context.Context, req *pb.IdReq, rsp *pb.FeedUserRsp) error {
	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		rsp.User = &user
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户查询失败"
		return err
	}
	followCount, _ := followModel.GetFollowingCnt(req.Id)
	followerCount, _ := followModel.GetFollowerCnt(req.Id)

	likeMicro := InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())

	totalFavorited, err := likeClient.TotalFavourite(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})
	favoritedCount, err := likeClient.FavouriteVideoCount(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})

	feedUser := model.FeedUser{
		Id:             req.Id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       false,
		TotalFavorited: totalFavorited.Count,
		FavoriteCount:  favoritedCount.Count,
	}

	var tmpUser *pb.FeedUser
	err = gconv.Struct(feedUser, &tmpUser)

	rsp.User = tmpUser
	rsp.StatusCode = 0
	rsp.StatusMsg = "用户查询成功"

	return nil
}

// GetFeedUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserByIdWithCurId(ctx context.Context, req *pb.CurIdReq, rsp *pb.FeedUserRsp) error {
	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		rsp.User = &user
		rsp.StatusCode = -1
		rsp.StatusMsg = "用户查询失败"
		return err
	}
	followCount, _ := followModel.GetFollowingCnt(req.Id)
	followerCount, _ := followModel.GetFollowerCnt(req.Id)

	followMicro := InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())

	isfollow, _ := followClient.IsFollowing(context.TODO(), &followService.UserTargetReq{
		UserId:   req.CurId,
		TargetId: req.Id,
	})

	likeMicro := InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())

	totalFavorited, _ := likeClient.TotalFavourite(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})

	favoritedCount, _ := likeClient.FavouriteVideoCount(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})
	tmpUser := model.FeedUser{
		Id:             req.Id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       isfollow.Flag,
		TotalFavorited: totalFavorited.Count,
		FavoriteCount:  favoritedCount.Count,
	}

	var feedUser *pb.FeedUser
	err = gconv.Struct(tmpUser, &feedUser)

	rsp.User = feedUser
	rsp.StatusCode = 0
	rsp.StatusMsg = "用户查询成功"

	return nil
}
