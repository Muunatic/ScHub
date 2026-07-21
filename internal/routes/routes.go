package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/handler"
	"github.com/muunatic/schub/internal/middleware"
)

func SetupRoutes(
	r *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	authHandler *handler.AuthHandler,
	oauthHandler *handler.OAuthHandler,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	userHandler *handler.UserHandler,
	viewHandler *handler.ViewHandler,
) {
	r.HEAD("/", viewHandler.HeadView)
	r.GET("/", viewHandler.HomeView)
	r.GET("/dashboard", authMiddleware.Restrict(), viewHandler.DashboardView)
	r.GET("/account", authMiddleware.Restrict(), viewHandler.AccountView)
	r.GET("/post", authMiddleware.Restrict(), viewHandler.PostView)
	r.GET("/comment", authMiddleware.Restrict(), viewHandler.CommentView)
	r.GET("/register", viewHandler.RegisterView)
	r.GET("/login", viewHandler.LoginView)
	r.GET("/privacy", viewHandler.PrivacyView)
	r.GET("/terms", viewHandler.TermsView)
	// r.GET("/verify", viewHandler.VerifyView)

	v1 := r.Group("/v1")
	{
		v1.POST("/login", authHandler.Login)
		v1.POST("/register", authHandler.Register)
		v1.POST("/logout", authMiddleware.Restrict(), authHandler.Logout)
		// v1.POST("/verify", authHandler.Verify)
	}

	api := r.Group("/api/v1")
	api.Use(authMiddleware.Restrict())
	{
		api.POST("/post", middleware.FilterQueryParams(), userHandler.UploadPost)
		api.GET("/post", userHandler.ListPost)
		api.GET("/comment", userHandler.ListComment)

		api.GET("/auth/twitter", oauthHandler.TwitterAuth)
		api.GET("/auth/twitter/callback", oauthHandler.TwitterCallback)
		api.GET("/auth/tiktok", oauthHandler.TikTokAuth)
		api.GET("/auth/tiktok/callback", oauthHandler.TikTokCallback)
		api.GET("/auth/youtube", oauthHandler.YouTubeAuth)
		api.GET("/auth/youtube/callback", oauthHandler.YouTubeCallback)

		platformRoutes := api.Group("/:platform")
		{
			platformRoutes.POST("/post", authMiddleware.RefreshToken(), postHandler.UploadVideo)
			platformRoutes.GET("/post", authMiddleware.RefreshToken(), postHandler.FetchPosts)
			platformRoutes.GET("/comment", authMiddleware.RefreshToken(), commentHandler.FetchComments)
		}
	}
}
