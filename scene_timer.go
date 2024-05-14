package timelinex

import (
	"time"

	"github.com/abmpio/libx/lang/tuple"
	"github.com/abmpio/timelinex/scheduler"
)

type ISceneTimer interface {
	// 启动一个新的计时器(一次性触发的)
	// delayInterval 延时多久后触发回调
	// action 回调
	StartNewOneTimer(delayInterval time.Duration, action func(), opts ...SceneTimerOption) string

	//启动一个新的计时器(一次性触发的)
	// delayInterval 延时多久后触发回调
	// action 回调
	// data 回调函数所需的数据
	StartNewOneTimerWithData(delayInterval time.Duration, action func(interface{}), data interface{}, opts ...SceneTimerOption) string

	//启动一个新的计时器(一直在运行的)
	// timerInterval 间隔时间
	// action:回调
	StartRecurNewTimer(timerInterval time.Duration, action func(), opts ...SceneTimerOption) string

	//移除一个定时器
	RemoveTimer(timerId string)
}

type SceneTimerOption = func(i *scheduler.TaskItem)

func SceneTimerOptionWithKey(key string) SceneTimerOption {
	return func(i *scheduler.TaskItem) {
		i.SetKey(key)
	}
}

var _ ISceneTimer = (*sceneTimer)(nil)

type sceneTimer struct {
	taskScheduler scheduler.ITaskScheduler

	timeline ITimeline
}

func newSceneTimer() ISceneTimer {
	t := &sceneTimer{
		taskScheduler: scheduler.NewTaskScheduler(),
	}
	return t
}

func (s *sceneTimer) setTimeline(timeline ITimeline) {
	s.timeline = timeline
}

// #region ISceneTimerService Members

// 启动一个新的计时器(一次性触发的)
// delayInterval 延时多久后触发回调
// timerCallback 回调
// key: 与此计时器关联的key，为空时将自动设置
func (s *sceneTimer) StartNewOneTimer(delayInterval time.Duration,
	action func(),
	opts ...SceneTimerOption) string {

	taskItem := scheduler.NewTaskItem()
	taskItem.Value = action
	for _, eachOpt := range opts {
		eachOpt(taskItem)
	}
	observer := s.taskScheduler.AfterFunc(delayInterval, taskItem, func(ti *scheduler.TaskItem) error {
		v := ti.Value.(func())
		// 将这个回调执行在时间轴中
		s.timeline.SubscribeAsOnTime(Observer(v), nil)
		return nil
	})
	return observer.GetKey()
}

// 启动一个新的计时器(一次性触发的)
// delayInterval 延时多久后触发回调
// action 回调
// data 回调函数所需的数据
func (s *sceneTimer) StartNewOneTimerWithData(delayInterval time.Duration,
	action func(interface{}),
	data interface{},
	opts ...SceneTimerOption) string {

	taskItem := scheduler.NewTaskItem()
	taskItemValue := tuple.New2(action, data)
	taskItem.Value = &taskItemValue
	for _, eachOpt := range opts {
		eachOpt(taskItem)
	}
	observer := s.taskScheduler.AfterFunc(delayInterval, taskItem, func(ti *scheduler.TaskItem) error {
		tValue := ti.Value.(tuple.T2[func(interface{}), interface{}])
		// 将这个回调执行在时间轴中
		s.timeline.SubscribeAsOnTime(ObserverFromActionWithT[any](tValue.V1, tValue.V2), nil)
		return nil
	})
	return observer.GetKey()
}

// 启动一个新的计时器(一直在运行的)
func (s *sceneTimer) StartRecurNewTimer(timerInterval time.Duration, action func(), opts ...SceneTimerOption) string {

	taskItem := scheduler.NewTaskItem()
	taskItem.Value = action
	for _, eachOpt := range opts {
		eachOpt(taskItem)
	}

	//增加到调度队列中
	observer := s.taskScheduler.SchedulerFuncOneByOne(timerInterval, taskItem, func(ti *scheduler.TaskItem) error {
		aValue := ti.Value.(func())
		// 将这个回调执行在时间轴中
		s.timeline.SubscribeAsOnTime(Observer(aValue), nil)
		return nil
	})
	return observer.GetKey()
}

func (s *sceneTimer) RemoveTimer(timerId string) {
	s.taskScheduler.StopScheduler(timerId)
}

// #endregion
