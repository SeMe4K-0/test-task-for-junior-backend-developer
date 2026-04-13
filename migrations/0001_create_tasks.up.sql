CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,

    repeated_type TEXT NOT NULL DEFAULT 'none',
    repeated_every_n_days INTEGER,
    repeated_day_of_month INTEGER,
    repeated_specific_dates TIMESTAMPTZ[],

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    CHECK (repeated_type IN ('none', 'every_n_days', 'monthly_day', 'specific_dates', 'even_days', 'odd_days'))
    CHECK (repeated_day_of_month IS NULL OR (repeated_day_of_month >= 1 AND repeated_day_of_month <= 30))
    CHECK (repeated_every_n_days IS NULL OR repeated_every_n_days > 0)
);


CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_repeated_type ON tasks (repeated_type);