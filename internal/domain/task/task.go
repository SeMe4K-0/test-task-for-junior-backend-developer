package task

import (
    "errors"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type PeriodType string

const (
	PeriodDaily    PeriodType = "daily"
	PeriodMonthly  PeriodType = "monthly"
	PeriodSpecific PeriodType = "specific"
	PeriodEvenOdd  PeriodType = "even_odd"
)

type EvenOddType string

const (
	Even EvenOddType = "even"
	Odd  EvenOddType = "odd"
)

type PeriodConf struct {
	Type          PeriodType  `json:"type"`
	IntervalDays  int         `json:"interval_days,omitempty"`
	MonthDay      int         `json:"month_day,omitempty"`
	SpecificDates []time.Time `json:"specific_dates,omitempty"`
	EvenOddType   EvenOddType `json:"even_odd_type,omitempty"`
}

type Task struct {
	ID            int64       `json:"id"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Status        Status      `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	PeriodConf    *PeriodConf `json:"period_conf,omitempty"`
	ScheduledDate *time.Time  `json:"scheduled_date,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

func (pc *PeriodConf) Validate() error {
	if pc == nil {
		return nil
	}
	switch pc.Type {
	case PeriodDaily:
		if pc.IntervalDays < 1 {
			return ErrInvalidPeriodConfig
		}
	case PeriodMonthly:
		if pc.MonthDay < 1 || pc.MonthDay > 30 {
			return ErrInvalidPeriodConfig
		}
	case PeriodSpecific:
		if len(pc.SpecificDates) == 0 {
			return ErrInvalidPeriodConfig
		}
	case PeriodEvenOdd:
		if pc.EvenOddType != Even && pc.EvenOddType != Odd {
			return ErrInvalidPeriodConfig
		}
	default:
		return ErrInvalidPeriodConfig
	}
	return nil
}

var ErrInvalidPeriodConfig = errors.New("invalid period configuration")
