// Package clock is the single time source for the order-execution path.
//
// Every time judgement the order-execution spec names — submission time windows,
// staleness thresholds, stabilisation intervals, SLO measurement points and
// trading-day boundaries — goes through the injected Clock here, and every
// market-relative judgement goes through the Market calendar in market.go. Two
// rules follow from that and are enforced by the tests next door:
//
//  1. No execution-path code calls time.Now() directly. A fake clock must be
//     able to drive a stabilisation loop or a staleness check to any instant
//     without wall-clock sleeps, otherwise those safety intervals are untestable
//     and therefore unverified.
//  2. No execution-path code interprets a wall-clock time in the process
//     timezone. The machine may run in UTC, KST or anything else; sessions are
//     defined in Asia/Seoul and America/New_York only.
//
// Clock.Now() always reports UTC so a persisted journal timestamp cannot carry a
// machine-local offset. Market conversions do the local-timezone work.
package clock

import (
	"context"
	"time"

	// The embedded timezone database is a fallback that time.LoadLocation uses
	// only when the host has no zoneinfo (scratch containers, distroless
	// images). Without it, a US session judgement on a DST boundary would fail
	// or silently fall back to UTC on such a host — for ~450KB of binary we
	// remove a whole class of "wrong market hours" incident.
	_ "time/tzdata"
)

// Clock is the injectable time source. Implementations must be safe for
// concurrent use.
type Clock interface {
	// Now returns the current instant in UTC.
	Now() time.Time
	// Since returns the duration elapsed since t according to this clock.
	Since(t time.Time) time.Duration
	// Sleep blocks for d or until ctx is done, returning ctx.Err() in the latter
	// case. A non-positive d returns immediately with ctx.Err().
	Sleep(ctx context.Context, d time.Duration) error
}

// leaseClock is an optional, process-local extension used only for elapsed
// leases. It deliberately does not change Clock: test clocks keep their
// deterministic Now/Since contract, while the system clock can retain Go's
// monotonic time reading without exposing it through persisted timestamps.
type leaseClock interface {
	leaseAnchor() time.Time
	leaseElapsed(time.Time) time.Duration
}

// LeaseAnchor captures an opaque process-local anchor for a short-lived lease.
// Callers must not persist it. Clocks without the optional lease extension use
// their ordinary Now/Since pair so test-controlled elapsed time remains exact.
func LeaseAnchor(c Clock) time.Time {
	if lease, ok := c.(leaseClock); ok {
		return lease.leaseAnchor()
	}
	return c.Now()
}

// LeaseElapsed reports elapsed time for an anchor returned by LeaseAnchor.
func LeaseElapsed(c Clock, anchor time.Time) time.Duration {
	if lease, ok := c.(leaseClock); ok {
		return lease.leaseElapsed(anchor)
	}
	return c.Since(anchor)
}

// System returns the real clock backed by the operating system.
func System() Clock { return systemClock{} }

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

func (systemClock) Since(t time.Time) time.Duration { return time.Since(t) }

// leaseAnchor intentionally retains the monotonic reading time.Now attaches.
// Now must remain UTC for journal timestamps, which strips that reading.
func (systemClock) leaseAnchor() time.Time { return time.Now() }

func (systemClock) leaseElapsed(anchor time.Time) time.Duration { return time.Since(anchor) }

func (systemClock) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
