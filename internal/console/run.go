package console

// run.go holds the one verification a console process may drive, and the web
// confirmer that gates it.
//
// # The confirmer is the whole point of the file
//
// verifylive.BatchConfirmer is a function that is handed the complete plan and
// returns nil to approve it. At a terminal it prints the plan and reads a line
// (verifylive.ConfirmBatch). Here it parks the plan where the page can render it
// and blocks until somebody types the nonce back through a form that already had
// to carry the session cookie and the CSRF token.
//
// The comparison itself is not reimplemented: verifylive.Batch.Verify is what
// decides, so the console cannot drift into accepting something the terminal
// would not. Its two refusals are kept apart for the same reason they are at a
// terminal — an expiry is a window that closed, a mismatch is a person who said
// no — and both mean the runner returns without sending anything.
//
// A mismatch aborts the run rather than offering another try. That is the TTY
// contract ("anything else aborts the run"), and re-asking would turn a nonce with
// a five-minute window into one with five minutes of guesses. Nothing was sent, so
// the cost of a typo is starting the form again.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// Refusals the console itself produces, before the runner is involved.
var (
	// errProcessSpent is the process boundary. It is not an inconvenience to be
	// worked around: conditional-persist measures whether the broker still holds
	// the protection after this process has exited, and a second run inside the
	// same process could not observe that.
	errProcessSpent = errors.New(
		"이 콘솔 프로세스는 이미 검증을 한 번 수행했다. 조건주문 존속(conditional-persist)은 " +
			"등록한 프로세스가 죽은 뒤에야 관측할 수 있으므로, 콘솔을 종료(Ctrl-C)하고 다시 시작한 뒤 " +
			"이어하기를 눌러라. 지금 상태에서 다시 시작해도 새 프로세스가 아니므로 측정이 되지 않는다")
	// errRunInProgress means a verification is already walking the catalogue.
	errRunInProgress = errors.New("검증이 이미 진행 중이다")
)

// maxLogLines bounds what one run keeps in memory. The evidence is the JSONL
// record; this is a progress view, and an unbounded slice behind an HTTP handler
// is a memory leak with a nice interface.
const maxLogLines = 2000

// runState is one verification in progress, and everything the page renders
// about it.
type runState struct {
	id        string
	startedAt time.Time
	cancelFn  context.CancelFunc
	done      chan struct{}
	now       func() time.Time
	// market is the market this run sends orders in. It is fixed at start: a run
	// that could change market mid-flight would be a run whose approved plan and
	// whose evidence record disagreed about what was measured.
	market string
	// redo is the re-measurement set this run was started with, empty for an
	// ordinary run. It is kept so the progress page can say which mode it is in
	// after the start screen is gone.
	redo []verifylive.StepID

	mu sync.Mutex
	// pending is the plan waiting for an answer, nil when nothing is.
	pending *verifylive.Batch
	// decided carries that answer from the HTTP handler to the runner. It is
	// buffered so a handler never blocks on a runner that has already given up.
	decided chan error
	// approved records that the plan was approved, and when.
	approved   bool
	approvedAt time.Time
	// notice is the last thing that went wrong, for the page.
	notice string

	lines   []string
	partial string

	finishedFlag bool
	summary      verifylive.Summary
	err          error
	steps        int
}

// runView is a consistent copy for a template. Templates render concurrently
// with the runner, so they never read runState directly.
type runView struct {
	ID        string
	StartedAt time.Time
	Lines     []string
	// Batch is the plan awaiting an answer, or the one that was approved.
	Batch *verifylive.Batch
	// Awaiting reports that the page must show the nonce form.
	Awaiting   bool
	Approved   bool
	ApprovedAt time.Time
	Notice     string
	Summary    verifylive.Summary
	Err        string
	Done       bool
	Steps      int
	// Redo names the steps this run was started to re-measure, empty otherwise.
	Redo []verifylive.StepID
	// Market is the market this run sends orders in.
	Market string
}

// Remeasuring reports the re-measurement mode, for the page.
func (v runView) Remeasuring() bool { return len(v.Redo) > 0 }

// RedoList is the set as one readable line.
func (v runView) RedoList() string {
	parts := make([]string, 0, len(v.Redo))
	for _, id := range v.Redo {
		parts = append(parts, string(id))
	}
	return strings.Join(parts, ", ")
}

// Active reports a run that is neither finished nor waiting for a person.
func (v runView) Active() bool { return !v.Done }

// AwaitingRestart reports the conditional-persistence halt: the property cannot
// be observed by the process that created it, so the next step needs a new one.
func (v runView) AwaitingRestart() bool {
	for _, o := range v.Summary.Outcomes {
		if o.Verdict == verifylive.VerdictAwaitingRestart {
			return true
		}
	}
	return false
}

func (r *runState) snapshot() runView {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := runView{
		ID:         r.id,
		StartedAt:  r.startedAt,
		Lines:      append([]string(nil), r.lines...),
		Batch:      r.pending,
		Awaiting:   r.pending != nil && r.decided != nil,
		Approved:   r.approved,
		ApprovedAt: r.approvedAt,
		Notice:     r.notice,
		Summary:    r.summary,
		Done:       r.finishedFlag,
		Steps:      r.steps,
		Redo:       append([]verifylive.StepID(nil), r.redo...),
		Market:     r.market,
	}
	if r.partial != "" {
		v.Lines = append(v.Lines, r.partial)
	}
	if r.err != nil {
		v.Err = r.err.Error()
	}
	return v
}

