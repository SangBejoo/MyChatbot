package usecases

import (
	"encoding/json"
	"fmt"
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/interfaces"
	"project_masAde/internal/repository"
	"strings"

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

// ProcessMessage handles incoming messages with priority-based response system
// Priority: 1. Greeting → 2. MENU → 3. Menu Selection → 4. Keyword → 5. Search → 6. AI (optional) → 7. Default
func (s *MessageService) ProcessMessage(msg entities.Message) error {
	content := strings.TrimSpace(msg.Content)
	contentLower := strings.ToLower(content)
	schema := msg.SchemaName
	if schema == "" {
		schema = "public"
	}

	// DEBUG: Log what we received
	fmt.Printf("[BOT] Received: '%s' (lower: '%s') from %s, schema: %s\n", content, contentLower, msg.From, schema)

	// 1. GREETING DETECTION
	if s.isGreeting(contentLower) {
		fmt.Printf("[BOT] Matched: GREETING\n")
		return s.sendReply(msg, s.getWelcomeMessage(schema))
	}

	// 2. MENU COMMAND - Show all available menus (flexible matching)
	if s.isMenuCommand(contentLower) {
		fmt.Printf("[BOT] Matched: MENU command\n")
		return s.sendReply(msg, s.getMenuList(schema))
	}

	// 3. DYNAMIC MENU HANDLING (exact match and keywords)
	if handled, err := s.handleDynamicMenu(msg); err != nil {
		fmt.Printf("Menu handling error: %v\n", err)
	} else if handled {
		return nil
	}

	// 4. DATASET SEARCH - "cari X", "search X", "harga X"
	if strings.HasPrefix(contentLower, "cari ") || strings.HasPrefix(contentLower, "search ") || strings.HasPrefix(contentLower, "harga ") {
		query := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(contentLower, "cari "), "search "), "harga ")
		return s.handleDatasetSearch(msg, schema, query)
	}

	// 5. CALCULATION (has weight like "2kg", "500g")
	if s.hasWeightPattern(contentLower) {
		// For now, just indicate calculation is possible - can be expanded
		return s.sendReply(msg, "🧮 *Untuk menghitung harga:*\nSilakan pilih produk dari MENU, lalu masukkan jumlah yang diinginkan.\n\nKetik *MENU* untuk melihat pilihan.")
	}

	// 6. AI RESPONSE (only if enabled)
	aiEnabled := s.isAIEnabled(schema)
	if aiEnabled && s.geminiClient != nil {
		return s.processWithAI(msg, schema)
	}

	// 7. DEFAULT FALLBACK
	return s.sendReply(msg, s.getDefaultResponse())
}

// isGreeting checks if message is a greeting
func (s *MessageService) isGreeting(content string) bool {
	greetings := []string{"halo", "hai", "hello", "hi", "selamat pagi", "selamat siang", "selamat sore", "selamat malam", "assalamualaikum", "asslmkm"}
	for _, g := range greetings {
		if strings.Contains(content, g) {
			return true
		}
	}
	return false
}

// isMenuCommand checks if message is a menu request
func (s *MessageService) isMenuCommand(content string) bool {
	menuCommands := []string{"menu", "help", "?", "daftar", "pilihan", "opsi"}
	for _, cmd := range menuCommands {
		if content == cmd || strings.HasPrefix(content, cmd+" ") {
			return true
		}
	}
	return false
}

// getWelcomeMessage returns configured or default welcome message
func (s *MessageService) getWelcomeMessage(schema string) string {
	if s.ConfigRepo != nil {
		if welcome, err := s.ConfigRepo.GetConfig(schema, "welcome_message"); err == nil && welcome != "" {
			return welcome
		}
	}
	return "👋 *Selamat datang!*\n\nSaya adalah asisten virtual.\nKetik *MENU* untuk melihat pilihan yang tersedia."
}

// getMenuList returns formatted list of available menus
func (s *MessageService) getMenuList(schema string) string {
	if s.ConfigRepo == nil {
		return "Menu tidak tersedia."
	}

	menus, err := s.ConfigRepo.GetAllMenus(schema)
	if err != nil || len(menus) == 0 {
		return "📋 *Menu*\n\nBelum ada menu yang dikonfigurasi.\nHubungi admin untuk setup."
	}

	var sb strings.Builder
	sb.WriteString("📋 *Menu Tersedia:*\n\n")
	
	for i, menu := range menus {
		sb.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, menu.Title))
		// Parse menu items
		itemsBytes, _ := json.Marshal(menu.Items)
		var items []repository.MenuItem
		if json.Unmarshal(itemsBytes, &items) == nil {
			for _, item := range items {
				sb.WriteString(fmt.Sprintf("   • %s\n", item.Label))
			}
		}
		sb.WriteString("\n")
	}
	
	sb.WriteString("_Ketik nama menu atau pilihan untuk melanjutkan_")
	return sb.String()
}

