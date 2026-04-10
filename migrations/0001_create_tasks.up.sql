CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

CREATE TABLE IF NOT EXISTS task_recurrence (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL UNIQUE,
    recur_type VARCHAR(20) NOT NULL,
    interval_days INT DEFAULT NULL CHECK (interval_days > 0),
    month_day INT DEFAULT NULL CHECK (month_day BETWEEN 1 AND 30),
    specific_dates DATE[] DEFAULT NULL,
    parity VARCHAR(5) DEFAULT NULL CHECK (parity IN ('even', 'odd')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_task_recurrence_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_recurrence_task_id ON task_recurrence(task_id);
CREATE INDEX IF NOT EXISTS idx_task_recurrence_recur_type ON task_recurrence(recur_type);