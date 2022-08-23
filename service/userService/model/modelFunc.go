package model

import (
	"log"
	pb "userService/proto"
)

// GetTableUserList 获取全部TableUser对象
func GetTableUserList() ([]*pb.User, error) {
	var tableUsers []*pb.User
	if err := Db.Find(&tableUsers).Error; err != nil {
		log.Println(err.Error())
		return tableUsers, err
	}
	return tableUsers, nil
}

// GetTableUserByUsername 根据username获得TableUser对象
func GetTableUserByUsername(name string) (*pb.User, error) {
	tableUser := pb.User{}
	if err := Db.Where("name = ?", name).First(&tableUser).Error; err != nil {
		log.Println(err.Error())
		return &tableUser, err
	}
	return &tableUser, nil
}

// GetTableUserById 根据user_id获得TableUser对象
func GetTableUserById(id int64) (*pb.User, error) {
	tableUser := pb.User{}
	if err := Db.Where("id = ?", id).First(&tableUser).Error; err != nil {
		log.Println(err.Error())
		return &tableUser, err
	}
	return &tableUser, nil
}

// InsertTableUser 将tableUser插入表内
func InsertTableUser(tableUser *pb.User) bool {
	if err := Db.Create(&tableUser).Error; err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
