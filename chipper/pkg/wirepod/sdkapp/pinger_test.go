package sdkapp

import (
	"testing"
)

// withEmptyPingerList clears the tracked-robot list and restores it afterwards.
// JdocsPingerBots is a package-level global with its own mutex, so tests
// touching it must not run in parallel.
func withEmptyPingerList(t *testing.T) {
	t.Helper()
	JdocsPingerBots.mu.Lock()
	restore := JdocsPingerBots.Robots
	JdocsPingerBots.Robots = nil
	JdocsPingerBots.mu.Unlock()

	t.Cleanup(func() {
		JdocsPingerBots.mu.Lock()
		JdocsPingerBots.Robots = restore
		JdocsPingerBots.mu.Unlock()
	})
}

// pingerState reports the tracked state for esn, or ok=false when untracked.
func pingerState(esn string) (stopped bool, since int, ok bool) {
	JdocsPingerBots.mu.Lock()
	defer JdocsPingerBots.mu.Unlock()
	for _, bot := range JdocsPingerBots.Robots {
		if bot.ESN == esn {
			return bot.Stopped, bot.TimeSinceLastCheck, true
		}
	}
	return false, 0, false
}

// The first check for a robot starts tracking it and pulls jdocs, which is what
// appends the escape-pod CA to the cert store.
func TestFirstCheckStartsTrackingAndPings(t *testing.T) {
	withEmptyPingerList(t)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	if !ShouldPingJdocsForESN("00e20145") {
		t.Fatal("the first check for an untracked robot must ping jdocs")
	}
	stopped, _, ok := pingerState("00e20145")
	if !ok {
		t.Fatal("the robot was not added to the tracked list")
	}
	if stopped {
		t.Error("a freshly tracked robot must not be marked stopped")
	}
}

// A robot checking in normally does not need a jdocs pull on every check; the
// call exists to reset its timer.
func TestSteadyCheckResetsTheTimerWithoutPinging(t *testing.T) {
	withEmptyPingerList(t)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	ShouldPingJdocsForESN("00e20145")

	JdocsPingerBots.mu.Lock()
	JdocsPingerBots.Robots[0].TimeSinceLastCheck = 9
	JdocsPingerBots.mu.Unlock()

	if ShouldPingJdocsForESN("00e20145") {
		t.Error("a robot already checking in must not trigger a jdocs pull every time")
	}
	if _, since, _ := pingerState("00e20145"); since != 0 {
		t.Errorf("TimeSinceLastCheck = %d, want 0; the check must reset the timer", since)
	}
}

// The case the whole mechanism exists for: the ticker marks a silent robot
// stopped, and the next check has to revive it and pull jdocs.
func TestCheckAfterAGapRevivesTheRobotAndPings(t *testing.T) {
	withEmptyPingerList(t)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	ShouldPingJdocsForESN("00e20145")
	JdocsPingerBots.mu.Lock()
	JdocsPingerBots.Robots[0].Stopped = true
	JdocsPingerBots.Robots[0].TimeSinceLastCheck = 16
	JdocsPingerBots.mu.Unlock()

	if !ShouldPingJdocsForESN("00e20145") {
		t.Fatal("a robot that had gone quiet must pull jdocs when it returns")
	}
	stopped, since, _ := pingerState("00e20145")
	if stopped {
		t.Error("the robot should no longer be marked stopped")
	}
	if since != 0 {
		t.Errorf("TimeSinceLastCheck = %d, want 0", since)
	}
}

// An ESN that names no stored robot must not create tracking state. connCheck
// is not adoption, and the list is keyed by ESN, so inventing entries here
// would leak an unbounded set of them.
func TestUnknownESNIsNotTracked(t *testing.T) {
	withEmptyPingerList(t)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	if ShouldPingJdocsForESN("00ffffff") {
		t.Error("an unknown ESN must not trigger a jdocs pull")
	}
	if _, _, ok := pingerState("00ffffff"); ok {
		t.Error("an unknown ESN must not be added to the tracked list")
	}
}

// Entries are matched case-sensitively against the tracked list (== at the
// lookup) but resolved case-insensitively against BotInfo (EqualFold). That
// only holds together because the stored spelling is canonicalised on the way
// in; if it ever stops being, a robot gains a duplicate entry per case variant
// and every check pulls jdocs afresh.
func TestCaseVariantESNResolvesToTheSameTrackedRobot(t *testing.T) {
	withEmptyPingerList(t)
	withRobots(t, struct{ esn, ip string }{"00e20145", "192.168.1.150"})

	ShouldPingJdocsForESN("00e20145")
	ShouldPingJdocsForESN("00E20145")

	JdocsPingerBots.mu.Lock()
	n := len(JdocsPingerBots.Robots)
	JdocsPingerBots.mu.Unlock()
	if n != 1 {
		t.Errorf("tracked list holds %d entries for one robot, want 1; a case "+
			"variant created a duplicate", n)
	}
}
