# ✅ Ingredient Suggest Localization - COMPLETE

## 🎯 Проблема

Эндпоинт `/api/admin/ingredients/suggest` возвращал всегда английские названия, игнорируя `Accept-Language` заголовок:

```json
{
  "name": "Salmon"  // ❌ всегда English
}
```

А должен был возвращать локализованные имена в зависимости от языка запроса:

```json
{
  "name": "Łosoś"  // ✅ Polish при Accept-Language: pl
}
```

## ✅ Решение

### 1. Обновлен DTO `IngredientSuggestion`

Добавлено поле `nutritionGroup` для полной совместимости с `/ingredients`:

```go
type IngredientSuggestion struct {
    ID             string `json:"id"`
    Name           string `json:"name"`           // Локализованное имя
    Category       string `json:"category"`
    NutritionGroup string `json:"nutritionGroup"` // НОВОЕ
    Unit           string `json:"unit"`
}
```

### 2. Добавлена локализация в Service Layer

#### Обновлен метод `SuggestIngredients`

Теперь принимает параметр `lang` и использует общий метод `getLocalizedName()`:

```go
func (s *adminService) SuggestIngredients(query string, limit int, lang string) ([]IngredientSuggestion, error) {
    // Всегда загружаем ВСЕ языковые поля из БД
    var ingredients []models.Ingredient
    err := s.db.Where(sqlQuery, pattern, pattern, pattern, pattern, pattern).
        Limit(limit).
        Order("name ASC").
        Find(&ingredients).Error
    
    // Формируем локализованный ответ
    for _, ing := range ingredients {
        displayName := s.getLocalizedName(ing, lang)
        
        suggestions = append(suggestions, IngredientSuggestion{
            ID:             ing.ID,
            Name:           displayName,  // Локализовано!
            Category:       ing.Category,
            NutritionGroup: ing.NutritionGroup,
            Unit:           ing.Unit,
        })
    }
}
```

#### Добавлен метод `getLocalizedName()`

Единая логика локализации для `/ingredients` и `/suggest`:

```go
func (s *adminService) getLocalizedName(ing models.Ingredient, lang string) string {
    switch lang {
    case "pl":
        if ing.NamePL != nil && *ing.NamePL != "" {
            return *ing.NamePL
        }
    case "ru":
        if ing.NameRU != nil && *ing.NameRU != "" {
            return *ing.NameRU
        }
    case "en":
        if ing.NameEN != nil && *ing.NameEN != "" {
            return *ing.NameEN
        }
    }
    
    // Fallback chain: EN → PL → RU → name
    if ing.NameEN != nil && *ing.NameEN != "" {
        return *ing.NameEN
    }
    if ing.NamePL != nil && *ing.NamePL != "" {
        return *ing.NamePL
    }
    if ing.NameRU != nil && *ing.NameRU != "" {
        return *ing.NameRU
    }
    return ing.Name
}
```

### 3. Обновлен Handler

#### Извлечение языка из заголовка

```go
func (h *AdminHandlers) SuggestIngredients(w http.ResponseWriter, r *http.Request) {
    // Получаем язык из Accept-Language заголовка
    acceptLang := r.Header.Get("Accept-Language")
    lang := normalizeLang(acceptLang)
    
    fmt.Printf("📥 Request: GET /suggest?q=%s (Accept-Language: %s → %s)\n", 
        query, acceptLang, lang)
    
    // Вызываем service с указанием языка
    suggestions, err := h.service.SuggestIngredients(query, limit, lang)
}
```

#### Функция нормализации языка

```go
func normalizeLang(acceptLang string) string {
    acceptLang = strings.ToLower(strings.TrimSpace(acceptLang))
    
    // Парсим первый язык из списка (например: "pl-PL,en;q=0.9" → "pl")
    if idx := strings.Index(acceptLang, ","); idx > 0 {
        acceptLang = acceptLang[:idx]
    }
    
    // Проверяем префикс языка
    switch {
    case strings.HasPrefix(acceptLang, "pl"):
        return "pl"
    case strings.HasPrefix(acceptLang, "ru"):
        return "ru"
    case strings.HasPrefix(acceptLang, "en"):
        return "en"
    default:
        return "en"  // Default: English
    }
}
```

## 🧪 Тестирование

### Test 1: Polish (Accept-Language: pl)

```bash
curl "$API/ingredients/suggest?q=лосось" \
  -H "Accept-Language: pl"
```

**Результат:**
```json
{
  "data": [{
    "id": "fe1c7431-...",
    "name": "Łosoś",          // ✅ Polish
    "category": "fish",
    "nutritionGroup": "protein",
    "unit": "g"
  }]
}
```

### Test 2: English (Accept-Language: en)

```bash
curl "$API/ingredients/suggest?q=лосось" \
  -H "Accept-Language: en"
```

**Результат:**
```json
{
  "data": [{
    "id": "fe1c7431-...",
    "name": "Salmon",         // ✅ English
    "category": "fish",
    "nutritionGroup": "protein",
    "unit": "g"
  }]
}
```

