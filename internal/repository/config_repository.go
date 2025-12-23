package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BotConfig struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MenuItem struct {
	Label   string `json:"label"`
	Action  string `json:"action"`
	Payload string `json:"payload"`
}

type Menu struct {
	ID        int             `json:"id"`
	Slug      string          `json:"slug"`
	Title     string          `json:"title"`
	Items     json.RawMessage `json:"items"` // Flexible JSON structure
	CreatedAt time.Time       `json:"created_at"`
}

type ConfigRepository struct {
	db *pgxpool.Pool
}

func NewConfigRepository(db *pgxpool.Pool) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// GetConfig returns a config value by key
func (r *ConfigRepository) GetConfig(key string) (string, error) {
	var value string
	err := r.db.QueryRow(context.Background(), "SELECT value FROM bot_config WHERE key=$1", key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil // Not found is not strictly an error
		}
		return "", err
	}
	return value, nil
}

// SetConfig sets a config value
func (r *ConfigRepository) SetConfig(key, value string) error {
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO bot_config (key, value, updated_at) 
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()
	`, key, value)
	return err
}

// GetAllConfigs returns all configs
func (r *ConfigRepository) GetAllConfigs() ([]BotConfig, error) {
	rows, err := r.db.Query(context.Background(), "SELECT key, value, updated_at FROM bot_config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []BotConfig
	for rows.Next() {
		var c BotConfig
		if err := rows.Scan(&c.Key, &c.Value, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// GetMenu returns a menu by slug
func (r *ConfigRepository) GetMenu(slug string) (*Menu, error) {
	var m Menu
	err := r.db.QueryRow(context.Background(), "SELECT id, slug, title, items, created_at FROM menus WHERE slug=$1", slug).Scan(&m.ID, &m.Slug, &m.Title, &m.Items, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateMenu creates a new menu
func (r *ConfigRepository) CreateMenu(m *Menu) error {
	return r.db.QueryRow(context.Background(), `
		INSERT INTO menus (slug, title, items, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`, m.Slug, m.Title, m.Items).Scan(&m.ID)
}

// UpdateMenu updates an existing menu
func (r *ConfigRepository) UpdateMenu(m *Menu) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menus SET title=$1, items=$2 WHERE slug=$3
	`, m.Title, m.Items, m.Slug)
	return err
}

// DeleteMenu deletes a menu
func (r *ConfigRepository) DeleteMenu(slug string) error {
	_, err := r.db.Exec(context.Background(), "DELETE FROM menus WHERE slug=$1", slug)
	return err
}

// GetAllMenus list all menus (lightweight)
func (r *ConfigRepository) GetAllMenus() ([]Menu, error) {
	rows, err := r.db.Query(context.Background(), "SELECT id, slug, title, items, created_at FROM menus")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.Slug, &m.Title, &m.Items, &m.CreatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}
