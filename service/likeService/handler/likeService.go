package handler

import (
	"context"
	log "go-micro.dev/v4/logger"
	"likeService/model"
	pb "likeService/proto"
)

type LikeService struct{}

// IsFavorite 根据userId,videoId查询点赞状态 这边可以快一点,通过查询两个Redis DB;
func (e *LikeService) IsFavorite(ctx context.Context, req *pb.VideoUserReq, rsp *pb.BoolRsp) error {
	log.Infof("Received LikeService.IsFavorite request: %v", req)
	exist, err := model.IsFavorite(req.VideoId, req.UserId)
	rsp.Flag = exist
	return err
}

//FavouriteCount 根据videoId获取对应点赞数量;
func (e *LikeService) FavouriteCount(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.FavouriteCount request: %v", req)
	count, err := model.FavouriteCount(req.Id)
	rsp.Count = count
	return err
}

// FavouriteAction 根据userId，videoId,actionType对视频进行点赞或者取消赞操作;
func (e *LikeService) FavouriteAction(ctx context.Context, req *pb.ActionReq, rsp *pb.ActionRsp) error {
	log.Infof("Received LikeService.FavouriteAction request: %v", req)
	err := model.FavouriteAction(req.VideoId, req.UserId, req.ActionType)
	return err

}

//GetFavouriteList 根据userId，curId(当前用户Id),返回userId的点赞列表;
func (e *LikeService) GetFavouriteList(ctx context.Context, req *pb.UserCurReq, rsp *pb.FavouriteListRsp) error {
	log.Infof("Received LikeService.GetFavouriteList request: %v", req)
	favouriteList, err := model.GetFavouriteList(req.UserId, req.CurId)
	rsp.Video = favouriteList
	return err
}

//TotalFavourite 根据userId获取这个用户总共被点赞数量
func (e *LikeService) TotalFavourite(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.TotalFavourite request: %v", req)
	sum, err := model.TotalFavourite(req.Id)
	rsp.Count = sum
	return err
}

//FavouriteVideoCount 根据userId获取这个用户点赞视频数量
func (e *LikeService) FavouriteVideoCount(ctx context.Context, req *pb.IdReq, rsp *pb.CountRsp) error {
	log.Infof("Received LikeService.FavouriteVideoCount request: %v", req)
	count, err := model.FavouriteVideoCount(req.Id)
	rsp.Count = count
	return err
}
