package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type repeatedDTO struct {
	Type          taskdomain.PeriodType `json:"type"`
	EveryNDays    int                   `json:"every_n_days,omitempty"`
	DayOfMonth    int                   `json:"day_of_month,omitempty"`
	SpecificDates []time.Time           `json:"specific_dates,omitempty"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Repeated    repeatedDTO       `json:"repeated,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Repeated    repeatedDTO       `json:"repeated"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newRecurrenceDTO(r taskdomain.Repeated) repeatedDTO {
	return repeatedDTO{
		Type:          r.Type,
		EveryNDays:    r.EveryNDays,
		DayOfMonth:    r.DayOfMonth,
		SpecificDates: r.SpecificDates,
	}
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Repeated:    newRecurrenceDTO(task.Repeated),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
