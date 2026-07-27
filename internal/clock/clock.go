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

// System returns the real clock backed by the operating system.
func System() Clock { return systemClock{} }

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

func (systemClock) Since(t time.Time) time.Duration { return time.Since(t) }

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
