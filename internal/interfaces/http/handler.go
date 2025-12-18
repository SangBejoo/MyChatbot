package http

import (
	"project_masAde/internal/entities"
	"project_masAde/internal/usecases"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	messageService *usecases.MessageService
}

func NewHandler(service *usecases.MessageService) *Handler {
	return &Handler{messageService: service}
}

func SetupRoutes(r *gin.Engine, service *usecases.MessageService) {
	h := NewHandler(service)
	r.POST("/webhook/whatsapp", h.HandleWhatsAppWebhook)
	r.POST("/webhook/web", h.HandleWebMessage)
}

func (h *Handler) HandleWhatsAppWebhook(c *gin.Context) {
	// Parse WhatsApp Business API webhook payload
	var payload struct {
		Object string `json:"object"`
		Entry  []struct {
			Changes []struct {
				Value struct {
					Messages []struct {
						From string `json:"from"`
						Text struct {
							Body string `json:"body"`
						} `json:"text"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Process each message
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, message := range change.Value.Messages {
				msg := entities.Message{
					From:     message.From,
					Content:  message.Text.Body,
					Platform: "whatsapp",
				}
				go h.messageService.ProcessMessage(msg)
			}
		}
	}

	c.JSON(200, gin.H{"status": "received"})
}

func (h *Handler) HandleWebMessage(c *gin.Context) {
	var payload struct {
		From    string `json:"from"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	msg := entities.Message{
		From:     payload.From,
		Content:  payload.Content,
		Platform: "web",
	}

	go h.messageService.ProcessMessage(msg)
	c.JSON(200, gin.H{"status": "received"})
}