package task_test

import (
	"testing"
	"time"

	"example.com/taskservice/internal/domain/task"
)

// day возвращает полночь UTC для заданной даты год-месяц-день.
func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

// Helpers — собирают минимальный Task для тестирования конкретного типа периодичности.

func taskDaily(interval int) *task.Task {
	rt := task.RecurrenceDaily
	return &task.Task{RecurrenceType: &rt, RecurrenceDailyInterval: &interval}
}

func taskMonthlyDays(days []int) *task.Task {
	rt := task.RecurrenceMonthlyDays
	return &task.Task{RecurrenceType: &rt, RecurrenceMonthlyDays: days}
}

func taskSpecificDates(dates []time.Time) *task.Task {
	rt := task.RecurrenceSpecificDates
	return &task.Task{RecurrenceType: &rt, RecurrenceSpecificDates: dates}
}

func taskWeekdayParity(parity task.Parity) *task.Task {
	rt := task.RecurrenceDayParity
	return &task.Task{RecurrenceType: &rt, RecurrenceDayParity: &parity}
}

// ---------------------------------------------------------------------------
// ValidateRecurrence — daily
// ---------------------------------------------------------------------------

func TestValidateRecurrence_Daily(t *testing.T) {
	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{"interval 1 is valid", taskDaily(1), false},
		{"interval 7 is valid", taskDaily(7), false},
		{"interval 0 is invalid", taskDaily(0), true},
		{"negative interval is invalid", taskDaily(-1), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := task.ValidateRecurrence(tc.task)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateRecurrence() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateRecurrence — monthly_days
// ---------------------------------------------------------------------------

func TestValidateRecurrence_MonthlyDays(t *testing.T) {
	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{"valid single day", taskMonthlyDays([]int{15}), false},
		{"valid multiple days", taskMonthlyDays([]int{1, 15, 30}), false},
		{"empty days", taskMonthlyDays([]int{}), true},
		{"day 0 out of range", taskMonthlyDays([]int{0}), true},
		{"day 31 out of range", taskMonthlyDays([]int{31}), true},
		{"duplicate days", taskMonthlyDays([]int{5, 5}), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := task.ValidateRecurrence(tc.task)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateRecurrence() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateRecurrence — specific_dates
// ---------------------------------------------------------------------------

func TestValidateRecurrence_SpecificDates(t *testing.T) {
	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{"non-empty list", taskSpecificDates([]time.Time{day(2026, 6, 1)}), false},
		{"empty list", taskSpecificDates([]time.Time{}), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := task.ValidateRecurrence(tc.task)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateRecurrence() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateRecurrence — weekday_parity
// ---------------------------------------------------------------------------

func TestValidateRecurrence_WeekdayParity(t *testing.T) {
	tests := []struct {
		name    string
		task    *task.Task
		wantErr bool
	}{
		{"even is valid", taskWeekdayParity(task.ParityEven), false},
		{"odd is valid", taskWeekdayParity(task.ParityOdd), false},
		{
			"empty parity invalid",
			func() *task.Task {
				rt := task.RecurrenceDayParity
				p := task.Parity("")
				return &task.Task{RecurrenceType: &rt, RecurrenceDayParity: &p}
			}(),
			true,
		},
		{
			"unknown parity invalid",
			func() *task.Task {
				rt := task.RecurrenceDayParity
				p := task.Parity("both")
				return &task.Task{RecurrenceType: &rt, RecurrenceDayParity: &p}
			}(),
			true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := task.ValidateRecurrence(tc.task)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateRecurrence() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateOccurrences — daily
// ---------------------------------------------------------------------------

func TestDailyOccurrences(t *testing.T) {
	tests := []struct {
		name     string
		interval int
		from     time.Time
		to       time.Time
		want     []time.Time
	}{
		{
			name:     "interval=1 consecutive days",
			interval: 1,
			from:     day(2026, 1, 1),
			to:       day(2026, 1, 3),
			want:     []time.Time{day(2026, 1, 1), day(2026, 1, 2), day(2026, 1, 3)},
		},
		{
			name:     "interval=7 weekly from epoch-aligned date",
			interval: 7,
			from:     day(2000, 1, 1), // якорная эпоха
			to:       day(2000, 1, 22),
			want:     []time.Time{day(2000, 1, 1), day(2000, 1, 8), day(2000, 1, 15), day(2000, 1, 22)},
		},
		{
			name:     "interval=7 window starts mid-sequence",
			interval: 7,
			from:     day(2000, 1, 5), // между Jan-1 и Jan-8
			to:       day(2000, 1, 20),
			want:     []time.Time{day(2000, 1, 8), day(2000, 1, 15)},
		},
		{
			name:     "empty range from > to",
			interval: 1,
			from:     day(2026, 1, 5),
			to:       day(2026, 1, 3),
			want:     []time.Time{},
		},
		{
			name:     "single day range hits",
			interval: 1,
			from:     day(2026, 3, 10),
			to:       day(2026, 3, 10),
			want:     []time.Time{day(2026, 3, 10)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := task.GenerateOccurrences(taskDaily(tc.interval), tc.from, tc.to)
			assertDates(t, got, tc.want)
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateOccurrences — monthly_days
// ---------------------------------------------------------------------------

func TestMonthlyDaysOccurrences(t *testing.T) {
	tests := []struct {
		name string
		days []int
		from time.Time
		to   time.Time
		want []time.Time
	}{
		{
			// 30-е число не существует в феврале — должно быть пропущено.
			name: "day 30 in February skipped",
			days: []int{30},
			from: day(2026, 2, 1),
			to:   day(2026, 3, 31),
			want: []time.Time{day(2026, 3, 30)},
		},
		{
			// Числа 1 и 15 с пересечением границы месяца.
			name: "days 1 and 15, range crosses month boundary",
			days: []int{1, 15},
			from: day(2026, 1, 10),
			to:   day(2026, 2, 20),
			want: []time.Time{
				day(2026, 1, 15),
				day(2026, 2, 1),
				day(2026, 2, 15),
			},
		},
		{
			// from == to, совпадает с перечисленным числом.
			name: "single day range exact hit",
			days: []int{15},
			from: day(2026, 4, 15),
			to:   day(2026, 4, 15),
			want: []time.Time{day(2026, 4, 15)},
		},
		{
			// from == to, но число не входит в список.
			name: "single day range miss",
			days: []int{15},
			from: day(2026, 4, 10),
			to:   day(2026, 4, 10),
			want: []time.Time{},
		},
		{
			// Пустой диапазон.
			name: "empty range from > to",
			days: []int{1},
			from: day(2026, 3, 5),
			to:   day(2026, 3, 1),
			want: []time.Time{},
		},
		{
			// 29-е — есть в високосном 2028, но не в 2026.
			name: "day 29 skipped in non-leap February",
			days: []int{29},
			from: day(2026, 2, 1),
			to:   day(2026, 3, 31),
			want: []time.Time{day(2026, 3, 29)},
		},
		{
			// Числа переданы в убывающем порядке — вывод должен быть отсортирован.
			name: "days 20 and 5 are sorted in output",
			days: []int{20, 5},
			from: day(2026, 1, 1),
			to:   day(2026, 1, 31),
			want: []time.Time{day(2026, 1, 5), day(2026, 1, 20)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := task.GenerateOccurrences(taskMonthlyDays(tc.days), tc.from, tc.to)
			assertDates(t, got, tc.want)
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateOccurrences — specific_dates
// ---------------------------------------------------------------------------

func TestSpecificDatesOccurrences(t *testing.T) {
	tests := []struct {
		name  string
		dates []time.Time
		from  time.Time
		to    time.Time
		want  []time.Time
	}{
		{
			name:  "all dates inside range",
			dates: []time.Time{day(2026, 6, 1), day(2026, 7, 15)},
			from:  day(2026, 5, 1),
			to:    day(2026, 8, 1),
			want:  []time.Time{day(2026, 6, 1), day(2026, 7, 15)},
		},
		{
			name:  "all dates outside range",
			dates: []time.Time{day(2025, 1, 1), day(2030, 1, 1)},
			from:  day(2026, 1, 1),
			to:    day(2026, 12, 31),
			want:  []time.Time{},
		},
		{
			name:  "boundary dates included (inclusive on both ends)", // граничные даты включаются
			dates: []time.Time{day(2026, 1, 1), day(2026, 12, 31)},
			from:  day(2026, 1, 1),
			to:    day(2026, 12, 31),
			want:  []time.Time{day(2026, 1, 1), day(2026, 12, 31)},
		},
		{
			name:  "empty range from > to",
			dates: []time.Time{day(2026, 6, 1)},
			from:  day(2026, 7, 1),
			to:    day(2026, 6, 1),
			want:  []time.Time{},
		},
		{
			// Входные даты не отсортированы — вывод должен быть отсортирован.
			name:  "unsorted dates are returned sorted",
			dates: []time.Time{day(2026, 9, 15), day(2026, 6, 1)},
			from:  day(2026, 1, 1),
			to:    day(2026, 12, 31),
			want:  []time.Time{day(2026, 6, 1), day(2026, 9, 15)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := task.GenerateOccurrences(taskSpecificDates(tc.dates), tc.from, tc.to)
			assertDates(t, got, tc.want)
		})
	}
}

// ---------------------------------------------------------------------------
// GenerateOccurrences — weekday_parity
// ---------------------------------------------------------------------------

func TestWeekdayParityOccurrences(t *testing.T) {
	tests := []struct {
		name   string
		parity task.Parity
		from   time.Time
		to     time.Time
		want   []time.Time
	}{
		{
			// Чётные числа месяца с 30 днями (апрель): 2, 4, ..., 30.
			name:   "even days in April (30 days)",
			parity: task.ParityEven,
			from:   day(2026, 4, 1),
			to:     day(2026, 4, 30),
			want: func() []time.Time {
				var d []time.Time
				for i := 2; i <= 30; i += 2 {
					d = append(d, day(2026, 4, i))
				}
				return d
			}(),
		},
		{
			// Нечётные числа месяца с 31 днём (март). День 31 всегда пропускается.
			// Нечётные: 1, 3, 5, ..., 29. День 31 пропущен.
			name:   "odd days in March (31 days) — day 31 skipped",
			parity: task.ParityOdd,
			from:   day(2026, 3, 1),
			to:     day(2026, 3, 31),
			want: func() []time.Time {
				var d []time.Time
				for i := 1; i <= 29; i += 2 {
					d = append(d, day(2026, 3, i))
				}
				// день 31 не включается
				return d
			}(),
		},
		{
			// Чётные числа месяца с 31 днём — день 31 пропускается (31 нечётный,
			// но здесь явно проверяем правило пропуска).
			name:   "even days in March (31 days) — day 31 absent",
			parity: task.ParityEven,
			from:   day(2026, 3, 1),
			to:     day(2026, 3, 31),
			want: func() []time.Time {
				var d []time.Time
				for i := 2; i <= 30; i += 2 {
					d = append(d, day(2026, 3, i))
				}
				return d
			}(),
		},
		{
			// Нечётные числа месяца с 30 днями. День 30 чётный, последнее нечётное — 29.
			name:   "odd days in April (30 days) — last odd is 29",
			parity: task.ParityOdd,
			from:   day(2026, 4, 1),
			to:     day(2026, 4, 30),
			want: func() []time.Time {
				var d []time.Time
				for i := 1; i <= 29; i += 2 {
					d = append(d, day(2026, 4, i))
				}
				return d
			}(),
		},
		{
			// Пустой диапазон.
			name:   "empty range from > to",
			parity: task.ParityEven,
			from:   day(2026, 4, 10),
			to:     day(2026, 4, 5),
			want:   []time.Time{},
		},
		{
			// Один чётный день — совпадение.
			name:   "single even day",
			parity: task.ParityEven,
			from:   day(2026, 4, 4),
			to:     day(2026, 4, 4),
			want:   []time.Time{day(2026, 4, 4)},
		},
		{
			// Один нечётный день при чётной чётности — нет совпадения.
			name:   "single odd day with even parity",
			parity: task.ParityEven,
			from:   day(2026, 4, 5),
			to:     day(2026, 4, 5),
			want:   []time.Time{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := task.GenerateOccurrences(taskWeekdayParity(tc.parity), tc.from, tc.to)
			assertDates(t, got, tc.want)
		})
	}
}

// ---------------------------------------------------------------------------
// assertDates сравнивает два среза дат поэлементно.
// ---------------------------------------------------------------------------

func assertDates(t *testing.T, got, want []time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d, want %d\n  got:  %v\n  want: %v",
			len(got), len(want), formatDates(got), formatDates(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("index %d: got %s, want %s", i, got[i].Format("2006-01-02"), want[i].Format("2006-01-02"))
		}
	}
}

func formatDates(dates []time.Time) []string {
	s := make([]string, len(dates))
	for i, d := range dates {
		s[i] = d.Format("2006-01-02")
	}
	return s
}
