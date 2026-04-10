package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Status        taskdomain.Status     `json:"status"`
	StartDateTime time.Time             `json:"start_date,omitempty"`
	Rec           taskdomain.Recurrence `json:"recurrence"`
}

type taskDTO struct {
	ID            int64                 `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Status        taskdomain.Status     `json:"status"`
	Rec           taskdomain.Recurrence `json:"recurrence"`
	StartDateTime time.Time             `json:"start_date,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Rec:         task.Rec,
		StartDateTime: task.StartDateTime,
	}
}
