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

type PublishVideo struct {
	Id            int64                  `json:"id"`
	Author        *userModel.PublishUser `protobuf:"bytes,2,opt,name=author,proto3" json:"author,omitempty"`
	PlayUrl       string                 `json:"play_url"`
	CoverUrl      string                 `json:"cover_url"`
	FavoriteCount int64                  `json:"favorite_count"`
	CommentCount  int64                  `json:"comment_count"`
	IsFavorite    bool                   `json:"is_favorite"`
	Title         string                 `json:"title"`
}

type FeedVideo struct {
	Video
	Author        userModel.FeedUser `json:"author"`
	FavoriteCount int64              `json:"favorite_count"`
	CommentCount  int64              `json:"comment_count"`
	IsFavorite    bool               `json:"is_favorite"`
}
