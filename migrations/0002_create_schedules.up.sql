CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    recurrence_type VARCHAR NOT NULL,
    recurrence_params JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_created_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
    ADD COLUMN schedule_id BIGINT REFERENCES schedules(id) ON DELETE SET NULL