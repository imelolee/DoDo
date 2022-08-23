package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4/cmd/protoc-gen-micro/plugin/micro"
	"videoService/handler"
	"videoService/model"
	pb "videoService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
)

var (
	service = "videoService"
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
	err := pb.RegisterVideoServiceHandler(srv.Server(), new(handler.VideoService))
	if err != nil {
		fmt.Println("RegisterHandler err: ", err)
		return
	}
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
