package task

import (
	"errors"
	"time"
)

type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenDays      RecurrenceType = "even_days"
	RecurrenceOddDays       RecurrenceType = "odd_days"
)

type Recurrence struct {
	Type RecurrenceType `json:"type"`

	IntervalDays *int        `json:"interval_days,omitempty"`
	DaysOfMonth  []int       `json:"days_of_month,omitempty"`
	Dates        []time.Time `json:"dates,omitempty"`
}

func (r *Recurrence) Validate() error {
	if r == nil {
		return nil
	}

	switch r.Type {
	case RecurrenceDaily:
		if r.IntervalDays == nil || *r.IntervalDays <= 0 {
			return errors.New("interval_days must be > 0")
		}

	case RecurrenceMonthly:
		if len(r.DaysOfMonth) == 0 {
			return errors.New("days_of_month required")
		}

	case RecurrenceSpecificDates:
		if len(r.Dates) == 0 {
			return errors.New("dates required")
		}

	case RecurrenceEvenDays, RecurrenceOddDays:
		// ok

	default:
		return errors.New("invalid recurrence type")
	}

	return nil
}