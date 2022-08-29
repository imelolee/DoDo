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
		return err
	}
	rsp.User = tableUsers
	return nil
}

func (e *UserService) GetTableUserByUsername(ctx context.Context, req *pb.UsernameReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserByUsername request: %v\n", req)
	tableUser, err := model.GetTableUserByUsername(req.Name)
	if err != nil {
		rsp.User = nil
	}
	rsp.User = tableUser
	return nil
}

func (e *UserService) GetTableUserById(ctx context.Context, req *pb.IdReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		rsp.User = nil

	}
	rsp.User = tableUser
	return nil
}

func (e *UserService) InsertTableUser(ctx context.Context, req *pb.UserReq, rsp *pb.BoolRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	success := model.InsertTableUser(req.User)
	if success == false {
		rsp.Flag = false
		return nil
	} else {
		rsp.Flag = success
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
		return err
	}

	followMicro := InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())

	followRsp, err := followClient.GetFollowingCnt(context.TODO(), &followService.UserIdReq{
		UserId: req.Id,
	})
	if err != nil {
		rsp.User = nil
		return err
	}
	followerRsp, err := followClient.GetFollowerCnt(context.TODO(), &followService.UserIdReq{
		UserId: req.Id,
	})
	if err != nil {
		rsp.User = nil
		return err
	}

	likeMicro := InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())

	totalRsp, err := likeClient.TotalFavourite(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})
	countRsp, err := likeClient.FavouriteVideoCount(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})

	feedUser := model.FeedUser{
		Id:             req.Id,
		Name:           tableUser.Name,
		FollowCount:    followRsp.Count,
		FollowerCount:  followerRsp.Count,
		IsFollow:       false,
		TotalFavorited: totalRsp.Count,
		FavoriteCount:  countRsp.Count,
	}

	var tmpUser *pb.FeedUser
	tmpUser = new(pb.FeedUser)
	err = gconv.Struct(feedUser, &tmpUser)

	rsp.User = tmpUser

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
		return err
	}
	followCount, _ := followModel.GetFollowingCnt(req.Id)
	followerCount, _ := followModel.GetFollowerCnt(req.Id)

	followMicro := InitMicro()
	followClient := followService.NewFollowService("followService", followMicro.Client())

	followRsp, _ := followClient.IsFollowing(context.TODO(), &followService.UserTargetReq{
		UserId:   req.CurId,
		TargetId: req.Id,
	})

	likeMicro := InitMicro()
	likeClient := likeService.NewLikeService("likeService", likeMicro.Client())

	totalRsp, _ := likeClient.TotalFavourite(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})

	countRsp, _ := likeClient.FavouriteVideoCount(context.TODO(), &likeService.IdReq{
		Id: req.Id,
	})
	tmpUser := model.FeedUser{
		Id:             req.Id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       followRsp.Flag,
		TotalFavorited: totalRsp.Count,
		FavoriteCount:  countRsp.Count,
	}

	var feedUser *pb.FeedUser
	err = gconv.Struct(tmpUser, &feedUser)

	rsp.User = feedUser

	return nil
}
