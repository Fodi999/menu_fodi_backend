# Chef Academy Backend

A modern Go backend service for the Chef Academy platform with AI-powered culinary features.

## 🏗️ Architecture

The backend follows **Domain-Driven Design (DDD)** principles with a modular, microservice-friendly architecture:

```
internal/
├── modules/           # 20 independent domain modules
│   ├── ai/           # AI assistant features
│   ├── ai_core/      # Core AI engine (Groq LLM integration)
│   ├── academy/      # Online courses and certificates
│   ├── auth/         # Authentication & JWT
│   ├── business/     # Business management
│   ├── fridge/       # Smart fridge inventory
│   ├── marketplace/  # Recipe marketplace
│   ├── nutrition/    # Nutrition calculation
│   ├── recipes/      # Social recipe features
│   └── [16 more...]
├── app/              # Application setup
├── middleware/       # HTTP middleware (Auth, Admin)
├── database/         # Database layer & repositories
└── platform/         # Shared utilities & config
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL / SQLite
- Groq API Key (for AI features)

### Setup

1. **Clone and install dependencies:**
   ```bash
   go mod download
   ```

2. **Set environment variables:**
   ```bash
   export DATABASE_URL="postgresql://user:pass@localhost/chef_academy"
   export JWT_SECRET="your-secret-key"
   export GROQ_API_KEY="your-groq-api-key"
   export HTTP_PORT="8080"
   export ENV="development"
   ```

3. **Run migrations:**
   ```bash
   go run cmd/migrate/main.go
   ```

4. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

Server runs on `http://localhost:8080`

## 📚 API Documentation

See [ROUTES_DOCUMENTATION.md](./ROUTES_DOCUMENTATION.md) for complete API route reference.

### Key Endpoints

**Authentication:**
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user

**AI Features:**
- `POST /api/ai/chef-mentor` - Chat with AI chef mentor
- `POST /api/ai/recipe-generator` - Generate recipes
- `POST /api/ai/meal-plan` - Generate meal plans (Protected)

**User:**
- `GET /api/user/profile` - Get user profile (Protected)
- `GET /api/user/achievements` - Get user achievements (Protected)

**Admin:**
- `GET /api/admin/stats` - Get admin statistics (Admin only)
- `GET /api/admin/users` - List all users (Admin only)

## 🏗️ Module Structure

Each module contains:

```
module_name/
├── module.go              # Module initialization
├── transport/http/        # HTTP handlers
├── service/               # Business logic
├── repo/                  # Data repositories
└── dto/                   # Data transfer objects
```

### Example: Academy Module

```go
// Register routes
academyModule := academy.NewModule(db)
r.Route("/api", func(r chi.Router) {
    academyModule.RegisterRoutes(r, middleware.AuthMiddleware)
})

// Available endpoints
GET    /api/academy/courses
GET    /api/academy/courses/{courseId}
POST   /api/academy/enroll (Protected)
POST   /api/academy/certificates/generate (Protected)
```

## 🤖 AI Features

### Chef Mentor (Groq LLama 3-70B)
- Interactive culinary consulting
- Recipe recommendations
- Cooking techniques guidance
- Multi-language support (Polish, Ukrainian, English)

### Recipe Analysis
- Taste balance assessment
- Difficulty level estimation
- Price estimation
- Allergen detection
- Nutritional recommendations

### Meal Planning
- AI-generated meal plans
- Calorie tracking
- Personalized recommendations
- Smart fridge integration

## 🗄️ Database

PostgreSQL with automatic migrations:

```sql
-- Key tables
users                    -- User accounts & profiles
user_profiles           -- Extended user data
courses                 -- Academy courses
recipes                 -- Social recipes
user_fridge            -- Smart fridge inventory
chef_mentor_sessions   -- AI conversation history
```

## 🔒 Security

- **JWT Authentication**: Stateless token-based auth
- **Password Hashing**: bcrypt with 12 cost factor
- **CORS**: Configurable cross-origin access
- **Admin Middleware**: Role-based access control

## 📊 Project Statistics

- **20 Domain Modules** - Clean separation of concerns
- **~5000+ Lines of Code** - Focused and maintainable
- **100% Type-Safe** - Go's strong typing
- **DDD Architecture** - Domain-driven design principles
- **Groq AI Integration** - Latest LLM capabilities

## 🛠️ Development

### Build
```bash
go build ./cmd/server/
```

### Test
```bash
go test ./...
```

### Database Seed
```bash
go run cmd/migrate/main.go  # Run migrations
```

## 📝 Key Files

- `cmd/server/main.go` - Application entry point
- `internal/app/server.go` - Server initialization
- `internal/app/routes_modular.go` - Route registration
- `internal/middleware/auth.go` - Authentication middleware
- `migrations/` - Database migrations

## 🔄 Recent Refactoring

✅ **DDD Migration Complete**
- ✅ 7 heavy handlers extracted to modules
- ✅ JWT moved to auth module service layer
- ✅ AI core separated into `ai_core` module
- ✅ All shared services migrated to module ownership
- ✅ Old flat handler structure removed
- ✅ Modular routes as primary routing system

## 📦 Dependencies

Key packages:
- `github.com/go-chi/chi/v5` - HTTP router
- `gorm.io/gorm` - ORM
- `github.com/golang-jwt/jwt/v5` - JWT
- `github.com/google/uuid` - UUID generation
- `golang.org/x/crypto/bcrypt` - Password hashing

## 🐛 Troubleshooting

**"database connection failed"**
- Check DATABASE_URL environment variable
- Ensure PostgreSQL is running
- Run migrations: `go run cmd/migrate/main.go`

**"AI requests failing"**
- Verify GROQ_API_KEY is set
- Check internet connection
- Test with `curl -H "Authorization: Bearer $GROQ_API_KEY" https://api.groq.com/health`

**"JWT token invalid"**
- Ensure JWT_SECRET is consistent
- Check token expiration (24 hours)
- Verify Authorization header format: `Bearer <token>`

## 📖 Further Reading

- [Routes Documentation](./ROUTES_DOCUMENTATION.md) - Complete API reference
- [DDD Pattern](https://en.wikipedia.org/wiki/Domain-driven_design) - Architecture pattern
- [Go Best Practices](https://golang.org/doc/effective_go)
- [Chi Router Docs](https://github.com/go-chi/chi)

## 📞 Support

For issues or questions, please check:
1. Environment variables are correctly set
2. Database is running and accessible
3. API keys (Groq) are valid
4. Go version is 1.21 or higher

## 📄 License

All rights reserved © 2024
