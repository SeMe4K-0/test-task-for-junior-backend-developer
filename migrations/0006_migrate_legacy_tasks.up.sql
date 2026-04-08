DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'tasks'
	) THEN
		IF NOT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'tasks' AND column_name = 'schedule'
		) THEN
			ALTER TABLE tasks ADD COLUMN schedule JSONB;
		END IF;

		INSERT INTO task_templates (id, title, description, schedule, created_at, updated_at)
		SELECT id, title, description, schedule, created_at, updated_at
		FROM tasks
		ON CONFLICT (id) DO NOTHING;

		INSERT INTO task_occurrences (
			template_id,
			title,
			description,
			status,
			scheduled_for,
			created_at,
			updated_at,
			completed_at
		)
		SELECT
			id,
			title,
			description,
			status,
			created_at::date,
			created_at,
			updated_at,
			CASE WHEN status = 'done' THEN updated_at ELSE NULL END
		FROM tasks
		ON CONFLICT (template_id, scheduled_for) DO NOTHING;

		PERFORM setval(
			pg_get_serial_sequence('task_templates', 'id'),
			GREATEST(COALESCE((SELECT MAX(id) FROM task_templates), 1), 1),
			true
		);

		DROP TABLE tasks;
	END IF;
END $$;
