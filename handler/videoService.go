package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	pb "github.com/Coreychen4444/shortvideo"
	"github.com/gin-gonic/gin"
)

type GetVideoFlowResponse struct {
	NextTime   *int64      `json:"next_time"`   // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
	StatusCode int64       `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  *string     `json:"status_msg"`  // 返回状态描述
	VideoList  []*pb.Video `json:"video_list"`  // 视频列表
}

// 处理获取用户视频列表请求
func (h *ServiceHandler) GetUserVideoList(c *gin.Context) {
	userID := c.Query("user_id")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误", "video_list": nil})
		return
	}
	resp, err := h.vc.GetUserVideoList(context.Background(), &pb.UserVideoListRequest{Id: id})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "获取视频失败" {
			respCode = http.StatusInternalServerError
		} else if err.Error() == "该用户不存在" {
			respCode = http.StatusNotFound
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "video_list": nil})
		return
	}
	if len(resp.VideoList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "该用户没有发布任何视频", "video_list": resp.VideoList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取用户视频列表成功", "video_list": resp.VideoList})
}

func (h *ServiceHandler) GetVideoFlow(c *gin.Context) {
	latest_time := c.Query("latest_time")
	var latestTime int64
	if latest_time == "" {
		latestTime = time.Now().Unix()
	} else {
		latest_Time, err := strconv.ParseInt(latest_time, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "请求参数错误", "video_list": nil})
			return
		}
		latestTime = latest_Time
	}
	token := c.Query("token")
	http_resp := &GetVideoFlowResponse{
		StatusCode: 1,
		StatusMsg:  nil,
		VideoList:  nil,
		NextTime:   nil,
	}
	resp, err := h.vc.GetVideoFlow(context.Background(), &pb.VideoFlowRequest{LatestTime: latestTime, Token: token})
	if err != nil {
		respCode := http.StatusBadRequest
		errMsg := err.Error()
		if errMsg == "获取视频失败" {
			respCode = http.StatusInternalServerError
		}
		http_resp.StatusMsg = &errMsg
		c.JSON(respCode, resp)
		return
	}
	http_resp.StatusCode = 0
	if len(resp.VideoList) == 0 {
		resMsg := "没有更多视频了"
		http_resp.StatusMsg = &resMsg
		http_resp.VideoList = resp.VideoList
		http_resp.NextTime = nil
		c.JSON(http.StatusOK, resp)
		return
	}
	http_resp.VideoList = resp.VideoList
	http_resp.NextTime = &resp.NextTime
	http_resp.StatusMsg = nil
	c.JSON(http.StatusOK, resp)
}

// 发布视频

func (h *ServiceHandler) PublishVideo(c *gin.Context) {
	fileHeader, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": err.Error()})
		return
	}
	//multipart变字节流
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": err.Error()})
		return
	}
	defer file.Close()
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status_code": 1, "status_msg": "Unable to read file."})
		return
	}
	token, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	tokenId, ok := token.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	title := c.PostForm("title")
	_, err = h.vc.PublishVideo(context.Background(), &pb.PublishVideoRequest{TokenUserId: int64(tokenId), Title: title, Content: fileBytes})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "创建视频失败" || err.Error() == "保存文件失败" || err.Error() == "生成封面失败" ||
			err.Error() == "发布视频失败" || err.Error() == "创建存储客户端失败" || err.Error() == "打开封面文件失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "发布视频成功"})
}

// 点赞视频和取消点赞
func (h *ServiceHandler) LikeVideo(c *gin.Context) {
	token, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	tokenId, ok := token.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	video_id := c.Query("video_id")
	videoId, err := strconv.ParseInt(video_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "视频id格式错误"})
		return
	}
	action_type := c.Query("action_type")
	_, err = h.vc.LikeVideo(context.Background(), &pb.LikeVideoRequest{TokenUserId: int64(tokenId), Id: videoId, ActionType: action_type})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效, 请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "视频出错" || err.Error() == "请求参数错误" {
			respCode = http.StatusBadRequest
		} else if err.Error() == "点赞视频失败" || err.Error() == "取消点赞失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "操作成功"})
}

// 获取用户点赞的视频列表
func (h *ServiceHandler) GetUserLike(c *gin.Context) {
	user_id := c.Query("user_id")
	userId, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误", "video_list": nil})
		return
	}
	resp, err := h.vc.GetUserLike(context.Background(), &pb.UserLikeRequest{Id: userId})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效, 请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "用户id格式错误" {
			respCode = http.StatusBadRequest
		} else if err.Error() == "获取列表失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "video_list": nil})
		return
	}
	if len(resp.VideoList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "该用户没有点赞的视频", "video_list": resp.VideoList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取列表成功", "video_list": resp.VideoList})
}

// 评论视频和删除评论
func (h *ServiceHandler) CommentVideo(c *gin.Context) {
	token, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	tokenId, ok := token.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "video_list": nil})
		return
	}
	video_id := c.Query("video_id")
	videoId, err := strconv.ParseInt(video_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "视频id格式错误"})
		return
	}
	action_type := c.Query("action_type")
	if action_type == "1" {
		comment_text := c.Query("comment_text")
		resp, err := h.vc.CommentVideo(context.Background(), &pb.CommentVideoRequest{TokenUserId: int64(tokenId), Id: videoId, Content: comment_text})
		if err != nil {
			respCode := http.StatusBadRequest
			if err.Error() == "token无效, 请重新登录" {
				respCode = http.StatusUnauthorized
			} else if err.Error() == "视频出错" || err.Error() == "请求参数错误" {
				respCode = http.StatusBadRequest
			} else if err.Error() == "评论视频失败" {
				respCode = http.StatusInternalServerError
			} else if err.Error() == "评论发表成功, 但返回评论信息失败" {
				c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": err.Error(), "comment": nil})
				return
			}
			c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "评论成功", "comment": resp.Comment})
	}
	if action_type == "2" {
		comment_id := c.Query("comment_id")
		commentId, err := strconv.ParseInt(comment_id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "评论id格式错误"})
			return
		}
		_, err = h.vc.DeleteComment(context.Background(), &pb.CommentVideoRequest{Id: commentId})
		if err != nil {
			respCode := http.StatusBadRequest
			if err.Error() == "token无效, 请重新登录" {
				respCode = http.StatusUnauthorized
			} else if err.Error() == "视频出错" || err.Error() == "请求参数错误" {
				respCode = http.StatusBadRequest
			} else if err.Error() == "删除评论失败" {
				respCode = http.StatusInternalServerError
			}
			c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "删除评论成功"})
	}
}

// 获取视频评论列表
func (h *ServiceHandler) GetVideoComment(c *gin.Context) {
	video_id := c.Query("video_id")
	videoId, err := strconv.ParseInt(video_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "视频id格式错误", "comment_list": nil})
		return
	}
	resp, err := h.vc.GetVideoComment(context.Background(), &pb.VideoCommentRequest{Id: videoId})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效, 请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "视频出错" {
			respCode = http.StatusBadRequest
		} else if err.Error() == "获取评论列表失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "comment_list": nil})
		return
	}
	if len(resp.CommentList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "该视频没有评论", "comment_list": resp.CommentList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取视频评论成功", "comment_list": resp.CommentList})
}
