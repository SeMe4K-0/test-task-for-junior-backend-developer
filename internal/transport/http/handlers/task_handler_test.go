package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/infrastructure/logger"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...zap.Field)  {}
func (m *MockLogger) Info(msg string, fields ...zap.Field)   {}
func (m *MockLogger) Warn(msg string, fields ...zap.Field)   {}
func (m *MockLogger) Error(msg string, fields ...zap.Field)  {}
func (m *MockLogger) Fatal(msg string, fields ...zap.Field)  {}
func (m *MockLogger) With(fields ...zap.Field) logger.Logger { return m }

type MockUsecase struct {
	CreateFunc  func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error)
	GetByIDFunc func(ctx context.Context, id int64) (*taskdomain.Task, error)
	UpdateFunc  func(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error)
	DeleteFunc  func(ctx context.Context, id int64) error
	ListFunc    func(ctx context.Context) ([]taskdomain.Task, error)
}

func (m *MockUsecase) Create(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
	return m.CreateFunc(ctx, input)
}

func (m *MockUsecase) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockUsecase) Update(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
	return m.UpdateFunc(ctx, id, input)
}

func (m *MockUsecase) Delete(ctx context.Context, id int64) error {
	return m.DeleteFunc(ctx, id)
}

func (m *MockUsecase) List(ctx context.Context) ([]taskdomain.Task, error) {
	return m.ListFunc(ctx)
}

func TestHandlerCreate_Success(t *testing.T) {
	usecase := &MockUsecase{
		CreateFunc: func(ctx context.Context, input taskusecase.CreateInput) (*taskdomain.Task, error) {
			return &taskdomain.Task{
				ID:          1,
				Title:       input.Title,
				Description: input.Description,
				Status:      input.Status,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)

	body := []byte(`{"title":"Test","description":"Test Task","status":"new"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create() status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestHandlerCreate_InvalidJSON(t *testing.T) {
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(&MockUsecase{}, mockLogger)

	body := []byte(`invalid`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlerGetByID_Success(t *testing.T) {
	usecase := &MockUsecase{
		GetByIDFunc: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
			return &taskdomain.Task{
				ID:        id,
				Title:     "Task 1",
				Status:    taskdomain.StatusNew,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id:[0-9]+}", handler.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlerGetByID_NotFound(t *testing.T) {
	usecase := &MockUsecase{
		GetByIDFunc: func(ctx context.Context, id int64) (*taskdomain.Task, error) {
			return nil, taskdomain.ErrNotFound
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id:[0-9]+}", handler.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetByID() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandlerUpdate_Success(t *testing.T) {
	usecase := &MockUsecase{
		UpdateFunc: func(ctx context.Context, id int64, input taskusecase.UpdateInput) (*taskdomain.Task, error) {
			return &taskdomain.Task{
				ID:          id,
				Title:       input.Title,
				Description: input.Description,
				Status:      input.Status,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id:[0-9]+}", handler.Update).Methods(http.MethodPut)

	body := []byte(`{"title":"Updated","description":"Updated Task","status":"done"}`)
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Update() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandlerDelete_Success(t *testing.T) {
	usecase := &MockUsecase{
		DeleteFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)
	router := mux.NewRouter()
	router.HandleFunc("/tasks/{id:[0-9]+}", handler.Delete).Methods(http.MethodDelete)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandlerList_Success(t *testing.T) {
	usecase := &MockUsecase{
		ListFunc: func(ctx context.Context) ([]taskdomain.Task, error) {
			return []taskdomain.Task{
				{ID: 1, Title: "Task 1", Status: taskdomain.StatusNew, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: 2, Title: "Task 2", Status: taskdomain.StatusDone, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, nil
		},
	}
	mockLogger := &MockLogger{}
	handler := NewTaskHandler(usecase, mockLogger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List() status = %d, want %d", w.Code, http.StatusOK)
	}

	var results []taskDTO
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Errorf("List() failed to decode response: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("List() returned %d tasks, want 2", len(results))
	}
}

func TestParseOptionalDate(t *testing.T) {
	tests := []struct {
		name      string
		input     *string
		wantNil   bool
		wantError bool
	}{
		{"nil input", nil, true, false},
		{"empty string", strPtr(""), true, false},
		{"valid date", strPtr("2026-04-08"), false, false},
		{"invalid date", strPtr("04/08/2026"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalDate(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("parseOptionalDate() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantNil && got != nil {
				t.Errorf("parseOptionalDate() got %v, want nil", got)
			}

			if !tt.wantNil && !tt.wantError && got == nil {
				t.Errorf("parseOptionalDate() got nil, want non-nil")
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
