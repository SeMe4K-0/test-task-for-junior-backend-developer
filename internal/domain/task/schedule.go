package task

import (
	"fmt"
	"slices"
	"sort"
	"time"
)

type ScheduleType string

const (
	ScheduleDaily    ScheduleType = "daily"
	ScheduleMonthly  ScheduleType = "monthly"
	ScheduleSpecific ScheduleType = "specific_dates"
	ScheduleOddDays  ScheduleType = "odd_days"
	ScheduleEvenDays ScheduleType = "even_days"
)

type Schedule struct {
	Type       ScheduleType `json:"type"`
	EveryDays  int          `json:"every_days,omitempty"`
	DayOfMonth int          `json:"day_of_month,omitempty"`
	Dates      []Date       `json:"dates,omitempty"`
	StartDate  *Date        `json:"start_date,omitempty"`
	TimeZone   string       `json:"timezone,omitempty"`
}

func (s *Schedule) Validate() error {
	if s == nil {
		return nil
	}

	if _, err := s.location(); err != nil {
		return fmt.Errorf("invalid timezone %q", s.TimeZone)
	}

	switch s.Type {
	case ScheduleDaily:
		if s.EveryDays <= 0 {
			return fmt.Errorf("every_days must be greater than zero")
		}
		if s.DayOfMonth != 0 || len(s.Dates) > 0 {
			return fmt.Errorf("daily schedule accepts only every_days and optional start_date")
		}
	case ScheduleMonthly:
		if s.DayOfMonth < 1 || s.DayOfMonth > 30 {
			return fmt.Errorf("day_of_month must be between 1 and 30")
		}
		if s.EveryDays != 0 || len(s.Dates) > 0 {
			return fmt.Errorf("monthly schedule accepts only day_of_month and optional start_date")
		}
	case ScheduleSpecific:
		if len(s.Dates) == 0 {
			return fmt.Errorf("dates must not be empty")
		}
		if s.EveryDays != 0 || s.DayOfMonth != 0 || s.StartDate != nil {
			return fmt.Errorf("specific_dates schedule accepts only dates")
		}
	case ScheduleOddDays, ScheduleEvenDays:
		if s.EveryDays != 0 || s.DayOfMonth != 0 || len(s.Dates) > 0 {
			return fmt.Errorf("odd/even schedule accepts only optional start_date")
		}
	default:
		return fmt.Errorf("unsupported schedule type")
	}

	if len(s.Dates) > 1 {
		copied := append([]Date(nil), s.Dates...)
		sort.Slice(copied, func(i, j int) bool {
			return copied[i].Before(copied[j].Time)
		})

		for i := 1; i < len(copied); i++ {
			if copied[i].Equal(copied[i-1].Time) {
				return fmt.Errorf("dates must be unique")
			}
		}

		s.Dates = copied
	}

	return nil
}

func (s *Schedule) OccursOn(date Date, createdAt time.Time) bool {
	if s == nil {
		return false
	}

	switch s.Type {
	case ScheduleDaily:
		start := s.resolveStartDate(createdAt)
		if date.Before(start.Time) {
			return false
		}

		diffDays := int(date.Sub(start.Time).Hours() / 24)
		return diffDays%s.EveryDays == 0
	case ScheduleMonthly:
		start := s.resolveStartDate(createdAt)
		return !date.Before(start.Time) && date.Day() == s.DayOfMonth
	case ScheduleSpecific:
		return slices.Contains(s.Dates, date)
	case ScheduleOddDays:
		start := s.resolveStartDate(createdAt)
		return !date.Before(start.Time) && date.Day()%2 == 1
	case ScheduleEvenDays:
		start := s.resolveStartDate(createdAt)
		return !date.Before(start.Time) && date.Day()%2 == 0
	default:
		return false
	}
}

func (s *Schedule) OccurrencesBetween(from, to Date, createdAt time.Time) []Date {
	if s == nil || to.Before(from.Time) {
		return nil
	}

	switch s.Type {
	case ScheduleDaily:
		return s.dailyOccurrencesBetween(from, to, createdAt)
	case ScheduleMonthly:
		return s.monthlyOccurrencesBetween(from, to, createdAt)
	case ScheduleSpecific:
		return s.specificOccurrencesBetween(from, to)
	case ScheduleOddDays, ScheduleEvenDays:
		return s.oddEvenOccurrencesBetween(from, to, createdAt)
	default:
		return nil
	}
}

