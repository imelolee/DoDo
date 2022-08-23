package main

import (
	"commentService/handler"
	"commentService/model"
	pb "commentService/proto"
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"log"
)

var (
	service = "commentService"
	version = "latest"
)

func main() {
	model.InitRedis()
	model.InitDb()

	consulReg := consul.NewRegistry()

	// Create service
	srv := micro.NewService(
		micro.Name(service),
		micro.Version(version),
		micro.Registry(consulReg),
	)

	// Register handler
	err := pb.RegisterCommentServiceHandler(srv.Server(), new(handler.CommentService))
	if err != nil {
		fmt.Println("RegisterHandler err: ", err)
		return
	}
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
