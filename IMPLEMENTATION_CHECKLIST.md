# ✅ Implementation Checklist - Backend as Source of Truth

Use this checklist to track implementation progress.

---

## Phase 1: Foundation (Day 1) ✅

- [x] Create `internal/models/response.go` - Unified response helpers
- [x] Create `internal/models/errors.go` - Error code constants  
- [x] Create `internal/middleware/request_id.go` - Request ID tracking
- [x] Test compilation: `go build ./cmd/server` ✅
- [ ] **TODO:** Add RequestIDMiddleware to `cmd/server/main.go`
- [ ] **TODO:** Test middleware works (curl with/without X-Request-ID)
- [ ] **TODO:** Verify logs include request_id

---

## Phase 2: High-Priority Modules (Days 2-3)

### Auth Module 🔥 **HIGHEST PRIORITY**
**File:** `internal/modules/auth/transport/http/handler.go`

- [ ] Migrate `Register()` handler
- [ ] Migrate `Login()` handler  
- [ ] Migrate `VerifyToken()` handler
- [ ] Migrate `GetCurrentUser()` handler
- [ ] Test auth flow end-to-end
- [ ] Update auth tests

**Why First:** Every authenticated request depends on this.

---

### Admin Ingredients 🔥 **HIGH PRIORITY**
**File:** `internal/modules/admin/transport/http/ingredients.go`

- [ ] Migrate `GetIngredients()` (list with pagination)
- [ ] Migrate `GetIngredientByID()` (single item)
- [ ] Migrate `SuggestIngredients()` (autocomplete) ⭐ Used by frontend
- [ ] Migrate `CreateIngredient()` (with AI classification)
- [ ] Migrate `GetIngredientStats()` (statistics)
- [ ] Test ingredients flow end-to-end
- [ ] Update ingredient tests

**Why Second:** Frontend admin panel depends on this (especially `/suggest`).

---

### Admin Recipes 🔥 **HIGH PRIORITY**
**File:** `internal/modules/admin/transport/http/recipes.go`

- [ ] Migrate `GetRecipes()` (list)
- [ ] Migrate `GetRecipeStats()` (statistics)
- [ ] Migrate `PreviewAI()` (AI recipe preview) ⭐ Recently fixed
- [ ] Migrate `CreateAI()` (AI recipe creation + save) ⭐ Recently fixed
- [ ] Test AI recipe flow end-to-end
- [ ] Update recipe tests

**Why Third:** AI recipe feature is critical and was just fixed.

---

## Phase 3: User-Facing Modules (Days 4-5)

### Fridge Module
**File:** `internal/modules/fridge/transport/http/handlers.go`

