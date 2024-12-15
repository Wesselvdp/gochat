package services

import (
	"context"
	"encoding/json"
	"fmt"
	database "gochat/internal/db"
	"gochat/internal/schema"
)

type EventService struct {
	queries *schema.Queries
	user    string
}

type EventType string
type EventMetadata map[string]interface{}

const (
	EventLogin     EventType = "login"
	EventMessage   EventType = "message"
	UnknownAccount EventType = "unknownAccount"
	Evil           EventType = "evil"
)

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventLogin, EventMessage, UnknownAccount, Evil:
		return true
	}
	return false
}

func NewEventService(userId string) *EventService {
	queries, _, err := database.Init()
	if err != nil {
		fmt.Println("Error initializing queries for event service: " + err.Error())
		return nil
	}
	return &EventService{queries: queries, user: userId}
}

func (es *EventService) Create(event EventType, metadata interface{}) (*schema.Event, error) {
	// Validate event type
	if !event.IsValid() {
		return nil, fmt.Errorf("invalid event type: %s", event)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v", err)
	}

	savedEvent, err := es.queries.CreateEvent(context.Background(), schema.CreateEventParams{
		User:     es.user,
		Event:    string(event),
		Metadata: string(metadataJSON),
	})

	if err != nil {
		fmt.Printf("failed to create event: %s", err)
		return nil, err
	}
	return &savedEvent, nil
}
func (es *EventService) Get(id int64) *schema.Event {
	event, err := es.queries.GetEvent(context.Background(), id)
	if err != nil {
		fmt.Printf("failed to GET event: %s", err)
		return nil
	}
	return &event
}
