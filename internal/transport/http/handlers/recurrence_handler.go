package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	recurrenceusecase "example.com/taskservice/internal/usecase/recurrence"
)

type RecurrenceHandler struct {
	usecase recurrenceusecase.Usecase
}

func NewRecurrenceHandler(usecase recurrenceusecase.Usecase) *RecurrenceHandler {
	return &RecurrenceHandler{usecase: usecase}
}

func (h *RecurrenceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRecurrenceRuleDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := recurrenceusecase.CreateInput{
		TaskTitle:       req.TaskTitle,
		TaskDescription: req.TaskDescription,
		Recurrence: recurrenceusecase.RecurrenceParams{
			Type:       req.Recurrence.Type,
			EveryNDays: req.Recurrence.EveryNDays,
			DayOfMonth: req.Recurrence.DayOfMonth,
			Dates:      parseDates(req.Recurrence.Dates),
			EvenOdd:    req.Recurrence.EvenOdd,
		},
		StartDate: parseDate(req.StartDate),
		EndDate:   parseDate(req.EndDate),
	}

	output, err := h.usecase.Create(r.Context(), input)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	tasks := make([]taskDTO, 0, len(output.Tasks))
	for i := range output.Tasks {
		tasks = append(tasks, newTaskDTO(&output.Tasks[i]))
	}

	writeJSON(w, http.StatusCreated, createRecurrenceRuleResponseDTO{
		Rule:  newRecurrenceRuleDTO(output.Rule),
		Tasks: tasks,
	})
}

func (h *RecurrenceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getRecurrenceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rule, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newRecurrenceRuleDTO(rule))
}

func (h *RecurrenceHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.usecase.List(r.Context())
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	response := make([]recurrenceRuleDTO, 0, len(rules))
	for i := range rules {
		response = append(response, newRecurrenceRuleDTO(&rules[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *RecurrenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getRecurrenceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req createRecurrenceRuleDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := recurrenceusecase.UpdateInput{
		TaskTitle:       req.TaskTitle,
		TaskDescription: req.TaskDescription,
		Recurrence: recurrenceusecase.RecurrenceParams{
			Type:       req.Recurrence.Type,
			EveryNDays: req.Recurrence.EveryNDays,
			DayOfMonth: req.Recurrence.DayOfMonth,
			Dates:      parseDates(req.Recurrence.Dates),
			EvenOdd:    req.Recurrence.EvenOdd,
		},
		StartDate: parseDate(req.StartDate),
		EndDate:   parseDate(req.EndDate),
	}

	output, err := h.usecase.Update(r.Context(), id, input)
	if err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	tasks := make([]taskDTO, 0, len(output.Tasks))
	for i := range output.Tasks {
		tasks = append(tasks, newTaskDTO(&output.Tasks[i]))
	}

	writeJSON(w, http.StatusOK, createRecurrenceRuleResponseDTO{
		Rule:  newRecurrenceRuleDTO(output.Rule),
		Tasks: tasks,
	})
}

func (h *RecurrenceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getRecurrenceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	deleteTasks := r.URL.Query().Get("delete_tasks") == "true"

	if err := h.usecase.Delete(r.Context(), id, deleteTasks); err != nil {
		writeRecurrenceUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getRecurrenceIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing recurrence rule id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid recurrence rule id")
	}

	if id <= 0 {
		return 0, errors.New("invalid recurrence rule id")
	}

	return id, nil
}

func writeRecurrenceUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, recurrencedomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, recurrenceusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func parseDates(ss []string) []time.Time {
	dates := make([]time.Time, 0, len(ss))
	for _, s := range ss {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			dates = append(dates, t)
		}
	}
	return dates
}
