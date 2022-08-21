package model

import (
	"time"
	userModel "userService/model"
)

type Video struct {
	Id          int64 `json:"id"`
	AuthorId    int64
	PlayUrl     string `json:"play_url"`
	CoverUrl    string `json:"cover_url"`
	PublishTime time.Time
	Title       string `json:"title"` //视频名，5.23添加
}

type FeedVideo struct {
	Video
	Author        userModel.FeedUser `json:"author"`
	FavoriteCount int64              `json:"favorite_count"`
	CommentCount  int64              `json:"comment_count"`
	IsFavorite    bool               `json:"is_favorite"`
}
