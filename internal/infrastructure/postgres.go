package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	Pool *pgxpool.Pool
}

func NewPostgresClient(connString string) (*PostgresClient, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Pool configuration
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	client := &PostgresClient{Pool: pool}
	
	// Auto-migrate schema
	if err := client.Migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return client, nil
}

func (p *PostgresClient) Migrate() error {
	ctx := context.Background()

	// Users Table
	_, err := p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) DEFAULT 'user',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	// Products Table (replacing CSV)
	_, err = p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			code VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(50),
			type VARCHAR(20),            -- New: export/import
			raw_price VARCHAR(50),       -- New: stored string price from CSV
			currency VARCHAR(10),        -- New: IDR/USD
			price_min DECIMAL(15, 2),
			price_max DECIMAL(15, 2),
			min_order_qty INT,
			unit VARCHAR(20),
			weight_kg DECIMAL(10, 3) DEFAULT 0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create products table: %w", err)
	}
	
	// Add columns if they don't exist (migrations for existing table)
	// Simple dirty migration for dev environment
	p.Pool.Exec(ctx, "ALTER TABLE products ADD COLUMN IF NOT EXISTS type VARCHAR(20);")
	p.Pool.Exec(ctx, "ALTER TABLE products ADD COLUMN IF NOT EXISTS raw_price VARCHAR(50);")
	p.Pool.Exec(ctx, "ALTER TABLE products ADD COLUMN IF NOT EXISTS currency VARCHAR(10);")

	// Dynamic Tables Registry
	_, err = p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS dynamic_tables (
			id SERIAL PRIMARY KEY,
			table_name VARCHAR(255) UNIQUE NOT NULL,
			display_name VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create dynamic_tables registry: %w", err)
	}

	// Seed Admin User (root/root) if not exists
	// Password is bcrypt hash of "root"
	// Cost: 10, Hash: $2a$10$tM.y.y... (generated for 'root')
	// We'll generate a fresh one in the auth service, but let's seed a basic one only if table empty
	var count int
	err = p.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// "root" hash (bcrypt cost 10) for initial seeding
		// This logic effectively defers proper user creation to the AuthUsecase.EnsureAdmin call in main.go
		// which handles the hashing correctly. We just log here.
		log.Println("Database initialized. Users table empty. Admin will be ensured by application logic.")
	}

	// Bot Configuration Table
	_, err = p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bot_config (
			key VARCHAR(50) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create bot_config table: %w", err)
	}

	// Dynamic Menus Table
	_, err = p.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS menus (
			id SERIAL PRIMARY KEY,
			slug VARCHAR(50) UNIQUE NOT NULL,
			title VARCHAR(100) NOT NULL,
			items JSONB NOT NULL, -- structured menu options
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("create menus table: %w", err)
	}

	return nil
}

func (p *PostgresClient) Close() {
	p.Pool.Close()
}
