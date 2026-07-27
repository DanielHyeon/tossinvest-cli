package obs

// heartbeat.go is the liveness publish (harden-execution-base task 4.3,
// engine-safety: "죽은 프로세스는 스스로 통지할 수 없으므로 heartbeat … 방식을
// 사용한다").
//
// # Why the alarm lives on the other side
//
// Every failure mode this package handles assumes the engine is running to handle
// it. The one it cannot handle that way is the engine not running: a process that
// has been OOM-killed, a machine that rebooted, a container that never came back.
// No error handling inside the process covers that.
//
// So the engine publishes on a fixed interval and says nothing else, and the
// *receiver* is configured to alarm when a beat does not arrive on time (ntfy
// does this natively). The responsibility for noticing silence is thereby moved
// to something that is still alive.
//
// # Why a heartbeat is best-effort and its absence is not
//
// A missed publish is not an incident: the next one lands and the receiver's
// window is longer than one interval. Escalating a single failed heartbeat into a
// critical alert would make the liveness signal the noisiest thing in the system,
// which is how liveness signals get muted.

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// DefaultHeartbeatInterval is how often the engine says it is alive.
//
// A minute is short enough that an operator learns about a dead engine inside the
// window where a position can still be managed, and long enough to be nothing
// against the query budget (§0.4 — it is not a broker call at all).
const DefaultHeartbeatInterval = time.Minute

// Heartbeat publishes a periodic liveness beat.
type Heartbeat struct {
	// Publisher sends the beat. Required; nil makes Run a no-op.
	Publisher Publisher
	// Log records each beat and each failure. Optional.
	Log *Logger
	// Clock drives the interval. Defaults to clock.System().
	Clock clock.Clock
	// Interval is the publish period. Zero uses DefaultHeartbeatInterval.
	Interval time.Duration
	// Status is called for each beat to describe the engine's condition. Optional;
	// nil sends an empty body.
	Status func() string
}

// Beat publishes one heartbeat.
//
// It reports the publish error for a caller that wants it (the tests do), but
// nothing in the engine treats it as an incident — see the file comment.
func (h *Heartbeat) Beat(ctx context.Context) error {
	if h.Publisher == nil {
		return nil
	}
	body := ""
	if h.Status != nil {
		body = h.Status()
	}
	err := h.Publisher.Publish(ctx, Notification{
		Type:     EventEngineHeartbeat,
		Severity: SeverityNormal,
		Title:    "tossos engine alive",
		Body:     body,
		// Low priority: this one is meant to arrive silently and be noticed only
		// by its absence.
		Priority: 1,
	})
	if h.Log != nil {
		if err != nil {
			h.Log.Warn(EventEngineHeartbeat, FieldError, err.Error())
		} else {
			h.Log.Event(EventEngineHeartbeat)
		}
	}
	return err
}

// Run beats until the context ends.
//
// The first beat is immediate. An engine that starts, publishes nothing for a
// minute and then dies would otherwise be indistinguishable from one that never
// started — and the receiver's window would not have opened yet.
func (h *Heartbeat) Run(ctx context.Context) error {
	if h.Publisher == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	clk := h.Clock
	if clk == nil {
		clk = clock.System()
	}
	interval := h.Interval
	if interval <= 0 {
		interval = DefaultHeartbeatInterval
	}

	for {
		_ = h.Beat(ctx)
		if err := clk.Sleep(ctx, interval); err != nil {
			return err
		}
	}
}
