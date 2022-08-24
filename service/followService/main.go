package main

import (
	"fmt"
	"followService/handler"
	"followService/model"
	pb "followService/proto"
	"github.com/go-micro/plugins/v4/registry/consul"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
)

var (
	service = "followService"
	version = "latest"
)

func main() {
	model.InitRedis()
	model.InitDb()
	model.InitRabbitMQ()

	consulReg := consul.NewRegistry()

	// Create service
	srv := micro.NewService(
		micro.Name(service),
		micro.Version(version),
		micro.Registry(consulReg),
	)

	// Register handler
	err := pb.RegisterFollowServiceHandler(srv.Server(), new(handler.FollowService))
	if err != nil {
		fmt.Println("RegisterHandler err: ", err)
		return
	}
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
