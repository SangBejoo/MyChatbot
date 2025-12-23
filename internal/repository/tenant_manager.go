package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TenantManager struct {
	db *pgxpool.Pool
}

func NewTenantManager(db *pgxpool.Pool) *TenantManager {
	return &TenantManager{db: db}
}

// sanitizeSchemaName ensures schema name is safe for SQL
func sanitizeSchemaName(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return strings.ToLower(reg.ReplaceAllString(name, "_"))
}

// CreateTenantSchema creates a new schema for a user with all required tables
func (t *TenantManager) CreateTenantSchema(userID int) (string, error) {
	ctx := context.Background()
	schemaName := fmt.Sprintf("tenant_%d", userID)

	// Start transaction
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// Create schema
	_, err = tx.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	if err != nil {
		return "", fmt.Errorf("failed to create schema: %w", err)
	}

	// Create tenant-specific tables
	tables := []string{
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.bot_config (
				id SERIAL PRIMARY KEY,
				key VARCHAR(64) UNIQUE NOT NULL,
				value TEXT,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, schemaName),
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.menus (
				id SERIAL PRIMARY KEY,
				slug VARCHAR(64) UNIQUE NOT NULL,
				title VARCHAR(256) NOT NULL,
				items JSONB DEFAULT '[]',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, schemaName),
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.dynamic_tables (
				id SERIAL PRIMARY KEY,
				table_name VARCHAR(128) UNIQUE NOT NULL,
				display_name VARCHAR(256) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, schemaName),
	}

	for _, ddl := range tables {
		if _, err := tx.Exec(ctx, ddl); err != nil {
			return "", fmt.Errorf("failed to create table: %w", err)
		}
	}

	return schemaName, tx.Commit(ctx)
}

// DropTenantSchema removes a user's schema and all data
func (t *TenantManager) DropTenantSchema(schemaName string) error {
	ctx := context.Background()
	schemaName = sanitizeSchemaName(schemaName)
	
	_, err := t.db.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName))
	return err
}

// GetTenantConnection returns a connection with search_path set to tenant schema
func (t *TenantManager) SetSearchPath(conn *pgxpool.Conn, schemaName string) error {
	schemaName = sanitizeSchemaName(schemaName)
	_, err := conn.Exec(context.Background(), fmt.Sprintf("SET search_path TO %s, public", schemaName))
	return err
}
