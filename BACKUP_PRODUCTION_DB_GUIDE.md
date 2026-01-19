# Backup Production Database - Инструкция

## ❌ Проблема: Нет прямого доступа к pg_dump

У нас нет локального доступа с паролем к Neon PostgreSQL production БД.

## ✅ Решение: Используем Neon Console Snapshots

Neon PostgreSQL автоматически создаёт snapshots каждый день + при каждом изменении schema.

### Шаг 1: Проверить существующие snapshots

1. Откройте [Neon Console](https://console.neon.tech)
2. Выберите проект `ep-soft-mud-agon8wu3`
3. Перейдите в **Branches** tab
4. Проверьте что `main` branch имеет recent snapshots

### Шаг 2: Создать backup branch (РЕКОМЕНДУЕТСЯ)

```bash
# Создать backup branch перед миграцией
# Neon Console → Branches → "Create branch from main"
# Name: "backup_before_canonical_migration_2026_01_18"
```

Это создаст **point-in-time snapshot** текущего состояния БД.

### Шаг 3: Alternative - Local backup через API

Если нужен локальный backup:

```bash
# Получить DATABASE_URL из Koyeb
# Koyeb Dashboard → Service → Settings → Environment Variables
# Скопировать значение DATABASE_URL

export DATABASE_URL="postgresql://fodi999:YOUR_PASSWORD@ep-soft-mud-agon8wu3.c-2.eu-central-1.aws.neon.tech/recipe_matching_db"

# Создать backup
pg_dump "$DATABASE_URL" > backup_before_canonical_migration.sql
```

## ⚠️ ВАЖНО: Не выполнять миграцию без backup

Без backup нельзя откатить изменения если что-то пойдёт не так.

## Текущий статус

- ✅ Commit pushed (4f94695)
- ⏳ **Backup pending** (ждём подтверждения)
- ⏳ Dry-run миграции
- ⏳ Выполнение миграции
- ⏳ Constraints

## Следующий шаг

**Получить подтверждение от пользователя:**
- Есть ли доступ к Neon Console?
- Создан ли backup branch?
- Или нужна помощь с pg_dump?
