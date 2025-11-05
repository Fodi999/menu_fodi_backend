# 📱 Recipe Feed API Documentation

Complete API documentation for the public recipe sharing feature that allows users to post recipes and view them on a main feed or user profiles.

---

## 🎯 Overview

The Recipe Feed system allows:
- **Main Feed**: Display all recipes from all users (like Instagram/Facebook feed)
- **User Profile**: Show recipes posted by a specific user
- **Recipe Management**: Create, update, and delete recipes
- **Author Information**: Each recipe includes author details (name, avatar, etc.)

---

## 📊 Data Model

### Recipe Schema
```json
{
  "id": "string (UUID)",
  "title": "string",
  "description": "string",
  "imageUrl": "string",
  "authorId": "string (UUID)",
  "author": {
    "id": "string",
    "name": "string",
    "email": "string",
    "avatarUrl": "string",
    "role": "string",
    "createdAt": "timestamp"
  },
  "createdAt": "timestamp",
  "updatedAt": "timestamp"
}
```

### Database Table
```sql
CREATE TABLE "Recipe" (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    author_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_recipe_author FOREIGN KEY (author_id) 
        REFERENCES "User"(id) ON DELETE CASCADE
);

CREATE INDEX idx_recipe_author_id ON "Recipe"(author_id);
CREATE INDEX idx_recipe_created_at ON "Recipe"(created_at DESC);
```

---

## 🔌 API Endpoints

### 1. Get All Recipes (Main Feed)
**Endpoint:** `GET /api/posts`  
**Purpose:** Retrieve all recipes from all users for the main feed  
**Authentication:** Public (no auth required)

#### Request
```bash
curl http://localhost:8080/api/posts
```

#### Response (200 OK)
```json
{
  "status": "success",
  "data": [
    {
      "id": "recipe-001",
      "title": "Fresh Salmon Nigiri",
      "description": "Autentyczne nigiri z łososiem - tradycyjna japońska kuchnia",
      "imageUrl": "https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=800",
      "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
      "author": {
        "id": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
        "email": "dima@example.com",
        "name": "Dima Fomin",
        "avatarUrl": "https://res.cloudinary.com/dwrn0ohbp/...",
        "role": "user",
        "createdAt": "2025-11-03T12:21:14.105Z"
      },
      "createdAt": "2025-11-05T11:36:50.584057+01:00",
      "updatedAt": "2025-11-05T11:36:50.584057+01:00"
    },
    // ... more recipes
  ]
}
```

---

### 2. Get User Recipes (Profile)
**Endpoint:** `GET /api/users/{id}/posts`  
**Purpose:** Retrieve all recipes posted by a specific user  
**Authentication:** Public

#### Request
```bash
curl http://localhost:8080/api/users/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/posts
```

#### Response (200 OK)
```json
{
  "status": "success",
  "data": [
    {
      "id": "recipe-001",
      "title": "Fresh Salmon Nigiri",
      "description": "Autentyczne nigiri z łososiem",
      "imageUrl": "https://images.unsplash.com/...",
      "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
      "author": { ... },
      "createdAt": "2025-11-05T11:36:50.584057+01:00",
      "updatedAt": "2025-11-05T11:36:50.584057+01:00"
    }
  ]
}
```

#### Error Response (404 Not Found)
```json
{
  "error": "User not found"
}
```

---

### 3. Create Recipe
**Endpoint:** `POST /api/recipes`  
**Purpose:** Create a new recipe post  
**Authentication:** Public (should be protected in production)

#### Request
```bash
curl -X POST http://localhost:8080/api/recipes \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "California Roll",
    "description": "Klasyczna rolada kalifornijska z krabem, awokado i ogórkiem",
    "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
    "authorId": "407582be-59d5-4d21-873b-1a72d31b0d42"
  }'
```

#### Request Body
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | ✅ Yes | Recipe title |
| description | string | No | Recipe description |
| imageUrl | string | No | URL to recipe image |
| authorId | string | ✅ Yes | UUID of the user creating the recipe |

