# Input Sanitization - Quick Reference

## 🧹 Purpose
Clean user input before sending to AI to prevent garbage and improve classification accuracy.

## 📍 Implementation
**File**: `internal/modules/admin/service/service.go`  
**Function**: `sanitizeIngredientName(input string) string`

## 🎯 What It Does

### Removes:
1. **Numbers**: 123, 999, 2024
2. **Test/Debug words** (case-insensitive):
   - test, testing
   - prod, production
   - demo, sample, example
   - debug

### Limits:
- **Max 3 words** (prevents AI overload)

### Preserves:
- Actual ingredient words
- Multi-language names
- Special characters (é, ñ, ó, etc.)

## ✅ Examples

| Input | Output |
|-------|--------|
| `Kiwi production test 12345` | `Kiwi` |
| `Fresh green testing cucumber 999 demo` | `Fresh green cucumber` |
| `Pomidor 2024 sample` | `Pomidor` |
| `Масло оливковое extra virgin debug 777` | `Масло оливковое extra` |

## 🔄 Flow

```
User Input → sanitizeIngredientName() → AI Classification → DB
"Kiwi test 123"    "Kiwi"              {pl/en/ru, category, unit}
```

## 📊 Impact

### Before Sanitization:
```json
Input: "Kiwi production test"
AI Response: null or incorrect (confused by "production test")
```

### After Sanitization:
```json
Input: "Kiwi production test"
Sanitized: "Kiwi"
AI Response: {
  "name_en": "kiwi",
  "category": "fruit",
  "unit": "g",
  "normalized_value": "kiwi"
}
```

## 🐛 Bug Fixed
**Production Issue**: API returned `data: null` when input contained test/debug words.  
**Root Cause**: AI confused by non-food words in prompt.  
**Solution**: Strip garbage before sending to AI.

## 📝 Logging
```
📝 Input sanitized: 'Kiwi production test 12345' → 'Kiwi'
```

## 🚀 Deployment
- Committed: 2026-01-07
- Deployed: Production (Koyeb)
- Status: ✅ Working

## 🔮 Future Enhancements
1. **Configurable banned words** (admin panel)
2. **Language-specific filters** (e.g., "тест" for Russian)
3. **Autocorrect typos** (e.g., "tomatoo" → "tomato")
4. **Emoji removal** (🍅 → tomato)
