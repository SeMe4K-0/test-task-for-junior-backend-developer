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
	if !r.Type.IsValid() {
		return fmt.Errorf("invalid recurrence type")
	}

	if r.Type == None {
		return nil
	}

	switch r.Type {

	case Daily:
		if r.IntervalDays <= 0 {
			return fmt.Errorf("interval_days must be > 0")
		}
		if r.DayOfMonth != 0 || len(r.SpecificDates) > 0 {
			return fmt.Errorf("invalid fields for daily recurrence")
		}

	case Monthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 31 {
			return fmt.Errorf("day_of_month must be 1-31")
		}
		if r.IntervalDays != 0 || len(r.SpecificDates) > 0 {
			return fmt.Errorf("invalid fields for monthly recurrence")
		}

	case Parity:
		if r.IntervalDays != 0 || r.DayOfMonth != 0 || len(r.SpecificDates) > 0 {
			return fmt.Errorf("invalid fields for parity recurrence")
		}

	case Specific:
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates required")
		}
		if r.IntervalDays != 0 || r.DayOfMonth != 0 {
			return fmt.Errorf("invalid fields for specific recurrence")
		}
	}

	return nil
}