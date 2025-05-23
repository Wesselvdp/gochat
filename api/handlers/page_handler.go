package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/auth"
	"gochat/internal/rag"

	"gochat/internal/services"
	views "gochat/views"
	"gochat/views/components"
	"io"
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

// TODO move somewhere else
func notify(str string) {
	url := "https://ntfy.sh/alberttorgon"
	data := []byte(str)

	resp, err := http.Post(url, "text/plain", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

}

func getUserData(ctx *gin.Context) (*services.UserDto, error) {
	userID := ctx.GetString("user")
	userService := services.NewUserService()
	user, err := userService.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func impersonateAccount(ctx *gin.Context, impersonationID string) (*services.UserDto, error) {
	userID := ctx.GetString("user")
	userService := services.NewUserService()
	user, err := userService.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	// If it's not me, we have a joker
	if user.Email != "wessel@torgon.io" {
		metadata := map[string]interface{}{
			"ip":         ctx.ClientIP(),
			"user_agent": ctx.GetHeader("User-Agent"),
			"method":     ctx.Request.Method,
			"url":        ctx.Request.RequestURI,
			"headers":    ctx.Request.Header,
			"user_id":    userID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}

		notify("joker alert")

		// Read and reset body if necessary
		if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" {
			body, _ := io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body for further use
			metadata["body"] = string(body)
		}

		eventService := services.NewEventService(userID)
		_, err := eventService.Create(services.Evil, metadata)

		if err != nil {
			fmt.Println("Error creating Evil event", err)
		}

		return nil, errors.New("not found")
	}

	accountService := services.NewAccountService()
	if err != nil {
		return nil, err
	}
	account, err := accountService.Get(ctx, impersonationID)
	if err != nil {
		fmt.Println("accountservice2 error:", err)

		return nil, err
	}
	if account == nil {
		return nil, nil
	}

	user.Account.ID = impersonationID

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

func ImpersonateIndexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		impersonationID := ctx.Param("id")

		user, err := impersonateAccount(ctx, impersonationID)
		if err != nil {
			fmt.Println("err", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

			return
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
		var err error
		var errorMessage string

		// Capture user ID
		userID, exists := ctx.Get("user")

		defer func() {
			if exists {

				eventService := services.NewEventService(userID.(string))
				_, logErr := eventService.Create(services.EventMessage, services.EventMetadata{
					"conversation":  data.ConversationID,
					"hasFiles":      data.HasFiles,
					"err":           errorMessage,
					"executionTime": time.Since(start).Seconds(), // Full handler duration
				})

				if errorMessage != "" {
					str := fmt.Sprintf("error in sendMessage: %s", errorMessage)
					notify(str)
				}

				if logErr != nil {
					// Consider using a proper logger in production
					fmt.Printf("Failed to log execution time: %v\n", logErr)
				}
			}
		}()

		// Bind JSON and handle potential error
		if err = ctx.ShouldBindJSON(&data); err != nil {
			errorMessage = err.Error()
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Process message based on file existence
		//if data.HasFiles {
		//	aiResponse, err = rag.GetRaggedAnswer(ctx, data.Messages, data.ConversationID)
		//} else {
		aiResponse, err = ai.GetCompletion(data.Messages)
		//}
		// Handle potential processing error
		if err != nil {
			errorMessage = err.Error()
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

func ThreadPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		threadID := ctx.Param("id")
		user, err := getUserData(ctx)
		if err != nil {
			fmt.Println("err", err)
		}
		isHTMX := ctx.GetHeader("HX-Request") != ""

		render(ctx, http.StatusOK, views.ThreadPage(threadID, user, isHTMX))
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
			fmt.Println("err", err)
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
