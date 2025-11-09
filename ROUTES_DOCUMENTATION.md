# Backend API Routes Documentation

## Architecture
The backend uses a **Domain-Driven Design (DDD)** modular architecture with Chi router.

All routes are grouped under `/api` prefix and are organized into 20 domain modules.

## Module Routes Structure

### 1. **Auth Module** (`/api/auth`)
- Authentication, registration, login
- **Public routes**
- JWT token management

### 2. **Wallet Module** (`/api/wallet`)
- User wallet and balance management
- Token balance operations

### 3. **User Module** (`/api/user`) - Protected
- User profile management
- User progress and dashboard
- User achievements

### 4. **Fridge Module** (`/api/fridge`) - Protected
- User fridge item management
- Add, update, delete fridge items
- Get available items

### 5. **AI Module** (`/api/ai`)
- Chef Mentor AI chat
- Recipe generation
- AI-powered recommendations
- Meal plan generation (Protected)

### 6. **Marketplace Module** (`/api/marketplace`)
- Recipe marketplace
- Leaderboard and seller stats
- Purchase recipes (Protected)

### 7. **Academy Module** (`/api/academy`)
- Online courses and lessons
- Course enrollment (Protected)
- Quiz submissions (Protected)
- Certificate generation (Protected)

### 8. **Hint Module** (`/api/hint`) - Protected
- Cooking hints and tips

### 9. **Ingredients Module** (`/api/ingredients`) - Protected
- Ingredient database and management
- Stock movements tracking

### 10. **Leaderboard Module** (`/api/leaderboard`)
- User rankings and leaderboard

### 11. **Admin Module** (`/api/admin`) - Protected + Admin
- User management
- Order management
- System statistics
- **Requires admin middleware**

### 12. **Business Module** (`/api/business`) - Protected
- Business profile management
- Business analytics
- Subscription management

### 13. **Meal Plan Module** (`/api/mealplan`) - Protected
- AI-generated meal plans
- Meal plan management

### 14. **Metrics Module** (`/api/metrics`) - Protected
- Business KPI calculations
- Performance metrics

### 15. **Nutrition Module** (`/api/nutrition`)
- Recipe nutrition analysis
- Custom nutrition calculations
- Ingredient macro data

### 16. **Recipes Module** (`/api/recipes`) - Protected
- Social recipe posting
- Recipe CRUD operations
- Recipe viewing and interaction
- Token reward system

### 17. **Semi-Finished Module** (`/api/semi-finished`) - Protected
- Semi-finished ingredient management
- Stock operations

### 18. **Stats Module** (`/api/stats`) - Protected
- Admin statistics
- Recent orders tracking

### 19. **Health Module** (`/health`)
- Health check endpoint
- Database status

### 20. **Contact Module** (`/contact`)
- Contact form submissions
- **Public endpoint**

## Key Features

### Authentication
- JWT-based authentication
- Middleware stack: `AuthMiddleware`, `AdminMiddleware`
- All protected routes require valid JWT token

### Response Format
All endpoints use standardized JSON response:
```json
{
  "success": true,
  "data": { /* endpoint-specific data */ }
}
```

Error responses:
```json
{
  "code": 400,
  "message": "Error description",
  "success": false
}
```

### Middleware
- **AuthMiddleware**: Validates JWT token
- **AdminMiddleware**: Checks for admin role
- Global: Request ID, Real IP, Logger, Recoverer, CORS, Timeout (60s)

### CORS
- Allowed Origins: `*` (all)
- Allowed Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
- Allowed Headers: Accept, Authorization, Content-Type, X-CSRF-Token
- Max Age: 300 seconds

## Module Registration

All modules are initialized and registered in `/internal/app/routes_modular.go`:

```go
func (a *App) setupModularRoutes() http.Handler {
    // Initialize 20 modules
    // Register all module routes within r.Route("/api", ...)
}
```

## Protected vs Public

- **Public routes**: No authentication required
- **Protected routes**: Require valid JWT token (requires middleware)
- **Admin routes**: Require valid JWT + admin role

## Architecture Benefits

1. **Clean Separation of Concerns**: Each module owns its domain logic
2. **Scalability**: Easy to add new modules without affecting existing ones
3. **Maintainability**: Clear structure and organization
4. **DDD Principles**: Modules follow Domain-Driven Design patterns
5. **Service Layer**: Each module has independent business logic
6. **Repository Pattern**: Data access is encapsulated in repositories
7. **DTO Pattern**: Request/response validation and transformation

## Module Structure

Each module follows this structure:
```
module_name/
├── module.go              (initialization & route registration)
├── transport/http/        (HTTP handlers)
├── service/               (business logic)
├── repo/                  (data access - if needed)
└── dto/                   (data structures)
```
