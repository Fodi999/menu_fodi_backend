# 🚀 API Quick Start for Frontend

**For:** Frontend Development Team
**Date:** November 10, 2024
**Status:** Production Ready

---

## ⭐ MOST IMPORTANT ENDPOINTS

### 1️⃣ User Authentication

#### Login
```bash
POST /api/auth/login
```
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```
**Returns:** `token` - Save this! Use in `Authorization: Bearer {token}`

#### Register
```bash
POST /api/auth/register
```
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "John Doe"
}
```
**Returns:** `token` - Same as login

---

### 2️⃣ Chat with AI Chef

#### Start/Continue Recipe Chat
```bash
POST /api/ai/chef-mentor
```
```json
{
  "message": "I want to make pasta carbonara",
  "language": "en",
  "currentRecipe": null,
  "conversationHistory": []
}
```

**Response includes:**
- `message` - AI's response
- `recipe` - Recipe being built
- `isComplete` - true when recipe is complete
- `suggestedActions` - ["save_recipe", "save_ingredients_to_fridge", "generate_meal_plan"]

---

### 3️⃣ 🆕 SAVE INGREDIENTS TO FRIDGE

#### Save Recipe Ingredients
```bash
POST /api/ai/save-ingredients
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```
```json
{
  "ingredients": [
    {"name": "Pasta", "amount": 400, "unit": "g"},
    {"name": "Eggs", "amount": 3, "unit": "pcs"},
    {"name": "Bacon", "amount": 200, "unit": "g"}
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "ingredients saved to fridge",
  "count": 3
}
```

✅ **THIS IS THE NEW FEATURE!** Use this when user clicks "Save to Fridge" button

---

### 4️⃣ View Fridge

#### Get All Fridge Items
```bash
GET /api/fridge/
Authorization: Bearer {JWT_TOKEN}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "product": "Pasta",
      "quantity": 400,
      "unit": "g",
      "available": true,
      "addedAt": "2024-11-10T12:34:56Z"
    }
  ]
}
```

---

### 5️⃣ Get Recipe Ideas

#### Fridge Recommendations
```bash
POST /api/ai/fridge-recommendations
Authorization: Bearer {JWT_TOKEN}
```
```json
{
  "cuisine": "italian",
  "maxTime": 30
}
```

**Response:** List of recipes that can be made with fridge items

---

### 6️⃣ Meal Planning

#### Generate Meal Plan
```bash
POST /api/ai/meal-plan
Authorization: Bearer {JWT_TOKEN}
```
```json
{
  "days": 3,
  "targetCalories": 2000,
  "language": "en"
}
```

**Response:** 3-day meal plan with breakfast, lunch, dinner

---

### 7️⃣ 🆕 UPLOAD IMAGE

#### Upload Image to Cloudinary
```bash
POST /api/upload/image
Authorization: Bearer {JWT_TOKEN}
Content-Type: multipart/form-data
```
```
Form Data:
  image: <File>  (JPEG, PNG, WebP, GIF, or SVG, max 10MB)
```

**Response:**
```json
{
  "success": true,
  "url": "http://res.cloudinary.com/...",
  "secureUrl": "https://res.cloudinary.com/...",
  "publicId": "culinary-academy/uuid",
  "message": "Image uploaded successfully"
}
```

✅ **NEW FEATURE!** Use for recipe photos, profile pictures, etc.

**JavaScript Example:**
```javascript
const formData = new FormData();
formData.append('image', fileInput.files[0]);

const response = await fetch('/api/upload/image', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('token')}`
  },
  body: formData
});

