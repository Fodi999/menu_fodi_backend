# Cloudinary Image Upload - Frontend Integration Guide

## ✅ Implementation Status

**Backend Status:** ✅ Complete and deployed
- ✅ Cloudinary SDK integrated
- ✅ Admin endpoints implemented (POST/DELETE)
- ✅ Database migration applied (imageUrl, imagePublicId columns)
- ✅ RecipeCatalog model updated with image fields
- ✅ API returns imageUrl in responses
- ✅ Transactional integrity (cleanup on failures)

**Production URL:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`

---

## 📋 API Endpoints

### 1. Upload Recipe Image

**Endpoint:** `POST /api/admin/recipes/:id/image`

**Authentication:** Admin role required (JWT token in Authorization header)

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -F "image=@/path/to/image.jpg" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/6b8628ef-ef1e-42eb-a166-924566bb9b7b/image
```

**Request Body:** `multipart/form-data`
- Field name: `image`
- Max size: 5MB
- Allowed types: JPEG, PNG, WebP

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "imageUrl": "https://res.cloudinary.com/dwrn0ohbp/image/upload/v1768818751/recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b.webp",
    "imagePublicId": "recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b"
  },
  "message": "Image uploaded successfully"
}
```

**Error Responses:**
- `401 Unauthorized` - No JWT token or not admin
- `400 Bad Request` - No file, file too large (>5MB), or invalid format
- `404 Not Found` - Recipe not found
- `500 Internal Server Error` - Upload or database error

---

### 2. Delete Recipe Image

**Endpoint:** `DELETE /api/admin/recipes/:id/image`

**Authentication:** Admin role required

**Request:**
```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/6b8628ef-ef1e-42eb-a166-924566bb9b7b/image
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Image deleted successfully"
}
```

---

### 3. Get Recipes (with images)

**Endpoint:** `GET /api/recipes`

**Authentication:** Optional (public endpoint)

**Response:**
```json
{
  "success": true,
  "data": {
    "recipes": [
      {
        "id": "6b8628ef-ef1e-42eb-a166-924566bb9b7b",
        "canonicalName": "fried_salmon",
        "localName": "Smażony łosoś",
        "country": "Poland",
        "category": "main",
        "difficulty": "easy",
        "timeMinutes": 30,
        "servings": 4,
        "imageUrl": "https://res.cloudinary.com/dwrn0ohbp/image/upload/v1768818751/recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b.webp"
      }
    ],
    "count": 1
  }
}
```

**Note:** `imageUrl` field only appears if recipe has an uploaded image.

---

## 🖼️ Image Specifications

### Upload Transformations (automatic)
- **Format:** Auto-converted to WebP for optimal performance
- **Size:** 1200x800 pixels (16:9 aspect ratio)
- **Crop:** `fill` mode (maintains aspect ratio, crops if needed)
- **Quality:** Auto-optimized by Cloudinary
- **Overwrite:** Yes (uploading new image replaces old one)

### Public ID Pattern
```
recipes/recipe_{recipeID}
```

Example:
```
recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b
```

---

## 🎨 Frontend Implementation

### TypeScript Types

```typescript
// Response types
interface ImageUploadResponse {
  success: boolean;
  data: {
    imageUrl: string;
    imagePublicId: string;
  };
  message: string;
}

interface Recipe {
  id: string;
  canonicalName: string;
  localName: string;
  country: string;
  category: string;
  difficulty: string;
  timeMinutes: number;
  servings: number;
  imageUrl?: string; // Optional - only present if image exists
}
```

### React Upload Component Example

```tsx
import { useState } from 'react';

interface ImageUploadProps {
  recipeId: string;
  adminToken: string;
  onSuccess?: (imageUrl: string) => void;
}

