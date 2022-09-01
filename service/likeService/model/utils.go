package model

import (
	"fmt"
	"github.com/gogf/gf/util/gconv"
	log "go-micro.dev/v4/logger"
	pb "likeService/proto"
	"sync"
	videoModel "videoService/model"
)

// 根据videoId,登录用户curId，添加视频对象到点赞列表空间
func addFavouriteVideoList(videoId int64, curId int64, favoriteVideoList *[]*pb.Video, wg *sync.WaitGroup) {

	defer wg.Done()
	//调用videoService接口，GetVideo：根据videoId，当前用户id:curId，返回Video类型对象
	video, err := videoModel.GetVideo(videoId, curId)
	if err != nil {
		log.Infof("videoModel.GetVideo err:", err)
	}
	var tmpVideo *pb.Video
	gconv.Struct(video, &tmpVideo)

	*favoriteVideoList = append(*favoriteVideoList, tmpVideo)
}

// 根据videoId，将该视频点赞数加入对应提前开辟好的空间内
func addVideoLikeCount(videoId int64, videoLikeCountList *[]int64, wg *sync.WaitGroup) {
	defer wg.Done()
	//调用FavouriteCount：根据videoId,获取点赞数

	count, err := FavouriteCount(videoId)
	if err != nil {
		fmt.Println("likeModel.FavouriteCount err:", err)
		return
	}
	*videoLikeCountList = append(*videoLikeCountList, count)
}
