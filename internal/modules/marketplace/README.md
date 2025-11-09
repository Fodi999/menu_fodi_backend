# Marketplace Module

## Overview
The **Marketplace Module** provides a complete recipe marketplace system where users can buy and sell personal recipes using ChefTokens. The module handles complex transactions with platform commission, seller statistics, and a global leaderboard.

## Architecture

```
marketplace/
├── dto/
│   └── requests.go      # DTOs for marketplace operations
├── repo/
│   └── repository.go    # Data access layer with complex queries
├── service/
│   └── service.go       # Business logic with transaction handling
├── transport/
│   └── http/
│       └── handlers.go  # HTTP handlers with authentication
├── module.go            # Route registration
└── README.md           # This file
```

## Business Model

### Platform Commission
- **Commission Rate**: 10% of every sale
- **Buyer Pays**: Full recipe price
- **Seller Receives**: 90% of price (after commission)
- **Platform Receives**: 10% commission

### Transaction Flow
1. Buyer initiates purchase
2. System validates:
   - Recipe exists
   - Buyer has sufficient balance
   - Buyer doesn't own the recipe
   - Recipe not already purchased
3. Transaction executes atomically:
   - Deduct full price from buyer's wallet
   - Add net amount (90%) to seller's wallet
   - Create purchase record
   - Increment recipe purchase counter
   - Create wallet transaction records for audit
4. Return purchase confirmation with commission details

## Features

### 1. Recipe Marketplace
Browse recipes with advanced filtering and sorting:
- **Filters**: Category, difficulty, price range, minimum rating
- **Sorting**: Popular, newest, rating, price
- **Enrichment**: Author info, review statistics
- **Pagination**: Configurable limit (default 50)

### 2. Purchase System
Secure transaction handling:
- **Validation**: Own recipe check, duplicate prevention
- **Balance Check**: Insufficient funds detection
- **Atomic Operations**: Transaction rollback on error
- **Audit Trail**: Complete wallet transaction history
- **Commission Calculation**: Automatic 10% platform fee

### 3. Seller Statistics
Comprehensive seller analytics:
- Total sales count
- Total revenue earned
- Average rating across all recipes
- Top-selling recipe identification

### 4. Global Leaderboard
Multi-dimensional ranking system:
- **Sort Options**: XP, sales, rating, revenue
- **Filters**: Language preference
- **Data**: User stats with achievements, recipes, sales
- **Complex Aggregation**: 4-table JOIN with calculated fields

### 5. Purchase History
User purchase tracking:
- List of all purchased recipes
- Recipe details enrichment
- Purchase timestamp
- Easy access to owned content

## API Endpoints

### Public Endpoints

#### GET /api/marketplace/recipes
Browse marketplace recipes with filters.

**Query Parameters:**
- `category` (string): Filter by recipe category
- `difficulty` (string): Filter by difficulty level
- `maxPrice` (int): Maximum price in ChefTokens
- `minRating` (float): Minimum rating (1-5)
- `sortBy` (string): Sort order (popular, newest, rating, price)
- `limit` (int): Results per page (default 50)

