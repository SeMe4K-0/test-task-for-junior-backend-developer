package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// dateOnly serializes/deserializes time.Time as a plain "YYYY-MM-DD" string
// so that scheduled dates look readable in the API.
type dateOnly struct {
	time.Time
}

func (d *dateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}
	d.Time = t.UTC()
	return nil
}

func (d dateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format("2006-01-02"))
}

// ---- request DTOs -------------------------------------------------------

type taskMutationDTO struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	ScheduledDate *dateOnly         `json:"scheduled_date,omitempty"`
}

type recurrenceRequestDTO struct {
	Type      taskdomain.RecurrenceType `json:"type"`
	Interval  int                       `json:"interval,omitempty"`
	MonthDays []int                     `json:"month_days,omitempty"`
	Dates     []dateOnly                `json:"dates,omitempty"`
	Parity    taskdomain.EvenOddParity  `json:"parity,omitempty"`
}

type createRecurringTaskDTO struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      taskdomain.Status    `json:"status"`
	Recurrence  recurrenceRequestDTO `json:"recurrence"`
	StartDate   *dateOnly            `json:"start_date,omitempty"`
	EndDate     *dateOnly            `json:"end_date,omitempty"`
}

// ---- response DTOs ------------------------------------------------------

type recurrenceResponseDTO struct {
	Type      taskdomain.RecurrenceType `json:"type"`
	Interval  int                       `json:"interval,omitempty"`
	MonthDays []int                     `json:"month_days,omitempty"`
	Dates     []dateOnly                `json:"dates,omitempty"`
	Parity    taskdomain.EvenOddParity  `json:"parity,omitempty"`
}

type taskDTO struct {
	ID            int64                  `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Status        taskdomain.Status      `json:"status"`
	ScheduledDate *dateOnly              `json:"scheduled_date,omitempty"`
	Recurrence    *recurrenceResponseDTO `json:"recurrence,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	if task.ScheduledDate != nil {
		dto.ScheduledDate = &dateOnly{*task.ScheduledDate}
	}

	if task.Recurrence != nil {
		dto.Recurrence = fromDomainRecurrence(task.Recurrence)
	}

	return dto
}

func fromDomainRecurrence(r *taskdomain.Recurrence) *recurrenceResponseDTO {
	dto := &recurrenceResponseDTO{
		Type:      r.Type,
		Interval:  r.Interval,
		MonthDays: r.MonthDays,
		Parity:    r.Parity,
	}
	for _, d := range r.Dates {
		dto.Dates = append(dto.Dates, dateOnly{d})
	}
	return dto
}

func toDomainRecurrence(dto recurrenceRequestDTO) taskdomain.Recurrence {
	r := taskdomain.Recurrence{
		Type:      dto.Type,
		Interval:  dto.Interval,
		MonthDays: dto.MonthDays,
		Parity:    dto.Parity,
	}
	for _, d := range dto.Dates {
		r.Dates = append(r.Dates, d.Time)
	}
	return r
}
