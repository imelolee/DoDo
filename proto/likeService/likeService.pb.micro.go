// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto/likeService.proto

package likeService

import (
	fmt "fmt"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "go-micro.dev/v4/api"
	client "go-micro.dev/v4/client"
	server "go-micro.dev/v4/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for LikeService service

func NewLikeServiceEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for LikeService service

type LikeService interface {
	IsFavorite(ctx context.Context, in *VideoUserReq, opts ...client.CallOption) (*BoolRsp, error)
	FavouriteCount(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error)
	TotalFavourite(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error)
	FavouriteVideoCount(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error)
	FavouriteAction(ctx context.Context, in *ActionReq, opts ...client.CallOption) (*ActionRsp, error)
	GetFavouriteList(ctx context.Context, in *UserCurReq, opts ...client.CallOption) (*FavouriteListRsp, error)
}

type likeService struct {
	c    client.Client
	name string
}

func NewLikeService(name string, c client.Client) LikeService {
	return &likeService{
		c:    c,
		name: name,
	}
}

func (c *likeService) IsFavorite(ctx context.Context, in *VideoUserReq, opts ...client.CallOption) (*BoolRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.IsFavorite", in)
	out := new(BoolRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *likeService) FavouriteCount(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.FavouriteCount", in)
	out := new(CountRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *likeService) TotalFavourite(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.TotalFavourite", in)
	out := new(CountRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *likeService) FavouriteVideoCount(ctx context.Context, in *IdReq, opts ...client.CallOption) (*CountRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.FavouriteVideoCount", in)
	out := new(CountRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *likeService) FavouriteAction(ctx context.Context, in *ActionReq, opts ...client.CallOption) (*ActionRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.FavouriteAction", in)
	out := new(ActionRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *likeService) GetFavouriteList(ctx context.Context, in *UserCurReq, opts ...client.CallOption) (*FavouriteListRsp, error) {
	req := c.c.NewRequest(c.name, "LikeService.GetFavouriteList", in)
	out := new(FavouriteListRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for LikeService service

type LikeServiceHandler interface {
	IsFavorite(context.Context, *VideoUserReq, *BoolRsp) error
	FavouriteCount(context.Context, *IdReq, *CountRsp) error
	TotalFavourite(context.Context, *IdReq, *CountRsp) error
	FavouriteVideoCount(context.Context, *IdReq, *CountRsp) error
	FavouriteAction(context.Context, *ActionReq, *ActionRsp) error
	GetFavouriteList(context.Context, *UserCurReq, *FavouriteListRsp) error
}

func RegisterLikeServiceHandler(s server.Server, hdlr LikeServiceHandler, opts ...server.HandlerOption) error {
	type likeService interface {
		IsFavorite(ctx context.Context, in *VideoUserReq, out *BoolRsp) error
		FavouriteCount(ctx context.Context, in *IdReq, out *CountRsp) error
		TotalFavourite(ctx context.Context, in *IdReq, out *CountRsp) error
		FavouriteVideoCount(ctx context.Context, in *IdReq, out *CountRsp) error
		FavouriteAction(ctx context.Context, in *ActionReq, out *ActionRsp) error
		GetFavouriteList(ctx context.Context, in *UserCurReq, out *FavouriteListRsp) error
	}
	type LikeService struct {
		likeService
	}
	h := &likeServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&LikeService{h}, opts...))
}

type likeServiceHandler struct {
	LikeServiceHandler
}

func (h *likeServiceHandler) IsFavorite(ctx context.Context, in *VideoUserReq, out *BoolRsp) error {
	return h.LikeServiceHandler.IsFavorite(ctx, in, out)
}

func (h *likeServiceHandler) FavouriteCount(ctx context.Context, in *IdReq, out *CountRsp) error {
	return h.LikeServiceHandler.FavouriteCount(ctx, in, out)
}

func (h *likeServiceHandler) TotalFavourite(ctx context.Context, in *IdReq, out *CountRsp) error {
	return h.LikeServiceHandler.TotalFavourite(ctx, in, out)
}

func (h *likeServiceHandler) FavouriteVideoCount(ctx context.Context, in *IdReq, out *CountRsp) error {
	return h.LikeServiceHandler.FavouriteVideoCount(ctx, in, out)
}

func (h *likeServiceHandler) FavouriteAction(ctx context.Context, in *ActionReq, out *ActionRsp) error {
	return h.LikeServiceHandler.FavouriteAction(ctx, in, out)
}

func (h *likeServiceHandler) GetFavouriteList(ctx context.Context, in *UserCurReq, out *FavouriteListRsp) error {
	return h.LikeServiceHandler.GetFavouriteList(ctx, in, out)
}
