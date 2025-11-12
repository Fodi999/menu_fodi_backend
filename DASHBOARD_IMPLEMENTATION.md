# 🏠 Dashboard Implementation - Admin vs User

Полное описание реализации личных кабинетов для администратора и обычного пользователя на бекенде.

---

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Admin Dashboard](#admin-dashboard)
3. [User Dashboard](#user-dashboard)
4. [Authentication & Authorization](#authentication--authorization)
5. [Data Flow](#data-flow)
6. [Comparison](#comparison)
7. [Database Structure](#database-structure)

---

## Architecture Overview

### Project Structure

```
internal/modules/
├── admin/                          # Admin panel module
│   ├── service/
│   │   ├── service.go             # AdminService interface & implementation
│   │   └── policy.go              # RBAC policy (access control)
│   └── transport/http/
│       └── handlers.go            # HTTP handlers for admin endpoints
│
└── user/                           # User dashboard module
    ├── service/
    │   └── service.go             # UserService interface & implementation
    ├── repo/
    │   └── repository.go          # Data access layer
    ├── dto/
    │   └── dtos.go                # Data transfer objects
    └── transport/http/
        └── handlers.go            # HTTP handlers for user endpoints
```

---

## Admin Dashboard

### 1️⃣ Architecture Layers

```
┌─────────────────────────────────────────┐
│  HTTP Layer (handlers.go)               │
│  - GetAdminProfile                      │
│  - GetAllUsers, UpdateUser, DeleteUser  │
│  - GetAllOrders, UpdateOrderStatus      │
│  - GetAllTokenBanks, AllocateTokens     │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Service Layer (service.go)             │
│  - AdminService interface               │
│  - Business logic & validations         │
│  - Permission checks via AdminPolicy    │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Repository Layer (db layer)            │
│  - UserRepository                       │
│  - OrderRepository                      │
│  - TokenBankRepository                  │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Database (PostgreSQL)                  │
│  - User table                           │
│  - Order table                          │
│  - token_bank table                     │
└─────────────────────────────────────────┘
```

### 2️⃣ Key Components

#### AdminService Interface

```go
type AdminService interface {
    // Users Management
    GetAllUsers() ([]models.User, error)
    UpdateUser(userID string, name, email string) (*models.User, error)
    DeleteUser(userID string) error
    UpdateUserRole(userID, role string) error

    // Orders Management
    GetAllOrders() ([]models.Order, error)
    GetRecentOrders(limit int) ([]models.Order, error)
    UpdateOrderStatus(orderID, status string) error

    // Statistics
    GetAdminStats() (map[string]interface{}, error)

    // Admin Profile
    GetAdminProfile(adminID string) (map[string]interface{}, error)

    // Token Bank
    GetAllTokenBanks() ([]models.TokenBank, error)
    AllocateTokens(userID string, amount int64) error
    RevokeTokens(userID string, amount int64) error
    // ... more token bank methods
}
```

#### AdminPolicy (RBAC)

```go
type AdminPolicy interface {
    CanViewUsers(admin *models.User) bool
    CanUpdateUser(admin *models.User) bool
    CanDeleteUser(admin *models.User) bool
    CanManageOrders(admin *models.User) bool
    CanManageTokens(admin *models.User) bool
}

// Implementation
func (p *adminPolicy) CanViewUsers(admin *models.User) bool {
    return admin.Role == "admin"
}
```

### 3️⃣ GetAdminProfile Endpoint

**Endpoint:** `GET /api/admin/profile`

**Handler Implementation:**

```go
func (h *AdminHandlers) GetAdminProfile(w http.ResponseWriter, r *http.Request) {
    // 1. Extract admin ID from JWT context
    adminID := r.Context().Value("user_id")
    if adminID == nil {
        utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    adminIDStr, ok := adminID.(string)
    if !ok {
        utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user ID format")
        return
    }

    // 2. Call service layer
    profile, err := h.service.GetAdminProfile(adminIDStr)
    if err != nil {
        if err.Error() == "admin not found" {
            utils.RespondWithError(w, http.StatusNotFound, "Admin not found")
        } else {
            utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch admin profile")
        }
        return
    }

    // 3. Return response
    utils.RespondWithJSON(w, http.StatusOK, profile)
}
```

**Service Implementation:**

```go
func (s *adminService) GetAdminProfile(adminID string) (map[string]interface{}, error) {
    // 1. Get admin user data
    var user models.User
    if err := s.db.First(&user, "id = ?", adminID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("admin not found")
        }
        return nil, err
    }

    // 2. Count total users managed
    var userCount int64
    if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
        return nil, err
    }

    // 3. Count total orders managed
    var orderCount int64
    if err := s.db.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
        return nil, err
    }

    // 4. Return admin profile with statistics
    return map[string]interface{}{
        "id":              user.ID,
        "name":            user.Name,
        "email":           user.Email,
        "role":            user.Role,
        "createdAt":       user.CreatedAt,
        "managedUsers":    userCount,      // Total users in system
        "managedOrders":   orderCount,     // Total orders in system
        "totalStats": map[string]interface{}{
            "users":  userCount,
            "orders": orderCount,
        },
    }, nil
}
```

**Response Example:**

```json
{
  "id": "admin-uuid-123",
  "name": "John Admin",
  "email": "admin@example.com",
  "role": "admin",
  "createdAt": "2024-01-01T00:00:00Z",
  "managedUsers": 156,
  "managedOrders": 2340,
  "totalStats": {
    "users": 156,
    "orders": 2340
  }
}
```

### 4️⃣ Admin Dashboard Data

Admin личный кабинет показывает:

- ✅ **Admin Info:** ID, Name, Email, Role, Created Date
- ✅ **Managed Resources:**
  - Total users in system
  - Total orders in system
- ✅ **System Statistics:**
  - User growth metrics
  - Order completion rates
  - Revenue information (if applicable)
- ✅ **Token Bank Stats:**
  - Total tokens allocated
  - Total tokens used
  - User distribution
- ✅ **Quick Actions:**
  - View all users (GET /api/admin/users)
  - View all orders (GET /api/admin/orders)
  - Manage token banks (GET /api/admin/token-bank)

---

## User Dashboard

### 1️⃣ Architecture Layers

```
┌─────────────────────────────────────────┐
│  HTTP Layer (handlers.go)               │
│  - GetProfile, UpdateProfile            │
│  - GetProgress, GetStats                │
│  - GetCourses, GetWallet                │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Service Layer (service.go)             │
│  - UserService interface                │
│  - Profile management                   │
│  - Progress tracking                    │
│  - Stats aggregation                    │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Repository Layer (repo/)               │
│  - UserRepository                       │
│  - CourseRepository                     │
│  - WalletRepository                     │
│  - ProgressRepository                   │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│  Database (PostgreSQL)                  │
│  - User table                           │
│  - UserProfile table                    │
│  - UserCourse table                     │
│  - UserProgress table                   │
│  - WalletTransaction table              │
└─────────────────────────────────────────┘
```

### 2️⃣ Key Components

#### UserService Interface

```go
type UserService interface {
    // Profile Management
    GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error)
    UpdateProfile(userID uuid.UUID, req dto.UpdateProfileRequest) error
    UploadAvatar(userID uuid.UUID, file []byte) (string, error)

    // Progress & Stats
    GetProgress(userID uuid.UUID) (*dto.UserProgressResponse, error)
    GetStats(userID uuid.UUID) (*dto.UserStatsResponse, error)

    // Courses
    GetEnrolledCourses(userID uuid.UUID, limit, offset int) ([]dto.CourseResponse, error)

    // Wallet
    GetWallet(userID uuid.UUID) (*dto.WalletResponse, error)
}
```

### 3️⃣ GetProfile Endpoint

**Endpoint:** `GET /api/user/profile`

**Handler Implementation:**

```go
func (h *UserHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
    // 1. Extract user ID from JWT context
    userIDPtr := middleware.GetUserID(r)
    if userIDPtr == nil {
        logger.Error("user ID not found in context")
        httpx.Unauthorized(w, "unauthorized")
        return
    }
    userID := *userIDPtr

    // 2. Call service layer
    profile, err := h.service.GetProfile(userID)
    if err != nil {
        logger.Error("failed to get profile",
            zap.Error(err),
            zap.String("user_id", userID.String()))
        httpx.InternalError(w, "failed to get profile")
        return
    }

    // 3. Return response
    httpx.Success(w, profile)
}
```

**Service Implementation:**

```go
func (s *userService) GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error) {
    // 1. Get user data
    user, err := s.repo.GetUser(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // 2. Get user profile
    profile, err := s.repo.GetProfile(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get profile: %w", err)
    }

    // 3. Get completed courses count
    coursesCount, err := s.repo.GetCompletedCoursesCount(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to count courses: %w", err)
    }

    // 4. Get wallet balance
    wallet, err := s.repo.GetWallet(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get wallet: %w", err)
    }

    // 5. Build response
    return &dto.UserProfileResponse{
        UserID:           user.ID.String(),
        Name:             user.Name,
        Email:            user.Email,
        Level:            profile.Level,
        Stars:            profile.Stars,
        XP:               profile.XP,
        Role:             user.Role,
        Language:         profile.Language,
        AvatarURL:        profile.AvatarURL,
        CompletedCourses: coursesCount,
        WalletBalance:    wallet.Balance,
        CreatedAt:        user.CreatedAt,
        UpdatedAt:        user.UpdatedAt,
    }, nil
}
```

**DTO Definition:**

```go
type UserProfileResponse struct {
    UserID           string    `json:"userId"`
    Name             string    `json:"name"`
    Email            string    `json:"email"`
    Level            int       `json:"level"`
    Stars            int       `json:"stars"`
    XP               int       `json:"xp"`
    Role             string    `json:"role"`
    Language         string    `json:"language"`
    AvatarURL        string    `json:"avatarUrl"`
    CompletedCourses int       `json:"completedCourses"`
    WalletBalance    int64     `json:"walletBalance"`
    CreatedAt        time.Time `json:"createdAt"`
    UpdatedAt        time.Time `json:"updatedAt"`
}
```

**Response Example:**

```json
{
  "userId": "user-uuid-456",
  "name": "John Doe",
  "email": "user@example.com",
  "level": 5,
  "stars": 150,
  "xp": 2500,
  "role": "user",
  "language": "en",
  "avatarUrl": "https://example.com/avatars/user123.jpg",
  "completedCourses": 3,
  "walletBalance": 1500,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-11-10T15:45:00Z"
}
```

### 4️⃣ User Dashboard Data

User личный кабинет показывает:

- ✅ **User Info:** Name, Email, Role, Avatar, Language
- ✅ **Learning Progress:**
  - Current level
  - XP points earned
  - Stars collected
  - Courses completed
- ✅ **Personal Statistics:**
  - Study hours
  - Quiz scores
  - Streak days
  - Achievements & badges
- ✅ **Financial:**
  - Wallet balance
  - Token bank balance
  - Transaction history
- ✅ **Quick Actions:**
  - Update profile
  - Upload avatar
  - View courses
  - Check progress

---

## Authentication & Authorization

### Admin Authentication Flow

```
1. Admin logs in
   POST /api/auth/login
   {email, password} → JWT token

2. Admin requests profile
   GET /api/admin/profile
   Headers: Authorization: Bearer {JWT}

3. Middleware validates JWT
   ├─ Verify signature
   ├─ Check expiration
   ├─ Extract user_id
   └─ Store in context

4. AdminMiddleware checks role
   ├─ Get user from DB
   ├─ Check role == "admin"
   └─ Allow request or return 403

5. Handler processes request
   ├─ Get user_id from context
   ├─ Call service.GetAdminProfile()
   └─ Return profile data
```

### User Authentication Flow

```
1. User logs in
   POST /api/auth/login
   {email, password} → JWT token

2. User requests profile
   GET /api/user/profile
   Headers: Authorization: Bearer {JWT}

3. Middleware validates JWT
   ├─ Verify signature
   ├─ Check expiration
   ├─ Extract user_id
   └─ Store in context

4. Handler processes request
   ├─ Get user_id from context
   ├─ Call service.GetProfile()
   ├─ Aggregate user data
   └─ Return profile data
```

### Middleware Stack

```
Request
  ↓
┌─────────────────────────────────┐
│ Request ID Middleware           │
│ (Adds unique request ID)        │
└──────────────┬──────────────────┘
               ↓
┌─────────────────────────────────┐
│ CORS Middleware                 │
│ (Handles cross-origin requests) │
└──────────────┬──────────────────┘
               ↓
┌─────────────────────────────────┐
│ Auth Middleware                 │
│ (Validates JWT token)           │
│ (Extracts user_id to context)   │
└──────────────┬──────────────────┘
               ↓
┌─────────────────────────────────┐
│ Admin Middleware (for /api/admin) │
│ (Checks role == "admin")        │
│ (Returns 403 if not admin)      │
└──────────────┬──────────────────┘
               ↓
         Handler
```

---

## Data Flow

### Admin Profile Request Flow

```
┌─────────────────────────────────────────────────────┐
│ Client Request                                      │
│ GET /api/admin/profile                              │
│ Headers: Authorization: Bearer {JWT}                │
└──────────────────┬──────────────────────────────────┘
                   │
        ┌──────────▼──────────┐
        │ Auth Middleware     │
        │ - Validate JWT      │
        │ - Extract user_id   │
        │ - Store in context  │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ Admin Middleware    │
        │ - Check role admin  │
        │ - Return 403 if not │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────────────────┐
        │ GetAdminProfile Handler         │
        │ - Get admin_id from context     │
        │ - Validate permissions (policy) │
        └──────────┬─────────────────────┘
                   │
        ┌──────────▼──────────────────────┐
        │ AdminService.GetAdminProfile()  │
        │ - Get user data                 │
        │ - Count users (SELECT COUNT)    │
        │ - Count orders (SELECT COUNT)   │
        │ - Build response                │
        └──────────┬─────────────────────┘
                   │
        ┌──────────▼──────────┐
        │ Database Query      │
        │ - Get User record   │
        │ - Count User rows   │
        │ - Count Order rows  │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────┐
        │ Return Response     │
        │ {                   │
        │   id, name, email,  │
        │   managedUsers,     │
        │   managedOrders     │
        │ }                   │
        └─────────────────────┘
```

### User Profile Request Flow

```
┌─────────────────────────────────────────────────────┐
│ Client Request                                      │
│ GET /api/user/profile                               │
│ Headers: Authorization: Bearer {JWT}                │
└──────────────────┬──────────────────────────────────┘
                   │
        ┌──────────▼──────────┐
        │ Auth Middleware     │
        │ - Validate JWT      │
        │ - Extract user_id   │
        │ - Store in context  │
        └──────────┬──────────┘
                   │
        ┌──────────▼──────────────────────┐
        │ GetProfile Handler              │
        │ - Get user_id from context      │
        │ - Call service.GetProfile()     │
        └──────────┬─────────────────────┘
                   │
        ┌──────────▼──────────────────────┐
        │ UserService.GetProfile()        │
        │ - Get user data                 │
        │ - Get profile record            │
        │ - Count completed courses       │
        │ - Get wallet balance            │
        │ - Aggregate data                │
        └──────────┬─────────────────────┘
                   │
        ┌──────────▼────────────────────────────┐
        │ Repository Queries                   │
        │ - SELECT * FROM User WHERE id = ?    │
        │ - SELECT * FROM UserProfile WHERE id │
        │ - SELECT COUNT(*) FROM UserCourse    │
        │ - SELECT balance FROM wallet         │
        └──────────┬──────────────────────────┘
                   │
        ┌──────────▼──────────┐
        │ Return Response     │
        │ {                   │
        │   name, email,      │
        │   level, stars, xp, │
        │   avatar, courses,  │
        │   wallet            │
        │ }                   │
        └─────────────────────┘
```

---

## Comparison

| Feature | Admin Dashboard | User Dashboard |
|---------|-----------------|----------------|
| **Endpoint** | `GET /api/admin/profile` | `GET /api/user/profile` |
| **Access Control** | Role == "admin" | Authenticated only |
| **Middleware** | Auth + Admin + Policy | Auth only |
| **Data Source** | User + Order + Stats | User + Profile + Courses + Wallet |
| **Query Count** | 3 queries (User, Count Users, Count Orders) | 4+ queries (User, Profile, Courses, Wallet) |
| **Cache Needed** | Yes (stats can be cached) | Yes (profile can be cached) |
| **Response Size** | Small (~200 bytes) | Medium (~500 bytes) |
| **Update Frequency** | Low (stats don't change often) | High (progress updates frequently) |
| **Critical Data** | System statistics | Personal progress & wallet |

---

## Database Structure

### User Table

```sql
CREATE TABLE "User" (
  id UUID PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,  -- bcrypt hashed
  role VARCHAR(50) DEFAULT 'user', -- 'user' or 'admin'
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### UserProfile Table

```sql
CREATE TABLE user_profile (
  id UUID PRIMARY KEY,
  user_id UUID UNIQUE NOT NULL REFERENCES "User"(id),
  level INT DEFAULT 1,
  stars INT DEFAULT 0,
  xp INT DEFAULT 0,
  language VARCHAR(10) DEFAULT 'en',
  avatar_url VARCHAR(1000),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Order Table (for Admin)

```sql
CREATE TABLE order (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES "User"(id),
  status VARCHAR(50) DEFAULT 'pending',
  total_price DECIMAL(10, 2),
  items_count INT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Token Bank Table (for Admin)

```sql
CREATE TABLE token_bank (
  id UUID PRIMARY KEY,
  user_id UUID UNIQUE NOT NULL REFERENCES "User"(id),
  balance BIGINT DEFAULT 0,
  total_allocated BIGINT DEFAULT 0,
  total_used BIGINT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## Key Differences Summary

### Admin Dashboard

```
Purpose: System Management
Scope: View all users, orders, tokens
Access: Only admins
Data: Aggregated system statistics
Updates: Infrequent (manual admin actions)
Performance: Optimized for read-heavy operations
Cache: Yes (stats cache recommended)
```

### User Dashboard

```
Purpose: Personal Learning Portal
Scope: View own profile, progress, courses
Access: All authenticated users
Data: Personal user information & progress
Updates: Frequent (user activity updates)
Performance: Optimized for personalized data
Cache: Yes (profile cache with short TTL)
```

---

## Implementation Checklist

### Admin Dashboard
- ✅ AdminService interface defined
- ✅ GetAdminProfile() method implemented
- ✅ AdminMiddleware for role checking
- ✅ AdminPolicy for RBAC
- ✅ Handlers with proper error handling
- ✅ Documentation with examples
- ✅ Tests for RBAC scenarios

### User Dashboard
- ✅ UserService interface defined
- ✅ GetProfile() method implemented
- ✅ DTOs for type-safe responses
- ✅ Repository pattern for data access
- ✅ Handlers with proper error handling
- ✅ Update profile functionality
- ✅ Avatar upload functionality
- ✅ Progress & stats endpoints

---

## Best Practices Applied

1. **Separation of Concerns:**
   - HTTP layer (handlers) separate from business logic (service)
   - Service layer separate from data access (repository)

2. **Dependency Injection:**
   - Services injected into handlers
   - Repositories injected into services
   - Database connection passed to repositories

3. **Error Handling:**
   - Proper HTTP status codes
   - Descriptive error messages
   - Logging with context (zap logger)

4. **Security:**
   - JWT token validation
   - Role-based access control (RBAC)
   - Password hashing (bcrypt)
   - Middleware for authentication

5. **Performance:**
   - Efficient database queries
   - Connection pooling
   - Response caching (where applicable)
   - Index optimization on frequently queried columns

6. **Code Quality:**
   - Type-safe DTOs
   - Interface-based design
   - Unit testable architecture
   - Comprehensive error handling
