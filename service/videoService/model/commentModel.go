package model

import (
	"errors"
	"log"
	"strconv"
	"time"
	"videoService/config"
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

// Count 使用video id 查询Comment数量
func Count(videoId int64) (int64, error) {
	log.Println("CommentDao-Count: running") //函数已运行
	//Init()
	var count int64
	//数据库中查询评论数量
	err := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": videoId, "cancel": 0}).Count(&count).Error
	if err != nil {
		log.Println("CommentDao-Count: return count failed") //函数返回提示错误信息
		return -1, errors.New("find comments count failed")
	}
	log.Println("CommentDao-Count: return count success") //函数执行成功，返回正确信息
	return count, nil
}

//CommentIdList 根据视频id获取评论id 列表
func CommentIdList(videoId int64) ([]string, error) {
	var commentIdList []string
	err := Db.Model(Comment{}).Select("id").Where("video_id = ?", videoId).Find(&commentIdList).Error
	if err != nil {
		log.Println("CommentIdList:", err)
		return nil, err
	}
	return commentIdList, nil
}

// CountFromVideoId 使用video id 查询Comment数量
func CountFromVideoId(id int64) (int64, error) {
	//先在缓存中查
	cnt, err := RdbVCid.SCard(Ctx, strconv.FormatInt(id, 10)).Result()
	if err != nil { //若查询缓存出错，则打印log
		return 0, err
	}
	//1.缓存中查到了数量，则返回数量值-1（去除0值）
	if cnt != 0 {
		return cnt - 1, nil
	}
	//2.缓存中查不到则去数据库查
	cnt, err = Count(id)
	if err != nil {
		return 0, err
	}
	//将评论id切片存入redis-第一次存储 V-C set 值：
	go func() {
		//查询评论id list
		cList, _ := CommentIdList(id)
		//先在redis中存储一个-1值，防止脏读
		_, err := RdbVCid.SAdd(Ctx, strconv.Itoa(int(id)), -1).Result()
		if err != nil { //若存储redis失败，则直接返回
			return
		}
		//设置key值过期时间
		_, err = RdbVCid.Expire(Ctx, strconv.Itoa(int(id)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			return
		}
		//评论id循环存入redis
		for _, commentId := range cList {
			insertRedisVideoCommentId(strconv.Itoa(int(id)), commentId)
		}
	}()
	return cnt, err
}

// 在redis中存储video_id对应的comment_id
func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := RdbVCid.SAdd(Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-有err，则直接删除key
		log.Println("redis save send: vId - cId failed, key deleted")
		RdbVCid.Del(Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = RdbCVid.Set(Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Println("Redis save cId - vId failed.")
	}
}
