#!/bin/bash

# Simple test: Check if add-missing-ingredients endpoint works
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI1NWViMmRiOC1hYWViLTQ1NWEtODFkMy1hZGNjMTEzNjI0ZWYiLCJlbWFpbCI6InJlY2lwZXRlc3RAZm9kaS5hcHAiLCJyb2xlIjoiaG9tZV9jaGVmIiwiZXhwIjoxNzY2NDA2MDE2LCJpYXQiOjE3NjYzMTk2MTZ9.EnB54La8LmmHG0BQ6G3UAjDHcXwUbr-YPTnCKG4nZYg"
BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "🧪 Testing /api/ai/add-missing-ingredients endpoint"
echo ""
echo "Request:"
echo '{
  "ingredients": [
    {"name": "Olej roślinny", "quantity": 15, "unit": "ml"},
    {"name": "Sól", "quantity": 3, "unit": "g"},
    {"name": "NonExistentItem", "quantity": 100, "unit": "g"}
  ]
}'
echo ""

echo "Response:"
curl -s -X POST "$BASE_URL/api/ai/add-missing-ingredients" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "ingredients": [
      {"name": "Olej roślinny", "quantity": 15, "unit": "ml"},
      {"name": "Sól", "quantity": 3, "unit": "g"},
      {"name": "NonExistentItem", "quantity": 100, "unit": "g"}
    ]
  }' | jq '.'

echo ""
echo "✅ Test complete"
