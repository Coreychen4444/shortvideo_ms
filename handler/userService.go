package handler

import (
	"context"
	"net/http"
	"strconv"

	pb "github.com/Coreychen4444/shortvideo"
	"github.com/gin-gonic/gin"
)

type RegisterResponse struct {
	StatusCode int64  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
	Token      string `json:"token"`       // 用户鉴权token
	UserID     int64  `json:"user_id"`     // 用户id
}

// 处理注册请求
func (h *ServiceHandler) Register(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	resp, err := h.uc.Register(context.Background(), &pb.UserRequest{Username: username, Password: password})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "生成token时出错" || err.Error() == "查找用户时出错" || err.Error() == "创建用户时出错" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "token": "", "user_id": -1})
		return
	}
	http_resp := &RegisterResponse{
		StatusCode: 0,
		StatusMsg:  "注册成功",
		Token:      resp.Token,
		UserID:     resp.Id,
	}
	c.JSON(http.StatusOK, http_resp)
}

// 处理登录请求
func (h *ServiceHandler) Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	resp, err := h.uc.Login(context.Background(), &pb.UserRequest{Username: username, Password: password})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "生成token时出错" || err.Error() == "验证密码时出错" || err.Error() == "查找用户时出错" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "token": "", "user_id": -1})
		return
	}
	http_resp := &RegisterResponse{
		StatusCode: 0,
		StatusMsg:  "登录成功",
		Token:      resp.Token,
		UserID:     resp.Id,
	}
	c.JSON(http.StatusOK, http_resp)
}

type GetUserInfoResponse struct {
	StatusCode int64    `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  *string  `json:"status_msg"`  // 返回状态描述
	User       *pb.User `json:"user"`        // 用户信息
}

// 处理获取用户信息请求
func (h *ServiceHandler) GetUserInfo(c *gin.Context) {
	user_id := c.Query("user_id")
	userId, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误", "user": nil})
		return
	}
	token_user_id, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "user": nil})
		return
	}
	tokenId, ok := token_user_id.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "user": nil})
		return
	}
	resp, err := h.uc.GetUserInfo(context.Background(), &pb.UserInfoRequest{Id: userId, TokenUserId: int64(tokenId)})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "该用户不存在" {
			respCode = http.StatusNotFound
		} else if err.Error() == "查找用户时出错" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "user": nil})
		return
	}
	statusMsg := "获取用户信息成功"
	http_resp := &GetUserInfoResponse{
		StatusCode: 0,
		StatusMsg:  &statusMsg,
		User:       resp.User,
	}
	c.JSON(http.StatusOK, http_resp)
}

// 关注或取消关注
func (h *ServiceHandler) FollowOrCancel(c *gin.Context) {
	token_user_id, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录"})
		return
	}
	tokenId, ok := token_user_id.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录"})
		return
	}
	to_user_id := c.Query("to_user_id")
	toUserId, err := strconv.ParseInt(to_user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误"})
		return
	}
	action_type := c.Query("action_type")
	_, err = h.uc.FollowOrCancel(context.Background(), &pb.FollowOrCancelRequest{TokenUserId: int64(tokenId), Id: toUserId, ActionType: action_type})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "关注失败" || err.Error() == "取消关注失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
		return
	}
	if action_type == "1" {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "关注成功"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "已取消关注"})
}

// 获取用户关注列表
func (h *ServiceHandler) GetFollowings(c *gin.Context) {
	user_id := c.Query("user_id")
	userId, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误"})
		return
	}
	resp, err := h.uc.GetFollowings(context.Background(), &pb.GetFollowingsRequest{Id: userId})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "获取关注列表失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "user_list": nil})
		return
	}
	if len(resp.UserList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "暂无关注", "user_list": resp.UserList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取关注列表成功", "user_list": resp.UserList})
}

// 获取用户粉丝列表
func (h *ServiceHandler) GetFollowers(c *gin.Context) {
	user_id := c.Query("user_id")
	userId, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误"})
		return
	}
	resp, err := h.uc.GetFollowers(context.Background(), &pb.GetFollowersRequest{Id: userId})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "获取粉丝列表失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "user_list": nil})
		return
	}
	if len(resp.UserList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "暂无粉丝", "user_list": resp.UserList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取粉丝列表成功", "user_list": resp.UserList})
}

// 获取用户好友列表
func (h *ServiceHandler) GetFriends(c *gin.Context) {
	user_id := c.Query("user_id")
	userId, err := strconv.ParseInt(user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误"})
		return
	}
	resp, err := h.uc.GetFriends(context.Background(), &pb.GetFriendsRequest{Id: userId})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "获取好友列表失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "user_list": nil})
		return
	}
	if len(resp.UserList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "暂无好友", "user_list": resp.UserList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取好友列表成功", "user_list": resp.UserList})
}

// 获取聊天记录
func (h *ServiceHandler) GetChatMessages(c *gin.Context) {
	to_user_id := c.Query("to_user_id")
	toUserId, err := strconv.ParseInt(to_user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误", "message_list": nil})
		return
	}
	token, ok := c.Get("UserId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "message_list": nil})
		return
	}
	tokenId, ok := token.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录", "message_list": nil})
		return
	}
	pre_msg_time := c.Query("pre_msg_time")
	if pre_msg_time == "" {
		pre_msg_time = "0"
	}
	resp, err := h.uc.GetChatMessages(context.Background(), &pb.GetChatMessagesRequest{TokenUserId: int64(tokenId), ToUserId: toUserId, PreMsgTime: pre_msg_time})
	if err != nil {
		respCode := http.StatusBadRequest
		if err.Error() == "token无效,请重新登录" {
			respCode = http.StatusUnauthorized
		} else if err.Error() == "获取聊天记录失败" {
			respCode = http.StatusInternalServerError
		}
		c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error(), "message_list": nil})
		return
	}
	if len(resp.ChatMessageList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "已是最新消息", "message_list": resp.ChatMessageList})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "获取聊天记录成功", "message_list": resp.ChatMessageList})
}

type SendMessageRequest struct {
	ActionType string `json:"action_type"` // 1-发送消息
	Content    string `json:"content"`     // 消息内容
	ToUserID   string `json:"to_user_id"`  // 对方用户id
	Token      string `json:"token"`       // 用户鉴权token
}

// 发送消息
func (h *ServiceHandler) SendMessage(c *gin.Context) {
	token, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录"})
		return
	}
	tokenId, ok := token.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"status_code": 1, "status_msg": "token无效,请重新登录"})
		return
	}
	to_user_id := c.Query("to_user_id")
	toUserId, err := strconv.ParseInt(to_user_id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status_code": 1, "status_msg": "用户id格式错误"})
		return
	}
	action_type := c.Query("action_type")
	content := c.Query("content")
	if action_type == "1" {
		_, err = h.uc.SendMessage(context.Background(), &pb.SendMessageRequest{TokenUserId: int64(tokenId), ToUserId: toUserId, Content: content})
		if err != nil {
			respCode := http.StatusBadRequest
			if err.Error() == "token无效,请重新登录" {
				respCode = http.StatusUnauthorized
			} else if err.Error() == "发送消息失败" {
				respCode = http.StatusInternalServerError
			}
			c.JSON(respCode, gin.H{"status_code": 1, "status_msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status_code": 0, "status_msg": "发送消息成功"})
	}
}
