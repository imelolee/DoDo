package model

import (
	"time"
	"videoService/config"
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
	//Init()
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
