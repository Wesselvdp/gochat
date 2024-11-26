package services

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	database "gochat/internal/db"
	"gochat/internal/schema"
)

type FileService struct {
	queries *schema.Queries
	owner   string
}

func NewFileService(ctx *gin.Context) (*FileService, error) {
	queries, _, err := database.Init()
	if err != nil {
		return nil, fmt.Errorf("Error initializing queries for event service: " + err.Error())
	}
	//ctx.Set("user", "sientje")
	owner, exist := ctx.Get("user")
	if !exist {
		return nil, fmt.Errorf("user not found in context")
	}
	return &FileService{queries: queries, owner: owner.(string)}, nil
}

// Create Creates a file in the database
func (fs *FileService) Create(ctx context.Context, fileName string) (*schema.File, error) {
	id := uuid.New().String()
	savedFile, err := fs.queries.CreateFile(ctx, schema.CreateFileParams{
		Owner: fs.owner,
		ID:    id,
		Name:  fileName,
	})

	if err != nil {
		fmt.Printf("failed to create file: %s", err)
		return nil, err
	}
	return &savedFile, nil
}

func (fs *FileService) Get(id int64) *schema.Event {
	file, err := fs.queries.GetEvent(context.Background(), id)
	if err != nil {
		fmt.Printf("failed to GET file: %s", err)
		return nil
	}
	return &file
}
