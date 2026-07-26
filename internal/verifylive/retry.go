package verifylive

// retry.go is the one place this tool tries the same request twice, and it is
// restricted to reads on purpose.
//
// # Why it exists
//
// The first live run (2026-07-26) lost three steps to a broker 429 rather than to
// anything about the account: the client resolves its account sequence lazily on
// the first account-scoped GET, a soak cycle had used the same endpoints minutes
// earlier, and the resolution failed three times in a row inside 300ms
// (measurements.md M4). The window size and the aggregation unit behind that 429
// are still unmeasured — task 1.3 — so the response here is the most conservative
// thing that helps at all: a couple of long waits on a read, and nothing else.
//
// # Why a mutation is never retried
//
// A retried GET costs a request. A retried POST /api/v1/orders can cost a second
// live order: the broker may have accepted the first one and lost the response, and
// this tool cannot tell that apart from a refusal. The idempotency key that would
// make the difference is exactly what task 2.7 has not measured yet — replay is
// reported DISABLED until it is — so an automatic re-send would be the tool
// assuming the property it exists to verify.
//
// WORKFLOW §0.1 also settles it independently: the operator approved a list of N
// live requests. Sending one of them twice because the network was slow is not on
// that list.
//
// So the retry lives in a helper that only read paths call, mutate.go does not
// import it, and a test drives PlaceOrder into a 429 and asserts it was sent once.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// readRetryExtraAttempts is how many more times a read-only call is sent after a
// 429. Three attempts in total.
//
// Two, because the ceiling has to be a number somebody can multiply by the number
// of reads in a run and still recognise as a bounded delay. Until 1.3 measures the
// broker's actual window, more attempts would be guessing with the operator's rate
// budget.
const readRetryExtraAttempts = 2

// readRetryBackoff is the wait before extra attempt n, counted from zero.
//
// Fifteen and thirty seconds. The observed burst was three failures inside 300ms,
// so anything on a sub-second scale would simply reproduce it; these are chosen to
// step over a window measured in tens of seconds without turning a stalled read
// into a stalled afternoon.
func readRetryBackoff(extra int) time.Duration {
	if extra <= 0 {
		return 15 * time.Second
	}
	return 30 * time.Second
}

// readRetry performs one read-only broker call and retries it only when the
// broker answered 429.
//
// Every attempt is logged as its own Call, including the ones that failed: the
// evidence record is what task 1.3 reads to size the retry matrix, and a record
// that showed only the attempt that eventually worked would hide the measurement.
//
// sr may be nil — the plan's own read runs before any step exists — in which case
// the attempts are not logged, because there is no step to log them against.
func readRetry[T any](
	ctx context.Context,
	r *Runner,
	sr *stepRun,
	endpoint string,
	req any,
	do func(context.Context) (T, error),
	summarise func(T) any,
) (T, error) {
	var (
		out T
		err error
	)
	for extra := 0; ; extra++ {
		started := r.now()
		out, err = do(ctx)
		if sr != nil {
			var resp any
			if summarise != nil {
				resp = summarise(out)
			}
			sr.logCall(endpoint, req, started, r.now(), resp, err)
		}

		if !errors.Is(err, official.ErrRateLimited) || extra >= readRetryExtraAttempts {
			return out, err
		}
		wait := readRetryBackoff(extra)
		fmt.Fprintf(r.out, "  %s was rate limited; waiting %s and reading it again (%d of %d)\n",
			endpoint, wait, extra+1, readRetryExtraAttempts)
		if sleepErr := r.sleep(ctx, wait); sleepErr != nil {
			// The operator interrupted, or the deadline passed. The rate limit is
			// still the reason the read has no answer, so that is what is returned.
			return out, err
		}
	}
}
