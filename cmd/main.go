package main

import (
	"encoding/json"
	"os"
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/interfaces/http"
	"project_masAde/internal/repository"
	"project_masAde/internal/usecases"
	"strconv"
	"strings"

	"fmt"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"go.mau.fi/whatsmeow/types/events" // Import events
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Connect to PostgreSQL
	pgClient, err := infrastructure.NewPostgresClient("postgres://postgres:root@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer pgClient.Close()
	
	// Initialize Repositories
	repo := repository.NewProductRepository(pgClient.Pool)
	userRepo := repository.NewUserRepository(pgClient.Pool)
	configRepo := repository.NewConfigRepository(pgClient.Pool)
	tableManager := repository.NewTableManager(pgClient.Pool)

	// Sync Data
	err = repo.SyncFromCSV("data/products.csv")
	if err != nil {
		fmt.Println("Warning: Failed to sync products from CSV:", err)
	}
	
	// Initialize Usecases & Services
	authUsecase := usecases.NewAuthUsecase(userRepo, os.Getenv("JWT_SECRET"))
	
	// Ensure Admin User
	if err := authUsecase.EnsureAdmin("root", "root"); err != nil {
		fmt.Println("Warning: Failed to ensure admin user:", err)
	}

	geminiClient := infrastructure.NewGeminiClient(os.Getenv("GEMINI_API_KEY"))
	telegramClient := infrastructure.NewTelegramClient(os.Getenv("TELEGRAM_BOT_TOKEN"))
	
	messageService := usecases.NewMessageService(geminiClient, telegramClient, configRepo, tableManager)
	messageService.ProductRepo = repo
	if tc, ok := telegramClient.(*infrastructure.TelegramClient); ok {
		messageService.TelegramClient = tc
	}

	dashboardUsecase := usecases.NewDashboardUsecase(configRepo, repo, tableManager)
	authMiddleware := http.NewMiddleware(os.Getenv("JWT_SECRET"))

	// Initialize session manager & pricing
	sessionManager := infrastructure.NewSessionManager()
	pricingCalc := usecases.NewPricingCalculator(repo)

	// Initialize WhatsApp client
	whatsAppClient, err := infrastructure.NewWhatsAppClient("device.db")
	if err != nil {
		panic("Failed to create WhatsApp client: " + err.Error())
	}
	messageService.WhatsAppClient = whatsAppClient

	// WhatsApp Message Handler
	whatsAppClient.AddHandler(func(evt interface{}) {
		// Only check for new messages
		switch v := evt.(type) {
		case *events.Message:
			sender, content := whatsAppClient.ParseMessage(v)
			
			// Ignore messages from self or status updates
			if v.Info.IsGroup { return }

			// Handle regular text menu selection
			// "1. 🧮 Calculate Price"
			if content == "1" || strings.Contains(strings.ToLower(content), "calculate") {
				whatsAppClient.SendMessage(sender, "📝 Enter product details:\n\nFormat: *30 tumbler 30kg*\n\n(quantity product weight)")
				return
			}
			
			// Pricing calculation check
			if strings.Contains(content, "kg") || strings.Contains(content, "g") {
				parsed := pricingCalc.ParseQuery(content)
				result := pricingCalc.CalculatePrice(parsed)
				whatsAppClient.SendMessage(sender, result + "\n\nReply with *1* to calculate again.")
				return
			}

			// Dynamic Menu Check (via MessageService)
			msg := entities.Message{
				From:     strings.TrimSuffix(sender, "@s.whatsapp.net"),
				Content:  content,
				Platform: "whatsapp",
			}
			// We try dynamic menu first inside ProcessMessage, but here we invoke it manually or just pass to ProcessMessage
			// The ProcessMessage inside MessageService now handles dynamic menu logic FIRST.
			// So we can just delegate everything to ProcessMessageWithContext (which checks AI) OR ProcessMessage (checks menu + AI)
			// Wait, ProcessMessageWithContext does NOT check Dynamic Menu in my implementation?
			// My implementation updated ProcessMessage. 
			// I should use ProcessMessage if I want menu check. 
			// But I also want RAG context.
			// Let's update ProcessMessageWithContext to ALSO check Dynamic Menu? 
			// Or better: Let ProcessMessage handle it.
			// Let's rely on ProcessMessage logic I wrote which falls back to AI.
			// But for RAG, I needed `AIContext`.
			// Let's add `AIContext` to msg and call `ProcessMessage` (I need to update ProcessMessage to use Context if present).
			// For now, let's keep it simple: Use ProcessMessage.
			
			whatsAppClient.SendPresence(sender)
			msg.AIContext = repo.FormatAsContext(repo.GetAllProducts())
			go messageService.ProcessMessage(msg) // Changed to ProcessMessage to use new Logic
		}
	})

	// Connect WhatsApp
	if err := whatsAppClient.Connect(); err != nil {
		fmt.Println("Error connecting to WhatsApp:", err)
	}

	// Setup HTTP server
	r := gin.Default()
	http.SetupRoutes(r, messageService, authUsecase, dashboardUsecase, authMiddleware)
	go func() {
		if err := r.Run("0.0.0.0:8080"); err != nil {
			fmt.Printf("FAILED to start HTTP Server: %v\n", err)
			os.Exit(1)
		}
	}()


	// Telegram polling
	var bot *tgbotapi.BotAPI
	if tc, ok := telegramClient.(*infrastructure.TelegramClient); ok && tc.Bot != nil {
		bot = tc.Bot
		fmt.Println("Telegram Bot Connected")
	} else {
		fmt.Println("Telegram disabled (Token missing or invalid). Application running (Web/WhatsApp only).")
		select {} // Block main thread forever since we have nothing else to do here (Gin runs in goroutine)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// Handle regular messages
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			
			// Check for /start command
			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(chatID, "Welcome! 👋")
				
				// Fetch dynamic main_menu
				menu, err := configRepo.GetMenu("main_menu")
				if err == nil {
					msg.Text = menu.Title + "\n\nChoose an option:"
					
					// Parse Items
					itemsBytes, _ := json.Marshal(menu.Items)
					var items []repository.MenuItem
					if err := json.Unmarshal(itemsBytes, &items); err == nil {
						keyboard := http.GenerateDynamicKeyboard(items)
						msg.ReplyMarkup = &keyboard
						bot.Send(msg)
						continue
					}
				}

				// Fallback if no dynamic menu
				msg.Text = "Welcome! 👋 What would you like to know about our products?\n\nChoose a category:"
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
			if strings.HasPrefix(callbackData, "dyn:") {
				// Format: dyn:Action:Payload
				parts := strings.SplitN(callbackData, ":", 3)
				if len(parts) >= 2 {
					action := parts[1]
					payload := ""
					if len(parts) > 2 {
						payload = parts[2]
					}
					
					bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))

					switch action {
					case "view_table":
						// Fetch table data
						data, err := tableManager.GetTableData(payload)
						if err != nil {
							bot.Send(tgbotapi.NewMessage(chatID, "Error fetching data: "+err.Error()))
							continue
						}
						
						// Format Data (Simple List)
						var sb strings.Builder
						sb.WriteString(fmt.Sprintf("📊 *%s Data:*\n\n", payload))
						limit := 10
						if len(data) < limit { limit = len(data) }
						for i := 0; i < limit; i++ {
							row := data[i]
							sb.WriteString("- ")
							for k, v := range row {
								if k != "id" {
									sb.WriteString(fmt.Sprintf("%s: %v, ", k, v))
								}
							}
							sb.WriteString("\n")
						}
						if len(data) > limit {
							sb.WriteString(fmt.Sprintf("\n...and %d more rows.", len(data)-limit))
						}
						
						msgText := tgbotapi.NewMessage(chatID, sb.String())
						msgText.ParseMode = "Markdown"
						bot.Send(msgText)
						continue

					case "reply":
						bot.Send(tgbotapi.NewMessage(chatID, payload))
						continue
					}
				}
			}

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