package api

import (
	"github.com/gin-gonic/gin"
	"gochat/api/handlers"
	"gochat/internal/auth"
)

func AddRoutes(r *gin.Engine) {

	//r.Static("/static", "./frontend/dist")templ
	r.Static("/static", "./frontend/dist")
	// Auth
	r.GET("login", handlers.LoginPageHandler())
	r.GET("/login/microsoft", auth.LoginHandler())
	r.GET("/login/google", auth.OauthGoogleLogin())
	r.GET("/oauth/redirect/azure", handlers.OAuthRedirectAzure(r))
	r.GET("/oauth/redirect/google", handlers.OAuthRedirectGoogle(r))

	protected := r.Group("")
	protected.Use(auth.JWTMiddleware())
	{
		protected.GET("", handlers.IndexPageHandler())
		protected.GET("c/:id", handlers.ChatPageHandler())
		protected.GET("new", handlers.NewChatHandler())
		protected.GET("component/:componentName", handlers.ComponentHandler())
		protected.POST("send-message", handlers.SendMessageHandler())

		//protected.GET("/table", handlers.Table())
	}
	// Analysis

	//r.GET("/login", handlers.LoginPageHandler(sm))

}
