#!/bin/bash
# Quick script to check ingredient in catalog
INGREDIENT="$1"
psql "$DATABASE_URL" -c "SELECT id, name, category, default_unit FROM \"Ingredient\" WHERE LOWER(name) LIKE LOWER('%$INGREDIENT%') ORDER BY name;" 2>/dev/null || echo "Need DATABASE_URL env variable"
