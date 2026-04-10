package task

import "time"

type Status string

type RecurrenceType string

type Parity string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"

	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeParity        RecurrenceType = "parity"

	ParityEven Parity = "even"
	ParityOdd  Parity = "odd"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	IntervalDays  int            `json:"interval_days,omitempty"`
	MonthDay      int            `json:"month_day,omitempty"`
	SpecificDates []string       `json:"specific_dates,omitempty"`
	Parity        Parity         `json:"parity,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (r RecurrenceType) Valid() bool {
	switch r {
	case RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeParity:
		return true
	default:
		return false
	}
}

func (p Parity) Valid() bool {
	switch p {
	case ParityEven, ParityOdd:
		return true
	default:
		return false
	}
}
