package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// ExpiredItemMetadata represents metadata for expired fridge items
type ExpiredItemMetadata struct {
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	Cost           float64 `json:"cost"`
	PricePerUnit   float64 `json:"price_per_unit"`
	Currency       string  `json:"currency"`
	ExpiryDate     string  `json:"expiry_date"`
	ArrivedAt      string  `json:"arrived_at"`
	DaysInFridge   int     `json:"days_in_fridge"`
	Reason         string  `json:"reason"`
	Context        string  `json:"context"`
}

// ExpiredItem represents a fridge item that has expired
type ExpiredItem struct {
	ID             string
	UserID         string
	IngredientID   string
	IngredientName string
	Quantity       float64
	Unit           string
	PricePerUnit   float64
	Currency       string
	ExpiresAt      time.Time
	ArrivedAt      time.Time
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("\n🧹 Starting expired items cleanup...")
	fmt.Println("==================================================")

	// Find expired items
	expiredItems, err := findExpiredItems(db)
	if err != nil {
		log.Fatalf("❌ Failed to find expired items: %v", err)
	}

	if len(expiredItems) == 0 {
		fmt.Println("\n✨ No expired items found - your fridges are fresh!")
		return
	}

	fmt.Printf("\n🗑️  Found %d expired items\n\n", len(expiredItems))

	// Process each expired item
	totalCost := 0.0
	successCount := 0
	failCount := 0

	for i, item := range expiredItems {
		fmt.Printf("[%d/%d] Processing: %s (%.2f %s)\n",
			i+1, len(expiredItems), item.IngredientName, item.Quantity, item.Unit)

		// Calculate cost
		cost := item.Quantity * item.PricePerUnit
		totalCost += cost

		// Create history event
		if err := createHistoryEvent(db, item, cost); err != nil {
			fmt.Printf("  ⚠️  Failed to create history event: %v\n", err)
			failCount++
			continue
		}

		// Delete from fridge
		if err := deleteFridgeItem(db, item.ID); err != nil {
			fmt.Printf("  ⚠️  Failed to delete fridge item: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("  ✅ Removed | Cost: %.2f %s | Expired: %s\n",
			cost, item.Currency, item.ExpiresAt.Format("2006-01-02"))
		successCount++
	}

	// Summary
	fmt.Println("\n==================================================")
	fmt.Println("📊 Cleanup Summary:")
	fmt.Printf("   ✅ Successfully processed: %d items\n", successCount)
	fmt.Printf("   ⚠️  Failed: %d items\n", failCount)
	fmt.Printf("   💰 Total loss: %.2f PLN\n", totalCost)
	fmt.Println("\n🎉 Cleanup completed!")
}

func findExpiredItems(db *sql.DB) ([]ExpiredItem, error) {
	query := `
		SELECT 
			ufi.id,
			ufi.user_id,
			ufi.ingredient_id,
			i.name as ingredient_name,
			ufi.quantity,
			ufi.unit,
			COALESCE(ufi.current_price_per_unit, i.default_price_per_unit, 0.01) as price_per_unit,
			COALESCE(ufi.current_price_currency, 'PLN') as currency,
			ufi.expires_at,
			ufi.arrived_at
		FROM user_fridge_items ufi
		JOIN "Ingredient" i ON i.id = ufi.ingredient_id
		WHERE ufi.expires_at IS NOT NULL
		  AND ufi.expires_at < NOW()
		ORDER BY ufi.expires_at ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExpiredItem
	for rows.Next() {
		var item ExpiredItem
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.IngredientID,
			&item.IngredientName,
			&item.Quantity,
			&item.Unit,
			&item.PricePerUnit,
			&item.Currency,
			&item.ExpiresAt,
			&item.ArrivedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func createHistoryEvent(db *sql.DB, item ExpiredItem, cost float64) error {
	// Calculate days in fridge
	daysInFridge := int(time.Since(item.ArrivedAt).Hours() / 24)

	// Create metadata
	metadata := ExpiredItemMetadata{
		IngredientID:   item.IngredientID,
		IngredientName: item.IngredientName,
		Quantity:       item.Quantity,
		Unit:           item.Unit,
		Cost:           cost,
		PricePerUnit:   item.PricePerUnit,
		Currency:       item.Currency,
		ExpiryDate:     item.ExpiresAt.Format(time.RFC3339),
		ArrivedAt:      item.ArrivedAt.Format(time.RFC3339),
		DaysInFridge:   daysInFridge,
		Reason:         "expiry_date_passed",
		Context:        "fridge_cleanup",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert history event
	query := `
		INSERT INTO history_events (
			id,
			user_id,
			event_type,
			source_type,
			source_id,
			metadata,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err = db.Exec(
		query,
		uuid.New().String(),
		item.UserID,
		"expired", // event_type
		"auto",    // source_type
		item.ID,   // source_id (fridge item ID)
		metadataJSON,
	)

	return err
}

func deleteFridgeItem(db *sql.DB, itemID string) error {
	query := `DELETE FROM user_fridge_items WHERE id = $1`
	_, err := db.Exec(query, itemID)
	return err
}
