package model

// User 对应数据库User表结构的结构体
type User struct {
	Id       int64
	Name     string
	Password string
}

// FeedUser 最终封装后,controller返回的User结构体
type FeedUser struct {
	Id             int64  `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	FollowCount    int64  `json:"follow_count"`
	FollowerCount  int64  `json:"follower_count"`
	IsFollow       bool   `json:"is_follow"`
	TotalFavorited int64  `json:"total_favorited,omitempty"`
	FavoriteCount  int64  `json:"favorite_count,omitempty"`
}
