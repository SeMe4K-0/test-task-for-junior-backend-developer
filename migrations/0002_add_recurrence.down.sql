DROP INDEX IF EXISTS idx_tasks_parent_task_id;
DROP INDEX IF EXISTS idx_tasks_template_scheduled;

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS chk_recurrence_day_parity,
    DROP CONSTRAINT IF EXISTS chk_recurrence_specific_dates,
    DROP CONSTRAINT IF EXISTS chk_recurrence_monthly_days,
    DROP CONSTRAINT IF EXISTS chk_recurrence_daily,
    DROP CONSTRAINT IF EXISTS chk_recurrence_null_params,
    DROP CONSTRAINT IF EXISTS chk_recurrence_type_values,
    DROP COLUMN IF EXISTS recurrence_day_parity,
    DROP COLUMN IF EXISTS recurrence_specific_dates,
    DROP COLUMN IF EXISTS recurrence_monthly_days,
    DROP COLUMN IF EXISTS recurrence_daily_interval,
    DROP COLUMN IF EXISTS recurrence_type,
    DROP COLUMN IF EXISTS parent_task_id,
    DROP COLUMN IF EXISTS scheduled_at;
