package service

import (
	"errors"
	"fmt"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"gorm.io/gorm"
)

// ============================================================================
// СЕРВИС КАНОНИЧЕСКИХ ПРОДУКТОВ
// Единственное место для создания и поиска продуктов
// ============================================================================

type CanonicalIngredientService struct {
	db *gorm.DB
}

func NewCanonicalIngredientService(db *gorm.DB) *CanonicalIngredientService {
	return &CanonicalIngredientService{
		db: db,
	}
}

// ============================================================================
// ПОИСК ПРОДУКТОВ
// ============================================================================

// FindByNormalizedName ищет продукт по нормализованному названию (через alias)
// Возвращает канонический продукт, если найден алиас
func (s *CanonicalIngredientService) FindByNormalizedName(name string) (*models.CanonicalIngredient, error) {
	normalized := utils.NormalizeName(name)

	var alias models.IngredientAlias
	err := s.db.
		Preload("CanonicalIngredient").
		Preload("CanonicalIngredient.Aliases").
		Where("normalizedName = ?", normalized).
		First(&alias).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Не найдено - это норм
		}
		return nil, fmt.Errorf("failed to find alias: %w", err)
	}

	return alias.CanonicalIngredient, nil
}

// FindByCanonicalKey ищет продукт по каноническому ключу
func (s *CanonicalIngredientService) FindByCanonicalKey(key string) (*models.CanonicalIngredient, error) {
	var ingredient models.CanonicalIngredient
	err := s.db.
		Preload("Aliases").
		Where("canonicalKey = ? AND status = ?", key, models.IngredientStatusActive).
		First(&ingredient).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ingredient: %w", err)
	}

	return &ingredient, nil
}

// SearchByQuery ищет продукты по запросу (для автокомплита)
// Возвращает топ N совпадений
func (s *CanonicalIngredientService) SearchByQuery(query string, language string, limit int) ([]*models.IngredientSearchResult, error) {
	normalized := utils.NormalizeName(query)

	var aliases []models.IngredientAlias
	err := s.db.
		Preload("CanonicalIngredient").
		Where("normalizedName LIKE ? AND canonicalIngredientId IN (SELECT id FROM \"CanonicalIngredient\" WHERE status = ?)",
							normalized+"%", models.IngredientStatusActive).
		Order("LENGTH(normalizedName) ASC"). // Сначала короткие (точнее совпадают)
		Limit(limit).
		Find(&aliases).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search aliases: %w", err)
	}

	results := make([]*models.IngredientSearchResult, 0, len(aliases))
	seen := make(map[string]bool) // Убираем дубли по canonicalKey

	for _, alias := range aliases {
		if alias.CanonicalIngredient == nil {
			continue
		}

		key := alias.CanonicalIngredient.CanonicalKey
		if seen[key] {
			continue
		}
		seen[key] = true

		displayName := alias.CanonicalIngredient.GetNameForLanguage(language)

		results = append(results, &models.IngredientSearchResult{
			ID:           alias.CanonicalIngredient.ID,
			CanonicalKey: alias.CanonicalIngredient.CanonicalKey,
			DisplayName:  displayName,
			Category:     alias.CanonicalIngredient.Category,
			Unit:         alias.CanonicalIngredient.BaseUnit,
			MatchedAlias: alias.Name,
		})
	}

	return results, nil
}

// ============================================================================
// СОЗДАНИЕ ПРОДУКТОВ
// ============================================================================

// CreateOrFindIngredient - ГЛАВНАЯ ФУНКЦИЯ
// Пытается найти существующий продукт, если не нашёл - создаёт новый
// AI и UI должны ВСЕГДА использовать эту функцию!
func (s *CanonicalIngredientService) CreateOrFindIngredient(input *CreateIngredientInput) (*models.CanonicalIngredient, bool, error) {
	// 1. Проверяем, существует ли уже такой продукт
	existing, err := s.FindByNormalizedName(input.Name)
	if err != nil {
		return nil, false, err
	}

	if existing != nil {
		// Продукт найден! Возвращаем его
		return existing, false, nil
	}

	// 2. Продукта нет - создаём новый
	ingredient, err := s.CreateNewIngredient(input)
	if err != nil {
		return nil, false, err
	}

	return ingredient, true, nil
}

