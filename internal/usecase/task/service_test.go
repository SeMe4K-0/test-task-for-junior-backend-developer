package task

import (
	"context"
	"sort"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type repoStub struct {
	nextTemplateID   int64
	nextOccurrenceID int64
	templates        map[int64]taskdomain.Template
	occurrences      map[int64]taskdomain.Task
}

func newRepoStub() *repoStub {
	return &repoStub{
		nextTemplateID:   1,
		nextOccurrenceID: 1,
		templates:        make(map[int64]taskdomain.Template),
		occurrences:      make(map[int64]taskdomain.Task),
	}
}

func (r *repoStub) Ping(context.Context) error { return nil }

func (r *repoStub) CreateTemplate(_ context.Context, template *taskdomain.Template) (*taskdomain.Template, error) {
	created := *template
	created.ID = r.nextTemplateID
	r.nextTemplateID++
	r.templates[created.ID] = created
	return &created, nil
}

func (r *repoStub) GetTemplateByID(_ context.Context, id int64) (*taskdomain.Template, error) {
	template, ok := r.templates[id]
	if !ok {
		return nil, taskdomain.ErrNotFound
	}
	return &template, nil
}

func (r *repoStub) UpdateTemplate(_ context.Context, template *taskdomain.Template) (*taskdomain.Template, error) {
	current, ok := r.templates[template.ID]
	if !ok {
		return nil, taskdomain.ErrNotFound
	}

	current.Title = template.Title
	current.Description = template.Description
	current.Schedule = template.Schedule
	current.UpdatedAt = template.UpdatedAt
	r.templates[current.ID] = current

	return &current, nil
}

func (r *repoStub) DeleteTemplate(_ context.Context, id int64) error {
	if _, ok := r.templates[id]; !ok {
		return taskdomain.ErrNotFound
	}
	delete(r.templates, id)
	for occurrenceID, occurrence := range r.occurrences {
		if occurrence.TemplateID == id {
			delete(r.occurrences, occurrenceID)
		}
	}
	return nil
}

func (r *repoStub) ListTemplates(_ context.Context, filter TemplateListInput) ([]taskdomain.Template, int, error) {
	items := make([]taskdomain.Template, 0, len(r.templates))
	for _, template := range r.templates {
		if filter.ScheduleType != nil {
			if template.Schedule == nil || template.Schedule.Type != *filter.ScheduleType {
				continue
			}
		}
		items = append(items, template)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID > items[j].ID
	})

	total := len(items)
	items = paginateTemplates(items, filter.Offset, filter.Limit)

	return items, total, nil
}

func (r *repoStub) ListTemplatesForSync(_ context.Context, templateID *int64) ([]taskdomain.Template, error) {
	items := make([]taskdomain.Template, 0, len(r.templates))
	for _, template := range r.templates {
		if templateID != nil && template.ID != *templateID {
			continue
		}
		items = append(items, template)
	}
	return items, nil
}

func (r *repoStub) UpsertOccurrences(_ context.Context, tasks []taskdomain.Task) error {
	for _, task := range tasks {
		if existingID, ok := r.findOccurrenceID(task.TemplateID, task.ScheduledFor); ok {
			_ = existingID
			continue
		}

		created := task
		created.ID = r.nextOccurrenceID
		r.nextOccurrenceID++
		r.occurrences[created.ID] = created
	}
	return nil
}

func (r *repoStub) GetOccurrenceByID(_ context.Context, id int64) (*taskdomain.Task, error) {
	task, ok := r.occurrences[id]
	if !ok {
		return nil, taskdomain.ErrNotFound
	}

	if template, ok := r.templates[task.TemplateID]; ok {
		task.Schedule = template.Schedule
	}

	return &task, nil
}

func (r *repoStub) GetOccurrenceByTemplateAndDate(_ context.Context, templateID int64, scheduledFor taskdomain.Date) (*taskdomain.Task, error) {
	for _, occurrence := range r.occurrences {
		if occurrence.TemplateID == templateID && occurrence.ScheduledFor.Equal(scheduledFor.Time) {
			task := occurrence
			if template, ok := r.templates[task.TemplateID]; ok {
				task.Schedule = template.Schedule
			}
			return &task, nil
		}
	}
	return nil, taskdomain.ErrNotFound
}

func (r *repoStub) UpdateOccurrence(_ context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	current, ok := r.occurrences[task.ID]
	if !ok {
		return nil, taskdomain.ErrNotFound
	}

	current.Title = task.Title
	current.Description = task.Description
	current.Status = task.Status
	current.UpdatedAt = task.UpdatedAt
	current.CompletedAt = task.CompletedAt
	r.occurrences[current.ID] = current

	if template, ok := r.templates[current.TemplateID]; ok {
		current.Schedule = template.Schedule
	}

	return &current, nil
}

func (r *repoStub) DeleteOccurrence(_ context.Context, id int64) error {
	if _, ok := r.occurrences[id]; !ok {
		return taskdomain.ErrNotFound
	}
	delete(r.occurrences, id)
	return nil
}

func (r *repoStub) DeleteOccurrencesFrom(_ context.Context, templateID int64, from taskdomain.Date) error {
	for id, occurrence := range r.occurrences {
		if occurrence.TemplateID == templateID && !occurrence.ScheduledFor.Before(from.Time) {
			delete(r.occurrences, id)
		}
	}
	return nil
}

