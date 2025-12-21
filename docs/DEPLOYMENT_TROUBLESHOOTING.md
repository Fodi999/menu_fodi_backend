# Deployment Troubleshooting 🚀

## Issue: Backend Returns 500 Error

### Symptoms
```
[AI][ERROR] Raw response: 
[AI][ERROR] Is JSON: false, Parse error: empty response
panic: runtime error: invalid memory address or nil pointer dereference
-> handlers.go:746
```

### Root Causes

#### 1. **Deployment Lag** ⏳
**Problem:** Koyeb auto-deploy takes 2-3 minutes  
**Solution:** Wait for deployment to complete

Check deployment status:
```bash
# View recent commits
git log --oneline -5

# Check if Koyeb is still deploying
# Visit: https://app.koyeb.com
# Look for: "Deploying" → "Healthy"
```

Expected timeline:
- Push to main: 0 seconds
- Koyeb detects: 10-30 seconds
- Build starts: 30-60 seconds
- Deploy starts: 1-2 minutes
- **Service ready: 2-3 minutes** ✅

#### 2. **Empty AI Response** 🤖
**Problem:** Groq API returns empty string  
**Causes:**
- Rate limit exceeded
- API timeout
- Network hiccup
- Model unavailable

**Logs:**
```
[AI][ERROR] Raw response: 
[AI][ERROR] Is JSON: false, Parse error: empty response
[AI][RETRY] First attempt failed, trying self-repair...
[AI][RETRY] Self-repair also failed
```

