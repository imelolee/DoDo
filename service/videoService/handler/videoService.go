package handler

import (
	"context"
	"github.com/gogf/gf/util/gconv"
	"github.com/satori/go.uuid"
	log "go-micro.dev/v4/logger"
	"time"
	"videoService/config"
	"videoService/model"

	pb "videoService/proto"
)

type VideoService struct{}

// Feed 通过传入时间戳，当前用户的id，返回对应的视频数组，以及视频数组中最早的发布时间
func (e *VideoService) Feed(ctx context.Context, req *pb.FeedReq, rsp *pb.FeedRsp) error {
	log.Infof("Received VideoService.Feed request: %v", req)
	//创建对应返回视频的切片数组，提前将切片的容量设置好，可以减少切片扩容的性能
	videos := make([]model.FeedVideo, 0, config.VideoCount)
	//根据传入的时间，获得传入时间前n个视频，可以通过config.videoCount来控制
	lastTime := time.Unix(req.LatestTime, 0)
	tableVideos, err := model.GetVideosByLastTime(lastTime)
	if err != nil {
		rsp.NextTime = 0
		rsp.VideoList = nil
		return err
	}
	//将数据通过copyVideos进行处理，在拷贝的过程中对数据进行组装
	err = copyVideos(&videos, &tableVideos, req.UserId)
	if err != nil {
		rsp.NextTime = 0
		rsp.VideoList = nil
		return err
	}
	//返回数据，同时获得视频中最早的时间返回
	var tmpVideo []*pb.Video
	gconv.Struct(videos, &tmpVideo)

	if len(tableVideos) == 0 {
		rsp.NextTime = time.Now().Unix()
	} else {
		rsp.NextTime = tableVideos[0].PublishTime.Unix()
	}

	rsp.VideoList = tmpVideo

	return nil
}

func (e *VideoService) GetVideo(ctx context.Context, req *pb.GetVideoReq, rsp *pb.GetVideoRsp) error {
	log.Infof("Received VideoService.GetVideo request: %v", req)
	//初始化video对象
	var video model.FeedVideo
	//从数据库中查询数据，如果查询不到数据，就直接失败返回，后续流程就不需要执行了
	data, err := model.GetVideoByVideoId(req.VideoId)
	if err != nil {
		rsp.Video = nil
		return err
	}

	//插入从数据库中查到的数据
	creatVideo(&video, &data, req.UserId)
	var tmpVideo *pb.Video
	gconv.Struct(video, &tmpVideo)

	rsp.Video = tmpVideo
	return nil
}

func (e *VideoService) Publish(ctx context.Context, req *pb.PublishReq, rsp *pb.PublishRsp) error {
	log.Infof("Received VideoService.Publish request: %v", req.Title)
	//将视频流上传到视频服务器，保存视频链接
	file := req.Data
	//生成一个uuid作为视频的名字
	videoName := uuid.NewV4().String() + req.FileExt

	err := uploadQiniu(file, videoName, req.FileSize)
	if err != nil {
		return err
	}

	//在服务器上执行ffmpeg 从视频流中获取第一帧截图，并上传图片服务器，保存图片链接
	imageName := uuid.NewV4().String()

	//组装并持久化
	err = model.Save(videoName, imageName, req.UserId, req.Title)
	if err != nil {
		return err
	}
	return nil
}

func (e *VideoService) GetPublishList(ctx context.Context, req *pb.PublishListReq, rsp *pb.PublishListRsp) error {
	log.Infof("Received VideoService.GetPublishList request: %v", req)
	//依据用户id查询所有的视频，获取视频列表
	data, err := model.GetVideosByAuthorId(req.UserId)
	if err != nil {
		rsp.Video = nil
		return err
	}
	//提前定义好切片长度
	result := make([]model.FeedVideo, 0, len(data))
	//调用拷贝方法，将数据进行转换
	err = copyVideos(&result, &data, req.CurId)
	if err != nil {
		rsp.Video = nil
		return err
	}
	//如果数据没有问题，则直接返回
	var tmpVideo []*pb.Video
	gconv.Struct(result, &tmpVideo)

	rsp.Video = tmpVideo
	return nil
}

func (e *VideoService) GetVideoIdList(ctx context.Context, req *pb.VideoIdReq, rsp *pb.VideoIdRsp) error {
	log.Infof("Received VideoService.GetVideoIdList request: %v", req)
	var id []int64
	//通过pluck来获得单独的切片
	result := model.Db.Model(&model.Video{}).Where("author_id = ?", req.UserId).Pluck("id", &id)
	//如果出现问题，返回对应到空，并且返回error
	if result.Error != nil {
		rsp.VideoId = nil
		return result.Error
	}

	rsp.VideoId = id

	return nil
}
