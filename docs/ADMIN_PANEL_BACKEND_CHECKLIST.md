# 🎯 Admin Panel Backend - Minimal Checklist

## ✅ Реализовано

### 1. User Statistics API
- **Endpoint**: `GET /api/admin/users/stats`
- **Status**: ✅ **ГОТОВ**
- **Response**:
  ```json
  {
    "total": 54,
    "active_today": 0,
    "blocked": 0,
    "premium": 0
  }
  ```
- **Features**:
  - ✅ Использует `DATE_TRUNC` для стабильных цифр
  - ✅ FILTER clause для эффективности
  - ✅ Защищён authMiddleware + adminMiddleware

### 2. User List API
- **Endpoint**: `GET /api/admin/users`
- **Status**: ✅ **ГОТОВ (базовая версия)**
- **Response**:
  ```json
  {
    "users": [...],
    "meta": {
      "total": 54,
      "page": 1,
      "limit": 20,
      "totalPages": 3
    }
  }
  ```
- **Current Features**:
  - ✅ Пагинация (query: `?page=1&limit=20`)
  - ✅ Meta информация
  - ✅ Возвращает: id, name, email, role, status, lastLogin, createdAt

---

## ⏳ TODO: Обязательные фичи

### 3. User List Filters
- **Status**: ⏳ **TODO**
- **Required Query Params**:
  ```
  GET /api/admin/users?role=home_chef
  GET /api/admin/users?status=blocked
  GET /api/admin/users?search=john
  GET /api/admin/users?sort=lastLogin&order=desc
  ```
- **Implementation**:
  - [ ] Filter by role (`home_chef`, `admin`, `investor`)
  - [ ] Filter by status (`active`, `blocked`, `pending`)
  - [ ] Search by name/email
  - [ ] Sort by: name, email, createdAt, lastLogin
  - [ ] Order: asc/desc

### 4. Block/Unblock User
- **Status**: ⏳ **TODO**
- **Endpoints**:
  ```
  POST /api/admin/users/:id/block
  POST /api/admin/users/:id/unblock
  ```
- **Implementation**:
  - [ ] `BlockUser(userID)` - set status='blocked'
  - [ ] `UnblockUser(userID)` - set status='active'
  - [ ] Validation: cannot block yourself
  - [ ] Return updated user

### 5. Change User Role
- **Status**: ⏳ **TODO**
- **Endpoint**:
  ```
  PATCH /api/admin/users/:id/role
  Body: { "role": "admin" }
  ```
- **Implementation**:
  - [ ] `ChangeUserRole(userID, newRole)`
  - [ ] Update User.role
  - [ ] Sync to UserProfile.role
  - [ ] Validation: valid role enum
  - [ ] Return updated user

### 6. Audit Log Integration
- **Status**: ⏳ **TODO**
- **Endpoint**:
  ```
  GET /api/admin/users/:id/audit
  ```
- **Implementation**:
  - [ ] Link audit_log table to users
  - [ ] Track: block/unblock, role change, status change
  - [ ] Filter logs by user_id
  - [ ] Show: timestamp, admin_id, action, old_value, new_value

---

## 📋 Implementation Priority

### **Phase 1: Filters (высокий приоритет)**
Админ должен видеть:
- Заблокированных пользователей
- Пользователей по ролям
- Поиск по имени/email

**Effort**: 2-3 hours

### **Phase 2: Block/Unblock (критично)**
Админ должен иметь возможность блокировать пользователей

**Effort**: 1-2 hours

### **Phase 3: Role Change (средний приоритет)**
Админ должен менять роли (home_chef → admin)

**Effort**: 1-2 hours

### **Phase 4: Audit Log (low priority, но важно)**
Прозрачность действий админа

**Effort**: 3-4 hours

---

## 🔍 Текущая структура

### Existing Files
```
internal/modules/admin/
├── module.go              # Роуты
├── service/
│   └── service.go         # ✅ GetAllUsers, GetUsersStats
└── transport/http/
    └── handlers.go        # ✅ GetAllUsers, GetUsersStats
```

### Что нужно добавить

#### 1. Service Methods (`service/service.go`)
```go
type AdminService interface {
    // ✅ Existing
    GetAllUsers() ([]models.User, error)
    GetUsersStats() (map[string]interface{}, error)
    
    // ⏳ TODO
    GetUsersWithFilters(filters UserFilters, page, limit int) (*UserListResponse, error)
    BlockUser(userID string, adminID string) error
    UnblockUser(userID string, adminID string) error
    ChangeUserRole(userID string, newRole string, adminID string) error
    GetUserAuditLog(userID string) ([]AuditLogEntry, error)
}
```