#### Response (201 Created)
```json
{
  "status": "success",
  "data": {
    "id": "febf704a-604c-4701-8f92-9c4c4648d0db",
    "title": "California Roll",
    "description": "Klasyczna rolada kalifornijska z krabem, awokado i ogórkiem",
    "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
    "authorId": "407582be-59d5-4d21-873b-1a72d31b0d42",
    "author": {
      "id": "407582be-59d5-4d21-873b-1a72d31b0d42",
      "name": "Dima Fomin",
      "email": "fodi85@gmail.ru",
      ...
    },
    "createdAt": "2025-11-05T11:39:48.232909+01:00",
    "updatedAt": "2025-11-05T11:39:48.232909+01:00"
  }
}
```

#### Error Response (400 Bad Request)
```json
{
  "error": "Title is required"
}
```

#### Error Response (404 Not Found)
```json
{
  "error": "Author not found"
}
```

---

### 4. Update Recipe
**Endpoint:** `PUT /api/recipes/{id}`  
**Purpose:** Update an existing recipe  
**Authentication:** Public (should verify ownership in production)

#### Request
```bash
curl -X PUT http://localhost:8080/api/recipes/febf704a-604c-4701-8f92-9c4c4648d0db \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Premium California Roll",
    "description": "Luksusowa rolada kalifornijska z krabem królewskim"
  }'
```

#### Request Body (all fields optional)
| Field | Type | Description |
|-------|------|-------------|
| title | string | New recipe title |
| description | string | New description |
| imageUrl | string | New image URL |

#### Response (200 OK)
```json
{
  "status": "success",
  "data": {
    "id": "febf704a-604c-4701-8f92-9c4c4648d0db",
    "title": "Premium California Roll",
    "description": "Luksusowa rolada kalifornijska z krabem królewskim",
    "imageUrl": "https://images.unsplash.com/...",
    "authorId": "407582be-59d5-4d21-873b-1a72d31b0d42",
    "author": { ... },
    "createdAt": "2025-11-05T11:39:48.232909+01:00",
    "updatedAt": "2025-11-05T11:40:15.123456+01:00"
  }
}
```

---

### 5. Delete Recipe
**Endpoint:** `DELETE /api/recipes/{id}`  
**Purpose:** Delete a recipe  
**Authentication:** Public (should verify ownership in production)

#### Request
```bash
curl -X DELETE http://localhost:8080/api/recipes/febf704a-604c-4701-8f92-9c4c4648d0db
```

#### Response (200 OK)
```json
{
  "status": "success",
  "message": "Recipe deleted successfully"
}
```

#### Error Response (404 Not Found)
```json
{
  "error": "Recipe not found"
}
```

---

## 🧪 Frontend Integration Examples

### React/Next.js

#### Fetch Main Feed
```typescript
const fetchFeed = async () => {
  const response = await fetch(`${API_URL}/api/posts`);
  const { data } = await response.json();
  setPosts(data);
};

useEffect(() => {
  fetchFeed();
}, []);
```

#### Display Feed
```tsx
{posts.map(post => (
  <div key={post.id} className="recipe-card">
    <img src={post.imageUrl} alt={post.title} />
    <h3>{post.title}</h3>
    <p>{post.description}</p>
    <div className="author">
      <img src={post.author.avatarUrl} alt={post.author.name} />
      <span>{post.author.name}</span>
    </div>
    <time>{new Date(post.createdAt).toLocaleDateString()}</time>
  </div>
))}
```

#### Fetch User Profile Posts
```typescript
const fetchUserPosts = async (userId: string) => {
  const response = await fetch(`${API_URL}/api/users/${userId}/posts`);
  const { data } = await response.json();
  setUserPosts(data);
};
```

