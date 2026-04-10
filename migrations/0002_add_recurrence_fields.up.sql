ALTER TABLE tasks ADD COLUMN due_date DATE;
ALTER TABLE tasks ADD COLUMN is_template BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tasks ADD COLUMN parent_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE tasks ADD COLUMN recurrence_config JSONB;

CREATE INDEX idx_tasks_due_date ON tasks (due_date);
CREATE INDEX idx_tasks_parent_id ON tasks (parent_id);
CREATE INDEX idx_tasks_is_template ON tasks (is_template);
