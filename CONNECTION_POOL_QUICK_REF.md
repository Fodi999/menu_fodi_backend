# Connection Pool Quick Reference

## ⚡ Quick Fix Summary

**Problem:** `FATAL: prepared statement name is already in use (SQLSTATE 08P01)`

**Solution:**
```go
// internal/database/db.go

DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
    PrepareStmt: false, // ← FIX #1: Disable prepared statement caching
    ...
})

sqlDB, err := DB.DB()
sqlDB.SetMaxOpenConns(15)              // ← FIX #2: Limit connections
sqlDB.SetMaxIdleConns(5)               // ← FIX #3: Keep 5 idle
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // ← FIX #4: Rotate every 5min
sqlDB.SetConnMaxIdleTime(2 * time.Minute)  // ← FIX #5: Close idle after 2min
```

## 🎯 Values for Different Databases

| Database | MaxOpen | MaxIdle | MaxLifetime | MaxIdleTime |
|----------|---------|---------|-------------|-------------|
| **Neon Serverless** | 15 | 5 | 5m | 2m |
| Supabase | 12 | 4 | 10m | 3m |
| AWS RDS | 25 | 10 | 15m | 5m |
| Local PostgreSQL | 50 | 20 | 30m | 10m |

## 📊 Current Configuration (Koyeb + Neon)

```
✅ MaxOpenConns:     15
✅ MaxIdleConns:     5
✅ ConnMaxLifetime:  5 minutes
✅ ConnMaxIdleTime:  2 minutes
✅ PrepareStmt:      false (disabled)
```

## 🔍 Monitoring

```bash
# Check for errors
koyeb logs menu-fodi-backend | grep "SQLSTATE 08P01"

# Verify pool config
koyeb logs menu-fodi-backend | grep "connection pool configured"
```

## 🚀 Deployment Status

- **Commit:** `3980b44` ✅
- **Status:** Deployed to production
- **Date:** 2026-01-16 20:59 UTC

## 📖 Full Documentation

See: `docs/DATABASE_CONNECTION_POOL_FIX.md`
