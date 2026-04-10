-- Adds scheduling metadata to tasks.
--
-- scheduled_date: the calendar day this specific task instance is assigned to.
--                 NULL for tasks created without a recurrence schedule.
-- recurrence:     the recurrence rule that generated this task, stored as JSONB
--                 so the original settings are preserved alongside each instance.
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS scheduled_date DATE,
    ADD COLUMN IF NOT EXISTS recurrence     JSONB;

-- Index speeds up queries like "show me all tasks for this week".
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks (scheduled_date);
