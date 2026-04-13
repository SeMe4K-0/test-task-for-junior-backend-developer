CREATE TABLE IF NOT EXISTS task_rules (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,                 
    description TEXT NOT NULL DEFAULT '',
    recurrence_type TEXT NOT NULL, 
    recurrence_value JSONB,        
    start_date DATE NOT NULL,      
    end_date DATE,                 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
    execution_date DATE NOT NULL,  
    rule_id BIGINT REFERENCES task_rules(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tasks_rule_date UNIQUE (rule_id, execution_date)

);


CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_execution_date ON tasks (execution_date);