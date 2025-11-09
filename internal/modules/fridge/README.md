# Fridge Module

Fridge Module управляет виртуальным холодильником пользователя - хранилищем ингредиентов и продуктов.

## Архитектура

```
fridge/
├── dto/                   # Data Transfer Objects
│   └── requests.go       # Request/Response DTOs
├── repo/                  # Repository Layer (Database)
│   └── repository.go     # Fridge data access
├── service/              # Business Logic Layer
│   └── service.go        # Fridge business rules
├── transport/            # Transport Layer
│   └── http/
│       └── handlers.go   # HTTP handlers
├── module.go             # Module initialization
└── README.md            # This file
```

## Функциональность

### 1. Item Management
- **GetUserFridge** - получение всех предметов холодильника
- **AddFridgeItem** - добавление нового ингредиента
- **UpdateFridgeItem** - обновление количества или доступности
- **DeleteFridgeItem** - удаление предмета
- **GetAvailableItems** - получение только доступных предметов

## API Endpoints

### Get User Fridge
```http
GET /api/fridge
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "items": [
    {
      "id": "uuid",
      "userId": "uuid",
      "product": "Молоко",
      "quantity": 1.5,
      "unit": "л",
      "available": true,
      "createdAt": "2025-01-15T10:00:00Z",
      "updatedAt": "2025-01-15T10:00:00Z"
    }
  ],
  "count": 1
}
```

### Add Fridge Item
```http
POST /api/fridge
Authorization: Bearer {token}
Content-Type: application/json

{
  "product": "Молоко",
  "quantity": 1.5,
  "unit": "л"
}
```

**Response:**
```json
{
  "success": true,
  "message": "item added to fridge"
}
```

### Update Fridge Item
```http
PUT /api/fridge/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "quantity": 1.0,
  "available": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "item updated successfully",
  "item": {
    "id": "uuid",
    "userId": "uuid",
    "product": "Молоко",
    "quantity": 1.0,
    "unit": "л",
    "available": true,
    "createdAt": "2025-01-15T10:00:00Z",
    "updatedAt": "2025-01-15T12:00:00Z"
  }
}
```

### Delete Fridge Item
```http
DELETE /api/fridge/{id}
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "item deleted successfully"
}
```

### Get Available Items
```http
GET /api/fridge/available
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "items": [
    {
      "id": "uuid",
      "userId": "uuid",
      "product": "Молоко",
      "quantity": 1.5,
      "unit": "л",
      "available": true,
      "createdAt": "2025-01-15T10:00:00Z",
      "updatedAt": "2025-01-15T10:00:00Z"
    }
  ],
  "count": 1
}
```

## Business Logic

### Validation Rules
- **Product name**: не может быть пустым, триммится
- **Quantity**: должно быть > 0
- **Unit**: не может быть пустым, триммится
- **Updates**: хотя бы одно поле должно обновляться

### Item States
- **Available**: true - предмет есть в наличии
- **Available**: false - предмет использован/израсходован

### Authorization
- Пользователь может видеть только свои предметы
- Операции update/delete проверяют принадлежность предмета пользователю

## Database Schema

### UserFridge Table
```sql
CREATE TABLE user_fridge (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  product VARCHAR(255) NOT NULL,
  quantity DECIMAL(10,2) NOT NULL,
  unit VARCHAR(20) NOT NULL,
  available BOOLEAN DEFAULT true,
  category VARCHAR(50),
  expiry_date TIMESTAMP,
  added_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  INDEX idx_user_id (user_id)
);
```

## Error Handling

### Custom Errors
- `ErrItemNotFound` - предмет не найден
- `ErrUnauthorized` - нет доступа к предмету
- `ErrInvalidQuantity` - количество <= 0
- `ErrEmptyProduct` - пустое название продукта
- `ErrEmptyUnit` - пустая единица измерения
- `ErrNoUpdates` - нет данных для обновления

### HTTP Status Codes
- **200 OK** - успешный запрос
- **400 Bad Request** - неверные данные
- **401 Unauthorized** - не авторизован
- **404 Not Found** - предмет не найден
- **500 Internal Server Error** - ошибка сервера

## Dependencies

### External
- `github.com/google/uuid` - UUID support
- `github.com/go-chi/chi/v5` - HTTP router
- `go.uber.org/zap` - структурированное логирование
- `gorm.io/gorm` - ORM для работы с БД

### Internal
- `backend/internal/models` - модели данных
- `backend/internal/middleware` - JWT middleware
- `backend/internal/platform/httpx` - HTTP helpers
- `backend/internal/platform/logger` - логгер

## Usage Example

```go
import (
    "gorm.io/gorm"
    "backend/internal/modules/fridge"
)

// Initialize module
fridgeModule := fridge.NewModule(db)

// Register routes
fridgeModule.RegisterRoutes(router, jwtMiddleware)
```

## Testing

```bash
# Get fridge items
curl -X GET http://localhost:8080/api/fridge \
  -H "Authorization: Bearer YOUR_TOKEN"

# Add item
curl -X POST http://localhost:8080/api/fridge \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product":"Молоко","quantity":1.5,"unit":"л"}'

# Update item
curl -X PUT http://localhost:8080/api/fridge/ITEM_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":1.0}'

# Delete item
curl -X DELETE http://localhost:8080/api/fridge/ITEM_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get available items
curl -X GET http://localhost:8080/api/fridge/available \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Security

- ✅ Все эндпоинты требуют JWT аутентификацию
- ✅ User ID извлекается из JWT токена
- ✅ Проверка принадлежности предмета пользователю
- ✅ Валидация всех входных данных
- ✅ Структурированное логирование операций
- ✅ Защита от SQL injection через GORM

## Integration with Other Modules

### AI Module
- Fridge data используется для генерации рецептов
- Рекомендации рецептов основаны на доступных ингредиентах

### Recipe Module
- При приготовлении рецепта уменьшается количество ингредиентов
- Создаются FridgeTransaction записи для истории

## Future Improvements

- [ ] Автоматическое отслеживание сроков годности
- [ ] Уведомления о скором истечении срока
- [ ] Категоризация продуктов
- [ ] Умные рекомендации по закупкам
- [ ] История использования ингредиентов (FridgeTransaction)
- [ ] Интеграция с shopping list
- [ ] Barcode scanning для быстрого добавления
