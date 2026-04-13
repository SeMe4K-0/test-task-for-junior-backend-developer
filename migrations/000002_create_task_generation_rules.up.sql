CREATE TABLE IF NOT EXISTS task_generation_rules (
    id SERIAL PRIMARY KEY,
    -- Task info
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    -- Generation settings
    start_date DATE NOT NULL DEFAULT NOW(),
    end_date DATE,
    interval_days INTEGER,
    month_days INTEGER [],
    specific_dates DATE [],
    even_odd VARCHAR(20),
    -- Meta
    type VARCHAR(40),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_generated_at TIMESTAMPTZ
);

ALTER TABLE tasks 
    ADD COLUMN rule_id INTEGER REFERENCES task_generation_rules(id),
    ADD COLUMN due_date TIMESTAMPTZ;