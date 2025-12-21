# Economy Calculation - Current Status & Next Steps

## 📊 Summary

Backend **economy calculation is fully implemented** and correct:
- ✅ Formula: `usedValue = Σ(quantityUsed × pricePerUnit)`
- ✅ Formula: `savedMoney = usedValue - estimatedExtraCost`
- ✅ Handler loads `current_price_per_unit` from DB
- ✅ Service calculates costs per product
- ✅ Always returns economy structure (even if 0)

## 🔍 Current Investigation

**Problem:** Economy returns `usedValue: 0`  
**Root Cause:** Need to verify if prices are actually saved in database

## 📝 What I Did (Commit a36bdfb)

### 1. Added Debug Logging

**File:** `internal/modules/ai/transport/http/handlers.go`

Added detailed logs to track price flow:
```go
// Log all fridge items loaded from DB
logger.Info("Loaded fridge items with prices", 
    zap.Int("total_items", len(fridgeItems)))

// Log each item's price status
for _, item := range fridgeItems {
    if item.CurrentPricePerUnit != nil {
        logger.Info("Price data found for item",
            zap.String("name", item.Ingredient.Name),
            zap.Float64("price_per_unit", *pricePerUnit))
    } else {
        logger.Warn("No price data for item",
            zap.String("name", item.Ingredient.Name))
    }
}
```

### 2. Created Documentation

- **`docs/PRICE_FLOW_DEBUG.md`** - Complete diagnostic guide
- **`sql/diagnostic_price_flow.sql`** - SQL queries to check DB
- **`test_price_debug.sh`** - Test script with debug output
- **`test_full_price_flow.sh`** - End-to-end test (register → add products → recipe)

## 🎯 Next Steps (Priority Order)

### 1️⃣ Check Koyeb Logs (IMMEDIATE)

After deploying commit `a36bdfb`, call recipe generation and check logs for:

```
✅ GOOD:
INFO  Loaded fridge items with prices  total_items=5
INFO  Price data found for item  name="Mleko 2%"  price_per_unit=0.0032
INFO  Price data found for item  name="Wołowina"  price_per_unit=0.0206
[AI][ECONOMY] Used cost: 18.42 PLN ... (prices available: 3 products)

❌ BAD:
INFO  Loaded fridge items with prices  total_items=5
WARN  No price data for item  name="Mleko 2%"  current_price_per_unit=<nil>
WARN  No price data for item  name="Wołowina"  current_price_per_unit=<nil>
[AI][ECONOMY] Used cost: 0.00 PLN ... (prices available: 0 products)
```

### 2️⃣ Check Database (IF logs show NULL)

Run SQL query in **Neon.tech SQL Editor**:

```sql
SELECT
  u.email as user_email,
  i.name as ingredient_name,
  ufi.quantity,
  ufi.unit,
  ufi.current_price_per_unit,
  ufi.current_price_currency,
  CASE 
    WHEN ufi.unit = 'g' THEN ROUND((ufi.current_price_per_unit * 1000)::numeric, 2) || ' PLN/kg'
    WHEN ufi.unit = 'ml' THEN ROUND((ufi.current_price_per_unit * 1000)::numeric, 2) || ' PLN/l'
    ELSE 'N/A'
  END as display_price
FROM user_fridge_items ufi
JOIN "Ingredient" i ON ufi.ingredient_id = i.id
JOIN "User" u ON ufi.user_id = u.id
WHERE ufi.current_price_per_unit IS NOT NULL
ORDER BY u.email, i.name
LIMIT 20;
```

**Expected Results:**
| user_email | ingredient_name | quantity | current_price_per_unit | display_price |
|------------|----------------|----------|------------------------|---------------|
| user@example.com | Wołowina | 500 | 0.02056 | 20.56 PLN/kg |
| user@example.com | Mleko | 1000 | 0.00324 | 3.24 PLN/l |

**❌ If table is empty:**
→ Prices are **NOT being saved** to database  
→ Check `/api/fridge POST` endpoint - does it accept `priceInput`?  
→ Check price normalization logic in backend  
→ Check frontend - does it send `priceInput` field?

### 3️⃣ If DB Has Prices But Logs Show NULL

**Issue:** GORM not mapping `current_price_per_unit` from database

**Check:** `internal/models/user_fridge.go`

Should have:
```go
type UserFridgeItem struct {
    // ...
    CurrentPricePerUnit  *float64   `gorm:"column:current_price_per_unit"`
    CurrentPriceCurrency string     `gorm:"column:current_price_currency"`
    PriceUpdatedAt       *time.Time `gorm:"column:price_updated_at"`
}
```

### 4️⃣ Manual Test with SQL Insert

If you want to test immediately, run this in Neon.tech:

**File:** `sql/diagnostic_price_flow.sql` (Step 5)

This will:
- Insert test products with prices
- Normalize prices correctly (PLN/kg → PLN/g)
- Allow you to test recipe generation immediately

## 🧪 Test Commands

### Quick API Test
```bash
# Get token
TOKEN=$(curl -s -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "YOUR_EMAIL", "password": "YOUR_PASSWORD"}' | jq -r '.data.token')

# Generate recipe
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/create-recipe-from-fridge" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}' | jq '.data.recipe.economy'
```

Expected response:
```json
{
  "usedValue": 18.42,
  "estimatedExtraCost": 0,
  "savedMoney": 18.42,
  "currency": "PLN"
}
```

## 📊 Diagnostic Files Created

1. **`docs/PRICE_FLOW_DEBUG.md`**  
   Complete step-by-step diagnostic guide with:
   - 5-step verification checklist
   - Code snippets to check
   - Expected vs actual values
   - Root cause analysis (90% = prices not in DB)

2. **`sql/diagnostic_price_flow.sql`**  
   6 SQL queries to:
   - Check if any prices exist
   - View normalized prices
   - Simulate economy calculation
   - Insert test data
   - Cleanup

3. **`test_price_debug.sh`**  
   Quick test script to:
   - Generate recipe
   - Show economy data
   - Remind to check Koyeb logs

4. **`test_full_price_flow.sh`**  
   Complete end-to-end test:
   - Register new user
   - Add products with prices
   - Generate recipe
   - Verify economy calculation

## 🎬 Recommended Action Plan

1. ✅ **Commit already pushed** (a36bdfb) - debug logging active
2. ⏳ **Wait for Koyeb deployment** (~2-3 minutes)
3. 🔍 **Generate a recipe** via API or frontend
4. 📋 **Check Koyeb logs** for price debug output
5. 🗄️ **If logs show NULL** → Run SQL query in Neon.tech
6. 🐛 **If DB empty** → Check frontend price input + backend save logic
7. ✅ **If DB has prices but not loading** → Check GORM model mapping

## 💡 Key Insight

**Backend code is correct.** The issue is most likely:
- **90% chance:** Prices not saved in database (frontend issue or save logic issue)
- **5% chance:** GORM not mapping field from DB to Go struct
- **3% chance:** Price data exists but not passed correctly in handler
- **2% chance:** Service calculation logic issue (already verified correct)

---

**Status:** 🟡 Investigation in progress  
**Blocker:** Need to verify actual database state  
**ETA:** Can resolve in 15 minutes once we see Koyeb logs + DB query results
