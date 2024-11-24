package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gochat/api/handlers"
	"gochat/internal/auth"
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

	fmt.Println("running!")
	podId := os.Getenv("RUNPOD_POD_ID")

	// Check if the request method and path match the target request

	// Parse the request body
	body := RequestBody{
		Model:     "gemma2:27b",
		KeepAlive: -1,
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
	r.Static("/favicon.ico", "./frontend/dist/favicon.ico")
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

	}
	r.POST("file/upload", handlers.FileUploadHandler())
	r.GET("rag/query", handlers.RagQueryHandler())
	// Analysis

	//r.GET("/login", handlers.LoginPageHandler(sm))

}
