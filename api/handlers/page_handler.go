package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
	views "gochat/views"
	"gochat/views/components"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Router *gin.Engine
}

type UserRequestData struct {
	messages []openai.ChatCompletionMessage
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
			"data":    data.messages,
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

func NewChatHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		id := uuid.New()
		// Set the HX-Push-Url header
		responseHeader := map[string]string{
			"path":   "/c/" + id.String() + "?create=true",
			"target": "#inner",
		}

		// Convert the map to a JSON string
		jsonResponseHeader, err := json.Marshal(responseHeader)
		if err != nil {
			// Handle error, e.g., log it or return an appropriate response
			log.Printf("Error marshalling JSON: %v", err)
			ctx.String(500, "Internal Server Error")
			return
		}

		// Set the "HX-Location" header with the JSON string
		ctx.Header("HX-Location", string(jsonResponseHeader))
		return
	}
}
