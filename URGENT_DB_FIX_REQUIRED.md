# 🔥 URGENT FIX: Menu Unique Constraint Still Broken

## Problem Status
- ✅ Code fixed (migrations created)
- ✅ Code pushed to repository
- ❌ **Database NOT updated** - still has old constraint with `status`

## Why It's Still Failing

Production database still has this constraint:
```sql
UNIQUE (user_id, recipe_id, planned_for, status)  ❌ WRONG
```

When you call `/complete`:
1. UPDATE tries to change status from `cooking` → `completed`
2. Postgres checks the UNIQUE constraint
3. **ERROR**: Now we have TWO rows for same (user_id, recipe_id, planned_for):
   - One with status = `cooking`
   - One with status = `completed` (trying to insert)
4. BOOM! Constraint violated!

## Solution

**MUST RUN MANUALLY** - Need direct DB access to:
1. Drop old table
2. Recreate with correct constraint

### Option A: If you have psql access

```bash
psql $DATABASE_URL < SQL_FIX_PRODUCTION_HOTFIX.sql
```

### Option B: Via database GUI

1. Connect to production database
2. Run SQL from `SQL_FIX_PRODUCTION_HOTFIX.sql`
3. Verify with: `\d user_menu_items`

### Option C: Automated (if CI/CD supports migrations)

```bash
# Ensure migration is applied before deployment
make migrate
```

## Files Created

1. **SQL_FIX_PRODUCTION_HOTFIX.sql** - Direct SQL fix (run immediately!)
2. **migrations/20260122_recreate_user_menu_items_fixed.sql** - Migration file
3. **MENU_UNIQUE_CONSTRAINT_FIX.md** - Full documentation

## Verification

After running the SQL fix:

```bash
# Check constraint
psql $DATABASE_URL -c "\d user_menu_items"

# Should show:
# Indexes:
#     "user_menu_items_pkey" PRIMARY KEY, btree (id)
#     "unique_user_recipe_today" UNIQUE CONSTRAINT, btree (user_id, recipe_id, planned_for)
```

## Expected Result After Fix

```
✅ GET /api/menu/today → 200 OK
✅ POST /api/menu/today → 201 CREATED
✅ POST /api/menu/{id}/start → 200 OK (status: cooking)
✅ POST /api/menu/{id}/complete → 200 OK (status: completed) 🎉
```

## Timeline

- ✅ 2026-01-22 20:14: Code changes pushed
- ❌ 2026-01-22 20:19: Still failing (DB not updated)
- **NOW**: Run SQL fix immediately!

---

**NEXT STEP:** Run `SQL_FIX_PRODUCTION_HOTFIX.sql` against production database!
