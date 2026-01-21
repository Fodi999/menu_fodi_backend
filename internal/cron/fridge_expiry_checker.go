package cron

import (
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	fridgeService "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/fridge/service"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// FridgeExpiryChecker CRON задача для проверки истекающих продуктов
type FridgeExpiryChecker struct {
	cron *cron.Cron
	db   *gorm.DB
}

// NewFridgeExpiryChecker создает новый экземпляр CRON задачи
func NewFridgeExpiryChecker(db *gorm.DB) *FridgeExpiryChecker {
	// 🌍 Europe/Warsaw timezone для корректного времени уведомлений
	location, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		// Fallback to UTC if timezone not found
		fmt.Printf("⚠️ Failed to load Europe/Warsaw timezone, using UTC: %v\n", err)
		location = time.UTC
	}

	c := cron.New(cron.WithLocation(location))

	return &FridgeExpiryChecker{
		cron: c,
		db:   db,
	}
}

// Start запускает CRON задачу
func (f *FridgeExpiryChecker) Start() error {
	// ⏰ Запуск каждый день в 08:00 по локальному времени (Europe/Warsaw)
	// Формат: "minute hour day month weekday"
	// "0 8 * * *" = каждый день в 08:00
	_, err := f.cron.AddFunc("0 8 * * *", f.checkAllUsers)

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	f.cron.Start()
	
	// Показываем в каком часовом поясе работаем
	location := f.cron.Location()
	fmt.Printf("🕐 CRON: Fridge expiry checker started (daily at 08:00 %s)\n", location)

	return nil
}

// Stop останавливает CRON задачу
func (f *FridgeExpiryChecker) Stop() {
	f.cron.Stop()
	fmt.Println("🛑 CRON: Fridge expiry checker stopped")
}

// checkAllUsers проверяет продукты всех пользователей
func (f *FridgeExpiryChecker) checkAllUsers() {
	startTime := time.Now()
	fmt.Printf("\n🔍 [%s] Starting daily fridge expiry check...\n", startTime.Format("2006-01-02 15:04:05"))

	// ✅ ПРАВИЛЬНО: используем UserFridgeItem (не FridgeItem!)
	// Получаем всех пользователей у которых есть продукты с датой истечения
	var userIDs []string
	err := f.db.Model(&models.UserFridgeItem{}).
		Select("DISTINCT user_id").
		Where("expires_at IS NOT NULL").
		Pluck("user_id", &userIDs).Error

	if err != nil {
		fmt.Printf("❌ Failed to fetch users with fridge items: %v\n", err)
		return
	}

	if len(userIDs) == 0 {
		fmt.Println("📊 No users with expiry dates found")
		return
	}

	fmt.Printf("📊 Found %d users with fridge items\n", len(userIDs))

	successCount := 0
	errorCount := 0
	// totalNotifications := 0 // TODO: add counter when CheckAndNotifyExpiringItems returns count

	// Обрабатываем каждого пользователя
	for _, userID := range userIDs {
		// ✅ Используем новую архитектуру: CheckAndNotifyExpiringItems
		if err := fridgeService.CheckAndNotifyExpiringItems(f.db, userID); err != nil {
			fmt.Printf("❌ Failed to check user %s: %v\n", userID, err)
			errorCount++
		} else {
			successCount++
			// TODO: можно добавить счетчик созданных уведомлений
			// totalNotifications += count
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Daily check completed in %v: %d users processed, %d errors\n\n", 
		duration, successCount, errorCount)
	
	// TODO: добавить метрики:
	// - Сколько уведомлений создано (critical, warning, info)
	// - Сколько продуктов проверено
	// - Performance metrics
}

// RunNow запускает проверку немедленно (для тестирования)
func (f *FridgeExpiryChecker) RunNow() {
	fmt.Println("🚀 Running fridge expiry check manually...")
	f.checkAllUsers()
}
