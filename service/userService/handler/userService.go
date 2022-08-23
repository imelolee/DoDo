package handler

import (
	"context"
	followModel "followService/model"
	"log"
	"userService/model"
	pb "userService/proto"
)

type UserService struct{}

func (e *UserService) GetTableUserList(ctx context.Context, req *pb.Req, rsp *pb.UserListRsp) error {
	log.Printf("Received UserService.GetTableUserList request: %v\n", req)
	tableUsers, err := model.GetTableUserList()
	if err != nil {
		log.Printf("Err:", err.Error())
	}
	rsp.User = tableUsers
	return nil
}

func (e *UserService) GetTableUserByUsername(ctx context.Context, req *pb.UsernameReq, rsp *pb.UserRsp) error {
	log.Printf("Received UserService.GetTableUserByUsername request: %v\n", req)
	tableUser, err := model.GetTableUserByUsername(req.Name)
	if err != nil {
		log.Printf("Err:", err.Error())
	}
	rsp.User = tableUser
	return nil
}

func (e *UserService) GetTableUserById(ctx context.Context, req *pb.IdReq, rsp *pb.UserRsp) error {
	log.Printf("Received UserService.GetTableUserById request: %v\n", req)
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		log.Printf("Err:", err.Error())
	}
	rsp.User = tableUser
	return nil
}

func (e *UserService) InsertTableUser(ctx context.Context, req *pb.UserReq, rsp *pb.BoolRsp) error {
	log.Printf("Received UserService.GetTableUserById request: %v\n", req)
	model.InsertTableUser(req.User)

	rsp.Flag = true
	return nil
}

// GetFeedUserById 未登录情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserById(ctx context.Context, req *pb.IdReq, rsp *pb.FeedUserRsp) error {
	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		rsp.User = &user
	}
	log.Println("Query User Success")
	followCount, _ := followModel.GetFollowingCnt(req.Id)
	if err != nil {
		log.Println("Err:", err.Error())
	}
	followerCount, _ := followModel.GetFollowerCnt(req.Id)
	if err != nil {
		log.Println("Err:", err.Error())
	}
	u := GetLikeService() //解决循环依赖
	totalFavorited, _ := u.TotalFavourite(id)
	favoritedCount, _ := u.FavouriteVideoCount(id)
	user = User{
		Id:             id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       false,
		TotalFavorited: totalFavorited,
		FavoriteCount:  favoritedCount,
	}
	return user, nil
}

// GetUserByIdWithCurId 已登录(curID)情况下,根据user_id获得User对象
func (e *UserService) GetFeedUserByIdWithCurId(ctx context.Context, req *pb.CurIdReq, rsp *pb.FeedUserRsp) error {
	user := pb.FeedUser{
		Id:             0,
		Name:           "",
		FollowCount:    0,
		FollowerCount:  0,
		IsFollow:       false,
		TotalFavorited: 0,
		FavoriteCount:  0,
	}
	tableUser, err := model.GetTableUserById(req.Id)
	if err != nil {
		log.Println("Err:", err.Error())
		log.Println("User Not Found")
		rsp.User = &user
	}
	log.Println("Query User Success")
	followCount, err := followModel.GetFollowingCnt(id)
	if err != nil {
		log.Println("Err:", err.Error())
	}
	followerCount, err := usi.GetFollowerCnt(id)
	if err != nil {
		log.Println("Err:", err.Error())
	}
	isfollow, err := usi.IsFollowing(curId, id)
	if err != nil {
		log.Println("Err:", err.Error())
	}
	u := GetLikeService() //解决循环依赖
	totalFavorited, _ := u.TotalFavourite(id)
	favoritedCount, _ := u.FavouriteVideoCount(id)
	user = User{
		Id:             id,
		Name:           tableUser.Name,
		FollowCount:    followCount,
		FollowerCount:  followerCount,
		IsFollow:       isfollow,
		TotalFavorited: totalFavorited,
		FavoriteCount:  favoritedCount,
	}
	return user, nil
}
