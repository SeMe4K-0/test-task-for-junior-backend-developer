package handlers

import (
	"encoding/json"

	"time"

	scheduledomain "example.com/taskservice/internal/domain/schedule"
)

type scheduleMutationDTO struct {
	Title            string                        `json:"title"`
	Description      string                        `json:"description"`
	RecurrenceType   scheduledomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceParams json.RawMessage               `json:"recurrence_params"`
	IsActive         bool                          `json:"is_active"`
}

type scheduleDTO struct {
	ID               int64                         `json:"id"`
	Title            string                        `json:"title"`
	Description      string                        `json:"description"`
	RecurrenceType   scheduledomain.RecurrenceType `json:"recurrence_type"`
	RecurrenceParams json.RawMessage               `json:"recurrence_params"`
	IsActive         bool                          `json:"is_active"`
	LastCreatedDate  *time.Time                    `json:"last_created_date"`
	CreatedAt        time.Time                     `json:"created_at"`
	UpdatedAt        time.Time                     `json:"updated_at"`
}

func newScheduleDTO(schedule *scheduledomain.Schedule) scheduleDTO {
	return scheduleDTO{
		ID:               schedule.ID,
		Title:            schedule.Title,
		Description:      schedule.Description,
		RecurrenceType:   schedule.RecurrenceType,
		RecurrenceParams: schedule.RecurrenceParams,
		IsActive:         schedule.IsActive,
		LastCreatedDate:  schedule.LastCreatedDate,
		CreatedAt:        schedule.CreatedAt,
		UpdatedAt:        schedule.UpdatedAt,
	}
}
