package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Periodicity taskdomain.Periodicity `json:"periodicity"`
	Status      taskdomain.Status      `json:"status"`
	ScheduledAt []time.Time            `json:"scheduled_at"`
	Frequency   int64                  `json:"frequency"`
}

type taskDTO struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Periodicity taskdomain.Periodicity `json:"periodicity"`
	ScheduledAt []time.Time            `json:"scheduled_at"`
	Frequency   int64                  `json:"frequency"`
	Status      taskdomain.Status      `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Periodicity: task.Periodicity,
		ScheduledAt: task.ScheduledAt,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
