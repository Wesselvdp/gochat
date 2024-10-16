package main_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"os"
	"rboai/internal/apiSchema"
	"rboai/internal/models"
	"rboai/internal/run"

	"testing"
	"time"
)

////go:embed ../schema/apiSchema.sql
//var ddl string

type Response struct {
	Message  string               `json:"message"`
	Analyses []apiSchema.Analysis `json:"analyses"` // Assuming Analysis is a defined type
}

func TestPingRoute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Cleanup(cancel)

	port := "8080"
	args := []string{
		"--port", port,
	}

	now := time.Now()

	err := os.Setenv("dataSourceName", fmt.Sprintf("test-%s.db", now))
	//_, db, err := database.Init()
	if err != nil {
		fmt.Errorf("failed to initialize database: %w", err)
	}

	// create tables
	//if _, err := db.ExecContext(ctx, ddl); err != nil {
	//	fmt.Errorf("failed creating tables: %w", err)
	//}

	url := fmt.Sprintf("http://localhost:%s", port)

	// Run server in new go routine
	go func() {
		defer cancel()
		err := run.Run(ctx, os.Stdout, args)
		if err != nil {
			t.Errorf("Failed to run server: %v", err)
		}
	}()

	// Make the request
	respHealth, err := http.Get(url + "/health")

	var params [2]models.CreateAnalysisParams

	params[0] = models.CreateAnalysisParams{
		Source: "gesprek",
		Status: "Positief",
	}
	params[1] = models.CreateAnalysisParams{
		Source: "email",
		Status: "Negatief",
	}

	// Marshal the data to JSON
	jsonCreateData, err := json.Marshal(params)
	jsonSearchData, err := json.Marshal([]string{"ABC"})

	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	respCreate, err := http.Post(url+"/analyses", "application/json", bytes.NewBuffer(jsonCreateData))
	respSearch, err := http.Post(url+"/analyses/search", "application/json", bytes.NewBuffer(jsonSearchData))
	respList, err := http.Get(url + "/analyses")

	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	// Check the response
	assert.Equal(t, 200, respHealth.StatusCode)
	assert.Equal(t, 200, respCreate.StatusCode)
	assert.Equal(t, 200, respList.StatusCode)

	// Read the response body
	createBody, err := io.ReadAll(respCreate.Body)
	listBody, err := io.ReadAll(respList.Body)
	searchBody, err := io.ReadAll(respSearch.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}

	// Attempt to parse the JSON
	type Response struct {
		Message  string               `json:"message"`
		Analyses []apiSchema.Analysis `json:"analyses"`
	}
	var createResponse Response
	var listResponse Response
	var searchResponse Response
	err = json.Unmarshal(createBody, &createResponse)
	err = json.Unmarshal(searchBody, &searchResponse)
	err = json.Unmarshal(listBody, &listResponse)
	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}

	// Assertions
	assert.Equal(t, "success", createResponse.Message)
	assert.Equal(t, 2, len(createResponse.Analyses))
	assert.Equal(t, "success", listResponse.Message)
	assert.Equal(t, 2, len(listResponse.Analyses))
	assert.Equal(t, 1, len(searchResponse.Analyses))

	cancel()
}
