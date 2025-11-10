# API Endpoints for Frontend - Complete Reference

**Last Updated**: November 10, 2024
**Base URL**: `https://api.example.com` (adjust for your environment)
**Authentication**: JWT Bearer Token (where required)

---

## Table of Contents

1. [Authentication Endpoints](#authentication-endpoints)
2. [AI & Chef Mentor Endpoints](#ai--chef-mentor-endpoints)
3. [Fridge Endpoints](#fridge-endpoints)
4. [Recommendations Endpoints](#recommendations-endpoints)
5. [Health & Status Endpoints](#health--status-endpoints)
6. [Error Handling](#error-handling)
7. [Code Examples](#code-examples)

---

## Authentication Endpoints

### 1. Login
```
POST /api/auth/login
```

**No Auth Required**

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

**Error Response** (401):
```json
{
  "success": false,
  "message": "invalid credentials"
}
```

---

### 2. Register
```
POST /api/auth/register
```

**No Auth Required**

**Request Body**:
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "New User"
}
```

**Success Response** (201 Created):
```json
{
  "success": true,
  "message": "registration successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "newuser@example.com",
      "name": "New User"
    }
  }
}
```

---

### 3. Refresh Token
```
POST /api/auth/refresh
```

**Auth Required**: Yes (JWT Token)

**Request Body**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

## AI & Chef Mentor Endpoints

### 1. Chef Mentor Chat
```
POST /api/ai/chef-mentor
```

**Auth Required**: No

**Request Body**:
```json
{
  "message": "I want to make pasta carbonara",
  "language": "en",
  "currentRecipe": null,
  "conversationHistory": []
}
```

**Parameters**:
- `message` (string, required): User's message or query
- `language` (string, optional): "en", "ua", "ru", "pl" (default: "ua")
- `currentRecipe` (object, optional): Current recipe being built
- `conversationHistory` (array, optional): Previous messages in conversation

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "Great! Let me help you make pasta carbonara...",
  "recipe": {
    "title": "Pasta Carbonara",
    "ingredients": [
      {
        "name": "Pasta",
        "amount": 400,
        "unit": "g"
      }
    ],
    "steps": []
  },
  "nextQuestion": "What ingredients do you have?",
  "isComplete": false,
  "suggestedActions": null
}
```

**Complete Recipe Response**:
```json
{
  "success": true,
  "message": "Your recipe is complete!",
  "recipe": {
    "title": "Pasta Carbonara",
    "ingredients": [
      {"name": "Pasta", "amount": 400, "unit": "g"},
      {"name": "Eggs", "amount": 3, "unit": "pcs"},
      {"name": "Bacon", "amount": 200, "unit": "g"},
      {"name": "Parmesan", "amount": 100, "unit": "g"}
    ],
    "steps": [
      "Cook pasta al dente",
      "Fry bacon until crispy",
      "Mix eggs with cheese",
      "Combine everything"
    ]
  },
  "nextQuestion": null,
  "isComplete": true,
  "suggestedActions": [
    "save_recipe",
    "save_ingredients_to_fridge",
    "generate_meal_plan"
  ]
}
```

---

### 2. Generate Recipe from Title
```
POST /api/ai/recipe-generator
```

**Auth Required**: No

**Request Body**:
```json
{
  "title": "Margherita Pizza",
  "language": "en"
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "recipe": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "title": "Margherita Pizza",
    "description": "Classic Italian pizza with fresh mozzarella and basil",
    "difficulty": "intermediate",
    "time": 30,
    "portions": 4,
    "ingredients": [
      {"name": "Flour", "amount": 300, "unit": "g"},
      {"name": "Mozzarella", "amount": 250, "unit": "g"},
      {"name": "Tomato", "amount": 400, "unit": "g"},
      {"name": "Basil", "amount": 20, "unit": "g"}
    ],
    "steps": [
      "Prepare dough",
      "Add sauce",
      "Add cheese",
      "Bake at 220°C for 15 minutes"
    ]
  }
}
```

---

### 3. Save Recipe Ingredients to Fridge ⭐ NEW
```
POST /api/ai/save-ingredients
```

**Auth Required**: Yes (JWT Token)

**Headers**:
```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body**:
```json
{
  "ingredients": [
    {
      "name": "Pasta",
      "amount": 400,
      "unit": "g"
    },
    {
      "name": "Eggs",
      "amount": 3,
      "unit": "pcs"
    },
    {
      "name": "Bacon",
      "amount": 200,
      "unit": "g"
    },
    {
      "name": "Parmesan Cheese",
      "amount": 100,
      "unit": "g"
    }
  ]
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "ingredients saved to fridge",
  "count": 4
}
```

**Error Response - Missing Auth** (401):
```json
{
  "error": "missing or invalid authentication token"
}
```

**Error Response - Empty Ingredients** (400):
```json
{
  "error": "ingredients list cannot be empty"
}
```

**Error Response - DB Error** (500):
```json
{
  "error": "failed to save ingredients to database"
}
```

---

## Fridge Endpoints

### 1. Get All Fridge Items
```
GET /api/fridge/
```

**Auth Required**: Yes (JWT Token)

**Headers**:
```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Query Parameters**:
- `page` (integer, optional): Page number (default: 1)
- `limit` (integer, optional): Items per page (default: 20)
- `available` (boolean, optional): Filter by availability

**Success Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "product": "Pasta",
      "quantity": 400,
      "unit": "g",
      "available": true,
      "category": "grains",
      "addedAt": "2024-11-10T12:34:56Z",
      "updatedAt": "2024-11-10T12:34:56Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "product": "Eggs",
      "quantity": 3,
      "unit": "pcs",
      "available": true,
      "category": "dairy",
      "addedAt": "2024-11-10T12:34:56Z",
      "updatedAt": "2024-11-10T12:34:56Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 2
  }
}
```

---

### 2. Get Available Fridge Items
```
GET /api/fridge/available
```

**Auth Required**: Yes (JWT Token)

**Success Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "product": "Pasta",
      "quantity": 400,
      "unit": "g",
      "available": true
    }
  ]
}
```

---

### 3. Add Item to Fridge
```
POST /api/fridge/
```

**Auth Required**: Yes (JWT Token)

**Request Body**:
```json
{
  "product": "Tomato",
  "quantity": 500,
  "unit": "g",
  "category": "vegetables",
  "available": true
}
```

**Success Response** (201 Created):
```json
{
  "success": true,
  "message": "item added to fridge",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440099",
    "product": "Tomato",
    "quantity": 500,
    "unit": "g",
    "available": true,
    "addedAt": "2024-11-10T13:00:00Z"
  }
}
```

---

### 4. Update Fridge Item
```
PUT /api/fridge/{id}
```

**Auth Required**: Yes (JWT Token)

**URL Parameters**:
- `id`: Item UUID

**Request Body**:
```json
{
  "quantity": 300,
  "available": false
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "item updated",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "product": "Pasta",
    "quantity": 300,
    "unit": "g",
    "available": false,
    "updatedAt": "2024-11-10T13:05:00Z"
  }
}
```

---

### 5. Delete Fridge Item
```
DELETE /api/fridge/{id}
```

**Auth Required**: Yes (JWT Token)

**URL Parameters**:
- `id`: Item UUID

**Success Response** (200 OK):
```json
{
  "success": true,
  "message": "item deleted"
}
```

---

## Recommendations Endpoints

### 1. Get Fridge Recommendations
```
POST /api/ai/fridge-recommendations
```

**Auth Required**: Yes (JWT Token)

**Request Body**:
```json
{
  "cuisine": "italian",
  "maxTime": 30,
  "language": "en"
}
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "recommendations": [
    {
      "recipeName": "Pasta Carbonara",
      "description": "Classic Italian pasta dish",
      "matchPercentage": 95,
      "missingItems": ["Bacon", "Eggs"],
      "prepTime": 20,
      "difficulty": "easy"
    },
    {
      "recipeName": "Tomato Pasta",
      "description": "Simple and delicious",
      "matchPercentage": 85,
      "missingItems": ["Tomato sauce"],
      "prepTime": 15,
      "difficulty": "easy"
    }
  ]
}
```

---

### 2. Generate Meal Plan
```
POST /api/ai/meal-plan
```

**Auth Required**: Yes (JWT Token)

**Request Body**:
```json
{
  "days": 3,
  "targetCalories": 2000,
  "language": "en"
}
```

**Parameters**:
- `days` (integer, required): 1-14 days
- `targetCalories` (integer, optional): Target daily calories (default: 2000)
- `language` (string, optional): "en", "ua", "ru", "pl"

**Success Response** (200 OK):
```json
{
  "success": true,
  "plan": [
    {
      "day": "Day 1",
      "breakfast": "Scrambled eggs with toast",
      "lunch": "Pasta carbonara",
      "dinner": "Grilled chicken with vegetables",
      "totalCalories": 2150
    },
    {
      "day": "Day 2",
      "breakfast": "Oatmeal with fruits",
      "lunch": "Pizza margherita",
      "dinner": "Fish with salad",
      "totalCalories": 2050
    },
    {
      "day": "Day 3",
      "breakfast": "Pancakes with jam",
      "lunch": "Burger with fries",
      "dinner": "Pasta with tomato sauce",
      "totalCalories": 2200
    }
  ],
  "totalCalories": 6400,
  "avgPerDay": 2133
}
```

---

## Health & Status Endpoints

### 1. Health Check
```
GET /health
```

**Auth Required**: No

**Success Response** (200 OK):
```json
{
  "status": "healthy",
  "uptime": "24h",
  "timestamp": "2024-11-10T14:00:00Z"
}
```

---

### 2. API Status
```
GET /api/status
```

**Auth Required**: No

**Success Response** (200 OK):
```json
{
  "status": "operational",
  "version": "1.0.0",
  "services": {
    "database": "connected",
    "ai": "available",
    "cache": "working"
  }
}
```

---

## Error Handling

### Common Error Responses

**400 Bad Request**:
```json
{
  "error": "invalid request body",
  "details": "missing required field: ingredients"
}
```

**401 Unauthorized**:
```json
{
  "error": "missing or invalid authentication token",
  "hint": "add Authorization header with valid JWT token"
}
```

**403 Forbidden**:
```json
{
  "error": "you do not have permission to access this resource"
}
```

**404 Not Found**:
```json
{
  "error": "resource not found",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**500 Internal Server Error**:
```json
{
  "error": "internal server error",
  "requestId": "abc123def456"
}
```

---

## Code Examples

### JavaScript/TypeScript (Fetch API)

```javascript
// 1. Login
async function login(email, password) {
  const response = await fetch('https://api.example.com/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  const data = await response.json();
  return data.data.token; // Save this token
}

// 2. Chef Mentor Chat
async function chefMentorChat(message, token) {
  const response = await fetch('https://api.example.com/api/ai/chef-mentor', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message: message,
      language: 'en',
      currentRecipe: null,
      conversationHistory: []
    })
  });
  return await response.json();
}

