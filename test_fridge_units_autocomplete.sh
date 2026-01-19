#!/bin/bash

# 🧪 Тест автозаполнения единиц измерения в холодильнике
# Проверяет, что API возвращает unit для автокомплита

echo "🧪 Тест 1: Autocomplete для 'молоко' (RU)"
echo "============================================"
curl -s -H "Accept-Language: ru" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=молок&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "🧪 Тест 2: Autocomplete для 'яйца' (RU)"
echo "============================================"
curl -s -H "Accept-Language: ru" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=яйц&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "🧪 Тест 3: Autocomplete для 'мука' (RU)"
echo "============================================"
curl -s -H "Accept-Language: ru" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=мук&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "🧪 Тест 4: Autocomplete для 'масло' (RU)"
echo "============================================"
curl -s -H "Accept-Language: ru" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=масл&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "🧪 Тест 5: Autocomplete на польском (PL)"
echo "============================================"
curl -s -H "Accept-Language: pl" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=mleko&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "🧪 Тест 6: Autocomplete на английском (EN)"
echo "============================================"
curl -s -H "Accept-Language: en" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/ingredients/suggest?q=milk&limit=3" \
  | jq '.data[] | {name, unit, category}'

echo ""
echo "✅ Проверьте, что все ответы содержат поле 'unit' (ml/g/pcs)"
echo ""
echo "📊 Ожидаемые единицы:"
echo "  - Молоко: ml (миллилитры)"
echo "  - Яйца: pcs (штуки)"
echo "  - Мука: g (граммы)"
echo "  - Масло: ml (миллилитры)"