**Response:**
```json
{
  "recipes": [
    {
      "id": "uuid",
      "title": "Italian Carbonara",
      "description": "Authentic Roman pasta",
      "category": "italian",
      "difficulty": "medium",
      "price": 500,
      "userId": "uuid",
      "authorName": "Chef Mario",
      "authorLevel": 12,
      "authorAvatar": "https://...",
      "reviewCount": 45,
      "avgReview": 4.7,
      "purchases": 230,
      "views": 1523,
      "createdAt": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

#### GET /api/marketplace/leaderboard
Global leaderboard with rankings.

**Query Parameters:**
- `sortBy` (string): Sort field (xp, sales, rating, revenue) - default "xp"
- `language` (string): Filter by language preference
- `limit` (int): Number of entries (default 50, max 100)

**Response:**
```json
{
  "leaderboard": [
    {
      "rank": 1,
      "userId": "uuid",
      "username": "ChefMaster",
      "avatarUrl": "https://...",
      "level": 25,
      "xp": 15000,
      "totalSales": 450,
      "totalRevenue": 125000,
      "avgRating": 4.8,
      "recipesCount": 12,
      "achievementsCount": 18
    }
  ],
  "total": 1
}
```

#### GET /api/marketplace/stats/{userId}
Get seller statistics for a user.

**Path Parameters:**
- `userId` (uuid): User ID to get stats for

**Response:**
```json
{
  "userId": "uuid",
  "totalSales": 450,
  "totalRevenue": 125000,
  "avgRating": 4.8,
  "topRecipe": {
    "id": "uuid",
    "title": "Perfect Risotto",
    "purchases": 180
  }
}
```

### Protected Endpoints (JWT Required)

#### POST /api/marketplace/purchase
Purchase a recipe.

**Request:**
```json
{
  "recipeId": "uuid"
}
```

**Response:**
```json
{
  "success": true,
  "purchase": {
    "id": "uuid",
    "recipeId": "uuid",
    "buyerId": "uuid",
    "price": 500,
    "purchasedAt": "2024-01-20T14:30:00Z"
  },
  "commission": 50,
  "sellerAmount": 450,
  "message": "Recipe purchased successfully"
}
```

**Error Responses:**
- `404 Not Found`: Recipe doesn't exist
- `400 Bad Request`: 
  - Cannot buy own recipe
  - Already purchased this recipe
  - Insufficient funds

#### GET /api/marketplace/purchases
Get user's purchased recipes.

**Response:**
```json
{
  "purchases": [
    {
      "id": "uuid",
      "recipeId": "uuid",
      "recipe": {
        "id": "uuid",
        "title": "Italian Carbonara",
        "description": "Authentic Roman pasta",
        "category": "italian",
        "price": 500
      },
      "purchasedAt": "2024-01-20T14:30:00Z"
    }
  ]
}
```

## Database Schema

### Tables Used

#### PersonalRecipe
- `id` (uuid): Recipe ID
- `user_id` (uuid): Seller ID
- `title` (string): Recipe title
- `description` (text): Recipe description
- `category` (string): Recipe category
- `difficulty` (string): Difficulty level
- `price` (int): Price in ChefTokens
- `purchases` (int): Purchase counter
- `views` (int): View counter
- `created_at` (timestamp): Creation time

#### RecipePurchase
- `id` (uuid): Purchase ID
- `recipe_id` (uuid): Purchased recipe
- `buyer_id` (uuid): Buyer user ID
- `price` (int): Purchase price
- `purchased_at` (timestamp): Purchase time

#### UserProfile
- `id` (uuid): User ID
- `name` (string): User name
- `level` (int): User level
- `xp` (int): Experience points
- `wallet_balance` (decimal): ChefToken balance
- `avatar_url` (string): Avatar URL

#### WalletTransaction
- `id` (uuid): Transaction ID
- `user_id` (uuid): User ID
- `amount` (decimal): Amount (positive/negative)
- `transaction_type` (string): Type (recipe_purchase, recipe_sale)
- `reference_id` (uuid): Related recipe/purchase ID
- `created_at` (timestamp): Transaction time

#### PersonalRecipeReview
- `recipe_id` (uuid): Reviewed recipe
- `user_id` (uuid): Reviewer
- `rating` (int): Rating (1-5)
- `created_at` (timestamp): Review time

## Repository Layer

### MarketplaceRepository Interface

```go
type MarketplaceRepository interface {
    // Recipe marketplace
    GetMarketRecipes(filters *dto.MarketplaceFilters) ([]*dto.RecipeWithAuthor, int64, error)
    GetRecipeByID(recipeID uuid.UUID) (*model.PersonalRecipe, error)
    IncrementPurchases(recipeID uuid.UUID) error
    
    // Purchase management
    CheckPurchaseExists(buyerID, recipeID uuid.UUID) (bool, error)
    CreatePurchase(purchase *model.RecipePurchase) error
    GetUserPurchases(userID uuid.UUID) ([]*dto.UserPurchase, error)
    
    // Wallet operations
    GetUserProfile(userID uuid.UUID) (*model.UserProfile, error)
    UpdateWalletBalance(userID uuid.UUID, amount float64) error
    CreateWalletTransaction(tx *model.WalletTransaction) error
    
    // Statistics
    GetSellerStats(userID uuid.UUID) (*dto.SellerStats, error)
    GetLeaderboard(sortBy, language string, limit int) ([]*dto.LeaderboardEntry, error)
}
```

### Key Repository Methods

#### GetMarketRecipes
- Applies category, difficulty, price, rating filters
- Enriches with author info via JOIN
- Aggregates review statistics
- Supports 4 sort modes: popular, newest, rating, price
- Returns paginated results

#### PurchaseRecipe (Service Layer)
Transaction steps:
1. Fetch recipe (validate exists)
2. Check ownership (cannot buy own)
3. Check duplicate purchase
4. Verify buyer/seller profiles
5. Check buyer balance
6. Calculate commission (10%)
7. Create purchase record
8. Update buyer wallet (-price)
9. Update seller wallet (+netAmount)
10. Increment purchase counter
11. Create transaction records (2x)

#### GetLeaderboard
Complex query with 4 LEFT JOINs:
- UserProfile (base data)
- RecipePurchase (sales aggregation)
- PersonalRecipe (recipe counts)
- UserAchievement (achievement counts)
- Dynamic ORDER BY based on sortBy
- COALESCE for null handling
- Optional language filter

## Service Layer

### Business Rules

#### Purchase Validation
1. **Recipe Exists**: Must be valid marketplace recipe
2. **Ownership**: Cannot purchase own recipe
3. **Duplicate**: Cannot purchase same recipe twice
4. **Balance**: Buyer must have sufficient ChefTokens

#### Commission Calculation
```go
const PlatformCommissionRate = 0.10