// 3. Save Ingredients to Fridge
async function saveIngredientsToFridge(ingredients, token) {
  const response = await fetch('https://api.example.com/api/ai/save-ingredients', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ ingredients })
  });
  return await response.json();
}

// 4. Get Fridge Items
async function getFridgeItems(token) {
  const response = await fetch('https://api.example.com/api/fridge/', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  return await response.json();
}

// 5. Get Recommendations
async function getFridgeRecommendations(cuisine, maxTime, token) {
  const response = await fetch('https://api.example.com/api/ai/fridge-recommendations', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      cuisine,
      maxTime,
      language: 'en'
    })
  });
  return await response.json();
}

// 6. Generate Meal Plan
async function generateMealPlan(days, calories, token) {
  const response = await fetch('https://api.example.com/api/ai/meal-plan', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      days,
      targetCalories: calories,
      language: 'en'
    })
  });
  return await response.json();
}

// Usage Example
(async () => {
  const token = await login('user@example.com', 'password123');
  
  const chefResponse = await chefMentorChat('I want pasta', token);
  console.log(chefResponse);
  
  if (chefResponse.isComplete) {
    const saveResult = await saveIngredientsToFridge(
      chefResponse.recipe.ingredients,
      token
    );
    console.log(`Saved ${saveResult.count} ingredients!`);
  }
})();
```

### React Hook Example

```typescript
import { useState } from 'react';

