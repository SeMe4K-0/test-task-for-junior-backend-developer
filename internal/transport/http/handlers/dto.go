package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

//здесь в дто добавил PeriodConf 

type taskMutationDTO struct {
	Title       string            		`json:"title"`
	Description string            		`json:"description"`
	Status      taskdomain.Status 		`json:"status"`
	PeriodConf	*taskdomain.PeriodConf 	`json:"period_conf"`
}

type taskDTO struct {
	ID          int64             					`json:"id"`
	Title       string            					`json:"title"`
	Description string            					`json:"description"`
	Status      taskdomain.Status 					`json:"status"`
	CreatedAt   time.Time         					`json:"created_at"`
	UpdatedAt   time.Time         					`json:"updated_at"`
	PeriodConf  *taskdomain.PeriodConf 				`json:"period_conf"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		PeriodConf:  task.PeriodConf, 

	}
}
