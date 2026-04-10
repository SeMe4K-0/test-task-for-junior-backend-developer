DROP INDEX IF EXISTS idx_tasks_scheduled_date;
ALTER TABLE tasks DROP COLUMN scheduled_date;
ALTER TABLE tasks DROP COLUMN period_conf;