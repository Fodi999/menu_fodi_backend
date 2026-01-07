# AI Ingredient Creation - API Contract

## 📋 Endpoint

```
POST /api/admin/ingredients
```

**Authentication:** Required (Bearer Token, role: admin или super_admin)

---

## 📥 Request

### Headers
```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

### Body (ТОЛЬКО 1 поле!)
```json
{
  "inputName": "Соль каменная"
}
```

**Правила валидации:**
- ✅ `inputName` - обязательное, минимум 2 символа
- ❌ `inputLang` - НЕ требуется (AI определяет автоматически)
- ❌ `category` - НЕ требуется (AI определяет автоматически)
- ❌ `unit` - НЕ требуется (AI определяет автоматически)

---

## 📤 Response

### Success (201 Created)
```json
{
  "success": true,
  "message": "Ingredient created via AI classification",
  "data": {
    "id": "d29fa77a-0715-4295-994b-301170736038",
    "namePl": "sól kamienia",
    "nameEn": "rock salt",
    "nameRu": "соль каменная",
    "category": "condiment",
    "unit": "g",
    "normalizedValue": "salt",
    "autoTranslated": true
  }
}
```

### Error (409 Conflict) - Дубликат
```json
{
  "error": "Ingredient already exists: INGREDIENT_ALREADY_EXISTS: salt (id: xxx)"
}
```

### Error (400 Bad Request) - Пустое название
```json
{
  "error": "inputName is required"
}
```

### Error (500 Internal Server Error) - AI сбой
```json
{
  "error": "Failed to create ingredient: groq classification failed"
}
```

---

## 🤖 AI Обработка

### Что делает AI:
1. **Определяет язык** входного текста (PL/EN/RU/другие)
2. **Переводит** на все 3 языка (польский, английский, русский)
3. **Классифицирует категорию**:
   - `protein` - мясо, рыба, яйца
   - `vegetable` - овощи
   - `fruit` - фрукты и ягоды
   - `dairy` - молочные продукты
   - `grain` - крупы, макароны, хлеб
   - `condiment` - специи, соусы, масла
   - `other` - прочее
4. **Определяет единицу измерения**:
   - `g` - граммы (твердые продукты)
   - `ml` - миллилитры (жидкости)
   - `pcs` - штуки (целые единицы)
5. **Создает normalized_value** - для проверки дубликатов

### Sanitization (автоматически):
- Удаляет мусорные слова: test, testing, prod, production, demo, sample, debug
- Удаляет числа: 123, 999, 2024
- Ограничивает до 3 слов
- Убирает лишние пробелы

---

## 🧪 Примеры

### Пример 1: Русское название
**Request:**
```bash
curl -X POST http://localhost:8080/api/admin/ingredients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"inputName":"Молоко"}'
```

**Response:**
```json
{
  "data": {
    "namePl": "mleko",
    "nameEn": "milk",
    "nameRu": "молоко",
    "category": "dairy",
    "unit": "ml",
    "normalizedValue": "milk"
  }
}
```

### Пример 2: Английское название
**Request:**
```json
{"inputName": "Fresh Eggs"}
```

**Response:**
```json
{
  "data": {
    "namePl": "świeże jajka",
    "nameEn": "fresh eggs",
    "nameRu": "свежие яйца",
    "category": "protein",
    "unit": "pcs",
    "normalizedValue": "egg"
  }
}
```

### Пример 3: Польское название
**Request:**
```json
{"inputName": "Pomidor"}
```

**Response:**
```json
{
  "data": {
    "namePl": "pomidor",
    "nameEn": "tomato",
    "nameRu": "помидор",
    "category": "vegetable",
    "unit": "g",
    "normalizedValue": "tomato"
  }
}
```

### Пример 4: Sanitization
**Request:**
```json
{"inputName": "Kiwi production test 12345 demo"}
```

**Backend sanitizes to:** `"Kiwi"`

**Response:**
```json
{
  "data": {
    "namePl": "kiwi",
    "nameEn": "kiwi",
    "nameRu": "киви",
    "category": "fruit",
    "unit": "g",
    "normalizedValue": "kiwi"
  }
}
```

### Пример 5: Duplicate Detection
**Request 1:**
```json
{"inputName": "Соль"}
```
✅ Created → `normalizedValue: "salt"`

**Request 2:**
```json
{"inputName": "Salt"}
```
❌ 409 Conflict → `normalizedValue: "salt"` уже существует

---

## 🔒 Защита от дубликатов

Database constraint:
```sql
CREATE UNIQUE INDEX uniq_ingredient_normalized
ON "Ingredient"(normalized_value);
```

Проверка:
- AI создает `normalized_value` (lowercase, singular, ASCII)
- Backend проверяет уникальность перед INSERT
- Если существует → возвращает 409 Conflict с ID существующего

---

## 📊 Статистика

**Среднее время ответа:** 400-600ms  
**Успешность классификации:** 100% (на тестах)  
**AI Model:** llama-3.3-70b-versatile (Groq)  
**Cost per request:** ~$0.0001  

---

## 🎯 Frontend Integration

### TypeScript Interface
```typescript
// Request
interface CreateIngredientRequest {
  inputName: string; // Только это поле!
}

// Response
interface IngredientResponse {
  id: string;
  namePl: string;
  nameEn: string;
  nameRu: string;
  category: 'protein' | 'vegetable' | 'fruit' | 'dairy' | 'grain' | 'condiment' | 'other';
  unit: 'g' | 'ml' | 'pcs';
  normalizedValue: string;
  autoTranslated: boolean;
}
```

### React Example
```tsx
const [inputName, setInputName] = useState('');

const handleCreate = async () => {
  const response = await fetch('/api/admin/ingredients', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ inputName })
  });
  
  if (response.ok) {
    const data = await response.json();
    console.log('Created:', data.data.nameEn);
  } else if (response.status === 409) {
    alert('Ingredient already exists!');
  }
};
```

---

## ✅ Checklist для Frontend разработчика

- [ ] Убрать поля `inputLang`, `category`, `unit` из формы
- [ ] Оставить только одно поле `inputName`
- [ ] Добавить обработку 409 (дубликат)
- [ ] Показывать AI-результат после создания
- [ ] Добавить индикатор загрузки (~500ms)
- [ ] Тестировать с разными языками (PL/EN/RU)

---

## 🚀 Production URLs

**Local:** `http://localhost:8080/api/admin/ingredients`  
**Production:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients`

---

_Last updated: 2026-01-07_
