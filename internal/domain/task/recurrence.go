package task

import (
	"fmt"
	"time"
)

type RecurrenceType string

const (
	None     RecurrenceType = "none"
	Daily    RecurrenceType = "daily"
	Monthly  RecurrenceType = "monthly"
	Parity   RecurrenceType = "parity"
	Specific RecurrenceType = "specific"
)

type Recurrence struct {
	Type          RecurrenceType `json:"type"`
	IntervalDays  int            `json:"interval_days,omitempty"`
	DayOfMonth    int            `json:"day_of_month,omitempty"`
	ParityEven    bool           `json:"parity_even,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
}

func (r RecurrenceType) IsValid() bool {
	switch r {
	case None, Daily, Monthly, Parity, Specific:
		return true
	default:
		return false
	}
}

func (r Recurrence) Validate() error {
	switch r.Type {

	case None:
		return nil

	case Daily:
		if r.IntervalDays <= 0 {
			return fmt.Errorf("interval_days must be > 0")
		}

	case Monthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 30 {
			return fmt.Errorf("day_of_month must be in range 1-30")
		}

	case Specific:
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates is required")
		}

		for _, d := range r.SpecificDates {
			if d.IsZero() {
				return fmt.Errorf("specific_dates contains invalid date")
			}
		}

	case Parity:
		return nil

	default:
		return fmt.Errorf("unknown recurrence type")
	}

	return nil
}
