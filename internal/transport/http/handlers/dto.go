package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title             string                        `json:"title"`
	Description       string                        `json:"description"`
	Status            taskdomain.Status             `json:"status"`
	PeriodicityType   taskdomain.PeriodicityType    `json:"periodicity_type,omitempty"`
	PeriodicityConfig *taskdomain.PeriodicityConfig `json:"periodicity_config,omitempty"`
}

type taskDTO struct {
	ID                int64                        `json:"id"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	Status            taskdomain.Status            `json:"status"`
	PeriodicityType   taskdomain.PeriodicityType   `json:"periodicity_type"`
	PeriodicityConfig taskdomain.PeriodicityConfig `json:"periodicity_config"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:                task.ID,
		Title:             task.Title,
		Description:       task.Description,
		Status:            task.Status,
		PeriodicityType:   task.PeriodicityType,
		PeriodicityConfig: task.PeriodicityConfig,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}
