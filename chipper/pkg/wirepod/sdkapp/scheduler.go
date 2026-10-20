package sdkapp

// Per-robot active hours and play mode.
//
// Outside a robot's active hours it is sent to its charger, then wire-pod holds
// RESERVE_CONTROL (the lowest-priority SDK control level, meant for disabling
// idle behaviors) so it stays quiet. During active hours with play mode on, a
// built-in explore/social behavior is started every PlayIntervalMinutes.

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const (
	defaultPlayIntervalMinutes = 20
	// how long to let a robot drive home before holding control anyway
	chargerTimeout = 5 * time.Minute
)

// Built-in behaviors started in play mode. Exploring and seeking people are
// listed more than once so they come up more often than tricks.
var playIntents = []string{
	"intent_explore_start",
	"intent_explore_start",
	"intent_imperative_come",
	"intent_imperative_come",
	"intent_imperative_lookatme",
	"intent_imperative_lookatme",
	"intent_play_anytrick",
	"intent_play_fistbump",
	"intent_imperative_dance",
}

type schedState struct {
	vec          *vector.Vector
	quietSince   time.Time // zero while active
	holding      bool
	releaseHold  context.CancelFunc
	nextPlay     time.Time
	lastSchedule vars.RobotSchedule
	busy         bool // in a play-together session, owned by that goroutine
}

var (
	schedMu     sync.Mutex
	schedStates = map[string]*schedState{}
)

// schedLog goes to the debug log and the web UI's log page.
func schedLog(msg string) {
	logger.Println("Scheduler: " + msg)
	logger.LogUI("Scheduler: " + msg)
}

func InitScheduler() {
	go func() {
		for {
			schedulerTick(time.Now())
			time.Sleep(time.Minute)
		}
	}()
}

// ParseClock parses "HH:MM" into minutes after midnight.
func ParseClock(clock string) (int, error) {
	var h, m int
	if _, err := fmt.Sscanf(strings.TrimSpace(clock), "%d:%d", &h, &m); err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, fmt.Errorf("invalid time %q, expected HH:MM", clock)
	}
	return h*60 + m, nil
}

// InActiveHours reports whether now falls in [start, end). A window with
// start > end spans midnight; start == end means always active.
func InActiveHours(now time.Time, start, end string) bool {
	s, err1 := ParseClock(start)
	e, err2 := ParseClock(end)
	if err1 != nil || err2 != nil || s == e {
		return true
	}
	m := now.Hour()*60 + now.Minute()
	if s < e {
		return m >= s && m < e
	}
	return m >= s || m < e
}

func schedulerTick(now time.Time) {
	schedMu.Lock()
	defer schedMu.Unlock()
	seen := map[string]bool{}
	for _, bot := range vars.BotInfo.Robots {
		esn := bot.Esn
		seen[esn] = true
		st := schedStates[esn]
		if st == nil {
			st = &schedState{}
			schedStates[esn] = st
		}
		if st.busy {
			continue
		}
		var sched vars.RobotSchedule
		if bot.Schedule != nil {
			sched = *bot.Schedule
		}
		if sched != st.lastSchedule {
			// settings changed: play soon instead of waiting out the old interval
			st.nextPlay = time.Time{}
			st.lastSchedule = sched
		}
		if sched.ActiveHoursEnabled && !InActiveHours(now, sched.ActiveStart, sched.ActiveEnd) {
			goQuiet(esn, st, now)
			continue
		}
		wakeUp(esn, st)
		if sched.PlayEnabled && !now.Before(st.nextPlay) {
			interval := sched.PlayIntervalMinutes
			if interval <= 0 {
				interval = defaultPlayIntervalMinutes
			}
			if st.nextPlay.IsZero() {
				// first run: don't fire every robot the instant wire-pod starts
				st.nextPlay = now.Add(time.Duration(rand.Intn(interval)+1) * time.Minute)
				continue
			}
			if mate := ""; sched.PlayTogether && rand.Intn(2) == 0 {
				if mate = findPlaymate(esn, now); mate != "" {
					ms := schedStates[mate]
					st.busy, ms.busy = true, true
					ms.nextPlay = now.Add(time.Duration(interval) * time.Minute)
					go playTogether(esn, mate, st, ms)
				}
				if mate == "" {
					play(esn, st)
				}
			} else {
				play(esn, st)
			}
			// +/- 25% jitter so multiple robots don't move in lockstep
			jitter := time.Duration(rand.Intn(interval*30+1)-interval*15) * time.Second
			st.nextPlay = now.Add(time.Duration(interval)*time.Minute + jitter)
		}
	}
	// robots removed from wire-pod
	for esn, st := range schedStates {
		if !seen[esn] {
			wakeUp(esn, st)
			delete(schedStates, esn)
		}
	}
}

