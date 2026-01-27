# 🧾 Recipe Cost Calculation API

## Обзор

**Цель:** Рассчитать себестоимость рецепта БЕЗ создания блюда (dish), без UUID-конфликтов, без 500 ошибок.

**Архитектура:** Отдельный calculation endpoint, полностью независим от CRUD операций.

**Дата добавления:** 27 Января 2026

---

## 📌 Новый Endpoint

### GET /api/admin/recipes/{recipeId}/cost

Рассчитывает себестоимость рецепта на одну порцию.

**Требования:**
- ✅ Admin или Super Admin
- ✅ Valid JWT token
- ✅ Valid recipeId (UUID format)

**Параметры пути:**
```
{recipeId} - UUID рецепта из каталога
```

**Пример запроса:**
```bash
curl -X GET "https://api.example.com/api/admin/recipes/550e8400-e29b-41d4-a716-446655440000/cost" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json"
```

---

## 📤 Response (200 OK)

```json
{
  "recipeId": "550e8400-e29b-41d4-a716-446655440000",
  "recipeTitle": "Паста Болоньезе",
  "totalCost": 28.50,
  "costPerServing": 28.50,
  "ingredientsCount": 12,
  "missingPrice": false,
  "details": [
    {
      "ingredientName": "Макароны",
      "quantity": 400,
      "unit": "g",
      "pricePerUnit": 0.0125,
      "itemCost": 5.0,
      "optional": false
    },
    {
      "ingredientName": "Фарш говяжий",
      "quantity": 500,
      "unit": "g",
      "pricePerUnit": 0.045,
      "itemCost": 22.5,
      "optional": false
    },
    {
      "ingredientName": "Помидоры консервированные",
      "quantity": 400,
      "unit": "g",
      "pricePerUnit": 0.01,
      "itemCost": 4.0,
      "optional": false
    },
    {
      "ingredientName": "Чеснок",
      "quantity": 2,
      "unit": "szt",
      "pricePerUnit": 0.5,
      "itemCost": 1.0,
      "optional": true
    }
  ]
}
```

---

## 📊 Поля ответа

| Поле | Тип | Описание |
|------|-----|---------|
| `recipeId` | string (UUID) | ID рецепта |
| `recipeTitle` | string | Название рецепта |
| `totalCost` | number | **Сумма без опциональных ингредиентов (PLN)** |
| `costPerServing` | number | Alias для totalCost (для ясности) |
| `ingredientsCount` | integer | Количество ингредиентов |
| `missingPrice` | boolean | **true** если некоторые ингредиенты без цены |
| `details` | array | Массив деталей ингредиентов |

### Структура details[]:

| Поле | Тип | Описание |
|------|-----|---------|
| `ingredientName` | string | Локализованное название (PL → EN → RU) |
| `quantity` | number | Количество |
| `unit` | string | Единица (g, ml, szt, l, kg и т.д.) |
| `pricePerUnit` | number | Цена за единицу из `defaultPricePerUnit` (PLN) |
| `itemCost` | number | quantity × pricePerUnit |
| `optional` | boolean | Опциональный ли ингредиент? |

---

## ❌ Ошибки

### 400 Bad Request
```json
{
  "error": "recipeId is required",
  "code": 400
}
```
**Причины:**
- recipeId отсутствует
- recipeId некорректный UUID

### 401 Unauthorized
```json
{
  "error": "User not authenticated",
  "code": 401
}
```
**Причины:**
- JWT token отсутствует
- JWT token невалидный
- JWT token истёк

### 403 Forbidden
```json
{
  "error": "Access denied",
  "code": 403
}
```
**Причины:**
- User не admin
- User не super_admin

