# 🎉 Recipe Feed API - Production Deployment

**Status:** ✅ DEPLOYED  
**Date:** November 5, 2025  
**Commit:** `2831016`

---

## 🌐 Production Endpoints

### Base URL
```
https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
```

### Available Endpoints

#### 1. Main Feed (All Recipes)
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/posts
```

#### 2. User Profile Recipes
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/users/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/posts
```

#### 3. Create Recipe
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "New Recipe",
    "description": "Description here",
    "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
    "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
  }'
```

#### 4. Update Recipe
```bash
curl -X PUT https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/{recipe-id} \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Updated Title",
    "description": "Updated description"
  }'
```

#### 5. Delete Recipe
```bash
curl -X DELETE https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/{recipe-id}
```

---

## 🧪 Production Test Results

### Health Check
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/health
```
```json
{
  "status": "ok",
  "data": {
    "database": "connected",
    "service": "menu-fodifood-backend"
  }
}
```

### Recipe Count
```bash
curl -s https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/posts | jq '.data | length'
# Result: 6 recipes
```

### Sample Recipe
```bash
curl -s https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/posts | jq '.data[0]'
```
```json
{
  "id": "febf704a-604c-4701-8f92-9c4c4648d0db",
  "title": "Premium California Roll",
  "description": "Luksusowa rolada kalifornijska z krabem królewskim i awokado",
  "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
  "authorId": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "author": {
    "id": "407582be-59d5-4d21-873b-1a72d31b0d42",
    "email": "fodi85@gmail.ru",
    "name": "Dima Fomin",
    "role": "user",
    "createdAt": "2025-11-04T09:50:11.472Z"
  },
  "createdAt": "2025-11-05T11:39:48.232909+01:00",
  "updatedAt": "2025-11-05T11:40:15.123456+01:00"
}
```

---

## 🎨 Frontend Integration

### Environment Variable
Add to your `.env.local`:
```bash
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
```

### React/Next.js Example
```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL;

// Fetch main feed
const fetchFeed = async () => {
  const response = await fetch(`${API_URL}/api/posts`);
  const { data } = await response.json();
  return data;
};

// Fetch user recipes
const fetchUserRecipes = async (userId: string) => {
  const response = await fetch(`${API_URL}/api/users/${userId}/posts`);
  const { data } = await response.json();
  return data;
};

// Create recipe
const createRecipe = async (recipe: RecipeInput) => {
  const response = await fetch(`${API_URL}/api/recipes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(recipe)
  });
  const { data } = await response.json();
  return data;
};
```

---

## 📊 Production Database

### Connection
- **Platform:** Neon PostgreSQL
- **Status:** Connected ✅
- **Tables:** Recipe, User (with foreign key relationship)

### Sample Data
- 5 pre-seeded recipes from 3 different users
- 1 test recipe created via API
- Total: 6 recipes in production

### Test Users
```json
{
  "dima": {
    "id": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
    "email": "dima@example.com",
    "password": "password123"
  },
  "anna": {
    "id": "fba50be3-e3c5-4d73-8ed8-cfb6422f7034",
    "email": "anna@example.com",
    "password": "password123"
  },
  "maksym": {
    "id": "407582be-59d5-4d21-873b-1a72d31b0d42",
    "email": "fodi85@gmail.ru",
    "password": "password123"
  }
}
```

---

## 📚 Documentation

- **[RECIPE_FEED_API.md](RECIPE_FEED_API.md)** - Complete API documentation
- **[RECIPE_FEED_QUICK_START.md](RECIPE_FEED_QUICK_START.md)** - Quick reference guide
- **[QUICK_START.md](QUICK_START.md)** - General API quick start
- **[ALL_ENDPOINTS.md](ALL_ENDPOINTS.md)** - All 88 API endpoints (83 + 5 new)

---

## 🚀 Deployment Info

- **Platform:** Koyeb
- **Auto-deploy:** Enabled (GitHub main branch)
- **Build Time:** ~2 minutes
- **Health Check:** `/api/health`
- **CORS:** Configured for dima-fomin.pl

---

## ✅ Production Checklist

- [x] Code compiled without errors
- [x] Database migration applied
- [x] 5 routes added to server
- [x] All endpoints tested locally
- [x] Pushed to GitHub (commit 2831016)
- [x] Auto-deployed to Koyeb
- [x] Health check passed
- [x] Recipe feed working on production
- [x] Author details properly loaded
- [x] CORS configured
- [x] Documentation complete

---

**Last Updated:** November 5, 2025  
**Status:** 🟢 Production Ready  
**Next Steps:** Frontend integration with dima-fomin.pl