function FridgeChatComponent() {
  const [token, setToken] = useState<string>('');
  const [message, setMessage] = useState('');
  const [recipe, setRecipe] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleChat = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/ai/chef-mentor', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message,
          language: 'en',
          currentRecipe: recipe,
          conversationHistory: []
        })
      });
      const data = await response.json();
      setRecipe(data.recipe);
    } catch (error) {
      console.error('Error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveToFridge = async () => {
    if (!recipe?.ingredients) return;
    
    try {
      const response = await fetch('/api/ai/save-ingredients', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ ingredients: recipe.ingredients })
      });
      const data = await response.json();
      if (data.success) {
        alert(`${data.count} ingredients saved!`);
      }
    } catch (error) {
      console.error('Error saving ingredients:', error);
    }
  };

  return (
    <div>
      <input
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        placeholder="Tell me what you want to cook..."
      />
      <button onClick={handleChat} disabled={loading}>
        {loading ? 'Thinking...' : 'Chat with Chef'}
      </button>

      {recipe && (
        <>
          <h2>{recipe.title}</h2>
          <button onClick={handleSaveToFridge}>
            Save to Fridge
          </button>
        </>
      )}
    </div>
  );
}

export default FridgeChatComponent;
```

---

## API Summary Table

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/auth/login` | POST | No | User login |
| `/api/auth/register` | POST | No | User registration |
| `/api/auth/refresh` | POST | Yes | Refresh JWT token |
| `/api/ai/chef-mentor` | POST | No | Chat with AI chef |
| `/api/ai/recipe-generator` | POST | No | Generate recipe from title |
| `/api/ai/save-ingredients` | POST | **Yes** | **Save ingredients to fridge** |
| `/api/fridge/` | GET | Yes | Get all fridge items |
| `/api/fridge/available` | GET | Yes | Get available items only |
| `/api/fridge/` | POST | Yes | Add item to fridge |
| `/api/fridge/{id}` | PUT | Yes | Update fridge item |
| `/api/fridge/{id}` | DELETE | Yes | Delete fridge item |
| `/api/ai/fridge-recommendations` | POST | Yes | Get recipe recommendations |
| `/api/ai/meal-plan` | POST | Yes | Generate meal plan |
| `/health` | GET | No | Health check |
| `/api/status` | GET | No | API status |