// CreateNewIngredient создаёт НОВЫЙ канонический продукт с алиасами
// ⚠️ Не проверяет дубликаты! Используйте CreateOrFindIngredient
func (s *CanonicalIngredientService) CreateNewIngredient(input *CreateIngredientInput) (*models.CanonicalIngredient, error) {
	canonicalKey := utils.GenerateCanonicalKey(input.Name)

	// Создаём канонический продукт
	ingredient := &models.CanonicalIngredient{
		CanonicalKey:         canonicalKey,
		CanonicalName:        input.CanonicalName,
		Category:             input.Category,
		NutritionGroup:       input.NutritionGroup,
		BaseUnit:             input.BaseUnit,
		DefaultShelfLifeDays: input.DefaultShelfLifeDays,
		DefaultPricePerUnit:  input.DefaultPricePerUnit,
		Status:               models.IngredientStatusActive,
	}

	// Транзакция: создаём продукт + алиасы
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Создаём продукт
		if err := tx.Create(ingredient).Error; err != nil {
			return fmt.Errorf("failed to create ingredient: %w", err)
		}

		// Создаём primary alias (основное название)
		primaryAlias := &models.IngredientAlias{
			CanonicalIngredientID: ingredient.ID,
			Name:                  input.Name,
			NormalizedName:        utils.NormalizeName(input.Name),
			Language:              input.Language,
			AliasType:             models.AliasTypePrimary,
		}

		if err := tx.Create(primaryAlias).Error; err != nil {
			return fmt.Errorf("failed to create primary alias: %w", err)
		}

		// Создаём дополнительные алиасы (переводы, синонимы)
		if input.AdditionalAliases != nil {
			for _, aliasInput := range input.AdditionalAliases {
				alias := &models.IngredientAlias{
					CanonicalIngredientID: ingredient.ID,
					Name:                  aliasInput.Name,
					NormalizedName:        utils.NormalizeName(aliasInput.Name),
					Language:              aliasInput.Language,
					AliasType:             aliasInput.AliasType,
				}

				if err := tx.Create(alias).Error; err != nil {
					// Игнорируем ошибки дублей алиасов
					continue
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Перезагружаем с алиасами
	if err := s.db.Preload("Aliases").First(ingredient, "id = ?", ingredient.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload ingredient: %w", err)
	}

	return ingredient, nil
}

// AddAlias добавляет новый алиас к существующему продукту
// Используется для добавления переводов, синонимов
func (s *CanonicalIngredientService) AddAlias(ingredientID string, aliasInput *AliasInput) error {
	normalized := utils.NormalizeName(aliasInput.Name)

	// Проверяем, не занят ли уже этот алиас
	var existing models.IngredientAlias
	err := s.db.Where("normalizedName = ?", normalized).First(&existing).Error
	if err == nil {
		return fmt.Errorf("alias already exists for ingredient %s", existing.CanonicalIngredientID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check alias: %w", err)
	}

	// Создаём алиас
	alias := &models.IngredientAlias{
		CanonicalIngredientID: ingredientID,
		Name:                  aliasInput.Name,
		NormalizedName:        normalized,
		Language:              aliasInput.Language,
		AliasType:             aliasInput.AliasType,
	}

	if err := s.db.Create(alias).Error; err != nil {
		return fmt.Errorf("failed to create alias: %w", err)
	}

	return nil
}

// ============================================================================
// MERGE / АРХИВИРОВАНИЕ
// ============================================================================

// MergeIngredients объединяет дубликаты: targetID становится главным
// Все алиасы sourceIDs переносятся на target
// Source продукты архивируются
func (s *CanonicalIngredientService) MergeIngredients(targetID string, sourceIDs []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Переносим все алиасы на target
		for _, sourceID := range sourceIDs {
			if err := tx.Model(&models.IngredientAlias{}).
				Where("canonicalIngredientId = ?", sourceID).
				Update("canonicalIngredientId", targetID).Error; err != nil {
				return fmt.Errorf("failed to move aliases from %s: %w", sourceID, err)
			}

			// Архивируем source
			if err := tx.Model(&models.CanonicalIngredient{}).
				Where("id = ?", sourceID).
				Update("status", models.IngredientStatusArchived).Error; err != nil {
				return fmt.Errorf("failed to archive ingredient %s: %w", sourceID, err)
			}
		}

		return nil
	})
}

// ArchiveIngredient архивирует продукт (НЕ удаляет!)
func (s *CanonicalIngredientService) ArchiveIngredient(id string) error {
	return s.db.Model(&models.CanonicalIngredient{}).
		Where("id = ?", id).
		Update("status", models.IngredientStatusArchived).Error
}

// ============================================================================
// DTO
// ============================================================================

type CreateIngredientInput struct {
	Name                 string        `json:"name" binding:"required"`
	CanonicalName        string        `json:"canonicalName"` // Если пусто, используется Name
	Category             string        `json:"category" binding:"required"`
	NutritionGroup       string        `json:"nutritionGroup" binding:"required"`
	BaseUnit             string        `json:"baseUnit" binding:"required"`
	DefaultShelfLifeDays *int          `json:"defaultShelfLifeDays"`
	DefaultPricePerUnit  *float64      `json:"defaultPricePerUnit"`
	Language             *string       `json:"language"` // pl, en, ru
	AdditionalAliases    []*AliasInput `json:"additionalAliases"`
}

type AliasInput struct {
	Name      string  `json:"name" binding:"required"`
	Language  *string `json:"language"`
	AliasType string  `json:"aliasType"` // translation, synonym, typo
}
