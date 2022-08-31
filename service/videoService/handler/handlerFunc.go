package handler

import (
	commentService "commentService/proto"
	"context"
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/gogf/gf/util/gconv"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/sms/bytes"
	"github.com/qiniu/go-sdk/v7/storage"
	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
	likeService "likeService/proto"
	"sync"
	userModel "userService/model"
	userService "userService/proto"
	"videoService/config"
	"videoService/model"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	consulReg := consul.NewRegistry()
	return micro.NewService(micro.Registry(consulReg))
}

// 该方法可以将数据进行拷贝和转换，并从其他方法获取对应的数据
func copyVideos(result *[]model.FeedVideo, data *[]model.Video, userId int64) error {
	for _, temp := range *data {
		var video model.FeedVideo
		//将video进行组装，添加想要的信息,插入从数据库中查到的数据
		creatVideo(&video, &temp, userId)
		*result = append(*result, video)
	}
	return nil
}

//将video进行组装，添加想要的信息,插入从数据库中查到的数据
func creatVideo(video *model.FeedVideo, data *model.Video, userId int64) {
	//建立协程组，当这一组的携程全部完成后，才会结束本方法
	var wg sync.WaitGroup
	wg.Add(4)
	video.Video = *data
	//插入Author，这里需要将视频的发布者和当前登录的用户传入，才能正确获得isFollow，
	//如果出现错误，不能直接返回失败，将默认值返回，保证稳定

	go func() {
		userMicro := InitMicro()
		userClient := userService.NewUserService("userService", userMicro.Client())
		userRsp, err := userClient.GetFeedUserByIdWithCurId(context.TODO(), &userService.CurIdReq{
			Id:    data.AuthorId,
			CurId: userId,
		}, config.Opts)

		if err != nil {
			log.Infof("userClient.GetFeedUserByIdWithCurId err:", err)
		}

		var tmpUser *userModel.FeedUser
		gconv.Struct(userRsp.User, &tmpUser)

		video.Author = *tmpUser
		wg.Done()
	}()

	//插入点赞数量，同上所示，不将nil直接向上返回，数据没有就算了，给一个默认就行了
	go func() {
		likeMicro := InitMicro()
		likeClient := likeService.NewLikeService("likeService", likeMicro.Client())
		likeRsp, _ := likeClient.FavouriteCount(context.TODO(), &likeService.IdReq{
			Id: data.Id,
		}, config.Opts)
		video.FavoriteCount = likeRsp.Count
		wg.Done()
	}()

	//获取该视屏的评论数字
	go func() {
		commentMicro := InitMicro()
		commentClient := commentService.NewCommentService("commentService", commentMicro.Client())
		commentRsp, _ := commentClient.CountFromVideoId(context.TODO(), &commentService.IdReq{
			Id: data.Id,
		}, config.Opts)
		video.CommentCount = commentRsp.Count
		wg.Done()
	}()

	//获取当前用户是否点赞了该视频
	go func() {
		likeMicro := InitMicro()
		likeClient := likeService.NewLikeService("likeService", likeMicro.Client())
		likeRsp, err := likeClient.IsFavorite(context.TODO(), &likeService.VideoUserReq{
			VideoId: video.Id,
			UserId:  userId,
		}, config.Opts)
		if err != nil {
			log.Infof("likeClient.IsFavorite err:", err)
		}

		video.IsFavorite = likeRsp.Flag
		wg.Done()
	}()

	wg.Wait()
}

// 七牛云上传
func uploadQiniu(file []byte, fileName string, fileSize int64) error {

	putPolicy := storage.PutPolicy{
		Scope: config.Bucket,
	}
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{
		Zone:          &storage.ZoneHuabei,
		UseCdnDomains: false,
		UseHTTPS:      false,
	}

	putExtra := storage.PutExtra{}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	fileReader := bytes.NewReader(file)
	err := formUploader.Put(context.Background(), &ret, upToken, fileName, fileReader, fileSize, &putExtra)

	return err
}