// handleDatasetSearch searches all tables for matching data
func (s *MessageService) handleDatasetSearch(msg entities.Message, schema, query string) error {
	if s.TableManager == nil {
		return s.sendReply(msg, "Fitur pencarian tidak tersedia.")
	}

	// Get all tables
	tables, err := s.TableManager.ListTables(schema)
	if err != nil || len(tables) == 0 {
		return s.sendReply(msg, "Tidak ada dataset untuk dicari.")
	}

	var results strings.Builder
	results.WriteString(fmt.Sprintf("🔍 *Hasil pencarian \"%s\":*\n\n", query))
	totalFound := 0

	for _, table := range tables {
		data, err := s.TableManager.GetTableData(schema, table.TableName)
		if err != nil {
			continue
		}

		// Search in each row
		for _, row := range data {
			for _, value := range row {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), query) {
					totalFound++
					results.WriteString(fmt.Sprintf("📦 *%s*: ", table.DisplayName))
					for k, v := range row {
						if k != "id" {
							results.WriteString(fmt.Sprintf("%s=%v ", k, v))
						}
					}
					results.WriteString("\n")
					break // Only show row once
				}
			}
			if totalFound >= 5 { // Limit results
				break
			}
		}
		if totalFound >= 5 {
			break
		}
	}

	if totalFound == 0 {
		return s.sendReply(msg, fmt.Sprintf("❌ Tidak ditemukan hasil untuk \"%s\".\n\nKetik *MENU* untuk melihat pilihan.", query))
	}

	return s.sendReply(msg, results.String())
}

// hasWeightPattern checks if message contains weight pattern like "2kg" or "500g"
func (s *MessageService) hasWeightPattern(content string) bool {
	// Simple check for kg or g with number
	for i, c := range content {
		if c >= '0' && c <= '9' {
			rest := content[i:]
			if strings.Contains(rest, "kg") || strings.Contains(rest, " kg") ||
				strings.Contains(rest, "g ") || strings.HasSuffix(rest, "g") {
				return true
			}
		}
	}
	return false
}

// isAIEnabled checks if AI is enabled for this tenant
func (s *MessageService) isAIEnabled(schema string) bool {
	if s.ConfigRepo == nil {
		return false // Default: AI disabled
	}
	enabled, err := s.ConfigRepo.GetConfig(schema, "ai_enabled")
	if err != nil {
		return false // Default: AI disabled if config not found
	}
	return enabled == "true" || enabled == "1"
}

// processWithAI handles AI-powered response (only called if AI enabled)
func (s *MessageService) processWithAI(msg entities.Message, schema string) error {
	systemPrompt := `You are a helpful assistant. 
Keep responses SHORT and CONCISE.
Only answer based on the provided context. 
If the question is outside your knowledge, politely decline.
Do NOT make up information.`

	if s.ConfigRepo != nil {
		if customPrompt, err := s.ConfigRepo.GetConfig(schema, "ai_system_prompt"); err == nil && customPrompt != "" {
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
		// AI failed - return friendly error
		return s.sendReply(msg, "⚠️ Maaf, sistem AI sedang sibuk. Silakan coba lagi nanti.\n\nKetik *MENU* untuk melihat pilihan manual.")
	}

	return s.sendReply(msg, aiResponse)
}

// getDefaultResponse returns default fallback message
func (s *MessageService) getDefaultResponse() string {
	return "🤔 Maaf, saya tidak mengerti pesan Anda.\n\n" +
		"Silakan coba:\n" +
		"• Ketik *MENU* untuk melihat pilihan\n" +
		"• Ketik *CARI [nama]* untuk mencari produk\n" +
		"• Atau pilih dari menu yang tersedia"
}

// sendReply sends message back to user based on platform
func (s *MessageService) sendReply(msg entities.Message, text string) error {
	if s.WhatsAppClient != nil && msg.Platform == "whatsapp" {
		return s.WhatsAppClient.SendMessage(msg.From, text)
	}
	if s.TelegramClient != nil && msg.Platform == "telegram" {
		return s.TelegramClient.SendMessage(msg.From, text)
	}
	if s.messengerClient != nil {
		return s.messengerClient.SendMessage(msg.From, text)
	}
	return fmt.Errorf("no messaging client available")
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