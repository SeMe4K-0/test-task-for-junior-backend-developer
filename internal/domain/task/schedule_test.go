package task

import (
	"testing"
	"time"
)

func TestScheduleOccursOn(t *testing.T) {
	createdAt := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	t.Run("daily interval", func(t *testing.T) {
		schedule := &Schedule{Type: ScheduleDaily, EveryDays: 2}

		if !schedule.OccursOn(NewDate(time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC)), createdAt) {
			t.Fatal("expected occurrence on every second day")
		}

		if schedule.OccursOn(NewDate(time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)), createdAt) {
			t.Fatal("did not expect occurrence on off day")
		}
	})

	t.Run("monthly day", func(t *testing.T) {
		schedule := &Schedule{Type: ScheduleMonthly, DayOfMonth: 15}

		if !schedule.OccursOn(NewDate(time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC)), createdAt) {
			t.Fatal("expected monthly occurrence")
		}
	})

	t.Run("specific dates", func(t *testing.T) {
		dateA := NewDate(time.Date(2026, time.April, 10, 0, 0, 0, 0, time.UTC))
		schedule := &Schedule{Type: ScheduleSpecific, Dates: []Date{dateA}}

		if !schedule.OccursOn(dateA, createdAt) {
			t.Fatal("expected explicit date occurrence")
		}
	})

	t.Run("timezone shifts created day", func(t *testing.T) {
		schedule := &Schedule{
			Type:      ScheduleDaily,
			EveryDays: 1,
			TimeZone:  "Europe/Moscow",
		}
		lateUTC := time.Date(2026, time.April, 1, 22, 30, 0, 0, time.UTC)

		if schedule.OccursOn(NewDate(time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)), lateUTC) {
			t.Fatal("did not expect occurrence on previous local day")
		}

		if !schedule.OccursOn(NewDate(time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)), lateUTC) {
			t.Fatal("expected occurrence on created local day")
		}
	})
}

func TestScheduleNextOccurrence(t *testing.T) {
	createdAt := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	schedule := &Schedule{Type: ScheduleEvenDays}
	next := schedule.NextOccurrence(NewDate(time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC)), createdAt)
	if next == nil || next.String() != "2026-04-04" {
		t.Fatalf("expected 2026-04-04, got %#v", next)
	}
}

func TestScheduleValidateTimezone(t *testing.T) {
	schedule := &Schedule{
		Type:      ScheduleDaily,
		EveryDays: 1,
		TimeZone:  "Mars/Base",
	}

	if err := schedule.Validate(); err == nil {
		t.Fatal("expected invalid timezone error")
	}
}