---

## Quick Reference

### Authentication Header
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Standard Response Format
```json
{
  "success": boolean,
  "message": string,
  "data": object | array,
    "error": string (if failed)
}
```

### Units Commonly Used
- Weight: g, kg, mg
- Volume: ml, l, cl
- Count: pcs, dozen
- Other: tbsp, tsp, cup, bowl

---

## Image Upload Endpoints

### POST /api/upload/image

Upload an image file to Cloudinary and get the image URL.

**Authentication**: Required (JWT Bearer Token)

**Request Type**: `multipart/form-data`

**Request Parameters**:
```
image: File (required)
  - Supported formats: JPEG, PNG, WebP, GIF, SVG
  - Max size: 10 MB
```

**Success Response** (200 OK):
```json
{
  "success": true,
  "url": "http://res.cloudinary.com/...",
  "secureUrl": "https://res.cloudinary.com/...",
  "publicId": "culinary-academy/550e8400-e29b-41d4-a716-446655440000",
  "message": "Image uploaded successfully"
}
```

**Error Responses**:

Unauthorized (401):
```json
{
  "error": "unauthorized"
}
```

Bad Request (400) - Missing file:
```json
{
  "error": "image file is required"
}
```

Bad Request (400) - Invalid file type:
```json
{
  "error": "invalid file type - only JPEG, PNG, WebP, GIF, and SVG are allowed"
}
```

