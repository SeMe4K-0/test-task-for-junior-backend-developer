package postgres

import (
	"encoding/json"
	taskdomain "example.com/taskservice/internal/domain/task"
)

func encodeRecurrence(r *taskdomain.Recurrence) ([]byte, error) {
	if r == nil {
		return nil, nil
	}

	if r.Type == "" {
		return nil, nil
	}

	return json.Marshal(r)
}

func decodeRecurrence(data []byte) (taskdomain.Recurrence, error) {
	var r taskdomain.Recurrence

	if len(data) == 0 {
		return r, nil
	}

	err := json.Unmarshal(data, &r)
	return r, err
}
