package database

import (
	"errors"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository репозиторий для работы с заданиями
type TaskRepository struct{}

// ============================================
// Task CRUD Operations
// ============================================

// CreateTask создаёт новое задание
func (r *TaskRepository) CreateTask(task *models.Task) error {
	return DB.Create(task).Error
}

// GetTaskByID возвращает задание по ID
func (r *TaskRepository) GetTaskByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	result := DB.Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, result.Error
	}
	return &task, nil
}

// GetAllTasks возвращает все задания
func (r *TaskRepository) GetAllTasks(activeOnly bool) ([]models.Task, error) {
	var tasks []models.Task
	query := DB.Order("created_at DESC")
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// UpdateTask обновляет задание
func (r *TaskRepository) UpdateTask(task *models.Task) error {
	return DB.Save(task).Error
}

// DeleteTask удаляет задание
func (r *TaskRepository) DeleteTask(taskID uuid.UUID) error {
	result := DB.Delete(&models.Task{}, "id = ?", taskID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

// ToggleTaskActive переключает статус активности задания
func (r *TaskRepository) ToggleTaskActive(taskID uuid.UUID, isActive bool) error {
	result := DB.Model(&models.Task{}).
		Where("id = ?", taskID).
		Update("is_active", isActive)
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found")
	}
	return nil
}

// ============================================
// UserTask Operations
// ============================================

// StartTask начинает выполнение задания пользователем
func (r *TaskRepository) StartTask(userID string, taskID uuid.UUID) (*models.UserTask, error) {
	// Проверяем, существует ли задание и активно ли оно
	var task models.Task
	if err := DB.Where("id = ? AND is_active = ?", taskID, true).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found or inactive")
		}
		return nil, err
	}

	// Проверяем, не начато ли уже это задание пользователем
	var existingUserTask models.UserTask
	result := DB.Where("user_id = ? AND task_id = ?", userID, taskID).First(&existingUserTask)
	if result.Error == nil {
		// Задание уже начато
		return &existingUserTask, nil
	}

	// Создаём новую запись UserTask
	userTask := &models.UserTask{
		UserID:   userID,
		TaskID:   taskID,
		Status:   "in_progress",
		Progress: 0,
	}

	if err := DB.Create(userTask).Error; err != nil {
		return nil, err
	}

	return userTask, nil
}

// UpdateTaskProgress обновляет прогресс выполнения задания
func (r *TaskRepository) UpdateTaskProgress(userID string, taskID uuid.UUID, progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	var userTask models.UserTask
	if err := DB.Where("user_id = ? AND task_id = ?", userID, taskID).First(&userTask).Error; err != nil {
		return errors.New("user task not found")
	}

	userTask.Progress = progress
	if progress == 100 {
		userTask.MarkAsCompleted()
	}

	return DB.Save(&userTask).Error
}

// CompleteTask отмечает задание как завершённое
func (r *TaskRepository) CompleteTask(userID string, taskID uuid.UUID) error {
	var userTask models.UserTask
	if err := DB.Where("user_id = ? AND task_id = ?", userID, taskID).First(&userTask).Error; err != nil {
		return errors.New("user task not found")
	}

	userTask.MarkAsCompleted()
	return DB.Save(&userTask).Error
}

// ClaimTaskReward выдаёт награду за выполненное задание
func (r *TaskRepository) ClaimTaskReward(userID string, taskID uuid.UUID) (int64, error) {
	var userTask models.UserTask
	if err := DB.Preload("Task").Where("user_id = ? AND task_id = ?", userID, taskID).First(&userTask).Error; err != nil {
		return 0, errors.New("user task not found")
	}

	// Проверяем, можно ли получить награду
	if !userTask.CanClaimReward() {
		if userTask.RewardClaimed {
			return 0, errors.New("reward already claimed")
		}
		return 0, errors.New("task not completed yet")
	}

	// Получаем награду
	reward := userTask.Task.Reward

	// Используем транзакцию для атомарной операции
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. Отмечаем награду как полученную
		if err := userTask.ClaimReward(); err != nil {
			return err
		}
		if err := tx.Save(&userTask).Error; err != nil {
			return err
		}

		// 2. Выделяем токены из Treasury
		tokenRepo := &TokenBankRepository{}
		if err := tokenRepo.AllocateQuestReward(userID, taskID.String(), reward); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return reward, nil
}

// GetUserTasks возвращает все задания пользователя
func (r *TaskRepository) GetUserTasks(userID string, status string) ([]models.UserTask, error) {
	var userTasks []models.UserTask
	query := DB.Preload("Task").Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&userTasks).Error; err != nil {
		return nil, err
	}

	return userTasks, nil
}

