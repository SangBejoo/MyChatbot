package http

import (
	"net/http"
	"project_masAde/internal/entities"
	"project_masAde/internal/usecases"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

type Handler struct {
	messageService   *usecases.MessageService
	dashboardUsecase *usecases.DashboardUsecase
}

func NewHandler(service *usecases.MessageService, dashboard *usecases.DashboardUsecase) *Handler {
	return &Handler{
		messageService:   service,
		dashboardUsecase: dashboard,
	}
}

func SetupRoutes(r *gin.Engine, service *usecases.MessageService, auth *usecases.AuthUsecase, dashboard *usecases.DashboardUsecase, middleware *Middleware) {
	h := NewHandler(service, dashboard)
	
	// Apply CORS
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
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			token, err := auth.Login(loginReq.Username, loginReq.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": token})
		})
	}
	
	// Protected Dashboard Routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	api.Use(middleware.RateLimitPerUser(5, 10))
	{
		api.GET("/dashboard/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "user": c.GetFloat64("user_id")})
		})
		
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
	}

	
	// WhatsApp Management Routes (Protected in future, currently public for QR access)
	wa := r.Group("/whatsapp")
	{
		wa.GET("/login", h.ServeLoginPage) // HTML Page
		wa.GET("/qr", h.GetQRCode)         // Returns PNG
		wa.GET("/status", h.GetStatus)     // JSON Status
		wa.POST("/logout", h.Logout)
	}
}

func (h *Handler) GetQRCode(c *gin.Context) {
	waClient := h.messageService.WhatsAppClient
	if waClient == nil {
		c.String(http.StatusServiceUnavailable, "WhatsApp client not initialized")
		return
	}

	qrCodeString := waClient.GetQR()
	if qrCodeString == "" {
		if waClient.IsLoggedIn() {
			c.String(http.StatusOK, "Already logged in")
			return
		}
		c.String(http.StatusServiceUnavailable, "QR code not yet available. Please wait or check logs.")
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

func (h *Handler) GetStatus(c *gin.Context) {
	waClient := h.messageService.WhatsAppClient
	if waClient == nil {
		c.JSON(http.StatusOK, gin.H{"connected": false, "error": "Client not initialized"})
		return
	}
	
	phone, name := waClient.GetUserInfo()
	c.JSON(http.StatusOK, gin.H{
		"connected": waClient.IsLoggedIn(),
		"phone":     phone,
		"name":      name,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	waClient := h.messageService.WhatsAppClient
	if waClient != nil {
		waClient.Logout()
		c.JSON(http.StatusOK, gin.H{"status": "logged_out"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No client"})
	}
}

func (h *Handler) ServeLoginPage(c *gin.Context) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>WhatsApp Login</title>
		<style>
			body { font-family: sans-serif; text-align: center; padding: 50px; }
			img { border: 1px solid #ddd; padding: 10px; }
			.btn { padding: 10px 20px; background: #ff4444; color: white; text-decoration: none; border-radius: 5px; cursor: pointer; border: none; }
		</style>
	</head>
	<body>
		<h1>WhatsApp Login</h1>
		<div id="status">Checking status...</div>
		<div id="qr-container" style="display:none;">
			<p>Scan this QR Code with WhatsApp (Linked Devices):</p>
			<img src="/whatsapp/qr" alt="QR Code" />
		</div>
		<div id="logout-container" style="display:none;">
			<p>✅ Bot is Connected</p>
			<button class="btn" onclick="logout()">Logout</button>
		</div>

		<script>
			function checkStatus() {
				fetch('/whatsapp/status').then(r => r.json()).then(data => {
					document.getElementById('status').style.display = 'none';
					if (data.connected) {
						document.getElementById('logout-container').style.display = 'block';
						document.getElementById('qr-container').style.display = 'none';
					} else {
						document.getElementById('logout-container').style.display = 'none';
						document.getElementById('qr-container').style.display = 'block';
						// Refresh QR image timestamp to force browser reload
						const img = document.querySelector('#qr-container img');
						img.src = '/whatsapp/qr?t=' + new Date().getTime();
					}
				});
			}

			// Check status every 3 seconds
			setInterval(checkStatus, 3000);
			checkStatus();

			function logout() {
				fetch('/whatsapp/logout', { method: 'POST' }).then(() => checkStatus());
			}
		</script>
	</body>
	</html>
	`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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