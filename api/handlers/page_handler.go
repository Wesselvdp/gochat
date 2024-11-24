package handlers

import (
	"bufio"
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/auth"
	"gochat/internal/rag"
	views "gochat/views"
	"gochat/views/components"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
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

		aiResponse, err := ai.GetCompletion(data.Messages)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"content": "Oeps, er is iets mis. We sturen er een ontwikkelaar op af"})
		} else {
			//fmt.Println(aiResponse)
			response := gin.H{
				"content": aiResponse,
				"data":    data.Messages,
			}
			ctx.JSON(http.StatusOK, response)

		}

	}
}

func ChatPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		id := ctx.Param("id")
		//isNew := ctx.DefaultQuery("create", "f") == "true"

		isHTMX := ctx.GetHeader("HX-Request") != ""
		if isHTMX {
			// Serve partial HTML for HTMX requests
			render(ctx, http.StatusOK, views.Chat(id))
		} else {
			render(ctx, http.StatusOK, views.ChatPage(id))
		}
	}
}

func handlePDFFile(file *multipart.FileHeader) {
	// Add your PDF processing logic here
	fmt.Printf("Processing PDF file: %s\n", file.Filename)
}

func GetTextFromFile(file *multipart.FileHeader) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	// Read the content
	content := ""
	reader := bufio.NewReader(src)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading file: %v", err)
		}
		content += string(buffer[:n])
	}
	return content, nil
}

func handleTXTFile(ctx context.Context, file *multipart.FileHeader) {
	// Add your TXT processing logic here
	fmt.Printf("Processing TXT file: %s\n", file.Filename)
	content, err := GetTextFromFile(file)
	if err != nil {
		fmt.Printf("Error processing TXT file: %v\n", err)
	}

	if err != nil {
		fmt.Printf("Error processing TXT file: %v\n", err)
	}

	err = rag.SaveEmbedding(ctx, content)
	if err != nil {
		fmt.Printf("Error processing TXT file: %v\n", err)
	}
	fmt.Printf("Processed TXT file: %s\n", file.Filename)
	//fmt.Println("embeddings:", embeddings)

}

func RagQueryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryEmbedding, err := ai.GetEmbedding("Who is elara?")
		if err != nil {
			fmt.Printf("Error getting embedding: %v\n", err)
		}
		searchResult, err := rag.SearchSimilarChunks(c, queryEmbedding, 2)
		if err != nil {
			fmt.Printf("Error getting embedding: %v\n", err)
		}
		fmt.Printf("Search result: %v\n", searchResult)
	}
}

func FileUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get file from form data
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
			})
			return
		}

		// Get the file extension
		ext := filepath.Ext(file.Filename)

		// Process file based on type
		switch ext {
		case ".pdf":
			handlePDFFile(file)
		case ".txt":
			handleTXTFile(c, file)
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported file type",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s uploaded successfully", file.Filename),
			"type":    ext,
		})
	}

}

//curl https://cqzqss48xndt4b-11434.proxy.runpod.net/api/embeddings -d '{
//"model": "mxbai-embed-large",
//"prompt": "Llamas are members of the camelid family"
//}'