**Solution:** User retries (self-repair can't fix empty response)

#### 3. **Old Code Still Running** 🔄
**Problem:** Koyeb hasn't picked up latest commit  
**How to verify:**

Check commit hash in logs:
```bash
# Latest local commit
git rev-parse --short HEAD

# vs

# Commit in Koyeb logs
# Look for: "Deployed commit: abc1234"
```

If mismatch → wait for auto-deploy or trigger manual deploy

#### 4. **Nil Pointer Not Fixed** ⚠️
**Problem:** Code fix not deployed yet  
**Evidence:**
```
panic: runtime error: invalid memory address or nil pointer dereference
-> handlers.go:746
```

**Fixed in:** commit `5cce607` (Dec 21, 13:xx)
**Contains:**
- Nil checks for response.Recipe
- Safe zap logging
- Token limit 1024 → 4096

**Verify fix deployed:**
```bash
# Check if fix commit is in deployment
git log --oneline | grep "5cce607"

# If yes → wait for Koyeb to deploy
# If no → push again
```

---

## Quick Diagnostic Checklist

### Step 1: Check Deployment Status
- [ ] Visit Koyeb dashboard
- [ ] Look for "Deploying" or "Healthy"
- [ ] If deploying → wait 2-3 minutes
- [ ] If healthy → check commit hash

### Step 2: Verify Commit Deployed
```bash
# Latest 3 commits
git log --oneline -3

# Should see:
# ebc2769 feat(migrations): Add 'Olej roślinny'
# eda23da docs: Add recipe recalculation API
# 688a7db feat(ai): Add recipe recalculation endpoint
```

### Step 3: Check Logs Pattern

**OLD deployment (broken):**
```
panic: runtime error: invalid memory address or nil pointer dereference
-> handlers.go:746
```

**NEW deployment (fixed):**
```
[AI][RETRY] First attempt failed, trying self-repair...
[AI][RETRY] Self-repair succeeded!
or
[AI][RETRY] Self-repair also failed
→ returns error to user (no panic)
```

### Step 4: Test Recipe Generation
```bash
# Wait 3 minutes after push
# Try recipe generation on frontend
# Check new logs (should be after 13:59:36)
```

---

## Timeline Reference

**Current timestamp:** 2025-12-21 13:59:36 (Instance stopped)

**Fixes deployed:**
- 13:48:xx - `5cce607` Nil pointer fix + token limit
- 13:52:xx - `caf900e` Self-repair retry
- 14:02:xx - `688a7db` Recalculation endpoint
- 14:05:xx - `ebc2769` Olej roślinny migration

**Expected ready:** 14:08:00 (3 min after last push)

---

## Expected Behavior After Fix

### Scenario 1: AI Returns Valid JSON
```
[AI][SUCCESS] Recipe parsed successfully: Recipe name
[AI][ECONOMY] ✅ AFTER override - recipe.Economy = {UsedValue:10.43 ...}
→ 200 OK with recipe
```

### Scenario 2: AI Returns Truncated JSON
```
[AI][ERROR] Failed to parse AI response as JSON
[AI][RETRY] First attempt failed, trying self-repair...
[AI][RETRY] ✅ Self-repair succeeded!
[AI][SUCCESS] Recipe parsed successfully: Recipe name
→ 200 OK with recipe
```

### Scenario 3: AI Returns Empty (Rate Limit)
```
[AI][ERROR] Raw response: 
[AI][RETRY] First attempt failed, trying self-repair...
[AI][RETRY] Self-repair also failed
→ 200 OK with {"success": false, "message": "Try again"}
→ NO PANIC, clean error handling
```

### Scenario 4: AI Second Attempt Succeeds
```
[AI][ERROR] Failed to parse AI response as JSON
[AI][RETRY] Trying self-repair...
[AI][RETRY] ✅ Self-repair succeeded!
→ 200 OK with recipe
```

---

## Common Mistakes

### ❌ DON'T: Test immediately after push
```bash
git push origin main
# Wait 0 seconds
curl /api/ai/create-recipe  # ← OLD CODE STILL RUNNING
```

### ✅ DO: Wait for deployment
```bash
git push origin main
# Wait 2-3 minutes
# Check Koyeb dashboard
curl /api/ai/create-recipe  # ← NEW CODE RUNNING
```

### ❌ DON'T: Assume instant deployment
- Koyeb auto-deploy is NOT instant
- Need build time (Go compilation)
- Need container restart

### ✅ DO: Verify deployment completed
1. Check Koyeb status: "Healthy"
2. Check logs timestamp > push time
3. Try request

---

## Koyeb Auto-Deploy Process

```
1. GitHub detects push to main
   ↓ (10-30 seconds)
2. Webhook notifies Koyeb
   ↓ (5-10 seconds)
3. Koyeb pulls new code
   ↓ (10-20 seconds)
4. Koyeb runs: go build ./cmd/server
   ↓ (30-60 seconds)
5. Koyeb builds Docker image
   ↓ (20-40 seconds)
6. Koyeb stops old container
   ↓ (5-10 seconds)
7. Koyeb starts new container
   ↓ (10-20 seconds)
8. Health check passes
   ↓
9. ✅ NEW CODE LIVE
```

**Total:** ~2-3 minutes

---

## Debug Commands

### Check latest commits
```bash
git log --oneline -5
```

### Check if specific fix deployed
```bash
git log --oneline | grep "nil pointer"
git log --oneline | grep "self-repair"
```

### View deployment timeline
```bash
git log --pretty=format:"%h %ai %s" -5
```

### Check current branch
```bash
git branch
# Should show: * main
```

---

## Next Steps

1. ⏳ **Wait 2-3 minutes** after last push (14:05:xx + 3min = 14:08:xx)
2. 🔍 **Check Koyeb dashboard** - status should be "Healthy"
3. 🧪 **Try recipe generation** on frontend
4. 📊 **Check NEW logs** (timestamp > 14:08:00)
5. ✅ **Verify fix:**
   - No panic
   - Self-repair logs visible
   - Recipe generated OR clean error message

---

**Related docs:**
- [AI_SELF_REPAIR_PATTERN.md](./AI_SELF_REPAIR_PATTERN.md)
- [PRICE_FLOW_DEBUG.md](./PRICE_FLOW_DEBUG.md)
