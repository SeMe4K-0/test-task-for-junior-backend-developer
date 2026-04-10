CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	
	recurrence JSONB NOT NULL,
	start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_DATE,

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recurrence_type ON tasks ((recurrence->>'type'));
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
