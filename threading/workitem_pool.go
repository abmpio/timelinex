package threading

type IWorkItemPool interface {
	WorkItemIsList() bool

	// 获取线程池中的线程的工作项列表
	GetWorkItemQueueCount() int

	// 获取下一个工作项
	GetNextWorkItem() []IWorkItem
}
