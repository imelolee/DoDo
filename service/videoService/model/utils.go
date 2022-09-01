package model

import (
	"context"
	"github.com/gogf/gf/util/gconv"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/sms/bytes"
	"github.com/qiniu/go-sdk/v7/storage"
	log "go-micro.dev/v4/logger"
	"sync"
	userModel "userService/model"
	"videoService/config"
)

// 该方法可以将数据进行拷贝和转换，并从其他方法获取对应的数据
func copyVideos(result *[]FeedVideo, data *[]Video, userId int64) error {
	for _, temp := range *data {
		var video FeedVideo
		//将video进行组装，添加想要的信息,插入从数据库中查到的数据
		creatVideo(&video, &temp, userId)
		*result = append(*result, video)
	}
	return nil
}

//将video进行组装，添加想要的信息,插入从数据库中查到的数据
func creatVideo(video *FeedVideo, data *Video, userId int64) {
	//建立协程组，当这一组的携程全部完成后，才会结束本方法
	var wg sync.WaitGroup
	wg.Add(4)
	video.Video = *data
	//插入Author，这里需要将视频的发布者和当前登录的用户传入，才能正确获得isFollow，
	//如果出现错误，不能直接返回失败，将默认值返回，保证稳定

	go func() {
		user, err := userModel.GetFeedUserByIdWithCurId(data.AuthorId, userId)

		if err != nil {
			log.Infof("userModel.GetFeedUserByIdWithCurId err:", err)
		}

		var tmpUser *userModel.FeedUser
		gconv.Struct(user, &tmpUser)

		video.Author = *tmpUser
		wg.Done()
	}()

	//插入点赞数量，同上所示，不将nil直接向上返回，数据没有就算了，给一个默认就行了
	go func() {
		likeCount, err := FavouriteCount(data.Id)
		if err != nil {
			log.Infof("likeModel.FavouriteCount err:", err)
		}
		video.FavoriteCount = likeCount
		wg.Done()
	}()

	//获取该视屏的评论数字
	go func() {
		commentCount, err := CountFromVideoId(data.Id)
		if err != nil {
			log.Infof("commentModel.CountFromVideoId err:", err)
		}
		video.CommentCount = commentCount
		wg.Done()
	}()

	//获取当前用户是否点赞了该视频
	go func() {
		like, err := IsFavorite(video.Id, userId)
		if err != nil {
			log.Infof(" likeModel.IsFavorite:", err)
		}

		video.IsFavorite = like
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
