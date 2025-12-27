package http

import (
	"fmt"
	"net/http"
	"project_masAde/internal/entities"
	"project_masAde/internal/infrastructure"
	"project_masAde/internal/repository"
	"project_masAde/internal/usecases"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type Handler struct {
	messageService   *usecases.MessageService
	dashboardUsecase *usecases.DashboardUsecase
	waManager        *infrastructure.WhatsAppManager
	usageRepo        *repository.UsageRepository
	userRepo         *repository.UserRepository
}

func NewHandler(service *usecases.MessageService, dashboard *usecases.DashboardUsecase, waManager *infrastructure.WhatsAppManager, usageRepo *repository.UsageRepository, userRepo *repository.UserRepository) *Handler {
	return &Handler{
		messageService:   service,
		dashboardUsecase: dashboard,
		waManager:        waManager,
		usageRepo:        usageRepo,
		userRepo:         userRepo,
	}
}

func SetupRoutes(r *gin.Engine, service *usecases.MessageService, auth *usecases.AuthUsecase, dashboard *usecases.DashboardUsecase, waManager *infrastructure.WhatsAppManager, tgManager *infrastructure.TelegramBotManager, userRepo *repository.UserRepository, usageRepo *repository.UsageRepository, middleware *Middleware) {
	h := NewHandler(service, dashboard, waManager, usageRepo, userRepo)
	adminHandler := NewAdminHandler(userRepo, waManager)
	telegramHandler := NewTelegramHandler(tgManager, userRepo)
	
	// Apply Security Middleware
	r.Use(SecurityHeaders())
	r.Use(RequestSizeLimiter(10 << 20)) // 10MB max request size
	r.Use(middleware.CORSMiddleware())
	
	// Public Routes
	r.POST("/webhook/web", h.HandleWebMessage)
	
	// Public Auth Routes
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) {
			var loginReq struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&loginReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			token, err := auth.Login(loginReq.Username, loginReq.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": token})
		})
		
		authGroup.POST("/register", func(c *gin.Context) {
			var regReq struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := c.ShouldBindJSON(&regReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			// Validate inputs
			if !ValidSlug(regReq.Username) || len(regReq.Password) < 6 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password (min 6 chars)"})
				return
			}
			if err := auth.Register(regReq.Username, regReq.Password); err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"status": "registered"})
		})
	}
	
	// Protected Dashboard Routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	api.Use(middleware.RateLimitPerUser(5, 10))
	{
		api.GET("/dashboard/stats", h.GetUserStats)
		
		// Config Routes
		api.GET("/config", h.GetAllConfigs)
		api.POST("/config", h.SetConfig)
		
		// Menu Routes
		api.GET("/menus", h.GetAllMenus)
		api.GET("/menus/:slug", h.GetMenu)
		api.POST("/menus", h.CreateMenu)
		api.PUT("/menus/:slug", h.UpdateMenu)
		api.DELETE("/menus/:slug", h.DeleteMenu)
		
		// Product Routes (JSON)
		api.GET("/products", h.GetAllProducts)
		
		// Dynamic Table Routes
		api.GET("/tables", h.ListTables)
		api.GET("/tables/:name/data", h.GetTableData)
		api.POST("/tables/import", h.ImportTable)
		api.DELETE("/tables/:name", h.DeleteTable)
		api.PUT("/tables/:name/row", h.UpdateRow)
		api.DELETE("/tables/:name/row", h.DeleteRow)
		
		// WhatsApp Management Routes - DISABLED (using Telegram)
		// api.GET("/whatsapp/qr", h.GetUserQRCode)
		// api.GET("/whatsapp/status", h.GetUserWhatsAppStatus)
		// api.POST("/whatsapp/connect", h.ConnectUserWhatsApp)
		// api.POST("/whatsapp/logout", h.LogoutUserWhatsApp)
		
		// Telegram Management Routes (per-user bots)
		telegramHandler.RegisterRoutes(api)
	}
	
	// Admin-only Routes
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/users", adminHandler.GetAllUsers)
		admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
		admin.PUT("/users/:id/whatsapp", adminHandler.UpdateWAEnabled)
		admin.PUT("/users/:id/limits", adminHandler.UpdateUserLimits)
		admin.POST("/users/:id/disconnect-wa", adminHandler.DisconnectUserWA)
	}
}
// ========================================
// Per-User WhatsApp Handlers
// ========================================

