package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/services"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetCourses возвращает список всех опубликованных курсов
func GetCourses(w http.ResponseWriter, r *http.Request) {
	var courses []models.Course
	query := database.DB.Where("is_published = ?", true).Order("level ASC, created_at DESC")

	// Фильтр по языку
	if lang := r.URL.Query().Get("language"); lang != "" {
		query = query.Where("language = ?", lang)
	}

	// Фильтр по категории
	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&courses).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, courses)
}

// GetCourse возвращает детали курса
func GetCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["courseId"]

	var course models.Course
	if err := database.DB.Where("id = ?", courseID).First(&course).Error; err != nil {
		utils.RespondError(w, http.StatusNotFound, "Course not found", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, course)
}

// GetCourseLessons возвращает уроки курса
func GetCourseLessons(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["courseId"]

	var lessons []models.Lesson
	if err := database.DB.Where("course_id = ? AND is_published = ?", courseID, true).Order("\"order\" ASC").Find(&lessons).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, lessons)
}

// GetQuiz возвращает случайные вопросы для теста
func GetQuiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["courseId"]

	// Получаем все вопросы для курса
	var allQuestions []models.QuizQuestion
	if err := database.DB.Where("course_id = ?", courseID).Find(&allQuestions).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if len(allQuestions) == 0 {
		utils.RespondError(w, http.StatusNotFound, "No quiz questions found", "This course has no quiz yet")
		return
	}

	// Выбираем случайные 10 вопросов (или меньше, если вопросов меньше 10)
	questionCount := 10
	if len(allQuestions) < questionCount {
		questionCount = len(allQuestions)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allQuestions), func(i, j int) {
		allQuestions[i], allQuestions[j] = allQuestions[j], allQuestions[i]
	})

	selectedQuestions := allQuestions[:questionCount]

	utils.RespondJSON(w, http.StatusOK, selectedQuestions)
}

// SubmitQuiz принимает ответы на тест и возвращает результат
func SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["courseId"]

	var input struct {
		UserID  string  `json:"userId"`
		Answers []int32 `json:"answers"` // индексы выбранных ответов
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Получаем вопросы теста
	var questions []models.QuizQuestion
	if err := database.DB.Where("course_id = ?", courseID).Find(&questions).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if len(questions) == 0 {
		utils.RespondError(w, http.StatusNotFound, "No quiz found", "This course has no quiz")
		return
	}

	// Подсчитываем правильные ответы
	correctCount := 0
	for i, answer := range input.Answers {
		if i < len(questions) && int(answer) == questions[i].CorrectAnswer {
			correctCount++
		}
	}

	totalQuestions := len(input.Answers)
	score := (correctCount * 100) / totalQuestions

	// Рассчитываем звёзды (0-5)
	stars := 0
	if score >= 90 {
		stars = 5
	} else if score >= 80 {
		stars = 4
	} else if score >= 70 {
		stars = 3
	} else if score >= 60 {
		stars = 2
	} else if score >= 50 {
		stars = 1
	}

	// Сохраняем результат
	parsedUserID, _ := uuid.Parse(input.UserID)
	parsedCourseID, _ := uuid.Parse(courseID)

	userQuiz := models.UserQuiz{
		UserID:         parsedUserID,
		CourseID:       parsedCourseID,
		Score:          score,
		TotalQuestions: totalQuestions,
		CorrectAnswers: correctCount,
		Answers:        input.Answers,
		StarsEarned:    stars,
		CompletedAt:    time.Now(),
	}

	if err := database.DB.Create(&userQuiz).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to save quiz result", err.Error())
		return
	}

	// Обновляем профиль ученика
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", input.UserID).First(&profile).Error; err == nil {
		profile.Stars += stars
		profile.XP += score // XP = процент правильных ответов
		database.DB.Save(&profile)

		// Награда в ChefToken (1 звезда = 10 токенов)
		if stars > 0 {
			reward := float64(stars * 10)
			profile.WalletBalance += reward
			database.DB.Save(&profile)

			// Записываем транзакцию
			transaction := models.WalletTransaction{
				UserID:      parsedUserID,
				Amount:      reward,
				Type:        "reward",
				Description: fmt.Sprintf("Quiz completion reward: %d stars", stars),
				RelatedID:   parsedCourseID,
			}
			database.DB.Create(&transaction)
		}
	}

	response := map[string]interface{}{
		"score":          score,
		"correctAnswers": correctCount,
		"totalQuestions": totalQuestions,
		"stars":          stars,
		"reward":         stars * 10,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

// GetLesson возвращает детали урока
func GetLesson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lessonID := vars["lessonId"]

	var lesson models.Lesson
	if err := database.DB.Where("id = ?", lessonID).First(&lesson).Error; err != nil {
		utils.RespondError(w, http.StatusNotFound, "Lesson not found", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, lesson)
}

// GenerateCertificateHandler POST /api/academy/certificate/{courseId} - генерация PDF-сертификата
func GenerateCertificateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["courseId"]

	var req struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	// Проверяем курс
	var course models.Course
	if err := database.DB.Where("id = ?", courseID).First(&course).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Course not found",
		})
		return
	}

	// Проверяем профиль студента
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", req.UserID).First(&profile).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "User profile not found",
		})
		return
	}

	// Проверяем прогресс по курсу
	var progress models.UserProgress
	if err := database.DB.Where("user_id = ? AND course_id = ?", req.UserID, courseID).First(&progress).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Course progress not found",
		})
		return
	}

	// Проверка: курс должен быть завершён
	if !progress.IsCompleted {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Course not completed yet",
			"progress": map[string]interface{}{
				"completedLessons": progress.CompletedLessons,
				"totalLessons":     progress.TotalLessons,
				"quizScore":        progress.QuizScore,
			},
		})
		return
	}

	// Проверяем, есть ли уже сертификат
	var existingCert models.Certificate
	err := database.DB.Where("user_id = ? AND course_id = ?", req.UserID, courseID).First(&existingCert).Error
	if err == nil {
		// Сертификат уже существует
		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "Certificate already exists",
			"data": map[string]interface{}{
				"certificateId": existingCert.ID,
				"pdfUrl":        existingCert.PDFURL,
				"issuedAt":      existingCert.IssuedAt,
			},
		})
		return
	}

	// Генерируем PDF-сертификат через сервис
	certService := services.NewCertificateService()
	certData := services.CertificateData{
		StudentName:    profile.Name,
		CourseName:     course.Title,
		Level:          profile.Level,
		Stars:          progress.StarsEarned,
		CompletionDate: time.Now(),
		Language:       profile.Language,
		QuizScore:      progress.QuizScore,
	}

	pdfPath, err := certService.GenerateCertificate(certData)
	if err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to generate certificate PDF",
			"error":   err.Error(),
		})
		return
	}

	// Создаём запись в базе данных
	certificate := models.Certificate{
		UserID:     uuid.MustParse(req.UserID),
		CourseID:   uuid.MustParse(courseID),
		CourseName: course.Title,
		UserName:   profile.Name,
		Level:      profile.Level,
		Stars:      progress.StarsEarned,
		PDFURL:     pdfPath,
		Signature:  "Chef Dima Fomin - Culinary Academy AI",
	}

	if err := database.DB.Create(&certificate).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to save certificate record",
		})
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"certificateId": certificate.ID,
			"pdfUrl":        certificate.PDFURL,
			"studentName":   certificate.UserName,
			"courseName":    certificate.CourseName,
			"issuedAt":      certificate.IssuedAt,
			"message":       certData.AIPersonalizedMessage,
		},
	})
}
