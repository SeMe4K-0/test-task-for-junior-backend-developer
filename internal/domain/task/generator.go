package task

import (
	"encoding/json"
	"time"
)

func MidnightUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

type DailyRecurrence struct {
	Interval int `json:"interval"`
}

type MonthlyRecurrence struct {
	Days []int `json:"days"`
}

type SpecificDatesRecurrence struct {
	Dates []string `json:"dates"`
}

func GenerateDatesForRule(rule TaskRule, from, to time.Time) []time.Time {
	var dates []time.Time

	start := MidnightUTC(from)
	if MidnightUTC(rule.StartDate).After(start) {
		start = MidnightUTC(rule.StartDate)
	}

	end := MidnightUTC(to)
	if rule.EndDate != nil && MidnightUTC(*rule.EndDate).Before(end) {
		end = MidnightUTC(*rule.EndDate)
	}

	if start.After(end) {
		return dates
	}

	var (
		dailyInterval  int
		monthlyDays    map[int]bool
		specificDatesM map[string]bool
	)

	switch rule.RecurrenceType {
	case RecurrenceDaily:
		var p DailyRecurrence
		if err := json.Unmarshal(rule.RecurrenceValue, &p); err == nil && p.Interval > 0 {
			dailyInterval = p.Interval
		} else {
			dailyInterval = 1
		}
	case RecurrenceMonthly:
		var p MonthlyRecurrence
		monthlyDays = make(map[int]bool)
		if err := json.Unmarshal(rule.RecurrenceValue, &p); err == nil {
			for _, d := range p.Days {
				monthlyDays[d] = true
			}
		}
	case RecurrenceSpecificDates:
		var p SpecificDatesRecurrence
		specificDatesM = make(map[string]bool)
		if err := json.Unmarshal(rule.RecurrenceValue, &p); err == nil {
			for _, dStr := range p.Dates {
				specificDatesM[dStr] = true
			}
		}
	}

	ruleStart := MidnightUTC(rule.StartDate)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dDate := MidnightUTC(d)

		switch rule.RecurrenceType {
		case RecurrenceEvenDays:
			if dDate.Day()%2 == 0 {
				dates = append(dates, dDate)
			}
		case RecurrenceOddDays:
			if dDate.Day()%2 != 0 {
				dates = append(dates, dDate)
			}
		case RecurrenceDaily:
			daysDiff := int(dDate.Sub(ruleStart).Hours() / 24)
			if daysDiff >= 0 && daysDiff%dailyInterval == 0 {
				dates = append(dates, dDate)
			}
		case RecurrenceMonthly:
			if monthlyDays[dDate.Day()] {
				dates = append(dates, dDate)
			}
		case RecurrenceSpecificDates:
			if specificDatesM[dDate.Format("2006-01-02")] {
				dates = append(dates, dDate)
			}
		}
	}

	return dates
}