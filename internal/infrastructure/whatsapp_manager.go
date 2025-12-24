package infrastructure

import (
	"fmt"
	"os"
	"sync"
)

// WhatsAppManager manages per-user WhatsApp clients
type WhatsAppManager struct {
	clients  map[int]*WhatsAppClient
	mu       sync.RWMutex
	baseDir  string
	
	// Callback for registering message handlers per client
	HandlerFactory func(userID int, schemaName string) func(interface{})
}

// NewWhatsAppManager creates a new manager for per-user WhatsApp clients
func NewWhatsAppManager(baseDir string) *WhatsAppManager {
	// Ensure devices directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Warning: could not create devices directory: %v\n", err)
	}
	
	return &WhatsAppManager{
		clients: make(map[int]*WhatsAppClient),
		baseDir: baseDir,
	}
}

// GetClient returns existing client for user (nil if not exists or not connected)
func (m *WhatsAppManager) GetClient(userID int) *WhatsAppClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[userID]
}

// GetOrCreateClient gets existing client or creates new one for user
func (m *WhatsAppManager) GetOrCreateClient(userID int, schemaName string) (*WhatsAppClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if already exists
	if client, exists := m.clients[userID]; exists {
		return client, nil
	}
	
	// Create new client with user-specific DB
	dbPath := fmt.Sprintf("%s/user_%d.db", m.baseDir, userID)
	client, err := NewWhatsAppClientWithUser(dbPath, userID, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp client for user %d: %w", userID, err)
	}
	
	// Register handler if factory is set
	if m.HandlerFactory != nil {
		handler := m.HandlerFactory(userID, schemaName)
		client.AddHandler(handler)
	}
	
	m.clients[userID] = client
	return client, nil
}

// ConnectClient connects user's WhatsApp client (creates if needed)
func (m *WhatsAppManager) ConnectClient(userID int, schemaName string) (*WhatsAppClient, error) {
	client, err := m.GetOrCreateClient(userID, schemaName)
	if err != nil {
		return nil, err
	}
	
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect WhatsApp for user %d: %w", userID, err)
	}
	
	return client, nil
}

// DisconnectClient disconnects user's WhatsApp client
func (m *WhatsAppManager) DisconnectClient(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if client, exists := m.clients[userID]; exists {
		client.Disconnect()
		delete(m.clients, userID)
	}
}

// LogoutClient logs out user's WhatsApp (clears session, shows new QR)
// Returns nil if client doesn't exist or already logged out (graceful handling)
func (m *WhatsAppManager) LogoutClient(userID int) error {
	m.mu.RLock()
	client, exists := m.clients[userID]
	m.mu.RUnlock()
	
	// No client = already logged out, return success
	if !exists || client == nil {
		return nil
	}
	
	// Check if already disconnected
	if !client.IsLoggedIn() && !client.Client.IsConnected() {
		// Clean up the client from map
		m.mu.Lock()
		delete(m.clients, userID)
		m.mu.Unlock()
		return nil
	}
	
	// Attempt logout, ignore errors from already-disconnected state
	err := client.Logout()
	
	// Clean up client from map regardless of logout result
	m.mu.Lock()
	delete(m.clients, userID)
	m.mu.Unlock()
	
	return err
}

// GetAllConnectedUsers returns list of userIDs with active connections
func (m *WhatsAppManager) GetAllConnectedUsers() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var users []int
	for userID, client := range m.clients {
		if client.IsLoggedIn() {
			users = append(users, userID)
		}
	}
	return users
}

// DisconnectAll disconnects all clients (for graceful shutdown)
func (m *WhatsAppManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, client := range m.clients {
		client.Disconnect()
	}
	m.clients = make(map[int]*WhatsAppClient)
}
