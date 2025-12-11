package models

import (
	"time"

	"github.com/google/uuid"
)

// Task представляет задание для пользователя
type Task struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Reward      int64     `json:"reward" gorm:"type:bigint;not null;default:0"`
	IsActive    bool      `json:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName указывает имя таблицы в базе данных
func (Task) TableName() string {
	return "tasks"
}

// UserTask представляет связь между пользователем и заданием (выполнение задания)
type UserTask struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"type:text;not null;index"`
	TaskID      uuid.UUID `json:"task_id" gorm:"type:uuid;not null;index"`
	Status      string    `json:"status" gorm:"type:varchar(50);not null;default:'pending'"` // pending, in_progress, completed, failed
	Progress    int       `json:"progress" gorm:"type:int;default:0"`                         // Процент выполнения (0-100)
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:timestamp"`
	RewardClaimed bool     `json:"reward_claimed" gorm:"type:boolean;default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID;references:ID"`
}

// TableName указывает имя таблицы в базе данных
func (UserTask) TableName() string {
	return "user_tasks"
}

// TaskCategory представляет категорию заданий
type TaskCategory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null;unique"`
	Description string    `json:"description" gorm:"type:text"`
	Icon        string    `json:"icon" gorm:"type:varchar(100)"` // Иконка категории
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName указывает имя таблицы в базе данных
func (TaskCategory) TableName() string {
	return "task_categories"
}

// TaskWithCategory расширенная структура задания с категорией
type TaskWithCategory struct {
	Task
	CategoryID   *uuid.UUID    `json:"category_id,omitempty" gorm:"type:uuid"`
	Category     *TaskCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID"`
}

// TaskStats статистика по заданиям пользователя
type TaskStats struct {
	TotalTasks       int64 `json:"total_tasks"`
	CompletedTasks   int64 `json:"completed_tasks"`
	InProgressTasks  int64 `json:"in_progress_tasks"`
	PendingTasks     int64 `json:"pending_tasks"`
	TotalRewardsEarned int64 `json:"total_rewards_earned"`
	CompletionRate   float64 `json:"completion_rate"` // Процент выполненных заданий
}

// IsCompleted проверяет, завершено ли задание
func (ut *UserTask) IsCompleted() bool {
	return ut.Status == "completed"
}

// CanClaimReward проверяет, можно ли получить награду
func (ut *UserTask) CanClaimReward() bool {
	return ut.IsCompleted() && !ut.RewardClaimed
}

// MarkAsCompleted отмечает задание как завершённое
func (ut *UserTask) MarkAsCompleted() {
	now := time.Now()
	ut.Status = "completed"
	ut.Progress = 100
	ut.CompletedAt = &now
}

// ClaimReward отмечает награду как полученную
func (ut *UserTask) ClaimReward() error {
	if !ut.CanClaimReward() {
		return ErrRewardAlreadyClaimed
	}
	ut.RewardClaimed = true
	return nil
}

// Custom errors
var (
	ErrRewardAlreadyClaimed = ErrCustom{Message: "reward already claimed"}
	ErrTaskNotCompleted     = ErrCustom{Message: "task not completed yet"}
	ErrTaskNotFound         = ErrCustom{Message: "task not found"}
	ErrTaskAlreadyStarted   = ErrCustom{Message: "task already started"}
)

// ErrCustom кастомная ошибка
type ErrCustom struct {
	Message string
}

func (e ErrCustom) Error() string {
	return e.Message
}
