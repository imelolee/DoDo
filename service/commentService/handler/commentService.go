package handler

import (
	"commentService/config"
	"commentService/model"
	pb "commentService/proto"
	"context"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	"sort"
	"strconv"
	"sync"
	"time"
	userModel "userService/model"
	userService "userService/proto"
)

type CommentService struct{}

// CountFromVideoId 使用video id 查询Comment数量
func (e *CommentService) CountFromVideoId(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received CommentService.CountFromVideoId request: %v", req)
	//先在缓存中查
	cnt, err := model.RdbVCid.SCard(model.Ctx, strconv.FormatInt(req.Id, 10)).Result()
	if err != nil { //若查询缓存出错，则打印log
		rsp.Count = 0
		return err
	}
	//1.缓存中查到了数量，则返回数量值-1（去除0值）
	if cnt != 0 {
		rsp.Count = cnt - 1
		return nil
	}
	//2.缓存中查不到则去数据库查
	cnt, err = model.Count(req.Id)
	if err != nil {
		rsp.Count = 0
		return err
	}
	//将评论id切片存入redis-第一次存储 V-C set 值：
	go func() {
		//查询评论id list
		cList, _ := model.CommentIdList(req.Id)
		//先在redis中存储一个-1值，防止脏读
		_, err := model.RdbVCid.SAdd(model.Ctx, strconv.Itoa(int(req.Id)), config.DefaultRedisValue).Result()
		if err != nil { //若存储redis失败，则直接返回
			return
		}
		//设置key值过期时间
		_, err = model.RdbVCid.Expire(model.Ctx, strconv.Itoa(int(req.Id)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			return
		}
		//评论id循环存入redis
		for _, commentId := range cList {
			insertRedisVideoCommentId(strconv.Itoa(int(req.Id)), commentId)
		}
	}()
	//返回结果
	return nil
}

// Send 发表评论
func (e *CommentService) Send(ctx context.Context, req *pb.CommentReq, rsp *pb.CommentRsp) error {
	log.Infof("Received CommentService.Send request: %v", req)
	//数据准备
	var commentInfo model.Comment
	commentInfo.VideoId = req.Comment.VideoId                                       //评论视频id传入
	commentInfo.UserId = req.Comment.UserId                                         //评论用户id传入
	commentInfo.CommentText = req.Comment.CommentText                               //评论内容传入
	commentInfo.Cancel = config.ValidComment                                        //评论状态，0，有效
	commentInfo.CreateDate, _ = time.Parse(config.DateTime, req.Comment.CreateDate) //评论时间

	//1.评论信息存储：
	commentRtn, err := model.InsertComment(commentInfo)
	if err != nil {
		rsp.CommentInfo = &pb.CommentInfo{}
		return err
	}
	//2.查询用户信息
	userMicro := InitMicro()
	userClient := userService.NewUserService("userService", userMicro.Client())

	userRsp, _ := userClient.GetFeedUserByIdWithCurId(context.TODO(), &userService.CurIdReq{
		Id:    req.Comment.VideoId,
		CurId: req.Comment.UserId,
	})

	if err != nil {
		rsp.CommentInfo = &pb.CommentInfo{}
		return err
	}

	var user userModel.FeedUser
	gconv.Struct(userRsp.User, &user)

	//3.拼接
	commentData := model.CommentInfo{
		Id:         commentRtn.Id,
		UserInfo:   user,
		Content:    commentRtn.CommentText,
		CreateDate: commentRtn.CreateDate.Format(config.DateTime),
	}
	//将此发表的评论id存入redis
	go func() {
		insertRedisVideoCommentId(strconv.Itoa(int(req.Comment.VideoId)), strconv.Itoa(int(commentRtn.Id)))
	}()

	var comment pb.CommentInfo
	gconv.Struct(commentData, &comment)

	//返回结果
	rsp.CommentInfo = &comment
	return nil
}

// Delete 删除评论，传入评论id
func (e *CommentService) Delete(ctx context.Context, req *pb.IdReq, rsp *pb.DelRsp) error {
	log.Infof("Received CommentService.DelComment request: %v", req)
	//1.先查询redis，若有则删除，返回客户端-再go协程删除数据库；无则在数据库中删除，返回客户端。
	n, err := model.RdbCVid.Exists(model.Ctx, strconv.FormatInt(req.Id, 10)).Result()
	if err != nil {
		return err
	}
	if n > 0 { //在缓存中有此值，则找出来删除，然后返回
		vid, err := model.RdbCVid.Get(model.Ctx, strconv.FormatInt(req.Id, 10)).Result()
		if err != nil { //没找到，返回err
			return err
		}
		//删除，两个redis都要删除
		_, err = model.RdbCVid.Del(model.Ctx, strconv.FormatInt(req.Id, 10)).Result()
		if err != nil {
			return err
		}
		_, err = model.RdbVCid.SRem(model.Ctx, vid, strconv.FormatInt(req.Id, 10)).Result()
		if err != nil {
			return err
		}

		//使用mq进行数据库中评论的删除-评论状态更新
		//评论id传入消息队列
		model.RmqCommentDel.Publish(strconv.FormatInt(req.Id, 10))
		return nil
	}
	//不在内存中，则直接走数据库删除
	err = model.DeleteComment(req.Id)
	if err != nil {
		return err
	}
	return nil
}

// GetList 查看评论列表-返回评论list
func (e *CommentService) GetList(ctx context.Context, req *pb.VideoUserReq, rsp *pb.CommentListRsp) error {
	log.Infof("Received CommentService.GetList request: %v", req)
	//1.先查询评论列表信息
	commentList, err := model.GetCommentList(req.VideoId)
	if err != nil {
		rsp.CommentInfo = nil
		return err
	}
	//当前有0条评论
	if commentList == nil {
		rsp.CommentInfo = nil
		return nil
	}

	//提前定义好切片长度
	commentInfoList := make([]model.CommentInfo, len(commentList))

	wg := &sync.WaitGroup{}
	wg.Add(len(commentList))
	idx := 0
	for _, comment := range commentList {
		//2.调用方法组装评论信息，再append
		var commentData model.CommentInfo
		//将评论信息进行组装，添加想要的信息,插入从数据库中查到的数据
		go func(comment model.Comment) {
			oneComment(&commentData, &comment, req.UserId)
			//3.组装list
			//commentInfoList = append(commentInfoList, commentData)
			commentInfoList[idx] = commentData
			idx = idx + 1
			wg.Done()
		}(comment)
	}
	wg.Wait()
	//评论排序-按照主键排序
	sort.Sort(model.CommentSlice(commentInfoList))

	//协程查询redis中是否有此记录，无则将评论id切片存入redis
	go func() {
		//1.先在缓存中查此视频是否已有评论列表
		cnt, err := model.RdbVCid.SCard(model.Ctx, strconv.FormatInt(req.VideoId, 10)).Result()
		if err != nil { //若查询缓存出错，则打印log
			return
		}
		//2.缓存中查到了数量大于0，则说明数据正常，不用更新缓存
		if cnt > 0 {
			return
		}
		//3.缓存中数据不正确，更新缓存：
		//先在redis中存储一个-1 值，防止脏读
		_, err = model.RdbVCid.SAdd(model.Ctx, strconv.Itoa(int(req.VideoId)), config.DefaultRedisValue).Result()
		if err != nil { //若存储redis失败，则直接返回
			return
		}
		//设置key值过期时间
		_, err = model.RdbVCid.Expire(model.Ctx, strconv.Itoa(int(req.VideoId)),
			time.Duration(config.OneMonth)*time.Second).Result()
		if err != nil {
			return
		}
		//将评论id循环存入redis
		for _, _comment := range commentInfoList {
			insertRedisVideoCommentId(strconv.Itoa(int(req.VideoId)), strconv.Itoa(int(_comment.Id)))
		}
	}()

	var comInfoList []*pb.CommentInfo
	err = gconv.Struct(commentInfoList, &comInfoList)
	if err != nil {
		return err
	}

	rsp.CommentInfo = comInfoList
	return nil
}
