package task

import (
	"fmt"
	"time"
)

type RecurrenceType string

const (
	// RecurrenceTypeDaily repeats every N days starting from start_date.
	RecurrenceTypeDaily RecurrenceType = "daily"
	// RecurrenceTypeMonthly repeats on the specified day-numbers within each month.
	RecurrenceTypeMonthly RecurrenceType = "monthly"
	// RecurrenceTypeDates creates tasks only on the explicitly listed dates.
	RecurrenceTypeDates RecurrenceType = "dates"
	// RecurrenceTypeEvenOdd creates tasks on even or odd calendar days of the month.
	RecurrenceTypeEvenOdd RecurrenceType = "even_odd"
)

type EvenOddParity string

const (
	ParityEven EvenOddParity = "even"
	ParityOdd  EvenOddParity = "odd"
)

// MaxDateRangeDays is the upper bound on the start→end window to prevent
// accidentally generating thousands of tasks in one request.
const MaxDateRangeDays = 365

// Recurrence describes how a task repeats over time.
//
// Only the fields relevant to Type are used; others are ignored.
type Recurrence struct {
	Type RecurrenceType `json:"type"`

	// Interval is used with RecurrenceTypeDaily: repeat every Interval days (>= 1).
	Interval int `json:"interval,omitempty"`

	// MonthDays is used with RecurrenceTypeMonthly: list of day-of-month values (1–30).
	// Days that do not exist in a given month (e.g., 30 in February) are silently skipped.
	MonthDays []int `json:"month_days,omitempty"`

	// Dates is used with RecurrenceTypeDates: the exact dates to schedule tasks on.
	// start_date / end_date in the request are ignored for this type.
	Dates []time.Time `json:"dates,omitempty"`

	// Parity is used with RecurrenceTypeEvenOdd: "even" or "odd" days of the month.
	Parity EvenOddParity `json:"parity,omitempty"`
}

// Validate checks that the recurrence settings are internally consistent.
func (r *Recurrence) Validate() error {
	switch r.Type {
	case RecurrenceTypeDaily:
		if r.Interval < 1 {
			return fmt.Errorf("daily recurrence: interval must be >= 1")
		}

	case RecurrenceTypeMonthly:
		if len(r.MonthDays) == 0 {
			return fmt.Errorf("monthly recurrence: month_days must not be empty")
		}
		seen := make(map[int]struct{}, len(r.MonthDays))
		for _, d := range r.MonthDays {
			if d < 1 || d > 30 {
				return fmt.Errorf("monthly recurrence: month_day %d is out of range [1, 30]", d)
			}
			if _, dup := seen[d]; dup {
				return fmt.Errorf("monthly recurrence: duplicate month_day %d", d)
			}
			seen[d] = struct{}{}
		}

	case RecurrenceTypeDates:
		if len(r.Dates) == 0 {
			return fmt.Errorf("dates recurrence: dates list must not be empty")
		}
		if len(r.Dates) > MaxDateRangeDays {
			return fmt.Errorf("dates recurrence: too many dates (max %d)", MaxDateRangeDays)
		}

	case RecurrenceTypeEvenOdd:
		if r.Parity != ParityEven && r.Parity != ParityOdd {
			return fmt.Errorf("even_odd recurrence: parity must be %q or %q", ParityEven, ParityOdd)
		}

	default:
		return fmt.Errorf("unknown recurrence type: %q", r.Type)
	}

	return nil
}

// GenerateDates returns the set of dates on which a task instance should be created.
//
// start and end are inclusive and define the scheduling window.
// For RecurrenceTypeDates the window is ignored — r.Dates is used directly.
func (r *Recurrence) GenerateDates(start, end time.Time) []time.Time {
	switch r.Type {
	case RecurrenceTypeDaily:
		return generateDailyDates(start, end, r.Interval)
	case RecurrenceTypeMonthly:
		return generateMonthlyDates(start, end, r.MonthDays)
	case RecurrenceTypeDates:
		return r.Dates
	case RecurrenceTypeEvenOdd:
		return generateEvenOddDates(start, end, r.Parity)
	default:
		return nil
	}
}

func generateDailyDates(start, end time.Time, interval int) []time.Time {
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, interval) {
		dates = append(dates, d)
	}
	return dates
}

func generateMonthlyDates(start, end time.Time, monthDays []int) []time.Time {
	daySet := make(map[int]struct{}, len(monthDays))
	for _, d := range monthDays {
		daySet[d] = struct{}{}
	}

	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if _, ok := daySet[d.Day()]; ok {
			dates = append(dates, d)
		}
	}
	return dates
}

func generateEvenOddDates(start, end time.Time, parity EvenOddParity) []time.Time {
	var dates []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		isEven := d.Day()%2 == 0
		if (parity == ParityEven && isEven) || (parity == ParityOdd && !isEven) {
			dates = append(dates, d)
		}
	}
	return dates
}
