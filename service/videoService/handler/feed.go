package handler

import (
	"log"
	"sync"
	userService "userService/proto"
	"videoService/model"
)

var service VideoService{}

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
	var err error
	video.Video = *data

	microService := InitMicro()
	microClient := userService.NewUserService("userService", microService.Client())


	//插入Author，这里需要将视频的发布者和当前登录的用户传入，才能正确获得isFollow，
	//如果出现错误，不能直接返回失败，将默认值返回，保证稳定
	go func() {
		video.Author, err = microClient.GetUserByIdWithCurId(data.AuthorId, userId)
		if err != nil {
			log.Printf(err.Error())
		}

		if err != nil {
			log.Printf("方法videoService.GetUserByIdWithCurId(data.AuthorId, userId) 失败：%v", err)
		} else {
			log.Printf("方法videoService.GetUserByIdWithCurId(data.AuthorId, userId) 成功")
		}
		wg.Done()
	}()

	//插入点赞数量，同上所示，不将nil直接向上返回，数据没有就算了，给一个默认就行了
	go func() {
		video.FavoriteCount, err = videoService.FavouriteCount(data.Id)
		if err != nil {
			log.Printf("方法videoService.FavouriteCount(data.ID) 失败：%v", err)
		} else {
			log.Printf("方法videoService.FavouriteCount(data.ID) 成功")
		}
		wg.Done()
	}()

	//获取该视屏的评论数字
	go func() {
		video.CommentCount, err = videoService.CountFromVideoId(data.Id)
		if err != nil {
			log.Printf("方法videoService.CountFromVideoId(data.ID) 失败：%v", err)
		} else {
			log.Printf("方法videoService.CountFromVideoId(data.ID) 成功")
		}
		wg.Done()
	}()

	//获取当前用户是否点赞了该视频
	go func() {
		video.IsFavorite, err = videoService.IsFavourite(video.Id, userId)
		if err != nil {
			log.Printf("方法videoService.IsFavourit(video.Id, userId) 失败：%v", err)
		} else {
			log.Printf("方法videoService.IsFavourit(video.Id, userId) 成功")
		}
		wg.Done()
	}()

	wg.Wait()
}