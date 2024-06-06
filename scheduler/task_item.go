package scheduler

import (
	uuid "github.com/satori/go.uuid"
)

type TaskItem struct {
	key   string
	Value interface{}

	Properties map[string]interface{}
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

func (s *TaskItem) GetProperty(name string) interface{} {
	if len(s.Properties) <= 0 {
		return nil
	}
	v, ok := s.Properties[name]
	if !ok {
		return nil
	}
	return v
}

func (s *TaskItem) SetProperty(name string, v interface{}) {
	if s.Properties == nil {
		s.Properties = make(map[string]interface{})
	}
	if v == nil {
		return
	}
	s.Properties[name] = v
}
