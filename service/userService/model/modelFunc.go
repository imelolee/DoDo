package model

import (
	"fmt"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	pb "userService/proto"
)

// GetTableUserList 获取全部TableUser对象
func GetTableUserList() ([]*pb.User, error) {
	var tableUsers []*pb.User
	if err := Db.Find(&tableUsers).Error; err != nil {
		log.Infof(err.Error())
		return tableUsers, err
	}
	return tableUsers, nil
}

// GetTableUserByUsername 根据username获得TableUser对象
func GetTableUserByUsername(name string) (*pb.User, error) {
	tableUser := pb.User{}
	if err := Db.Where("name = ?", name).First(&tableUser).Error; err != nil {
		log.Infof(err.Error())
		return &tableUser, err
	}
	return &tableUser, nil
}

// GetTableUserById 根据user_id获得TableUser对象
func GetTableUserById(id int64) (*pb.User, error) {
	tableUser := pb.User{}
	if Db == nil {
		InitDb()
	}
	err := Db.Where("id = ?", id).First(&tableUser).Error
	if err != nil {
		fmt.Println(err.Error())
		return &tableUser, err
	}
	return &tableUser, nil
}

// InsertTableUser 将tableUser插入表内
func InsertTableUser(tableUser *pb.User) bool {
	if err := Db.Create(&tableUser).Error; err != nil {
		log.Infof(err.Error())
		return false
	}
	return true
}

// GetFeedUserById 未登录情况下,根据user_id获得User对象
func GetFeedUserById(id int64) (*pb.FeedUser, error) {
	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := GetTableUserById(id)
	if err != nil {
		return &user, err
	}

	followingCnt, err := GetFollowingCnt(id)
	if err != nil {
		return nil, err
	}
	followerCnt, err := GetFollowerCnt(id)
	if err != nil {
		return nil, err
	}

	totalCnt, err := TotalFavourite(id)
	if err != nil {
		return nil, err
	}
	favCnt, err := FavouriteVideoCount(id)
	if err != nil {
		return nil, err
	}

	feedUser := FeedUser{
		Id:             id,
		Name:           tableUser.Name,
		FollowCount:    followingCnt,
		FollowerCount:  followerCnt,
		IsFollow:       false,
		TotalFavorited: totalCnt,
		FavoriteCount:  favCnt,
	}

	var tmpUser *pb.FeedUser
	tmpUser = new(pb.FeedUser)
	err = gconv.Struct(feedUser, &tmpUser)

	return tmpUser, nil
}

// GetFeedUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func GetFeedUserByIdWithCurId(id int64, curId int64) (*pb.FeedUser, error) {
	if Db == nil {
		InitDb()
	}
	if RdbFollowing == nil {
		InitRedis()
	}

	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := GetTableUserById(id)
	if err != nil {
		return &user, err
	}
	followCount, err := GetFollowingCnt(id)
	if err != nil {
		return &user, err
	}
	followerCount, err := GetFollowerCnt(id)
	if err != nil {
		return &user, err
	}

	isFollowing, err := IsFollowing(curId, id)
	if err != nil {
		return &user, err
	}

	total, err := TotalFavourite(id)
	if err != nil {
		return &user, err
	}

	count, err := FavouriteVideoCount(id)
	if err != nil {
		return &user, err
	}
	tmpUser := FeedUser{
		Id:             id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       isFollowing,
		TotalFavorited: total,
		FavoriteCount:  count,
	}

	var feedUser *pb.FeedUser
	err = gconv.Struct(tmpUser, &feedUser)
	return feedUser, nil
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
