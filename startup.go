package timelinex

import "github.com/abmpio/timelinex/threading"

var (
	//默认的timeline
	_globalTimeline    ITimeline
	_globalSceneTimer  ISceneTimer
	_globalLogicThread *threading.LogicThread
)

func init() {
	timeline := newTimeline()
	sceneTimer := newSceneTimer()
	_globalLogicThread = threading.NewLogicThread()

	timeline.setSceneTimer(_globalSceneTimer)
	sceneTimer.setTimeline(_globalTimeline)

	_globalSceneTimer = sceneTimer
	_globalTimeline = timeline
}

func GlobalTimeline() ITimeline {
	return _globalTimeline
}

func GlobalSceneTimer() ISceneTimer {
	return _globalSceneTimer
}

func GlobalLogicThread() *threading.LogicThread {
	return _globalLogicThread
}
