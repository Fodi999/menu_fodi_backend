#!/bin/bash

# Apply migration 034: Add "Olej roślinny" to ingredient catalog
# Usage: ./apply_migration_034.sh

echo "=== Migration 034: Add Vegetable Oil ==="
echo ""
echo "📋 This will add 'Olej roślinny' to Ingredient catalog"
echo ""
echo "🔍 Run this SQL on Neon.tech dashboard:"
echo ""
cat migrations/034_add_vegetable_oil.sql
echo ""
echo "---"
echo ""
echo "✅ After applying, verify with:"
echo ""
echo "SELECT id, name, category, unit, \"defaultPricePerUnit\""
echo "FROM \"Ingredient\""
echo "WHERE name = 'Olej roślinny';"
echo ""
echo "Expected result: 1 row"
echo "  name: Olej roślinny"
echo "  category: condiment"
echo "  unit: ml"
echo "  defaultPricePerUnit: 0.005"
