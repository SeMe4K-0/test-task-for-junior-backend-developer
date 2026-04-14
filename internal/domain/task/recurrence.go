package task

import "time"

// RecurrenceType определяет тип периодичности задачи.
type RecurrenceType string

const (
	RecurrenceDaily         RecurrenceType = "daily"
	RecurrenceMonthly       RecurrenceType = "monthly"
	RecurrenceSpecificDates RecurrenceType = "specific_dates"
	RecurrenceEvenOdd       RecurrenceType = "even_odd"
)

// Valid проверяет корректность типа периодичности.
func (rt RecurrenceType) Valid() bool {
	switch rt {
	case RecurrenceDaily, RecurrenceMonthly, RecurrenceSpecificDates, RecurrenceEvenOdd:
		return true
	default:
		return false
	}
}

// EvenOdd — чётные или нечётные дни.
type EvenOdd string

const (
	Even EvenOdd = "even"
	Odd  EvenOdd = "odd"
)

// ValidEvenOdd проверяет корректность значения even/odd.
func (e EvenOdd) Valid() bool {
	return e == Even || e == Odd
}

// RecurrenceRule описывает правило периодичности, привязанное к задаче-шаблону.
type RecurrenceRule struct {
	ID            int64
	TaskID        int64
	Type          RecurrenceType
	IntervalDays  *int       // для daily: интервал в днях (≥1)
	DayOfMonth    *int       // для monthly: день месяца (1–30)
	SpecificDates []time.Time // для specific_dates: конкретные даты
	EvenOdd       *EvenOdd   // для even_odd: "even" или "odd"
	StartDate     time.Time  // начало действия правила
	EndDate       *time.Time // конец действия (nil = бессрочно)
	CreatedAt     time.Time
}

// ShouldCreateOn определяет, нужно ли создавать экземпляр задачи на указанную дату.
func (r *RecurrenceRule) ShouldCreateOn(date time.Time) bool {
	d := truncateToDate(date)

	// Проверяем диапазон действия правила.
	if d.Before(truncateToDate(r.StartDate)) {
		return false
	}

	if r.EndDate != nil && d.After(truncateToDate(*r.EndDate)) {
		return false
	}

	switch r.Type {
	case RecurrenceDaily:
		return r.shouldCreateDaily(d)
	case RecurrenceMonthly:
		return r.shouldCreateMonthly(d)
	case RecurrenceSpecificDates:
		return r.shouldCreateSpecificDate(d)
	case RecurrenceEvenOdd:
		return r.shouldCreateEvenOdd(d)
	default:
		return false
	}
}

func (r *RecurrenceRule) shouldCreateDaily(d time.Time) bool {
	interval := 1
	if r.IntervalDays != nil {
		interval = *r.IntervalDays
	}

	start := truncateToDate(r.StartDate)
	daysSinceStart := int(d.Sub(start).Hours() / 24)

	return daysSinceStart%interval == 0
}

func (r *RecurrenceRule) shouldCreateMonthly(d time.Time) bool {
	if r.DayOfMonth == nil {
		return false
	}

	dayOfMonth := *r.DayOfMonth

	// Если в месяце нет такого числа (например, 30 февраля) — не создаём.
	lastDay := lastDayOfMonth(d)
	if dayOfMonth > lastDay {
		return false
	}

	return d.Day() == dayOfMonth
}

func (r *RecurrenceRule) shouldCreateSpecificDate(d time.Time) bool {
	for _, sd := range r.SpecificDates {
		if truncateToDate(sd).Equal(d) {
			return true
		}
	}

	return false
}

func (r *RecurrenceRule) shouldCreateEvenOdd(d time.Time) bool {
	if r.EvenOdd == nil {
		return false
	}

	day := d.Day()
	isEven := day%2 == 0

	if *r.EvenOdd == Even {
		return isEven
	}

	return !isEven
}

// truncateToDate обнуляет время, оставляя только дату (UTC).
func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// lastDayOfMonth возвращает последний день месяца для заданной даты.
func lastDayOfMonth(t time.Time) int {
	y, m, _ := t.Date()
	// Первое число следующего месяца минус один день.
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
