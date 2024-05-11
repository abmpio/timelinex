package threading

import "sync"

type LogicThread struct {
	_thread *WorkItemThread

	rwLock sync.RWMutex
	//用于在线程池中执行的任务项队列
	addedWorkItemList []IWorkItem
	//正在执行的任各项队列
	workingItemList []IWorkItem
}

var _ IWorkItemPool = (*LogicThread)(nil)

func NewLogicThread() *LogicThread {
	t := &LogicThread{
		rwLock:            sync.RWMutex{},
		addedWorkItemList: make([]IWorkItem, 0),
		workingItemList:   make([]IWorkItem, 0),
	}
	t._thread = NewWorkItemThread(t)
	return t
}

func (t *LogicThread) Start() {
	t._thread.Start()
}

func (t *LogicThread) Stop() {
	t._thread.Stop()
}

// 往逻辑线程中放置一个任务
func (t *LogicThread) AttatchWorkItem(workItem IWorkItem) {
	t.rwLock.Lock()
	defer t.rwLock.Unlock()

	t.addedWorkItemList = append(t.addedWorkItemList, workItem)
	t.workingItemList = t.addedWorkItemList[:]
}

func (t *LogicThread) DetatchWorkItem(workItem IWorkItem) {
	t.rwLock.Lock()
	defer t.rwLock.Unlock()

	set := make([]IWorkItem, 0)
	for _, eachItem := range t.addedWorkItemList {
		if eachItem == workItem {
			//被排除的
			continue
		}
		set = append(set, eachItem)
	}
	t.addedWorkItemList = set
	t.workingItemList = t.addedWorkItemList[:]
}

// #region IWorkItemPool Members

func (t *LogicThread) WorkItemIsList() bool {
	return true
}

// 获取线程池中的线程的工作项列表
func (t *LogicThread) GetWorkItemQueueCount() int {
	return len(t.workingItemList)
}

// 获取下一个工作项
func (t *LogicThread) GetNextWorkItem() []IWorkItem {
	return t.workingItemList
}

// #endregion
