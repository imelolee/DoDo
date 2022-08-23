package model

import (
	"time"
	userModel "userService/model"
)

// Comment 评论信息-数据库中的结构体-dao层使用
type Comment struct {
	Id          int64     //评论id
	UserId      int64     //评论用户id
	VideoId     int64     //视频id
	CommentText string    //评论内容
	CreateDate  time.Time //评论发布的日期mm-dd
	Cancel      int32     //取消评论为1，发布评论为0
}

// CommentInfo 查看评论-传出的结构体-service
type CommentInfo struct {
	Id         int64              `json:"id,omitempty"`
	UserInfo   userModel.FeedUser `json:"user,omitempty"`
	Content    string             `json:"content,omitempty"`
	CreateDate string             `json:"create_date,omitempty"`
}

type CommentSlice []CommentInfo

func (a CommentSlice) Len() int { //重写Len()方法
	return len(a)
}
func (a CommentSlice) Swap(i, j int) { //重写Swap()方法
	a[i], a[j] = a[j], a[i]
}
func (a CommentSlice) Less(i, j int) bool { //重写Less()方法
	return a[i].Id > a[j].Id
}
