package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"project_masAde/internal/entities"
	"project_masAde/internal/interfaces"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GeminiClient struct {
	apiKey string
}

func NewGeminiClient(apiKey string) interfaces.AIClient {
	return &GeminiClient{apiKey: apiKey}
}

func (g *GeminiClient) GenerateResponse(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", g.apiKey)
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("http post error: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal error: %w", err)
	}

	if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{}); ok {
			if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
				if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
					return text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no text found in response")
}

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