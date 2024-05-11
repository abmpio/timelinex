package threading

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/abmpio/threadingx/lang"
	"github.com/abmpio/threadingx/rescue"
	"github.com/lithammer/shortuuid/v4"
)

type workItemThreadOptions struct {
	id string
	// 线程池请求执行项的超时时间,0为永不超时
	abortThreadTimeout time.Duration
	//当执行时间超过这个时间后将报警,单位毫秒
	warningWhenWorkItemDurationMs int64
	//线程每个方法执行的间隔时间
	threadWorkItemInterval time.Duration
}

func newWorkItemThreadOptions() *workItemThreadOptions {
	return &workItemThreadOptions{
		id:                            shortuuid.New(),
		abortThreadTimeout:            time.Duration(0),
		warningWhenWorkItemDurationMs: 500,
		threadWorkItemInterval:        16 * time.Microsecond,
	}
}

type ThreadOption func(o *workItemThreadOptions)

func ThreadOptionWithAbortThreadTimeoutMS(abortThreadTimeoutMS int32) ThreadOption {
	return func(o *workItemThreadOptions) {
		o.abortThreadTimeout = time.Duration(int64(time.Millisecond) * int64(abortThreadTimeoutMS))
	}
}

func ThreadOptionWithName(id string) ThreadOption {
	return func(o *workItemThreadOptions) {
		o.id = id
	}
}

type WorkItemThread struct {
	*workItemThreadOptions
	pool IWorkItemPool

	_shutdown  lang.SafeBool
	_running   bool
	rw         sync.RWMutex
	_lastStart *time.Time
}

// new NewWorkItemThread instance
func NewWorkItemThread(pool IWorkItemPool, opts ...ThreadOption) *WorkItemThread {
	options := newWorkItemThreadOptions()
	for _, eachOpt := range opts {
		eachOpt(options)
	}

	t := &WorkItemThread{
		workItemThreadOptions: options,
		pool:                  pool,

		_shutdown: lang.SafeBool{},
		_running:  false,
		rw:        sync.RWMutex{},
	}
	return t
}

// 请求执行工作项的超时时间,以豪秒为单位
// 0表示永不超时
func (t *WorkItemThread) SetOptions(opts ...ThreadOption) {
	for _, eachOpt := range opts {
		eachOpt(t.workItemThreadOptions)
	}
}

// 获取最后一个工作项启动时间
func (t *WorkItemThread) LastWorkItemStartTime() *time.Time {
	return t._lastStart
}

// 时间间隔(毫秒)，默认为16ms，即每秒60秒
func (t *WorkItemThread) ThreadWorkItemIntervalMs() int64 {
	return t.threadWorkItemInterval.Microseconds()
}

// 停此线程
func (t *WorkItemThread) Stop() {
	t._shutdown.Set(true)
}

// 启动线程
func (t *WorkItemThread) Start() {
	t.rw.Lock()
	defer t.rw.Unlock()

	if t._running {
		// running
		return
	}
	t._running = true
	t._shutdown.Set(false)
	//创建协程并等待协程被启动
	t.createGoRoutine()
}

// 启动协程
func (t *WorkItemThread) createGoRoutine() {
	go func() {
		t.workToDo()
	}()
}

func (t *WorkItemThread) workToDo() {

	for !t._shutdown.Get() {
		// 等待时间片断，默认为16毫秒，即每秒60帧
		time.Sleep(t.threadWorkItemInterval)
		nextWorkItems := t.pool.GetNextWorkItem()
		for {
			if len(nextWorkItems) <= 0 {
				break
			}
			if !t.pool.WorkItemIsList() {
				count := t.pool.GetWorkItemQueueCount()
				if count > 10 {
					fmt.Printf("线程池中堆积未处理的线程已经超过10个,当前数量:%d", count)
				}
				if !t._shutdown.Get() {
					t.doWorkItem(nextWorkItems[0])
					nextWorkItems = t.pool.GetNextWorkItem()
				} else {
					//已经shutdown，则直接退出
					break
				}
			} else {
				//一次拿多个工作项的，则循环拿到的工作项列表，如LogicThread就是这种方式
				for i := 0; i < len(nextWorkItems); i++ {
					if !t._shutdown.Get() {
						t.doWorkItem(nextWorkItems[i])
					}
				}
				//跳出循环，等待下一次进来
				break
			}
		}
	}

	t.rw.Lock()
	t._running = false
	t.rw.Unlock()
}

func (t *WorkItemThread) doWorkItem(workItem IWorkItem) {
	t.rw.Lock()
	now := time.Now()
	t._lastStart = &now
	t.rw.Unlock()

	defer rescue.Recover()

	if t.abortThreadTimeout <= 0 || t.abortThreadTimeout == math.MaxInt64 {
		workItem.DoWork()
	} else {
		err := DoWithTimeout(func() error {
			workItem.DoWork()
			return nil
		}, t.abortThreadTimeout)
		if err != nil {
			fmt.Printf("doWorkItem timeout,id:%s,err:%s",
				t.id,
				err.Error())
		}
	}
	workItemDurationMs := time.Since(*t._lastStart).Milliseconds()
	if workItemDurationMs >= t.warningWhenWorkItemDurationMs {
		fmt.Printf("线程性能警报,id:%s 任务耗时过长,任务: %s 耗时ms:%d",
			t.id,
			workItem.Description(),
			workItemDurationMs)
	}
}