// GetUserTaskByID возвращает конкретное задание пользователя
func (r *TaskRepository) GetUserTaskByID(userTaskID uuid.UUID) (*models.UserTask, error) {
	var userTask models.UserTask
	result := DB.Preload("Task").Preload("User").Where("id = ?", userTaskID).First(&userTask)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user task not found")
		}
		return nil, result.Error
	}
	return &userTask, nil
}

// GetAvailableTasks возвращает доступные задания для пользователя (не начатые)
func (r *TaskRepository) GetAvailableTasks(userID string) ([]models.Task, error) {
	var tasks []models.Task

	// Получаем ID всех заданий, которые пользователь уже начал
	var startedTaskIDs []uuid.UUID
	DB.Model(&models.UserTask{}).
		Where("user_id = ?", userID).
		Pluck("task_id", &startedTaskIDs)

	// Получаем активные задания, которые пользователь ещё не начал
	query := DB.Where("is_active = ?", true)
	if len(startedTaskIDs) > 0 {
		query = query.Where("id NOT IN ?", startedTaskIDs)
	}

	if err := query.Order("reward DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetUserTaskStats возвращает статистику заданий пользователя
func (r *TaskRepository) GetUserTaskStats(userID string) (*models.TaskStats, error) {
	var stats models.TaskStats

	// Общее количество заданий пользователя
	DB.Model(&models.UserTask{}).Where("user_id = ?", userID).Count(&stats.TotalTasks)

	// Завершённые задания
	DB.Model(&models.UserTask{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Count(&stats.CompletedTasks)

	// Задания в процессе
	DB.Model(&models.UserTask{}).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Count(&stats.InProgressTasks)

	// Ожидающие задания
	DB.Model(&models.UserTask{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&stats.PendingTasks)

	// Общая сумма полученных наград
	type RewardSum struct {
		Total int64
	}
	var rewardSum RewardSum
	DB.Model(&models.UserTask{}).
		Select("COALESCE(SUM(tasks.reward), 0) as total").
		Joins("JOIN tasks ON user_tasks.task_id = tasks.id").
		Where("user_tasks.user_id = ? AND user_tasks.reward_claimed = ?", userID, true).
		Scan(&rewardSum)
	stats.TotalRewardsEarned = rewardSum.Total

	// Процент выполнения
	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return &stats, nil
}

// ============================================
// Task Categories Operations
// ============================================

// GetTasksByCategory возвращает задания по категории
func (r *TaskRepository) GetTasksByCategory(categoryID uuid.UUID, activeOnly bool) ([]models.Task, error) {
	var tasks []models.Task
	query := DB.Where("category_id = ?", categoryID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("reward DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetAllCategories возвращает все категории заданий
func (r *TaskRepository) GetAllCategories() ([]models.TaskCategory, error) {
	var categories []models.TaskCategory
	if err := DB.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// ============================================
// Admin Operations
// ============================================

// GetAllUserTasks возвращает все задания всех пользователей (для админа)
func (r *TaskRepository) GetAllUserTasks() ([]models.UserTask, error) {
	var userTasks []models.UserTask
	if err := DB.Preload("Task").Preload("User").Order("created_at DESC").Find(&userTasks).Error; err != nil {
		return nil, err
	}
	return userTasks, nil
}

// GetTaskCompletionStats возвращает статистику выполнения конкретного задания
func (r *TaskRepository) GetTaskCompletionStats(taskID uuid.UUID) (map[string]interface{}, error) {
	var (
		totalStarted  int64
		completed     int64
		inProgress    int64
		averageProgress float64
	)

	// Общее количество пользователей, начавших задание
	DB.Model(&models.UserTask{}).Where("task_id = ?", taskID).Count(&totalStarted)

	// Завершили задание
	DB.Model(&models.UserTask{}).
		Where("task_id = ? AND status = ?", taskID, "completed").
		Count(&completed)

	// В процессе выполнения
	DB.Model(&models.UserTask{}).
		Where("task_id = ? AND status = ?", taskID, "in_progress").
		Count(&inProgress)

	// Средний прогресс
	type AvgProgress struct {
		Avg float64
	}
	var avg AvgProgress
	DB.Model(&models.UserTask{}).
		Select("COALESCE(AVG(progress), 0) as avg").
		Where("task_id = ?", taskID).
		Scan(&avg)
	averageProgress = avg.Avg

	completionRate := 0.0
	if totalStarted > 0 {
		completionRate = float64(completed) / float64(totalStarted) * 100
	}

	return map[string]interface{}{
		"task_id":          taskID,
		"total_started":    totalStarted,
		"completed":        completed,
		"in_progress":      inProgress,
		"average_progress": averageProgress,
		"completion_rate":  completionRate,
	}, nil
}
