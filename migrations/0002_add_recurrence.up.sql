-- Добавляем поддержку повторений задач (плоская модель колонок).
-- parent_task_id — самореференс FK с CASCADE удалением экземпляров.

ALTER TABLE tasks
    ADD COLUMN scheduled_at              TIMESTAMPTZ,
    ADD COLUMN parent_task_id            BIGINT REFERENCES tasks(id) ON DELETE CASCADE,
    ADD COLUMN recurrence_type           TEXT,
    ADD COLUMN recurrence_daily_interval INT,
    ADD COLUMN recurrence_monthly_days   INT[],
    ADD COLUMN recurrence_specific_dates DATE[],
    ADD COLUMN recurrence_day_parity TEXT;

-- recurrence_type: одно из 4 допустимых значений или NULL.
ALTER TABLE tasks
    ADD CONSTRAINT chk_recurrence_type_values
        CHECK (
            recurrence_type IS NULL
            OR recurrence_type IN ('daily', 'monthly_days', 'specific_dates', 'day_parity')
        );

-- Если recurrence_type IS NULL — все параметры тоже NULL.
ALTER TABLE tasks
    ADD CONSTRAINT chk_recurrence_null_params
        CHECK (
            recurrence_type IS NOT NULL
            OR (
                recurrence_daily_interval IS NULL
                AND recurrence_monthly_days IS NULL
                AND recurrence_specific_dates IS NULL
                AND recurrence_day_parity IS NULL
            )
        );

-- Для каждого типа повторения — только своя колонка параметров.
ALTER TABLE tasks
    ADD CONSTRAINT chk_recurrence_daily
        CHECK (
            recurrence_type != 'daily'
            OR (
                recurrence_daily_interval IS NOT NULL
                AND recurrence_monthly_days IS NULL
                AND recurrence_specific_dates IS NULL
                AND recurrence_day_parity IS NULL
            )
        ),
    ADD CONSTRAINT chk_recurrence_monthly_days
        CHECK (
            recurrence_type != 'monthly_days'
            OR (
                recurrence_monthly_days IS NOT NULL
                AND recurrence_daily_interval IS NULL
                AND recurrence_specific_dates IS NULL
                AND recurrence_day_parity IS NULL
            )
        ),
    ADD CONSTRAINT chk_recurrence_specific_dates
        CHECK (
            recurrence_type != 'specific_dates'
            OR (
                recurrence_specific_dates IS NOT NULL
                AND recurrence_daily_interval IS NULL
                AND recurrence_monthly_days IS NULL
                AND recurrence_day_parity IS NULL
            )
        ),
    ADD CONSTRAINT chk_recurrence_day_parity
        CHECK (
            recurrence_type != 'day_parity'
            OR (
                recurrence_day_parity IS NOT NULL
                AND recurrence_daily_interval IS NULL
                AND recurrence_monthly_days IS NULL
                AND recurrence_specific_dates IS NULL
            )
        );

CREATE UNIQUE INDEX idx_tasks_template_scheduled
    ON tasks (parent_task_id, scheduled_at)
    WHERE parent_task_id IS NOT NULL AND scheduled_at IS NOT NULL;

CREATE INDEX idx_tasks_parent_task_id ON tasks (parent_task_id);
