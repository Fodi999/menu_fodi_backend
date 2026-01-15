# 🎯 Canonical Ingredients System - Implementation Guide

> **Статус:** ✅ Готово к имплементации  
> **Дата:** 2026-01-15  
> **Цель:** Убрать дубликаты продуктов навсегда

---

## 📦 Что включено

1. **SQL миграция** - новые таблицы с индексами
2. **Go модели** - CanonicalIngredient + IngredientAlias
3. **Сервис нормализации** - canonicalKey generation
4. **CRUD сервис** - CreateOrFind, Search, Merge
5. **Защита от дублей** - unique indexes на БД уровне

---

## 🚀 Порядок внедрения

### Шаг 1: Применить миграцию (5 мин)

```bash
# Подключиться к production БД
psql $DATABASE_URL

# Выполнить миграцию
\i migrations/20260115_add_canonical_ingredients.sql

# Проверить
\d "CanonicalIngredient"
\d "IngredientAlias"
```

**Важно:** Старая таблица `Ingredient` НЕ удаляется! Работает параллельно.

---

### Шаг 2: Тест нормализации (локально)

```bash
cd /Users/dmitrijfomin/Desktop/backend

# Создать тестовый файл
cat > test_normalize.go << 'EOF'
package main

import (
	"fmt"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

func main() {
	tests := []string{
		"Лук репчатый",
		"лук",
		"ЛУК РЕПЧАТЫЙ",
		"Onion",
		"onion",
		"Cebula",
		"CEBULA",
		"Pierś z kurczaka",
	}
	
	fmt.Println("=== Тест нормализации ===\n")
	for _, test := range tests {
		normalized := utils.NormalizeName(test)
		canonicalKey := utils.GenerateCanonicalKey(test)
		fmt.Printf("%-20s -> normalized: %-20s key: %s\n", test, normalized, canonicalKey)
	}
}
EOF

# Запустить
go run test_normalize.go
rm test_normalize.go
```

**Ожидаемый результат:**
```
Лук репчатый        -> normalized: лук репчатый          key: лук-репчатый
лук                 -> normalized: лук                   key: лук
ЛУК РЕПЧАТЫЙ        -> normalized: лук репчатый          key: лук-репчатый
Onion               -> normalized: onion                 key: onion
onion               -> normalized: onion                 key: onion
Cebula              -> normalized: cebula                key: cebula
CEBULA              -> normalized: cebula                key: cebula
Pierś z kurczaka    -> normalized: piers z kurczaka      key: piers-z-kurczaka
```

---

### Шаг 3: Миграция существующих данных

Создадим скрипт миграции `migrations/migrate_to_canonical.go`:

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Подключение к БД
	dsn := "your_database_url"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	
	canonicalService := service.NewCanonicalIngredientService(db)
	
	// Получить все старые ингредиенты
	var oldIngredients []models.Ingredient
	if err := db.Find(&oldIngredients).Error; err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Найдено %d старых ингредиентов\n", len(oldIngredients))
	
	migrated := 0
	duplicates := 0
	
	for _, old := range oldIngredients {
		// Определяем основное имя
		primaryName := old.Name
		if old.NamePL != nil && *old.NamePL != "" {
			primaryName = *old.NamePL
		}
		
		// Пытаемся создать или найти канонический
		input := &service.CreateIngredientInput{
			Name:                 primaryName,
			CanonicalName:        primaryName,
			Category:             old.Category,
			NutritionGroup:       old.NutritionGroup,
			BaseUnit:             old.Unit,
			DefaultShelfLifeDays: old.DefaultShelfLifeDays,
			DefaultPricePerUnit:  old.DefaultPricePerUnit,
			Language:             strPtr("pl"),
		}
		
		// Добавляем алиасы для других языков
		if old.NameEN != nil && *old.NameEN != "" {
			input.AdditionalAliases = append(input.AdditionalAliases, &service.AliasInput{
				Name:      *old.NameEN,
				Language:  strPtr("en"),
				AliasType: models.AliasTypeTranslation,
			})
		}
		
		if old.NameRU != nil && *old.NameRU != "" {
			input.AdditionalAliases = append(input.AdditionalAliases, &service.AliasInput{
				Name:      *old.NameRU,
				Language:  strPtr("ru"),
				AliasType: models.AliasTypeTranslation,
			})
		}
		
		canonical, created, err := canonicalService.CreateOrFindIngredient(input)
		if err != nil {
			log.Printf("❌ Ошибка для '%s': %v\n", primaryName, err)
			continue
		}
		
		if created {
			log.Printf("✅ Создан: %s -> %s\n", primaryName, canonical.CanonicalKey)
			migrated++
		} else {
			log.Printf("♻️  Дубликат: %s -> существующий %s\n", primaryName, canonical.CanonicalKey)
			duplicates++
		}
	}
	
	log.Printf("\n=== Результат ===\n")
	log.Printf("Создано новых: %d\n", migrated)
	log.Printf("Найдено дублей: %d\n", duplicates)
	log.Printf("Всего обработано: %d\n", migrated+duplicates)
}

