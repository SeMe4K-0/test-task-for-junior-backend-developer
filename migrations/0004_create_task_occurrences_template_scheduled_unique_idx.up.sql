CREATE UNIQUE INDEX IF NOT EXISTS idx_task_occurrences_template_scheduled
ON task_occurrences (template_id, scheduled_for);
