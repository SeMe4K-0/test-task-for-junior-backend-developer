-- Тип периодичности
CREATE TYPE recurrence_type AS ENUM ('daily', 'monthly', 'specific_dates', 'even_odd');

-- Таблица настроек периодичности (привязана к задаче-шаблону 1:1)
CREATE TABLE IF NOT EXISTS recurrence_rules (
    id            BIGSERIAL PRIMARY KEY,
    task_id       BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    type          recurrence_type NOT NULL,

    -- Для daily: каждый n-й день (default=1 — каждый день)
    interval_days INT DEFAULT NULL,

    -- Для monthly: число месяца (1–30)
    day_of_month  INT DEFAULT NULL,

    -- Для specific_dates: массив конкретных дат
    specific_dates DATE[] DEFAULT NULL,

    -- Для even_odd: 'even' или 'odd'
    even_odd      TEXT DEFAULT NULL,

    -- Общие настройки
    start_date    DATE NOT NULL,
    end_date      DATE DEFAULT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_interval_days CHECK (interval_days IS NULL OR interval_days > 0),
    CONSTRAINT chk_day_of_month CHECK (day_of_month IS NULL OR (day_of_month >= 1 AND day_of_month <= 30)),
    CONSTRAINT chk_even_odd     CHECK (even_odd IS NULL OR even_odd IN ('even', 'odd'))
);

CREATE UNIQUE INDEX idx_recurrence_rules_task_id ON recurrence_rules(task_id);

-- Расширяем таблицу tasks: шаблон → экземпляр
ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT DEFAULT NULL REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE tasks ADD COLUMN scheduled_date DATE DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN is_template    BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);
CREATE INDEX idx_tasks_scheduled_date ON tasks(scheduled_date);
CREATE INDEX idx_tasks_is_template    ON tasks(is_template);
