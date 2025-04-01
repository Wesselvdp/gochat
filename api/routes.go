package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gochat/api/handlers"
	"gochat/internal/auth"
	"gochat/internal/services"
	"io/ioutil"
	"net/http"
	"os"
)

type RequestBody struct {
	Model     string `json:"model"`
	KeepAlive int    `json:"keep_alive"`
}

func afterRequestMiddleware(c *gin.Context) {
	// Execute the main handler
	c.Next()

	podId := os.Getenv("RUNPOD_POD_ID")

	// Check if the request method and path match the target request

	// Parse the request body
	body := RequestBody{
		Model:     "gemma2:27b",
		KeepAlive: 86400, // 24 hours in seconds
	}

	if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Make the API call
	apiURL := "https://" + podId + "-11434.proxy.runpod.net/api/generate"
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("API response:", string(respBody))

}

func AddRoutes(r *gin.Engine) {

	//r.Static("/static", "./frontend/dist")templ
	r.Static("/static", "./frontend/dist")
	r.Static("/e4694570-f591-4c52-bba9-a5865dc4ba09.ico", "./frontend/dist/e4694570-f591-4c52-bba9-a5865dc4ba09.ico")
	// Auth
	r.GET("login", handlers.LoginPageHandler())
	r.GET("/login/microsoft", auth.LoginHandler())
	r.GET("/login/google", auth.OauthGoogleLogin())
	r.GET("/oauth/redirect/azure", handlers.OAuthRedirectAzure(r))
	//r.GET("/oauth/redirect/google", handlers.OAuthRedirectGoogle(r))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/logout", handlers.LogoutPageHandler())

	protected := r.Group("")
	protected.Use(auth.JWTMiddleware())
	{
		protected.GET("", handlers.IndexPageHandler())
		protected.GET("c/:id", handlers.ChatPageHandler())
		protected.GET("component/:componentName", handlers.ComponentHandler())
		protected.POST("send-message", afterRequestMiddleware, handlers.SendMessageHandler())
		protected.POST("file/upload", handlers.FileUploadHandler())
		protected.POST("file/delete", handlers.FileDeleteHandler())
		protected.POST("conversation/delete", handlers.PartitionDeleteHandler())
		protected.GET("impersonate/:id", handlers.ImpersonateIndexPageHandler())

	}

	admin := r.Group("patron")
	authConfig := auth.LoadAdminAuthConfig()
	admin.Use(auth.AdminAuthMiddleware(authConfig))

	accountService := services.NewAccountService()
	accountHandlers := handlers.NewAccountHandlers(accountService)
	{
		admin.GET("user/:id", accountHandlers.GetUser())
		admin.GET("account/:id", accountHandlers.GetAccount())
		admin.POST("account/create", accountHandlers.CreateAccount())
		admin.POST("account/accountdomains/create", accountHandlers.AddDomain())
		admin.GET("account/accountdomains/delete/:domain", accountHandlers.DeleteAccountDomain())
		admin.POST("account/change-user-account", accountHandlers.ChangeUserAccount())
	}
}
