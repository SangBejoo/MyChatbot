package usecases

import (
	"encoding/json"
	"fmt"
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/interfaces"
	"project_masAde/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageService struct {
	geminiClient    interfaces.AIClient
	messengerClient interfaces.Messenger
	TelegramClient  *infrastructure.TelegramClient
	WhatsAppClient  *infrastructure.WhatsAppClient
	ProductRepo     *repository.ProductRepository
	ConfigRepo      *repository.ConfigRepository
	TableManager    *repository.TableManager
}

func NewMessageService(gemini interfaces.AIClient, messenger interfaces.Messenger, configRepo *repository.ConfigRepository, tableManager *repository.TableManager) *MessageService {
	return &MessageService{
		geminiClient:    gemini,
		messengerClient: messenger,
		ConfigRepo:      configRepo,
		TableManager:    tableManager,
	}
}


func (s *MessageService) ProcessMessage(msg entities.Message) error {
	// 1. Check Dynamic Menu Commands
	if handled, err := s.handleDynamicMenu(msg); err != nil {
		fmt.Printf("Menu handling error: %v\n", err)
	} else if handled {
		return nil
	}

	// 2. Fallback to AI
	// Fetch configurable system prompt
	systemPrompt := `You are a helpful assistant. 
Keep responses SHORT and CONCISE.
Only answer based on the provided context. 
If the question is outside your knowledge, politely decline.
Do NOT make up information.`

	if s.ConfigRepo != nil {
		if customPrompt, err := s.ConfigRepo.GetConfig("public", "ai_system_prompt"); err == nil && customPrompt != "" {
			systemPrompt = customPrompt
		}
	}

	var fullPrompt string
	if msg.AIContext != "" {
		fullPrompt = systemPrompt + "\n\nCONTEXT/DATABASE:\n" + msg.AIContext + "\n\nUser: " + msg.Content
	} else {
		fullPrompt = systemPrompt + "\n\nUser: " + msg.Content
	}

	aiResponse, err := s.geminiClient.GenerateResponse(fullPrompt)
	if err != nil {
		if s.WhatsAppClient != nil && msg.Platform == "whatsapp" {
			return s.WhatsAppClient.SendMessage(msg.From, "Error generating response: "+err.Error())
		}
		return s.messengerClient.SendMessage(msg.From, "Error generating response: "+err.Error())
	}

	if s.WhatsAppClient != nil && msg.Platform == "whatsapp" {
		return s.WhatsAppClient.SendMessage(msg.From, aiResponse)
	}
	response := entities.Response{Content: aiResponse}
	return s.messengerClient.SendMessage(msg.From, response.Content)
}

func (s *MessageService) handleDynamicMenu(msg entities.Message) (bool, error) {
	if s.ConfigRepo == nil {
		return false, nil
	}

	// Fetch 'main_menu' (hardcoded for now, could be contextual)
	menu, err := s.ConfigRepo.GetMenu("public", "main_menu")
	
	// If menu doesn't exist, ignore (or log)
	if err != nil {
		return false, nil // Not treated as error to allow flow to continue
	}

	// Parse Items
	// The 'Items' field in database is JSONB but repository returns struct where Items is interface{}?
	// Let's check repository.Menu definition.
	// Assuming Items is json.RawMessage or similar if coming from Postgres JSONB, 
	// or we need to marshal/unmarshal.
	// Actually repository.Menu.Items is defined as interface{} in previous context? 
	// Let's assume it maps to []MenuItem struct structure if Unmarshaled.
	
	// Safest way given Go types: Marshal back to bytes then Unmarshal to []MenuItem
	itemsBytes, _ := json.Marshal(menu.Items)
	var items []repository.MenuItem
	if err := json.Unmarshal(itemsBytes, &items); err != nil {
		return false, fmt.Errorf("invalid menu items json: %w", err)
	}

	for _, item := range items {
		// Simple Case-Insensitive Match
		// Note: User can type "1" or "Label". For now assume exact Label match or "1" if we implemented numbering.
		// Let's stick to Label match for simplicity of "customize menu".
		if item.Label == msg.Content { // Exact match for now
			// Execute Action
			switch item.Action {
			case "view_table":
				return s.handleViewTable(msg.From, item.Payload)
			case "reply":
				if s.WhatsAppClient != nil {
					return true, s.WhatsAppClient.SendMessage(msg.From, item.Payload)
				}
			}
		}
	}

	return false, nil
}

func (s *MessageService) handleViewTable(to, tableName string) (bool, error) {
	if s.TableManager == nil {
		return false, fmt.Errorf("table manager not initialized")
	}
	data, err := s.TableManager.GetTableData("public", tableName)
	if err != nil {
		return true, s.WhatsAppClient.SendMessage(to, fmt.Sprintf("Error fetching table '%s': %v", tableName, err))
	}

	if len(data) == 0 {
		return true, s.WhatsAppClient.SendMessage(to, fmt.Sprintf("Table '%s' is empty.", tableName))
	}

	// Format as simple list
	var sb string
	sb = fmt.Sprintf("*%s Data:*\n\n", tableName)
	
	// Limit to 10 rows for safety
	limit := 10
	if len(data) < limit {
		limit = len(data)
	}

	for i := 0; i < limit; i++ {
		row := data[i]
		sb += "- "
		// Iterate keys? Order is random in map. 
		// Ideally we use known columns but they are dynamic.
		// Just print all values.
		for k, v := range row {
			if k != "id" { // skip internal ID
				sb += fmt.Sprintf("%s: %v, ", k, v)
			}
		}
		sb += "\n"
	}
	
	if len(data) > limit {
		sb += fmt.Sprintf("\n...and %d more rows.", len(data)-limit)
	}

	return true, s.WhatsAppClient.SendMessage(to, sb)
}

func (s *MessageService) ProcessMessageWithContext(msg entities.Message) error {
	systemPrompt := `You are a helpful export-import sales assistant. 
IMPORTANT: Keep responses SHORT and CONCISE (1-2 sentences max).
Only answer based on the provided product database. 
Do NOT answer questions outside the database.

PRODUCT DATABASE:
` + msg.AIContext + `

Answer briefly, be professional, provide specific pricing from database.`

	fullPrompt := systemPrompt + "\n\nCustomer: " + msg.Content

	aiResponse, err := s.geminiClient.GenerateResponse(fullPrompt)
	if err != nil {
		if s.WhatsAppClient != nil && msg.Platform == "whatsapp" {
			return s.WhatsAppClient.SendMessage(msg.From, "Error generating response: "+err.Error())
		}
		return s.messengerClient.SendMessage(msg.From, "Error generating response: "+err.Error())
	}

	// WhatsApp Specific Logic
	if msg.Platform == "whatsapp" && s.WhatsAppClient != nil {
		// Append text-based menu
		menuText := "\n\n1. 🧮 Calculate Price\n2. ❓ Ask More\n3. 🏠 Back to Menu\n\n_Reply with number to choose_"
		return s.WhatsAppClient.SendMessage(msg.From, aiResponse+menuText)
	}

	// Telegram Specific Logic
	if s.TelegramClient != nil {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🧮 Calculate Price", "action_calculate"),
				tgbotapi.NewInlineKeyboardButtonData("❓ Ask More", "action_ask"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "action_menu"),
			),
		)
		return s.TelegramClient.SendMessageWithMenu(msg.From, aiResponse, keyboard)
	}

	// Fallback
	response := entities.Response{Content: aiResponse}
	return s.messengerClient.SendMessage(msg.From, response.Content)
}