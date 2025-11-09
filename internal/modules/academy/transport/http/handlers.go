package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AcademyHandlers обрабатывает HTTP-запросы для академии
type AcademyHandlers struct {
	service *service.AcademyService
	logger  *zap.Logger
}

// NewAcademyHandlers создает новые хендлеры
func NewAcademyHandlers(service *service.AcademyService) *AcademyHandlers {
	return &AcademyHandlers{
		service: service,
		logger:  logger.Log,
	}
}

// ============================================================================
// Courses (Public)
// ============================================================================

// GetCourses godoc
// @Summary Получить список курсов
// @Description Возвращает список опубликованных курсов с фильтрацией
// @Tags academy
// @Accept json
// @Produce json
// @Param language query string false "Фильтр по языку"
// @Param category query string false "Фильтр по категории"
// @Param level query int false "Фильтр по уровню"
// @Success 200 {array} dto.CourseResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/courses [get]
func (h *AcademyHandlers) GetCourses(w http.ResponseWriter, r *http.Request) {
	filters := &dto.CourseFilters{
		Language: r.URL.Query().Get("language"),
		Category: r.URL.Query().Get("category"),
	}

	courses, err := h.service.GetCourses(filters)
	if err != nil {
		h.logger.Error("Failed to get courses", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get courses")
		return
	}

	httpx.Success(w, courses)
}

// GetCourse godoc
// @Summary Получить детали курса
// @Description Возвращает информацию о конкретном курсе
// @Tags academy
// @Accept json
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200 {object} dto.CourseResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/courses/{courseId} [get]
func (h *AcademyHandlers) GetCourse(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	course, err := h.service.GetCourse(courseID)
	if err != nil {
		if err == repo.ErrCourseNotFound {
			httpx.Error(w, http.StatusNotFound, "Course not found")
			return
		}
		h.logger.Error("Failed to get course", zap.Error(err), zap.String("courseId", courseIDStr))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get course")
		return
	}

	httpx.Success(w, course)
}

// GetCourseLessons godoc
// @Summary Получить уроки курса
// @Description Возвращает список уроков для курса
// @Tags academy
// @Accept json
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200 {array} dto.LessonResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/courses/{courseId}/lessons [get]
func (h *AcademyHandlers) GetCourseLessons(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	lessons, err := h.service.GetCourseLessons(courseID)
	if err != nil {
		h.logger.Error("Failed to get lessons", zap.Error(err), zap.String("courseId", courseIDStr))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get lessons")
		return
	}

	httpx.Success(w, lessons)
}

// GetLesson godoc
// @Summary Получить детали урока
// @Description Возвращает информацию о конкретном уроке
// @Tags academy
// @Accept json
// @Produce json
// @Param lessonId path string true "ID урока"
// @Success 200 {object} dto.LessonResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/lessons/{lessonId} [get]
func (h *AcademyHandlers) GetLesson(w http.ResponseWriter, r *http.Request) {
	lessonIDStr := chi.URLParam(r, "lessonId")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	lesson, err := h.service.GetLesson(lessonID)
	if err != nil {
		if err == repo.ErrLessonNotFound {
			httpx.Error(w, http.StatusNotFound, "Lesson not found")
			return
		}
		h.logger.Error("Failed to get lesson", zap.Error(err), zap.String("lessonId", lessonIDStr))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get lesson")
		return
	}

	httpx.Success(w, lesson)
}

// GetQuiz godoc
// @Summary Получить вопросы теста
// @Description Возвращает случайные вопросы для теста курса
// @Tags academy
// @Accept json
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200 {array} dto.QuizQuestionResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/quizzes/{courseId} [get]
func (h *AcademyHandlers) GetQuiz(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	questions, err := h.service.GetQuiz(courseID)
	if err != nil {
		if err == repo.ErrNoQuizQuestions {
			httpx.Error(w, http.StatusNotFound, "No quiz questions found")
			return
		}
		h.logger.Error("Failed to get quiz", zap.Error(err), zap.String("courseId", courseIDStr))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get quiz")
		return
	}

	httpx.Success(w, questions)
}

// ============================================================================
// Protected Endpoints (JWT Required)
// ============================================================================

// EnrollInCourse godoc
// @Summary Записаться на курс
// @Description Регистрирует пользователя на курс
// @Tags academy
// @Accept json
// @Produce json
// @Param request body dto.EnrollRequest true "Запрос на запись"
// @Success 200 {object} httpx.SuccessResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/enroll [post]
func (h *AcademyHandlers) EnrollInCourse(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.EnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.EnrollInCourse(*userID, req.CourseID); err != nil {
		if err == repo.ErrAlreadyEnrolled {
			httpx.Error(w, http.StatusBadRequest, "Already enrolled in this course")
			return
		}
		if err == repo.ErrCourseNotFound {
			httpx.Error(w, http.StatusNotFound, "Course not found")
			return
		}
		h.logger.Error("Failed to enroll", zap.Error(err), zap.String("userId", userID.String()))
		httpx.Error(w, http.StatusInternalServerError, "Failed to enroll")
		return
	}

	h.logger.Info("User enrolled in course",
		zap.String("userId", userID.String()),
		zap.String("courseId", req.CourseID.String()))

	httpx.Success(w, map[string]interface{}{
		"message": "Successfully enrolled in course",
	})
}

// CompleteLesson godoc
// @Summary Завершить урок
// @Description Отмечает урок как завершенный
// @Tags academy
// @Accept json
// @Produce json
// @Param request body dto.CompleteLessonRequest true "Запрос на завершение"
// @Success 200 {object} httpx.SuccessResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/lessons/complete [post]
func (h *AcademyHandlers) CompleteLesson(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CompleteLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.CompleteLesson(*userID, req.LessonID); err != nil {
		if err == repo.ErrLessonNotFound {
			httpx.Error(w, http.StatusNotFound, "Lesson not found")
			return
		}
		h.logger.Error("Failed to complete lesson", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to complete lesson")
		return
	}

	httpx.Success(w, map[string]interface{}{
		"message": "Lesson completed successfully",
	})
}

// SubmitQuiz godoc
// @Summary Отправить ответы теста
// @Description Проверяет ответы и начисляет награды
// @Tags academy
// @Accept json
// @Produce json
// @Param request body dto.QuizRequest true "Ответы теста"
// @Success 200 {object} dto.QuizResultResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/quizzes/submit [post]
func (h *AcademyHandlers) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.QuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.service.SubmitQuiz(*userID, &req)
	if err != nil {
		if err == repo.ErrNoQuizQuestions {
			httpx.Error(w, http.StatusNotFound, "No quiz found")
			return
		}
		h.logger.Error("Failed to submit quiz", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to submit quiz")
		return
	}

	h.logger.Info("Quiz submitted",
		zap.String("userId", userID.String()),
		zap.String("courseId", req.CourseID.String()),
		zap.Int("score", result.Score),
		zap.Int("stars", result.StarsEarned))

	httpx.Success(w, result)
}

// GetUserProgress godoc
// @Summary Получить прогресс пользователя
// @Description Возвращает прогресс пользователя по курсу
// @Tags academy
// @Accept json
// @Produce json
// @Param courseId path string true "ID курса"
// @Success 200 {object} dto.UserProgressResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/progress/{courseId} [get]
func (h *AcademyHandlers) GetUserProgress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	progress, err := h.service.GetUserProgress(*userID, courseID)
	if err != nil {
		if err == repo.ErrNotEnrolled {
			httpx.Error(w, http.StatusNotFound, "Not enrolled in this course")
			return
		}
		h.logger.Error("Failed to get progress", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get progress")
		return
	}

	httpx.Success(w, progress)
}

// GenerateCertificate godoc
// @Summary Сгенерировать сертификат
// @Description Генерирует PDF-сертификат для завершенного курса
// @Tags academy
// @Accept json
// @Produce json
// @Param request body dto.CertificateRequest true "Запрос на сертификат"
// @Success 200 {object} dto.CertificateResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/certificates/generate [post]
func (h *AcademyHandlers) GenerateCertificate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CertificateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cert, err := h.service.GenerateCertificate(*userID, req.CourseID)
	if err != nil {
		if err == repo.ErrNotEnrolled {
			httpx.Error(w, http.StatusBadRequest, "Not enrolled in this course")
			return
		}
		if err == repo.ErrCourseNotCompleted {
			httpx.Error(w, http.StatusBadRequest, "Course not completed yet")
			return
		}
		h.logger.Error("Failed to generate certificate", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to generate certificate")
		return
	}

	h.logger.Info("Certificate generated",
		zap.String("userId", userID.String()),
		zap.String("courseId", req.CourseID.String()),
		zap.String("certificateId", cert.ID.String()))

	httpx.Success(w, cert)
}

// GetUserCertificates godoc
// @Summary Получить сертификаты пользователя
// @Description Возвращает все сертификаты пользователя
// @Tags academy
// @Accept json
// @Produce json
// @Success 200 {object} dto.CertificateListResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/academy/certificates [get]
func (h *AcademyHandlers) GetUserCertificates(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == nil {
		httpx.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	certs, err := h.service.GetUserCertificates(*userID)
	if err != nil {
		h.logger.Error("Failed to get certificates", zap.Error(err))
		httpx.Error(w, http.StatusInternalServerError, "Failed to get certificates")
		return
	}

	httpx.Success(w, certs)
}
