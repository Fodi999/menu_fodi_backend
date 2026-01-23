# 🔴 URGENT: CORS Fix Required

## Problem
Frontend (`https://dima-fomin.pl`) cannot access backend API due to CORS policy:

```
Access to fetch at 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories' 
from origin 'https://dima-fomin.pl' has been blocked by CORS policy
```

## Root Cause
Domain `https://dima-fomin.pl` is NOT in `ALLOWED_ORIGINS` environment variable.

## Solution

### 1. Update Koyeb Environment Variable

Go to Koyeb Dashboard → Your App → Settings → Environment Variables

**Update:**
```
ALLOWED_ORIGINS
```

**From:**
```
http://localhost:3000,http://localhost:3001,https://menu-fodifood.vercel.app
```

**To:**
```
http://localhost:3000,http://localhost:3001,https://menu-fodifood.vercel.app,https://dima-fomin.pl
```

### 2. Redeploy Backend

After updating environment variable, redeploy the service in Koyeb.

### 3. Verify Fix

After redeployment, check CORS headers:

```bash
curl -I -X OPTIONS \
  -H "Origin: https://dima-fomin.pl" \
  -H "Access-Control-Request-Method: GET" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories
```

Should return:
```
Access-Control-Allow-Origin: https://dima-fomin.pl
```

---

## Additional Issue: Frontend Localhost Connection

Frontend is trying to connect to `localhost:8080` instead of production API.

### Check Vercel Environment Variables

Ensure these are set in Vercel dashboard:

```env
NEXT_PUBLIC_SITE_URL=https://dima-fomin.pl
NEXT_PUBLIC_API_BASE=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
```

**Remove any localhost references!**

---

## Quick Fix Steps

1. ✅ Go to Koyeb → Environment Variables
2. ✅ Add `https://dima-fomin.pl` to ALLOWED_ORIGINS
3. ✅ Redeploy backend
4. ✅ Check Vercel environment variables (no localhost!)
5. ✅ Redeploy frontend if needed
6. ✅ Test: `https://dima-fomin.pl` should work without CORS errors

---

**Status:** Waiting for Koyeb environment variable update
**Priority:** 🔴 HIGH (blocking production frontend)
