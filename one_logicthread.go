package timelinex

import "github.com/abmpio/timelinex/threading"

type OneLogicThread struct {
	ITimeline
	ISceneTimer

	logicThread *threading.LogicThread
}

func NewOneLogicThread() *OneLogicThread {
	t := &OneLogicThread{
		ITimeline:   newTimeline(),
		ISceneTimer: newSceneTimer(),

		logicThread: threading.NewLogicThread(),
	}

	t.ITimeline.(*timeline).setSceneTimer(t.ISceneTimer)
	t.ISceneTimer.(*sceneTimer).setTimeline(t.ITimeline)
	t.logicThread.AttatchWorkItem(t.ITimeline.(threading.IWorkItem))

	t.logicThread.Start()
	return t
}

func (t *OneLogicThread) Shutdown() {
	if t.ISceneTimer != nil {
		t.ISceneTimer.(*sceneTimer).Stop()
	}
	t.logicThread.Stop()
}