const data = await response.json();
const imageUrl = data.secureUrl;  // Use this URL
```

---

## 🔑 Key Points

### Token Usage
Every request that requires auth must include:
```
Authorization: Bearer {your_token_from_login}
```

### Required Headers
```
Content-Type: application/json
Authorization: Bearer {token}  (where needed)
```

### Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad request (check your input)
- `401` - Not authenticated (add token)
- `500` - Server error (contact support)

---

## 📱 Frontend Flow

### 1. User Opens App
```
→ Check if token exists (localStorage)
→ If yes, show home
→ If no, redirect to login
```

### 2. User Logs In
```
POST /api/auth/login
← Get token
→ Save token in localStorage
→ Show home/chat screen
```

### 3. User Starts Recipe
```
POST /api/ai/chef-mentor
← Get AI response
→ Show message + suggested next question
```

### 4. User Continues Chat
```
POST /api/ai/chef-mentor (with updated recipe)
← Get AI response
→ Update recipe on screen
→ When isComplete=true, show action buttons
```

### 5. User Saves to Fridge ⭐ NEW
```
POST /api/ai/save-ingredients
← Get count of saved items
→ Show "4 ingredients saved!" message
→ Refresh fridge view
```

### 6. User Views Fridge
```
GET /api/fridge/
← Get all items
→ Display in list/grid
→ Allow edit/delete
```

### 7. User Gets Ideas
```
POST /api/ai/fridge-recommendations
← Get recipe suggestions
→ Show recommendations
→ Click to start new chat
```

---

## 💻 Code Snippet

### React Example
```typescript
import { useState } from 'react';

export function FridgeChat() {
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [message, setMessage] = useState('');
  const [recipe, setRecipe] = useState(null);

  // Send message to Chef Mentor
  const sendMessage = async () => {
    const res = await fetch('/api/ai/chef-mentor', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        language: 'en',
        currentRecipe: recipe,
        conversationHistory: []
      })
    });
    const data = await res.json();
    setRecipe(data.recipe);
  };

  // Save ingredients to fridge
  const saveToFridge = async () => {
    const res = await fetch('/api/ai/save-ingredients', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ ingredients: recipe.ingredients })
    });
    const data = await res.json();
    console.log(`${data.count} ingredients saved!`);
  };

  return (
    <div>
      <input value={message} onChange={e => setMessage(e.target.value)} />
      <button onClick={sendMessage}>Ask Chef</button>
      
      {recipe?.isComplete && (
        <button onClick={saveToFridge}>Save to Fridge</button>
      )}
    </div>
  );
}
```

---

## 🚨 Common Issues & Solutions

### Issue: "401 Unauthorized"
**Solution:** 
- Check if token exists in localStorage
- Check if token is in Authorization header
- Get new token by logging in again

### Issue: "400 Bad Request"
**Solution:**
- Check JSON format is correct
- Verify all required fields are present
- Check for typos in field names

### Issue: Ingredients not saving
**Solution:**
- Verify token is valid
- Check ingredients array is not empty
- Check network request in browser DevTools

### Issue: Recipe not showing as complete
**Solution:**
- Ensure recipe has title, ingredients AND steps
- Check if AI parsed recipe correctly
- Try sending recipe data again

---

## 📝 Ingredients Format

Common units:
```
Weight: g, kg, mg, oz, lb
Volume: ml, l, cl, fl oz, cup, tbsp, tsp
Count: pcs, dozen, bunch
```

Example ingredient:
```json
{
  "name": "Pasta",
  "amount": 400,
  "unit": "g"
}
```

---

## 🔄 Conversation History

To continue conversation, pass previous messages:
```json
{
  "message": "next message",
  "conversationHistory": [
    {
      "role": "user",
      "content": "I want to make pasta"
    },
    {
      "role": "assistant",
      "content": "Great! What ingredients do you have?"
    }
  ]
}
```

---

## ✅ Deployment Checklist for Frontend

- [ ] Update API_BASE_URL to production URL
- [ ] Implement token storage (localStorage/sessionStorage)
- [ ] Add error handling for all API calls
- [ ] Show loading states during requests
- [ ] Handle 401 errors (redirect to login)
- [ ] Implement token refresh logic
- [ ] Test all endpoints in browser
- [ ] Add request timeouts
- [ ] Log errors to monitoring service

---

## 📞 Getting Help

**Documentation Files:**
1. `API_ENDPOINTS_FOR_FRONTEND.md` - Complete API reference
2. `FRIDGE_CHAT_INTEGRATION.md` - Feature guide
3. `ARCHITECTURE.md` - System architecture
4. `TEST_REPORT.md` - Testing results

**Contact:** Backend team available for questions

---

## 🎯 Next Steps

1. ✅ Review this document
2. ✅ Implement auth flow (login/token storage)
3. ✅ Build chat interface
4. ✅ Add save to fridge button when recipe complete
5. ✅ Display fridge items
6. ✅ Add recommendations feature
7. ✅ Test all endpoints
8. ✅ Deploy to production

---

**Last Updated:** November 10, 2024
**API Version:** 1.0.0
**Status:** Ready for Frontend Integration
