package task_test

import (
	"testing"
	"time"

	"example.com/taskservice/internal/domain/task"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestRecurrence_Validate(t *testing.T) {
	tests := []struct {
		name    string
		r       task.Recurrence
		wantErr bool
	}{
		{
			name:    "daily valid",
			r:       task.Recurrence{Type: task.RecurrenceTypeDaily, Interval: 1},
			wantErr: false,
		},
		{
			name:    "daily zero interval",
			r:       task.Recurrence{Type: task.RecurrenceTypeDaily, Interval: 0},
			wantErr: true,
		},
		{
			name:    "monthly valid",
			r:       task.Recurrence{Type: task.RecurrenceTypeMonthly, MonthDays: []int{1, 15}},
			wantErr: false,
		},
		{
			name:    "monthly out of range day",
			r:       task.Recurrence{Type: task.RecurrenceTypeMonthly, MonthDays: []int{31}},
			wantErr: true,
		},
		{
			name:    "monthly duplicate day",
			r:       task.Recurrence{Type: task.RecurrenceTypeMonthly, MonthDays: []int{1, 1}},
			wantErr: true,
		},
		{
			name:    "monthly empty days",
			r:       task.Recurrence{Type: task.RecurrenceTypeMonthly},
			wantErr: true,
		},
		{
			name:    "dates valid",
			r:       task.Recurrence{Type: task.RecurrenceTypeDates, Dates: []time.Time{date(2026, 4, 1)}},
			wantErr: false,
		},
		{
			name:    "dates empty",
			r:       task.Recurrence{Type: task.RecurrenceTypeDates},
			wantErr: true,
		},
		{
			name:    "even_odd even",
			r:       task.Recurrence{Type: task.RecurrenceTypeEvenOdd, Parity: task.ParityEven},
			wantErr: false,
		},
		{
			name:    "even_odd odd",
			r:       task.Recurrence{Type: task.RecurrenceTypeEvenOdd, Parity: task.ParityOdd},
			wantErr: false,
		},
		{
			name:    "even_odd invalid parity",
			r:       task.Recurrence{Type: task.RecurrenceTypeEvenOdd, Parity: "both"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			r:       task.Recurrence{Type: "weekly"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.r.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestRecurrence_GenerateDates_Daily(t *testing.T) {
	r := task.Recurrence{Type: task.RecurrenceTypeDaily, Interval: 2}
	start := date(2026, 4, 1)
	end := date(2026, 4, 7)

	got := r.GenerateDates(start, end)

	want := []time.Time{
		date(2026, 4, 1),
		date(2026, 4, 3),
		date(2026, 4, 5),
		date(2026, 4, 7),
	}

	assertDates(t, got, want)
}

func TestRecurrence_GenerateDates_Monthly(t *testing.T) {
	// Day 29 and 30 in February should be skipped.
	r := task.Recurrence{Type: task.RecurrenceTypeMonthly, MonthDays: []int{1, 29, 30}}
	start := date(2026, 2, 1)
	end := date(2026, 3, 1)

	got := r.GenerateDates(start, end)

	want := []time.Time{
		date(2026, 2, 1),
		date(2026, 3, 1),
	}

	assertDates(t, got, want)
}

func TestRecurrence_GenerateDates_EvenOdd(t *testing.T) {
	r := task.Recurrence{Type: task.RecurrenceTypeEvenOdd, Parity: task.ParityEven}
	start := date(2026, 4, 1)
	end := date(2026, 4, 5)

	got := r.GenerateDates(start, end)

	want := []time.Time{
		date(2026, 4, 2),
		date(2026, 4, 4),
	}

	assertDates(t, got, want)
}

func TestRecurrence_GenerateDates_Dates(t *testing.T) {
	explicit := []time.Time{date(2026, 4, 10), date(2026, 4, 20)}
	r := task.Recurrence{Type: task.RecurrenceTypeDates, Dates: explicit}

	// start/end are ignored for this type
	got := r.GenerateDates(time.Time{}, time.Time{})

	assertDates(t, got, explicit)
}

func TestRecurrence_GenerateDates_EmptyRange(t *testing.T) {
	// start > end should produce no dates
	r := task.Recurrence{Type: task.RecurrenceTypeDaily, Interval: 1}
	got := r.GenerateDates(date(2026, 4, 10), date(2026, 4, 9))

	if len(got) != 0 {
		t.Errorf("expected empty result for inverted range, got %d dates", len(got))
	}
}

func assertDates(t *testing.T, got, want []time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, len(want) = %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("date[%d]: got %s, want %s", i, got[i].Format("2006-01-02"), want[i].Format("2006-01-02"))
		}
	}
}
