package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrenceDTO struct {
	Type      taskdomain.RecurrenceType `json:"type"`
	Interval  *int                      `json:"interval,omitempty"`
	Day       *int                      `json:"day,omitempty"`
	Dates     []time.Time               `json:"dates,omitempty"`
	Parity    *string                   `json:"parity,omitempty"`
	StartDate time.Time                 `json:"start_date"`
	EndDate   time.Time                 `json:"end_date"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	DueDate     *time.Time        `json:"due_date"`
	Recurrence  *recurrenceDTO    `json:"recurrence"`
}

type taskDTO struct {
	ID               int64                        `json:"id"`
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Status           taskdomain.Status            `json:"status"`
	DueDate          *time.Time                   `json:"due_date"`
	IsTemplate       bool                         `json:"is_template"`
	ParentID         *int64                       `json:"parent_id"`
	RecurrenceConfig *taskdomain.RecurrenceConfig `json:"recurrence_config"`
	CreatedAt        time.Time                    `json:"created_at"`
	UpdatedAt        time.Time                    `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:               task.ID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		DueDate:          task.DueDate,
		IsTemplate:       task.IsTemplate,
		ParentID:         task.ParentID,
		RecurrenceConfig: task.RecurrenceConfig,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}
