package handlers

import (
	"encoding/json"
	"errors"
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

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := taskusecase.CreateInput{
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		ExecutionDate: req.ExecutionDate,
	}

	if req.Recurrence != nil {
		input.Recurrence = &taskusecase.RecurrenceInput{
			Type:      req.Recurrence.Type,
			Value:     req.Recurrence.Value,
			StartDate: req.Recurrence.StartDate,
			EndDate:   req.Recurrence.EndDate,
		}
	}

	created, err := h.usecase.Create(r.Context(), input)
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

	var req taskUpdateDTO 
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateInput{
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		ExecutionDate: req.ExecutionDate, 
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
	now := time.Now().UTC()
	from := now.Truncate(24 * time.Hour) 
	to := from.AddDate(0, 1, 0)          

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = taskdomain.MidnightUTC(parsed)
		} else {
			writeError(w, http.StatusBadRequest, errors.New("invalid 'from' date format, expected YYYY-MM-DD"))
			return
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = taskdomain.MidnightUTC(parsed)
		} else {
			writeError(w, http.StatusBadRequest, errors.New("invalid 'to' date format, expected YYYY-MM-DD"))
			return
		}
	}

	if from.After(to) {
		writeError(w, http.StatusBadRequest, errors.New("'from' date cannot be after 'to' date"))
		return
	}

	tasks, err := h.usecase.List(r.Context(), from, to)
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

func (h *TaskHandler) Materialize(w http.ResponseWriter, r *http.Request) {
	var req materializeDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RuleID <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("rule_id is required for materialization"))
		return
	}

	input := taskusecase.MaterializeInput{
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		ExecutionDate: req.ExecutionDate,
		RuleID:        req.RuleID,
	}

	created, err := h.usecase.Materialize(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO(created))
}

func (h *TaskHandler) ListRules(w http.ResponseWriter, r *http.Request) {
    rules, err := h.usecase.ListRules(r.Context())
    if err != nil {
        writeUsecaseError(w, err)
        return
    }

    response := make([]taskRuleDTO, 0, len(rules))
    for _, rule := range rules {
        response = append(response, newTaskRuleDTO(rule))
    }

    writeJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.usecase.DeleteRule(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, taskdomain.ErrAlreadyExists): 
        writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
