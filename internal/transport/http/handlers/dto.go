package handlers

import (
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	ScheduledAt *time.Time        `json:"scheduled_at,omitempty"`

	RecurrenceType          *taskdomain.RecurrenceType `json:"recurrence_type,omitempty"`
	RecurrenceDailyInterval *int                       `json:"recurrence_daily_interval,omitempty"`
	RecurrenceMonthlyDays   []int                      `json:"recurrence_monthly_days,omitempty"`
	RecurrenceSpecificDates []string                   `json:"recurrence_specific_dates,omitempty"` // YYYY-MM-DD
	RecurrenceDayParity *taskdomain.Parity         `json:"recurrence_day_parity,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ScheduledAt  *time.Time        `json:"scheduled_at,omitempty"`
	ParentTaskID *int64            `json:"parent_task_id,omitempty"`

	RecurrenceType          *taskdomain.RecurrenceType `json:"recurrence_type,omitempty"`
	RecurrenceDailyInterval *int                       `json:"recurrence_daily_interval,omitempty"`
	RecurrenceMonthlyDays   []int                      `json:"recurrence_monthly_days,omitempty"`
	RecurrenceSpecificDates []string                   `json:"recurrence_specific_dates,omitempty"` // YYYY-MM-DD
	RecurrenceDayParity *taskdomain.Parity         `json:"recurrence_day_parity,omitempty"`
}

func newTaskDTO(t *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:                      t.ID,
		Title:                   t.Title,
		Description:             t.Description,
		Status:                  t.Status,
		CreatedAt:               t.CreatedAt,
		UpdatedAt:               t.UpdatedAt,
		ScheduledAt:             t.ScheduledAt,
		ParentTaskID:            t.ParentTaskID,
		RecurrenceType:          t.RecurrenceType,
		RecurrenceDailyInterval: t.RecurrenceDailyInterval,
		RecurrenceMonthlyDays:   t.RecurrenceMonthlyDays,
		RecurrenceDayParity: t.RecurrenceDayParity,
	}
	if len(t.RecurrenceSpecificDates) > 0 {
		dto.RecurrenceSpecificDates = make([]string, len(t.RecurrenceSpecificDates))
		for i, d := range t.RecurrenceSpecificDates {
			dto.RecurrenceSpecificDates[i] = d.UTC().Format("2006-01-02")
		}
	}
	return dto
}

// parseDates конвертирует строки YYYY-MM-DD в []time.Time (полночь UTC).
func parseDates(strs []string) ([]time.Time, error) {
	if len(strs) == 0 {
		return nil, nil
	}
	dates := make([]time.Time, 0, len(strs))
	for _, s := range strs {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, fmt.Errorf("recurrence_specific_dates: invalid date %q", s)
		}
		dates = append(dates, d.UTC())
	}
	return dates, nil
}
