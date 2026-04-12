UPDATE tasks
SET recurrence = jsonb_build_object(
    'type', recurrence_type,
    'interval_days', recurrence_interval_days,
    'day_of_month', recurrence_day_of_month,
    'parity_even', recurrence_parity_even,
    'specific_dates', recurrence_specific_dates,
    'end_date', recurrence_end_date
);