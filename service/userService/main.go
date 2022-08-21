package main

import (
	"fmt"
	"github.com/go-micro/plugins/v4/registry/consul"
	"userService/handler"
	"userService/model"
	pb "userService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"
)

var (
	service = "userService"
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
	err := pb.RegisterUserServiceHandler(srv.Server(), new(handler.UserService))
	if err != nil {
		fmt.Println("RegisterHandler err: ", err)
		return
	}
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
