ALTER TABLE tasks 
ADD COLUMN scheduled_for TIMESTAMP NULL;

ALTER TABLE task_generation_rules 
DROP COLUMN last_generated_at;

CREATE UNIQUE INDEX unique_rule_scheduled_for 
ON tasks (rule_id, scheduled_for) 
WHERE rule_id IS NOT NULL AND scheduled_for IS NOT NULL;