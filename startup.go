package timelinex

import "github.com/abmpio/timelinex/threading"

var (
	//默认的timeline
	_globalTimeline    ITimeline
	_globalSceneTimer  ISceneTimer
	_globalLogicThread *threading.LogicThread
)

func init() {
	_globalTimeline = newTimeline()
	_globalSceneTimer = newSceneTimer()
	_globalLogicThread = threading.NewLogicThread()

	_globalTimeline.(*timeline).setSceneTimer(_globalSceneTimer)
	_globalSceneTimer.(*sceneTimer).setTimeline(_globalTimeline)
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
