package repository

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TableMetadata struct {
	ID          int       `json:"id"`
	TableName   string    `json:"table_name"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type TableManager struct {
	db *pgxpool.Pool
}

func NewTableManager(db *pgxpool.Pool) *TableManager {
	return &TableManager{db: db}
}

// sanitizeTableName cleans strings to be safe for SQL table names (alphanumeric + underscore)
func sanitizeTableName(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return strings.ToLower(reg.ReplaceAllString(name, "_"))
}

// ImportCSV creates a dynamic table and imports CSV data transactionally
func (m *TableManager) ImportCSV(displayName string, csvData io.Reader) error {
	ctx := context.Background()
	tableName := "dt_" + sanitizeTableName(displayName) + "_" + time.Now().Format("20060102150405")

	// Parse CSV
	reader := csv.NewReader(csvData)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	if len(rows) < 1 {
		return fmt.Errorf("csv is empty")
	}

	headers := rows[0]
	if len(headers) == 0 {
		return fmt.Errorf("no headers found")
	}

	// Start Transaction
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Create Table
	var colDefs []string
	safeHeaders := make([]string, len(headers))
	for i, h := range headers {
		safeH := sanitizeTableName(h)
		if safeH == "" {
			safeH = fmt.Sprintf("col_%d", i)
		}
		safeHeaders[i] = safeH
		colDefs = append(colDefs, fmt.Sprintf("%s TEXT", safeH)) // Default to TEXT for flexibility
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, %s);", tableName, strings.Join(colDefs, ", "))
	if _, err := tx.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	// 2. Insert Data
	// Prepare statement for bulk insert? Or simple loop for now (easier for dynamic cols)
	// Using pgx CopyFrom is most efficient for bulk, but let's stick to simple INSERT for simplicity with dynamic cols locally
	
	colNames := strings.Join(safeHeaders, ", ")
	placeholders := make([]string, len(safeHeaders))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, colNames, strings.Join(placeholders, ", "))

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		// Handle row length mismatch
		if len(row) != len(headers) {
			// Pad or truncate? Let's skip or error. 
			// For robustness, let's pad with empty strings if short
			if len(row) < len(headers) {
				for len(row) < len(headers) {
					row = append(row, "")
				}
			} else {
				row = row[:len(headers)]
			}
		}

		// Convert []string to []interface{}
		args := make([]interface{}, len(row))
		for j, v := range row {
			args[j] = v
		}

		if _, err := tx.Exec(ctx, insertSQL, args...); err != nil {
			return fmt.Errorf("row %d insert failed: %w", i, err)
		}
	}

	// 3. Register Table in Registry
	_, err = tx.Exec(ctx, "INSERT INTO dynamic_tables (table_name, display_name) VALUES ($1, $2)", tableName, displayName)
	if err != nil {
		return fmt.Errorf("failed to register table: %w", err)
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListTables returns all registered dynamic tables
func (m *TableManager) ListTables() ([]TableMetadata, error) {
	rows, err := m.db.Query(context.Background(), "SELECT id, table_name, display_name, created_at FROM dynamic_tables ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableMetadata
	for rows.Next() {
		var t TableMetadata
		if err := rows.Scan(&t.ID, &t.TableName, &t.DisplayName, &t.CreatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// GetTableDatafetches data from a dynamic table
func (m *TableManager) GetTableData(tableName string) ([]map[string]interface{}, error) {
	// Verify table exists in registry to prevent SQL Injection via tableName
	var exists bool
	err := m.db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM dynamic_tables WHERE table_name=$1)", tableName).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("table not found or unauthorized")
	}

	// Fetch data
	// Since cols are dynamic, we need rows.FieldDescriptions()
	rows, err := m.db.Query(context.Background(), fmt.Sprintf("SELECT * FROM %s", tableName)) // Safe because authorized via registry
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, fd := range fieldDescs {
			colName := string(fd.Name)
			rowMap[colName] = values[i]
		}
		results = append(results, rowMap)
	}
	return results, nil
}

// DeleteTable removes a dynamic table and its registry entry
func (m *TableManager) DeleteTable(tableName string) error {
	ctx := context.Background()
	
	// Verify table exists in registry
	var exists bool
	err := m.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dynamic_tables WHERE table_name=$1)", tableName).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("table not found")
	}

	// Start transaction
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Drop the actual table
	if _, err := tx.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)); err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	// Remove from registry
	if _, err := tx.Exec(ctx, "DELETE FROM dynamic_tables WHERE table_name=$1", tableName); err != nil {
		return fmt.Errorf("failed to remove registry entry: %w", err)
	}

	return tx.Commit(ctx)
}

// UpdateRow updates a single row in a dynamic table
func (m *TableManager) UpdateRow(tableName string, rowID int, data map[string]interface{}) error {
	ctx := context.Background()
	
	// Verify table exists
	var exists bool
	err := m.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dynamic_tables WHERE table_name=$1)", tableName).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("table not found")
	}

	// Build UPDATE query dynamically
	if len(data) == 0 {
		return fmt.Errorf("no data provided")
	}

	var setClauses []string
	var args []interface{}
	i := 1
	for col, val := range data {
		// Sanitize column name
		safeCol := sanitizeTableName(col)
		if safeCol == "" || safeCol == "id" {
			continue // Skip ID or invalid columns
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", safeCol, i))
		args = append(args, val)
		i++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no valid columns to update")
	}

	args = append(args, rowID)
	updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(setClauses, ", "), i)

	_, err = m.db.Exec(ctx, updateSQL, args...)
	return err
}

// DeleteRow deletes a single row from a dynamic table
func (m *TableManager) DeleteRow(tableName string, rowID int) error {
	ctx := context.Background()
	
	// Verify table exists
	var exists bool
	err := m.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dynamic_tables WHERE table_name=$1)", tableName).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("table not found")
	}

	_, err = m.db.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName), rowID)
	return err
}
