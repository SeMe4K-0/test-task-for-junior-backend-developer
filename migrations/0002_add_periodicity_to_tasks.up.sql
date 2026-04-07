-- Добавляем коллоны для поддержки периодичности задач
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS periodicity_type TEXT NOT NULL DEFAULT 'none';
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS periodicity_config JSONB NOT NULL DEFAULT '{}';

-- Создаём индекс для быстрого поиска по типу периодичности
CREATE INDEX IF NOT EXISTS idx_tasks_periodicity_type ON tasks (periodicity_type);
