package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/task/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/task/service"
	"github.com/go-chi/chi/v5"
)

type TaskHandlers struct {
	service service.TaskService
}

func NewTaskHandlers(service service.TaskService) *TaskHandlers {
	return &TaskHandlers{
		service: service,
	}
}

// ============================================
// Task Endpoints
// ============================================

// GetAllTasks возвращает все задания
// GET /api/tasks?active_only=true
func (h *TaskHandlers) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	tasks, err := h.service.GetAllTasks(activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTaskByID возвращает задание по ID
// GET /api/tasks/{taskID}
func (h *TaskHandlers) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	task, err := h.service.GetTaskByID(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// CreateTask создаёт новое задание (admin only)
// POST /api/admin/tasks
func (h *TaskHandlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(req.Title, req.Description, req.Reward)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask обновляет задание (admin only)
// PUT /api/admin/tasks/{taskID}
func (h *TaskHandlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.service.UpdateTask(taskID, req.Title, req.Description, req.Reward)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask удаляет задание (admin only)
// DELETE /api/admin/tasks/{taskID}
func (h *TaskHandlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	if err := h.service.DeleteTask(taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// User Task Endpoints
// ============================================

// StartTask начинает выполнение задания
// POST /api/user/tasks/start
func (h *TaskHandlers) StartTask(w http.ResponseWriter, r *http.Request) {
	var req dto.StartTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Получить userID из JWT токена (r.Context().Value(middleware.UserContextKey))
	userID := r.Header.Get("X-User-ID") // временно

	userTask, err := h.service.StartTask(userID, req.TaskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userTask)
}

// UpdateProgress обновляет прогресс выполнения задания
// PATCH /api/user/tasks/{taskID}/progress
func (h *TaskHandlers) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	var req dto.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	if err := h.service.UpdateTaskProgress(userID, taskID, req.Progress); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Progress updated"})
}

// CompleteTask отмечает задание как завершённое
// POST /api/user/tasks/{taskID}/complete
func (h *TaskHandlers) CompleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	if err := h.service.CompleteTask(userID, taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task completed"})
}

// ClaimReward забирает награду за выполненное задание
// POST /api/user/tasks/{taskID}/claim
func (h *TaskHandlers) ClaimReward(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	reward, err := h.service.ClaimTaskReward(userID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Reward claimed successfully",
		"reward":  reward,
	})
}

// GetMyTasks возвращает задания текущего пользователя
// GET /api/user/tasks?status=completed
func (h *TaskHandlers) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	tasks, err := h.service.GetUserTasks(userID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetAvailableTasks возвращает доступные задания для пользователя
// GET /api/user/tasks/available
func (h *TaskHandlers) GetAvailableTasks(w http.ResponseWriter, r *http.Request) {
	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	tasks, err := h.service.GetAvailableTasks(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetMyStats возвращает статистику заданий пользователя
// GET /api/user/tasks/stats
func (h *TaskHandlers) GetMyStats(w http.ResponseWriter, r *http.Request) {
	// TODO: Получить userID из JWT токена
	userID := r.Header.Get("X-User-ID") // временно

	stats, err := h.service.GetUserTaskStats(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ============================================
// Admin Endpoints
// ============================================

// ApproveTaskCompletion одобряет выполнение задания и начисляет награду
// POST /api/admin/tasks/approve (JSON body)
func (h *TaskHandlers) ApproveTaskCompletion(w http.ResponseWriter, r *http.Request) {
	var req dto.ApproveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ApproveTaskCompletion(req.UserID, req.TaskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Task completion approved and reward paid",
		"user_id": req.UserID,
		"task_id": req.TaskID,
	})
}

// ApproveTaskCompletionByPath одобряет выполнение задания через URL параметры
// POST /api/admin/tasks/{taskID}/approve?user_id=xxx
// или с JSON body: {"user_id": "xxx"}
func (h *TaskHandlers) ApproveTaskCompletionByPath(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	// Пробуем получить user_id из query параметра
	userID := r.URL.Query().Get("user_id")

	// Если не в query, пробуем из body
	if userID == "" {
		var req struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			userID = req.UserID
		}
	}

	if userID == "" {
		http.Error(w, "user_id is required (in query or body)", http.StatusBadRequest)
		return
	}

	if err := h.service.ApproveTaskCompletion(userID, taskID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Task completion approved and reward paid",
		"user_id": userID,
		"task_id": taskID,
	})
}

// GetAllUserTasks возвращает все задания всех пользователей (admin)
// GET /api/admin/tasks/users
func (h *TaskHandlers) GetAllUserTasks(w http.ResponseWriter, r *http.Request) {
	userTasks, err := h.service.GetAllUserTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userTasks)
}

// GetTaskStats возвращает статистику по конкретному заданию (admin)
// GET /api/admin/tasks/{taskID}/stats
func (h *TaskHandlers) GetTaskStats(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")

	stats, err := h.service.GetTaskCompletionStats(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
