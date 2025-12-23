package repository

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID        string
	Name      string
	Category  string
	Type      string // "export" or "import"
	Price     string
	Currency  string
	Details   string
}

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// SyncFromCSV loads products from a CSV file and upserts them into Postgres
func (r *ProductRepository) SyncFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	ctx := context.Background()

	// Skip header row
	for i := 1; i < len(records); i++ {
		if len(records[i]) >= 7 {
			code := records[i][0]
			name := records[i][1]
			category := records[i][2]
			pType := records[i][3]
			price := records[i][4]
			currency := records[i][5]
			details := records[i][6]

			_, err := r.db.Exec(ctx, `
				INSERT INTO products (code, name, category, description, type, raw_price, currency) 
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (code) DO UPDATE 
				SET name = EXCLUDED.name, 
				    category = EXCLUDED.category, 
					description = EXCLUDED.description,
					type = EXCLUDED.type,
					raw_price = EXCLUDED.raw_price,
					currency = EXCLUDED.currency;
			`, code, name, category, details, pType, price, currency)
			
			if err != nil {
				fmt.Printf("Failed to sync product %s: %v\n", code, err)
			}
		}
	}
	return nil
}

// GetByCategory returns all products in a category
func (r *ProductRepository) GetByCategory(category string) []Product {
	rows, err := r.db.Query(context.Background(), "SELECT code, name, category, description, type, raw_price, currency FROM products WHERE category ILIKE $1", category)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()

	return scanProducts(rows)
}

// GetByType returns all products of a type (export/import)
func (r *ProductRepository) GetByType(typeStr string) []Product {
	rows, err := r.db.Query(context.Background(), "SELECT code, name, category, description, type, raw_price, currency FROM products WHERE type ILIKE $1", typeStr)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	return scanProducts(rows)
}

// GetByTypeAndCategory returns filtered products
func (r *ProductRepository) GetByTypeAndCategory(typeStr, category string) []Product {
	rows, err := r.db.Query(context.Background(), "SELECT code, name, category, description, type, raw_price, currency FROM products WHERE type ILIKE $1 AND category ILIKE $2", typeStr, category)
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	return scanProducts(rows)
}

// GetAllCategories returns unique categories
func (r *ProductRepository) GetAllCategories() []string {
	rows, err := r.db.Query(context.Background(), "SELECT DISTINCT category FROM products")
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		cats = append(cats, c)
	}
	return cats
}

// SearchByName searches products by name
func (r *ProductRepository) SearchByName(query string) []Product {
	rows, err := r.db.Query(context.Background(), "SELECT code, name, category, description, type, raw_price, currency FROM products WHERE name ILIKE $1", "%"+query+"%")
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	return scanProducts(rows)
}

// GetAllProducts returns all products
func (r *ProductRepository) GetAllProducts() []Product {
	rows, err := r.db.Query(context.Background(), "SELECT code, name, category, description, type, raw_price, currency FROM products")
	if err != nil {
		return []Product{}
	}
	defer rows.Close()
	return scanProducts(rows)
}

func scanProducts(rows pgx.Rows) []Product {
	var products []Product
	for rows.Next() {
		var p Product
		// We use pointers to handles potentially null fields if DB allows, but schema text says NOT NULL usually.
		// Scan directly
		err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Details, &p.Type, &p.Price, &p.Currency)
		if err == nil {
			products = append(products, p)
		}
	}
	return products
}

// FormatAsContext returns products formatted for AI context
func (r *ProductRepository) FormatAsContext(products []Product) string {
	if len(products) == 0 {
		return "No products found in this category."
	}
	var sb strings.Builder
	sb.WriteString("Available Products:\n\n")
	for _, p := range products {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", p.Name, p.Category))
		sb.WriteString(fmt.Sprintf("  Type: %s | Price: %s %s\n", p.Type, p.Price, p.Currency))
		sb.WriteString(fmt.Sprintf("  Details: %s\n\n", p.Details))
	}
	return sb.String()
}
