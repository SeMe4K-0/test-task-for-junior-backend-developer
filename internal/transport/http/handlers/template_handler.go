package handlers

import (
	"errors"
	"net/http"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type TemplateHandler struct {
	usecase taskusecase.TemplateUsecase
}

func NewTemplateHandler(usecase taskusecase.TemplateUsecase) *TemplateHandler {
	return &TemplateHandler{usecase: usecase}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req templateMutationDTO
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.CreateTemplate(r.Context(), taskusecase.CreateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Schedule:    req.Schedule,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTemplateDTO(created))
}

func (h *TemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	template, err := h.usecase.GetTemplateByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTemplateDTO(template))
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req templateMutationDTO
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.UpdateTemplate(r.Context(), id, taskusecase.UpdateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Schedule:    req.Schedule,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTemplateDTO(updated))
}

func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeleteTemplate(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	input, err := buildTemplateListInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := h.usecase.ListTemplates(r.Context(), input)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writePageHeaders(w, result.Total, result.Limit, result.Offset)

	response := make([]templateDTO, 0, len(result.Items))
	for i := range result.Items {
		response = append(response, newTemplateDTO(&result.Items[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func buildTemplateListInput(r *http.Request) (taskusecase.TemplateListInput, error) {
	limit, err := parseOptionalInt(r.URL.Query().Get("limit"))
	if err != nil {
		return taskusecase.TemplateListInput{}, errors.New("invalid limit query parameter")
	}
	offset, err := parseOptionalInt(r.URL.Query().Get("offset"))
	if err != nil {
		return taskusecase.TemplateListInput{}, errors.New("invalid offset query parameter")
	}

	var scheduleType *taskdomain.ScheduleType
	if rawScheduleType := r.URL.Query().Get("schedule_type"); rawScheduleType != "" {
		value := taskdomain.ScheduleType(rawScheduleType)
		scheduleType = &value
	}

	return taskusecase.TemplateListInput{
		Limit:        limit,
		Offset:       offset,
		Search:       r.URL.Query().Get("search"),
		ScheduleType: scheduleType,
	}, nil
}
