package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// /api/v1/tasks	///////////////////////////////////////////////////////////////////////////////////////////
func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// /api/v1/tasks/{id}	//////////////////////////////////////////////////////////////////////////////////////////////
func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

//

// /api/v1/tasks/batch	//////////////////////////////////////////////////////////////////////////////////////////////
// входит первичная инфа
func (s *Service) CreatePereodic(ctx context.Context, input CreatePereodicInput) (*taskdomain.Task, error) {
	normalized, err := validateCreatePereodicInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
	}
	var strbldr strings.Builder
	strbldr.WriteString(model.Title)
	strbldr.WriteString(" (Повторное)")

	if input.Repetition != 0 {
		model.Title = strbldr.String()
	}

	//now := s.now()
	model.CreatedAt = input.StartDate
	model.UpdatedAt = input.StartDate
	//

	created, err := s.repo.CreatePereodic(ctx, model)
	if err != nil {
		return nil, err
	}
	model.CreatedAt, err = NextPereodicOcurrence(input)
	model.UpdatedAt, err = NextPereodicOcurrence(input)

	for i := 1; i < input.Repetition; i++ {
		created, err = s.repo.CreatePereodic(ctx, model)
		model.CreatedAt, err = NextPereodicOcurrence(input)
		model.UpdatedAt, err = NextPereodicOcurrence(input)
	}

	return created, nil
}

func validateCreatePereodicInput(input CreatePereodicInput) (CreatePereodicInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	//
	if input.Repetition < 1 {
		//return CreatePereodicInput{}, fmt.Errorf("%w: invalid repetition", ErrInvalidInput)
		input.Repetition = 1
	}

	if input.DayAmount < 1 {
		//return CreatePereodicInput{}, fmt.Errorf("%w: invalid DayAmount", ErrInvalidInput)
		input.DayAmount = 1
	}

	switch input.RecurrType {
	case taskdomain.OnceNDay, taskdomain.OnceOtherDay, taskdomain.OnceAMonth:
		//valid
	default:
		//return CreatePereodicInput{}, fmt.Errorf("%w: invalid RecurrType", ErrInvalidInput)
		input.RecurrType = taskdomain.OnceNDay
	}

	if input.StartDate.IsZero() {
		//return CreatePereodicInput{}, fmt.Errorf("%w: invalid StartDate", ErrInvalidInput)
		input.StartDate = time.Now()
	}

	//

	if input.Title == "" {
		return CreatePereodicInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreatePereodicInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func NextPereodicOcurrence(input CreatePereodicInput) (time.Time, error) {
	DiffTime := input.StartDate
	DayA := input.DayAmount

	switch input.RecurrType {
	case taskdomain.OnceNDay:
		return DiffTime.AddDate(0, 0, DayA), nil
	case taskdomain.OnceOtherDay:
		return DiffTime.AddDate(0, 0, 2), nil
	case taskdomain.OnceAMonth:
		return DiffTime.AddDate(0, 1, 0), nil
	default:
		return time.Now(), fmt.Errorf("%w: invalid input", ErrInvalidInput)
	}
}

// other //////////////////////////////////////////////////////////////////////////
func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
