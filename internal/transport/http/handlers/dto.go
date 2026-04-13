package handlers

import (
	"time"
	"encoding/json"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	Status        taskdomain.Status  `json:"status"`
	ExecutionDate time.Time          `json:"execution_date"`  
	Recurrence    *recurrenceRuleDTO `json:"recurrence,omitempty"`  
}

type taskUpdateDTO struct {
    Title         string            `json:"title"`
    Description   string            `json:"description"`
    Status        taskdomain.Status `json:"status"`
    ExecutionDate time.Time         `json:"execution_date"`
}

type taskDTO struct {
	ID            *int64            `json:"id"`         
	IsVirtual     bool              `json:"is_virtual"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	ExecutionDate time.Time         `json:"execution_date"` 
	RuleID        *int64            `json:"rule_id,omitempty"`
	CreatedAt     *time.Time        `json:"created_at"`
	UpdatedAt     *time.Time        `json:"updated_at"`
}

type taskRuleDTO struct {
    ID              int64                     `json:"id"`
    Title           string                    `json:"title"`
    Description     string                    `json:"description"`
    RecurrenceType  taskdomain.RecurrenceType `json:"recurrence_type"`
    RecurrenceValue json.RawMessage           `json:"recurrence_value"`
    StartDate       time.Time                 `json:"start_date"`
    EndDate         *time.Time                `json:"end_date,omitempty"`
    CreatedAt       time.Time                 `json:"created_at"`
}

type recurrenceRuleDTO struct {
	Type      taskdomain.RecurrenceType `json:"type"`             
	Value     json.RawMessage           `json:"value"`            
	StartDate time.Time                 `json:"start_date"`       
	EndDate   *time.Time                `json:"end_date,omitempty"` 
}

type materializeDTO struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	ExecutionDate time.Time         `json:"execution_date"`    
	RuleID        int64             `json:"rule_id"` 
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		Title:         task.Title,
		Description:   task.Description,
		Status:        task.Status,
		ExecutionDate: task.ExecutionDate,
		RuleID:        task.RuleID,
	}

	if task.ID == 0 {
		dto.IsVirtual = true
		dto.ID = nil
		dto.CreatedAt = nil
		dto.UpdatedAt = nil
		return dto
	}

	id := task.ID
	createdAt := task.CreatedAt
	updatedAt := task.UpdatedAt

	dto.ID = &id
	dto.IsVirtual = false
	dto.CreatedAt = &createdAt
	dto.UpdatedAt = &updatedAt

	return dto
}

func newTaskRuleDTO(rule taskdomain.TaskRule) taskRuleDTO {
    return taskRuleDTO{
        ID:              rule.ID,
        Title:           rule.Title,
        Description:     rule.Description,
        RecurrenceType:  rule.RecurrenceType,
        RecurrenceValue: json.RawMessage(rule.RecurrenceValue),
        StartDate:       rule.StartDate,
        EndDate:         rule.EndDate,
        CreatedAt:       rule.CreatedAt,
    }
}
