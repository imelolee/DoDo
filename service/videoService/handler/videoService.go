package handler

import (
	"context"
	log "go-micro.dev/v4/logger"
	"videoService/model"

	pb "videoService/proto"
)

type VideoService struct{}

// Feed 通过传入时间戳，当前用户的id，返回对应的视频数组，以及视频数组中最早的发布时间
func (e *VideoService) Feed(ctx context.Context, req *pb.FeedReq, rsp *pb.FeedRsp) error {
	log.Infof("Received VideoService.Feed request: %v", req)
	nextTime, videoList, err := model.Feed(req.LatestTime, req.UserId)
	rsp.VideoList = videoList
	rsp.NextTime = nextTime
	return err
}

func (e *VideoService) GetVideo(ctx context.Context, req *pb.GetVideoReq, rsp *pb.GetVideoRsp) error {
	log.Infof("Received VideoService.GetVideo request: %v", req)
	video, err := model.GetVideo(req.VideoId, req.UserId)

	rsp.Video = video
	return err
}

func (e *VideoService) Publish(ctx context.Context, req *pb.PublishReq, rsp *pb.PublishRsp) error {
	log.Infof("Received VideoService.Publish request: %v", req.Title)
	err := model.Publish(req)
	return err
}

func (e *VideoService) GetPublishList(ctx context.Context, req *pb.PublishListReq, rsp *pb.PublishListRsp) error {
	log.Infof("Received VideoService.GetPublishList request: %v", req)
	video, err := model.GetPublishList(req.UserId, req.CurId)

	rsp.Video = video
	return err
}

func (e *VideoService) GetVideoIdList(ctx context.Context, req *pb.VideoIdReq, rsp *pb.VideoIdRsp) error {
	log.Infof("Received VideoService.GetVideoIdList request: %v", req)
	videoId, err := model.GetVideoIdList(req.UserId)
	rsp.VideoId = videoId
	return err
}
