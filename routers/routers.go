// 微服务API网关路由配置
package routers

import (
	pb "github.com/Coreychen4444/shortvideo"
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/auth"
	"github.com/Coreychen4444/shortvideo_ms-api_gateway/handler"
	"github.com/gin-gonic/gin"
)

// InitRouter initialize routing information
func InitRouter(UC *pb.UserServiceClient, VC *pb.VideoServiceClient) *gin.Engine {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	handler := handler.NewServiceHandler(UC, VC)
	needAuth := r.Group("/douyin")
	needAuth.Use(auth.TokenAuthMiddleware())
	{
		needAuth.GET("/user/", handler.GetUserInfo)
		needAuth.POST("/publish/action/", handler.PublishVideo)
		needAuth.GET("/publish/list/", handler.GetUserVideoList)
		needAuth.POST("/favorite/action/", handler.LikeVideo)
		needAuth.GET("/favorite/list/", handler.GetUserLike)
		needAuth.POST("/comment/action/", handler.CommentVideo)
		needAuth.GET("/comment/list/", handler.GetVideoComment)
		needAuth.GET("/message/chat/", handler.GetChatMessages)
		needAuth.POST("/message/action/", handler.SendMessage)
	}
	user := r.Group("/douyin/user")
	{
		user.POST("/register/", handler.Register)
		user.POST("/login/", handler.Login)
	}
	r.GET("/douyin/feed", handler.GetVideoFlow)
	relation := r.Group("/douyin/relation")
	relation.Use(auth.TokenAuthMiddleware())
	{
		relation.POST("/action/", handler.FollowOrCancel)
		relation.GET("/follow/list/", handler.GetFollowings)
		relation.GET("/follower/list/", handler.GetFollowers)
		relation.GET("/friend/list/", handler.GetFriends)
	}
	return r
}
