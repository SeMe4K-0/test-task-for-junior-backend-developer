package handlers

import (
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

// --- Task DTOs ---

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID               int64             `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Status           taskdomain.Status `json:"status"`
	RecurrenceRuleID *int64            `json:"recurrence_rule_id,omitempty"`
	ScheduledDate    *string           `json:"scheduled_date,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:               task.ID,
		Title:            task.Title,
		Description:      task.Description,
		Status:           task.Status,
		RecurrenceRuleID: task.RecurrenceRuleID,
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
	if task.ScheduledDate != nil {
		s := task.ScheduledDate.Format("2006-01-02")
		dto.ScheduledDate = &s
	}
	return dto
}

// --- Recurrence Rule DTOs ---

type recurrenceParamsDTO struct {
	Type       recurrencedomain.RecurrenceType `json:"type"`
	EveryNDays *int                            `json:"every_n_days,omitempty"`
	DayOfMonth *int                            `json:"day_of_month,omitempty"`
	Dates      []string                        `json:"dates,omitempty"`
	EvenOdd    recurrencedomain.EvenOddType    `json:"even_odd,omitempty"`
}

type createRecurrenceRuleDTO struct {
	TaskTitle       string              `json:"task_title"`
	TaskDescription string              `json:"task_description"`
	Recurrence      recurrenceParamsDTO `json:"recurrence"`
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
}

type recurrenceRuleDTO struct {
	ID              int64                           `json:"id"`
	Type            recurrencedomain.RecurrenceType `json:"type"`
	EveryNDays      *int                            `json:"every_n_days,omitempty"`
	DayOfMonth      *int                            `json:"day_of_month,omitempty"`
	SpecificDates   []string                        `json:"specific_dates,omitempty"`
	EvenOdd         recurrencedomain.EvenOddType    `json:"even_odd,omitempty"`
	StartDate       string                          `json:"start_date"`
	EndDate         string                          `json:"end_date"`
	TaskTitle       string                          `json:"task_title"`
	TaskDescription string                          `json:"task_description"`
	CreatedAt       time.Time                       `json:"created_at"`
	UpdatedAt       time.Time                       `json:"updated_at"`
}

func newRecurrenceRuleDTO(rule *recurrencedomain.RecurrenceRule) recurrenceRuleDTO {
	dto := recurrenceRuleDTO{
		ID:              rule.ID,
		Type:            rule.Type,
		EveryNDays:      rule.EveryNDays,
		DayOfMonth:      rule.DayOfMonth,
		EvenOdd:         rule.EvenOdd,
		StartDate:       rule.StartDate.Format("2006-01-02"),
		EndDate:         rule.EndDate.Format("2006-01-02"),
		TaskTitle:       rule.TaskTitle,
		TaskDescription: rule.TaskDescription,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       rule.UpdatedAt,
	}
	if len(rule.SpecificDates) > 0 {
		dto.SpecificDates = make([]string, 0, len(rule.SpecificDates))
		for _, d := range rule.SpecificDates {
			dto.SpecificDates = append(dto.SpecificDates, d.Format("2006-01-02"))
		}
	}
	return dto
}

type createRecurrenceRuleResponseDTO struct {
	Rule  recurrenceRuleDTO `json:"rule"`
	Tasks []taskDTO         `json:"tasks"`
}

type paginatedResponse struct {
	Items any `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
