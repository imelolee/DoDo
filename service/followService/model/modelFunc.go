package model

import (
	"fmt"
	"log"
)

// GetFollowingCnt 给定当前用户id，查询follow表中该用户关注了多少人。
func GetFollowingCnt(userId int64) (int64, error) {
	// 用于存储当前用户关注了多少人。
	var cnt int64
	if Db == nil {
		InitDb()
	}
	// 查询出错，日志打印err msg，并return err
	err := Db.Model(Follow{}).Where("follower_id = ?", userId).Where("cancel = ?", 0).Count(&cnt)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error
	}
	// 查询成功，返回人数。
	return cnt, nil
}

// GetFollowerCnt 给定当前用户id，查询follow表中该用户的粉丝数。
func GetFollowerCnt(userId int64) (int64, error) {
	// 用于存储当前用户粉丝数的变量
	var cnt int64
	if Db == nil {
		InitDb()
	}
	// 当查询出现错误的情况，日志打印err msg，并返回err.
	if err := Db.
		Model(Follow{}).
		Where("user_id = ?", userId).
		Where("cancel = ?", 0).
		Count(&cnt).Error; nil != err {
		log.Println(err.Error())
		return 0, err
	}
	// 正常情况，返回取到的粉丝数。
	return cnt, nil
}

// FindRelation 给定当前用户和目标用户id，查询follow表中相应的记录。
func FindRelation(userId int64, targetId int64) (*Follow, error) {
	// follow变量用于后续存储数据库查出来的用户关系。
	follow := Follow{}
	//当查询出现错误时，日志打印err msg，并return err.
	if err := Db.
		Where("user_id = ?", targetId).
		Where("follower_id = ?", userId).
		Where("cancel = ?", 0).
		Take(&follow).Error; nil != err {
		// 当没查到数据时，gorm也会报错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		log.Println(err.Error())
		return nil, err
	}
	//正常情况，返回取到的值和空err.
	return &follow, nil
}

// GetFollowersIds 给定用户id，查询他关注了哪些人的id。
func GetFollowersIds(userId int64) ([]int64, error) {
	var ids []int64
	if err := Db.Model(Follow{}).
		Where("user_id = ?", userId).
		Where("cancel = ?", 0).
		Pluck("follower_id", &ids).Error; nil != err {
		// 没有粉丝，但是不能算错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		// 查询出错。
		log.Println(err.Error())
		return nil, err
	}
	// 查询成功。
	return ids, nil
}

// GetFollowingIds 给定用户id，查询他关注了哪些人的id。
func GetFollowingIds(userId int64) ([]int64, error) {
	var ids []int64
	if err := Db.Model(Follow{}).
		Where("follower_id = ?", userId).
		Pluck("user_id", &ids).Error; nil != err {
		// 没有关注任何人，但是不能算错。
		if "record not found" == err.Error() {
			return nil, nil
		}
		// 查询出错。
		log.Println(err.Error())
		return nil, err
	}
	// 查询成功。
	return ids, nil
}
