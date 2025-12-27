package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"project_masAde/internal/entities"
	"project_masAde/internal/interfaces"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GeminiClient removed - AI handled by separate microservice
// See: AI Service project for AI integration

type WhatsAppBusinessClient struct {
	accessToken   string
	phoneNumberID string
}

func NewWhatsAppBusinessClient(accessToken, phoneNumberID string) interfaces.Messenger {
	return &WhatsAppBusinessClient{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
	}
}

func (w *WhatsAppBusinessClient) SendMessage(to, content string) error {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", w.phoneNumberID)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": content,
		},
	}
	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+w.accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (w *WhatsAppBusinessClient) ReceiveMessage() (entities.Message, error) {
	// Webhook-based; handled in handler
	return entities.Message{}, nil
}

type TelegramClient struct {
	Bot *tgbotapi.BotAPI
}

func NewTelegramClient(token string) interfaces.Messenger {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		fmt.Printf("Warning: Telegram Bot Token issue: %v. Telegram features disabled.\n", err)
		return &TelegramClient{Bot: nil}
	}
	return &TelegramClient{Bot: bot}
}

func (t *TelegramClient) SendMessage(to, content string) error {
	chatID, _ := strconv.ParseInt(to, 10, 64)
	msg := tgbotapi.NewMessage(chatID, content)
	msg.ParseMode = "Markdown"
	_, err := t.Bot.Send(msg)
	return err
}

// SendMessageWithMenu sends message with inline keyboard menu
func (t *TelegramClient) SendMessageWithMenu(to, content string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	chatID, _ := strconv.ParseInt(to, 10, 64)
	msg := tgbotapi.NewMessage(chatID, content)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	_, err := t.Bot.Send(msg)
	return err
}

func (t *TelegramClient) ReceiveMessage() (entities.Message, error) {
	// Polling-based; handled in main loop
	return entities.Message{}, nil
}