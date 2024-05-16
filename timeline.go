package timelinex

import (
	"sync"
	"time"

	"github.com/abmpio/threadingx/collection"
	"github.com/abmpio/threadingx/lang"
	threadingx "github.com/abmpio/threadingx/threading"
	"github.com/abmpio/timelinex/threading"
)

// 时间轴服务,以每秒60帧的速率轮循(即16毫秒一次轮循)
type ITimeline interface {
	// 订阅时间轴轮循通知，以用来接收时间轴通知，这里订阅的将会一直工作在这个时间轴中，直到调用Unsubscribe(ITimelineObserver)方法取消订阅为止
	Subscribe(timelineObserver ITimelineObserver) ITimelineObserver

	// 订阅时间轴轮轮循通知，只通知一次，通知到达一次后在下次的时间轴中将不会再次通知，此方法会自动执行取消订阅
	// delayTime: 延时再执行，如果不设置，则立即执行
	SubscribeAsOneTime(timelineObserver ITimelineObserver, delayTime *time.Duration)

	// 取消原有的订阅
	Unsubscribe(timelineObserver ITimelineObserver)
}

// 主timeline
var (
	// 确保timeline实现了ITimeline与threading.IWorkItem两个接口
	_ ITimeline           = (*timeline)(nil)
	_ threading.IWorkItem = (*timeline)(nil)
)

// ITimeline的默认实现
type timeline struct {
	description string

	//只执行一次的observer队列
	registedOneTimeObserverQueue *collection.Queue
	//一直订阅的observer列表
	registedObserverList []ITimelineObserver
	// 这个列表来源于registedObserverList，不直接使用registedObserverList是防止多线程下的安全性
	workingObserverList []ITimelineObserver
	rwLock              sync.RWMutex

	isChanged          lang.SafeBool
	previousUpdateTime *time.Time
	scenseTimer        ISceneTimer
}

func newTimeline() ITimeline {
	timelineService := &timeline{
		description: "timeline",

		registedOneTimeObserverQueue: collection.NewQueue(10),
		registedObserverList:         make([]ITimelineObserver, 0),
		workingObserverList:          make([]ITimelineObserver, 0),
		rwLock:                       sync.RWMutex{},

		isChanged: lang.SafeBool{},
	}
	return timelineService
}

func (t *timeline) setSceneTimer(timer ISceneTimer) {
	t.scenseTimer = timer
}

// #region ITimeline Members

// 订阅时间轴轮轮循通知，以用来接收时间轴通知，这里订阅的将会一直工作在这个时间轴中，直到调用Unsubscribe(ITimelineObserver)方法取消订阅为止
func (t *timeline) Subscribe(timelineObserver ITimelineObserver) ITimelineObserver {
	if timelineObserver == nil {
		return nil
	}
	t.rwLock.Lock()
	defer t.rwLock.Unlock()

	t.registedObserverList = append(t.registedObserverList, timelineObserver)
	t.isChanged.Set(true)
	return timelineObserver
}

// 订阅时间轴轮轮循通知，只通知一次，通知到达一次后在下次的时间轴中将不会再次通知，此方法会自动执行取消订阅
func (t *timeline) SubscribeAsOneTime(timelineObserver ITimelineObserver, delayTime *time.Duration) {
	if timelineObserver == nil {
		return
	}

	if delayTime == nil || delayTime.Milliseconds() <= 0 {
		//立即执行，不延时
		t.registedOneTimeObserverQueue.Put(timelineObserver)
		return
	}
	t.scenseTimer.StartNewOneTimer(*delayTime, func() {
		t.registedOneTimeObserverQueue.Put(timelineObserver)
	})
}

// 取消原有的订阅
func (t *timeline) Unsubscribe(timelineObserver ITimelineObserver) {
	if timelineObserver == nil {
		return
	}
	t.rwLock.Lock()
	defer t.rwLock.Unlock()
	if len(t.registedObserverList) <= 0 {
		return
	}

	removeIndex := -1
	for i, eachObserver := range t.registedObserverList {
		if eachObserver == timelineObserver {
			removeIndex = i
			break
		}
	}
	if removeIndex > -1 {
		t.registedObserverList = removeSliceByIndex(t.registedObserverList, removeIndex)
	}
}

// #endregion

// #region threading.IWorkItem Members

func (t *timeline) Description() string {
	return t.description
}

// 通知observer
func (t *timeline) DoWork() {
	now := time.Now()
	if t.previousUpdateTime == nil {
		t.previousUpdateTime = &now
	}
	lastTime := t.previousUpdateTime
	t.previousUpdateTime = &now
	duration := time.Since(*lastTime)
	// 通知各个observer
	t._notifyRegistedObserver(float64(duration.Milliseconds()))
}

// #endregion

// / <summary>
// / 通知所有的订阅者
// / </summary>
func (t *timeline) _notifyRegistedObserver(deltaMS float64) {
	if t.isChanged.Get() {
		// 已经改变
		t.rwLock.Lock()
		t.workingObserverList = t.registedObserverList[:]
		t.isChanged.Set(false)
		t.rwLock.Unlock()
	}

	// 通知一次性订阅的observer
	oneTimeObserver := t.dequeueOneTimelineObserver()
	for oneTimeObserver != nil {
		threadingx.RunSafe(func() {
			oneTimeObserver.OnNext(deltaMS)
		})
		oneTimeObserver = t.dequeueOneTimelineObserver()
	}

	// 通知所有一直在订阅的observer
	workingObserverList := t.workingObserverList[:]
	for _, eachObserver := range workingObserverList {
		threadingx.RunSafe(func() {
			eachObserver.OnNext(deltaMS)
		})
	}
}

func (t *timeline) dequeueOneTimelineObserver() ITimelineObserver {
	v, ok := t.registedOneTimeObserverQueue.Take()
	if !ok {
		return nil
	}
	observer := v.(ITimelineObserver)
	return observer
}
