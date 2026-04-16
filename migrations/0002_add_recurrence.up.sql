CREATE TABLE IF NOT EXISTS recurrence_rules (
    id               BIGSERIAL PRIMARY KEY,
    type             TEXT NOT NULL,
    every_n_days     INT,
    day_of_month     INT,
    specific_dates   DATE[],
    even_odd         TEXT,
    start_date       DATE NOT NULL,
    end_date         DATE NOT NULL,
    task_title       TEXT NOT NULL,
    task_description TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks ADD COLUMN IF NOT EXISTS recurrence_rule_id BIGINT REFERENCES recurrence_rules(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS scheduled_date DATE;

CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_rule_id ON tasks (recurrence_rule_id);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks (scheduled_date);
