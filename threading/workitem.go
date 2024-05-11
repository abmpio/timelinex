package threading

type IWorkItem interface {
	// 工作项的描述，可为空
	Description() string

	// 如何执行
	DoWork()
}

var _ IWorkItem = (*WorkItem)(nil)
var _ IWorkItem = (*WorkItemT[any])(nil)

type WorkItem struct {
	action      func()
	description string
}

func (i *WorkItem) SetDescription(description string) {
	i.description = description
}

// #region IWorkItem Members

func (i *WorkItem) Description() string {
	return i.description
}

func (i *WorkItem) DoWork() {
	if i.action == nil {
		return
	}
	i.action()
}

// #endregion

// 带泛型的IWorkItem实现
type WorkItemT[T any] struct {
	IWorkItem
	data T
}

func (i *WorkItemT[T]) Data() T {
	return i.data
}

func NewWorkItem(fn func()) IWorkItem {
	return &WorkItem{
		action: fn,
	}
}

func NewWorkItemT[T any](fn func(v T), data T) IWorkItem {
	return &WorkItemT[T]{
		IWorkItem: NewWorkItem(
			func() {
				if fn != nil {
					fn(data)
				}
			},
		),
		data: data,
	}
}
