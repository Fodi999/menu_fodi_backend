package task

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/task/service"
	taskhttp "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/task/transport/http"
	"github.com/go-chi/chi/v5"
)

// Module представляет модуль заданий
type Module struct {
	service  service.TaskService
	handlers *taskhttp.TaskHandlers
}

// NewModule создаёт новый модуль заданий
func NewModule() *Module {
	taskService := service.NewTaskService()
	handlers := taskhttp.NewTaskHandlers(taskService)

	return &Module{
		service:  taskService,
		handlers: handlers,
	}
}

// RegisterRoutes регистрирует все роуты модуля заданий
func (m *Module) RegisterRoutes(r chi.Router) {
	// Public routes - все задания (для неавторизованных пользователей)
	r.Get("/tasks", m.handlers.GetAllTasks)
	r.Get("/tasks/{taskID}", m.handlers.GetTaskByID)

	// User routes - требуют авторизации (TODO: добавить middleware)
	r.Route("/user/tasks", func(r chi.Router) {
		// TODO: r.Use(authMiddleware)
		r.Get("/", m.handlers.GetMyTasks)
		r.Get("/available", m.handlers.GetAvailableTasks)
		r.Get("/stats", m.handlers.GetMyStats)
		r.Post("/start", m.handlers.StartTask)
		r.Patch("/{taskID}/progress", m.handlers.UpdateProgress)
		r.Post("/{taskID}/complete", m.handlers.CompleteTask)
		r.Post("/{taskID}/claim", m.handlers.ClaimReward)
	})

	// Admin routes - требуют admin роли (TODO: добавить adminMiddleware)
	r.Route("/admin/tasks", func(r chi.Router) {
		// TODO: r.Use(authMiddleware)
		// TODO: r.Use(adminMiddleware)

		// Task management
		r.Post("/", m.handlers.CreateTask) // POST /api/admin/tasks
		r.Put("/{taskID}", m.handlers.UpdateTask)
		r.Delete("/{taskID}", m.handlers.DeleteTask)

		// User tasks monitoring
		r.Get("/users", m.handlers.GetAllUserTasks)
		r.Get("/{taskID}/stats", m.handlers.GetTaskStats)

		// Main admin endpoints: Approve task completion and pay reward
		r.Post("/approve", m.handlers.ApproveTaskCompletion)                // POST /api/admin/tasks/approve (JSON body)
		r.Post("/{taskID}/approve", m.handlers.ApproveTaskCompletionByPath) // POST /api/admin/tasks/{taskID}/approve
	})
}

// GetService возвращает сервис заданий (для использования в других модулях)
func (m *Module) GetService() service.TaskService {
	return m.service
}
