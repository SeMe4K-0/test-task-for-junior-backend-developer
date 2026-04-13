package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type PeriodType string

const (
	PeriodNone          PeriodType = "none"
	PeriodEveryNDays    PeriodType = "every_n_days"
	PeriodMonthlyDay    PeriodType = "monthly_day"
	PeriodSpecificDates PeriodType = "specific_dates"
	PeriodEvenDays      PeriodType = "even_days"
	PeriodOddDays       PeriodType = "odd_days"
)

type Repeated struct {
	Type          PeriodType  `json:"type"`
	EveryNDays    int         `json:"every_n_days,omitempty"`
	DayOfMonth    int         `json:"day_of_month,omitempty"`
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
}

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Repeated    Repeated  `json:"repeated"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (p PeriodType) Valid() bool {
	switch p {
	case PeriodNone, PeriodEveryNDays, PeriodMonthlyDay, PeriodSpecificDates, PeriodEvenDays, PeriodOddDays:
		return true
	default:
		return false
	}
}

func (r Repeated) Valid() bool {
	if !r.Type.Valid() {
		return false
	}

	switch r.Type {
	case PeriodNone:
		return true

	case PeriodEveryNDays:
		return r.EveryNDays > 0

	case PeriodMonthlyDay:
		return r.DayOfMonth >= 1 && r.DayOfMonth <= 30

	case PeriodSpecificDates:
		return len(r.SpecificDates) > 0

	case PeriodEvenDays, PeriodOddDays:
		return true

	default:
		return false
	}
}
