package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

const maxRequestBodyBytes = 1 << 20

type TaskHandler struct {
	usecase taskusecase.TaskUsecase
}

func NewTaskHandler(usecase taskusecase.TaskUsecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), taskusecase.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Schedule:    req.Schedule,
	})
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
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, taskusecase.UpdateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Schedule:    req.Schedule,
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
	input, err := buildTaskListInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.usecase.List(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writePageHeaders(w, result.Total, result.Limit, result.Offset)

	response := make([]taskDTO, 0, len(result.Items))
	for i := range result.Items {
		response = append(response, newTaskDTO(&result.Items[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func buildTaskListInput(r *http.Request) (taskusecase.TaskListInput, error) {
	limit, err := parseOptionalInt(r.URL.Query().Get("limit"))
	if err != nil {
		return taskusecase.TaskListInput{}, errors.New("invalid limit query parameter")
	}
	offset, err := parseOptionalInt(r.URL.Query().Get("offset"))
	if err != nil {
		return taskusecase.TaskListInput{}, errors.New("invalid offset query parameter")
	}

	var status *taskdomain.Status
	if rawStatus := r.URL.Query().Get("status"); rawStatus != "" {
		value := taskdomain.Status(rawStatus)
		status = &value
	}

	var templateID *int64
	if rawTemplateID := r.URL.Query().Get("template_id"); rawTemplateID != "" {
		value, err := strconv.ParseInt(rawTemplateID, 10, 64)
		if err != nil {
			return taskusecase.TaskListInput{}, errors.New("invalid template_id query parameter")
		}
		templateID = &value
	}

	dateFrom, err := parseOptionalDate(r.URL.Query().Get("date_from"))
	if err != nil {
		return taskusecase.TaskListInput{}, errors.New("invalid date_from query parameter, expected YYYY-MM-DD")
	}
	dateTo, err := parseOptionalDate(r.URL.Query().Get("date_to"))
	if err != nil {
		return taskusecase.TaskListInput{}, errors.New("invalid date_to query parameter, expected YYYY-MM-DD")
	}

	return taskusecase.TaskListInput{
		Limit:      limit,
		Offset:     offset,
		Status:     status,
		TemplateID: templateID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}, nil
}

func parseOptionalInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func parseOptionalDate(raw string) (*taskdomain.Date, error) {
	if raw == "" {
		return nil, nil
	}

	date, err := taskdomain.ParseDate(raw)
	if err != nil {
		return nil, err
	}

	return &date, nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
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

func writePageHeaders(w http.ResponseWriter, total, limit, offset int) {
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
}
