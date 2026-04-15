package timelinex

import "testing"

type testTimelineObserver struct {
	hits int
}

func (o *testTimelineObserver) OnNext(deltaMS float64) {
	o.hits++
}

func TestTimelineUnsubscribeStopsFutureNotifications(t *testing.T) {
	timeline := newTimeline().(*timeline)

	observerA := &testTimelineObserver{}
	observerB := &testTimelineObserver{}

	timeline.Subscribe(observerA)
	timeline.Subscribe(observerB)

	timeline._notifyRegistedObserver(16)
	if observerA.hits != 1 || observerB.hits != 1 {
		t.Fatalf("expected both observers to be called once, got A=%d B=%d", observerA.hits, observerB.hits)
	}

	timeline.Unsubscribe(observerA)
	timeline._notifyRegistedObserver(16)

	if observerA.hits != 1 {
		t.Fatalf("expected unsubscribed observer to stop receiving notifications, got %d hits", observerA.hits)
	}
	if observerB.hits != 2 {
		t.Fatalf("expected remaining observer to continue receiving notifications, got %d hits", observerB.hits)
	}
}

func TestTimelineWorkingObserverSnapshotIsDetachedFromRegistry(t *testing.T) {
	timeline := newTimeline().(*timeline)

	observerA := &testTimelineObserver{}
	observerB := &testTimelineObserver{}
	observerC := &testTimelineObserver{}

	timeline.Subscribe(observerA)
	timeline.Subscribe(observerB)
	timeline.Subscribe(observerC)
	timeline._notifyRegistedObserver(16)

	if len(timeline.workingObserverList) != 3 {
		t.Fatalf("expected working snapshot size 3, got %d", len(timeline.workingObserverList))
	}
	if timeline.workingObserverList[0] != observerA ||
		timeline.workingObserverList[1] != observerB ||
		timeline.workingObserverList[2] != observerC {
		t.Fatalf("unexpected initial working observer snapshot")
	}

	timeline.Unsubscribe(observerB)

	// 当前 working snapshot 应该保持稳定，直到下一个 tick 刷新。
	if len(timeline.workingObserverList) != 3 {
		t.Fatalf("expected detached working snapshot to keep size 3 before refresh, got %d", len(timeline.workingObserverList))
	}
	if timeline.workingObserverList[0] != observerA ||
		timeline.workingObserverList[1] != observerB ||
		timeline.workingObserverList[2] != observerC {
		t.Fatalf("expected unsubscribe not to mutate current working snapshot in place")
	}

	timeline._notifyRegistedObserver(16)

	if len(timeline.workingObserverList) != 2 {
		t.Fatalf("expected refreshed working snapshot size 2, got %d", len(timeline.workingObserverList))
	}
	if timeline.workingObserverList[0] != observerA || timeline.workingObserverList[1] != observerC {
		t.Fatalf("expected refreshed working snapshot to contain remaining observers in order")
	}
}