commission := float64(recipe.Price) * PlatformCommissionRate
netAmount := float64(recipe.Price) - commission
```

#### Transaction Safety
- All wallet updates are atomic via `gorm.Expr`
- Database transactions ensure consistency
- Rollback on any error
- Audit trail for all monetary changes

## Error Handling

### Custom Errors
```go
var (
    ErrRecipeNotFound       = errors.New("recipe not found")
    ErrCannotBuyOwnRecipe   = errors.New("cannot purchase your own recipe")
    ErrAlreadyPurchased     = errors.New("you have already purchased this recipe")
    ErrInsufficientFunds    = errors.New("insufficient funds")
)
```

### HTTP Status Codes
- `200 OK`: Successful operation
- `400 Bad Request`: Validation errors, business rule violations
- `401 Unauthorized`: Missing/invalid JWT token
- `404 Not Found`: Recipe not found
- `500 Internal Server Error`: Database/system errors

## Security

### Authentication
- JWT middleware protects purchase and purchase history endpoints
- Public endpoints for browsing and leaderboard
- User ID extracted from JWT token (cannot be spoofed)

### Authorization
- Purchase endpoint auto-sets `buyerId` from JWT
- Cannot manipulate buyer ID in request
- Seller receives payment automatically

### Transaction Security
- Atomic wallet updates prevent race conditions
- Database transactions ensure consistency
- Balance checks prevent overdraft
- Duplicate purchase prevention

## Integration

### Dependencies
- **Wallet Module**: ChefToken balance management
- **User Module**: User profiles and levels
- **Recipe Module**: Personal recipe storage
- **Review Module**: Recipe ratings and reviews

### Used By
- Mobile app marketplace screen
- Web platform recipe store
- Leaderboard displays
- Seller dashboard analytics

## Testing

### Unit Tests
```go
func TestPurchaseRecipe(t *testing.T) {
    // Test cases:
    // - Successful purchase
    // - Insufficient funds
    // - Own recipe rejection
    // - Duplicate purchase prevention
    // - Commission calculation
}
```

### Integration Tests
```bash
# Test purchase flow
curl -X POST http://localhost:8080/api/marketplace/purchase \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"recipeId": "recipe-uuid"}'

# Verify balance updated
curl http://localhost:8080/api/wallet/balance \
  -H "Authorization: Bearer $JWT"
```

## Performance Considerations

### Database Indexes
Recommended indexes for optimal performance:
```sql
CREATE INDEX idx_personal_recipe_price ON PersonalRecipe(price);
CREATE INDEX idx_personal_recipe_category ON PersonalRecipe(category);
CREATE INDEX idx_recipe_purchase_buyer ON RecipePurchase(buyer_id);
CREATE INDEX idx_recipe_purchase_recipe ON RecipePurchase(recipe_id);
CREATE INDEX idx_wallet_transaction_user ON WalletTransaction(user_id);
```

### Query Optimization
- Leaderboard query uses JOINs efficiently
- Recipe filtering uses indexed columns
- Pagination limits result set size
- Aggregations computed in database

## Monitoring

### Key Metrics
- Total marketplace transactions
- Average purchase price
- Platform commission earned
- Top-selling recipes
- Conversion rate (views to purchases)
- Average seller revenue

### Logging
All operations logged with structured fields:
- User ID
- Recipe ID
- Transaction amount
- Commission amount
- Error messages

## Future Enhancements

### Planned Features
1. **Refund System**: 24-hour refund window
2. **Recipe Bundles**: Multi-recipe packages
3. **Subscription Model**: Monthly recipe access
4. **Seller Tiers**: Reduced commission for top sellers
5. **Promotional Sales**: Discounted pricing
6. **Gift System**: Buy recipes for others
7. **Review Integration**: Purchase verification for reviews
8. **Analytics Dashboard**: Detailed seller insights

### Technical Improvements
1. Caching for popular recipes
2. Search with Elasticsearch
3. Real-time purchase notifications
4. Event-driven architecture for transactions
5. Read replicas for leaderboard queries

## Examples

### Browse Marketplace
```bash
# Get Italian recipes under 1000 tokens
curl "http://localhost:8080/api/marketplace/recipes?category=italian&maxPrice=1000&sortBy=rating"
```

### Purchase Recipe
```bash
# Purchase a recipe
curl -X POST http://localhost:8080/api/marketplace/purchase \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "recipeId": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### View Leaderboard
```bash
# Get top sellers by revenue
curl "http://localhost:8080/api/marketplace/leaderboard?sortBy=revenue&limit=10"
```

### Check Seller Stats
```bash
# Get seller statistics
curl "http://localhost:8080/api/marketplace/stats/550e8400-e29b-41d4-a716-446655440000"
```

## Conclusion

The Marketplace module provides a complete e-commerce solution for recipe trading with:
- ✅ Secure transaction handling
- ✅ Platform commission management
- ✅ Comprehensive analytics
- ✅ Global leaderboard system
- ✅ Purchase history tracking
- ✅ Atomic wallet operations

Built with DDD principles and clean architecture for maintainability and scalability.
