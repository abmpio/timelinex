package scheduler

import (
	uuid "github.com/satori/go.uuid"
)

type TaskItem struct {
	key   string
	Value interface{}
}

func NewTaskItem() *TaskItem {
	return &TaskItem{}
}

func (i *TaskItem) SetKey(key string) {
	i.key = key
}

func (i *TaskItem) GetKey() string {
	return i.key
}

func (s *TaskItem) ensureHasKey() {
	if len(s.key) > 0 {
		return
	}

	s.key = uuid.NewV4().String()
}
