# 📸 Recipe Image Upload Guide

## Overview

Users can now attach photos to AI-generated recipes using Cloudinary integration.

---

## 🔄 Workflow

### 1. Create Recipe via Chef Mentor
```bash
POST /api/ai/chef-mentor/session
{
  "message": "Каліфорнія рол",
  "language": "ua"
}
```

**Response:**
```json
{
  "data": {
    "sessionId": "abc-123",
    "recipe": { "title": "Каліфорнія рол" }
  }
}
```

### 2. Complete Recipe
Continue conversation until `isComplete: true`.

Recipe auto-saves to `ai_generated_recipes` table with ID.

### 3. Upload Image to Cloudinary
```bash
POST /api/upload/image
Content-Type: multipart/form-data

file: [recipe_photo.jpg]
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "url": "https://res.cloudinary.com/your-cloud/image/upload/v123/recipe.jpg",
    "publicId": "recipe_abc123",
    "width": 1920,
    "height": 1080
  }
}
```

### 4. Attach Image to Recipe
```bash
POST /api/ai/recipes/{recipeId}/image
Content-Type: application/json

{
  "imageUrl": "https://res.cloudinary.com/your-cloud/image/upload/v123/recipe.jpg"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Recipe image updated",
  "imageUrl": "https://res.cloudinary.com/..."
}
```

---

## 📋 API Endpoints

### Upload Image
```http
POST /api/upload/image
Content-Type: multipart/form-data
```

**Parameters:**
- `file` (required): Image file (JPEG, PNG, WebP)

**Returns:**
- `url`: Public Cloudinary URL
- `publicId`: Cloudinary public ID
- `width`, `height`: Image dimensions

---

### Update Recipe Image
```http
POST /api/ai/recipes/{id}/image
Content-Type: application/json
```

**Body:**
```json
{
  "imageUrl": "https://res.cloudinary.com/..."
}
```

**Returns:**
```json
{
  "status": "success",
  "message": "Recipe image updated",
  "imageUrl": "..."
}
```

---

## 🗄️ Database Schema

### Updated AIGeneratedRecipe Model
```go
type AIGeneratedRecipe struct {
    ID          uuid.UUID
    Title       string
    Category    string
    ImageURL    string     // ✨ NEW FIELD
    Ingredients JSONB
    Nutrition   JSONB
    // ... other fields
}
```

**Migration:** Auto-applied via GORM AutoMigrate.

---

## 🎨 Frontend Integration

### React Example
```tsx
// 1. Upload image
const uploadImage = async (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch('/api/upload/image', {
    method: 'POST',
    body: formData
  });
  
  const data = await response.json();
  return data.data.url;
};

// 2. Attach to recipe
const attachImageToRecipe = async (recipeId: string, imageUrl: string) => {
  await fetch(`/api/ai/recipes/${recipeId}/image`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ imageUrl })
  });
};

// 3. Complete flow
const handleImageUpload = async (file: File, recipeId: string) => {
  try {
    const imageUrl = await uploadImage(file);
    await attachImageToRecipe(recipeId, imageUrl);
    console.log('✅ Image attached!');
  } catch (error) {
    console.error('❌ Upload failed:', error);
  }
};
```

---

## 🧪 Testing

### Test Image Upload
```bash
curl -X POST http://localhost:8080/api/upload/image \
  -F "file=@recipe_photo.jpg"
```

### Test Attach to Recipe
```bash
RECIPE_ID="abc-123-def"
IMAGE_URL="https://res.cloudinary.com/..."

curl -X POST "http://localhost:8080/api/ai/recipes/$RECIPE_ID/image" \
  -H "Content-Type: application/json" \
  -d "{\"imageUrl\":\"$IMAGE_URL\"}"
```

### Verify in Database
```bash
curl "http://localhost:8080/api/ai/recipes/$RECIPE_ID" | jq '.data.imageUrl'
```

---

## 🚀 Production

**Cloudinary Setup:**
1. Create account: https://cloudinary.com
2. Get credentials: `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`
3. Add to `.env`:
   ```env
   CLOUDINARY_CLOUD_NAME=your-cloud-name
   CLOUDINARY_API_KEY=your-api-key
   CLOUDINARY_API_SECRET=your-api-secret
   ```

**Image Optimization:**
- Automatic compression
- WebP conversion
- Responsive sizing
- CDN delivery

---

## ✅ Features

- ✅ Upload images via Cloudinary
- ✅ Attach to AI-generated recipes
- ✅ Store image URL in database
- ✅ Display in recipe details
- ✅ Show in marketplace
- ✅ Auto-migration support

---

## 📝 Notes

1. **Image URL is optional** - recipes can exist without photos
2. **Update anytime** - can change recipe image by posting new URL
3. **Cloudinary handles** - optimization, CDN, transformations
4. **Frontend can** - crop, resize, apply filters before upload

---

**Status:** ✅ Ready for production  
**Version:** 1.0  
**Updated:** 2025-11-06
