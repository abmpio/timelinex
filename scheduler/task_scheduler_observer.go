package scheduler

import (
	"time"

	"github.com/abmpio/threadingx/timingwheel"
)

type timeIntervalScheduler struct {
	interval time.Duration

	stopped bool
}

func (s *timeIntervalScheduler) Next(prev time.Time) time.Time {
	if s.stopped {
		//已经停止
		return time.Time{}
	}
	return prev.Add(s.interval)
}

func (s *timeIntervalScheduler) stop() {
	s.stopped = true
}

type ITaskSchedulerObserver interface {
	GetKey() string
	//获取值
	GetTaskItemValue() interface{}
	//如果回调返回了error,用来获取其error
	Error() error
	AddCompleteCallbacks(callbacks ...func(ITaskSchedulerObserver))
	IsStopped() bool

	Stop() bool
}

type taskSchedulerObserver struct {
	timer     *timingwheel.Timer
	host      *taskScheduler
	scheduler *timeIntervalScheduler

	completeCallbackList []func(ITaskSchedulerObserver)
	taskItem             *TaskItem
	err                  error
}

var _ ITaskSchedulerObserver = (*taskSchedulerObserver)(nil)

func newTaskSchedulerObserver(host *taskScheduler) *taskSchedulerObserver {
	return &taskSchedulerObserver{
		host:                 host,
		completeCallbackList: make([]func(ITaskSchedulerObserver), 0),
	}
}

// #region ITaskSchedulerObserver Members

func (o *taskSchedulerObserver) GetKey() string {
	return o.timer.GetKey()
}

// 获取值
func (o *taskSchedulerObserver) GetTaskItemValue() interface{} {
	return o.taskItem.Value
}

// 如果回调返回了error,用来获取其error
func (o *taskSchedulerObserver) Error() error {
	return o.err
}

func (o *taskSchedulerObserver) IsStopped() bool {
	return o.scheduler == nil || o.scheduler.stopped
}

func (o *taskSchedulerObserver) Stop() bool {
	if o.scheduler != nil {
		o.scheduler.stop()
	}

	if o.timer == nil {
		o.host.schedulerObserverList.Del(o.taskItem.key)
		return true
	}
	result := o.timer.Stop()
	if result {
		o.host.schedulerObserverList.Del(o.taskItem.key)
	}
	return result
}

func (o *taskSchedulerObserver) AddCompleteCallbacks(callbacks ...func(ITaskSchedulerObserver)) {
	if callbacks == nil || len(callbacks) <= 0 {
		return
	}
	o.completeCallbackList = append(o.completeCallbackList, callbacks...)
}

func (o *taskSchedulerObserver) notifyCompleted() {
	for _, eachCallback := range o.completeCallbackList {
		eachCallback(o)
	}
}

// #endregion
