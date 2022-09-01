package model

import (
	"github.com/gogf/gf/util/gconv"
	uuid "github.com/satori/go.uuid"
	"time"
	"videoService/config"
	pb "videoService/proto"
)

// GetVideosByLastTime 依据一个时间，来获取这个时间之前的一些视频
func GetVideosByLastTime(lastTime time.Time) ([]Video, error) {
	videos := make([]Video, config.VideoCount)
	result := Db.Where("publish_time<?", lastTime).Order("publish_time desc").Limit(config.VideoCount).Find(&videos)
	if result.Error != nil {
		return videos, result.Error
	}
	return videos, nil
}

// GetVideoByVideoId 依据VideoId来获得视频信息
func GetVideoByVideoId(videoId int64) (Video, error) {

	var tableVideo Video
	if Db == nil {
		InitDb()
	}
	result := Db.Where("id = ?", videoId).Order("id desc").Limit(config.VideoCount).Find(&tableVideo)
	if result.Error != nil {
		return tableVideo, result.Error
	}
	return tableVideo, nil

}

// Save 保存视频记录
func Save(videoName string, imageName string, authorId int64, title string) error {
	var video Video
	video.PublishTime = time.Now()
	video.PlayUrl = config.Domain + videoName
	video.CoverUrl = video.PlayUrl
	video.AuthorId = authorId
	video.Title = title
	result := Db.Save(&video)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetVideosByAuthorId 根据作者的id来查询对应数据库数据，并TableVideo返回切片
func GetVideosByAuthorId(authorId int64) ([]Video, error) {
	//建立结果集接收
	var data []Video
	//初始化db
	//Init()
	result := Db.Where(&Video{AuthorId: authorId}).Find(&data)
	//如果出现问题，返回对应到空，并且返回error
	if result.Error != nil {
		return nil, result.Error
	}
	return data, nil
}

// Feed 通过传入时间戳，当前用户的id，返回对应的视频数组，以及视频数组中最早的发布时间
func Feed(latestTime int64, userId int64) (int64, []*pb.Video, error) {
	//创建对应返回视频的切片数组，提前将切片的容量设置好，可以减少切片扩容的性能
	videos := make([]FeedVideo, 0, config.VideoCount)
	//根据传入的时间，获得传入时间前n个视频，可以通过config.videoCount来控制
	lastTime := time.Unix(latestTime, 0)
	tableVideos, err := GetVideosByLastTime(lastTime)
	if err != nil {
		return 0, nil, err
	}
	//将数据通过copyVideos进行处理，在拷贝的过程中对数据进行组装
	err = copyVideos(&videos, &tableVideos, userId)
	if err != nil {
		return 0, nil, err
	}
	//返回数据，同时获得视频中最早的时间返回
	var tmpVideo []*pb.Video
	gconv.Struct(videos, &tmpVideo)

	var nextTime int64
	if len(tableVideos) == 0 {
		nextTime = time.Now().Unix()
	} else {
		nextTime = tableVideos[0].PublishTime.Unix()
	}

	return nextTime, tmpVideo, nil
}

func GetVideo(videoId int64, userId int64) (*pb.Video, error) {
	//初始化video对象
	var video FeedVideo
	//从数据库中查询数据，如果查询不到数据，就直接失败返回，后续流程就不需要执行了
	data, err := GetVideoByVideoId(videoId)
	if err != nil {
		return nil, err
	}

	//插入从数据库中查到的数据
	creatVideo(&video, &data, userId)
	var tmpVideo *pb.Video
	gconv.Struct(video, &tmpVideo)

	return tmpVideo, nil
}

func Publish(req *pb.PublishReq) error {
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
	err = Save(videoName, imageName, req.UserId, req.Title)
	if err != nil {
		return err
	}
	return nil
}

func GetPublishList(userId int64, curId int64) ([]*pb.Video, error) {
	//依据用户id查询所有的视频，获取视频列表
	data, err := GetVideosByAuthorId(userId)
	if err != nil {
		return nil, err
	}
	//提前定义好切片长度
	result := make([]FeedVideo, 0, len(data))
	//调用拷贝方法，将数据进行转换
	err = copyVideos(&result, &data, curId)
	if err != nil {
		return nil, err
	}
	//如果数据没有问题，则直接返回
	var tmpVideo []*pb.Video
	gconv.Struct(result, &tmpVideo)

	return tmpVideo, nil
}

func GetVideoIdList(userId int64) ([]int64, error) {
	var id []int64
	//通过pluck来获得单独的切片
	result := Db.Model(&Video{}).Where("author_id = ?", userId).Pluck("id", &id)
	//如果出现问题，返回对应到空，并且返回error
	if result.Error != nil {
		return nil, result.Error
	}

	return id, nil
}
