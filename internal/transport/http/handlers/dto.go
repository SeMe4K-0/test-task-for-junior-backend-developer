package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Recurrence  *recurrenceInputDTO `json:"recurrence,omitempty"`
}

type recurrenceInputDTO struct {
	Type          string   `json:"type"`
	IntervalDays  *int     `json:"interval_days,omitempty"`
	DayOfMonth    *int     `json:"day_of_month,omitempty"`
	SpecificDates []string `json:"specific_dates,omitempty"`
	EvenOdd       *string  `json:"even_odd,omitempty"`
	StartDate     string   `json:"start_date"`
	EndDate       *string  `json:"end_date,omitempty"`
}

type taskDTO struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	ParentTaskID  *int64            `json:"parent_task_id,omitempty"`
	ScheduledDate *string           `json:"scheduled_date,omitempty"`
	IsTemplate    bool              `json:"is_template"`
}

type recurrenceDTO struct {
	ID            int64    `json:"id"`
	TaskID        int64    `json:"task_id"`
	Type          string   `json:"type"`
	IntervalDays  *int     `json:"interval_days,omitempty"`
	DayOfMonth    *int     `json:"day_of_month,omitempty"`
	SpecificDates []string `json:"specific_dates,omitempty"`
	EvenOdd       *string  `json:"even_odd,omitempty"`
	StartDate     string   `json:"start_date"`
	EndDate       *string  `json:"end_date,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		ParentTaskID: task.ParentTaskID,
		IsTemplate:   task.IsTemplate,
	}

	if task.ScheduledDate != nil {
		s := task.ScheduledDate.Format("2006-01-02")
		dto.ScheduledDate = &s
	}

	return dto
}

func newRecurrenceDTO(rule *taskdomain.RecurrenceRule) recurrenceDTO {
	dto := recurrenceDTO{
		ID:        rule.ID,
		TaskID:    rule.TaskID,
		Type:      string(rule.Type),
		IntervalDays:  rule.IntervalDays,
		DayOfMonth:    rule.DayOfMonth,
		StartDate: rule.StartDate.Format("2006-01-02"),
	}

	if rule.EvenOdd != nil {
		s := string(*rule.EvenOdd)
		dto.EvenOdd = &s
	}

	if rule.EndDate != nil {
		s := rule.EndDate.Format("2006-01-02")
		dto.EndDate = &s
	}

	for _, d := range rule.SpecificDates {
		dto.SpecificDates = append(dto.SpecificDates, d.Format("2006-01-02"))
	}

	return dto
}
