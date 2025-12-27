package entities

type User struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	PasswordHash  string `json:"-"`
	Role          string `json:"role"`
	SchemaName    string `json:"schema_name"`    // Tenant schema
	IsActive      bool   `json:"is_active"`      // Account enabled
	WAEnabled     bool   `json:"wa_enabled"`     // WhatsApp enabled
	TelegramToken string `json:"telegram_token"` // User's Telegram bot token
	DailyLimit    int    `json:"daily_limit"`    // Max messages per day (0 = unlimited)
	MonthlyLimit  int    `json:"monthly_limit"`  // Max messages per month (0 = unlimited)
}
