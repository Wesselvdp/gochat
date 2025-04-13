package rag

import (
	"context"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// CreateFAQCollection creates the FAQ collection in Milvus if it doesn't exist
func CreateDocumentsCollection(ctx context.Context, client client.Client) error {
	// Check if collection exists first
	has, err := client.HasCollection(ctx, "documents")
	if err != nil {
		return err
	}

	if has {
		// Collection already exists, no need to create
		return nil
	}

	// Define schema for the collection
	schema := &entity.Schema{
		CollectionName: "documents",
		Description:    "documents collection",
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				AutoID:     true,
				PrimaryKey: true,
			},
			{
				Name:     "fileId",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "1024",
				},
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": "1024",
				},
			},
			{
				Name:     "text",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "2048",
				},
			},
		},
	}

	// Create collection
	err = client.CreateCollection(ctx, schema, int32(2))
	if err != nil {
		return err
	}

	// Create index on the embedding field
	idx, err := entity.NewIndexFlat(entity.COSINE) // metric_type: COSINE
	if err != nil {
		return err
	}

	err = client.CreateIndex(ctx, "documents", "embedding", idx, false)
	if err != nil {
		return err
	}

	// Load collection to memory
	err = client.LoadCollection(ctx, "documents", false)
	if err != nil {
		return err
	}

	return nil
}
