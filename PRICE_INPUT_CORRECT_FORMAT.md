# 💰 ПРАВИЛЬНЫЙ формат цены: за кг, литр, штуку (НЕ за грамм/мл!)

## 🎯 Правило

**Цена ВСЕГДА вводится за крупные единицы:**
- ✅ **За килограмм (kg)** - для продуктов в граммах
- ✅ **За литр (l)** - для продуктов в миллилитрах
- ✅ **За штуку (pcs/szt)** - для штучных продуктов

❌ **НЕ за грамм (g), миллилитр (ml)** - это неудобно для пользователя!

---

## 📊 Примеры правильного использования

### Пример 1: Молоко (unit = "ml")

**Фронтенд отправляет:**
```json
{
  "ingredientId": "uuid",
  "quantity": 1000,          // 1 литр = 1000 мл
  "unit": "ml",
  "priceInput": {
    "value": 4.40,           // 4.40 PLN за 1 ЛИТР
    "per": "l"               // 👈 За ЛИТР, не за мл!
  }
}
```

**Backend пересчитывает:**
```
4.40 PLN / l → 0.0044 PLN / ml
Сохраняется: 0.0044 PLN/ml в user_fridge_price_history
```

---

### Пример 2: Лосось (unit = "g")

**Фронтенд отправляет:**
```json
{
  "ingredientId": "uuid",
  "quantity": 500,           // 500 грамм
  "unit": "g",
  "priceInput": {
    "value": 32.00,          // 32 PLN за 1 КИЛОГРАММ
    "per": "kg"              // 👈 За КИЛОГРАММ, не за грамм!
  }
}
```

**Backend пересчитывает:**
```
32.00 PLN / kg → 0.032 PLN / g
Сохраняется: 0.032 PLN/g в user_fridge_price_history
```

---

### Пример 3: Яйца (unit = "pcs")

**Фронтенд отправляет:**
```json
{
  "ingredientId": "uuid",
  "quantity": 10,            // 10 штук
  "unit": "pcs",
  "priceInput": {
    "value": 1.20,           // 1.20 PLN за 1 ШТУКУ
    "per": "pcs"             // 👈 За ШТУКУ
  }
}
```

**Backend пересчитывает:**
```
1.20 PLN / pcs → 1.20 PLN / pcs (без изменений)
Сохраняется: 1.20 PLN/pcs в user_fridge_price_history
```

---

## 🔧 Mapping: unit → priceInput.per

| unit в БД | priceInput.per | Пример |
|-----------|----------------|--------|
| `g` | `kg` | Мука, мясо, сахар → цена за кг |
| `ml` | `l` | Молоко, масло, сок → цена за литр |
| `pcs` | `pcs` | Яйца, помидоры → цена за штуку |
| `szt` | `szt` | Польские штучные товары |

---

## 📝 Код для фронтенда

```typescript
// Определяем правильную единицу для цены
const getPriceUnit = (unit: string): string => {
  switch (unit) {
    case 'g':
      return 'kg';   // Цена за килограмм
    case 'ml':
      return 'l';    // Цена за литр
    case 'pcs':
    case 'szt':
      return 'pcs';  // Цена за штуку
    default:
      return unit;
  }
};

// При отправке формы
const handleSubmit = () => {
  const priceUnit = getPriceUnit(selectedIngredient.unit);
  
  const payload = {
    ingredientId: selectedIngredient.id,
    quantity: parseFloat(quantity),
    unit: selectedIngredient.unit,  // ml, g, pcs (из БД)
    priceInput: {
      value: parseFloat(price),
      per: priceUnit                // kg, l, pcs (для цены!)
    },
    expiresAt: expiryDate.toISOString()
  };

  await fridgeApi.addItem(payload);
};
```

---

## 🎨 UI для ввода цены

### ✅ Правильный UI:

