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
