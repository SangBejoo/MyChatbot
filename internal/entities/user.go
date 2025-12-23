package entities

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	SchemaName   string `json:"schema_name"` // Tenant schema
	IsActive     bool   `json:"is_active"`   // Account enabled
	WAEnabled    bool   `json:"wa_enabled"`  // WhatsApp enabled
}
