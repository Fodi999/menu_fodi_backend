# 🧪 Backend Development Scripts

Быстрые тесты без UI для разработки и отладки.

## 📁 Структура

```
scripts/
├── login_user.sh           # Логин как home_chef
├── login_admin.sh          # Логин как super_admin
├── test_egg_scenario.sh    # E2E тест: сценарий "Яичница"
└── quick_test.sh           # Быстрые команды для ежедневной работы
```

---

## 🚀 Быстрый старт

### 1. Получить токены

```bash
# Пользователь (home_chef)
./scripts/login_user.sh

# Админ (super_admin)
./scripts/login_admin.sh
```

Скопируй токен из вывода:
```bash
export USER_TOKEN="eyJhbGc..."
export ADMIN_TOKEN="eyJhbGc..."
```

---

### 2. Быстрые тесты (quick_test.sh)

#### Холодильник
```bash
./scripts/quick_test.sh fridge-list              # Список продуктов
./scripts/quick_test.sh fridge-add               # Добавить чеснок (5g)
./scripts/quick_test.sh fridge-add <id> 10.5     # Добавить продукт (количество)
./scripts/quick_test.sh fridge-delete <id>       # Удалить продукт
```

#### Уведомления
```bash
./scripts/quick_test.sh notif-count              # Количество непрочитанных
./scripts/quick_test.sh notif-list               # Список уведомлений
./scripts/quick_test.sh notif-read <id>          # Отметить прочитанным
./scripts/quick_test.sh notif-read-all           # Отметить все
```

#### AI
```bash
./scripts/quick_test.sh ai-cook                  # Что я могу приготовить?
./scripts/quick_test.sh ai-fridge                # Умные инсайты холодильника
```

#### Ингредиенты
```bash
./scripts/quick_test.sh ingredient-search eggs   # Поиск ингредиентов
./scripts/quick_test.sh ingredient-search czosnek
```

#### Flow тесты
```bash
./scripts/quick_test.sh flow-add-delete          # Полный цикл: добавить → удалить → уведомления
```

---

### 3. E2E тест: Сценарий "Яичница"

**Цель:** Проверить, что AI использует каталог рецептов (не генерирует).

```bash
./scripts/test_egg_scenario.sh
```

**Что тестируется:**
1. ✅ Админ создаёт профессиональный рецепт "Яичница"
2. ✅ Пользователь добавляет все ингредиенты в холодильник
3. ✅ AI находит рецепт из каталога (source = professional)
4. ✅ Coverage = 100%, без AI-генерации

**Ожидаемый результат:**
```
✅ ✅ ✅ ТЕСТ ПРОЙДЕН! ✅ ✅ ✅

🎉 Система работает правильно:
   ✓ Рецепт взят из каталога (source = professional)
   ✓ Все ингредиенты есть (coverage = 100%)
   ✓ AI работает как диспетчер, а не как генератор
   ✓ Пользователь получает лучшее решение
```

---

## 🛠️ Makefile Commands

Для ещё большего удобства:

```bash
make help                    # Показать все команды
make login-user              # Логин как пользователь
make login-admin             # Логин как админ
make test-egg                # E2E тест "Яичница"
make test-notifications      # Тест уведомлений
make build                   # Собрать бинарник
make deploy                  # Закоммитить и задеплоить
```

---

## 📊 Примеры использования

### Пример 1: Тест жизненного цикла продукта

```bash
# 1. Проверить текущие уведомления
./scripts/quick_test.sh notif-count
# → {"count": 5}

# 2. Добавить продукт
./scripts/quick_test.sh fridge-add
# → Product added: d79d05cf-...

# 3. Проверить уведомление о добавлении
./scripts/quick_test.sh notif-count
# → {"count": 6}  ✅ +1

# 4. Удалить продукт
./scripts/quick_test.sh fridge-delete d79d05cf-...
# → Deleted successfully

# 5. Проверить уведомление об удалении
./scripts/quick_test.sh notif-count
# → {"count": 7}  ✅ +1
```

### Пример 2: AI рекомендации

```bash
# 1. Добавить ингредиенты
./scripts/quick_test.sh fridge-add <eggs_id> 2
./scripts/quick_test.sh fridge-add <oil_id> 20
./scripts/quick_test.sh fridge-add <salt_id> 5

# 2. Спросить AI
./scripts/quick_test.sh ai-cook
# → {
#     "title": "Яичница",
#     "coverage": 100,
#     "source": "professional"
#   }  ✅ Рецепт из каталога!
```

### Пример 3: Автоматический flow

```bash
./scripts/quick_test.sh flow-add-delete
# → Автоматически:
#    1. Добавляет продукт
#    2. Проверяет уведомление (+1)
#    3. Удаляет продукт
#    4. Проверяет уведомление (+1)
#    5. Валидирует результат
```

---

## 🎯 Философия тестирования

### ✅ Правильный подход (что мы делаем)

```
Тестируем РЕАЛЬНУЮ продуктовую логику:
  ↓
AI = диспетчер (выбирает источник)
  ↓
Каталог рецептов = база знаний
  ↓
Пользователь выбирает ЦЕЛЬ, не ТИП рецепта
```

### ❌ Неправильный подход (чего избегаем)

```
Кликать в UI → медленно
Вызывать AI напрямую → обход Rules Engine
Генерировать когда есть в каталоге → неэффективно
```

---

## 🔧 Настройка

### Обновить токены

Если токены истекли:

```bash
# 1. Получить новый токен
./scripts/login_user.sh

# 2. Обновить в quick_test.sh
vim scripts/quick_test.sh
# Найти: USER_TOKEN="..."
# Вставить новый токен
```

### Изменить BASE_URL

```bash
# Для локального тестирования
export BASE_URL="http://localhost:8080"
./scripts/quick_test.sh notif-count

# Для продакшена (по умолчанию)
unset BASE_URL
./scripts/quick_test.sh notif-count
```

---

## 📋 Чеклист тестирования

Перед деплоем проверь:

- [ ] `make test-egg` → ✅ PASSED
- [ ] `./scripts/quick_test.sh flow-add-delete` → ✅ PASSED
- [ ] `./scripts/quick_test.sh ai-cook` → source = professional
- [ ] `./scripts/quick_test.sh notif-count` → работает
- [ ] Логи на Koyeb без ошибок

---

## 🎉 Результат

**Скорость разработки:**
- ❌ Было: открыть UI → войти → 10 кликов → 2 минуты
- ✅ Стало: `./scripts/quick_test.sh flow-add-delete` → 3 секунды

**Качество тестов:**
- ✅ Реальная продуктовая логика
- ✅ Полное покрытие flow
- ✅ Автоматическая валидация
- ✅ Понятные сообщения об ошибках

---

## 📚 Дополнительно

- `NOTIFICATION_LIFECYCLE_COMPLETE.md` - Документация по уведомлениям
- `NOTIFICATIONS_QUICK_REF.md` - Быстрая справка по API
- `Makefile` - Все команды в одном месте
