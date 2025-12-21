#!/bin/bash

# Quick SQL check for "Olej roślinny" in database
# Copy-paste this SQL into Neon.tech SQL Editor

echo "=== Quick Check: Olej roślinny in Database ==="
echo ""
echo "📋 Copy this SQL and run on Neon.tech:"
echo ""
echo "---SQL START---"
cat << 'EOF'
-- Check if "Olej roślinny" exists
SELECT 
    id,
    name,
    category,
    unit as default_unit,
    "defaultPricePerUnit" as price,
    "createdAt"
FROM "Ingredient"
WHERE name = 'Olej roślinny';
EOF
echo "---SQL END---"
echo ""
echo "Expected result:"
echo "  ✅ 1 row: Olej roślinny | condiment | ml | 0.005"
echo "  ❌ 0 rows: Need to run migration 034"
echo ""
echo "---"
echo ""
echo "Check ALL oils in database:"
echo ""
echo "---SQL START---"
cat << 'EOF'
-- List all oils in catalog
SELECT 
    name,
    category,
    unit,
    "defaultPricePerUnit"
FROM "Ingredient"
WHERE LOWER(name) LIKE '%olej%'
   OR LOWER(name) LIKE '%oil%'
   OR LOWER(name) LIKE '%oliwa%'
ORDER BY name;
EOF
echo "---SQL END---"
echo ""
echo "Expected oils:"
echo "  - Olej rzepakowy (rapeseed)"
echo "  - Olej słonecznikowy (sunflower)"
echo "  - Olej roślinny (vegetable) ← NEW"
echo "  - Oliwa z oliwek (olive oil)"
