package repository

import (
	"context"
	"project_masAde/internal/entities"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

type PlatformStats struct {
	TotalUsers      int `json:"total_users"`
	ActiveUsers     int `json:"active_users"`
	WAEnabledUsers  int `json:"wa_enabled_users"`
	AdminCount      int `json:"admin_count"`
}

type UserListItem struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	SchemaName   string    `json:"schema_name"`
	IsActive     bool      `json:"is_active"`
	WAEnabled    bool      `json:"wa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	DailyLimit   int       `json:"daily_limit"`
	MonthlyLimit int       `json:"monthly_limit"`
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *entities.User) (int, error) {
	var id int
	err := r.db.QueryRow(context.Background(),
		"INSERT INTO users (username, password_hash, role, schema_name, is_active, wa_enabled) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		user.Username, user.PasswordHash, user.Role, user.SchemaName, true, true).Scan(&id)
	return id, err
}

func (r *UserRepository) GetByUsername(username string) (*entities.User, error) {
	var user entities.User
	var schemaName *string
	var isActive, waEnabled *bool
	var dailyLimit, monthlyLimit *int
	err := r.db.QueryRow(context.Background(),
		`SELECT id, username, password_hash, role, schema_name, 
		 COALESCE(is_active, true), COALESCE(wa_enabled, true),
		 COALESCE(daily_limit, 200), COALESCE(monthly_limit, 5000)
		 FROM users WHERE username = $1`,
		username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &schemaName, &isActive, &waEnabled, &dailyLimit, &monthlyLimit)
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if schemaName != nil {
		user.SchemaName = *schemaName
	}
	if isActive != nil {
		user.IsActive = *isActive
	} else {
		user.IsActive = true
	}
	if waEnabled != nil {
		user.WAEnabled = *waEnabled
	} else {
		user.WAEnabled = true
	}
	if dailyLimit != nil {
		user.DailyLimit = *dailyLimit
	} else {
		user.DailyLimit = 200
	}
	if monthlyLimit != nil {
		user.MonthlyLimit = *monthlyLimit
	} else {
		user.MonthlyLimit = 5000
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id int) (*entities.User, error) {
	var user entities.User
	var schemaName *string
	var isActive, waEnabled *bool
	var dailyLimit, monthlyLimit *int
	err := r.db.QueryRow(context.Background(),
		`SELECT id, username, password_hash, role, schema_name, 
		 COALESCE(is_active, true), COALESCE(wa_enabled, true),
		 COALESCE(daily_limit, 200), COALESCE(monthly_limit, 5000)
		 FROM users WHERE id = $1`,
		id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &schemaName, &isActive, &waEnabled, &dailyLimit, &monthlyLimit)
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if schemaName != nil {
		user.SchemaName = *schemaName
	}
	if isActive != nil {
		user.IsActive = *isActive
	} else {
		user.IsActive = true
	}
	if waEnabled != nil {
		user.WAEnabled = *waEnabled
	} else {
		user.WAEnabled = true
	}
	if dailyLimit != nil {
		user.DailyLimit = *dailyLimit
	} else {
		user.DailyLimit = 200
	}
	if monthlyLimit != nil {
		user.MonthlyLimit = *monthlyLimit
	} else {
		user.MonthlyLimit = 5000
	}
	return &user, nil
}

func (r *UserRepository) UpdateSchemaName(userID int, schemaName string) error {
	_, err := r.db.Exec(context.Background(),
		"UPDATE users SET schema_name = $1 WHERE id = $2",
		schemaName, userID)
	return err
}

// Admin methods

func (r *UserRepository) GetAllUsers() ([]UserListItem, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT id, username, role, COALESCE(schema_name, ''), COALESCE(is_active, true), COALESCE(wa_enabled, true), COALESCE(created_at, NOW()), COALESCE(daily_limit, 200), COALESCE(monthly_limit, 5000) 
		 FROM users ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	users := []UserListItem{}
	for rows.Next() {
		var u UserListItem
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.SchemaName, &u.IsActive, &u.WAEnabled, &u.CreatedAt, &u.DailyLimit, &u.MonthlyLimit); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) GetStats() (*PlatformStats, error) {
	var stats PlatformStats
	
	err := r.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}
	
	err = r.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE COALESCE(is_active, true) = true").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}
	
	err = r.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE COALESCE(wa_enabled, true) = true").Scan(&stats.WAEnabledUsers)
	if err != nil {
		return nil, err
	}
	
	err = r.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&stats.AdminCount)
	if err != nil {
		return nil, err
	}
	
	return &stats, nil
}

func (r *UserRepository) UpdateUserStatus(userID int, isActive bool) error {
	_, err := r.db.Exec(context.Background(),
		"UPDATE users SET is_active = $1 WHERE id = $2",
		isActive, userID)
	return err
}

func (r *UserRepository) UpdateWAEnabled(userID int, enabled bool) error {
	_, err := r.db.Exec(context.Background(),
		"UPDATE users SET wa_enabled = $1 WHERE id = $2",
		enabled, userID)
	return err
}

func (r *UserRepository) UpdateUserLimits(userID int, dailyLimit, monthlyLimit int) error {
	_, err := r.db.Exec(context.Background(),
		"UPDATE users SET daily_limit = $1, monthly_limit = $2 WHERE id = $3",
		dailyLimit, monthlyLimit, userID)
	return err
}

// UpdateTelegramToken updates user's Telegram bot token
func (r *UserRepository) UpdateTelegramToken(userID int, token string) error {
	_, err := r.db.Exec(context.Background(),
		"UPDATE users SET telegram_token = $1 WHERE id = $2",
		token, userID)
	return err
}

// GetTelegramToken returns user's Telegram bot token
func (r *UserRepository) GetTelegramToken(userID int) (string, error) {
	var token *string
	err := r.db.QueryRow(context.Background(),
		"SELECT telegram_token FROM users WHERE id = $1",
		userID).Scan(&token)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", nil
	}
	return *token, nil
}

