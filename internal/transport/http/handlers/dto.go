package handlers

import (
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type recurrence struct {
	StartDate     *time.Time                `json:"start_date,omitempty"`
	EndDate       *time.Time                `json:"end_date,omitempty"`
	IntervalDays  *int                      `json:"interval_days,omitempty"`
	MonthDays     *[]int                    `json:"month_days,omitempty"`
	SpecificDates *[]time.Time              `json:"specific_dates,omitempty"`
	EvenOdd       *recurrencedomain.EvenOdd `json:"even_odd,omitempty"`
}

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     *time.Time        `json:"due_date"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrence       `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     *time.Time        `json:"due_date"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type recurrenceDTO struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Recurrence  recurrence `json:"recurrence"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func newReccurenceDTO(recurrenceData *recurrencedomain.Recurrence) recurrenceDTO {
	return recurrenceDTO{
		ID:          recurrenceData.ID,
		Title:       recurrenceData.Title,
		Description: recurrenceData.Description,
		Recurrence: recurrence{
			StartDate:     recurrenceData.StartDate,
			EndDate:       recurrenceData.EndDate,
			IntervalDays:  recurrenceData.IntervalDays,
			MonthDays:     recurrenceData.MonthDays,
			SpecificDates: recurrenceData.SpecificDates,
			EvenOdd:       recurrenceData.EvenOdd,
		},
		CreatedAt: recurrenceData.CreatedAt,
		UpdatedAt: recurrenceData.UpdatedAt,
	}
}
