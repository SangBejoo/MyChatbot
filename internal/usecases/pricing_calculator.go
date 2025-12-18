package usecases

import (
	"fmt"
	"project_masAde/internal/repository"
	"regexp"
	"strconv"
	"strings"
)

type PricingCalculator struct {
	productRepo *repository.ProductRepository
}

func NewPricingCalculator(repo *repository.ProductRepository) *PricingCalculator {
	return &PricingCalculator{productRepo: repo}
}

// QueryParse represents a parsed product query
type QueryParse struct {
	ProductName string
	Quantity    int
	WeightGrams int
	Total       float64
	Currency    string
	Error       string
}

// ParseQuery extracts product, quantity, and weight from user input
// Examples: "30 tumbler 30kg", "50 steel rods 1000g", "10 coffee 5kg"
func (pc *PricingCalculator) ParseQuery(input string) QueryParse {
	result := QueryParse{}
	
	input = strings.ToLower(strings.TrimSpace(input))
	
	// Extract quantity (first number)
	quantityRegex := regexp.MustCompile(`(\d+)\s+([a-z\s]+)\s+(\d+)([kg|g]*)`)
	matches := quantityRegex.FindStringSubmatch(input)
	
	if len(matches) < 3 {
		result.Error = "Format not recognized. Please use: '30 tumbler 30kg'"
		return result
	}
	
	// Parse quantity
	quantity, err := strconv.Atoi(matches[1])
	if err != nil {
		result.Error = "Invalid quantity"
		return result
	}
	result.Quantity = quantity
	
	// Parse product name
	result.ProductName = strings.TrimSpace(matches[2])
	
	// Parse weight
	weight, err := strconv.Atoi(matches[3])
	if err != nil {
		result.Error = "Invalid weight"
		return result
	}
	
	// Convert to grams
	unit := strings.TrimSpace(matches[4])
	if unit == "kg" || strings.Contains(unit, "kg") {
		result.WeightGrams = weight * 1000
	} else {
		result.WeightGrams = weight
	}
	
	return result
}

// CalculatePrice finds matching product and calculates total cost
func (pc *PricingCalculator) CalculatePrice(query QueryParse) string {
	if query.Error != "" {
		return query.Error
	}
	
	// Search for product by name
	products := pc.productRepo.SearchByName(query.ProductName)
	if len(products) == 0 {
		return fmt.Sprintf("❌ Product '%s' not found in catalog.", query.ProductName)
	}
	
	product := products[0]
	
	// Parse price
	price, err := strconv.ParseFloat(product.Price, 64)
	if err != nil {
		return "❌ Error parsing product price."
	}
	
	// Calculate total based on product details
	var total float64
	var calculation string
	weightKg := float64(query.WeightGrams) / 1000
	
	if strings.Contains(strings.ToLower(product.Details), "per kg") {
		total = price * weightKg
		calculation = fmt.Sprintf("%.2f (per kg) × %.2f kg", price, weightKg)
	} else if strings.Contains(strings.ToLower(product.Details), "per meter") {
		// Assume weight as length in meters for textiles
		total = price * weightKg // Assuming 1kg = 1m for simplicity; adjust if needed
		calculation = fmt.Sprintf("%.2f (per meter) × %.2f m", price, weightKg)
	} else {
		// Per unit pricing
		total = price * float64(query.Quantity)
		calculation = fmt.Sprintf("%.2f (per unit) × %d units", price, query.Quantity)
	}
	
	result := fmt.Sprintf(`
✅ **Order Summary**
📦 Product: %s
📂 Category: %s | Type: %s
📊 Quantity: %d units | Weight: %dg
💰 Calculation: %s
🏷️ **Total: %.2f %s**
	`,
		product.Name,
		product.Category,
		product.Type,
		query.Quantity,
		query.WeightGrams,
		calculation,
		total,
		product.Currency,
	)
	
	return result
}
