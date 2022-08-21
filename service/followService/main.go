package main

import (

	"followService/handler"
	pb "followService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"

)

var (
	service = "followservice"
	version = "latest"
)

func main() {
	// Create service
	srv := micro.NewService(
		micro.Name(service),
		micro.Version(version),
	)
	srv.Init()

	// Register handler
	pb.RegisterFollowServiceHandler(srv.Server(), new(handler.FollowService))
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
