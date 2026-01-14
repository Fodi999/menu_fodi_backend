# 🗑️ Recipe Deletion API

## Endpoint
```
DELETE /api/admin/recipes/{id}
```

## Authentication
**Required:** Yes (Admin only)

## Request
```bash
# Headers
Authorization: Bearer <admin_jwt_token>

# URL Parameter
id - Recipe UUID
```

## Response

### Success (200 OK)
```json
{
  "success": true,
  "message": "Recipe deleted successfully"
}
```

### Not Found (404)
```json
{
  "error": "Recipe not found"
}
```

### Bad Request (400)
```json
{
  "error": "recipe ID is required"
}
```

### Unauthorized (401)
```json
{
  "error": "Authorization header required"
}
```

## What Gets Deleted

The deletion is **transactional** and removes:

1. ✅ Recipe main record
2. ✅ Recipe ingredients (cascade)
3. ✅ Recipe allergens (many-to-many)
4. ✅ Recipe diet tags (many-to-many)
5. ✅ All related associations

**If any step fails, the entire transaction is rolled back.**

## Example Usage

### Frontend (TypeScript/React)
```typescript
const deleteRecipe = async (recipeId: string) => {
  try {
    const response = await fetch(`/api/admin/recipes/${recipeId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${adminToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to delete recipe');
    }

    const data = await response.json();
    console.log(data.message); // "Recipe deleted successfully"
    
    // Refresh recipe list
    await fetchRecipes();
  } catch (error) {
    console.error('Delete failed:', error);
  }
};
```

### cURL Test
```bash
# Get admin token first
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.token')

# Delete recipe
curl -X DELETE "http://localhost:8080/api/admin/recipes/RECIPE_ID" \
  -H "Authorization: Bearer $TOKEN"
```

## Security Notes

- ⚠️ **Admin only** - Regular users cannot delete recipes
- ⚠️ **No soft delete** - Deletion is permanent
- ✅ **Transaction safety** - All-or-nothing deletion
- ✅ **Cascade handling** - All related data cleaned up

## Backend Logs

When deleting a recipe, you'll see:
```
🗑️  Deleting recipe: 696a2507-afe7-4a95-906e-cae7f7225e2f
✅ Recipe deleted: Лосось на Сковороде с Травами [696a2507-afe7-4a95-906e-cae7f7225e2f]
```

---

**Status:** ✅ **IMPLEMENTED**  
**Version:** 1.0.0  
**Last Updated:** 2026-01-11
