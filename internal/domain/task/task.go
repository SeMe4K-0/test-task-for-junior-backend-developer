package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Periodicity string

const (
	PeriodicityOnce     Periodicity = "once"
	PeriodicitySetDates Periodicity = "set_dates"
	PeriodicityInterval Periodicity = "interval"
	PeriodicityMonthly  Periodicity = "monthly"
	PeriodicityParity   Periodicity = "parity"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Periodicity Periodicity `json:"periodicity"`
	ScheduledAt []time.Time `json:"scheduled_at"`
	Status      Status      `json:"status"`
	Frequency   int64       `json:"frequency"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (p Periodicity) Valid() bool {
	switch p {
	case PeriodicityOnce, PeriodicitySetDates, PeriodicityInterval, PeriodicityMonthly, PeriodicityParity:
		return true
	default:
		return false
	}
}
