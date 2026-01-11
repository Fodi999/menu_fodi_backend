# 🧪 Backend AI Language Test - Manual Commands

**Purpose:** Test that backend properly reads `Accept-Language` and `language` field

---

## Prerequisites

1. ✅ Backend server running on `localhost:8080`
2. ✅ Valid JWT token (get from login)
3. ✅ Ingredient IDs from database

---

## Get JWT Token (if needed)

```bash
# Login as super admin
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "dmitrij.fomin.fodi@gmail.com",
    "password": "your_password"
  }' | jq -r '.token'
```

Save token as environment variable:
```bash
export TOKEN="paste_token_here"
```

---

## Test 1: Russian Language

### Request:
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Content-Type: application/json" \
  -H "Accept-Language: ru" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Жареный лосось",
    "language": "ru",
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "quantity": 150,
        "unit": "g"
      },
      {
        "ingredientId": "9ff773d2-a3ee-4f4b-bc45-4cfe0d7f680b",
        "quantity": 20,
        "unit": "ml"
      }
    ],
    "rawCookingText": "Обжарить лосось на масле до золотистой корочки"
  }' | jq '.'
```

### Expected Response:
```json
{
  "success": true,
  "message": "Recipe preview generated",
  "data": {
    "title": "Жареный лосось",
    "language": "ru",
    "description": "Классическое блюдо с сочным лососем...",
    "steps": [
      {
        "order": 1,
        "text": "Обжарить лосось на масле...",
        "time": 5
      }
    ],
    "servings": 1,
    "timeMinutes": 10,
    "difficulty": "easy",
    "calories": 250
  }
}
```

### ✅ Success Criteria:
- `language` field = `"ru"`
- `description` contains Cyrillic characters (А-Я, а-я)
- `steps[].text` contains Cyrillic characters

### Backend Logs to Check:
```bash
tail -f server_test.log | grep "Language"
```

Expected:
```
🌐 Language from Accept-Language: ru → ru
🔧 Enriched 2 ingredients for AI (lang=ru)
```

---

## Test 2: English Language

### Request:
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Content-Type: application/json" \
  -H "Accept-Language: en" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Grilled Salmon",
    "language": "en",
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "quantity": 150,
        "unit": "g"
      }
    ],
    "rawCookingText": "Grill salmon in oil until golden"
  }' | jq '.'
```

### ✅ Success Criteria:
- `language` field = `"en"`
- `description` is in English (no Cyrillic)
- `steps[].text` is in English

---

## Test 3: Polish Language

### Request:
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Content-Type: application/json" \
  -H "Accept-Language: pl" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Smażony łosoś",
    "language": "pl",
    "ingredients": [
      {
        "ingredientId": "fe1c7431-b1b7-4d36-94bf-74276481983e",
        "quantity": 150,
        "unit": "g"
      }
    ],
    "rawCookingText": "Smażyć łososia na oleju do złocistego"
  }' | jq '.'
```

### ✅ Success Criteria:
- `language` field = `"pl"`
- `description` is in Polish
- Contains Polish characters (ł, ą, ę, ć, ń, ó, ś, ź, ż)

---

## Test 4: Body language overrides Accept-Language

### Request (Accept-Language=en, but body language=ru):
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Content-Type: application/json" \
  -H "Accept-Language: en" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Тестовый рецепт",
    "language": "ru",
    "ingredients": [...],
    "rawCookingText": "..."
  }' | jq '.data.language'
```

### Expected:
```
"ru"
```

Backend log should show:
```
🌐 Language from body: ru
```

---

## Test 5: Fallback to Accept-Language if body is empty

### Request (no language in body):
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Content-Type: application/json" \
  -H "Accept-Language: pl" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Test",
    "ingredients": [...],
    "rawCookingText": "..."
  }' | jq '.data.language'
```

### Expected:
```
"pl"
```

Backend log should show:
```
🌐 Language from Accept-Language: pl → pl
```

---

## Debug Commands

### Check if server is running:
```bash
curl -s http://localhost:8080/health | jq '.'
```

### Watch logs in real-time:
```bash
tail -f server_test.log
```

### Filter language-related logs:
```bash
grep "Language\|Enriched\|Calling AI" server_test.log
```

### Check last 10 requests:
```bash
tail -20 server_test.log | grep "POST.*preview-ai"
```

---

## Troubleshooting

### Problem: "Unauthorized" error
**Solution:** Get fresh JWT token with login endpoint

### Problem: Language always = "en"
**Check:**
1. Handler reads `Accept-Language`: `grep "Language from" server_test.log`
2. Service receives language: `grep "Enriched.*lang=" server_test.log`
3. AI prompt includes language: Check `recipe_ai.go:211` system prompt

### Problem: AI returns English even with language="ru"
**Check:** AI system prompt must include:
```
Return the recipe in the language specified: ru
Description must be in ru language
```

---

## Success Checklist

- [ ] ✅ Test 1 (Russian): AI generates in Russian
- [ ] ✅ Test 2 (English): AI generates in English
- [ ] ✅ Test 3 (Polish): AI generates in Polish
- [ ] ✅ Test 4: Body language overrides header
- [ ] ✅ Test 5: Falls back to Accept-Language
- [ ] ✅ Backend logs show correct language detection
- [ ] ✅ AI prompt includes language parameter

If all tests pass → **Language support is WORKING!** ✅
