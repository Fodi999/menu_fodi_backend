package cron

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// DishAvailabilityChecker CRON задача для проверки доступности блюд
// Проверяет, можно ли приготовить блюдо на основе наличия ингредиентов
type DishAvailabilityChecker struct {
	cron *cron.Cron
	db   *gorm.DB
}

// NewDishAvailabilityChecker создает новый экземпляр CRON задачи
func NewDishAvailabilityChecker(db *gorm.DB) *DishAvailabilityChecker {
	// 🌍 Europe/Warsaw timezone
	location, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		fmt.Printf("⚠️ Failed to load Europe/Warsaw timezone, using UTC: %v\n", err)
		location = time.UTC
	}

	c := cron.New(cron.WithLocation(location))

	return &DishAvailabilityChecker{
		cron: c,
		db:   db,
	}
}

// Start запускает CRON задачу
func (d *DishAvailabilityChecker) Start() error {
	// ⏰ Запуск каждые 30 минут
	// Формат: "minute hour day month weekday"
	// "*/30 * * * *" = каждые 30 минут
	_, err := d.cron.AddFunc("*/30 * * * *", d.checkAllDishes)

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	d.cron.Start()
	
	location := d.cron.Location()
	fmt.Printf("🕐 CRON: Dish availability checker started (every 30 minutes, %s)\n", location)

	return nil
}

// Stop останавливает CRON задачу
func (d *DishAvailabilityChecker) Stop() {
	d.cron.Stop()
	fmt.Println("🛑 CRON: Dish availability checker stopped")
}

// checkAllDishes проверяет доступность всех опубликованных блюд
func (d *DishAvailabilityChecker) checkAllDishes() {
	startTime := time.Now()
	fmt.Printf("\n🔍 [%s] Starting dish availability check...\n", startTime.Format("2006-01-02 15:04:05"))

	// Получаем все опубликованные блюда
	var dishes []models.Dish
	err := d.db.
		Preload("Recipe").
		Preload("Recipe.Ingredients").
		Preload("Recipe.Ingredients.Ingredient").
		Where("status = ?", models.DishStatusPublished).
		Find(&dishes).Error

	if err != nil {
		fmt.Printf("❌ Failed to fetch published dishes: %v\n", err)
		return
	}

	if len(dishes) == 0 {
		fmt.Println("📊 No published dishes found")
		return
	}

	fmt.Printf("📊 Found %d published dishes to check\n", len(dishes))

	availableCount := 0
	unavailableCount := 0
	errorCount := 0
	changedCount := 0

	// Проверяем каждое блюдо
	for _, dish := range dishes {
		// TODO: Реализовать реальную проверку доступности ингредиентов
		// Сейчас используем простую проверку: если у рецепта есть ингредиенты
		isAvailable := len(dish.Recipe.Ingredients) > 0

		// Если статус изменился - обновляем
		if dish.IsAvailable != isAvailable {
			if err := d.db.Model(&dish).Update("is_available", isAvailable).Error; err != nil {
				fmt.Printf("❌ Failed to update dish %s availability: %v\n", dish.ID, err)
				errorCount++
				continue
			}

			changedCount++
			if isAvailable {
				fmt.Printf("✅ Dish '%s' is now AVAILABLE\n", dish.Title)
			} else {
				fmt.Printf("⚠️ Dish '%s' is now UNAVAILABLE\n", dish.Title)
			}
		}

		if isAvailable {
			availableCount++
		} else {
			unavailableCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Availability check completed in %v\n", duration)
	fmt.Printf("📊 Stats: %d available, %d unavailable, %d changed, %d errors\n\n",
		availableCount, unavailableCount, changedCount, errorCount)
}

// RunNow запускает проверку немедленно (для тестирования)
func (d *DishAvailabilityChecker) RunNow() {
	fmt.Println("🚀 Running dish availability check manually...")
	d.checkAllDishes()
}
