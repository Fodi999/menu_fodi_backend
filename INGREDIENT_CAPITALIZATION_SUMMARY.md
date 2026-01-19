# Капитализация ингредиентов - Краткая памятка

## ✅ Что сделано

**Проблема:** AI возвращал названия в произвольном регистре (`bacon`, `BACON`, `Bacon`)

**Решение:** Backend автоматически капитализирует все названия перед сохранением в БД

## 🔠 Правила капитализации

| Поле | Формат | Пример |
|------|--------|--------|
| `name_pl` | **Capitalize** | `Boczek` |
| `name_en` | **Capitalize** | `Bacon` |
| `name_ru` | **Capitalize** | `Бекон` |
| `normalized_value` | **lowercase** | `bacon` |
| `category` | **lowercase** | `meat` |
| `unit` | **lowercase** | `g` |
| `nutrition_group` | **lowercase** | `protein` |

## 📁 Файлы

```
pkg/utils/text.go               # Функция Capitalize()
pkg/utils/text_test.go          # 12 тестов (все прошли)
internal/modules/admin/service/service.go  # Применено в CreateIngredientWithAI
```

## 🔧 Использование

```go
import "your_project/pkg/utils"

// Любой UTF-8 текст
utils.Capitalize("bacon")    // → "Bacon"
utils.Capitalize("бекон")    // → "Бекон"
utils.Capitalize("łosoś")    // → "Łosoś"
utils.Capitalize("  test  ") // → "Test" (trim spaces)
```

## 🧪 Примеры трансформации

```
Input (AI)     →  Output (DB)
------------      ------------
"bacon"        →  "Bacon"
"EGGS"         →  "EGGS"  (keeps rest as-is)
"бекон"        →  "Бекон"
"łosoś"        →  "Łosoś"
"ёлка"         →  "Ёлка"
"olive oil"    →  "Olive oil"
""             →  ""
"  test  "     →  "Test"
```

## ✨ Преимущества

- ✅ Консистентность в UI
- ✅ Независимость от AI
- ✅ UTF-8 поддержка (RU/PL/EN/любые)
- ✅ Не ломает диакритику
- ✅ Покрыто тестами
- ✅ Без зависимостей

## 🎯 Результат

**До:** 
```
bacon / EGGS / Salmon / рис / ŁOSOŚ
```

**После:** 
```
Bacon / Eggs / Salmon / Рис / Łosoś
```

---

**Status:** ✅ Production ready  
**Commit:** `923e145`  
**Tests:** 12/12 passed
