package handlers

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	views "gochat/views"
	"gochat/views/components"
	"net/http"
	"time"
)

type Config struct {
	Router *gin.Engine
}

type UserRequestData struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

const appTimeout = time.Second * 10

func render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}

func IndexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.Index())
	}
}

func LoginPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.LoginPage())
	}
}

func ComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		componentName := ctx.Param("componentName")
		render(ctx, http.StatusOK, components.Component(componentName))
	}
}

func SendMessageHandler() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		var data UserRequestData
		if err := ctx.ShouldBindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{
			"content": "success",
			"data":    data.Messages,
		}

		ctx.JSON(http.StatusOK, response)
	}
}

func ChatPageHandler() gin.HandlerFunc {
	fmt.Println("running")
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		id := ctx.Param("id")
		isNew := ctx.DefaultQuery("create", "f") == "true"
		fmt.Println(isNew)
		isHTMX := ctx.GetHeader("HX-Request") != ""
		if isHTMX {
			// Serve partial HTML for HTMX requests
			render(ctx, http.StatusOK, views.Chat(id, isNew))
		} else {
			render(ctx, http.StatusOK, views.ChatPage(id, isNew))
		}
	}
}
