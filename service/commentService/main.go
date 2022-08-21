package main

import (

	"commentService/handler"
	pb "commentService/proto"

	"go-micro.dev/v4"
	log "go-micro.dev/v4/logger"

)

var (
	service = "commentservice"
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
	pb.RegisterCommentServiceHandler(srv.Server(), new(handler.CommentService))
	// Run service
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
