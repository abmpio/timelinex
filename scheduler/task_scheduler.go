package scheduler

import (
	"time"

	"github.com/abmpio/threadingx/collection"
	"github.com/abmpio/threadingx/threading"
	"github.com/abmpio/threadingx/timingwheel"
)

const (
	slots                    = 300
	defaultTimeWheelInterval = time.Millisecond * 16
)

type ITaskScheduler interface {
	//多久后执行回调
	AfterFunc(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行,这个定时器的回调是可能会存在着并行执行的
	//返回用于此任务的调度key
	SchedulerFunc(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	// 调度一个函数，此函数按照interval时间定期执行,这个定时器的回调是不会存在着并行执行的
	// 下一个定时的触发机制是在这个回调执行完成后再开始计时
	//返回用于此任务的调度key
	SchedulerFuncOneByOne(interval time.Duration, taskItem *TaskItem, callback func(*TaskItem) error, completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver

	//停止指定的调度项,如果key不存在，则返回false
	StopScheduler(key string) bool
}

type taskScheduler struct {
	timingWheel *timingwheel.TimingWheel

	schedulerObserverList *collection.SafeMap
}

var _ ITaskScheduler = (*taskScheduler)(nil)

func NewTaskScheduler() ITaskScheduler {
	scheduler := &taskScheduler{
		schedulerObserverList: collection.NewSafeMap(),
	}
	scheduler.timingWheel = timingwheel.NewTimingWheel(time.Millisecond, slots)
	scheduler.timingWheel.Start()
	return scheduler
}

func (s *taskScheduler) _afterFunc(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	observer *taskSchedulerObserver) {
	taskItem.ensureHasKey()
	t := s.timingWheel.AfterFunc(interval, func() {
		defer func() {
			//执行完成后删除key
			s.schedulerObserverList.Del(taskItem.key)
		}()
		//触发回调
		threading.SafeCallFunc(func() {
			if callback != nil {
				err := callback(taskItem)
				observer.err = err
			}
		})
		threading.SafeCallFunc(observer.notifyCompleted)
	}).SetKey(taskItem.key)
	observer.timer = t
	//增加到待执行的列表中
	s.schedulerObserverList.Set(taskItem.key, observer)
}

// #region ITaskScheduler Members

// 调度一个函数,指定时间后执行
// 只执行一次
func (s *taskScheduler) AfterFunc(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.AddCompleteCallbacks(completeOpts...)

	s._afterFunc(interval, taskItem, callback, observer)
	return observer
}

// 调度一个函数，此函数按照interval时间定期执行
// 返回用于此任务的调度key
// 这个函数，可能存在着callback并发执行的情况
func (s *taskScheduler) SchedulerFunc(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	taskItem.ensureHasKey()
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.scheduler = scheduler
	observer.AddCompleteCallbacks(completeOpts...)

	t := s.timingWheel.ScheduleFuncWith(scheduler, taskItem.key, func() {
		//触发回调
		threading.SafeCallFunc(func() {
			if callback != nil {
				err := callback(taskItem)
				observer.err = err
			}
		})
		threading.SafeCallFunc(observer.notifyCompleted)
	})
	if t == nil {
		return nil
	}
	s.schedulerObserverList.Set(taskItem.key, observer)
	return observer
}

func (s *taskScheduler) SchedulerFuncOneByOne(interval time.Duration,
	taskItem *TaskItem,
	callback func(*TaskItem) error,
	completeOpts ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	taskItem.ensureHasKey()
	scheduler := &timeIntervalScheduler{
		interval: interval,
	}
	observer := newTaskSchedulerObserver(s)
	observer.taskItem = taskItem
	observer.scheduler = scheduler
	observer.AddCompleteCallbacks(completeOpts...)

	compCallback := func(to ITaskSchedulerObserver) {
		//check IsStopped
		if to.IsStopped() {
			return
		}
		//再次启动
		s._afterFunc(interval, taskItem, callback, observer)
	}
	observer.AddCompleteCallbacks(compCallback)

	s._afterFunc(interval, taskItem, callback, observer)
	return observer
}

// 移除指定的调度项,如果key不存在，则返回false
func (s *taskScheduler) StopScheduler(key string) bool {
	observerValue, ok := s.schedulerObserverList.Get(key)
	if !ok {
		return false
	}
	observer := observerValue.(*taskSchedulerObserver)
	observer.Stop()
	return true
}

// #endregion