```tsx
<div>
  <label>💰 Цена (опционально)</label>
  <input 
    type="number" 
    step="0.01"
    value={price}
    onChange={(e) => setPrice(e.target.value)}
    placeholder="например, 4.40"
  />
  <small>
    PLN за {getPriceUnitLabel(selectedIngredient.unit)}
  </small>
</div>

// Функция для красивых лейблов
const getPriceUnitLabel = (unit: string): string => {
  switch (unit) {
    case 'g': return '1 kg (килограмм)';
    case 'ml': return '1 l (литр)';
    case 'pcs': return '1 szt (штуку)';
    default: return 'единицу';
  }
};
```

**Результат для молока:**
```
💰 Цена (опционально)
┌─────────────────────────┐
│ 4.40                    │
└─────────────────────────┘
PLN за 1 l (литр)
```

---

## ❌ Неправильные варианты

### Вариант 1: Цена за мелкие единицы

```json
{
  "priceInput": {
    "value": 0.0044,  // ❌ Неудобно вводить!
    "per": "ml"
  }
}
```

**Проблема:** Пользователь не знает цену за 1 мл, только за литр!

---

### Вариант 2: Общая цена вместо цены за единицу

```json
{
  "priceTotal": 4.40  // ❌ Потеряли информацию о цене за единицу!
}
```

**Проблема:** Нельзя рассчитать стоимость рецептов, отследить изменение цен!

---

## 🧮 Как backend рассчитывает общую стоимость

1. **Пользователь вводит:** 4.40 PLN за литр
2. **Backend нормализует:** 4.40 / 1000 = 0.0044 PLN/ml
3. **Backend сохраняет:** `current_price_per_unit = 0.0044`
4. **При расчете рецепта:** 
   - Нужно 500 мл молока
   - Стоимость = 500 * 0.0044 = 2.20 PLN

---

## ✅ Итоговая логика

**Фронтенд:**
1. Пользователь выбирает продукт → получает `unit` (g, ml, pcs)
2. Конвертирует `unit` → `priceUnit` (kg, l, pcs)
3. Показывает UI: "PLN за 1 kg" / "PLN за 1 l" / "PLN за 1 szt"
4. Отправляет `priceInput: {value, per: priceUnit}`

**Backend:**
1. Получает `priceInput: {value: 4.40, per: "l"}`
2. Нормализует: `4.40 / 1000 = 0.0044 PLN/ml`
3. Сохраняет в БД: `current_price_per_unit = 0.0044`
4. Использует для расчета стоимости рецептов

---

## 🚀 Поддерживаемые форматы

| priceInput.per | unit | Конвертация | Пример |
|----------------|------|-------------|--------|
| `kg` | `g` | ÷ 1000 | 32 PLN/kg → 0.032 PLN/g |
| `l` | `ml` | ÷ 1000 | 4.40 PLN/l → 0.0044 PLN/ml |
| `pcs` | `pcs` | без изменений | 1.20 PLN/pcs → 1.20 PLN/pcs |
| `szt` | `szt` | без изменений | 1.20 PLN/szt → 1.20 PLN/szt |

---

## ⚠️ Что НЕ поддерживается (специально!)

| priceInput.per | Причина |
|----------------|---------|
| `g` | ❌ Цена за грамм - неудобно для пользователя |
| `ml` | ❌ Цена за миллилитр - неудобно для пользователя |
| `total` | ❌ Общая цена - потеряем нормализацию |

**Backend вернет ошибку:**
```json
{
  "error": "invalid price unit: ml (must be: kg for grams, l for ml, pcs/szt for pieces)"
}
```

---

## 📌 Резюме

**Что изменить на фронтенде:**

1. ✅ Конвертировать `unit` → `priceUnit`:
   - `g` → `kg`
   - `ml` → `l`
   - `pcs` → `pcs`

2. ✅ Показывать в UI: "PLN за 1 kg", "PLN за 1 l"

3. ✅ Отправлять:
   ```json
   {
     "unit": "ml",           // Из БД
     "priceInput": {
       "value": 4.40,
       "per": "l"            // Конвертированная единица!
     }
   }
   ```

**Почему это правильно:**
- 💰 Удобно для пользователя (знают цену за кг/литр)
- 📊 Правильная нормализация для аналитики
- 🧮 Точный расчет стоимости рецептов
- 🌍 Сравнение цен между продуктами

---

**Готово! Теперь цены вводятся в естественном формате.** 🎉
