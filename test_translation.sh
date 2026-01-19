#!/bin/bash

# Тест автоперевода рецептов

echo "🧪 Testing recipe auto-translation..."
echo ""

# Создаём рецепт на русском
echo "1️⃣ Creating recipe in Russian..."
RECIPE_ID=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/create-ai \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyX2FkbWluIiwiaGFzUm9sZSI6dHJ1ZSwic3ViIjoiN2VjOGFiYTQtODE5NS00YmUxLWE5YTgtMDY3YzMwYWFlMzA2IiwiZXhwIjoxNzY4OTQ2MjEyLCJpYXQiOjE3Njg4NTk4MTJ9.NDBDgEz3Cgnj4LXQGYgakSMnfaptkjnJwm0c5xKXcpU" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "борщ",
    "rawCookingText": "сварить бульон добавить овощи",
    "language": "ru",
    "ingredients": [
      {
        "ingredientId": "37bf235a-5023-4e7a-915a-ef31c1cd3cd0",
        "quantity": 2,
        "unit": "pcs"
      }
    ]
  }' | jq -r '.data.id')

echo "✅ Recipe created: $RECIPE_ID"
echo ""

# Ждём 5 секунд
echo "⏳ Waiting 5 seconds for auto-translation..."
sleep 5

# Проверяем переводы
echo ""
echo "2️⃣ Checking translations in database..."
psql "postgresql://neondb_owner:npg_dz4Gl8ZhPLbX@ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech/neondb?sslmode=require" -c "
SELECT 
  title,
  name_pl,
  name_en,
  name_ru,
  CASE WHEN description_pl IS NOT NULL THEN '✅ YES' ELSE '❌ NO' END as desc_pl,
  CASE WHEN description_en IS NOT NULL THEN '✅ YES' ELSE '❌ NO' END as desc_en,
  CASE WHEN description_ru IS NOT NULL THEN '✅ YES' ELSE '❌ NO' END as desc_ru
FROM \"Recipe\"
WHERE id = '$RECIPE_ID';
"

echo ""
echo "3️⃣ If translations are missing, wait 10 more seconds..."
sleep 10

psql "postgresql://neondb_owner:npg_dz4Gl8ZhPLbX@ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech/neondb?sslmode=require" -c "
SELECT 
  title,
  name_pl,
  name_en,
  CASE WHEN description_pl IS NOT NULL THEN '✅' ELSE '❌' END as desc_pl,
  CASE WHEN description_en IS NOT NULL THEN '✅' ELSE '❌' END as desc_en
FROM \"Recipe\"
WHERE id = '$RECIPE_ID';
"

echo ""
echo "✅ Test complete!"
