package main

import (

	"likeService/handler"
	pb "likeService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"

)

var (
	service = "likeservice"
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
	pb.RegisterLikeServiceHandler(srv.Server(), new(handler.LikeService))
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
