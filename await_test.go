package await

import "testing"

func chanBlocked(t *testing.T, ch <-chan struct{}, name string) {
	select {
	case _ = <-ch:
		t.Errorf("Channel %s unblocked when it should be blocked")
	default:
	}
}

func chanUnblocked(t *testing.T, ch <-chan struct{}, name string) {
	select {
	case _ = <-ch:
	default:
		t.Errorf("Channel %s blocked when it should be unblocked")
	}
}

func TestSimple(t *testing.T) {
	as := NewAwaitServer("")
	a1 := as.New()
	a2 := as.New()

	chanBlocked(t, a1.Chan(), "a1")
	chanBlocked(t, a2.Chan(), "a2")

	if !as.wakeUp(a1.id) {
		t.Error("AwaitServer.wakeUp incorrectly returned false")
	}

	chanUnblocked(t, a1.Chan(), "a1")
	chanBlocked(t, a2.Chan(), "a2")

	if !as.wakeUp(a2.id) {
		t.Error("AwaitServer.wakeUp incorrectly returned false")
	}

	chanUnblocked(t, a1.Chan(), "a1")
	chanUnblocked(t, a2.Chan(), "a2")

	if as.wakeUp(a2.id) {
		t.Error("AwaitServer.wakeUp incorrectly returned true")
	}
}

func TestCancel(t *testing.T) {
	as := NewAwaitServer("")
	a := as.New()

	chanBlocked(t, a.Chan(), "a")

	a.Cancel()

	chanUnblocked(t, a.Chan(), "a")

	if as.wakeUp(a.id) {
		t.Error("AwaitServer.wakeUp returned true on a cancelled Await")
	}
}
