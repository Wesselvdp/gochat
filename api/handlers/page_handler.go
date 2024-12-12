package handlers

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/auth"
	"gochat/internal/rag"
	"gochat/internal/schema"
	"gochat/internal/services"
	views "gochat/views"
	"gochat/views/components"
	"net/http"
	"time"
)

type Config struct {
	Router *gin.Engine
}

type UserRequestData struct {
	Messages       []openai.ChatCompletionMessage `json:"messages"`
	HasFiles       bool                           `json:"hasFiles"`
	ConversationID string                         `json:"conversationId"`
}

const appTimeout = time.Second * 10

func render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}

func getUserData(ctx *gin.Context) (*schema.GetUserRow, error) {
	userID := ctx.GetString("user")
	userService := services.NewUserService()
	user, err := userService.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func IndexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()
		user, err := getUserData(ctx)
		if err != nil {
			fmt.Println("err", err)
		}

		render(ctx, http.StatusOK, views.Index(user))
	}
}

func LoginPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.LoginPage())
	}
}

func LogoutPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth.UnsetTokenCookie(ctx)
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.LoginPage())
	}
}

func ComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()
		userID := ctx.GetString("user")
		userService := services.NewUserService()
		user, err := userService.Get(ctx, userID)

		if err != nil {
			fmt.Println("err", err)
		}
		componentName := ctx.Param("componentName")

		ctx.Header("Cache-Control", "public, max-age=3600")
		render(ctx, http.StatusOK, components.Component(componentName, user))
	}
}

func SendMessageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timing
		start := time.Now()

		// Prepare variables
		var aiResponse string
		var data UserRequestData

		// Bind JSON and handle potential error
		if err := ctx.ShouldBindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Capture user ID
		userID, exists := ctx.Get("user")
		eventService := services.NewEventService(userID.(string))

		var err error
		// Process message based on file existence
		if data.HasFiles {
			aiResponse, err = rag.GetRaggedAnswer(ctx, data.Messages, data.ConversationID)
		} else {
			aiResponse, err = ai.GetCompletion(data.Messages)
		}

		// Log execution time if user exists
		if exists {
			_, logErr := eventService.Create(services.EventMessage, services.EventMetadata{
				"conversation":  data.ConversationID,
				"hasFiles":      data.HasFiles,
				"executionTime": time.Since(start).Seconds(), // Full handler duration
			})
			if logErr != nil {
				// Consider using a proper logger in production
				fmt.Printf("Failed to log execution time: %v\n", logErr)
			}
		}

		// Handle potential processing error
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"content": "Oeps, er is iets mis. We sturen er een ontwikkelaar op af",
				"error":   err.Error(),
			})
			return
		}

		// Return successful response
		response := gin.H{
			"content": aiResponse,
			"data":    data.Messages,
		}
		ctx.JSON(http.StatusOK, response)
	}
}

func ChatPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		conversationID := ctx.Param("id")
		user, err := getUserData(ctx)
		if err != nil {
			fmt.Println("err", err)
		}
		isHTMX := ctx.GetHeader("HX-Request") != ""
		if isHTMX {
			// Serve partial HTML for HTMX requests
			render(ctx, http.StatusOK, views.Chat(conversationID, user))
		} else {
			render(ctx, http.StatusOK, views.ChatPage(conversationID, user))
		}
	}
}

func FileUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")
		// Get file from form data
		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
			})
			return
		}
		// Save file entry locally
		fileService, err := services.NewFileService(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		dbEntry, err := fileService.Create(c, file.Filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		fileID := dbEntry.ID

		// Create embeddings and save to vector DB
		err = rag.HandleFileEmbedding(c, file, fileID, conversationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s uploaded successfully", file.Filename),
			"id":      fileID,
		})
	}

}
func FileDeleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")
		fileID := c.PostForm("fileId")

		// Get file from form data
		// Save file entry locally
		err := rag.RemoveDocumentsByFileId(c, fileID, conversationID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s successfully deleted", fileID),
		})
	}

}

func PartitionDeleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")

		// Get file from form data
		// Save file entry locally
		err := rag.RemovePartition(c, conversationID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Partition %s successfully deleted", conversationID),
		})
	}

}
