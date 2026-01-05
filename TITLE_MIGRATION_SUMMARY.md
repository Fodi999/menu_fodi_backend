# Title Column Migration Summary

## ✅ Completed

### Migration 063: Add Title Column to Recipe Table

**Date:** January 5, 2026  
**Commit:** 4c6c968

### What Was Done

1. **Database Migration**
   - Added `title TEXT NOT NULL` column to `Recipe` table
   - Populated from `name_pl` (Polish name) for all 10 recipes
   - Made NOT NULL after data population

2. **Go Model Update**
   - Updated `RecipeCatalog` struct with `Title` field
   - Type: `string` (NOT NULL)
   - JSON tag: `"title"`

3. **Data Verification**
   ```
   Greek Salad           → Sałatka grecka
   Scrambled Eggs        → Jajecznica
   Polish Meat Dumplings → Pierogi z mięsem
   Pierogi Ruskie        → Pierogi ruskie
   Spaghetti Carbonara   → Spaghetti alla Carbonara
   ```

### Schema Changes

**Before:**
```sql
- canonicalName (English)
- localName (DEPRECATED)
- name_pl, name_en, name_ru (multilingual)
```

**After:**
```sql
- canonicalName (English)
- title (NOT NULL, unified - typically Polish)
- localName (DEPRECATED)
- name_pl, name_en, name_ru (multilingual)
```

### API Response

Now includes `title` field:
```json
{
  "id": "uuid",
  "canonicalName": "Pierogi Ruskie",
  "title": "Pierogi ruskie",
  "namePl": "Pierogi ruskie",
  "nameEn": "Pierogi Ruskie",
  "nameRu": "Пироги рускье"
}
```

### Benefits

✅ **Unified Title** - Single primary title field (no more confusion)  
✅ **NOT NULL** - Guaranteed to always have a value  
✅ **Backend Match** - Code now matches database schema  
✅ **Backward Compatible** - Old fields (`name_pl`, etc.) still exist  
✅ **Populated** - All 10 recipes have titles  

### Deployment

- Migration applied: ✅
- Go model updated: ✅
- Build successful: ✅
- Deployed to production: ✅

### Next Steps

- Frontend can now use `recipe.title` as primary display name
- Consider deprecating `localName` completely
- Update API documentation to reflect new field

---

**Migration File:** `migrations/063_add_title_to_recipe.sql`  
**Model File:** `internal/models/recipe_catalog.go`
