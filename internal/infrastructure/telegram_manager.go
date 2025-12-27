package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"project_masAde/internal/repository"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotInstance represents a single user's Telegram bot
type TelegramBotInstance struct {
	Bot       *tgbotapi.BotAPI
	UserID    int
	Schema    string
	StopChan  chan struct{}
	IsRunning bool
	mu        sync.Mutex
}

// TelegramBotManager manages per-user Telegram bot instances
type TelegramBotManager struct {
	bots       map[int]*TelegramBotInstance
	mu         sync.RWMutex
	configRepo *repository.ConfigRepository
	tableManager *repository.TableManager
	
	// Handler factory for message processing
	MessageHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update, userID int, schema string)
}

// NewTelegramBotManager creates a new manager for per-user Telegram bots
func NewTelegramBotManager(configRepo *repository.ConfigRepository, tableManager *repository.TableManager) *TelegramBotManager {
	return &TelegramBotManager{
		bots:         make(map[int]*TelegramBotInstance),
		configRepo:   configRepo,
		tableManager: tableManager,
	}
}

// GetBot returns existing bot for user (nil if not connected)
func (m *TelegramBotManager) GetBot(userID int) *TelegramBotInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bots[userID]
}

// ValidateToken checks if a token is valid by creating a test bot
func (m *TelegramBotManager) ValidateToken(token string) (string, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}
	return bot.Self.UserName, nil
}

// ConnectBot creates and starts a bot for a user with their token
func (m *TelegramBotManager) ConnectBot(userID int, schema, token string) (*TelegramBotInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if already running
	if existing, ok := m.bots[userID]; ok && existing.IsRunning {
		return existing, nil
	}
	
	// Create new bot
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	
	instance := &TelegramBotInstance{
		Bot:       bot,
		UserID:    userID,
		Schema:    schema,
		StopChan:  make(chan struct{}),
		IsRunning: false,
	}
	
	m.bots[userID] = instance
	
	// Start polling in goroutine
	go m.startPolling(instance)
	
	return instance, nil
}

// startPolling runs the update loop for a user's bot
func (m *TelegramBotManager) startPolling(instance *TelegramBotInstance) {
	instance.mu.Lock()
	instance.IsRunning = true
	instance.mu.Unlock()
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := instance.Bot.GetUpdatesChan(u)
	
	fmt.Printf("[TG Bot] Started polling for user %d (@%s)\n", instance.UserID, instance.Bot.Self.UserName)
	
	for {
		select {
		case <-instance.StopChan:
			fmt.Printf("[TG Bot] Stopped polling for user %d\n", instance.UserID)
			instance.mu.Lock()
			instance.IsRunning = false
			instance.mu.Unlock()
			return
		case update := <-updates:
			if m.MessageHandler != nil {
				go m.MessageHandler(instance.Bot, update, instance.UserID, instance.Schema)
			} else {
				// Default simple handler
				m.defaultHandler(instance, update)
			}
		}
	}
}

// defaultHandler handles messages when no custom handler is set
func (m *TelegramBotManager) defaultHandler(instance *TelegramBotInstance, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	
	chatID := update.Message.Chat.ID
	
	// Handle /start command
	if update.Message.IsCommand() && update.Message.Command() == "start" {
		msg := tgbotapi.NewMessage(chatID, "Welcome! 👋")
		
		// Try to get dynamic menu
		if m.configRepo != nil {
			menu, err := m.configRepo.GetMenu(instance.Schema, "main_menu")
			if err == nil {
				msg.Text = menu.Title + "\n\nChoose an option:"
				
				// Build keyboard
				itemsBytes, _ := json.Marshal(menu.Items)
				var items []repository.MenuItem
				if err := json.Unmarshal(itemsBytes, &items); err == nil {
					keyboard := GenerateDynamicKeyboardFromItems(items)
					msg.ReplyMarkup = &keyboard
				}
			}
		}
		
		instance.Bot.Send(msg)
		return
	}
	
	// Echo for testing
	reply := tgbotapi.NewMessage(chatID, "Bot connected! Message received: "+update.Message.Text)
	instance.Bot.Send(reply)
}

// GenerateDynamicKeyboardFromItems creates keyboard from menu items
func GenerateDynamicKeyboardFromItems(items []repository.MenuItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	
	for i, item := range items {
		btnData := "dyn:" + item.Action + ":" + item.Payload
		btn := tgbotapi.NewInlineKeyboardButtonData(item.Label, btnData)
		row = append(row, btn)
		
		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// DisconnectBot stops a user's bot
func (m *TelegramBotManager) DisconnectBot(userID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if instance, ok := m.bots[userID]; ok {
		close(instance.StopChan)
		delete(m.bots, userID)
	}
}

// GetStatus returns connection status for a user
func (m *TelegramBotManager) GetStatus(userID int) (connected bool, botName string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if instance, ok := m.bots[userID]; ok && instance.IsRunning {
		return true, instance.Bot.Self.UserName
	}
	return false, ""
}

// DisconnectAll stops all bots (for graceful shutdown)
func (m *TelegramBotManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, instance := range m.bots {
		close(instance.StopChan)
	}
	m.bots = make(map[int]*TelegramBotInstance)
}

// SendMessage sends a message via a user's bot
func (m *TelegramBotManager) SendMessage(userID int, chatID int64, text string) error {
	m.mu.RLock()
	instance, ok := m.bots[userID]
	m.mu.RUnlock()
	
	if !ok || !instance.IsRunning {
		return fmt.Errorf("bot not connected for user %d", userID)
	}
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := instance.Bot.Send(msg)
	return err
}

// HandleCallbackQuery processes callback queries for per-user bots
func (m *TelegramBotManager) HandleCallbackQuery(instance *TelegramBotInstance, callback *tgbotapi.CallbackQuery, ctx context.Context) {
	// Acknowledge callback
	instance.Bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	
	// Handle callback data (can be extended)
	chatID := callback.Message.Chat.ID
	reply := tgbotapi.NewMessage(chatID, "Processing: "+callback.Data)
	instance.Bot.Send(reply)
}
