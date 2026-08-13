package reconcile

// ratelimit.go is what the restart recovery does when the broker says "not now"
// instead of "no" (a102 겹1).
//
// # Why a refusal is not an answer about the account
//
// The stabilisation rule asks one question: has the account stopped moving? Its
// evidence is snapshots. A 429 produces no snapshot — the reads never happened —
// so it is not evidence of instability, and counting it as a spent attempt
// answers a question nobody asked. That is why a refusal here consumes neither a
// stabilisation attempt nor a snapshot count; it consumes *time*, and time is
// what the budget below bounds.
//
// # Why waiting at all, and why exactly this long
//
// The same 429 is survivable everywhere else in this process. The periodic
// reconciliation hands it to the next period (2026-08-12T12:52:37Z, measured),
// and the read-only survey waits 15 seconds and reads again
// (internal/soak.ReadRetryBackoff). Recovery — the one path that holds the entry
// gate shut — was the only reader that treated it as fatal, and on
// 2026-08-13 02:03:30.545Z a boot-time 429 ended a restart recovery outright.
//
// A component that carries the protection must not give up before the one that
// only looks. So the backoff is the survey's first interval to the digit, and it
// is a floor rather than a tuned number.
//
// # Why a total budget rather than a retry count
//
// The number an operator can act on is "how long was there no protection", not
// "how many times did we ask". A budget also fails closed in the one case a
// count cannot describe: a broker that is refusing everything forever. When the
// budget is spent the recovery fails exactly as it does today, and the gate that
// New() latched shut stays shut.

import (
	"context"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// DefaultRateLimitBackoff is the wait after a rate-limited account read.
//
// It is internal/soak.ReadRetryBackoff(0) to the digit: fifteen seconds. The two
// readers hit the same broker with the same credential, and the one holding the
// entry gate shut must not be the impatient one. The value is duplicated rather
// than imported because internal/soak deliberately imports nothing that can
// place an order and this package is on the execution path.
const DefaultRateLimitBackoff = 15 * time.Second

// DefaultMaxRateLimitWait bounds the total time one recovery spends waiting out
// rate limits.
//
// Five minutes covers the KRX morning throttle window — measured in minutes, not
// hours — with room to spare, while still refusing to wait silently forever on a
// broker that is never going to answer.
const DefaultMaxRateLimitWait = 5 * time.Minute

// snapshotProgress is what one stabilisation run spent: readings that produced a
// snapshot, and time given up to rate limits.
//
// It travels back to Run so the Report can carry it. Nothing here is a new
// logging path — the numbers ride the report that already exists.
type snapshotProgress struct {
	// Taken is how many readings produced a snapshot. A refused read is not one.
	Taken int
	// RateLimitWaits is how many times a rate limit was waited out.
	RateLimitWaits int
	// RateLimitWaited is the total of those waits.
	RateLimitWaited time.Duration
}

// waitOutRateLimit spends one backoff on a rate-limited read, or explains why it
// will not.
//
// The budget is checked *before* sleeping, so the wait never overshoots what an
// operator was told the ceiling is. Cancellation goes through the same
// clock.Sleep seam the stabilisation interval uses, which is what keeps a
// shutdown immediate (안전 불변식 4).
func (r *Recovery) waitOutRateLimit(
	ctx context.Context,
	clk clock.Clock,
	progress *snapshotProgress,
) error {
	backoff := r.stab.RateLimitBackoff
	if spent := progress.RateLimitWaited + backoff; spent > r.stab.MaxRateLimitWait {
		return fmt.Errorf(
			"%w: the account read kept being refused with a rate limit; "+
				"waited %s of the %s budget and stopped rather than overspend it",
			ErrRecoveryIncomplete, progress.RateLimitWaited, r.stab.MaxRateLimitWait)
	}
	if err := clk.Sleep(ctx, backoff); err != nil {
		return fmt.Errorf("%w: waiting out a rate limit: %v", ErrRecoveryIncomplete, err)
	}
	progress.RateLimitWaits++
	progress.RateLimitWaited += backoff
	return nil
}
