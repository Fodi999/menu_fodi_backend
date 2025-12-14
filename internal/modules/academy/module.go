package academy

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/service"
	transporthttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/academy/transport/http"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Module представляет Academy модуль
type Module struct {
	handlers *transporthttp.AcademyHandlers
}

// NewModule создает новый Academy модуль
func NewModule(db *gorm.DB) *Module {
	academyRepo := repo.NewAcademyRepository(db)
	academyService := service.NewAcademyService(academyRepo)
	handlers := transporthttp.NewAcademyHandlers(academyService)

	return &Module{
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует роуты модуля
func (m *Module) RegisterRoutes(r chi.Router, jwtMiddleware func(next http.Handler) http.Handler) {
	r.Route("/academy", func(r chi.Router) {
		// Public routes
		r.Get("/courses", m.handlers.GetCourses)
		r.Get("/courses/{courseId}", m.handlers.GetCourse)
		r.Get("/courses/{courseId}/lessons", m.handlers.GetCourseLessons)
		r.Get("/lessons/{lessonId}", m.handlers.GetLesson)
		r.Get("/quizzes/{courseId}", m.handlers.GetQuiz)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware)

			r.Post("/enroll", m.handlers.EnrollInCourse)
			r.Post("/lessons/complete", m.handlers.CompleteLesson)
			r.Post("/quizzes/submit", m.handlers.SubmitQuiz)
			r.Get("/progress/{courseId}", m.handlers.GetUserProgress)
			r.Post("/certificates/generate", m.handlers.GenerateCertificate)
			r.Get("/certificates", m.handlers.GetUserCertificates)
		})
	})
}