func (r *repoStub) CountOccurrencesByTemplate(_ context.Context, templateID int64) (int, error) {
	total := 0
	for _, occurrence := range r.occurrences {
		if occurrence.TemplateID == templateID {
			total++
		}
	}
	return total, nil
}

func (r *repoStub) ListOccurrences(_ context.Context, filter TaskListInput) ([]taskdomain.Task, int, error) {
	items := make([]taskdomain.Task, 0, len(r.occurrences))
	for _, occurrence := range r.occurrences {
		if filter.TemplateID != nil && occurrence.TemplateID != *filter.TemplateID {
			continue
		}
		if filter.Status != nil && occurrence.Status != *filter.Status {
			continue
		}
		if filter.DateFrom != nil && occurrence.ScheduledFor.Before(filter.DateFrom.Time) {
			continue
		}
		if filter.DateTo != nil && occurrence.ScheduledFor.After(filter.DateTo.Time) {
			continue
		}
		if template, ok := r.templates[occurrence.TemplateID]; ok {
			occurrence.Schedule = template.Schedule
		}
		items = append(items, occurrence)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ScheduledFor.Equal(items[j].ScheduledFor.Time) {
			return items[i].ID > items[j].ID
		}
		return items[i].ScheduledFor.After(items[j].ScheduledFor.Time)
	})

	total := len(items)
	items = paginateTasks(items, filter.Offset, filter.Limit)

	return items, total, nil
}

func (r *repoStub) findOccurrenceID(templateID int64, scheduledFor taskdomain.Date) (int64, bool) {
	for id, occurrence := range r.occurrences {
		if occurrence.TemplateID == templateID && occurrence.ScheduledFor.Equal(scheduledFor.Time) {
			return id, true
		}
	}
	return 0, false
}

func paginateTasks(items []taskdomain.Task, offset, limit int) []taskdomain.Task {
	if offset >= len(items) {
		return []taskdomain.Task{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateTemplates(items []taskdomain.Template, offset, limit int) []taskdomain.Template {
	if offset >= len(items) {
		return []taskdomain.Template{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func TestCreateTaskCreatesTemplateAndOccurrence(t *testing.T) {
	repo := newRepoStub()
	service := NewService(repo)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC)
	}

	task, err := service.Create(context.Background(), CreateTaskInput{
		Title:       "One-off task",
		Description: "Call patient",
		Status:      taskdomain.StatusInProgress,
	})
	if err != nil {
		t.Fatal(err)
	}

	if task.TemplateID == 0 {
		t.Fatal("expected created task to reference a template")
	}
	if task.Status != taskdomain.StatusInProgress {
		t.Fatalf("expected status in_progress, got %s", task.Status)
	}
	if task.ScheduledFor.String() != "2026-04-08" {
		t.Fatalf("expected occurrence for today, got %s", task.ScheduledFor.String())
	}
}

func TestListSyncsRecurringOccurrencesByDateRange(t *testing.T) {
	repo := newRepoStub()
	repo.templates[1] = taskdomain.Template{
		ID:          1,
		Title:       "Daily calls",
		Description: "Follow-up",
		Schedule: &taskdomain.Schedule{
			Type:      taskdomain.ScheduleDaily,
			EveryDays: 2,
		},
		CreatedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC),
	}

	service := NewService(repo)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC)
	}

	dateFrom := taskdomain.NewDate(time.Date(2026, time.April, 9, 0, 0, 0, 0, time.UTC))
	dateTo := taskdomain.NewDate(time.Date(2026, time.April, 9, 0, 0, 0, 0, time.UTC))

	result, err := service.List(context.Background(), TaskListInput{
		DateFrom: &dateFrom,
		DateTo:   &dateTo,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Total != 1 {
		t.Fatalf("expected one occurrence, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one paginated item, got %d", len(result.Items))
	}
	if result.Items[0].ScheduledFor.String() != "2026-04-09" {
		t.Fatalf("expected synced occurrence on 2026-04-09, got %s", result.Items[0].ScheduledFor.String())
	}
}

func TestGetTemplateByIDComputesNextOccurrenceInTimezone(t *testing.T) {
	repo := newRepoStub()
	repo.templates[1] = taskdomain.Template{
		ID:          1,
		Title:       "Night shift follow-up",
		Description: "Daily check",
		Schedule: &taskdomain.Schedule{
			Type:      taskdomain.ScheduleDaily,
			EveryDays: 1,
			TimeZone:  "Europe/Moscow",
		},
		CreatedAt: time.Date(2026, time.April, 1, 22, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.April, 1, 22, 30, 0, 0, time.UTC),
	}

	service := NewService(repo)
	service.now = func() time.Time {
		return time.Date(2026, time.April, 1, 23, 0, 0, 0, time.UTC)
	}

	template, err := service.GetTemplateByID(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}

	if template.NextOccurrenceOn == nil || template.NextOccurrenceOn.String() != "2026-04-02" {
		t.Fatalf("expected next occurrence on 2026-04-02, got %#v", template.NextOccurrenceOn)
	}
}
