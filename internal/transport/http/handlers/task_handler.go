package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TaskHandler struct {
	usecase taskusecase.Usecase
}

func NewTaskHandler(usecase taskusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

// Recurrence

func (h *TaskHandler) ListRecurrence(w http.ResponseWriter, r *http.Request) {
	recurrences, err := h.usecase.ListRecurrence(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]recurrenceDTO, 0, len(recurrences))
	for i := range recurrences {
		response = append(response, newReccurenceDTO(&recurrences[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) CreateRecurrencedTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fromStr := query.Get("from")
	toStr := query.Get("to")

	layout := "2006-01-02"

	if fromStr == "" || toStr == "" {
		writeUsecaseError(w, errors.New("params 'from' and 'to' must be specified"))
		return
	}

	from, err := time.Parse(layout, fromStr)
	if err != nil {
		writeUsecaseError(w, errors.New("invalid 'from' format"))
		return
	}
	to, err := time.Parse(layout, toStr)
	if err != nil {
		writeUsecaseError(w, errors.New("invalid 'to' format"))
		return
	}
	maxRange := 30

	if to.Sub(from) > time.Duration(maxRange)*24*time.Hour {
		writeUsecaseError(w, fmt.Errorf("range exceeds maximum allowed %d days", maxRange))
	}

	err = h.usecase.CreateRecurrencedTasks(r.Context(), from, to)

	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, "ok")
}

// Tasks

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Recurrence != nil {
		recurrence := taskusecase.CreateRecurrenceInput{
			Title:       req.Title,
			Description: req.Description,
			Recurrence: taskusecase.Recurrence{
				StartDate:     req.Recurrence.StartDate,
				EndDate:       req.Recurrence.EndDate,
				IntervalDays:  req.Recurrence.IntervalDays,
				MonthDays:     req.Recurrence.MonthDays,
				SpecificDates: req.Recurrence.SpecificDates,
				EvenOdd:       req.Recurrence.EvenOdd,
			},
		}
		created, err := h.usecase.CreateRecurrence(r.Context(), recurrence)
		if err != nil {
			writeUsecaseError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, newReccurenceDTO(created))
		return
	}

	task := taskusecase.CreateInput{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     *req.DueDate,
		Status:      req.Status,
	}
	created, err := h.usecase.Create(r.Context(), task)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskDTO(&tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing task id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid task id")
	}

	if id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskusecase.ErrInvalidRecurrenceInput):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}
