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
	c := cron.New(cron.WithLocation(time.UTC))
	
	return &FridgeExpiryChecker{
		cron: c,
		db:   db,
	}
}

// Start запускает CRON задачу
func (f *FridgeExpiryChecker) Start() error {
	// Запуск каждый день в 08:00 UTC (11:00 по Москве, 09:00 по Варшаве)
	_, err := f.cron.AddFunc("0 8 * * *", f.checkAllUsers)
	
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	f.cron.Start()
	fmt.Println("🕐 CRON: Fridge expiry checker started (daily at 08:00 UTC)")
	
	return nil
}

// Stop останавливает CRON задачу
func (f *FridgeExpiryChecker) Stop() {
	f.cron.Stop()
	fmt.Println("🛑 CRON: Fridge expiry checker stopped")
}

// checkAllUsers проверяет продукты всех пользователей
func (f *FridgeExpiryChecker) checkAllUsers() {
	fmt.Printf("\n🔍 [%s] Starting daily fridge expiry check...\n", time.Now().Format("2006-01-02 15:04:05"))

	// Получаем всех пользователей у которых есть продукты в холодильнике
	var userIDs []string
	err := f.db.Model(&models.FridgeItem{}).
		Select("DISTINCT user_id").
		Where("status = ? AND expires_at IS NOT NULL", models.FridgeItemStatusFresh).
		Pluck("user_id", &userIDs).Error

	if err != nil {
		fmt.Printf("❌ Failed to fetch users with fridge items: %v\n", err)
		return
	}

	fmt.Printf("📊 Found %d users with fridge items\n", len(userIDs))

	successCount := 0
	errorCount := 0

	for _, userID := range userIDs {
		if err := fridgeService.CheckAndNotifyExpiringItems(f.db, userID); err != nil {
			fmt.Printf("❌ Failed to check user %s: %v\n", userID, err)
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("✅ Daily check completed: %d success, %d errors\n\n", successCount, errorCount)
}

// RunNow запускает проверку немедленно (для тестирования)
func (f *FridgeExpiryChecker) RunNow() {
	fmt.Println("🚀 Running fridge expiry check manually...")
	f.checkAllUsers()
}
