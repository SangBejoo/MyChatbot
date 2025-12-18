package usecases

import (
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/interfaces"
	"project_masAde/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageService struct {
	geminiClient    interfaces.AIClient
	messengerClient interfaces.Messenger
	TelegramClient  *infrastructure.TelegramClient // Direct access for menu sending (exported)
	ProductRepo     *repository.ProductRepository
}

func NewMessageService(gemini interfaces.AIClient, messenger interfaces.Messenger) *MessageService {
	return &MessageService{
		geminiClient:    gemini,
		messengerClient: messenger,
	}
}

func (s *MessageService) ProcessMessage(msg entities.Message) error {
	// Call AI to generate response (without context)
	aiResponse, err := s.geminiClient.GenerateResponse(msg.Content)
	if err != nil {
		// Send error message to user
		return s.messengerClient.SendMessage(msg.From, "Error generating response: "+err.Error())
	}

	// Send response via messenger
	response := entities.Response{Content: aiResponse}
	return s.messengerClient.SendMessage(msg.From, response.Content)
}

func (s *MessageService) ProcessMessageWithContext(msg entities.Message) error {
	// Call AI with context from CSV data
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
		// Send error message to user
		return s.messengerClient.SendMessage(msg.From, "Error generating response: "+err.Error())
	}

	// Send response with follow-up menu if Telegram client available
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

	// Fallback: send without menu
	response := entities.Response{Content: aiResponse}
	return s.messengerClient.SendMessage(msg.From, response.Content)
}