### Test 3: Russian (Accept-Language: ru)

```bash
curl "$API/ingredients/suggest?q=лосось" \
  -H "Accept-Language: ru"
```

**Результат:**
```json
{
  "data": [{
    "id": "fe1c7431-...",
    "name": "Лосось",         // ✅ Russian
    "category": "fish",
    "nutritionGroup": "protein",
    "unit": "g"
  }]
}
```

### Test 4: No Header (default to English)

```bash
curl "$API/ingredients/suggest?q=salmon"
```

**Результат:**
```json
{
  "data": [{
    "id": "fe1c7431-...",
    "name": "Salmon",         // ✅ English (default)
    "category": "fish",
    "nutritionGroup": "protein",
    "unit": "g"
  }]
}
```

### Test 5: Multi-language header

```bash
curl "$API/ingredients/suggest?q=rice" \
  -H "Accept-Language: pl-PL,en;q=0.9"
```

**Результат:**
```json
{
  "data": [{
    "id": "10be8c97-...",
    "name": "Makaron ryżowy", // ✅ Polish (first in list)
    "category": "grain",
    "nutritionGroup": "carbohydrate",
    "unit": "g"
  }]
}
```

## 📊 Архитектура

### Единая логика локализации

Теперь `/ingredients` и `/suggest` используют **одинаковую логику**:

```
1. Handler извлекает Accept-Language
2. Нормализует в pl/en/ru
3. Service загружает ВСЕ языковые поля из БД
4. getLocalizedName() выбирает нужный перевод
5. Fallback chain: EN → PL → RU → name
```

### Преимущества

✅ **Консистентность:** оба эндпоинта ведут себя одинаково  
✅ **Производительность:** один SQL запрос вместо нескольких  
✅ **Масштабируемость:** легко добавить новые языки  
✅ **DRY:** одна функция `getLocalizedName()` для всех  
✅ **Fallback:** всегда есть название на английском  

## 🚀 Интеграция с Frontend

### TypeScript интерфейс

```typescript
interface IngredientSuggestion {
  id: string;
  name: string;           // Локализовано автоматически
  category: string;
  nutritionGroup: string;
  unit: string;
}
```

### Пример использования

```typescript
const fetchSuggestions = async (query: string, lang: string) => {
  const response = await fetch(
    `/api/admin/ingredients/suggest?q=${query}`,
    {
      headers: {
        'Accept-Language': lang,  // 'pl', 'en', или 'ru'
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const { data } = await response.json();
  // data[0].name уже локализовано на нужном языке!
  return data;
};
```

### React пример (autocomplete)

```tsx
const IngredientAutocomplete = () => {
  const { i18n } = useTranslation();
  const [suggestions, setSuggestions] = useState<IngredientSuggestion[]>([]);
  
  const handleSearch = async (query: string) => {
    const data = await fetchSuggestions(query, i18n.language);
    setSuggestions(data);
  };
  
  return (
    <Autocomplete
      options={suggestions}
      getOptionLabel={(opt) => opt.name}  // Уже локализовано!
      renderOption={(props, opt) => (
        <li {...props}>
          {opt.name} ({opt.nutritionGroup}, {opt.unit})
        </li>
      )}
    />
  );
};
```

## 📝 Изменения в коде

### Файлы обновлены:

1. **`/internal/modules/admin/service/service.go`**
   - Обновлен `IngredientSuggestion` DTO (добавлено `nutritionGroup`)
   - Добавлен параметр `lang` в `SuggestIngredients()`
   - Добавлен метод `getLocalizedName()`
   - Обновлен интерфейс `AdminService`

2. **`/internal/modules/admin/transport/http/handlers.go`**
   - Добавлено извлечение `Accept-Language` заголовка
   - Добавлена функция `normalizeLang()`
   - Обновлены логи для отображения языка

### Файлы созданы:

3. **`/test_localization.sh`**
   - Автоматизированные тесты локализации
   - 5 тестовых сценариев (pl/en/ru/default/multi)

4. **`INGREDIENT_LOCALIZATION_COMPLETE.md`**
   - Полная документация реализации

## ✅ Результат

### До изменений:
```json
GET /suggest?q=лосось
Accept-Language: pl

{
  "name": "Salmon"  // ❌ всегда English
}
```

### После изменений:
```json
GET /suggest?q=лосось
Accept-Language: pl

{
  "id": "fe1c7431-...",
  "name": "Łosoś",          // ✅ Polish
  "category": "fish",
  "nutritionGroup": "protein",
  "unit": "g"
}
```

## 🎯 Что дальше?

### Необязательные улучшения:

- [ ] Кэширование переводов (Redis)
- [ ] Поддержка других языков (de, es, it)
- [ ] Приоритизация результатов по языку запроса
- [ ] A/B тестирование локализованного autocomplete

---

**🎉 Локализация ingredient suggest завершена и протестирована!**

Date: January 8, 2026  
Status: ✅ Production Ready  
Tests: All Passing (5/5)