func strPtr(s string) *string {
	return &s
}
```

**Запуск миграции:**
```bash
cd migrations
export DATABASE_URL="your_production_db_url"
go run migrate_to_canonical.go
```

---

### Шаг 4: Обновить API endpoints

В `internal/modules/admin/transport/http/handlers.go` добавить новые эндпоинты:

```go
// POST /api/admin/ingredients/v2 - Создание через каноническую систему
func (h *AdminHandlers) CreateIngredientV2(w http.ResponseWriter, r *http.Request) {
	var input service.CreateIngredientInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	
	// Валидация canonicalName
	if input.CanonicalName == "" {
		input.CanonicalName = input.Name
	}
	
	// CreateOrFind - автоматически проверяет дубликаты
	ingredient, created, err := h.canonicalService.CreateOrFindIngredient(&input)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	
	utils.RespondWithJSON(w, status, map[string]interface{}{
		"ingredient": ingredient.ToDTO(),
		"created":    created,
	})
}

// GET /api/admin/ingredients/v2/search?q=лук&lang=pl
func (h *AdminHandlers) SearchIngredientsV2(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	language := r.URL.Query().Get("lang")
	if language == "" {
		language = "pl"
	}
	
	results, err := h.canonicalService.SearchByQuery(query, language, 20)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
	})
}

// POST /api/admin/ingredients/v2/merge - Объединить дубликаты
func (h *AdminHandlers) MergeIngredients(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TargetID  string   `json:"targetId"`
		SourceIDs []string `json:"sourceIds"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	
	if err := h.canonicalService.MergeIngredients(input.TargetID, input.SourceIDs); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Ingredients merged successfully",
	})
}
```

**Регистрация роутов** в `internal/modules/admin/module.go`:
```go
r.Post("/ingredients/v2", m.handlers.CreateIngredientV2)
r.Get("/ingredients/v2/search", m.handlers.SearchIngredientsV2)
r.Post("/ingredients/v2/merge", m.handlers.MergeIngredients)
```

---

### Шаг 5: Интеграция с AI

Обновить `internal/modules/admin/service/recipe_ai.go`:

```go
// В функции CreateRecipeWithAI
for _, aiIng := range aiResponse.Ingredients {
	// ❌ БЫЛО: создавать напрямую
	// ingredient := &models.Ingredient{...}
	
	// ✅ СТАЛО: через каноническую систему
	input := &service.CreateIngredientInput{
		Name:           aiIng.Name,
		CanonicalName:  aiIng.Name,
		Category:       aiIng.Category,
		NutritionGroup: determineNutritionGroup(aiIng.Category),
		BaseUnit:       "g",
		Language:       strPtr("pl"),
	}
	
	canonical, _, err := h.canonicalService.CreateOrFindIngredient(input)
	if err != nil {
		log.Printf("AI ingredient error: %v", err)
		continue
	}
	
	// Использовать canonical.ID вместо старого ID
	recipeIngredients = append(recipeIngredients, RecipeIngredient{
		IngredientID: canonical.ID,
		Quantity:     aiIng.Quantity,
		Unit:         canonical.BaseUnit,
	})
}
```

---

## ✅ Проверка работоспособности

### Тест 1: Создание через API

```bash
curl -X POST http://localhost:8080/api/admin/ingredients/v2 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Лук репчатый",
    "category": "vegetable",
    "nutritionGroup": "vegetable",
    "baseUnit": "g",
    "language": "pl",
    "additionalAliases": [
      {"name": "Onion", "language": "en", "aliasType": "translation"},
      {"name": "Лук", "language": "ru", "aliasType": "synonym"}
    ]
  }'
```

**Ожидается:**
- Первый раз: `created: true`
- Второй раз с "лук" или "onion": `created: false` (найден существующий)

### Тест 2: Поиск

```bash
curl "http://localhost:8080/api/admin/ingredients/v2/search?q=лук&lang=pl"
curl "http://localhost:8080/api/admin/ingredients/v2/search?q=oni&lang=en"
```

---

## 🎯 Результат

После внедрения:

- ✅ **Физически невозможно** создать дубликат (unique index)
- ✅ **AI не создаёт мусор** (через CreateOrFind)
- ✅ **Все языки** - одна запись
- ✅ **Merge вместо delete** - история сохраняется
- ✅ **Быстрый поиск** - индексы на normalizedName

---

## 📊 Мониторинг

```sql
-- Сколько канонических продуктов
SELECT COUNT(*) FROM "CanonicalIngredient" WHERE status = 'active';

-- Сколько всего алиасов
SELECT COUNT(*) FROM "IngredientAlias";

-- Топ продуктов с наибольшим количеством алиасов
SELECT 
  ci."canonicalName", 
  COUNT(ia.id) as alias_count
FROM "CanonicalIngredient" ci
LEFT JOIN "IngredientAlias" ia ON ia."canonicalIngredientId" = ci.id
GROUP BY ci.id, ci."canonicalName"
ORDER BY alias_count DESC
LIMIT 10;

-- Проверка дублей (не должно быть!)
SELECT "normalizedName", COUNT(*) 
FROM "IngredientAlias" 
GROUP BY "normalizedName" 
HAVING COUNT(*) > 1;
```

---

## 🔄 Откат (если что-то пошло не так)

```sql
-- Откатить миграцию
DROP TABLE IF EXISTS "IngredientAlias";
DROP TABLE IF EXISTS "CanonicalIngredient";
DROP FUNCTION IF EXISTS update_canonical_ingredient_updated_at();
```

Старая таблица `Ingredient` останется нетронутой!

---

## 📝 Следующие шаги

1. ✅ Применить миграцию
2. ✅ Тест нормализации
3. ⏳ Миграция данных
4. ⏳ Обновить API
5. ⏳ Интеграция с AI
6. ⏳ Обновить фронтенд

**Статус:** Готово к имплементации! 🚀
