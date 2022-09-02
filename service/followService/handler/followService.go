package handler

import (
	"context"
	"followService/model"
	log "go-micro.dev/v4/logger"

	pb "followService/proto"
)

type FollowService struct{}

func (e *FollowService) IsFollowing(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.IsFollowing request: %v", req)
	flag, err := model.IsFollowing(req.UserId, req.TargetId)
	rsp.Flag = flag
	return err
}

func (e *FollowService) GetFollowerCnt(ctx context.Context, req *pb.UserIdReq, rsp *pb.CountRsp) error {
	log.Infof("Received FollowService.GetFollowerCnt request: %v", req)
	count, err := model.GetFollowerCnt(req.UserId)
	rsp.Count = count

	return err
}

func (e *FollowService) GetFollowingCnt(ctx context.Context, req *pb.UserIdReq, rsp *pb.CountRsp) error {
	log.Infof("Received FollowService.GetFollowingCnt request: %v", req)
	count, err := model.GetFollowingCnt(req.UserId)
	rsp.Count = count

	return err
}

func (e *FollowService) AddFollowRelation(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.AddFollowRelation request: %v", req)
	flag, err := model.AddFollowRelation(req.UserId, req.TargetId)

	rsp.Flag = flag
	return err
}

func (e *FollowService) DeleteFollowRelation(ctx context.Context, req *pb.UserTargetReq, rsp *pb.BoolRsp) error {
	log.Infof("Received FollowService.DeleteFollowRelation request: %v", req)
	flag, err := model.DeleteFollowRelation(req.UserId, req.TargetId)
	rsp.Flag = flag

	return err
}

func (e *FollowService) GetFollowing(ctx context.Context, req *pb.UserIdReq, rsp *pb.UserListRsp) error {
	log.Infof("Received FollowService.GetFollowing request: %v", req)
	followUser, err := model.GetFollowing(req.UserId)
	rsp.User = followUser

	return err
}

func (e *FollowService) GetFollowers(ctx context.Context, req *pb.UserIdReq, rsp *pb.UserListRsp) error {
	log.Infof("Received FollowService.GetFollowers request: %v", req)
	followerUser, err := model.GetFollowers(req.UserId)
	rsp.User = followerUser

	return err
}
