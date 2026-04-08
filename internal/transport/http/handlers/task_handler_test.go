package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type taskUsecaseStub struct {
	listResult taskusecase.TaskListResult
}

func (s *taskUsecaseStub) Create(context.Context, taskusecase.CreateTaskInput) (*taskdomain.Task, error) {
	return nil, nil
}

func (s *taskUsecaseStub) GetByID(context.Context, int64) (*taskdomain.Task, error) {
	return nil, nil
}

func (s *taskUsecaseStub) Update(context.Context, int64, taskusecase.UpdateTaskInput) (*taskdomain.Task, error) {
	return nil, nil
}

func (s *taskUsecaseStub) Delete(context.Context, int64) error {
	return nil
}

func (s *taskUsecaseStub) List(context.Context, taskusecase.TaskListInput) (taskusecase.TaskListResult, error) {
	return s.listResult, nil
}

func TestDecodeJSONRejectsMultipleObjects(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(`{"title":"a"}{"title":"b"}`))

	err := decodeJSON(recorder, request, &taskMutationDTO{})
	if err == nil {
		t.Fatal("expected decoder to reject multiple JSON objects")
	}
}

func TestTaskHandlerListWritesPaginationHeaders(t *testing.T) {
	handler := NewTaskHandler(&taskUsecaseStub{
		listResult: taskusecase.TaskListResult{
			Items: []taskdomain.Task{
				{
					ID:           1,
					TemplateID:   10,
					Title:        "Daily calls",
					Description:  "Follow-up",
					Status:       taskdomain.StatusNew,
					ScheduledFor: taskdomain.NewDate(time.Date(2026, time.April, 8, 0, 0, 0, 0, time.UTC)),
					CreatedAt:    time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
					UpdatedAt:    time.Date(2026, time.April, 8, 9, 0, 0, 0, time.UTC),
				},
			},
			Total:  17,
			Limit:  5,
			Offset: 10,
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?limit=5&offset=10", nil)

	handler.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Total-Count") != "17" {
		t.Fatalf("expected X-Total-Count header, got %q", recorder.Header().Get("X-Total-Count"))
	}

	var response []taskDTO
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response) != 1 || response[0].TemplateID != 10 {
		t.Fatalf("unexpected response payload: %#v", response)
	}
}
