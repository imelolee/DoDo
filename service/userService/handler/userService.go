package handler

import (
	"context"
	log "go-micro.dev/v4/logger"
	"userService/model"
	pb "userService/proto"
)

type UserService struct{}

func (e *UserService) GetTableUserList(ctx context.Context, req *pb.Req, rsp *pb.UserListRsp) error {
	log.Infof("Received UserService.GetTableUserList request: %v\n", req)
	tableUsers, err := model.GetTableUserList()
	rsp.User = tableUsers
	return err
}

func (e *UserService) GetTableUserByUsername(ctx context.Context, req *pb.UsernameReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserByUsername request: %v\n", req)
	tableUser, err := model.GetTableUserByUsername(req.Name)
	rsp.User = tableUser
	return err
}

func (e *UserService) GetTableUserById(ctx context.Context, req *pb.IdReq, rsp *pb.UserRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	tableUser, err := model.GetTableUserById(req.Id)
	rsp.User = tableUser
	return err
}

func (e *UserService) InsertTableUser(ctx context.Context, req *pb.UserReq, rsp *pb.BoolRsp) error {
	log.Infof("Received UserService.GetTableUserById request: %v\n", req)
	success := model.InsertTableUser(req.User)

	rsp.Flag = success
	return nil
}

// GetFeedUserById 未登录情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserById(ctx context.Context, req *pb.IdReq, rsp *pb.FeedUserRsp) error {
	log.Infof("Received UserService.GetFeedUserById request: %v\n", req)
	user, err := model.GetFeedUserById(req.Id)
	rsp.User = user
	return err
}

// GetFeedUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserByIdWithCurId(ctx context.Context, req *pb.CurIdReq, rsp *pb.FeedUserRsp) error {
	log.Infof("Received UserService.GetFeedUserByIdWithCurId request: %v\n", req)
	feedUser, err := model.GetFeedUserByIdWithCurId(req.CurId, req.Id)
	rsp.User = feedUser
	return err
}