Bad Request (400) - File too large:
```json
{
  "error": "file size exceeds 10MB limit"
}
```

Internal Error (500):
```json
{
  "error": "failed to upload image"
}
```

**Frontend Example**:

```javascript
// Using Fetch API
async function uploadImage(file, authToken) {
  const formData = new FormData();
  formData.append('image', file);

  const response = await fetch('/api/upload/image', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${authToken}`
    },
    body: formData
  });

  const data = await response.json();
  
  if (response.ok) {
    console.log('Image uploaded:', data.secureUrl);
    return data.secureUrl;
  } else {
    console.error('Upload failed:', data.error);
    return null;
  }
}

// Usage with localStorage token
const authToken = localStorage.getItem('authToken');
const imageUrl = await uploadImage(fileInput.files[0], authToken);
```

**React Component Example**:

```jsx
import { useState } from 'react';

function ImageUploader() {
  const [uploading, setUploading] = useState(false);
  const [imageUrl, setImageUrl] = useState(null);

  const handleUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('image', file);

      const response = await fetch('/api/upload/image', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error('Upload failed');
      }

      const data = await response.json();
      setImageUrl(data.secureUrl);
    } catch (error) {
      console.error('Error:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input 
        type="file" 
        accept="image/*" 
        onChange={handleUpload}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
      {imageUrl && <img src={imageUrl} alt="Uploaded" />}
    </div>
  );
}

export default ImageUploader;
```

**curl Example**:

```bash
# Create a test image
convert -size 100x100 xc:red test.jpg

# Upload with authentication
curl -X POST http://localhost:8080/api/upload/image \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "image=@test.jpg"
```

---

## Important Notes

1. **JWT Token**: Save the token from login response and include in Authorization header for protected endpoints
2. **Image Upload**: Use `secureUrl` (HTTPS) for secure image delivery
3. **File Storage**: Images are stored on Cloudinary (cloud-based, not local)
4. **Base URL**: Update `https://api.example.com` with your actual API base URL
5. **CORS**: Frontend should be whitelisted in backend CORS configuration
6. **Timeouts**: Consider implementing request timeout (30 seconds recommended for uploads)
7. **Rate Limiting**: API may have rate limits - implement retry logic
8. **Error Handling**: Always check `success` field and handle errors appropriately
9. **Image Formats**: JPEG and PNG are recommended for compatibility
10. **localStorage**: Store authToken in localStorage and use in all protected requests

---

## Support

For questions about these endpoints:
- Check `FRIDGE_CHAT_INTEGRATION.md` for feature details
- Check `IMAGE_UPLOAD_GUIDE.md` for detailed image upload documentation
- Review `ARCHITECTURE.md` for system design
- Check server logs for detailed error information

````
}
```

### Units Commonly Used
- Weight: g, kg, mg
- Volume: ml, l, cl
- Count: pcs, dozen
- Other: tbsp, tsp, cup, bowl

---

## Important Notes

1. **JWT Token**: Save the token from login response and include in Authorization header for protected endpoints
2. **Base URL**: Update `https://api.example.com` with your actual API base URL
3. **CORS**: Frontend should be whitelisted in backend CORS configuration
4. **Timeouts**: Consider implementing request timeout (30 seconds recommended)
5. **Rate Limiting**: API may have rate limits - implement retry logic
6. **Error Handling**: Always check `success` field and handle errors appropriately

---

## Support

For questions about these endpoints:
- Check `FRIDGE_CHAT_INTEGRATION.md` for feature details
- Review `ARCHITECTURE.md` for system design
- Check server logs for detailed error information
