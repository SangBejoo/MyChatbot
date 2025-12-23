package repository

import (
	"context"
	"encoding/json"
	"fmt"
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

// qualifyConfigTable returns schema-qualified table name
func qualifyConfigTable(schema, table string) string {
	if schema == "" || schema == "public" {
		return table
	}
	return fmt.Sprintf("%s.%s", schema, table)
}

// GetConfig returns a config value by key (schema-aware)
func (r *ConfigRepository) GetConfig(schemaName, key string) (string, error) {
	table := qualifyConfigTable(schemaName, "bot_config")
	var value string
	err := r.db.QueryRow(context.Background(), fmt.Sprintf("SELECT value FROM %s WHERE key=$1", table), key).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil // Not found is not strictly an error
		}
		return "", err
	}
	return value, nil
}

// SetConfig sets a config value (schema-aware)
func (r *ConfigRepository) SetConfig(schemaName, key, value string) error {
	table := qualifyConfigTable(schemaName, "bot_config")
	_, err := r.db.Exec(context.Background(), fmt.Sprintf(`
		INSERT INTO %s (key, value, updated_at) 
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()
	`, table), key, value)
	return err
}

// GetAllConfigs returns all configs (schema-aware)
func (r *ConfigRepository) GetAllConfigs(schemaName string) ([]BotConfig, error) {
	table := qualifyConfigTable(schemaName, "bot_config")
	rows, err := r.db.Query(context.Background(), fmt.Sprintf("SELECT key, value, updated_at FROM %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := []BotConfig{}
	for rows.Next() {
		var c BotConfig
		if err := rows.Scan(&c.Key, &c.Value, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// GetMenu returns a menu by slug (schema-aware)
func (r *ConfigRepository) GetMenu(schemaName, slug string) (*Menu, error) {
	table := qualifyConfigTable(schemaName, "menus")
	var m Menu
	err := r.db.QueryRow(context.Background(), fmt.Sprintf("SELECT id, slug, title, items, created_at FROM %s WHERE slug=$1", table), slug).Scan(&m.ID, &m.Slug, &m.Title, &m.Items, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateMenu creates a new menu (schema-aware)
func (r *ConfigRepository) CreateMenu(schemaName string, m *Menu) error {
	table := qualifyConfigTable(schemaName, "menus")
	return r.db.QueryRow(context.Background(), fmt.Sprintf(`
		INSERT INTO %s (slug, title, items, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`, table), m.Slug, m.Title, m.Items).Scan(&m.ID)
}

// UpdateMenu updates an existing menu (schema-aware)
func (r *ConfigRepository) UpdateMenu(schemaName string, m *Menu) error {
	table := qualifyConfigTable(schemaName, "menus")
	_, err := r.db.Exec(context.Background(), fmt.Sprintf(`
		UPDATE %s SET title=$1, items=$2 WHERE slug=$3
	`, table), m.Title, m.Items, m.Slug)
	return err
}

// DeleteMenu deletes a menu (schema-aware)
func (r *ConfigRepository) DeleteMenu(schemaName, slug string) error {
	table := qualifyConfigTable(schemaName, "menus")
	_, err := r.db.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s WHERE slug=$1", table), slug)
	return err
}

// GetAllMenus list all menus (schema-aware)
func (r *ConfigRepository) GetAllMenus(schemaName string) ([]Menu, error) {
	table := qualifyConfigTable(schemaName, "menus")
	rows, err := r.db.Query(context.Background(), fmt.Sprintf("SELECT id, slug, title, items, created_at FROM %s", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus := []Menu{}
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.Slug, &m.Title, &m.Items, &m.CreatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	return menus, nil
}
