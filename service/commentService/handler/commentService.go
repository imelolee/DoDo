package handler

import (
	"commentService/model"
	pb "commentService/proto"
	"context"
	log "go-micro.dev/v4/logger"
)

type CommentService struct{}

// CountFromVideoId 使用video id 查询Comment数量
func (e *CommentService) CountFromVideoId(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received CommentService.CountFromVideoId request: %v", req)
	count, err := model.CountFromVideoId(req.Id)
	rsp.Count = count
	return err
}

// Send 发表评论
func (e *CommentService) Send(ctx context.Context, req *pb.CommentReq, rsp *pb.CommentRsp) error {
	log.Infof("Received CommentService.Send request: %v", req)
	comment, err := model.Send(req.Comment)
	rsp.CommentInfo = comment
	return err
}

// Delete 删除评论，传入评论id
func (e *CommentService) Delete(ctx context.Context, req *pb.IdReq, rsp *pb.DelRsp) error {
	log.Infof("Received CommentService.DelComment request: %v", req)
	err := model.Delete(req.Id)
	return err
}

// GetList 查看评论列表-返回评论list
func (e *CommentService) GetList(ctx context.Context, req *pb.VideoUserReq, rsp *pb.CommentListRsp) error {
	log.Infof("Received CommentService.GetList request: %v", req)
	comInfoList, err := model.GetList(req.VideoId, req.UserId)
	rsp.CommentInfo = comInfoList
	return err
}
