package model

import (
	"commentService/config"
	pb "commentService/proto"
	"errors"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	"sort"
	"strconv"
	"sync"
	"time"
	userModel "userService/model"
)

// 在redis中存储video_id对应的comment_id
func insertRedisVideoCommentId(videoId string, commentId string) {
	//在redis-RdbVCid中存储video_id对应的comment_id
	_, err := RdbVCid.SAdd(Ctx, videoId, commentId).Result()
	if err != nil { //若存储redis失败-有err，则直接删除key
		log.Infof("insertRedisVideoCommentId err: ", err)
		RdbVCid.Del(Ctx, videoId)
		return
	}
	//在redis-RdbCVid中存储comment_id对应的video_id
	_, err = RdbCVid.Set(Ctx, commentId, videoId, 0).Result()
	if err != nil {
		log.Infof("insertRedisVideoCommentId err: ", err)
	}
}

// 此函数用于给一个评论赋值：评论信息+用户信息 填充
func oneComment(comment *CommentInfo, com *Comment, userId int64) {
	var wg sync.WaitGroup
	wg.Add(1)
	//根据评论用户id和当前用户id，查询评论用户信息

	user, err := userModel.GetFeedUserByIdWithCurId(com.VideoId, com.UserId)
	if err != nil {
		log.Infof("userModel.GetFeedUserByIdWithCurId err:", err)
		return
	}
	var userInfo userModel.FeedUser
	gconv.Struct(user, &userInfo)

	comment.Id = com.Id
	comment.Content = com.CommentText
	comment.CreateDate = com.CreateDate.Format(config.DateTime)
	comment.UserInfo = userInfo
	if err != nil {
		log.Infof("oneComment err:", err) //函数返回提示错误信息
	}
	wg.Done()
	wg.Wait()
}

// Count 使用video id 查询Comment数量
func Count(videoId int64) (int64, error) {
	var count int64
	//数据库中查询评论数量
	err := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.ValidComment}).Count(&count).Error
	if err != nil {
		log.Infof("CommentModel.Count err:", err)
		return -1, err
	}
	return count, nil
}

//CommentIdList 根据视频id获取评论id 列表
func CommentIdList(videoId int64) ([]string, error) {
	var commentIdList []string
	err := Db.Model(Comment{}).Select("id").Where("video_id = ?", videoId).Find(&commentIdList).Error
	if err != nil {
		log.Infof("CommentModel.CommentIdList err:", err)
		return nil, err
	}
	return commentIdList, nil
}

// InsertComment 发表评论
func InsertComment(comment Comment) (Comment, error) {
	//数据库中插入一条评论信息
	err := Db.Model(Comment{}).Create(&comment).Error
	if err != nil {
		log.Infof("CommentModel.InsertComment err:", err)
		return Comment{}, err
	}
	return comment, nil
}

// DeleteComment 删除评论，传入评论id
func DeleteComment(id int64) error {

	var commentInfo Comment
	//先查询是否有此评论
	result := Db.Model(Comment{}).Where(map[string]interface{}{"id": id, "cancel": config.ValidComment}).First(&commentInfo)
	if result.RowsAffected == 0 { //查询到此评论数量为0则返回无此评论
		log.Infof("CommentModel.DeleteComment err: Comment not exist.")
		return errors.New("Comment not exist.")
	}
	//数据库中删除评论-更新评论状态为-1
	err := Db.Model(Comment{}).Where("id = ?", id).Update("cancel", config.InvalidComment).Error
	if err != nil {
		log.Infof("CommentModel.DeleteComment err:", err)
		return err
	}

	return nil
}

// GetCommentList 根据视频id查询所属评论全部列表信息
func GetCommentList(videoId int64) ([]Comment, error) {
	//数据库中查询评论信息list
	var commentList []Comment
	result := Db.Model(Comment{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.ValidComment}).
		Order("create_date desc").Find(&commentList)
	//若此视频没有评论信息，返回空列表，不报错
	if result.RowsAffected == 0 {
		log.Infof("CommentModel.GetCommentList err: No comment.") //函数返回提示无评论
		return nil, nil
	}
	//若获取评论列表出错
	if result.Error != nil {
		log.Infof("CommentModel.GetCommentList err:", result.Error)
		return commentList, result.Error
	}
	return commentList, nil
}

