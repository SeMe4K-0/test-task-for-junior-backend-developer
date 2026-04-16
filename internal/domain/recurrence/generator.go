package recurrence

import (
	"fmt"
	"time"
)

const MaxGeneratedTasks = 1000

func GenerateDates(rule *RecurrenceRule) ([]time.Time, error) {
	if rule.StartDate.After(rule.EndDate) {
		return nil, fmt.Errorf("start_date must not be after end_date")
	}

	var dates []time.Time

	switch rule.Type {
	case TypeDaily:
		dates = generateDaily(rule)
	case TypeMonthly:
		dates = generateMonthly(rule)
	case TypeSpecificDates:
		dates = generateSpecificDates(rule)
	case TypeEvenOdd:
		dates = generateEvenOdd(rule)
	default:
		return nil, fmt.Errorf("unknown recurrence type: %s", rule.Type)
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("no tasks would be generated for the given parameters")
	}

	if len(dates) > MaxGeneratedTasks {
		return nil, fmt.Errorf("recurrence would generate %d tasks, maximum allowed is %d", len(dates), MaxGeneratedTasks)
	}

	return dates, nil
}

func generateDaily(rule *RecurrenceRule) []time.Time {
	n := 1
	if rule.EveryNDays != nil && *rule.EveryNDays > 0 {
		n = *rule.EveryNDays
	}

	var dates []time.Time
	for d := rule.StartDate; !d.After(rule.EndDate); d = d.AddDate(0, 0, n) {
		dates = append(dates, d)
	}
	return dates
}

func generateMonthly(rule *RecurrenceRule) []time.Time {
	if rule.DayOfMonth == nil {
		return nil
	}
	targetDay := *rule.DayOfMonth

	var dates []time.Time
	y, m, _ := rule.StartDate.Date()
	loc := rule.StartDate.Location()

	for {
		lastDay := daysInMonth(y, m)
		day := targetDay
		if day > lastDay {
			day = lastDay
		}
		candidate := time.Date(y, m, day, 0, 0, 0, 0, loc)

		if candidate.After(rule.EndDate) {
			break
		}
		if !candidate.Before(rule.StartDate) {
			dates = append(dates, candidate)
		}

		m++
		if m > 12 {
			m = 1
			y++
		}
	}
	return dates
}

func generateSpecificDates(rule *RecurrenceRule) []time.Time {
	var dates []time.Time
	for _, d := range rule.SpecificDates {
		if !d.Before(rule.StartDate) && !d.After(rule.EndDate) {
			dates = append(dates, d)
		}
	}
	return dates
}

func generateEvenOdd(rule *RecurrenceRule) []time.Time {
	wantEven := rule.EvenOdd == EvenDays

	var dates []time.Time
	for d := rule.StartDate; !d.After(rule.EndDate); d = d.AddDate(0, 0, 1) {
		day := d.Day()
		isEven := day%2 == 0
		if isEven == wantEven {
			dates = append(dates, d)
		}
	}
	return dates
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
