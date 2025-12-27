package usecases

import (
	"fmt"
	"project_masAde/internal/repository"
	"regexp"
	"strconv"
	"strings"
)

// DynamicCalculator performs calculations using data from user-imported datasets
type DynamicCalculator struct {
	tableManager *repository.TableManager
}

func NewDynamicCalculator(tm *repository.TableManager) *DynamicCalculator {
	return &DynamicCalculator{tableManager: tm}
}

// DynamicQuery represents a parsed calculation query
type DynamicQuery struct {
	Quantity    int
	ProductName string
	WeightGrams int
	Error       string
}

// ParseInput parses user input like "30 tumbler 30kg"
func (dc *DynamicCalculator) ParseInput(input string) DynamicQuery {
	result := DynamicQuery{}
	input = strings.ToLower(strings.TrimSpace(input))

	// Pattern: quantity product weight
	// Examples: "30 tumbler 30kg", "50 steel rods 1000g", "10 coffee"
	quantityRegex := regexp.MustCompile(`^(\d+)\s+(.+?)\s*(\d+)?\s*(kg|g)?$`)
	matches := quantityRegex.FindStringSubmatch(input)

	if len(matches) < 3 {
		// Try simpler pattern: quantity product
		simpleRegex := regexp.MustCompile(`^(\d+)\s+(.+)$`)
		simpleMatches := simpleRegex.FindStringSubmatch(input)
		if len(simpleMatches) >= 3 {
			qty, _ := strconv.Atoi(simpleMatches[1])
			result.Quantity = qty
			result.ProductName = strings.TrimSpace(simpleMatches[2])
			return result
		}
		result.Error = "Format tidak dikenali. Gunakan: `30 tumbler` atau `30 tumbler 30kg`"
		return result
	}

	// Parse quantity
	qty, err := strconv.Atoi(matches[1])
	if err != nil {
		result.Error = "Quantity tidak valid"
		return result
	}
	result.Quantity = qty
	result.ProductName = strings.TrimSpace(matches[2])

	// Parse weight if provided
	if len(matches) > 3 && matches[3] != "" {
		weight, _ := strconv.Atoi(matches[3])
		unit := strings.ToLower(matches[4])
		if unit == "kg" {
			result.WeightGrams = weight * 1000
		} else {
			result.WeightGrams = weight
		}
	}

	return result
}

// Calculate performs calculation by looking up product in dataset
func (dc *DynamicCalculator) Calculate(schemaName, tableName string, query DynamicQuery) string {
	if query.Error != "" {
		return "❌ " + query.Error
	}

	// Fetch data from dataset
	data, err := dc.tableManager.GetTableData(schemaName, tableName)
	if err != nil {
		return fmt.Sprintf("❌ Error mengambil data: %s", err.Error())
	}

	if len(data) == 0 {
		return "❌ Dataset kosong"
	}

	// Find product by name (fuzzy match)
	var matchedRow map[string]interface{}
	searchName := strings.ToLower(query.ProductName)

	for _, row := range data {
		// Check common column names for product name
		for _, colName := range []string{"name", "nama", "product", "produk", "item"} {
			if val, ok := row[colName]; ok {
				if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), searchName) {
					matchedRow = row
					break
				}
			}
		}
		if matchedRow != nil {
			break
		}
	}

	if matchedRow == nil {
		return fmt.Sprintf("❌ Produk '%s' tidak ditemukan di dataset", query.ProductName)
	}

	// Find price column
	var price float64
	var priceFound bool
	var currency string = "USD"

	for _, colName := range []string{"price", "harga", "unit_price", "cost"} {
		if val, ok := matchedRow[colName]; ok {
			priceStr := fmt.Sprintf("%v", val)
			// Remove currency symbols
			priceStr = strings.ReplaceAll(priceStr, "$", "")
			priceStr = strings.ReplaceAll(priceStr, "Rp", "")
			priceStr = strings.ReplaceAll(priceStr, ",", "")
			priceStr = strings.TrimSpace(priceStr)
			
			if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
				price = p
				priceFound = true
				break
			}
		}
	}

	if !priceFound {
		return "❌ Kolom harga tidak ditemukan di dataset. Pastikan ada kolom `price` atau `harga`."
	}

	// Check for currency column
	for _, colName := range []string{"currency", "mata_uang"} {
		if val, ok := matchedRow[colName]; ok {
			currency = fmt.Sprintf("%v", val)
			break
		}
	}

	// Get product display name
	productName := query.ProductName
	for _, colName := range []string{"name", "nama", "product", "produk"} {
		if val, ok := matchedRow[colName]; ok {
			productName = fmt.Sprintf("%v", val)
			break
		}
	}

	// Calculate total
	var total float64
	var calculation string

	if query.WeightGrams > 0 {
		// Weight-based calculation
		weightKg := float64(query.WeightGrams) / 1000
		total = price * weightKg
		calculation = fmt.Sprintf("%.2f × %.2f kg", price, weightKg)
	} else {
		// Quantity-based calculation
		total = price * float64(query.Quantity)
		calculation = fmt.Sprintf("%.2f × %d units", price, query.Quantity)
	}

	// Build result
	result := fmt.Sprintf(`✅ *Hasil Perhitungan*

📦 Produk: %s
📊 Quantity: %d
💰 Harga: %.2f %s
🧮 Perhitungan: %s

🏷️ *Total: %.2f %s*`,
		productName,
		query.Quantity,
		price,
		currency,
		calculation,
		total,
		currency,
	)

	return result
}

// CalculateFromInput is a convenience method that parses and calculates in one call
func (dc *DynamicCalculator) CalculateFromInput(schemaName, tableName, userInput string) string {
	query := dc.ParseInput(userInput)
	return dc.Calculate(schemaName, tableName, query)
}
