package recurrence

import "time"

type RecurrenceType string

const (
	TypeDaily         RecurrenceType = "daily"
	TypeMonthly       RecurrenceType = "monthly"
	TypeSpecificDates RecurrenceType = "specific_dates"
	TypeEvenOdd       RecurrenceType = "even_odd"
)

func (t RecurrenceType) Valid() bool {
	switch t {
	case TypeDaily, TypeMonthly, TypeSpecificDates, TypeEvenOdd:
		return true
	default:
		return false
	}
}

type EvenOddType string

const (
	EvenDays EvenOddType = "even"
	OddDays  EvenOddType = "odd"
)

func (e EvenOddType) Valid() bool {
	switch e {
	case EvenDays, OddDays:
		return true
	default:
		return false
	}
}

type RecurrenceRule struct {
	ID              int64          `json:"id"`
	Type            RecurrenceType `json:"type"`
	EveryNDays      *int           `json:"every_n_days,omitempty"`
	DayOfMonth      *int           `json:"day_of_month,omitempty"`
	SpecificDates   []time.Time    `json:"specific_dates,omitempty"`
	EvenOdd         EvenOddType    `json:"even_odd,omitempty"`
	StartDate       time.Time      `json:"start_date"`
	EndDate         time.Time      `json:"end_date"`
	TaskTitle       string         `json:"task_title"`
	TaskDescription string         `json:"task_description"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}
