package timelinex

// 根据一个回调来创建ITimelineObserver实例
func Observer(action func()) ITimelineObserver {
	return NewActionTimelineObserver(action)
}

// 根据一个回调来创建ITimelineObserver实例
func ObserverFromAction(action func(float64)) ITimelineObserver {
	return NewDefaultTimelineObserver(action)
}

// 根据一个回调来创建ITimelineObserver实例
func ObserverFromActionWithT[T any](action func(v T), data T) ITimelineObserver {
	return NewActionWithTimelineObserver[T](action, data)
}
