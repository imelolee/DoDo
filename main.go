package main

import (
	"github.com/genleel/DoDo/controller"
	"github.com/genleel/DoDo/middleware"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
)

func InitRouter(r *gin.Engine) {
	apiRouter := r.Group("/douyin")
	// basic apis
	apiRouter.GET("/feed/", middleware.AuthWithoutLogin(), controller.Feed)
	apiRouter.POST("/publish/action/", middleware.AuthBody(), controller.Publish)
	apiRouter.GET("/publish/list/", middleware.Auth(), controller.PublishList)
	apiRouter.GET("/user/", middleware.Auth(), controller.UserInfo)
	apiRouter.POST("/user/register/", controller.Register)
	apiRouter.POST("/user/login/", controller.Login)
	// extra apis - I
	apiRouter.POST("/favorite/action/", middleware.Auth(), controller.FavoriteAction)
	apiRouter.GET("/favorite/list/", middleware.Auth(), controller.GetFavouriteList)
	apiRouter.POST("/comment/action/", middleware.Auth(), controller.CommentAction)
	apiRouter.GET("/comment/list/", middleware.AuthWithoutLogin(), controller.CommentList)
	// extra apis - II
	apiRouter.POST("/relation/action/", middleware.Auth(), controller.RelationAction)
	apiRouter.GET("/relation/follow/list/", middleware.Auth(), controller.GetFollowing)
	apiRouter.GET("/relation/follower/list", middleware.Auth(), controller.GetFollowers)
}

// IP地址在app-Base_URL设置
func main() {
	// 日志
	log.SetOutput(ioutil.Discard)

	//gin
	r := gin.Default()
	InitRouter(r)
	//pprof
	pprof.Register(r)
	r.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