// getUserIDAndSchema extracts user_id and schema_name from JWT context
func getUserIDAndSchema(c *gin.Context) (int, string) {
	userIDFloat, _ := c.Get("user_id")
	schema := getSchemaName(c)
	
	userID := 0
	if uid, ok := userIDFloat.(float64); ok {
		userID = int(uid)
	}
	return userID, schema
}

// ConnectUserWhatsApp creates and connects WhatsApp client for the user
func (h *Handler) ConnectUserWhatsApp(c *gin.Context) {
	userID, schema := getUserIDAndSchema(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}
	
	if h.waManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WhatsApp not configured"})
		return
	}
	
	client, err := h.waManager.ConnectClient(userID, schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Return initial status
	phone, name := client.GetUserInfo()
	c.JSON(http.StatusOK, gin.H{
		"status":    "connecting",
		"connected": client.IsLoggedIn(),
		"phone":     phone,
		"name":      name,
	})
}

// GetUserQRCode returns QR code PNG for user's WhatsApp
func (h *Handler) GetUserQRCode(c *gin.Context) {
	userID, schema := getUserIDAndSchema(c)
	if userID == 0 {
		c.String(http.StatusUnauthorized, "Invalid user")
		return
	}
	
	if h.waManager == nil {
		c.String(http.StatusServiceUnavailable, "WhatsApp not configured")
		return
	}
	
	// Get or create client
	client, err := h.waManager.GetOrCreateClient(userID, schema)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create client: "+err.Error())
		return
	}
	
	// Connect if not already
	if client.Client.Store.ID == nil && !client.Client.IsConnected() {
		if err := client.Connect(); err != nil {
			c.String(http.StatusInternalServerError, "Failed to connect: "+err.Error())
			return
		}
	}
	
	qrCodeString := client.GetQR()
	if qrCodeString == "" {
		if client.IsLoggedIn() {
			c.String(http.StatusOK, "Already logged in")
			return
		}
		c.String(http.StatusAccepted, "QR code not yet available. Please wait...")
		return
	}
	
	// Generate PNG
	png, err := qrcode.Encode(qrCodeString, qrcode.Medium, 256)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to generate QR code")
		return
	}
	
	c.Data(http.StatusOK, "image/png", png)
}

// GetUserWhatsAppStatus returns WhatsApp connection status for user
func (h *Handler) GetUserWhatsAppStatus(c *gin.Context) {
	userID, schema := getUserIDAndSchema(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}
	
	if h.waManager == nil {
		c.JSON(http.StatusOK, gin.H{"connected": false, "error": "WhatsApp not configured"})
		return
	}
	
	client := h.waManager.GetClient(userID)
	if client == nil {
		// Try to get or create (but don't connect)
		client, _ = h.waManager.GetOrCreateClient(userID, schema)
	}
	
	if client == nil {
		c.JSON(http.StatusOK, gin.H{"connected": false, "initialized": false})
		return
	}
	
	phone, name := client.GetUserInfo()
	c.JSON(http.StatusOK, gin.H{
		"connected":   client.IsLoggedIn(),
		"initialized": true,
		"phone":       phone,
		"name":        name,
		"hasQR":       client.GetQR() != "",
	})
}

// LogoutUserWhatsApp logs out user's WhatsApp session
func (h *Handler) LogoutUserWhatsApp(c *gin.Context) {
	userID, _ := getUserIDAndSchema(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}
	
	if h.waManager == nil {
		c.JSON(http.StatusOK, gin.H{"status": "logged_out", "message": "WhatsApp not configured"})
		return
	}
	
	// Attempt logout - errors are logged but not returned to user
	if err := h.waManager.LogoutClient(userID); err != nil {
		// Log the error but return success to user (already logged out)
		fmt.Printf("WhatsApp logout warning for user %d: %v\n", userID, err)
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "logged_out"})
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