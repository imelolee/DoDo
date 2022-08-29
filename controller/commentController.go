package controller

import (
	"context"
	"fmt"
	"github.com/genleel/DoDo/model"
	"github.com/genleel/DoDo/proto/commentService"
	"github.com/genleel/DoDo/utils"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
	"log"
	"net/http"
	"strconv"
	"time"
)

// CommentListResponse 评论列表返回参数
type CommentListResponse struct {
	StatusCode  int32               `json:"status_code"`
	StatusMsg   string              `json:"status_msg,omitempty"`
	CommentList []model.CommentInfo `json:"comment_list,omitempty"`
}

// CommentActionResponse
// 发表评论返回参数
type CommentActionResponse struct {
	StatusCode int32             `json:"status_code"`
	StatusMsg  string            `json:"status_msg,omitempty"`
	Comment    model.CommentInfo `json:"comment"`
}

// CommentAction
// 发表 or 删除评论 comment/action/
func CommentAction(c *gin.Context) {
	fmt.Println("CommentController-Comment_Action: running") //函数已运行
	//获取userId
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)
	fmt.Printf("err:%v", err)
	fmt.Printf("userId:%v", userId)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment userId json invalid",
		})
		fmt.Println("CommentController-Comment_Action: return comment userId json invalid") //函数返回userId无效
		return
	}
	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		fmt.Println("CommentController-Comment_Action: return comment videoId json invalid") //函数返回视频id无效
		return
	}
	//获取操作类型
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
	//错误处理
	if err != nil || actionType < 1 || actionType > 2 {
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "comment actionType json invalid",
		})
		fmt.Println("CommentController-Comment_Action: return actionType json invalid") //评论类型数据无效
		return
	}
	//调用service层评论函数
	commentMicro := utils.InitMicro()
	commentClient := commentService.NewCommentService("commentService", commentMicro.Client())
	if actionType == 1 { //actionType为1，则进行发表评论操作
		content := c.Query("comment_text")

		//发表评论数据准备
		var sendComment model.Comment
		sendComment.UserId = userId
		sendComment.VideoId = videoId
		sendComment.CommentText = content
		timeNow := time.Now()
		sendComment.CreateDate = timeNow
		//发表评论
		commentRsp, err := commentClient.Send(context.TODO(), &commentService.CommentReq{})
		//发表评论失败
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "send comment failed",
			})
			fmt.Println("CommentController-Comment_Action: return send comment failed") //发表失败
			return
		}

		var tmpComment model.CommentInfo
		gconv.Struct(commentRsp.CommentInfo, &tmpComment)
		//发表评论成功:
		//返回结果
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "send comment success",
			Comment:    tmpComment,
		})
		fmt.Println("CommentController-Comment_Action: return Send success") //发表评论成功，返回正确信息
		return
	} else { //actionType为2，则进行删除评论操作
		//获取要删除的评论的id
		commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "delete commentId invalid",
			})
			fmt.Println("CommentController-Comment_Action: return commentId invalid") //评论id格式错误
			return
		}
		//删除评论操作
		_, err = commentClient.Delete(context.TODO(), &commentService.IdReq{
			Id: commentId,
		})
		if err != nil { //删除评论失败
			str := err.Error()
			c.JSON(http.StatusOK, CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  str,
			})
			log.Println("CommentController-Comment_Action: return delete comment failed") //删除失败
			return
		}
		//删除评论成功
		c.JSON(http.StatusOK, CommentActionResponse{
			StatusCode: 0,
			StatusMsg:  "delete comment success",
		})

		log.Println("CommentController-Comment_Action: return delete success") //函数执行成功，返回正确信息
		return
	}
}

// CommentList
// 查看评论列表 comment/list/
func CommentList(c *gin.Context) {
	log.Println("CommentController-Comment_List: running") //函数已运行
	//获取userId
	id, _ := c.Get("userId")
	userid, _ := id.(string)
	userId, err := strconv.ParseInt(userid, 10, 64)

	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		log.Println("CommentController-Comment_List: return videoId json invalid") //视频id格式有误
		return
	}
	log.Printf("videoId:%v", videoId)

	//调用service层评论函数
	commentMicro := utils.InitMicro()
	commentClient := commentService.NewCommentService("commentService", commentMicro.Client())
	listRsp, err := commentClient.GetList(context.TODO(), &commentService.VideoUserReq{
		VideoId: videoId,
		UserId:  userId,
	})
	//commentList, err := commentService.GetListFromRedis(videoId, userId)
	if err != nil { //获取评论列表失败
		c.JSON(http.StatusOK, CommentListResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
		log.Println("CommentController-Comment_List: return list false") //查询列表失败
		return
	}

	var tmpList []model.CommentInfo
	gconv.Struct(listRsp.CommentInfo, &tmpList)

	//获取评论列表成功
	c.JSON(http.StatusOK, CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "get comment list success",
		CommentList: tmpList,
	})
	log.Println("CommentController-Comment_List: return success") //成功返回列表
	return
}