- [ ] Migrate `GetUserItems()` (list user's fridge)
- [ ] Migrate `AddItem()` (add item to fridge)
- [ ] Migrate `UpdateItemQuantity()` (update quantity)
- [ ] Migrate `DeleteItem()` (remove from fridge)
- [ ] Migrate `AddPrice()` (event sourcing for prices)
- [ ] Migrate `GetPriceHistory()` (price history)
- [ ] Migrate `AddMissingIngredients()` (from recipe)
- [ ] Test fridge flow end-to-end

---

### Recipe Catalog
**File:** `internal/modules/recipes/transport/http/handler.go`

- [ ] Migrate `ListRecipes()` (public catalog)
- [ ] Migrate `GetRecipeByID()` (recipe details)
- [ ] Migrate `GetRecipeStats()` (statistics)
- [ ] Migrate `MatchRecipes()` (fridge matching)
- [ ] Migrate `GetAvailableRecipes()` (categorized by feasibility)
- [ ] Migrate `GetRecommendation()` (best recipe for user)
- [ ] Migrate `SaveRecipe()` (save to user collection)
- [ ] Migrate `GetSavedRecipes()` (user's saved recipes)
- [ ] Migrate `CookRecipe()` (deduct from fridge)
- [ ] Migrate `AdaptRecipe()` (AI adaptation)
- [ ] Test recipe flow end-to-end

---

### User Module
**File:** `internal/modules/user/transport/http/handler.go`

- [ ] Migrate `GetProfile()` (get user profile)
- [ ] Migrate `UpdateProfile()` (update user profile)
- [ ] Test user flow end-to-end

---

## Phase 4: Supporting Modules (Day 6)

### Admin Users
**File:** `internal/modules/admin/transport/http/users.go`

- [ ] Migrate `GetUsers()` (list all users)
- [ ] Migrate `GetUserStats()` (user statistics)
- [ ] Migrate `UpdateUser()` (update user)
- [ ] Migrate `UpdateRole()` (change user role - super_admin only)
- [ ] Migrate `DeleteUser()` (delete user - super_admin only)

---

### Token Economy
**Files:** `internal/modules/admin/transport/http/token_bank.go`, `treasury.go`

- [ ] Migrate `GetTokenBanks()` (all user banks)
- [ ] Migrate `GetTokenBank()` (specific user)
- [ ] Migrate `AllocateTokens()` (give tokens to user)
- [ ] Migrate `RevokeTokens()` (remove tokens from user)
- [ ] Migrate `GetTreasury()` (treasury info - admin)
- [ ] Migrate `GetPublicTreasury()` (treasury info - public)
- [ ] Migrate `AllocateTreasury()` (allocate from treasury)

---

### Marketplace
**File:** `internal/modules/marketplace/transport/http/handlers.go`

- [ ] Migrate `GetMarketRecipes()` (marketplace catalog)
- [ ] Migrate `GetLeaderboard()` (sellers leaderboard)
- [ ] Migrate `GetSellerStats()` (seller statistics)
- [ ] Migrate `PurchaseRecipe()` (buy recipe)
- [ ] Migrate `GetUserPurchases()` (user's purchases)
- [ ] Migrate `UploadImage()` (image upload)

---

### Academy
**File:** `internal/modules/academy/transport/http/handlers.go`

- [ ] Migrate `GetCourses()` (list courses)
- [ ] Migrate `EnrollCourse()` (enroll in course)

---

## Phase 5: Testing & Documentation (Day 7)

### Update Tests
- [ ] Update auth module tests
- [ ] Update ingredient module tests
- [ ] Update recipe module tests
- [ ] Update fridge module tests
- [ ] Update integration tests
- [ ] Run all tests: `go test ./...`

---

### Update Documentation
- [ ] Update `API_CONTRACT_COMPLETE.md` with examples ✅ (already done)
- [ ] Create migration guide for team ✅ (already done)
- [ ] Update README.md with new response format
- [ ] Create TypeScript SDK documentation ✅ (already done)

---

### Frontend Integration
- [ ] Share TypeScript types with frontend team
- [ ] Update frontend API client to parse new format
- [ ] Add error code handling in frontend
- [ ] Test error scenarios (401 → redirect to login, etc.)
- [ ] Add request_id to Sentry error reports

---

## Phase 6: Deployment (Day 8)

### Local Testing
- [ ] Test all migrated endpoints locally
- [ ] Test with frontend locally
- [ ] Check logs for request_id correlation
- [ ] Verify error responses are correct

---

### Production Deployment
- [ ] Deploy to Koyeb (auto-deploy from GitHub)
- [ ] Monitor Koyeb logs for errors
- [ ] Test production endpoints with curl
- [ ] Verify frontend works with production backend
- [ ] Monitor error rates in Sentry

---

### Post-Deployment
- [ ] Update API documentation links
- [ ] Announce new response format to team
- [ ] Archive old documentation
- [ ] Create changelog entry

---

## 🎯 Success Criteria

- [x] All new code files created ✅
- [x] Code compiles successfully ✅
- [ ] Request ID middleware active
- [ ] All high-priority handlers migrated (auth, ingredients, recipes)
- [ ] All user-facing handlers migrated (fridge, recipe catalog)
- [ ] All tests passing
- [ ] Frontend integrated and working
- [ ] Production deployment successful
- [ ] Zero breaking changes (backward compatibility maintained)

---

## 📊 Progress Tracking

**Overall Progress:** 20% (Foundation complete, migration pending)

### By Module:
- Foundation: ✅ 100% (response models, errors, middleware)
- Auth: ⏳ 0% (not started)
- Admin Ingredients: ⏳ 0% (not started)
- Admin Recipes: ⏳ 0% (not started)
- Fridge: ⏳ 0% (not started)
- Recipe Catalog: ⏳ 0% (not started)
- User: ⏳ 0% (not started)
- Token Economy: ⏳ 0% (not started)
- Marketplace: ⏳ 0% (not started)
- Academy: ⏳ 0% (not started)

---

## 🚀 Quick Start

**First Action:**
```bash
# Add RequestIDMiddleware to cmd/server/main.go
# See: QUICKSTART_REQUEST_ID.md
```

**Then pick ONE handler:**
```bash
# Easiest to start: Admin Ingredients - GetIngredientByID
# See: docs/MIGRATION_EXAMPLES.md - Example 2
```

---

## 📚 Resources

- `docs/BACKEND_AS_SOURCE_OF_TRUTH.md` - Philosophy & principles
- `docs/MIGRATION_EXAMPLES.md` - 6 detailed migration examples
- `API_CONTRACT_COMPLETE.md` - Updated API contracts
- `docs/API_TYPES_TYPESCRIPT.ts` - TypeScript types for frontend
- `QUICKSTART_REQUEST_ID.md` - 5-minute setup guide

---

**Last Updated:** 2026-01-11  
**Status:** ✅ Ready to start  
**Estimated Total Time:** 7-8 days  
**Priority:** 🔥 CRITICAL - Foundation for all future development
