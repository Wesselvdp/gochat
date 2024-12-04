package services_test

import (
	"github.com/stretchr/testify/assert"
	"gochat/internal/services"
	"os"
	"testing"
)

func TestEventService(t *testing.T) {
	os.Setenv("DB_PATH", "../../database.db")
	testUserID := "123ABCD"
	// New service
	eventService := services.NewEventService(testUserID)
	assert.NotNil(t, eventService)

	// Create Event
	event, err := eventService.Create(services.EventLogin, nil)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, event)
	assert.Equal(t, testUserID, event.User)

	// Get event
	savedEvent := eventService.Get(event.ID)
	assert.NotNil(t, savedEvent)
	assert.Equal(t, testUserID, event.User)

	// Event 2
	_, err = eventService.Create("ThisIsWrong", nil)
	assert.NotNil(t, err)

}
