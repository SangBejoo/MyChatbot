package http

import (
	"encoding/json"
	"net/http"
	"project_masAde/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetUserStats returns dashboard stats for the authenticated user's tenant
func (h *Handler) GetUserStats(c *gin.Context) {
	schema := getSchemaName(c)
	userID, _ := getUserIDAndSchema(c)
	
	// Get counts from repositories
	menus, _ := h.dashboardUsecase.GetAllMenus(schema)
	tables, _ := h.dashboardUsecase.ListTables(schema)
	configs, _ := h.dashboardUsecase.GetAllConfigs(schema)
	
	// Check WhatsApp status
	waConnected := false
	waPhone := ""
	waName := ""
	if h.waManager != nil {
		client := h.waManager.GetClient(userID)
		if client != nil && client.IsConnected() {
			waConnected = true
			waPhone = client.GetPhoneNumber()
			waName = client.GetName()
		}
	}
	
	// Get user quota limits
	user, _ := h.userRepo.GetByID(userID)
	dailyLimit := 200
	monthlyLimit := 5000
	if user != nil {
		dailyLimit = user.DailyLimit
		monthlyLimit = user.MonthlyLimit
	}
	
	// Get quota usage stats
	var quotaStats *repository.UserQuotaStatus
	if h.usageRepo != nil {
		quotaStats, _ = h.usageRepo.GetQuotaStatus(userID, dailyLimit, monthlyLimit)
	}
	
	response := gin.H{
		"menu_count":    len(menus),
		"table_count":   len(tables),
		"config_count":  len(configs),
		"wa_connected":  waConnected,
		"wa_phone":      waPhone,
		"wa_name":       waName,
		"schema_name":   schema,
	}
	
	// Add quota info if available
	if quotaStats != nil {
		response["quota"] = quotaStats
	}
	
	c.JSON(http.StatusOK, response)
}

// getSchemaName extracts schema_name from JWT context, defaults to "public"
func getSchemaName(c *gin.Context) string {
	schema, exists := c.Get("schema_name")
	if !exists || schema == nil {
		return "public"
	}
	if s, ok := schema.(string); ok && s != "" {
		return s
	}
	return "public"
}

// Config
func (h *Handler) GetAllConfigs(c *gin.Context) {
	schema := getSchemaName(c)
	configs, err := h.dashboardUsecase.GetAllConfigs(schema)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, configs)
}

func (h *Handler) SetConfig(c *gin.Context) {
	schema := getSchemaName(c)
	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Input validation
	if !ValidConfigKey(payload.Key) {
		c.JSON(400, gin.H{"error": "Invalid config key"})
		return
	}
	if !ValidateLength(payload.Value, 0, MaxConfigValLength) {
		c.JSON(400, gin.H{"error": "Config value too long"})
		return
	}
	payload.Value = SanitizeString(payload.Value)
	
	if err := h.dashboardUsecase.SetConfig(schema, payload.Key, payload.Value); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save config"})
		return
	}
	c.JSON(200, gin.H{"status": "updated"})
}

// Menus
func (h *Handler) GetAllMenus(c *gin.Context) {
	schema := getSchemaName(c)
	menus, err := h.dashboardUsecase.GetAllMenus(schema)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, menus)
}

func (h *Handler) GetMenu(c *gin.Context) {
	schema := getSchemaName(c)
	slug := c.Param("slug")
	if !ValidSlug(slug) {
		c.JSON(400, gin.H{"error": "Invalid menu slug"})
		return
	}
	menu, err := h.dashboardUsecase.GetMenu(schema, slug)
	if err != nil {
		c.JSON(404, gin.H{"error": "Menu not found"})
		return
	}
	c.JSON(200, menu)
}

func (h *Handler) CreateMenu(c *gin.Context) {
	schema := getSchemaName(c)
	var m struct {
		Slug  string          `json:"slug"`
		Title string          `json:"title"`
		Items json.RawMessage `json:"items"`
	}
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	
	// Validate inputs
	if !ValidSlug(m.Slug) {
		c.JSON(400, gin.H{"error": "Invalid slug format"})
		return
	}
	if !ValidateLength(m.Title, 1, MaxTitleLength) {
		c.JSON(400, gin.H{"error": "Invalid title length"})
		return
	}
	m.Title = SanitizeString(m.Title)
	
	if err := h.dashboardUsecase.CreateMenu(schema, &repository.Menu{Slug: m.Slug, Title: m.Title, Items: m.Items}); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create menu"})
		return
	}
	c.JSON(201, gin.H{"status": "created"})
}

func (h *Handler) UpdateMenu(c *gin.Context) {
	schema := getSchemaName(c)
	slug := c.Param("slug")
	var m struct {
		Title string          `json:"title"`
		Items json.RawMessage `json:"items"`
	}
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.dashboardUsecase.UpdateMenu(schema, &repository.Menu{Slug: slug, Title: m.Title, Items: m.Items}); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "updated"})
}

func (h *Handler) DeleteMenu(c *gin.Context) {
	schema := getSchemaName(c)
	slug := c.Param("slug")
	if err := h.dashboardUsecase.DeleteMenu(schema, slug); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}

// Products
func (h *Handler) GetAllProducts(c *gin.Context) {
	products := h.dashboardUsecase.GetAllProducts()
	c.JSON(200, products)
}

// -- Dynamic Data Handlers --

func (h *Handler) ListTables(c *gin.Context) {
	schema := getSchemaName(c)
	tables, err := h.dashboardUsecase.ListTables(schema)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, tables)
}

func (h *Handler) GetTableData(c *gin.Context) {
	schema := getSchemaName(c)
	name := c.Param("name")
	if !ValidTableName(name) {
		c.JSON(400, gin.H{"error": "Invalid table name"})
		return
	}
	data, err := h.dashboardUsecase.GetTableData(schema, name)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch data"})
		return
	}
	c.JSON(200, data)
}

func (h *Handler) ImportTable(c *gin.Context) {
	schema := getSchemaName(c)
	// Multipart form upload
	displayName := c.PostForm("display_name")
	if !ValidateLength(displayName, 1, MaxTitleLength) {
		c.JSON(400, gin.H{"error": "Invalid display name"})
		return
	}
	displayName = SanitizeString(displayName)
	
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Bad request: missing file"})
		return
	}
	defer file.Close()

	if err := h.dashboardUsecase.ImportTable(schema, displayName, file); err != nil {
		c.JSON(500, gin.H{"error": "Import failed: " + err.Error()})
		return
	}
	c.JSON(201, gin.H{"status": "imported"})
}

func (h *Handler) DeleteTable(c *gin.Context) {
	schema := getSchemaName(c)
	name := c.Param("name")
	if !ValidTableName(name) {
		c.JSON(400, gin.H{"error": "Invalid table name"})
		return
	}
	if err := h.dashboardUsecase.DeleteTable(schema, name); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete table"})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}

func (h *Handler) UpdateRow(c *gin.Context) {
	schema := getSchemaName(c)
	tableName := c.Param("name")
	var payload struct {
		RowID int                    `json:"row_id"`
		Data  map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.dashboardUsecase.UpdateRow(schema, tableName, payload.RowID, payload.Data); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "updated"})
}

func (h *Handler) DeleteRow(c *gin.Context) {
	schema := getSchemaName(c)
	tableName := c.Param("name")
	var payload struct {
		RowID int `json:"row_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.dashboardUsecase.DeleteRow(schema, tableName, payload.RowID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}
