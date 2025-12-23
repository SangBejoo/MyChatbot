package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

type WhatsAppClient struct {
	Client      *whatsmeow.Client
	HandlerFunc func(evt interface{})
	
	UserID     int    // Owner user ID for multi-tenancy
	SchemaName string // Tenant schema for data isolation
	
	qrCode      string
	qrLock      sync.RWMutex
}

func NewWhatsAppClient(dbPath string) (*WhatsAppClient, error) {
	// Initialize SQLite container
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "sqlite", "file:"+dbPath+"?_pragma=foreign_keys(1)", dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get the first device (or create one)
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %v", err)
	}

	// Create client with standard logging
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &WhatsAppClient{
		Client: client,
	}, nil
}

// NewWhatsAppClientWithUser creates a client for a specific user (multi-tenancy)
func NewWhatsAppClientWithUser(dbPath string, userID int, schemaName string) (*WhatsAppClient, error) {
	client, err := NewWhatsAppClient(dbPath)
	if err != nil {
		return nil, err
	}
	client.UserID = userID
	client.SchemaName = schemaName
	return client, nil
}

func (w *WhatsAppClient) Connect() error {
	if w.Client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := w.Client.GetQRChannel(context.Background())
		err := w.Client.Connect()
		if err != nil {
			return err
		}
		
		// Wait for QR code in a goroutine
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					// Update QR code safely
					w.qrLock.Lock()
					w.qrCode = evt.Code
					w.qrLock.Unlock()
					
					// Print QR code to ISO standard terminal output (fallback)
					fmt.Println("QR Code:", evt.Code)
				} else {
					fmt.Println("Login event:", evt.Event)
				}
			}
		}()
	} else {
		// Already logged in
		err := w.Client.Connect()
		if err != nil {
			return err
		}
		fmt.Println("WhatsApp Client Connected (Existing Session)")
	}
	return nil
}

func (w *WhatsAppClient) GetQR() string {
	w.qrLock.RLock()
	defer w.qrLock.RUnlock()
	return w.qrCode
}

func (w *WhatsAppClient) IsLoggedIn() bool {
	return w.Client.Store.ID != nil
}

// GetUserInfo returns connected user's phone number and push name
func (w *WhatsAppClient) GetUserInfo() (string, string) {
	if w.Client.Store.ID == nil {
		return "", ""
	}
	return w.Client.Store.ID.User, w.Client.Store.PushName
}

// IsConnected returns true if client is connected and logged in
func (w *WhatsAppClient) IsConnected() bool {
	return w.Client.IsConnected() && w.Client.Store.ID != nil
}

// GetPhoneNumber returns the connected phone number
func (w *WhatsAppClient) GetPhoneNumber() string {
	if w.Client.Store.ID == nil {
		return ""
	}
	return w.Client.Store.ID.User
}

// GetName returns the push name of connected user
func (w *WhatsAppClient) GetName() string {
	if w.Client.Store.ID == nil {
		return ""
	}
	return w.Client.Store.PushName
}

func (w *WhatsAppClient) Logout() error {
	w.qrLock.Lock()
	w.qrCode = ""
	w.qrLock.Unlock()
	
	err := w.Client.Logout(context.Background())
	if err != nil {
		return err
	}
	
	// Re-initiate connection to get new QR
	// We need to disconnect first to be safe
	w.Client.Disconnect()
	
	// Create new QR channel
	qrChan, _ := w.Client.GetQRChannel(context.Background())
	err = w.Client.Connect()
	if err != nil {
		fmt.Printf("Failed to reconnect after logout: %v\n", err)
		return err
	}

	// Listen for new QR in background
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				w.qrLock.Lock()
				w.qrCode = evt.Code
				w.qrLock.Unlock()
				fmt.Println("New QR Code Generated")
			}
		}
	}()

	return nil
}
func (w *WhatsAppClient) Disconnect() {
	w.Client.Disconnect()
}

func (w *WhatsAppClient) AddHandler(handler func(interface{})) {
	w.Client.AddEventHandler(handler)
}

func (w *WhatsAppClient) SendMessage(to string, content string) error {
	// Ensure JID format (users usually just say "6289...")
	// We need to convert it to JID
	jid, err := types.ParseJID(to + "@s.whatsapp.net")
	if err != nil {
		return fmt.Errorf("invalid number format: %v", err)
	}

	// Send text message
	_, err = w.Client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: &content,
	})
	
	return err
}

// Helper to broadcast presence/typing status
func (w *WhatsAppClient) SendPresence(to string) {
	jid, _ := types.ParseJID(to + "@s.whatsapp.net")
	w.Client.SendPresence(context.Background(), types.PresenceAvailable)
	w.Client.SendChatPresence(context.Background(), jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
}

// Convert event-based message to domain entity
func (w *WhatsAppClient) ParseMessage(evt *events.Message) (string, string) {
	sender := evt.Info.Sender.User // The phone number
	var content string
	
	if evt.Message.Conversation != nil {
		content = *evt.Message.Conversation
	} else if evt.Message.ExtendedTextMessage != nil {
		content = *evt.Message.ExtendedTextMessage.Text
	}
	
	return sender, content
}
