package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      taskdomain.Status    `json:"status"`
	Schedule    *taskdomain.Schedule `json:"schedule,omitempty"`
}

type templateMutationDTO struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Schedule    *taskdomain.Schedule `json:"schedule,omitempty"`
}

type taskDTO struct {
	ID           int64                `json:"id"`
	TemplateID   int64                `json:"template_id"`
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	Status       taskdomain.Status    `json:"status"`
	ScheduledFor taskdomain.Date      `json:"scheduled_for"`
	Schedule     *taskdomain.Schedule `json:"schedule,omitempty"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type templateDTO struct {
	ID               int64                `json:"id"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	Schedule         *taskdomain.Schedule `json:"schedule,omitempty"`
	NextOccurrenceOn *string              `json:"next_occurrence_on,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type healthDTO struct {
	Status string `json:"status"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		TemplateID:   task.TemplateID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		ScheduledFor: task.ScheduledFor,
		Schedule:     task.Schedule,
		CompletedAt:  task.CompletedAt,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

func newTemplateDTO(template *taskdomain.Template) templateDTO {
	var nextOccurrenceOn *string
	if template.NextOccurrenceOn != nil {
		value := template.NextOccurrenceOn.String()
		nextOccurrenceOn = &value
	}

	return templateDTO{
		ID:               template.ID,
		Title:            template.Title,
		Description:      template.Description,
		Schedule:         template.Schedule,
		NextOccurrenceOn: nextOccurrenceOn,
		CreatedAt:        template.CreatedAt,
		UpdatedAt:        template.UpdatedAt,
	}
}
