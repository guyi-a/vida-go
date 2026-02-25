package router

import (
	"vida-go/internal/api/handler"
	"vida-go/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 注册所有业务路由
func Setup(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	relationHandler *handler.RelationHandler,
	videoHandler *handler.VideoHandler,
	commentHandler *handler.CommentHandler,
	favoriteHandler *handler.FavoriteHandler,
	adminMiddleware gin.HandlerFunc,
) {
	v1 := r.Group("/api/v1")

	// --- 认证模块 ---
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		authRequired := auth.Group("", middleware.AuthRequired())
		{
			authRequired.POST("/logout", authHandler.Logout)
			authRequired.GET("/me", authHandler.Me)
		}
	}

	// --- 用户模块 ---
	users := v1.Group("/users", middleware.AuthRequired())
	{
		users.GET("/me", userHandler.GetMe)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", userHandler.UpdateUser)

		// 管理员接口
		admin := users.Group("", adminMiddleware)
		{
			admin.GET("", userHandler.ListUsers)
			admin.DELETE("/:id", userHandler.DeleteUser)
			admin.POST("/:id/restore", userHandler.RestoreUser)
			admin.POST("/:id/set-admin", userHandler.SetAdmin)
		}
	}

	// --- 关注关系模块 ---
	relations := v1.Group("/relations", middleware.AuthRequired())
	{
		relations.POST("/follow/:id", relationHandler.Follow)
		relations.POST("/unfollow/:id", relationHandler.Unfollow)

		relations.GET("/following/:id", relationHandler.GetFollowing)
		relations.GET("/followers/:id", relationHandler.GetFollowers)
		relations.GET("/following/:id/status", relationHandler.GetFollowStatus)

		relations.GET("/following/my/list", relationHandler.GetMyFollowing)
		relations.GET("/followers/my/list", relationHandler.GetMyFollowers)
		relations.GET("/mutual", relationHandler.GetMutualFollows)

		relations.POST("/batch/status", relationHandler.BatchFollowStatus)
	}

	// --- 视频模块 ---
	videos := v1.Group("/videos")
	{
		// 公开接口（不需要登录）
		videos.GET("/feed", videoHandler.GetFeed)

		// 需要登录的接口
		videosAuth := videos.Group("", middleware.AuthRequired())
		{
			videosAuth.POST("/upload", videoHandler.Upload)
			videosAuth.GET("/my/list", videoHandler.GetMyVideos)
			videosAuth.GET("/:id", videoHandler.GetDetail)
			videosAuth.PUT("/:id", videoHandler.UpdateVideo)
			videosAuth.DELETE("/:id", videoHandler.DeleteVideo)
		}
	}

	// --- 评论模块 ---
	comments := v1.Group("/comments")
	{
		commentsAuth := comments.Group("", middleware.AuthRequired())
		{
			commentsAuth.POST("/:video_id", commentHandler.Create)
			commentsAuth.PUT("/:id", commentHandler.Update)
			commentsAuth.DELETE("/:id", commentHandler.Delete)
			commentsAuth.GET("/video/:video_id", commentHandler.ListByVideo)
			commentsAuth.GET("/:id/replies", commentHandler.ListReplies)
			commentsAuth.GET("/my/list", commentHandler.ListMyComments)
		}
	}

	// --- 点赞模块 ---
	favorites := v1.Group("/favorites", middleware.AuthRequired())
	{
		favorites.POST("/:video_id", favoriteHandler.Favorite)
		favorites.DELETE("/:video_id", favoriteHandler.Unfavorite)
		favorites.GET("/:video_id/status", favoriteHandler.GetStatus)
		favorites.GET("/my/list", favoriteHandler.ListMyFavorites)
		favorites.GET("/my/videos", favoriteHandler.GetMyFavoritedVideos)
		favorites.GET("/video/:video_id/list", favoriteHandler.ListVideoFavorites)
		favorites.POST("/batch/status", favoriteHandler.BatchStatus)
	}
}
