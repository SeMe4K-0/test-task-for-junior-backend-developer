package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceParity        RecurrenceType = "parity"
)

type RecurrenceConfig struct {
	Type      RecurrenceType `json:"type"`
	Interval  *int           `json:"interval,omitempty"` // For "daily" (every n-th day)
	Day       *int           `json:"day,omitempty"`      // For "monthly" (1-30)
	Dates     []time.Time    `json:"dates,omitempty"`    // For "specific_dates"
	Parity    *string        `json:"parity,omitempty"`   // For "parity" ("even", "odd")
	StartDate time.Time      `json:"start_date"`
	EndDate   time.Time      `json:"end_date"`
}

type Task struct {
	ID               int64             `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Status           Status            `json:"status"`
	DueDate          *time.Time        `json:"due_date"`
	IsTemplate       bool              `json:"is_template"`
	ParentID         *int64            `json:"parent_id"`
	RecurrenceConfig *RecurrenceConfig `json:"recurrence_config"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
