package service

import (
	"errors"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskService интерфейс для бизнес-логики заданий
type TaskService interface {
	// Task Management
	GetAllTasks(activeOnly bool) ([]models.Task, error)
	GetTaskByID(taskID string) (*models.Task, error)
	CreateTask(title, description string, reward int64) (*models.Task, error)
	UpdateTask(taskID string, title, description string, reward int64) (*models.Task, error)
	DeleteTask(taskID string) error
	ToggleTaskActive(taskID string, isActive bool) error

	// User Task Operations
	StartTask(userID, taskID string) (*models.UserTask, error)
	UpdateTaskProgress(userID, taskID string, progress int) error
	CompleteTask(userID, taskID string) error
	ClaimTaskReward(userID, taskID string) (int64, error)
	GetUserTasks(userID string, status string) ([]models.UserTask, error)
	GetAvailableTasks(userID string) ([]models.Task, error)
	GetUserTaskStats(userID string) (*models.TaskStats, error)

	// Admin Operations
	ApproveTaskCompletion(userID, taskID string) error
	GetAllUserTasks() ([]models.UserTask, error)
	GetTaskCompletionStats(taskID string) (map[string]interface{}, error)
}

type taskService struct {
	db            *gorm.DB
	taskRepo      *database.TaskRepository
	tokenBankRepo *database.TokenBankRepository
}

// NewTaskService создаёт новый экземпляр сервиса заданий
func NewTaskService() TaskService {
	return &taskService{
		db:            database.GetDB(),
		taskRepo:      &database.TaskRepository{},
		tokenBankRepo: &database.TokenBankRepository{},
	}
}

// ============================================
// Task Management
// ============================================

// GetAllTasks возвращает все задания
func (s *taskService) GetAllTasks(activeOnly bool) ([]models.Task, error) {
	return s.taskRepo.GetAllTasks(activeOnly)
}

// GetTaskByID возвращает задание по ID
func (s *taskService) GetTaskByID(taskID string) (*models.Task, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}
	return s.taskRepo.GetTaskByID(id)
}

// CreateTask создаёт новое задание
func (s *taskService) CreateTask(title, description string, reward int64) (*models.Task, error) {
	if title == "" {
		return nil, errors.New("task title is required")
	}
	if reward < 0 {
		return nil, errors.New("reward cannot be negative")
	}

	task := &models.Task{
		Title:       title,
		Description: description,
		Reward:      reward,
		IsActive:    true,
	}

	if err := s.taskRepo.CreateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

// UpdateTask обновляет задание
func (s *taskService) UpdateTask(taskID string, title, description string, reward int64) (*models.Task, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}

	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	if title != "" {
		task.Title = title
	}
	if description != "" {
		task.Description = description
	}
	if reward >= 0 {
		task.Reward = reward
	}

	if err := s.taskRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask удаляет задание
func (s *taskService) DeleteTask(taskID string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return errors.New("invalid task ID format")
	}
	return s.taskRepo.DeleteTask(id)
}

// ToggleTaskActive переключает статус активности задания
func (s *taskService) ToggleTaskActive(taskID string, isActive bool) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return errors.New("invalid task ID format")
	}
	return s.taskRepo.ToggleTaskActive(id, isActive)
}

// ============================================
// User Task Operations
// ============================================

// StartTask начинает выполнение задания пользователем
func (s *taskService) StartTask(userID, taskID string) (*models.UserTask, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}
	return s.taskRepo.StartTask(userID, id)
}

// UpdateTaskProgress обновляет прогресс выполнения задания
func (s *taskService) UpdateTaskProgress(userID, taskID string, progress int) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return errors.New("invalid task ID format")
	}
	return s.taskRepo.UpdateTaskProgress(userID, id, progress)
}

// CompleteTask отмечает задание как завершённое
func (s *taskService) CompleteTask(userID, taskID string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return errors.New("invalid task ID format")
	}
	return s.taskRepo.CompleteTask(userID, id)
}

// ClaimTaskReward выдаёт награду за выполненное задание (пользователь сам забирает)
func (s *taskService) ClaimTaskReward(userID, taskID string) (int64, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return 0, errors.New("invalid task ID format")
	}
	return s.taskRepo.ClaimTaskReward(userID, id)
}

// GetUserTasks возвращает задания пользователя
func (s *taskService) GetUserTasks(userID string, status string) ([]models.UserTask, error) {
	return s.taskRepo.GetUserTasks(userID, status)
}

// GetAvailableTasks возвращает доступные задания для пользователя
func (s *taskService) GetAvailableTasks(userID string) ([]models.Task, error) {
	return s.taskRepo.GetAvailableTasks(userID)
}

// GetUserTaskStats возвращает статистику заданий пользователя
func (s *taskService) GetUserTaskStats(userID string) (*models.TaskStats, error) {
	return s.taskRepo.GetUserTaskStats(userID)
}

// ============================================
// Admin Operations
// ============================================

// ApproveTaskCompletion одобряет выполнение задания и начисляет токены
// Это главный метод для администратора, который начисляет награду пользователю
func (s *taskService) ApproveTaskCompletion(userID, taskID string) error {
	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		return errors.New("invalid task ID format")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Получаем задание
		task, err := s.taskRepo.GetTaskByID(taskUUID)
		if err != nil {
			return err
		}

		// 2. Получаем запись о выполнении задания пользователем
		var userTask models.UserTask
		if err := tx.Preload("Task").
			Where("user_id = ? AND task_id = ?", userID, taskUUID).
			First(&userTask).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user task not found")
			}
			return err
		}

		// 3. Проверяем, не была ли уже выплачена награда
		if userTask.RewardClaimed {
			return errors.New("reward already paid")
		}

		// 4. Начисляем токены из Treasury
		if err := s.tokenBankRepo.AllocateFromTreasury(userID, task.Reward); err != nil {
			return err
		}

		// 5. Обновляем статус задания
		now := time.Now()
		userTask.Status = "completed"
		userTask.Progress = 100
		userTask.RewardClaimed = true
		userTask.CompletedAt = &now

		if err := tx.Save(&userTask).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetAllUserTasks возвращает все задания всех пользователей (для админа)
func (s *taskService) GetAllUserTasks() ([]models.UserTask, error) {
	return s.taskRepo.GetAllUserTasks()
}

// GetTaskCompletionStats возвращает статистику выполнения конкретного задания
func (s *taskService) GetTaskCompletionStats(taskID string) (map[string]interface{}, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, errors.New("invalid task ID format")
	}
	return s.taskRepo.GetTaskCompletionStats(id)
}
