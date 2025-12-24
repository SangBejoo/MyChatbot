package http

import (
	"net/http"
	"strconv"

	"project_masAde/internal/infrastructure"
	"project_masAde/internal/repository"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userRepo  *repository.UserRepository
	waManager *infrastructure.WhatsAppManager
}

func NewAdminHandler(userRepo *repository.UserRepository, waManager *infrastructure.WhatsAppManager) *AdminHandler {
	return &AdminHandler{
		userRepo:  userRepo,
		waManager: waManager,
	}
}

// GetStats returns platform statistics
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.userRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}
	
	// Add active WhatsApp connections count
	activeWA := 0
	if h.waManager != nil {
		activeWA = len(h.waManager.GetAllConnectedUsers())
	}
	
	c.JSON(http.StatusOK, gin.H{
		"total_users":         stats.TotalUsers,
		"active_users":        stats.ActiveUsers,
		"wa_enabled_users":    stats.WAEnabledUsers,
		"active_wa_connections": activeWA,
		"admin_count":         stats.AdminCount,
	})
}

// GetAllUsers returns list of all users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	
	// Add WhatsApp connection status for each user
	connectedUsers := make(map[int]bool)
	if h.waManager != nil {
		for _, uid := range h.waManager.GetAllConnectedUsers() {
			connectedUsers[uid] = true
		}
	}
	
	result := make([]gin.H, len(users))
	for i, u := range users {
		result[i] = gin.H{
			"id":            u.ID,
			"username":      u.Username,
			"role":          u.Role,
			"schema_name":   u.SchemaName,
			"is_active":     u.IsActive,
			"wa_enabled":    u.WAEnabled,
			"wa_connected":  connectedUsers[u.ID],
			"created_at":    u.CreatedAt,
			"daily_limit":   u.DailyLimit,
			"monthly_limit": u.MonthlyLimit,
		}
	}
	
	c.JSON(http.StatusOK, result)
}

// UpdateUserStatus enables/disables a user account
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var payload struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	// Don't allow disabling self
	currentUserID, _ := c.Get("user_id")
	if int(currentUserID.(float64)) == userID && !payload.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable your own account"})
		return
	}
	
	if err := h.userRepo.UpdateUserStatus(userID, payload.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "updated", "is_active": payload.IsActive})
}

// UpdateWAEnabled enables/disables WhatsApp for a user
func (h *AdminHandler) UpdateWAEnabled(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var payload struct {
		WAEnabled bool `json:"wa_enabled"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	if err := h.userRepo.UpdateWAEnabled(userID, payload.WAEnabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	// If disabling, disconnect their WhatsApp
	if !payload.WAEnabled && h.waManager != nil {
		h.waManager.DisconnectClient(userID)
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "updated", "wa_enabled": payload.WAEnabled})
}

// DisconnectUserWA forcefully disconnects a user's WhatsApp
func (h *AdminHandler) DisconnectUserWA(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	if h.waManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WhatsApp not configured"})
		return
	}
	
	h.waManager.DisconnectClient(userID)
	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

// UpdateUserLimits sets message quotas for a user
func (h *AdminHandler) UpdateUserLimits(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var payload struct {
		DailyLimit   int `json:"daily_limit"`
		MonthlyLimit int `json:"monthly_limit"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	// Validate limits
	if payload.DailyLimit < 0 || payload.MonthlyLimit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limits cannot be negative"})
		return
	}
	
	if err := h.userRepo.UpdateUserLimits(userID, payload.DailyLimit, payload.MonthlyLimit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update limits"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":        "updated",
		"daily_limit":   payload.DailyLimit,
		"monthly_limit": payload.MonthlyLimit,
	})
}

