package timelinex

type ITimelineObserver interface {
	// 时间轴轮循通知
	// deltaSeconds 与上一次时间轴执行时相关的毫秒值,以ms为单位</param>
	OnNext(deltaMS float64)
}

var _ ITimelineObserver = (*DefaultTimelineObserver)(nil)
var _ ITimelineObserver = (*ActionTimelineObserver)(nil)
var _ ITimelineObserver = (*ActionWithTimelineObserver[any])(nil)

// 一个默认的ITimelineObserver实现
type DefaultTimelineObserver struct {
	action func(float64)
}

func NewDefaultTimelineObserver(action func(float64)) *DefaultTimelineObserver {
	return &DefaultTimelineObserver{
		action: action,
	}
}

// #region ITimelineObserver Members

func (o *DefaultTimelineObserver) OnNext(deltaMS float64) {
	o.action(deltaMS)
}

// #endregion

type ActionTimelineObserver struct {
	action func()
}

func NewActionTimelineObserver(action func()) *ActionTimelineObserver {
	return &ActionTimelineObserver{
		action: action,
	}
}

// #region ITimelineObserver Members

func (o *ActionTimelineObserver) OnNext(deltaMS float64) {
	if o.action == nil {
		return
	}
	o.action()
}

// #endregion

type ActionWithTimelineObserver[T any] struct {
	action func(v T)
	data   T
}

func NewActionWithTimelineObserver[T any](action func(v T), data T) *ActionWithTimelineObserver[T] {
	return &ActionWithTimelineObserver[T]{
		action: action,
		data:   data,
	}
}

// #region ITimelineObserver Members

func (o *ActionWithTimelineObserver[T]) OnNext(deltaMS float64) {
	if o.action == nil {
		return
	}
	o.action(o.data)
}

// #endregion
