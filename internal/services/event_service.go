package services

import (
	"context"
	"fmt"
	database "gochat/internal/db"
	"gochat/internal/schema"
)

type EventService struct {
	queries *schema.Queries
	user    string
}

type EventType string

const (
	EventLogin   EventType = "login"
	EventMessage EventType = "message"
)

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventLogin, EventMessage:
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

func (es *EventService) Create(event EventType) (*schema.Event, error) {
	// Validate event type
	if !event.IsValid() {
		return nil, fmt.Errorf("invalid event type: %s", event)
	}
	savedEvent, err := es.queries.CreateEvent(context.Background(), schema.CreateEventParams{
		User:  es.user,
		Event: string(event),
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
