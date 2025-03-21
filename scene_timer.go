package timelinex

import (
	"time"

	"github.com/abmpio/libx/lang/tuple"
	"github.com/abmpio/timelinex/scheduler"
)

const (
	taskItem_PropertiesKey_DontRunInTimelineThread = "dontRunInTimelineThread"
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

// 新的定时器回调不运行在时间轴线程，也就是逻辑线程
func SceneTimerOptionWithDontRunInTimelineThread() SceneTimerOption {
	return func(i *scheduler.TaskItem) {
		i.SetProperty(taskItem_PropertiesKey_DontRunInTimelineThread, true)
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

		pValue := ti.GetProperty(taskItem_PropertiesKey_DontRunInTimelineThread)
		dontRunInTimelineThread, ok := pValue.(bool)
		if ok && dontRunInTimelineThread {
			// 不运行在时间轴
			v()
		} else {
			// 将这个回调执行在时间轴中
			s.timeline.SubscribeAsOneTime(Observer(v), nil)
		}
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
	taskItem.Value = taskItemValue
	for _, eachOpt := range opts {
		eachOpt(taskItem)
	}
	observer := s.taskScheduler.AfterFunc(delayInterval, taskItem, func(ti *scheduler.TaskItem) error {
		tValue := ti.Value.(tuple.T2[func(interface{}), interface{}])

		pValue := ti.GetProperty(taskItem_PropertiesKey_DontRunInTimelineThread)
		dontRunInTimelineThread, ok := pValue.(bool)
		if ok && dontRunInTimelineThread {
			// 不运行在时间轴
			tValue.V1(tValue.V2)
		} else {
			// 将这个回调执行在时间轴中
			s.timeline.SubscribeAsOneTime(ObserverFromActionWithT[any](tValue.V1, tValue.V2), nil)
		}
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

		pValue := ti.GetProperty(taskItem_PropertiesKey_DontRunInTimelineThread)
		dontRunInTimelineThread, ok := pValue.(bool)
		if ok && dontRunInTimelineThread {
			// 不运行在时间轴
			aValue()
		} else {
			// 将这个回调执行在时间轴中
			s.timeline.SubscribeAsOneTime(Observer(aValue), nil)
		}
		return nil
	})
	return observer.GetKey()
}

func (s *sceneTimer) RemoveTimer(timerId string) {
	s.taskScheduler.StopScheduler(timerId)
}

// #endregion
