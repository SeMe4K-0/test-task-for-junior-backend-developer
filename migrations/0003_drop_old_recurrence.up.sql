ALTER TABLE tasks
DROP COLUMN recurrence_type,
DROP COLUMN recurrence_interval_days,
DROP COLUMN recurrence_day_of_month,
DROP COLUMN recurrence_parity_even,
DROP COLUMN recurrence_specific_dates,
DROP COLUMN recurrence_end_date;