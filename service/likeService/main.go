package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"likeService/handler"
	"likeService/model"
	pb "likeService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
)

var (
	service = "likeService"
	version = "latest"
)

func main() {
	model.InitRedis()
	model.InitDb()
	model.InitRabbitMQ()
	model.InitLikeRabbitMQ()

	consulReg := consul.NewRegistry()

	// Create service
	srv := micro.NewService(
		micro.Name(service),
		micro.Version(version),
		micro.Registry(consulReg),
	)

	// Register handler
	err := pb.RegisterLikeServiceHandler(srv.Server(), new(handler.LikeService))
	if err != nil {
		fmt.Println("RegisterHandler err: ", err)
		return
	}
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