func robotConn(esn string, st *schedState) (*vector.Vector, error) {
	if st.vec != nil {
		return st.vec, nil
	}
	v, err := NewWP(esn, false)
	if err != nil {
		return nil, err
	}
	st.vec = v
	return v, nil
}

func battery(esn string, st *schedState) (*vectorpb.BatteryStateResponse, error) {
	v, err := robotConn(esn, st)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	resp, err := v.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
	if err != nil {
		st.vec = nil // reconnect next time (robot may have rebooted or changed IP)
	}
	return resp, err
}

func sendIntent(esn string, st *schedState, intent string) error {
	v, err := robotConn(esn, st)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := v.Conn.AppIntent(ctx, &vectorpb.AppIntentRequest{Intent: intent}); err != nil {
		st.vec = nil
		return err
	}
	return nil
}

func goQuiet(esn string, st *schedState, now time.Time) {
	if st.holding {
		return
	}
	if st.quietSince.IsZero() {
		st.quietSince = now
		schedLog(esn + " is outside its active hours, sending it to the charger")
		if b, err := battery(esn, st); err != nil {
			schedLog(esn + ": " + err.Error())
			return
		} else if !b.IsOnChargerPlatform {
			if err := sendIntent(esn, st, "intent_system_charger"); err != nil {
				schedLog(esn + ": " + err.Error())
			}
			return
		}
	} else if now.Sub(st.quietSince) < chargerTimeout {
		// still driving home? holding control now would interrupt it
		if b, err := battery(esn, st); err != nil || !b.IsOnChargerPlatform {
			return
		}
	}
	holdReserveControl(esn, st)
}

// holdReserveControl keeps RESERVE_CONTROL until st.releaseHold is called or
// the stream breaks (then the next tick takes it again).
func holdReserveControl(esn string, st *schedState) {
	v, err := robotConn(esn, st)
	if err != nil {
		schedLog(esn + ": " + err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := v.Conn.BehaviorControl(ctx)
	if err == nil {
		err = stream.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{Priority: vectorpb.ControlRequest_RESERVE_CONTROL},
			},
		})
	}
	if err != nil {
		cancel()
		st.vec = nil
		schedLog(esn + ": could not reserve control: " + err.Error())
		return
	}
	st.holding = true
	st.releaseHold = func() {
		stream.Send(&vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{ControlRelease: &vectorpb.ControlRelease{}},
		})
		cancel()
	}
	schedLog(esn + " is now in quiet hours")
	go func() {
		// drain responses until the stream ends (released, or robot disconnected)
		for {
			if _, err := stream.Recv(); err != nil {
				break
			}
		}
		if ctx.Err() != nil {
			return // released by wakeUp
		}
		cancel()
		schedMu.Lock()
		defer schedMu.Unlock()
		schedLog("lost quiet-hours control of " + esn + ", will retry")
		st.vec = nil
		st.holding = false
		st.releaseHold = nil
	}()
}

func wakeUp(esn string, st *schedState) {
	if st.releaseHold != nil {
		st.releaseHold()
		schedLog(esn + " active hours started, releasing control")
	}
	st.releaseHold = nil
	st.holding = false
	st.quietSince = time.Time{}
}

func play(esn string, st *schedState) {
	b, err := battery(esn, st)
	if err != nil {
		schedLog(esn + ": " + err.Error())
		return
	}
	// leave a low or still-charging robot alone
	if b.BatteryLevel < vectorpb.BatteryLevel_BATTERY_LEVEL_NOMINAL ||
		(b.IsOnChargerPlatform && b.BatteryLevel != vectorpb.BatteryLevel_BATTERY_LEVEL_FULL) {
		schedLog(esn + " battery not ready (" + b.BatteryLevel.String() + fmt.Sprintf(", on charger: %v", b.IsOnChargerPlatform) + "), skipping play")
		return
	}
	intent := playIntents[rand.Intn(len(playIntents))]
	schedLog("starting " + intent + " on " + esn)
	if err := sendIntent(esn, st, intent); err != nil {
		schedLog(esn + ": " + err.Error())
	}
}
