package scheduler

import (
	"testing"
	"time"
)

func TestSchedulerFuncObserverStoresTimerAndKey(t *testing.T) {
	s := NewTaskScheduler().(*taskScheduler)
	taskItem := NewTaskItem()

	observerValue := s.SchedulerFunc(time.Hour, taskItem, nil)
	if observerValue == nil {
		t.Fatal("expected scheduler observer")
	}

	observer := observerValue.(*taskSchedulerObserver)
	if observer.timer == nil {
		t.Fatal("expected recurring scheduler observer to store timer")
	}
	if observer.GetKey() == "" {
		t.Fatal("expected recurring scheduler observer to expose a key")
	}
	if observer.GetKey() != taskItem.GetKey() {
		t.Fatalf("expected observer key %q to match task item key %q", observer.GetKey(), taskItem.GetKey())
	}

	if !observer.Stop() {
		t.Fatal("expected observer stop to succeed")
	}
}

func TestTaskSchedulerObserverNilSafeAccessors(t *testing.T) {
	observer := &taskSchedulerObserver{}

	if observer.GetKey() != "" {
		t.Fatalf("expected empty key for observer without timer or task item, got %q", observer.GetKey())
	}
	if observer.GetTaskItemValue() != nil {
		t.Fatalf("expected nil task item value for observer without task item, got %#v", observer.GetTaskItemValue())
	}
}