### 404 Not Found
```json
{
  "error": "Recipe not found",
  "code": 404
}
```
**Причины:**
- Рецепт с таким recipeId не существует
- Рецепт был удален

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "code": 500
}
```

---

## 🔍 Логика расчета

### Алгоритм:

1. **Получить рецепт** из `Recipe` таблицы с `Preload("Ingredients.Ingredient")`
2. **Итерировать по каждому ингредиенту:**
   - Получить `defaultPricePerUnit` из таблицы `Ingredient`
   - Рассчитать: `itemCost = quantity × pricePerUnit`
   - Если ингредиент НЕ optional → добавить в `totalCost`
3. **Нормализировать** до 2 знаков после запятой
4. **Вернуть** полный ответ с деталями

### Что НЕ происходит:

❌ Создание Dish  
❌ Создание UUID конфликтов  
❌ Изменение базы данных  
❌ Чтение stockItem цен (только defaultPricePerUnit)  

---

## 🛠 JavaScript Frontend Example

```javascript
// React Hook для получения стоимости рецепта
const getRecipeCost = async (recipeId) => {
  const token = localStorage.getItem('jwtToken');
  
  try {
    const response = await fetch(
      `/api/admin/recipes/${recipeId}/cost`,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const data = await response.json();
    
    console.log(`Себестоимость: ${data.totalCost} PLN`);
    console.log(`Ингредиентов: ${data.ingredientsCount}`);
    
    if (data.missingPrice) {
      console.warn('⚠️ Некоторые ингредиенты без цены!');
    }

    return data;
  } catch (error) {
    console.error('Ошибка при расчете стоимости:', error);
    throw error;
  }
};

// Использование
const handleShowCost = async (recipeId) => {
  const costData = await getRecipeCost(recipeId);
  
  // Показываем деньги
  setDishCost(costData.totalCost);
  setShowCostDetails(true);
};
```

---

## 📱 React Component Example

```jsx
import React, { useState } from 'react';

const RecipeCostCard = ({ recipeId, recipeName }) => {
  const [cost, setCost] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchCost = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await fetch(`/api/admin/recipes/${recipeId}/cost`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('jwtToken')}`
        }
      });

      if (!response.ok) throw new Error('Failed to calculate cost');

      const data = await response.json();
      setCost(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="recipe-cost-card">
      <h3>{recipeName}</h3>
      
      <button onClick={fetchCost} disabled={loading}>
        {loading ? 'Считаю...' : 'Рассчитать стоимость'}
      </button>

      {cost && (
        <div className="cost-info">
          <h4>Себестоимость: {cost.totalCost} PLN</h4>
          
          {cost.missingPrice && (
            <p className="warning">
              ⚠️ Внимание: {cost.ingredientsCount} ингредиентов без цены
            </p>
          )}

          <table>
            <thead>
              <tr>
                <th>Ингредиент</th>
                <th>Кол-во</th>
                <th>Цена за шт</th>
                <th>Сумма</th>
              </tr>
            </thead>
            <tbody>
              {cost.details.map((item, i) => (
                <tr key={i}>
                  <td>{item.ingredientName}</td>
                  <td>{item.quantity} {item.unit}</td>
                  <td>{item.pricePerUnit} PLN</td>
                  <td>{item.itemCost} PLN</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {error && <p className="error">Ошибка: {error}</p>}
    </div>
  );
};

export default RecipeCostCard;
```

---

## 📋 Чек-лист использования

- ✅ Убедиться что все ингредиенты рецепта имеют `defaultPricePerUnit`
- ✅ Использовать архитектурно правильный URL: `/recipes/{recipeId}/cost` (а не `/dishes/calculate-cost?recipeId=...`)
- ✅ Обрабатывать warning `missingPrice: true`
- ✅ Показывать детали в таблице для прозрачности
- ✅ Кэшировать результат если рецепт не менялся

---

## 🔗 Связанные Endpoints

| Метод | Путь | Описание |
|-------|------|---------|
| GET | `/api/admin/recipes` | Список рецептов |
| GET | `/api/admin/recipes/{id}` | Детали рецепта |
| POST | `/api/admin/dishes/generate-from-recipe` | Создать блюдо из рецепта |
| **GET** | **`/api/admin/recipes/{recipeId}/cost`** | **🆕 Расчет стоимости** |

---

## 📝 История

| Дата | Событие |
|------|---------|
| 27.01.2026 | ✅ Добавлен новый endpoint |
| 27.01.2026 | ✅ Роут расположен ДО {id} для избежания конфликтов |
| 27.01.2026 | ✅ Тестировано и задокументировано |

