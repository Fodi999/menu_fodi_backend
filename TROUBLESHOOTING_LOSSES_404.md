# 🐛 Troubleshooting: 404 на /api/history/losses

## Проблема
Frontend получал 404 при запросе `/api/history/losses?days=30`

## Причина
Koyeb не подхватил последние изменения из коммитов:
- `9eb322a` - добавлен endpoint GetFridgeLosses
- `16fa698` - исправлен формат ответа
- `d7a4784` - добавлена автоматическая очистка

## Решение
Создан пустой коммит `fe1ea90` для форсирования редеплоя:
```bash
git commit --allow-empty -m "chore: trigger Koyeb redeploy for losses endpoint"
git push origin main
```

## Проверка Деплоя

### 1. Дождаться завершения деплоя (2-5 минут)
Koyeb автоматически соберёт и задеплоит новую версию

### 2. Проверить health endpoint
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/health
# Должен вернуть: ok
```

### 3. Проверить losses endpoint (без авторизации)
```bash
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/history/losses?days=30"
# Должен вернуть 401 (unauthorized) - это нормально, endpoint существует
```

### 4. Проверить losses endpoint (с авторизацией)
```bash
TOKEN="your_jwt_token"
curl -H "Authorization: Bearer $TOKEN" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/history/losses?days=30"
# Должен вернуть JSON с events и summary
```

## Ожидаемый Ответ

```json
{
  "events": [
    {
      "id": "uuid",
      "name": "Молоко",
      "quantity": 1.0,
      "unit": "л",
      "loss": 4.5,
      "reason": "expiry_date_passed",
      "addedDate": "2024-12-20T10:30:00Z",
      "expiryDate": "2024-12-25T00:00:00Z",
      "daysInFridge": 5
    }
  ],
  "summary": {
    "totalProducts": 10,
    "totalValue": 45.0,
    "avgValue": 4.5,
    "currency": "PLN"
  }
}
```

## Debug: Проверка Регистрации Роута

### Локально
```bash
go run cmd/server/main.go
# В другом терминале:
curl http://localhost:8080/api/history/losses?days=30
```

### Код Регистрации
- **Module:** `internal/modules/history/module.go:33`
  ```go
  r.Get("/losses", m.handler.GetFridgeLosses)
  ```

- **Handler:** `internal/modules/history/transport/http/handler.go:187`
  ```go
  func (h *HistoryHandler) GetFridgeLosses(w http.ResponseWriter, r *http.Request)
  ```

- **App Routes:** `internal/app/routes_modular.go:136`
  ```go
  historyModule.RegisterRoutes(r, middleware.AuthMiddleware)
  ```

## Известные Проблемы

### Koyeb Cache
Иногда Koyeb кеширует старую версию. Решения:
1. Пустой коммит (как сделали)
2. Перезапуск сервиса в Koyeb UI
3. Изменение переменной окружения в Koyeb (триггерит ребилд)

### Build Errors
Если деплой не прошёл, проверить логи:
1. Зайти в Koyeb Dashboard
2. Открыть service `menu_fodi_backend`
3. Вкладка "Logs" → "Build logs"

## Timeline

- `2025-12-28 14:00` - Frontend сообщил о 404
- `2025-12-28 14:10` - Обнаружено: Koyeb не подхватил изменения
- `2025-12-28 14:12` - Создан force redeploy commit `fe1ea90`
- `2025-12-28 14:15-14:20` - Ожидается завершение деплоя

## Next Steps

1. ⏳ Дождаться деплоя (~5 минут)
2. ✅ Проверить endpoint с фронтенда
3. 🎉 Увидеть корзину отходов с просроченными продуктами!
