package rag

import (
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"gochat/internal/ai"
	"log"
)

const (
	// Milvus instance proxy address, may verify in your env/settings
	milvusAddr = `localhost:19530`

	collectionName      = `documents`
	dim                 = 1024
	idCol, embeddingCol = "id", "embedding"
)

type Document struct {
	Text      string    // Original text
	Embedding []float32 // Vector embedding
	ID        int64
}

func SaveEmbedding(ctx context.Context, content string) error {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: milvusAddr,
	})
	if err != nil {
		// handling error and exit, to make example simple here
		log.Fatal("failed to connect to milvus:", err.Error())
	}

	docs := chunkText(content)

	// Prepare the data columns
	numDocs := len(docs)

	texts := make([]string, numDocs)
	embeddings := make([][]float32, numDocs)

	// Split the data into columns
	for i, doc := range docs {
		texts[i] = doc.Text
		embeddings[i] = doc.Embedding
	}
	// Split the data into columns
	for i, doc := range docs {
		texts[i] = doc.Text
		embeddings[i] = doc.Embedding
	}

	// Create column-based data
	textCol := entity.NewColumnVarChar("text", texts)
	embeddingCol := entity.NewColumnFloatVector("embedding", 1024, embeddings)

	// Insert data
	_, err = milvusClient.Insert(
		ctx,
		collectionName,
		"", // partition name (empty for default)
		textCol,
		embeddingCol,
	)

	if err != nil {
		return err
	}

	// Optional: Flush to make the data immediately searchable
	err = milvusClient.Flush(ctx, collectionName, false)
	if err != nil {
		return err
	}

	return nil
}
func chunkText(text string) []Document {
	// Define chunk size and overlap (adjust as needed)
	chunkSize := 62
	overlap := 12
	//var chunks []string
	start := 0

	var docs []Document

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunk := text[start:end]
		vector, err := ai.GetEmbedding(chunk)
		if err != nil {
			fmt.Println("GetEmbedding err:", err.Error())
		}
		doc := Document{
			Text:      chunk,
			Embedding: vector,
			ID:        int64(chunkSize - overlap),
		}
		docs = append(docs, doc)
		start += chunkSize - overlap
	}

	return docs
}

// Expanded to include more context about the chunk
type SearchResult struct {
	Text       string  // The text chunk
	ChunkIndex int     // Position in original document
	Score      float32 // Similarity score
	Distance   float32 // Vector distance
}

func SearchSimilarChunks(
	ctx context.Context,
	queryEmbedding []float32,
	topK int64,
) ([]SearchResult, error) {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: milvusAddr,
	})
	if err != nil {
		// handling error and exit, to make example simple here
		log.Fatal("failed to connect to milvus:", err.Error())
	}
	err = milvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	sp, err := entity.NewIndexIvfFlatSearchParam(10)
	if err != nil {
		return nil, fmt.Errorf("failed to create search parameters: %w", err)
	}

	vectors := []entity.Vector{
		entity.FloatVector(queryEmbedding),
	}

	// Now also retrieving chunk_index
	searchResult, err := milvusClient.Search(
		ctx,
		collectionName,
		[]string{"_default"},
		"",
		[]string{"text"},
		vectors,
		"embedding",
		entity.L2,
		int(topK),
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	//fmt.Printf("%#v\n", searchResult)
	for i, sr := range searchResult {
		fmt.Println(sr.Fields.GetColumn("text").GetAsString(i))
		fmt.Println(sr.Scores)
	}

	results := make([]SearchResult, 0, topK)

	for i := 0; i < len(searchResult); i++ {
		current := searchResult[i]
		text, err := current.Fields.GetColumn("text").GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get text: %w", err)
		}

		results = append(results, SearchResult{
			Text:  text,
			Score: current.Scores[i],
		})
	}
	//
	return results, nil
}
