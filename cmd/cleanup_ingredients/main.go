package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Получаем DATABASE_URL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Подключаемся к базе данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("\n🧹 Starting ingredients cleanup...")

	// 1. Удаление тестовых продуктов
	fmt.Println("\n1️⃣ Deleting test products...")
	result, err := db.Exec(`
		DELETE FROM "Ingredient" 
		WHERE "name" LIKE '%Тестов%' 
		   OR "name" LIKE '%тест%'
		   OR "name" = 'Тестовый лосось через API'
		   OR "name" = 'Тестовый угорь'
	`)
	if err != nil {
		log.Printf("⚠️  Error deleting test products: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Deleted %d test products\n", rows)
	}

	// 2. Удаление русских дубликатов - лосось
	fmt.Println("\n2️⃣ Deleting Russian duplicates (salmon)...")
	result, err = db.Exec(`
		DELETE FROM "Ingredient" 
		WHERE "name" IN (
			'Лосось свежий',
			'Лосось норвежский', 
			'Лосось Фермерский',
			'Лосось фермерский',
			'Лосось чилийский'
		)
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Deleted %d salmon duplicates\n", rows)
	}

	// 3. Удаление русских дубликатов - креветки
	fmt.Println("\n3️⃣ Deleting Russian duplicates (shrimp)...")
	result, err = db.Exec(`
		DELETE FROM "Ingredient" 
		WHERE "name" IN (
			'Креветки Королевские',
			'Креветки тигровые'
		)
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Deleted %d shrimp duplicates\n", rows)
	}

	// 4. Удаление русских дубликатов - тунец
	fmt.Println("\n4️⃣ Deleting Russian duplicates (tuna)...")
	result, err = db.Exec(`
		DELETE FROM "Ingredient" 
		WHERE "name" IN (
			'Лещь',
			'Тунец',
			'Тунец Желтопёрый',
			'Тунец желтоперый',
			'Тунец свежий'
		)
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Deleted %d tuna duplicates\n", rows)
	}

	// 5. Удаление русских базовых продуктов
	fmt.Println("\n5️⃣ Deleting Russian basic products...")
	result, err = db.Exec(`
		DELETE FROM "Ingredient" 
		WHERE "name" IN (
			'Минеральная вода',
			'Мука',
			'Соль',
			'Яица'
		)
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Deleted %d basic product duplicates\n", rows)
	}

	// 6. Установка defaultShelfLifeDays для protein
	fmt.Println("\n6️⃣ Setting defaultShelfLifeDays for protein...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 3
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'protein'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d protein items\n", rows)
	}

	// 7. Установка defaultShelfLifeDays для vegetable
	fmt.Println("\n7️⃣ Setting defaultShelfLifeDays for vegetable...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 7
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'vegetable'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d vegetable items\n", rows)
	}

	// 8. Установка defaultShelfLifeDays для dairy
	fmt.Println("\n8️⃣ Setting defaultShelfLifeDays for dairy...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 14
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'dairy'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d dairy items\n", rows)
	}

	// 9. Установка defaultShelfLifeDays для grain
	fmt.Println("\n9️⃣ Setting defaultShelfLifeDays for grain...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 365
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'grain'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d grain items\n", rows)
	}

	// 10. Установка defaultShelfLifeDays для condiment
	fmt.Println("\n🔟 Setting defaultShelfLifeDays for condiment...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 365
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'condiment'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d condiment items\n", rows)
	}

	// 11. Установка defaultShelfLifeDays для other
	fmt.Println("\n1️⃣1️⃣ Setting defaultShelfLifeDays for other...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultShelfLifeDays" = 180
		WHERE "defaultShelfLifeDays" IS NULL 
		  AND "category" = 'other'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d other items\n", rows)
	}

	// 12. Установка defaultPricePerUnit для protein
	fmt.Println("\n1️⃣2️⃣ Setting defaultPricePerUnit for protein...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultPricePerUnit" = 0.02
		WHERE "defaultPricePerUnit" IS NULL 
		  AND "category" = 'protein'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d protein items\n", rows)
	}

	// 13. Установка defaultPricePerUnit для vegetable, grain, dairy
	fmt.Println("\n1️⃣3️⃣ Setting defaultPricePerUnit for vegetable/grain/dairy...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultPricePerUnit" = 0.01
		WHERE "defaultPricePerUnit" IS NULL 
		  AND "category" IN ('vegetable', 'grain', 'dairy')
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d items\n", rows)
	}

	// 14. Установка defaultPricePerUnit для condiment
	fmt.Println("\n1️⃣4️⃣ Setting defaultPricePerUnit for condiment...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultPricePerUnit" = 0.03
		WHERE "defaultPricePerUnit" IS NULL 
		  AND "category" = 'condiment'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d condiment items\n", rows)
	}

	// 15. Установка defaultPricePerUnit для other
	fmt.Println("\n1️⃣5️⃣ Setting defaultPricePerUnit for other...")
	result, err = db.Exec(`
		UPDATE "Ingredient" 
		SET "defaultPricePerUnit" = 0.01
		WHERE "defaultPricePerUnit" IS NULL 
		  AND "category" = 'other'
	`)
	if err != nil {
		log.Printf("⚠️  Error: %v", err)
	} else {
		rows, _ := result.RowsAffected()
		fmt.Printf("   ✅ Updated %d other items\n", rows)
	}

	// Финальная статистика
	fmt.Println("\n📊 Final Statistics:")
	fmt.Println("==================================================")

	rows, err := db.Query(`
		SELECT 
			"category",
			COUNT(*) as count
		FROM "Ingredient"
		GROUP BY "category"
		ORDER BY "category"
	`)
	if err != nil {
		log.Printf("⚠️  Error getting statistics: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("\nIngredients by category:")
		for rows.Next() {
			var category string
			var count int
			if err := rows.Scan(&category, &count); err == nil {
				fmt.Printf("  - %-12s: %3d items\n", category, count)
			}
		}
	}

	// Проверка на отсутствующие поля
	var missingShelfLife, missingPrice int
	db.QueryRow(`SELECT COUNT(*) FROM "Ingredient" WHERE "defaultShelfLifeDays" IS NULL`).Scan(&missingShelfLife)
	db.QueryRow(`SELECT COUNT(*) FROM "Ingredient" WHERE "defaultPricePerUnit" IS NULL`).Scan(&missingPrice)

	fmt.Printf("\n⚠️  Items without defaultShelfLifeDays: %d\n", missingShelfLife)
	fmt.Printf("⚠️  Items without defaultPricePerUnit: %d\n", missingPrice)

	// Общее количество
	var totalCount int
	db.QueryRow(`SELECT COUNT(*) FROM "Ingredient"`).Scan(&totalCount)
	fmt.Printf("\n✅ Total ingredients: %d\n", totalCount)

	fmt.Println("\n🎉 Cleanup completed successfully!")
}
