package scheduler

import "time"

var (
	_taskScheduler ITaskScheduler
)

// 调度一个函数，此函数按照interval时间定期执行,这个定时器的回调是不会存在着并行执行的
// 下一个定时的触发机制是在这个回调执行完成后再开始计时
// 返回用于此调度的observer对象
func SchedulerTask(d time.Duration,
	taskId string,
	timerFunc func(string) error,
	finishedCallback ...func(ITaskSchedulerObserver)) ITaskSchedulerObserver {

	taskItem := NewTaskItem()
	taskItem.Value = taskId
	observer := _taskScheduler.SchedulerFuncOneByOne(d, taskItem, func(ti *TaskItem) error {
		currentItemValue := ti.Value.(string)
		return timerFunc(currentItemValue)
	}, finishedCallback...)
	return observer
}

// 调度一个函数,指定时间后执行
// timerFunc: 要执行的回调
// finishedCallback: 函数执行完成后的回调
func AfterFunc(d time.Duration,
	taskId string,
	timerFunc func(string) error,
	finishedCallback ...func(string, error)) ITaskSchedulerObserver {

	taskItem := NewTaskItem()
	taskItem.Value = taskId
	observer := _taskScheduler.AfterFunc(d, taskItem, func(ti *TaskItem) error {
		currentItemValue := ti.Value.(string)
		return timerFunc(currentItemValue)
	}, func(iso ITaskSchedulerObserver) {
		if len(finishedCallback) <= 0 {
			return
		}
		currentItemValue := iso.GetTaskItemValue().(string)
		timerErr := iso.Error()
		for _, eachFunc := range finishedCallback {
			eachFunc(currentItemValue, timerErr)
		}
	})
	return observer
}