func (r *runState) finished() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.finishedFlag
}

func (r *runState) cancel() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}

// Write receives the runner's operator progress and turns it into lines.
func (r *runState) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	text := r.partial + string(p)
	parts := strings.Split(text, "\n")
	r.partial = parts[len(parts)-1]
	for _, line := range parts[:len(parts)-1] {
		r.lines = append(r.lines, strings.TrimRight(line, "\r"))
	}
	if excess := len(r.lines) - maxLogLines; excess > 0 {
		r.lines = append([]string(nil), r.lines[excess:]...)
	}
	return len(p), nil
}

// confirmer binds the web approval to one run.
func (r *runState) confirmer(ctx context.Context) verifylive.BatchConfirmer {
	return func(b verifylive.Batch) error { return r.confirmBatch(ctx, b) }
}

// confirmBatch parks the plan and waits for a person.
func (r *runState) confirmBatch(ctx context.Context, b verifylive.Batch) error {
	decided := make(chan error, 1)

	r.mu.Lock()
	r.pending = &b
	r.decided = decided
	r.notice = ""
	r.mu.Unlock()

	err := r.await(ctx, b, decided)

	r.mu.Lock()
	r.decided = nil
	if err == nil {
		r.approved, r.approvedAt = true, r.clock()
	} else {
		r.notice = err.Error()
	}
	r.mu.Unlock()
	return err
}

// await is the wait itself: an answer, the window closing, or the console being
// shut down. All three end with nothing sent.
func (r *runState) await(ctx context.Context, b verifylive.Batch, decided <-chan error) error {
	remaining := time.Until(b.ExpiresAt)
	if remaining <= 0 {
		return verifylive.ErrConfirmationExpired
	}
	timer := time.NewTimer(remaining)
	defer timer.Stop()

	select {
	case err := <-decided:
		return err
	case <-timer.C:
		return verifylive.ErrConfirmationExpired
	case <-ctx.Done():
		return verifylive.ErrRefused
	}
}

// deliver hands an answer to the waiting runner.
//
// It reports whether anybody was waiting: a second submission of the same form,
// or one that arrives after the window closed, must not be mistaken for an
// approval of the next plan.
func (r *runState) deliver(answer error) bool {
	r.mu.Lock()
	ch := r.decided
	r.decided = nil
	if ch != nil && answer != nil {
		r.notice = answer.Error()
	}
	r.mu.Unlock()
	if ch == nil {
		return false
	}
	ch <- answer
	return true
}

func (r *runState) note(text string) {
	r.mu.Lock()
	r.notice = text
	r.mu.Unlock()
}

func (r *runState) finish(summary verifylive.Summary, err error, steps int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.partial != "" {
		r.lines = append(r.lines, r.partial)
		r.partial = ""
	}
	r.summary, r.err, r.steps, r.finishedFlag = summary, err, steps, true
}

func (r *runState) clock() time.Time {
	if r.now == nil {
		return time.Now().UTC()
	}
	return r.now()
}

// --- the console's side ----------------------------------------------------------

// currentRun returns the run this console is showing, if any.
func (c *Console) currentRun() *runState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.run
}

// startRun begins one verification.
//
// redo is empty for an ordinary run and otherwise carries the re-measurement set
// the caller read off the record. It reaches the runner as verifylive's Redo and
// nothing more: the approval, the plan and every rail behind them are the same
// ones an ordinary run goes through.
func (c *Console) startRun(market string, redo []verifylive.StepID) (*runState, error) {
	c.mu.Lock()
	if c.spent {
		c.mu.Unlock()
		return nil, errProcessSpent
	}
	if c.run != nil && !c.run.finished() {
		run := c.run
		c.mu.Unlock()
		return run, errRunInProgress
	}

	ctx, cancel := context.WithCancel(context.Background())
	run := &runState{
		id:        newToken(6),
		startedAt: c.now(),
		cancelFn:  cancel,
		done:      make(chan struct{}),
		now:       c.now,
		market:    verifylive.NormalizeMarket(market),
		redo:      append([]verifylive.StepID(nil), redo...),
	}
	c.run = run
	c.mu.Unlock()

	go c.drive(ctx, run)
	return run, nil
}

// drive runs the verification and records what it cost this process.
func (c *Console) drive(ctx context.Context, run *runState) {
	defer close(run.done)
	defer run.cancel()

	summary, entries, err := c.opts.StartVerify(ctx, run.confirmer(ctx), run, run.market, run.redo)
	steps := verifylive.StepCount(entries)
	run.finish(summary, err, steps)

	if steps > 0 {
		// The catalogue was walked. Whatever the outcome, the next attempt needs a
		// new process — see errProcessSpent.
		c.mu.Lock()
		c.spent = true
		c.mu.Unlock()
	}
	fmt.Fprintf(c.out, "verification %s finished — %d step(s) recorded%s\n",
		run.id, steps, errSuffix(err))
}

func errSuffix(err error) string {
	if err == nil {
		return ""
	}
	return ", " + err.Error()
}
