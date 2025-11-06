# ✅ AI Recipe Generator - Quick Reference

## 🚀 Production Endpoints

**Base URL:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`

### 1. Generate AI Recipe
```bash
POST /api/ai/recipe-helper
Content-Type: application/json

{
  "title": "Recipe name",
  "language": "pl|en|ru|ua"
}
```

**Response:** Complete recipe with nutrition, ingredients, steps

### 2. Create Recipe
```bash
POST /api/recipes
Authorization: Bearer {token}

{
  "title": "...",
  "description": "...",
  "grossWeight": 500,
  "netWeight": 450,
  "calories": 850,
  "protein": 35.5,
  "fats": 28,
  "carbs": 95,
  "yield": 400,
  "cost": 45.5,
  "tokensReward": 25
}
```

### 3. Track Views
```bash
POST /api/recipes/{id}/view
```
Awards 1 ChefToken per 10 views automatically.

---

## 📊 Recipe Metrics

| Field | Type | Description |
|-------|------|-------------|
| grossWeight | int | Вес сырых продуктов (г) |
| netWeight | int | Вес после обработки (г) |
| calories | int | Калорийность (ккал) |
| protein | float | Белки (г) |
| fats | float | Жиры (г) |
| carbs | float | Углеводы (г) |
| yield | int | Выход готового блюда (г) |
| cost | float | Цена (PLN) |
| tokensReward | int | ChefTokens за создание (10-50) |
| viewsCount | int | Просмотры |
| tokensEarned | int | Заработано токенов |

---

## 🌍 Supported Languages

- 🇵🇱 **Polish** (`pl`) - Tested ✅
- 🇬🇧 **English** (`en`) - Tested ✅
- 🇷🇺 **Russian** (`ru`) - Tested ✅
- 🇺🇦 **Ukrainian** (`ua`) - Tested ✅

---

## 🧪 Quick Tests

### Polish:
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title":"Pierogi z kapustą","language":"pl"}' | jq .
```

### English:
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title":"Caesar Salad","language":"en"}' | jq .
```

### Ukrainian:
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/recipe-helper \
  -H "Content-Type: application/json" \
  -d '{"title":"Вареники з вишнею","language":"ua"}' | jq .
```

---

## 💾 Database

**Schema:** `Recipe` table
**Migrations:** 005, 006 (executed on Neon PostgreSQL)
**Indexes:** calories, tokens_earned

---

## 🎯 Status: ✅ PRODUCTION READY

All tests passed on November 5, 2025.