func (s *Schedule) NextOccurrence(from Date, createdAt time.Time) *Date {
	if s == nil {
		return nil
	}

	switch s.Type {
	case ScheduleDaily:
		start := s.resolveStartDate(createdAt)
		if from.Before(start.Time) {
			next := start
			return &next
		}

		diffDays := int(from.Sub(start.Time).Hours() / 24)
		remainder := diffDays % s.EveryDays
		if remainder == 0 {
			next := from
			return &next
		}

		next := from.AddDays(s.EveryDays - remainder)
		return &next
	case ScheduleMonthly:
		start := s.resolveStartDate(createdAt)
		cursor := from
		if cursor.Before(start.Time) {
			cursor = start
		}

		year, month := cursor.Year(), cursor.Month()
		for i := 0; i < 24; i++ {
			candidate := NewDate(time.Date(year, month, s.DayOfMonth, 0, 0, 0, 0, time.UTC))
			if candidate.Month() == month && !candidate.Before(start.Time) && !candidate.Before(cursor.Time) {
				return &candidate
			}
			month++
			if month > time.December {
				year++
				month = time.January
			}
		}
	case ScheduleSpecific:
		for i := range s.Dates {
			if !s.Dates[i].Before(from.Time) {
				next := s.Dates[i]
				return &next
			}
		}
	case ScheduleOddDays, ScheduleEvenDays:
		start := s.resolveStartDate(createdAt)
		cursor := from
		if cursor.Before(start.Time) {
			cursor = start
		}

		for i := 0; i < 62; i++ {
			if s.OccursOn(cursor, createdAt) {
				next := cursor
				return &next
			}
			cursor = cursor.AddDays(1)
		}
	}

	return nil
}

func (s *Schedule) resolveStartDate(createdAt time.Time) Date {
	if s.StartDate != nil {
		return *s.StartDate
	}

	return s.DateFromTime(createdAt)
}

func (s *Schedule) DateFromTime(ts time.Time) Date {
	location, err := s.location()
	if err != nil {
		return NewDate(ts)
	}

	return NewDateInLocation(ts, location)
}

func (s *Schedule) location() (*time.Location, error) {
	if s == nil || s.TimeZone == "" {
		return time.UTC, nil
	}

	return time.LoadLocation(s.TimeZone)
}

func (s *Schedule) dailyOccurrencesBetween(from, to Date, createdAt time.Time) []Date {
	start := s.resolveStartDate(createdAt)
	if to.Before(start.Time) {
		return nil
	}

	cursor := from
	if cursor.Before(start.Time) {
		cursor = start
	}

	diffDays := int(cursor.Sub(start.Time).Hours() / 24)
	remainder := diffDays % s.EveryDays
	if remainder != 0 {
		cursor = cursor.AddDays(s.EveryDays - remainder)
	}

	occurrences := make([]Date, 0)
	for !cursor.After(to.Time) {
		occurrences = append(occurrences, cursor)
		cursor = cursor.AddDays(s.EveryDays)
	}

	return occurrences
}

func (s *Schedule) monthlyOccurrencesBetween(from, to Date, createdAt time.Time) []Date {
	start := s.resolveStartDate(createdAt)
	if to.Before(start.Time) {
		return nil
	}

	cursor := from
	if cursor.Before(start.Time) {
		cursor = start
	}

	occurrences := make([]Date, 0)
	year, month := cursor.Year(), cursor.Month()
	for !NewDateFromParts(year, month, 1).After(to.Time) {
		candidate := NewDate(time.Date(year, month, s.DayOfMonth, 0, 0, 0, 0, time.UTC))
		if candidate.Month() == month && !candidate.Before(start.Time) && !candidate.Before(from.Time) && !candidate.After(to.Time) {
			occurrences = append(occurrences, candidate)
		}

		month++
		if month > time.December {
			year++
			month = time.January
		}
	}

	return occurrences
}

func (s *Schedule) specificOccurrencesBetween(from, to Date) []Date {
	occurrences := make([]Date, 0)
	for i := range s.Dates {
		if s.Dates[i].Before(from.Time) || s.Dates[i].After(to.Time) {
			continue
		}
		occurrences = append(occurrences, s.Dates[i])
	}

	return occurrences
}

func (s *Schedule) oddEvenOccurrencesBetween(from, to Date, createdAt time.Time) []Date {
	start := s.resolveStartDate(createdAt)
	if to.Before(start.Time) {
		return nil
	}

	cursor := from
	if cursor.Before(start.Time) {
		cursor = start
	}

	occurrences := make([]Date, 0)
	for !cursor.After(to.Time) {
		if s.OccursOn(cursor, createdAt) {
			occurrences = append(occurrences, cursor)
		}
		cursor = cursor.AddDays(1)
	}

	return occurrences
}
