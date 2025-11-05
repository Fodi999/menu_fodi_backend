# 🚀 Recipe Feed API - Quick Reference

Quick examples for the most common use cases.

---

## 📋 Test Credentials

```javascript
const TEST_USERS = {
  dima: {
    id: "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
    email: "dima@example.com",
    password: "password123"
  },
  anna: {
    id: "fba50be3-e3c5-4d73-8ed8-cfb6422f7034",
    email: "anna@example.com",
    password: "password123"
  },
  maksym: {
    id: "407582be-59d5-4d21-873b-1a72d31b0d42",
    email: "fodi85@gmail.ru",
    password: "password123"
  }
};
```

---

## ⚡ Quick Examples

### 1. Main Feed (Instagram-style)
```bash
curl http://localhost:8080/api/posts
```

### 2. User Profile
```bash
curl http://localhost:8080/api/users/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/posts
```

### 3. Create Recipe
```bash
curl -X POST http://localhost:8080/api/recipes \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Tempura Roll",
    "description": "Crispy shrimp tempura with avocado",
    "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
    "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
  }'
```

---

## 🎨 React Component Example

```tsx
import { useState, useEffect } from 'react';

interface Recipe {
  id: string;
  title: string;
  description: string;
  imageUrl: string;
  author: {
    name: string;
    avatarUrl: string;
  };
  createdAt: string;
}

export default function RecipeFeed() {
  const [recipes, setRecipes] = useState<Recipe[]>([]);

  useEffect(() => {
    fetch('http://localhost:8080/api/posts')
      .then(res => res.json())
      .then(({ data }) => setRecipes(data));
  }, []);

  return (
    <div className="feed">
      {recipes.map(recipe => (
        <div key={recipe.id} className="recipe-card">
          <img src={recipe.imageUrl} alt={recipe.title} />
          <h3>{recipe.title}</h3>
          <p>{recipe.description}</p>
          <div className="author">
            <img src={recipe.author.avatarUrl} />
            <span>{recipe.author.name}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
```

---

## 📱 Frontend Flow

### Step 1: Upload Image
```typescript
const uploadToCloudinary = async (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('upload_preset', 'ml_default');

  const res = await fetch(
    'https://api.cloudinary.com/v1_1/dwrn0ohbp/image/upload',
    { method: 'POST', body: formData }
  );
  
  const { secure_url } = await res.json();
  return secure_url;
};
```

### Step 2: Create Recipe
```typescript
const createRecipe = async (file: File, title: string, description: string) => {
  // 1. Upload image
  const imageUrl = await uploadToCloudinary(file);
  
  // 2. Create recipe
  const res = await fetch('http://localhost:8080/api/recipes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title,
      description,
      imageUrl,
      authorId: currentUser.id
    })
  });
  
  const { data } = await res.json();
  return data;
};
```

---

## 🧪 Test Scenarios

### Scenario 1: Main Feed
```bash
# Get all recipes
curl http://localhost:8080/api/posts | jq '.data | length'
# Expected: 6 (or more)
```

### Scenario 2: User Has 2 Recipes
```bash
# Get Dima's recipes
curl http://localhost:8080/api/users/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/posts \
  | jq '.data | length'
# Expected: 2
```

### Scenario 3: Create & Verify
```bash
# 1. Create recipe
NEW_ID=$(curl -s -X POST http://localhost:8080/api/recipes \
  -H 'Content-Type: application/json' \
  -d '{"title":"Test","authorId":"ef03cd81-71fd-429f-bb5f-8be5c9172ca8"}' \
  | jq -r '.data.id')

# 2. Verify it appears in feed
curl -s http://localhost:8080/api/posts | jq ".data[] | select(.id == \"$NEW_ID\")"
```

---

## 📊 Database Queries

### Check Recipe Count
```sql
SELECT COUNT(*) FROM "Recipe";
```

### Get Recipes with Authors
```sql
SELECT 
  r.id,
  r.title,
  u.name as author_name,
  r.created_at
FROM "Recipe" r
JOIN "User" u ON r.author_id = u.id
ORDER BY r.created_at DESC;
```

### User Recipe Statistics
```sql
SELECT 
  u.name,
  COUNT(r.id) as recipe_count
FROM "User" u
LEFT JOIN "Recipe" r ON u.id = r.author_id
GROUP BY u.id, u.name
ORDER BY recipe_count DESC;
```

---

## 🐛 Troubleshooting

### Issue: "Author not found"
```bash
# Check if user exists
curl http://localhost:8080/api/users/YOUR_USER_ID/profile
```

### Issue: Empty feed
```bash
# Verify database has recipes
psql $DATABASE_URL -c "SELECT COUNT(*) FROM \"Recipe\";"
```

### Issue: Missing author details
```go
// Ensure Preload is used in handlers
db.Preload("Author").Find(&recipes)
```

---

## 🔗 Routes Summary

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/posts` | Main feed (all recipes) |
| GET | `/api/users/{id}/posts` | User profile recipes |
| POST | `/api/recipes` | Create recipe |
| PUT | `/api/recipes/{id}` | Update recipe |
| DELETE | `/api/recipes/{id}` | Delete recipe |

---

## 📚 See Full Documentation
- **RECIPE_FEED_API.md** - Complete API documentation
- **QUICK_START.md** - General API quick start
- **ALL_ENDPOINTS.md** - All 83+ API endpoints
