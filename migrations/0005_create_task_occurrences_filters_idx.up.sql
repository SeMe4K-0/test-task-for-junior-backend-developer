CREATE INDEX IF NOT EXISTS idx_task_occurrences_scheduled_status
ON task_occurrences (scheduled_for DESC, status, template_id);
