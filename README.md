# 🍳 Menu Fodi Backend

Production backend for intelligent recipe recommendation system with fridge management.

## 📚 Documentation

### Core Architecture
- **[BACKEND_ARCHITECTURE_COMPLETE.md](./BACKEND_ARCHITECTURE_COMPLETE.md)** - Complete backend architecture, modules, and design patterns

### Recipe System
- **[RECIPE_RECOMMENDATION_MECHANISM.md](./RECIPE_RECOMMENDATION_MECHANISM.md)** - How recipes are matched to fridge ingredients
- **[RECIPE_RECOMMENDATION_VISUAL_GUIDE.md](./RECIPE_RECOMMENDATION_VISUAL_GUIDE.md)** - Visual diagrams and examples
- **[RECIPE_CATALOG_ENDPOINTS.md](./RECIPE_CATALOG_ENDPOINTS.md)** - All recipe catalog API endpoints

### Fridge System
- **[FRIDGE_API_DOCUMENTATION.md](./FRIDGE_API_DOCUMENTATION.md)** - Fridge management API reference

---

## 🚀 Quick Start

### Prerequisites
- Go 1.x
- PostgreSQL (Neon Cloud)
- Docker (optional)

### Build & Run

```bash
# Install dependencies
go mod download

# Build
make build

# Run locally
make run

# Run tests
make test
```

### Docker

```bash
docker build -t menu-fodi-backend .
docker run -p 8080:8080 -e DATABASE_URL=... menu-fodi-backend
```

---

## 🏗️ Project Structure

```
internal/
├── app/                    # Application setup
├── middleware/             # HTTP middleware
├── models/                 # Data models
└── modules/                # Feature modules
    ├── recipes/            # Recipe system
    ├── fridge/             # Fridge management
    ├── ai_recipe_recommendation/  # Recipe matching
    ├── menu/               # Kitchen pipeline
    ├── admin/              # Admin API
    └── ...

migrations/                # Database migrations
cmd/server/               # Main server entry point
pkg/utils/                # Shared utilities
```

---

## 🔌 Core APIs

### Recipe Recommendations
```
GET /api/recipe-recommendations?lang=ru&limit=10
Authorization: Bearer <JWT>
```
Matches recipes from catalog to user's fridge ingredients.

### Recipe Catalog
```
GET /api/recipes?category=soup&maxTime=60&lang=ru
```
Browse all recipes with filters.

### Fridge Management
```
GET /api/fridge
POST /api/fridge/items
DELETE /api/fridge/items/{id}
```
Manage user's fridge inventory.

---

## 🗄️ Database

PostgreSQL on Neon Cloud with GORM ORM.

**Key tables:**
- `"User"` - User accounts
- `"RecipeCatalog"` - Recipe database
- `"Ingredient"` - Ingredient master data
- `user_fridge_items` - User's fridge inventory
- `"UserMenuItem"` - Kitchen pipeline (cooking workflow)

---

## 🔐 Authentication

JWT-based authentication with role support:
- `home_chef` - Regular user
- `admin` - Admin user
- `super_admin` - System admin

```
Authorization: Bearer eyJhbGc...
```

---

## 📊 Key Features

✅ **Smart Recipe Matching** - Rules engine (not AI) for matching recipes to fridge  
✅ **Canonical Ingredients** - Group similar ingredients (e.g., all oils)  
✅ **Multilingual** - Support for Polish, English, Russian  
✅ **Kitchen Pipeline** - Track cooking workflow (planned → cooking → completed)  
✅ **Admin System** - Full admin API for recipe management  
✅ **Notifications** - Event-based notification system  

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test file
go test ./internal/modules/recipes/...
```

---

## 🚢 Deployment

### Production (Koyeb)
Automatic deployment from `main` branch to https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app

### Environment Variables
```
DATABASE_URL=postgresql://...
JWT_SECRET=...
ENVIRONMENT=production
```

---

## 📖 API Documentation

For detailed API documentation, see:
- [Recipe Catalog Endpoints](./RECIPE_CATALOG_ENDPOINTS.md)
- [Recipe Recommendation Mechanism](./RECIPE_RECOMMENDATION_MECHANISM.md)
- [Fridge API](./FRIDGE_API_DOCUMENTATION.md)

---

## 🤝 Contributing

1. Create feature branch: `git checkout -b feature/name`
2. Make changes and test
3. Commit: `git commit -am "feat: description"`
4. Push and create PR

---

## 📝 License

Internal use only.

---

## 👥 Team

- Backend: Dmitrij Fomin
- Frontend: (https://dima-fomin.pl)

---

## 📞 Support

For issues or questions, check the documentation or contact the backend team.