// CountFromVideoId 使用video id 查询Comment数量
func CountFromVideoId(id int64) (int64, error) {
	if RdbVCid == nil {
		InitRedis()
	}
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
		_, err := RdbVCid.SAdd(Ctx, strconv.Itoa(int(id)), config.DefaultRedisValue).Result()
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

// Send 发表评论
func Send(cmt *pb.Comment) (*pb.CommentInfo, error) {
	//数据准备
	var commentInfo Comment
	commentInfo.CreateDate = time.Now()
	commentInfo.VideoId = cmt.VideoId         //评论视频id传入
	commentInfo.UserId = cmt.UserId           //评论用户id传入
	commentInfo.CommentText = cmt.CommentText //评论内容传入
	commentInfo.Cancel = config.ValidComment  //评论状态，0，有效

	//1.评论信息存储：
	commentRtn, err := InsertComment(commentInfo)
	if err != nil {
		return &pb.CommentInfo{}, err
	}
	//2.查询用户信息

	user, err := userModel.GetFeedUserByIdWithCurId(cmt.UserId, cmt.UserId)
	if err != nil {
		log.Infof("userModel.GetFeedUserByIdWithCurId err:", err)
		return &pb.CommentInfo{}, err
	}

	var tmpUser userModel.FeedUser
	gconv.Struct(user, &tmpUser)

	//3.拼接
	commentData := CommentInfo{
		Id:         commentRtn.Id,
		UserInfo:   tmpUser,
		Content:    commentRtn.CommentText,
		CreateDate: commentRtn.CreateDate.Format(config.DateTime),
	}
	//将此发表的评论id存入redis
	go func() {
		insertRedisVideoCommentId(strconv.Itoa(int(cmt.VideoId)), strconv.Itoa(int(commentRtn.Id)))
	}()

	var comment pb.CommentInfo
	gconv.Struct(commentData, &comment)

	//返回结果
	return &comment, nil
}

// Delete 删除评论，传入评论id
func Delete(id int64) error {
	//1.先查询redis，若有则删除，返回客户端-再go协程删除数据库；无则在数据库中删除，返回客户端。
	n, err := RdbCVid.Exists(Ctx, strconv.FormatInt(id, 10)).Result()
	if err != nil {
		return err
	}
	if n > 0 { //在缓存中有此值，则找出来删除，然后返回
		vid, err := RdbCVid.Get(Ctx, strconv.FormatInt(id, 10)).Result()
		if err != nil { //没找到，返回err
			return err
		}
		//删除，两个redis都要删除
		_, err = RdbCVid.Del(Ctx, strconv.FormatInt(id, 10)).Result()
		if err != nil {
			return err
		}
		_, err = RdbVCid.SRem(Ctx, vid, strconv.FormatInt(id, 10)).Result()
		if err != nil {
			return err
		}

		//使用mq进行数据库中评论的删除-评论状态更新
		//评论id传入消息队列
		RmqCommentDel.Publish(strconv.FormatInt(id, 10))
		return nil
	}
	//不在内存中，则直接走数据库删除
	err = DeleteComment(id)
	if err != nil {
		return err
	}
	return nil
}

// GetList 查看评论列表-返回评论list
func GetList(videoId int64, userId int64) ([]*pb.CommentInfo, error) {
	//1.先查询评论列表信息
	commentList, err := GetCommentList(videoId)
	if err != nil {
		return nil, err
	}
	//当前有0条评论
	if commentList == nil {
		return nil, nil
	}

	//提前定义好切片长度
	commentInfoList := make([]CommentInfo, len(commentList))

	wg := &sync.WaitGroup{}
	wg.Add(len(commentList))
	idx := 0
	for _, comment := range commentList {
		//2.调用方法组装评论信息，再append
		var commentData CommentInfo
		//将评论信息进行组装，添加想要的信息,插入从数据库中查到的数据
		go func(comment Comment) {
			oneComment(&commentData, &comment, userId)
			//3.组装list
			//commentInfoList = append(commentInfoList, commentData)
			commentInfoList[idx] = commentData
			idx = idx + 1
			wg.Done()
		}(comment)
	}
	wg.Wait()
	//评论排序-按照主键排序
	sort.Sort(CommentSlice(commentInfoList))

	//协程查询redis中是否有此记录，无则将评论id切片存入redis
	go func() {
		//1.先在缓存中查此视频是否已有评论列表
		cnt, err := RdbVCid.SCard(Ctx, strconv.FormatInt(videoId, 10)).Result()
		if err != nil { //若查询缓存出错，则打印log
			return
		}
		//2.缓存中查到了数量大于0，则说明数据正常，不用更新缓存
		if cnt > 0 {
			return
		}
		//3.缓存中数据不正确，更新缓存：
		//先在redis中存储一个-1 值，防止脏读
		_, err = RdbVCid.SAdd(Ctx, strconv.Itoa(int(videoId)), config.DefaultRedisValue).Result()
		if err != nil { //若存储redis失败，则直接返回
			return
		}
		//设置key值过期时间
		_, err = RdbVCid.Expire(Ctx, strconv.Itoa(int(videoId)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			return
		}
		//将评论id循环存入redis
		for _, _comment := range commentInfoList {
			insertRedisVideoCommentId(strconv.Itoa(int(videoId)), strconv.Itoa(int(_comment.Id)))
		}
	}()

	var comInfoList []*pb.CommentInfo
	gconv.Struct(commentInfoList, &comInfoList)

	return comInfoList, nil
}
