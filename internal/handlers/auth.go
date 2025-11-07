package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/auth"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var userRepo = &database.UserRepository{}

// Register обработчик регистрации
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Проверка на существующего пользователя
	existingUser, _ := userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		utils.RespondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Server error")
		return
	}

	// Создание пользователя
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	// Сохранение в БД
	if err := userRepo.Create(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// 🎁 Нарахування початкових 100 токенів Chef при реєстрації
	if err := grantWelcomeTokens(user.ID); err != nil {
		// Логуємо помилку, але не блокуємо реєстрацію
		// log.Printf("Warning: Failed to grant welcome tokens to user %s: %v", user.ID, err)
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Server error")
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// Login обработчик входа
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Поиск пользователя в БД
	user, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Server error")
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// GetProfile получение профиля пользователя
func GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := userRepo.FindByID(claims.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// UpdateProfile обновление профиля
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := userRepo.FindByID(claims.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Обновление данных
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Сохранение в БД
	if err := userRepo.Update(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// VerifyTokenRequest структура запроса для верификации токена
type VerifyTokenRequest struct {
	Token string `json:"token"`
}

// VerifyTokenResponse структура ответа верификации токена
type VerifyTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
}

// VerifyTokenHandler обработчик верификации JWT токена
func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Валидация токена
	claims, err := auth.ValidateToken(req.Token)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusOK, VerifyTokenResponse{Valid: false})
		return
	}

	// 🧠 Попробуем найти пользователя в базе
	user, err := userRepo.FindByID(claims.UserID)
	if err != nil {
		// Если не нашли — всё равно возвращаем валидный токен, но без имени/email
		utils.RespondWithJSON(w, http.StatusOK, VerifyTokenResponse{
			Valid:  true,
			UserID: claims.UserID,
			Role:   claims.Role,
		})
		return
	}

	// ✅ Возвращаем всю информацию
	utils.RespondWithJSON(w, http.StatusOK, VerifyTokenResponse{
		Valid:  true,
		UserID: user.ID,
		Role:   user.Role,
		Name:   user.Name,
		Email:  user.Email,
	})
}

// UpdateRoleRequest структура запроса для обновления роли
type UpdateRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// UpdateUserRole обработчик обновления роли пользователя
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация роли
	if req.Role != "user" && req.Role != "admin" && req.Role != "business_owner" && req.Role != "investor" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid role. Must be: user, admin, business_owner, or investor")
		return
	}

	// Получить текущего пользователя из контекста (проверка прав)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверка прав: только админ может менять роли
	if claims.Role != "admin" {
		utils.RespondWithError(w, http.StatusForbidden, "Only admins can update user roles")
		return
	}

	// Найти пользователя по ID
	user, err := userRepo.FindByID(req.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Обновить роль
	oldRole := user.Role
	user.Role = req.Role

	if err := userRepo.Update(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user role")
		return
	}

	// Вернуть успешный ответ
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "✅ User role updated successfully",
		"user_id":    user.ID,
		"old_role":   oldRole,
		"new_role":   user.Role,
		"name":       user.Name,
		"email":      user.Email,
		"updated_by": claims.UserID,
	})
}

// grantWelcomeTokens нараховує початкові 100 токенів Chef новому користувачеві
func grantWelcomeTokens(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// Перевірка чи існує профіль користувача
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userUUID).First(&profile).Error; err != nil {
		// Якщо профілю немає, створюємо його
		profile = models.UserProfile{
			UserID:        userUUID,
			Name:          "", // буде оновлено пізніше
			Email:         "", // буде оновлено пізніше
			WalletBalance: 100.00,
		}
		if err := database.DB.Create(&profile).Error; err != nil {
			return err
		}
	} else {
		// Якщо профіль є, додаємо токени
		profile.WalletBalance += 100.00
		if err := database.DB.Save(&profile).Error; err != nil {
			return err
		}
	}

	// Створити транзакцію для історії
	transaction := models.WalletTransaction{
		UserID:      userUUID,
		Amount:      100.00,
		Type:        "bonus",
		Description: "🎁 Вітальний бонус! Ласкаво просимо до Chef Academy!",
	}

	return database.DB.Create(&transaction).Error
}