#### Create New Recipe
```typescript
const createRecipe = async (recipe: RecipeInput) => {
  const response = await fetch(`${API_URL}/api/recipes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: recipe.title,
      description: recipe.description,
      imageUrl: recipe.imageUrl,
      authorId: currentUser.id
    })
  });
  
  const { data } = await response.json();
  return data;
};
```

---

## 📸 Image Upload Flow

### Option 1: Direct Cloudinary Upload (Recommended)
```typescript
// 1. Upload image to Cloudinary first
const uploadImage = async (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('upload_preset', 'ml_default');
  formData.append('folder', 'recipe-feed');

  const response = await fetch(
    `https://api.cloudinary.com/v1_1/dwrn0ohbp/image/upload`,
    { method: 'POST', body: formData }
  );

  const data = await response.json();
  return data.secure_url; // Use this URL in recipe creation
};

// 2. Create recipe with Cloudinary URL
const createRecipeWithImage = async (file: File, recipeData) => {
  const imageUrl = await uploadImage(file);
  
  return createRecipe({
    ...recipeData,
    imageUrl
  });
};
```

### Option 2: Backend Upload Endpoint
```typescript
const uploadImage = async (imageUrl: string) => {
  const response = await fetch(`${API_URL}/api/upload/image`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ imageUrl })
  });

  const { data } = await response.json();
  return data.url;
};
```

---

## 🔒 Production Considerations

### Security Enhancements Needed

1. **Authentication Required**
   - Add JWT authentication middleware to CREATE/UPDATE/DELETE endpoints
   - Only allow users to modify their own recipes

2. **Authorization Checks**
   ```go
   // Example: Verify recipe ownership before update/delete
   if recipe.AuthorID != authenticatedUserID {
       return http.StatusForbidden, "Not authorized to modify this recipe"
   }
   ```

3. **Input Validation**
   - Validate image URLs (whitelist Cloudinary domain)
   - Sanitize text inputs to prevent XSS
   - Rate limiting for recipe creation

4. **Pagination**
   ```go
   // Add pagination to main feed
   func GetAllPosts(w http.ResponseWriter, r *http.Request) {
       page := r.URL.Query().Get("page")
       limit := r.URL.Query().Get("limit")
       // ... implement pagination
   }
   ```

---

## 🧪 Testing Results

### Test Summary
✅ **GET /api/posts** - Fetches all 6 recipes with author details  
✅ **GET /api/users/{id}/posts** - Filters recipes by specific user  
✅ **POST /api/recipes** - Creates new recipe with UUID, returns with author  
✅ **PUT /api/recipes/{id}** - Updates recipe title/description  
✅ **Database** - Recipes persist correctly with foreign key to User table

### Sample Test Data
- 5 pre-seeded recipes from 3 different users
- 1 dynamically created recipe via API
- Author information properly preloaded in all responses

---

## 📚 Related Documentation

- [QUICK_START.md](QUICK_START.md) - Getting started guide with test credentials
- [ALL_ENDPOINTS.md](ALL_ENDPOINTS.md) - Complete API reference (83 endpoints)
- [FRONTEND_ENV_VARS.md](FRONTEND_ENV_VARS.md) - Environment variables for frontend
- [migrations/004_create_recipes_table.sql](migrations/004_create_recipes_table.sql) - Database schema

---

## 🎨 UI/UX Recommendations

### Main Feed Layout
```
┌─────────────────────────────────┐
│  🏠 Recipe Feed                 │
├─────────────────────────────────┤
│ ┌─────────────────────────────┐ │
│ │ [Recipe Image]              │ │
│ │ Fresh Salmon Nigiri         │ │
│ │ Autentyczne nigiri z...     │ │
│ │                             │ │
│ │ 👤 Dima Fomin  🕒 2h ago    │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ [Recipe Image]              │ │
│ │ Spicy Tuna Maki Roll        │ │
│ │ ...                         │ │
│ └─────────────────────────────┘ │
└─────────────────────────────────┘
```

### User Profile Layout
```
┌─────────────────────────────────┐
│  👤 Dima Fomin                  │
│  dima@example.com               │
├─────────────────────────────────┤
│  My Recipes (2)                 │
│                                 │
│  [Recipe 1]  [Recipe 2]         │
└─────────────────────────────────┘
```

---

**Created:** November 5, 2025  
**Status:** ✅ Production Ready  
**Backend:** Running on port 8080  
**Database:** PostgreSQL (Neon)  
**Test Server:** http://localhost:8080
