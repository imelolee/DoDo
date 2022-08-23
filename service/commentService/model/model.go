package model

// Follow 用户关系结构，对应用户关系表。
type Follow struct {
	Id         int64
	UserId     int64
	FollowerId int64
	Cancel     int8
}
