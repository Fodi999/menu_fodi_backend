# 🔧 Koyeb JWT Authentication Fix - Checklist

## 🔴 Problem Summary

- **Issue**: HTTP 401 Unauthorized on all protected endpoints (`/api/user/*`)
- **Root Cause**: `JWT_SECRET` environment variable not set on Koyeb
- **Impact**: Login works, but token validation fails immediately after

## 📊 Diagnostic Evidence

### Test Results (22 Dec 2025, 01:02 UTC)

```bash
# Login works ✅
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"dima@example.com","password":"password123"}' \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login"

# Response: 200 OK with token

# Same token immediately rejected ❌
curl -H "Authorization: Bearer <TOKEN_FROM_LOGIN>" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/recipes/saved"

# Response: HTTP 401 Unauthorized
```

**Conclusion**: Backend creates token, then rejects it = **JWT_SECRET mismatch**

## ✅ Solution Steps

### Step 1: Set JWT_SECRET on Koyeb

1. Go to **Koyeb Dashboard**: https://app.koyeb.com/
2. Navigate to your service: `yeasty-madelaine-fodi999-671ccdf5`
3. Click **Settings** → **Environment Variables**
4. Add new variable:
   - **Name**: `JWT_SECRET`
   - **Value**: `8cf357ecef730a2fc1848ea523f058fb36d372a4ce549ff7e4b68b6c7338b751`
   - ⚠️ **Important**: NO quotes, just raw value

5. Click **Update Service** or **Redeploy**

### Step 2: Wait for Deployment

- Koyeb will auto-deploy from `main` branch
- Commit `73c829a` includes diagnostic logging
- Deployment takes ~2-3 minutes

### Step 3: Check Logs for Diagnostic Output

After deployment, check Koyeb logs for:

```
✅ JWT_SECRET loaded from environment (length: 64)
```

If you see:
```
⚠️  JWT_SECRET not set, using fallback secret (INSECURE!)
```

Then the environment variable was not set correctly.

### Step 4: Test Authentication

```bash
# 1. Get fresh token AFTER redeploy
curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"dima@example.com","password":"password123"}' \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  | jq -r '.data.token'

# 2. Save token to variable
export TOKEN="<paste_token_here>"

# 3. Test protected endpoint
curl -i \
  -H "Authorization: Bearer $TOKEN" \
  "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/recipes/saved"

# Expected: HTTP 200 OK
```

### Step 5: Check Diagnostic Logs

Look for these log entries in Koyeb:

**Good signs:**
```
🔐 AuthMiddleware: GET /api/user/recipes/saved
📋 Auth header present: true, length: 259
🎫 Token extracted, length: 252
🔑 ValidateToken: Using JWT_SECRET from env (len=64)
✅ Auth OK for user ef03cd81-71fd-429f-bb5f-8be5c9172ca8
```

**Bad signs (JWT_SECRET not set):**
```
⚠️  ValidateToken: Using fallback secret
❌ JWT parse error: signature is invalid
❌ JWT validation failed: signature is invalid
```

## 🔐 Environment Variables Checklist

Ensure these are set on Koyeb:

| Variable | Required | Value | Status |
|----------|----------|-------|--------|
| `JWT_SECRET` | ✅ Yes | `8cf357ecef730a2fc1848ea523f058fb36d372a4ce549ff7e4b68b6c7338b751` | ⚠️ TO SET |
| `DATABASE_URL` | ✅ Yes | `postgresql://neondb_owner:...` | ✅ Set |
| `GROQ_API_KEY` | ✅ Yes | `gsk_UMES3X4kwogZZn0kr1dqWGdyb3FYh...` | ✅ Set |
| `PORT` | No | `8080` | Optional |
| `CLOUDINARY_URL` | No | `cloudinary://...` | Optional |

## 🐛 Troubleshooting

### Still 401 After Setting JWT_SECRET?

1. **Check Koyeb logs** for JWT_SECRET length:
   ```
   ✅ JWT_SECRET loaded from environment (length: 64)
   ```

2. **Verify redeploy completed**:
   - Check deployment status in Koyeb dashboard
   - Deployment should show "Healthy" status

3. **Test with brand new token**:
   - Don't reuse old tokens from before JWT_SECRET was set
   - Get fresh token via `/api/auth/login`

4. **Check for typos in environment variable name**:
   - Must be exactly `JWT_SECRET` (case-sensitive)
   - No spaces, no quotes in the variable name

5. **Verify .env file updated locally**:
   ```bash
   grep "^JWT_SECRET=" .env
   # Should show: JWT_SECRET="8cf357ecef730a2fc1848ea523f058fb36d372a4ce549ff7e4b68b6c7338b751"
   ```

### Frontend Still Shows 401?

After backend fix, frontend users need to:
1. **Clear browser cache** and cookies
2. **Log out** and **log in again** to get fresh token
3. Old tokens created before JWT_SECRET change will be invalid

## 📝 Code Changes Made

### Commit `73c829a` - Diagnostic Logging

**Files Modified:**
- `internal/middleware/auth.go` - Improved Bearer token parsing + detailed logs
- `internal/modules/auth/service/jwt_service.go` - Log JWT_SECRET usage
- `.env` - Updated JWT_SECRET to strong 64-char hex string
- `test_koyeb_auth.sh` - Interactive test script

**Key Improvements:**
1. Proper `Bearer` token extraction (split instead of TrimPrefix)
2. Log JWT_SECRET presence/length at startup
3. Log every step of token validation
4. Detailed error messages for debugging

## ✅ Success Criteria

Authentication is fixed when:

- [ ] Koyeb logs show: `✅ JWT_SECRET loaded from environment (length: 64)`
- [ ] Login returns 200 with token
- [ ] Protected endpoints return 200 with same token
- [ ] No `⚠️  Using fallback secret` warnings in logs
- [ ] Frontend can access `/api/user/recipes/saved` without 401

## 🚀 Next Steps After Fix

1. **Update frontend to handle auth properly**
2. **Test all protected endpoints**: `/api/user/*`, `/api/recipes/*`
3. **Remove diagnostic logs** in production (or reduce verbosity)
4. **Add monitoring** for JWT validation errors
5. **Document JWT_SECRET rotation** procedure

## 📞 Support

If issues persist after following this checklist:
1. Check Koyeb logs for detailed error messages
2. Run `./test_koyeb_auth.sh` script for automated diagnosis
3. Verify all environment variables are set correctly
4. Test with curl commands to isolate frontend vs backend issues

---

**Last Updated**: 22 Dec 2025, 01:15 UTC  
**Commit**: `73c829a`  
**Status**: ⚠️ Awaiting JWT_SECRET configuration on Koyeb