#### 2. New DTOs (`dto/` or inline)
```go
type UserFilters struct {
    Role   *string
    Status *string
    Search *string
    SortBy string
    Order  string
}

type UserListResponse struct {
    Users []models.User         `json:"users"`
    Meta  PaginationMeta        `json:"meta"`
}

type PaginationMeta struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    TotalPages int `json:"totalPages"`
}

type AuditLogEntry struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    AdminID   string    `json:"adminId"`
    Action    string    `json:"action"`
    OldValue  string    `json:"oldValue,omitempty"`
    NewValue  string    `json:"newValue,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### 3. New Routes (`module.go`)
```go
// ✅ Existing
r.Get("/users", m.handlers.GetAllUsers)
r.Get("/users/stats", m.handlers.GetUsersStats)

// ⏳ TODO
r.Post("/users/{userID}/block", m.handlers.BlockUser)
r.Post("/users/{userID}/unblock", m.handlers.UnblockUser)
r.Patch("/users/{userID}/role", m.handlers.ChangeUserRole)
r.Get("/users/{userID}/audit", m.handlers.GetUserAuditLog)
```

---

## 🎯 Quick Start: Phase 1 (Filters)

### Step 1: Update GetAllUsers to support filters

**File**: `internal/modules/admin/service/service.go`

```go
func (s *adminService) GetUsersWithFilters(filters UserFilters, page, limit int) (*UserListResponse, error) {
    query := s.db.Model(&models.User{})
    
    // Apply filters
    if filters.Role != nil {
        query = query.Where("role = ?", *filters.Role)
    }
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.Search != nil {
        search := "%" + *filters.Search + "%"
        query = query.Where("name ILIKE ? OR email ILIKE ?", search, search)
    }
    
    // Count total
    var total int64
    query.Count(&total)
    
    // Apply sorting
    sortBy := "createdAt"
    if filters.SortBy != "" {
        sortBy = filters.SortBy
    }
    order := "DESC"
    if filters.Order != "" {
        order = strings.ToUpper(filters.Order)
    }
    query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
    
    // Apply pagination
    offset := (page - 1) * limit
    query = query.Offset(offset).Limit(limit)
    
    // Execute
    var users []models.User
    if err := query.Find(&users).Error; err != nil {
        return nil, err
    }
    
    return &UserListResponse{
        Users: users,
        Meta: PaginationMeta{
            Total:      int(total),
            Page:       page,
            Limit:      limit,
            TotalPages: int(math.Ceil(float64(total) / float64(limit))),
        },
    }, nil
}
```

### Step 2: Update Handler

**File**: `internal/modules/admin/transport/http/handlers.go`

```go
func (h *AdminHandlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
    // Parse filters from query params
    role := r.URL.Query().Get("role")
    status := r.URL.Query().Get("status")
    search := r.URL.Query().Get("search")
    sortBy := r.URL.Query().Get("sort")
    order := r.URL.Query().Get("order")
    
    // Parse pagination
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit < 1 || limit > 100 {
        limit = 20
    }
    
    filters := UserFilters{
        SortBy: sortBy,
        Order:  order,
    }
    if role != "" {
        filters.Role = &role
    }
    if status != "" {
        filters.Status = &status
    }
    if search != "" {
        filters.Search = &search
    }
    
    response, err := h.service.GetUsersWithFilters(filters, page, limit)
    if err != nil {
        http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

---

## 📊 Expected Timeline

| Phase | Feature | Effort | Priority |
|-------|---------|--------|----------|
| 1 | Filters + Search | 2-3h | 🔴 High |
| 2 | Block/Unblock | 1-2h | 🔴 High |
| 3 | Role Change | 1-2h | 🟡 Medium |
| 4 | Audit Log | 3-4h | 🟢 Low |
| **Total** | | **7-11h** | |

---

## 🚀 Next Steps

1. **Implement Phase 1** (filters) - most useful for admin
2. **Test with Postman/curl**
3. **Deploy to production**
4. **Frontend integration**
5. **Implement Phase 2-4** based on feedback

---

## 📝 Notes

- ✅ Authentication/Authorization уже работает (authMiddleware + adminMiddleware)
- ✅ Database schema готова (status, role, last_login)
- ✅ Базовая структура API готова
- ⏳ Нужно добавить только бизнес-логику фильтрации и управления

**Estimated total time to complete all phases: 1-2 days**
