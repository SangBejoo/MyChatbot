package repository

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
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
	products []Product
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: []Product{},
	}
}

// LoadFromCSV loads products from a CSV file
func (r *ProductRepository) LoadFromCSV(filePath string) error {
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

	// Skip header row
	for i := 1; i < len(records); i++ {
		if len(records[i]) >= 7 {
			r.products = append(r.products, Product{
				ID:       records[i][0],
				Name:     records[i][1],
				Category: records[i][2],
				Type:     records[i][3],
				Price:    records[i][4],
				Currency: records[i][5],
				Details:  records[i][6],
			})
		}
	}
	return nil
}

// GetByCategory returns all products in a category
func (r *ProductRepository) GetByCategory(category string) []Product {
	var filtered []Product
	for _, p := range r.products {
		if strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetByType returns all products of a type (export/import)
func (r *ProductRepository) GetByType(typeStr string) []Product {
	var filtered []Product
	for _, p := range r.products {
		if strings.EqualFold(p.Type, typeStr) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetByTypeAndCategory returns filtered products
func (r *ProductRepository) GetByTypeAndCategory(typeStr, category string) []Product {
	var filtered []Product
	for _, p := range r.products {
		if strings.EqualFold(p.Type, typeStr) && strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetAllCategories returns unique categories
func (r *ProductRepository) GetAllCategories() []string {
	categoryMap := make(map[string]bool)
	for _, p := range r.products {
		categoryMap[p.Category] = true
	}
	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	return categories
}

// SearchByName searches products by name
func (r *ProductRepository) SearchByName(query string) []Product {
	var filtered []Product
	query = strings.ToLower(query)
	for _, p := range r.products {
		if strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetAllProducts returns all products
func (r *ProductRepository) GetAllProducts() []Product {
	return r.products
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
