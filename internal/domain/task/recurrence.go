package task

import (
	"fmt"
	"sort"
	"time"
)

// RecurrenceType — дискриминатор типа, хранится в колонке recurrence_type.
type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthlyDays   RecurrenceType = "monthly_days"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceDayParity     RecurrenceType = "day_parity"
)

// Parity указывает, выбираются чётные или нечётные числа.
type Parity string

const (
	ParityEven Parity = "even"
	ParityOdd  Parity = "odd"
)

// ValidateRecurrence проверяет корректность полей периодичности задачи.
// Проверяет не только собственные поля типа, но и отсутствие полей чужих типов.
func ValidateRecurrence(t *Task) error {
	if t.RecurrenceType == nil {
		if t.RecurrenceDailyInterval != nil || len(t.RecurrenceMonthlyDays) > 0 ||
			len(t.RecurrenceSpecificDates) > 0 || t.RecurrenceDayParity != nil {
			return fmt.Errorf("recurrence params must be nil when recurrence_type is nil")
		}
		return nil
	}

	switch *t.RecurrenceType {
	case RecurrenceDaily:
		if t.RecurrenceDailyInterval == nil {
			return fmt.Errorf("daily: interval is required")
		}
		if *t.RecurrenceDailyInterval < 1 {
			return fmt.Errorf("daily: interval must be >= 1")
		}
		if len(t.RecurrenceMonthlyDays) > 0 || len(t.RecurrenceSpecificDates) > 0 || t.RecurrenceDayParity != nil {
			return fmt.Errorf("daily: unexpected params for other recurrence types")
		}

	case RecurrenceMonthlyDays:
		if len(t.RecurrenceMonthlyDays) == 0 {
			return fmt.Errorf("monthly_days: days must not be empty")
		}
		seen := make(map[int]struct{}, len(t.RecurrenceMonthlyDays))
		for _, d := range t.RecurrenceMonthlyDays {
			if d < 1 || d > 30 {
				return fmt.Errorf("monthly_days: day %d is out of range 1..30", d)
			}
			if _, ok := seen[d]; ok {
				return fmt.Errorf("monthly_days: duplicate day %d", d)
			}
			seen[d] = struct{}{}
		}
		if t.RecurrenceDailyInterval != nil || len(t.RecurrenceSpecificDates) > 0 || t.RecurrenceDayParity != nil {
			return fmt.Errorf("monthly_days: unexpected params for other recurrence types")
		}

	case RecurrenceSpecificDates:
		if len(t.RecurrenceSpecificDates) == 0 {
			return fmt.Errorf("specific_dates: dates must not be empty")
		}
		if t.RecurrenceDailyInterval != nil || len(t.RecurrenceMonthlyDays) > 0 || t.RecurrenceDayParity != nil {
			return fmt.Errorf("specific_dates: unexpected params for other recurrence types")
		}

	case RecurrenceDayParity:
		if t.RecurrenceDayParity == nil {
			return fmt.Errorf("day_parity: parity is required")
		}
		if *t.RecurrenceDayParity != ParityEven && *t.RecurrenceDayParity != ParityOdd {
			return fmt.Errorf("day_parity: parity must be \"even\" or \"odd\"")
		}
		if t.RecurrenceDailyInterval != nil || len(t.RecurrenceMonthlyDays) > 0 || len(t.RecurrenceSpecificDates) > 0 {
			return fmt.Errorf("day_parity: unexpected params for other recurrence types")
		}

	default:
		return fmt.Errorf("unknown recurrence type %q", *t.RecurrenceType)
	}

	return nil
}

// GenerateOccurrences возвращает отсортированный список дат из [from, to] (включительно),
// усечённых до полуночи UTC. Если t не является шаблоном — возвращает пустой срез.
func GenerateOccurrences(t *Task, from, to time.Time) []time.Time {
	if t.RecurrenceType == nil {
		return []time.Time{}
	}
	switch *t.RecurrenceType {
	case RecurrenceDaily:
		if t.RecurrenceDailyInterval == nil {
			return []time.Time{}
		}
		return dailyOccurrences(*t.RecurrenceDailyInterval, from, to)
	case RecurrenceMonthlyDays:
		return monthlyDaysOccurrences(t.RecurrenceMonthlyDays, from, to)
	case RecurrenceSpecificDates:
		return specificDatesOccurrences(t.RecurrenceSpecificDates, from, to)
	case RecurrenceDayParity:
		if t.RecurrenceDayParity == nil {
			return []time.Time{}
		}
		return dayParityOccurrences(*t.RecurrenceDayParity, from, to)
	default:
		return []time.Time{}
	}
}

// dailyOccurrences срабатывает каждые interval дней, считая от эпохи 2000-01-01 UTC.
func dailyOccurrences(interval int, from, to time.Time) []time.Time {
	from = TruncateToDay(from)
	to = TruncateToDay(to)
	if from.After(to) {
		return []time.Time{}
	}

	epoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	daysSinceEpoch := int(from.Sub(epoch).Hours() / 24)
	rem := daysSinceEpoch % interval
	firstOffset := 0
	if rem != 0 {
		firstOffset = interval - rem
	}
	current := from.AddDate(0, 0, firstOffset)

	result := []time.Time{}
	for !current.After(to) {
		result = append(result, current)
		current = current.AddDate(0, 0, interval)
	}
	return result
}

// monthlyDaysOccurrences срабатывает на каждое из перечисленных чисел месяца (1–30).
// Числа, которых нет в данном месяце, пропускаются.
func monthlyDaysOccurrences(days []int, from, to time.Time) []time.Time {
	from = TruncateToDay(from)
	to = TruncateToDay(to)
	if from.After(to) {
		return []time.Time{}
	}

	sorted := make([]int, len(days))
	copy(sorted, days)
	sort.Ints(sorted)

	result := []time.Time{}
	year, month := from.Year(), from.Month()
	lastYear, lastMonth := to.Year(), to.Month()

	for {
		for _, day := range sorted {
			candidate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			// Go нормализует переполнение (Feb 30 → Mar 1/2); детектируем по смене месяца.
			if candidate.Month() != month {
				continue
			}
			if candidate.Before(from) || candidate.After(to) {
				continue
			}
			result = append(result, candidate)
		}
		if year == lastYear && month == lastMonth {
			break
		}
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	return result
}

// specificDatesOccurrences срабатывает только на перечисленные даты.
func specificDatesOccurrences(dates []time.Time, from, to time.Time) []time.Time {
	from = TruncateToDay(from)
	to = TruncateToDay(to)
	if from.After(to) {
		return []time.Time{}
	}

	result := []time.Time{}
	for _, d := range dates {
		d = TruncateToDay(d)
		if !d.Before(from) && !d.After(to) {
			result = append(result, d)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Before(result[j]) })
	return result
}

// dayParityOccurrences срабатывает на числа месяца с заданной чётностью.
// День 31 всегда пропускается для единообразия между месяцами.
func dayParityOccurrences(parity Parity, from, to time.Time) []time.Time {
	from = TruncateToDay(from)
	to = TruncateToDay(to)
	if from.After(to) {
		return []time.Time{}
	}

	result := []time.Time{}
	current := from
	for !current.After(to) {
		day := current.Day()
		if day != 31 {
			isEven := day%2 == 0
			if (parity == ParityEven && isEven) || (parity == ParityOdd && !isEven) {
				result = append(result, current)
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return result
}

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
