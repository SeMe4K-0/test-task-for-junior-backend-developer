package task

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/infrastructure/logger"
)

type MockRepository struct {
	CreateFunc  func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByIDFunc func(ctx context.Context, id int64) (*taskdomain.Task, error)
	UpdateFunc  func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	DeleteFunc  func(ctx context.Context, id int64) error
	ListFunc    func(ctx context.Context) ([]taskdomain.Task, error)
}

func (m *MockRepository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	return m.CreateFunc(ctx, task)
}

func (m *MockRepository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockRepository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	return m.UpdateFunc(ctx, task)
}

func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	return m.DeleteFunc(ctx, id)
}

func (m *MockRepository) List(ctx context.Context) ([]taskdomain.Task, error) {
	return m.ListFunc(ctx)
}

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zap.Field)  {}
func (m *MockLogger) Info(msg string, fields ...zap.Field)   {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)   {}
func (m *MockLogger) Error(msg string, fields ...zap.Field)  {}
func (m *MockLogger) Fatal(msg string, fields ...zap.Field)  {}
func (m *MockLogger) With(fields ...zap.Field) logger.Logger { return m }

func TestServiceCreate(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
			task.ID = 1
			return task, nil
		},
	}

	// Mock logger for tests
	mockLogger := &MockLogger{}

	svc := NewService(repo, mockLogger)
	result, err := svc.Create(context.Background(), CreateInput{
		Title:       "Test",
		Description: "Test Task",
		Status:      taskdomain.StatusNew,
	})

	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil || result.Title != "Test" {
		t.Errorf("Create() returned unexpected result")
	}
}

func TestServiceCreate_InvalidInput(t *testing.T) {
	repo := &MockRepository{}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)

	_, err := svc.Create(context.Background(), CreateInput{
		Title: "   ",
	})

	if err == nil {
		t.Errorf("Create() should return error for empty title")
	}
}

func TestServiceGetByID(t *testing.T) {
	now := time.Now()
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
			return &taskdomain.Task{
				ID:        id,
				Title:     "Task 1",
				Status:    taskdomain.StatusNew,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	result, err := svc.GetByID(context.Background(), 1)

	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if result == nil || result.ID != 1 {
		t.Errorf("GetByID() returned unexpected result")
	}
}

func TestServiceGetByID_InvalidID(t *testing.T) {
	repo := &MockRepository{}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)

	_, err := svc.GetByID(context.Background(), 0)

	if err == nil {
		t.Errorf("GetByID() should return error for invalid ID")
	}
}

func TestServiceList(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context) ([]taskdomain.Task, error) {
			return []taskdomain.Task{
				{ID: 1, Title: "Task 1", Status: taskdomain.StatusNew},
				{ID: 2, Title: "Task 2", Status: taskdomain.StatusDone},
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	result, err := svc.List(context.Background())

	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("List() returned %d tasks, want 2", len(result))
	}
}

func TestServiceDelete(t *testing.T) {
	repo := &MockRepository{
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	err := svc.Delete(context.Background(), 1)

	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestServiceDelete_InvalidID(t *testing.T) {
	repo := &MockRepository{}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)

	err := svc.Delete(context.Background(), 0)

	if err == nil {
		t.Errorf("Delete() should return error for invalid ID")
	}
}

func TestServiceUpdate(t *testing.T) {
	repo := &MockRepository{
		UpdateFunc: func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
			task.ID = 1
			return task, nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	result, err := svc.Update(context.Background(), 1, UpdateInput{
		Title:       "Updated",
		Description: "Updated Task",
		Status:      taskdomain.StatusDone,
	})

	if err != nil {
		t.Errorf("Update() error = %v", err)
	}
	if result == nil || result.Title != "Updated" {
		t.Errorf("Update() returned unexpected result")
	}
}

func TestServiceCreate_WithDueDate(t *testing.T) {
	now := time.Now()
	dueDate := now.AddDate(0, 1, 0)
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
			task.ID = 1
			return task, nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	result, err := svc.Create(context.Background(), CreateInput{
		Title:   "Test",
		DueDate: &dueDate,
	})

	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil || result.DueDate == nil {
		t.Errorf("Create() did not preserve DueDate")
	}
}

func TestServiceCreate_WithRecurrence(t *testing.T) {
	repo := &MockRepository{
		CreateFunc: func(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
			task.ID = 1
			return task, nil
		},
	}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)
	result, err := svc.Create(context.Background(), CreateInput{
		Title: "Test",
		Recurrence: &taskdomain.Recurrence{
			Type:         taskdomain.RecurrenceTypeDaily,
			IntervalDays: 2,
		},
	})

	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if result == nil {
		t.Errorf("Create() returned nil")
	}
}

func TestServiceCreate_InvalidRecurrence(t *testing.T) {
	repo := &MockRepository{}
	mockLogger := &MockLogger{}
	svc := NewService(repo, mockLogger)

	_, err := svc.Create(context.Background(), CreateInput{
		Title: "Test",
		Recurrence: &taskdomain.Recurrence{
			Type:         taskdomain.RecurrenceTypeDaily,
			IntervalDays: 0,
		},
	})

	if err == nil {
		t.Errorf("Create() should return error for invalid recurrence")
	}
}
