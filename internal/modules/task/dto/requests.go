package dto

// CreateTaskRequest запрос на создание задания
type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description"`
	Reward      int64  `json:"reward" validate:"required,min=0"`
}

// UpdateTaskRequest запрос на обновление задания
type UpdateTaskRequest struct {
	Title       string `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Description string `json:"description,omitempty"`
	Reward      int64  `json:"reward,omitempty" validate:"omitempty,min=0"`
}

// StartTaskRequest запрос на начало выполнения задания
type StartTaskRequest struct {
	TaskID string `json:"task_id" validate:"required,uuid"`
}

// UpdateProgressRequest запрос на обновление прогресса
type UpdateProgressRequest struct {
	Progress int `json:"progress" validate:"required,min=0,max=100"`
}

// ApproveTaskRequest запрос на одобрение выполнения задания
type ApproveTaskRequest struct {
	UserID string `json:"user_id" validate:"required"`
	TaskID string `json:"task_id" validate:"required,uuid"`
}

// TaskResponse ответ с информацией о задании
type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Reward      int64  `json:"reward"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

// UserTaskResponse ответ с информацией о выполнении задания пользователем
type UserTaskResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	TaskID        string  `json:"task_id"`
	TaskTitle     string  `json:"task_title"`
	TaskReward    int64   `json:"task_reward"`
	Status        string  `json:"status"`
	Progress      int     `json:"progress"`
	RewardClaimed bool    `json:"reward_claimed"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// TaskStatsResponse ответ со статистикой заданий
type TaskStatsResponse struct {
	TotalTasks         int64   `json:"total_tasks"`
	CompletedTasks     int64   `json:"completed_tasks"`
	InProgressTasks    int64   `json:"in_progress_tasks"`
	PendingTasks       int64   `json:"pending_tasks"`
	TotalRewardsEarned int64   `json:"total_rewards_earned"`
	CompletionRate     float64 `json:"completion_rate"`
}
