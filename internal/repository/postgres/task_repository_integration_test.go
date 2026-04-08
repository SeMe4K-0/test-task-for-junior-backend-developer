package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

func TestRepositoryTemplateAndOccurrenceLifecycle(t *testing.T) {
	baseDSN := os.Getenv("TEST_DATABASE_DSN")
	if baseDSN == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := context.Background()
	schemaName := fmt.Sprintf("task_repo_test_%d", time.Now().UnixNano())

	adminPool, err := infrastructurepostgres.Open(ctx, baseDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer adminPool.Close()

	if _, err := adminPool.Exec(ctx, `CREATE SCHEMA `+schemaName); err != nil {
		t.Fatal(err)
	}
	defer adminPool.Exec(ctx, `DROP SCHEMA `+schemaName+` CASCADE`)

	testDSN, err := withSearchPath(baseDSN, schemaName)
	if err != nil {
		t.Fatal(err)
	}

	pool, err := infrastructurepostgres.Open(ctx, testDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := infrastructurepostgres.ApplyMigrations(ctx, pool, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	repo := New(pool)
	template, err := repo.CreateTemplate(ctx, &taskdomain.Template{
		Title:       "Daily calls",
		Description: "Follow-up",
		Schedule: &taskdomain.Schedule{
			Type:      taskdomain.ScheduleDaily,
			EveryDays: 1,
		},
		CreatedAt: time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpsertOccurrences(ctx, []taskdomain.Task{
		{
			TemplateID:   template.ID,
			Title:        template.Title,
			Description:  template.Description,
			Status:       taskdomain.StatusNew,
			ScheduledFor: taskdomain.NewDate(time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)),
			CreatedAt:    time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	items, total, err := repo.ListOccurrences(ctx, taskusecase.TaskListInput{
		Limit:      10,
		Offset:     0,
		TemplateID: &template.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one occurrence, got total=%d items=%d", total, len(items))
	}
	if items[0].Schedule == nil || items[0].Schedule.Type != taskdomain.ScheduleDaily {
		t.Fatalf("expected schedule to be joined from template, got %#v", items[0].Schedule)
	}
}

func withSearchPath(rawDSN, schema string) (string, error) {
	parsed, err := url.Parse(rawDSN)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}
