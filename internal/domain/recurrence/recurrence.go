package recurrence

import "time"

type EvenOdd string

const (
	Even EvenOdd = "even"
	Odd  EvenOdd = "odd"
)

type Recurrence struct {
	ID            int          `json:"id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	StartDate     *time.Time   `json:"start_date,omitempty"`
	EndDate       *time.Time   `json:"end_date,omitempty"`
	IntervalDays  *int         `json:"interval_days,omitempty"`
	MonthDays     *[]int       `json:"month_days,omitempty"`
	SpecificDates *[]time.Time `json:"specific_dates,omitempty"`
	EvenOdd       *EvenOdd     `json:"even_odd,omitempty"`
	Active        bool         `json:"active"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func New(
	title, description string,
	startDate *time.Time,
	endDate *time.Time,
	intervalDays *int,
	monthDays *[]int,
	specificDates *[]time.Time,
	evenOdd *EvenOdd,
) Recurrence {
	sd := time.Now().UTC()

	if startDate != nil {
		sd = startDate.UTC()
	}

	return Recurrence{
		Title:         title,
		Description:   description,
		StartDate:     &sd,
		EndDate:       endDate,
		IntervalDays:  intervalDays,
		MonthDays:     monthDays,
		SpecificDates: specificDates,
		EvenOdd:       evenOdd,
		Active:        true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (r Recurrence) Matches(date time.Time) error {
	date = truncateToDay(date)
	start := truncateToDay(*r.StartDate)

	if date.Before(start) {
		return ErrDateOutOfRange
	}

	if r.EndDate != nil {
		end := truncateToDay(*r.EndDate)
		if date.After(end) {
			return ErrDateOutOfRange
		}
	}

	switch {
	case r.IntervalDays != nil:
		return r.matchesInterval(date, start)
	case r.MonthDays != nil:
		return r.matchesMonthDays(date)
	case r.SpecificDates != nil:
		return r.matchesSpecificDates(date)
	case r.EvenOdd != nil:
		return r.matchesEvenOdd(date)
	}

	return ErrDateNotMatched
}

func (r Recurrence) matchesInterval(date, start time.Time) error {
	if *r.IntervalDays <= 0 {
		return ErrDateNotMatched
	}
	daysDiff := int(date.Sub(start).Hours() / 24)
	if daysDiff%*r.IntervalDays != 0 {
		return ErrDateNotMatched
	}
	return nil
}

func (r Recurrence) matchesMonthDays(date time.Time) error {
	day := date.Day()
	for _, d := range *r.MonthDays {
		if d == day {
			return nil
		}
	}
	return ErrDateNotMatched
}

func (r Recurrence) matchesSpecificDates(date time.Time) error {
	for _, sd := range *r.SpecificDates {
		if truncateToDay(sd).Equal(date) {
			return nil
		}
	}
	return ErrDateNotMatched
}

func (r Recurrence) matchesEvenOdd(date time.Time) error {
	day := date.Day()
	isEven := day%2 == 0
	switch *r.EvenOdd {
	case Even:
		if isEven {
			return nil
		}
	case Odd:
		if !isEven {
			return nil
		}
	}
	return ErrDateNotMatched
}

func truncateToDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}
