# 🔧 User Settings Column Migration - URGENT FIX

## 🔴 Problem

Frontend language switch causes page reload and errors:
```
ERROR: column "settings" of relation "User" does not exist (SQLSTATE 42703)
```

**Root Cause:** Migration `053_add_user_settings.sql` was created but **NOT APPLIED** to production database.

## ✅ Solution

Apply migration to add `settings` JSONB column to `"User"` table.

---

## 📋 Step-by-Step Instructions

### 1️⃣ Connect to Neon PostgreSQL

```bash
# Get connection string from Koyeb environment variables
# Or from Neon dashboard: https://console.neon.tech/

psql "postgresql://user:password@host/database?sslmode=require"
```

### 2️⃣ Verify Current State (BEFORE)

```sql
-- Check if settings column exists
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'User' AND column_name = 'settings';

-- Expected: (0 rows) - column doesn't exist yet
```

### 3️⃣ Apply Migration

**Execute:** `APPLY_USER_SETTINGS_MIGRATION.sql`

Or copy-paste this:

```sql
-- Add settings JSONB column
ALTER TABLE "User"
ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{
  "language": "pl",
  "timeFormat": "24h",
  "units": "metric",
  "aiStyle": "mentor"
}'::jsonb NOT NULL;

-- Add GIN index for performance
CREATE INDEX IF NOT EXISTS idx_user_settings_language 
ON "User" USING gin ((settings->'language'));
```

### 4️⃣ Verify Success (AFTER)

```sql
-- Check column was added
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns 
WHERE table_name = 'User' AND column_name = 'settings';

-- Expected output:
-- column_name | data_type | is_nullable | column_default
-- settings    | jsonb     | NO          | '{"language":"pl",...}'::jsonb

-- Check sample data
SELECT id, email, settings 
FROM "User" 
LIMIT 3;

-- Expected: All users have default settings object
```

---

## 🧪 Test After Migration

### Backend Test (curl)

```bash
# 1. Login to get token
TOKEN=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.data.token')

# 2. Update language to Russian
curl -X PATCH https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language":"ru"}' | jq .

# Expected: 200 OK with updated settings
```

### Frontend Test

1. **Login:** https://menu-fodi.vercel.app/
2. **Go to:** Settings page (`/profile/settings`)
3. **Click:** Russian language button (Русский 🇷🇺)
4. **Expected Results:**
   - ✅ Language changes instantly
   - ✅ NO page reload
   - ✅ Settings persist after refresh
   - ❌ NO console errors

---

## 📊 Migration Impact

### Before Migration ❌
- Frontend: Click language → page reload → error
- Backend: `UPDATE "User" SET settings = ...` → ERROR
- User Experience: Settings don't save

### After Migration ✅
- Frontend: Click language → instant change → no reload
- Backend: `UPDATE "User" SET settings = ...` → SUCCESS
- User Experience: Settings save immediately

---

## 🔍 Technical Details

### Database Schema Change

**Column:** `settings`  
**Type:** `JSONB`  
**Nullable:** `NO`  
**Default:** 
```json
{
  "language": "pl",
  "timeFormat": "24h",
  "units": "metric",
  "aiStyle": "mentor"
}
```

**Index:** `idx_user_settings_language` (GIN index on `settings->'language'`)

### Go Model (Already Correct)

```go
type User struct {
    // ... other fields
    Settings UserSettings `gorm:"type:jsonb;column:settings" json:"settings"`
}

type UserSettings struct {
    Language   Language   `json:"language"`   // pl | en | ru
    TimeFormat TimeFormat `json:"timeFormat"` // 12h | 24h
    Units      Units      `json:"units"`      // metric | kitchen
    AIStyle    AIStyle    `json:"aiStyle"`    // mentor | direct
}
```

### API Endpoint (Already Working)

**PATCH** `/api/settings`
```json
{
  "language": "ru",
  "timeFormat": "24h",
  "units": "metric",
  "aiStyle": "mentor"
}
```

---

## ⚠️ Rollback Plan (If Needed)

If something goes wrong:

```sql
-- Remove settings column
ALTER TABLE "User" DROP COLUMN IF EXISTS settings;

-- Remove index
DROP INDEX IF EXISTS idx_user_settings_language;
```

Then revert frontend to use separate language/timeFormat/units columns.

---

## ✅ Checklist

- [ ] Connect to Neon PostgreSQL
- [ ] Verify settings column doesn't exist
- [ ] Run migration SQL
- [ ] Verify column was added
- [ ] Check all users have default settings
- [ ] Test with curl (backend)
- [ ] Test in UI (frontend)
- [ ] Verify no page reload on language change
- [ ] Confirm settings persist after refresh

---

## 🎯 Expected Timeline

- **Migration time:** ~5 seconds
- **Downtime:** 0 seconds (non-breaking change)
- **Testing:** ~2 minutes
- **Total:** < 5 minutes

---

## 📞 Troubleshooting

### Issue: "column settings already exists"
**Solution:** Column already added, skip migration. Verify with:
```sql
SELECT * FROM "User" LIMIT 1;
```

### Issue: "permission denied"
**Solution:** Use database owner/admin credentials from Neon dashboard.

### Issue: Frontend still shows errors
**Solution:** 
1. Clear browser cache
2. Hard refresh (Cmd+Shift+R)
3. Check backend logs for errors

---

## 📝 Related Files

- Migration: `migrations/053_add_user_settings.sql`
- SQL Script: `APPLY_USER_SETTINGS_MIGRATION.sql`
- Model: `internal/models/user.go`
- Settings Model: `internal/models/user_settings.go`
- Repository: `internal/modules/user/repo/repository.go`

---

**Status:** Ready to apply ✅  
**Risk:** Low (non-breaking, backward compatible)  
**Required:** YES (blocks language switching feature)
