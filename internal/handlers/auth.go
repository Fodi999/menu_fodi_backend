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
