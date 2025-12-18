package main

import (
	"os"
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/interfaces/http"
	"project_masAde/internal/repository"
	"project_masAde/internal/usecases"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Initialize data repository
	repo := repository.NewProductRepository()
	err = repo.LoadFromCSV("data/products.csv")
	if err != nil {
		panic("Error loading products CSV: " + err.Error())
	}

	// Initialize session manager for spam prevention
	sessionManager := infrastructure.NewSessionManager()

	// Initialize dependencies
	geminiClient := infrastructure.NewGeminiClient(os.Getenv("GEMINI_API_KEY"))
	telegramClient := infrastructure.NewTelegramClient(os.Getenv("TELEGRAM_BOT_TOKEN"))
	messageService := usecases.NewMessageService(geminiClient, telegramClient)
	messageService.ProductRepo = repo // Attach repository
	messageService.TelegramClient = telegramClient.(*infrastructure.TelegramClient) // Attach for menu sending

	// Initialize pricing calculator
	pricingCalc := usecases.NewPricingCalculator(repo)

	// Setup HTTP server for web integration
	r := gin.Default()
	http.SetupRoutes(r, messageService)
	go r.Run(":8080") // Run in background

	// Telegram polling
	bot := telegramClient.(*infrastructure.TelegramClient).Bot
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Handle regular messages
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			
			// Check for /start command
			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(chatID, "Welcome! 👋 What would you like to know about our products?\n\nChoose a category:")
				keyboard := http.CreateCategoryKeyboard()
				msg.ReplyMarkup = &keyboard
				bot.Send(msg)
				continue
			}

			// Regular text message
			msg := entities.Message{
				From:     strconv.FormatInt(update.Message.Chat.ID, 10),
				Content:  update.Message.Text,
				Platform: "telegram",
				AIContext: repo.FormatAsContext(repo.GetAllProducts()),
			}
			
			// Check if this is a calculation request (format: "30 tumbler 30kg")
			if strings.Contains(update.Message.Text, "kg") || strings.Contains(update.Message.Text, "g") {
				// Try to parse as pricing query
				parsed := pricingCalc.ParseQuery(update.Message.Text)
				result := pricingCalc.CalculatePrice(parsed)
				
				// Send calculation result with follow-up menu
				msgText := tgbotapi.NewMessage(chatID, result)
				followUpKeyboard := http.CreateFollowUpMenu()
				msgText.ReplyMarkup = &followUpKeyboard
				msgText.ParseMode = "Markdown"
				bot.Send(msgText)
				continue
			}
			
			// Regular AI query with context
			go messageService.ProcessMessageWithContext(msg)
		}

		// Handle button clicks (callback queries)
		if update.CallbackQuery != nil {
			callbackData := update.CallbackQuery.Data
			chatID := update.CallbackQuery.Message.Chat.ID
			messageID := update.CallbackQuery.Message.MessageID

			// Get or create user session
			session := sessionManager.GetOrCreateSession(chatID)

			// Handle action callbacks first
			if strings.HasPrefix(callbackData, "action_") {
				action := strings.TrimPrefix(callbackData, "action_")
				bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				
				switch action {
				case "menu":
					// Back to main menu
					editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "👋 Choose a category:")
					menuKeyboard := http.CreateCategoryKeyboard()
					editMsg.ReplyMarkup = &menuKeyboard
					bot.Send(editMsg)
					continue
				case "calculate":
					// Ask for calculation input
					editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "📝 Enter product details:\n\nFormat: `30 tumbler 30kg`\n\n(quantity product weight)")
					bot.Send(editMsg)
					continue
				case "ask":
					// Ask another question
					editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "❓ Type your question about our products:")
					bot.Send(editMsg)
					continue
				}
			}

			// Check if click is allowed (debouncing & concurrent request prevention)
			if !session.IsAllowedClick() {
				// Silently ignore spam clicks
				bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "Please wait..."))
				continue
			}

			// Mark as processing to prevent concurrent requests
			session.StartProcessing()

			// Answer callback query immediately (shows loading)
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))

			// Edit message to show loading state
			editMsg := tgbotapi.NewEditMessageText(chatID, messageID, "⏳ Loading products...")
			bot.Send(editMsg)

			var categoryFilter, typeFilter string

			// Parse callback data
			if strings.HasPrefix(callbackData, "cat_") {
				categoryFilter = strings.TrimPrefix(callbackData, "cat_")
			} else if strings.HasPrefix(callbackData, "type_") {
				typeFilter = strings.TrimPrefix(callbackData, "type_")
			}

			// Get filtered products
			var products []repository.Product
			if categoryFilter != "" && typeFilter != "" {
				products = repo.GetByTypeAndCategory(typeFilter, categoryFilter)
			} else if categoryFilter != "" {
				products = repo.GetByCategory(categoryFilter)
			} else if typeFilter != "" {
				products = repo.GetByType(typeFilter)
			}

			// Format context and send to AI
			context := repo.FormatAsContext(products)
			userMessage := "Show me " + categoryFilter + typeFilter + " products with pricing and details. Keep response concise."

			msg := entities.Message{
				From:       strconv.FormatInt(chatID, 10),
				Content:    userMessage,
				Platform:   "telegram",
				AIContext:  context,
				IsCallback: true,
			}

			// Process in goroutine with cleanup
			go func() {
				defer session.FinishProcessing()
				messageService.ProcessMessageWithContext(msg)
			}()
		}
	}
}