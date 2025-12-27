package http

import (
	"net/http"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/repository"

	"github.com/gin-gonic/gin"
)

// TelegramHandler handles Telegram bot management endpoints
type TelegramHandler struct {
	tgManager *infrastructure.TelegramBotManager
	userRepo  *repository.UserRepository
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(tgManager *infrastructure.TelegramBotManager, userRepo *repository.UserRepository) *TelegramHandler {
	return &TelegramHandler{
		tgManager: tgManager,
		userRepo:  userRepo,
	}
}

// RegisterRoutes registers Telegram management routes
func (h *TelegramHandler) RegisterRoutes(api *gin.RouterGroup) {
	tg := api.Group("/telegram")
	{
		tg.GET("/status", h.GetStatus)
		tg.POST("/token", h.SaveToken)
		tg.POST("/connect", h.Connect)
		tg.POST("/disconnect", h.Disconnect)
		tg.POST("/validate", h.ValidateToken)
	}
}

// GetStatus returns the connection status of user's Telegram bot
func (h *TelegramHandler) GetStatus(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get token from DB
	token, err := h.userRepo.GetTelegramToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token"})
		return
	}

	connected, botName := h.tgManager.GetStatus(userID)

	c.JSON(http.StatusOK, gin.H{
		"has_token":  token != "",
		"connected":  connected,
		"bot_name":   botName,
	})
}

// SaveToken saves the user's Telegram bot token
func (h *TelegramHandler) SaveToken(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate token first
	if req.Token != "" {
		botName, err := h.tgManager.ValidateToken(req.Token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}
		
		// Save to DB
		if err := h.userRepo.UpdateTelegramToken(userID, req.Token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status":   "saved",
			"bot_name": botName,
		})
		return
	}

	// Clear token
	if err := h.userRepo.UpdateTelegramToken(userID, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}

// ValidateToken checks if a token is valid without saving
func (h *TelegramHandler) ValidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	botName, err := h.tgManager.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"bot_name": "@" + botName,
	})
}

// Connect starts the user's Telegram bot
func (h *TelegramHandler) Connect(c *gin.Context) {
	userID := getUserID(c)
	schema := getSchemaName(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get token from DB
	token, err := h.userRepo.GetTelegramToken(userID)
	if err != nil || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token configured. Please save your bot token first."})
		return
	}

	// Connect bot
	instance, err := h.tgManager.ConnectBot(userID, schema, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "connected",
		"bot_name": "@" + instance.Bot.Self.UserName,
	})
}

// Disconnect stops the user's Telegram bot
func (h *TelegramHandler) Disconnect(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	h.tgManager.DisconnectBot(userID)

	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

// getUserID extracts user ID from JWT context
func getUserID(c *gin.Context) int {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if uid, ok := userIDFloat.(float64); ok {
		return int(uid)
	}
	return 0
}
