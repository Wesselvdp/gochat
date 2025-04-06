package services // services/client_manager.go

import (
	"fmt"
	"sync"
)

// ClientManager keeps track of SSE connections per conversation
type ClientManager struct {
	clients map[string]chan string
	mutex   sync.RWMutex
}

// NewClientManager creates a new instance of ClientManager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]chan string),
	}
}

// RegisterClient registers a new client by conversation ID
func (m *ClientManager) RegisterClient(conversationID string) chan string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close existing client channel if it exists
	if ch, exists := m.clients[conversationID]; exists {
		close(ch)
	}

	// Create new channel for this conversation
	ch := make(chan string, 10)
	m.clients[conversationID] = ch
	return ch
}

// SendToConversation sends a message to a specific conversation
func (m *ClientManager) SendToConversation(conversationID string, message string) bool {
	fmt.Println("Sending message to conversation "+conversationID, "message: "+message)
	sseData := fmt.Sprintf("event: message\ndata: {\"content\":%q,\"isDone\":false}\n\n", message)
	m.mutex.RLock()
	ch, exists := m.clients[conversationID]
	m.mutex.RUnlock()

	if !exists {
		return false
	}

	// Try to send with non-blocking operation
	select {
	case ch <- sseData:
		return true
	default:
		return false
	}
}

func (m *ClientManager) SendRawEventToConversation(conversationID string, eventType string, data string) bool {
	// Format as SSE event
	//fmt.Println("data", data)
	sseEvent := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, data)

	m.mutex.RLock()
	ch, exists := m.clients[conversationID]
	m.mutex.RUnlock()

	if !exists {
		return false
	}

	// Try to send with non-blocking operation
	select {
	case ch <- sseEvent:
		return true
	default:
		return false
	}
}

// UnregisterClient removes a client
func (m *ClientManager) UnregisterClient(conversationID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if ch, exists := m.clients[conversationID]; exists {
		close(ch)
		delete(m.clients, conversationID)
	}
}
