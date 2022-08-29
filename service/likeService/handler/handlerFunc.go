package handler

import (
	"context"
	"github.com/go-micro/plugins/v4/registry/consul"
	"github.com/gogf/gf/util/gconv"
	"go-micro.dev/v4"
	pb "likeService/proto"
	"log"
	"sync"
	videoService "videoService/proto"
)

// InitMicro 初始化微服务
func InitMicro() micro.Service {
	consulReg := consul.NewRegistry()
	return micro.NewService(micro.Registry(consulReg))
}

// 根据videoId,登录用户curId，添加视频对象到点赞列表空间
func addFavouriteVideoList(videoId int64, curId int64, favoriteVideoList *[]*pb.Video, wg *sync.WaitGroup) {

	defer wg.Done()
	//调用videoService接口，GetVideo：根据videoId，当前用户id:curId，返回Video类型对象

	videoMicro := InitMicro()
	videoClient := videoService.NewVideoService("videoService", videoMicro.Client())

	rsp, _ := videoClient.GetVideo(context.TODO(), &videoService.GetVideoReq{
		VideoId: videoId,
		UserId:  curId,
	})
	var video pb.Video
	err := gconv.Struct(rsp.Video, &video)
	if err != nil {
		log.Printf("类型转换失败:", err)
	}

	*favoriteVideoList = append(*favoriteVideoList, &video)
}

// 根据videoId，将该视频点赞数加入对应提前开辟好的空间内
func addVideoLikeCount(videoId int64, videoLikeCountList *[]int64, wg *sync.WaitGroup) {
	defer wg.Done()
	//调用FavouriteCount：根据videoId,获取点赞数

	likeMicro := InitMicro()
	likeClient := pb.NewLikeService("likeService", likeMicro.Client())
	rsp, err := likeClient.FavouriteCount(context.TODO(), &pb.IdReq{
		Id: videoId,
	})
	if err != nil {
		//如果有错误，输出错误信息，并不加入该视频点赞数
		log.Printf(err.Error())
		return
	}
	*videoLikeCountList = append(*videoLikeCountList, rsp.Count)
}
