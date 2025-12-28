# 🐛 Bug Fix: History Module 404 Error

## Проблема
All history endpoints returned **404 Not Found**:
- `/api/history` → 404
- `/api/history/stats` → 404  
- `/api/history/recent` → 404
- `/api/history/losses?days=30` → 404

## Root Cause
**Double route prefix** in history module registration.

### Код ДО исправления:
```go
// internal/modules/history/module.go
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
    r.Route("/api/history", func(r chi.Router) {  // ❌ ОШИБКА: /api/ префикс
        r.Use(authMiddleware)
        r.Get("/", m.handler.GetHistory)
        r.Get("/losses", m.handler.GetFridgeLosses)
    })
}
```

### Контекст вызова:
```go
// internal/app/routes_modular.go
r.Route("/api", func(r chi.Router) {              // Уже добавляет /api/
    historyModule.RegisterRoutes(r, middleware.AuthMiddleware)
})
```

### Результат:
Фактический путь: `/api/api/history/losses` (двойной `/api`)  
Ожидаемый путь: `/api/history/losses`

## Решение

### Код ПОСЛЕ исправления:
```go
// internal/modules/history/module.go
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
    r.Route("/history", func(r chi.Router) {  // ✅ БЕЗ /api/ префикса
        r.Use(authMiddleware)
        r.Get("/", m.handler.GetHistory)
        r.Get("/losses", m.handler.GetFridgeLosses)
    })
}
```

### Паттерн других модулей:
```go
// internal/modules/fridge/module.go
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
    r.Route("/fridge", func(r chi.Router) {  // ✅ Правильно: без /api/
        r.Use(jwtMiddleware)
        r.Get("/items", m.handlers.GetUserItems)
    })
}
```

## Проверка

### До исправления:
```bash
$ curl http://localhost:8080/api/history/losses?days=30
404 page not found
```

### После исправления:
```bash
$ curl http://localhost:8080/api/history/losses?days=30
# HTTP 401 Unauthorized (endpoint exists, needs auth)
```

### Production проверка:
```bash
$ curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/history/losses?days=30"
# Should return 401 (not 404)
```

## Timeline
- **14:08** - Frontend сообщил о 404 на `/api/history/losses`
- **14:12** - Первая попытка: пустой коммит для redeploy
- **14:15** - Обнаружено: endpoint возвращает 404 даже после деплоя
- **14:20** - Локальная проверка: 404 даже на localhost
- **14:25** - Найдена причина: двойной `/api` prefix
- **14:30** - Исправлено: убран `/api/` из `r.Route()`
- **14:32** - Коммит `d845057` - fix pushed to production

## Commits
- `fe1ea90` - Force redeploy (не помогло)
- `d845057` - **FIX**: correct route registration ✅

## Lesson Learned
❗ При создании нового модуля **НЕ добавлять** `/api/` префикс в `r.Route()`, т.к. модуль уже вызывается внутри `/api` контекста в `routes_modular.go`.

✅ **Правильно:**
```go
r.Route("/history", func(r chi.Router) { ... })
```

❌ **Неправильно:**
```go
r.Route("/api/history", func(r chi.Router) { ... })
```

## Status
✅ **FIXED** - Ожидается деплой на Koyeb (~2 мин)
