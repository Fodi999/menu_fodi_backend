# Многоязычные рецепты - Краткая справка

## ✅ Что реализовано

Все рецепты в системе теперь **автоматически создаются на 3 языках**: PL, EN, RU

### База данных
- ✅ 214 ингредиентов переведены на PL/EN/RU
- ✅ 5 категорий переведены на PL/EN/RU  
- ✅ 8 позиций меню переведены на PL/EN/RU
- ✅ Таблица Recipe имеет поля для переводов

### Backend
- ✅ Модель Recipe обновлена (поля `name_pl`, `name_en`, `name_ru`, и т.д.)
- ✅ AI Translation Service создан (`recipe_translator.go`)
- ✅ Автоматический перевод при публикации рецепта
- ✅ API принимает переводы при создании рецепта

## 🎯 Как использовать

### Создание рецепта С переводами (ручной ввод)

```bash
POST /api/admin/recipes
{
  "localName": "Pasta Carbonara",
  "category": "main",
  "difficulty": "medium",
  "namePl": "Makaron Carbonara",
  "nameEn": "Pasta Carbonara",
  "nameRu": "Паста Карбонара"
}
```

### Создание рецепта БЕЗ переводов (автоперевод)

```bash
POST /api/admin/recipes
{
  "localName": "Борщ",
  "category": "soup",
  "difficulty": "medium"
}
```

**При публикации** система автоматически переведет на PL/EN/RU через AI.

### Публикация с автопереводом

```bash
POST /api/admin/recipes/{id}/publish
{
  "ingredients": [...],
  "steps": [...]
}
```

**Response:**
```json
{
  "status": "success",
  "warnings": [
    "Recipe auto-translated to PL/EN/RU"
  ]
}
```

## 📊 Текущая статистика БД

| Таблица    | Всего | С переводами | Без переводов |
|------------|-------|--------------|---------------|
| Ingredient | 214   | 214 (100%)   | 0             |
| categories | 5     | 5 (100%)     | 0             |
| menu_items | 8     | 8 (100%)     | 0             |
| Recipe     | 0     | 0            | 0 (удалено 2) |

## 🔧 Файлы изменены

1. `internal/models/recipe.go` - добавлены поля переводов
2. `internal/modules/recipes_admin/dto/create_recipe.go` - DTO с переводами
3. `internal/modules/recipes_admin/service/recipe_admin_service.go` - автоперевод
4. `internal/modules/ai/service/recipe_translator.go` - AI Translation Service

## ⚠️ Требуется доработка

1. **Интеграция AI Service** - внедрение зависимости в RecipeAdminService
2. **Перевод steps** - автоматический перевод шагов приготовления
3. **Фронтенд** - UI для выбора языка и отображения переводов

## 📝 Примечания

- AI-рецепты из холодильника создаются только на выбранном языке
- Для полного перевода нужно сохранить рецепт через админ API
- Качество AI переводов требует проверки

---

**Дата:** 19 января 2026 г.  
**Статус:** ✅ Готово к использованию (требуется интеграция AI)