export function RecipeImageUpload({ recipeId, adminToken, onSuccess }: ImageUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file size
    if (file.size > 5 * 1024 * 1024) {
      setError('File size must be less than 5MB');
      return;
    }

    // Validate file type
    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      setError('Only JPEG, PNG, and WebP images are allowed');
      return;
    }

    setUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('image', file);

      const response = await fetch(
        `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/${recipeId}/image`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${adminToken}`,
          },
          body: formData,
        }
      );

      const data: ImageUploadResponse = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Upload failed');
      }

      onSuccess?.(data.data.imageUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input
        type="file"
        accept="image/jpeg,image/png,image/webp"
        onChange={handleFileChange}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  );
}
```

### Responsive Image Display

```tsx
interface RecipeImageProps {
  imageUrl: string;
  alt: string;
}

export function RecipeImage({ imageUrl, alt }: RecipeImageProps) {
  // Generate Cloudinary thumbnail URLs
  const getThumbnailUrl = (size: 'small' | 'medium' | 'large') => {
    const dimensions = {
      small: { w: 200, h: 150 },
      medium: { w: 400, h: 300 },
      large: { w: 800, h: 600 },
    };

    const { w, h } = dimensions[size];
    // Insert transformation parameters before /upload/
    return imageUrl.replace('/upload/', `/upload/c_fill,w_${w},h_${h},q_auto/`);
  };

  return (
    <img
      src={imageUrl}
      srcSet={`
        ${getThumbnailUrl('small')} 200w,
        ${getThumbnailUrl('medium')} 400w,
        ${getThumbnailUrl('large')} 800w,
        ${imageUrl} 1200w
      `}
      sizes="(max-width: 600px) 200px, (max-width: 1200px) 400px, 800px"
      alt={alt}
      loading="lazy"
      style={{ width: '100%', height: 'auto' }}
    />
  );
}
```

### Delete Image Example

```typescript
async function deleteRecipeImage(recipeId: string, adminToken: string) {
  const response = await fetch(
    `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/${recipeId}/image`,
    {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${adminToken}`,
      },
    }
  );

  if (!response.ok) {
    const data = await response.json();
    throw new Error(data.message || 'Delete failed');
  }

  return await response.json();
}
```

---

## 🔐 Authentication

### Getting Admin JWT Token

1. **Login as admin:**
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }'
```

2. **Response contains JWT:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "...",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

3. **Use token in Authorization header:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## 🧪 Testing

### Test Image Upload
```bash
# 1. Login as admin
TOKEN=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.data.token')

# 2. Upload image
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -F "image=@test_image.jpg" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/YOUR_RECIPE_ID/image

# 3. Verify image appears in recipe list
curl -s https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes | jq '.data.recipes[] | select(.id == "YOUR_RECIPE_ID") | .imageUrl'
```

### Test Image Delete
```bash
curl -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/YOUR_RECIPE_ID/image
```

---

## 📊 Database Schema

### Recipe Table Columns
```sql
CREATE TABLE "Recipe" (
  ...existing columns...
  "imageUrl" TEXT,
  "imagePublicId" TEXT
);

CREATE INDEX idx_recipe_has_image ON "Recipe"(imageUrl) WHERE imageUrl IS NOT NULL;
```

---

## 🎯 Features

### ✅ Implemented
- Admin-only image upload/delete
- Automatic WebP conversion
- Image optimization (quality, size)
- Transactional integrity (cleanup on failures)
- Duplicate upload handling (overwrites old image)
- Both Recipe and RecipeCatalog models support images
- API returns imageUrl in catalog responses

### 🔄 Automatic Behaviors
- **Replace on upload:** Uploading new image automatically deletes old one
- **Cleanup on failure:** If DB save fails, uploaded image is deleted from Cloudinary
- **Graceful degradation:** Old image deletion failures don't block new upload
- **Responsive images:** Generate thumbnails on-the-fly using Cloudinary URL transformations

### 📏 Validation
- Max file size: 5MB
- Allowed formats: JPEG, PNG, WebP
- Admin role required
- Recipe must exist

---

## 🐛 Error Handling

### Frontend Error States
```typescript
type UploadError = 
  | 'FILE_TOO_LARGE'      // > 5MB
  | 'INVALID_FORMAT'      // Not JPEG/PNG/WebP
  | 'UNAUTHORIZED'        // No admin token
  | 'RECIPE_NOT_FOUND'    // Invalid recipe ID
  | 'UPLOAD_FAILED'       // Cloudinary error
  | 'DATABASE_ERROR';     // DB save failed

const errorMessages: Record<UploadError, string> = {
  FILE_TOO_LARGE: 'Image must be less than 5MB',
  INVALID_FORMAT: 'Only JPEG, PNG, and WebP images are allowed',
  UNAUTHORIZED: 'You must be an admin to upload images',
  RECIPE_NOT_FOUND: 'Recipe not found',
  UPLOAD_FAILED: 'Failed to upload image. Please try again.',
  DATABASE_ERROR: 'Image uploaded but failed to save. Contact support.',
};
```

---

## 🔗 Related Documentation

- **Transactional Integrity:** `IMAGE_UPLOAD_TRANSACTIONAL_INTEGRITY.md`
- **Backend Source of Truth:** `README_BACKEND_SOURCE_OF_TRUTH.md`
- **Cloudinary Docs:** https://cloudinary.com/documentation/go_integration

---

## ✅ Production Verification

**Test Recipe ID:** `6b8628ef-ef1e-42eb-a166-924566bb9b7b`

**Uploaded Image URL:**
```
https://res.cloudinary.com/dwrn0ohbp/image/upload/v1768818751/recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b.webp
```

**API Response:**
```bash
curl -s "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes" | \
  jq '.data.recipes[] | select(.id == "6b8628ef-ef1e-42eb-a166-924566bb9b7b") | {id, canonicalName, imageUrl}'
```

**Result:** ✅ Returns imageUrl correctly
