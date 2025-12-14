# 🚀 INGREDIENT API - QUICK REFERENCE

## 📖 CATALOG ENDPOINTS (ALL users)

### Search (Autocomplete)
```bash
GET /api/catalog/ingredients/search?query=ml

# Response: Array of ingredients (20 max)
[
  {
    "id": "uuid",
    "name": "Mleko 2%",
    "unit": "ml",
    "category": "dairy",
    "defaultShelfLifeDays": 7,
    "defaultPricePerUnit": 0.003
  }
]
```

### List with Filters
```bash
GET /api/catalog/ingredients
GET /api/catalog/ingredients?category=protein
GET /api/catalog/ingredients?search=kur
GET /api/catalog/ingredients?category=vegetable&search=pom

# Response: Structured with count
{
  "success": true,
  "count": 12,
  "items": [...]
}
```

**Categories:** `protein`, `vegetable`, `dairy`, `grain`, `condiment`, `other`

---

## 📦 STOCK ENDPOINTS (PRO_CHEF only)

```bash
GET    /api/stock              # List stock items
POST   /api/stock              # Add to stock
GET    /api/stock/{id}         # Item details
PUT    /api/stock/{id}         # Update quantity
DELETE /api/stock/{id}         # Remove from stock
GET    /api/stock/{id}/movements  # Movement history
```

---

## 🔐 ADMIN ENDPOINTS (ADMIN only)

### Bulk Import
```bash
POST /api/admin/ingredients/import
Content-Type: application/json

[
  {
    "name": "New Product",
    "unit": "g",
    "category": "protein",
    "defaultShelfLifeDays": 7,
    "defaultPricePerUnit": 0.025
  }
]

# Response
{
  "success": true,
  "imported": 1,
  "total": 1
}
```

---

## 🎯 KEY DIFFERENCES

| Feature | Catalog | Stock |
|---------|---------|-------|
| **Route** | `/catalog/ingredients/*` | `/stock/*` |
| **Access** | ALL authenticated | PRO_CHEF only |
| **Purpose** | Browse products | Manage inventory |
| **Model** | Ingredient | StockItem |
| **Operations** | Read-only | Full CRUD |

---

## ✅ TEST COMMANDS

```bash
# Set token
export TOKEN="your_jwt_token"

# Test autocomplete
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/catalog/ingredients/search?query=ml"

# Test category filter
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/catalog/ingredients?category=protein"

# Test stock (requires pro_chef)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/stock"

# Test admin import (requires admin)
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[{"name":"Test","unit":"g","category":"other"}]' \
  "http://localhost:8080/api/admin/ingredients/import"
```

---

**Updated:** 14 декабря 2025  
**Status:** ✅ Production Ready
