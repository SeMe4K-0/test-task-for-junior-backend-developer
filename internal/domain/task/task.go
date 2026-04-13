package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        Status     `json:"status"`
	ExecutionDate time.Time  `json:"execution_date"` 
	RuleID        *int64     `json:"rule_id"`        
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenDays      RecurrenceType = "even_days"
	RecurrenceOddDays       RecurrenceType = "odd_days"
)

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenDays, RecurrenceOddDays:
		return true
	default:
		return false
	}
}

type TaskRule struct {
	ID              int64          `json:"id"`
	Title           string         `json:"title"`        
	Description     string         `json:"description"`
	RecurrenceType  RecurrenceType `json:"recurrence_type"`
	RecurrenceValue []byte         `json:"recurrence_value"` 
	StartDate       time.Time      `json:"start_date"`
	EndDate         *time.Time     `json:"end_date"` 
	CreatedAt       time.Time      `json:"created_at"`
}

