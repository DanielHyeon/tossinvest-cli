package verifylive

// runner.go walks the step catalogue.
//
// The loop is deliberately dull. Interesting control flow in a driver that places
// live orders is a liability, so the whole policy is four questions asked of
// every step in the same order — has it already been settled, is it opted in, did
// its dependency pass, does the account have what it needs — and then one call
// into steps.go. Everything a step learns comes back on a stepRun, which becomes
// exactly one line of the record whatever the outcome.
//
// # Abort, refuse, resume
//
// Three different things, kept apart:
//
//	refuse    a person declined one confirmation. The step records
//	          VerdictRefused and the run continues with the steps that do not
//	          depend on it. Refusing the idempotency check should not cost you the
//	          conditional-order measurements.
//	abort     the context was cancelled (Ctrl-C), or a confirmation was asked for
//	          with no terminal to type it at. Everything recorded so far stands,
//	          and Run returns.
//	resume    the conditional-persistence step cannot pass inside the process that
//	          registered the conditional, so the run halts there and the operator
//	          starts a new one. The record is how the new process knows what the
//	          old one did.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Options configures a run.
type Options struct {
	// Broker is the live account. Required.
	Broker Broker
	// Recorder receives one entry per step. Required.
	Recorder *Recorder
	// Confirm gates every mutation. Required — a nil one is refused at
	// construction rather than defaulted to something permissive.
	Confirm Confirmer
	// Out receives operator progress.
	Out io.Writer
	// Now is the clock, injectable for the tests.
	Now func() time.Time
	// Sleep waits, cancellably. Only the validity-window step uses it.
	Sleep func(ctx context.Context, d time.Duration) error

	// AccountRef is the broker's account number. It is masked before it reaches
	// the record or a prompt.
	AccountRef string
	// Symbol is what the buy-side probes are placed against.
	Symbol string
	// HoldingSymbol is the held symbol the sell-side and conditional steps use.
	// Empty means the account had nothing to use and those steps are skipped.
	HoldingSymbol string
	// Offset is how far from the last trade orders are priced.
	Offset float64
	// MaxSellQuantity bounds the whole-holding sell probe.
	MaxSellQuantity float64
	// IncludeTTLEdge opts into the step that deliberately creates a second order.
	IncludeTTLEdge bool
	// TTLWait is how long the validity-window step waits before replaying. The
	// document says the key lives ten minutes; the default overshoots it.
	TTLWait time.Duration
	// Redo lists steps to run again even though the record has a verdict.
	Redo []StepID

	// Prior is everything already on the record.
	Prior []Entry
	// Process identifies this process. Zero value mints a fresh one.
	Process Process
}

// DefaultTTLWait overshoots the documented ten-minute idempotency window
// ("멱등성 키는 10분간 유효하며", OrderCreateRequest.clientOrderId,
// docs/migration/openapi.latest.json) by a minute, so an expiry that happens
// exactly on the boundary is observed as an expiry rather than as a coin toss.
const DefaultTTLWait = 11 * time.Minute

// DefaultMaxSellQuantity bounds the whole-holding sell probe.
//
// One share. A "sell the entire holding" boundary is meaningful on a one-share
// holding and is a resting order for somebody's whole position on a larger one,
// so past this the boundary is recorded as unverified instead of being placed.
const DefaultMaxSellQuantity = 1.0

// Runner executes the procedure.
type Runner struct {
	broker   Broker
	recorder *Recorder
	confirm  Confirmer
	out      io.Writer
	now      func() time.Time
	sleep    func(ctx context.Context, d time.Duration) error

	accountRef      string
	symbol          string
	holdingSymbol   string
	offset          float64
	maxSellQuantity float64
	includeTTLEdge  bool
	ttlWait         time.Duration
	redo            map[StepID]bool

	runID   string
	process Process
	prior   []Entry
	written []Entry
}

// New builds a runner.
func New(o Options) (*Runner, error) {
	if o.Broker == nil {
		return nil, errors.New("verifylive: a broker is required")
	}
	if o.Recorder == nil {
		return nil, errors.New("verifylive: a recorder is required — the evidence is the point of the exercise")
	}
	if o.Confirm == nil {
		return nil, errors.New("verifylive: a confirmer is required; there is no unconfirmed mode")
	}
	if strings.TrimSpace(o.AccountRef) == "" {
		return nil, errors.New("verifylive: the account could not be identified, so nothing can be attested about it")
	}
	offset, err := ValidateOffset(o.Offset)
	if err != nil {
		return nil, err
	}
	r := &Runner{
		broker:          o.Broker,
		recorder:        o.Recorder,
		confirm:         o.Confirm,
		out:             o.Out,
		now:             o.Now,
		sleep:           o.Sleep,
		accountRef:      strings.TrimSpace(o.AccountRef),
		symbol:          strings.TrimSpace(o.Symbol),
		holdingSymbol:   strings.TrimSpace(o.HoldingSymbol),
		offset:          offset,
		maxSellQuantity: o.MaxSellQuantity,
		includeTTLEdge:  o.IncludeTTLEdge,
		ttlWait:         o.TTLWait,
		redo:            map[StepID]bool{},
		prior:           append([]Entry(nil), o.Prior...),
		process:         o.Process,
	}
	if r.out == nil {
		r.out = io.Discard
	}
	if r.now == nil {
		r.now = func() time.Time { return time.Now().UTC() }
	}
	if r.sleep == nil {
		r.sleep = sleepFor
	}
	if r.maxSellQuantity <= 0 {
		r.maxSellQuantity = DefaultMaxSellQuantity
	}
	if r.ttlWait <= 0 {
		r.ttlWait = DefaultTTLWait
	}
	if r.process.InstanceID == "" {
		r.process = NewProcess(r.now())
	}
	for _, id := range o.Redo {
		r.redo[id] = true
	}
	r.runID = newToken("run")
	return r, nil
}

// RunID identifies this invocation in the record.
func (r *Runner) RunID() string { return r.runID }

// Outcome is one step's result, for the caller's summary.
type Outcome struct {
	Step    StepID  `json:"step"`
	Title   string  `json:"title"`
	Verdict Verdict `json:"verdict"`
	Reason  string  `json:"reason,omitempty"`
	// AlreadySettled reports a step this run did not touch because the record
	// already had a verdict for it.
	AlreadySettled bool `json:"already_settled,omitempty"`
}

// Summary is what a run reports back.
type Summary struct {
	RunID    string    `json:"run_id"`
	Outcomes []Outcome `json:"outcomes"`
	// Halted reports that the run stopped before the end of the catalogue, and
	// Halt says why and what to do next.
	Halted bool   `json:"halted"`
	Halt   string `json:"halt,omitempty"`
	// Outstanding are the live objects the account is left holding.
	Outstanding []Artifact `json:"outstanding,omitempty"`
}

// Entries returns the entries this run appended.
func (r *Runner) Entries() []Entry { return append([]Entry(nil), r.written...) }

// Run walks the catalogue.
//
// It returns an error only when the account is left in a state a person has to
// deal with — an outstanding order that could not be cancelled — or when the
// context was cancelled. A step that failed is a measurement result, not a
// program error, and it comes back in the summary.
func (r *Runner) Run(ctx context.Context) (Summary, error) {
	summary := Summary{RunID: r.runID}
	fmt.Fprintf(r.out, "live verification — run %s, process %d\n", r.runID, r.process.PID)
	fmt.Fprintf(r.out, "  account %s, buy-side symbol %s, holding symbol %s\n",
		maskedAccount(r.accountRef), orNone(r.symbol), orNone(r.holdingSymbol))
	fmt.Fprintf(r.out, "  every live mutation is confirmed by typing an expiring string; nothing here has a --yes\n\n")

	for _, step := range Steps() {
		if err := ctx.Err(); err != nil {
			summary.Halted = true
			summary.Halt = "interrupted; everything recorded so far stands"
			summary.Outstanding = r.outstanding()
			return summary, err
		}

		if settled, verdict := r.settled(step.ID); settled {
			summary.Outcomes = append(summary.Outcomes, Outcome{
				Step: step.ID, Title: step.Title, Verdict: verdict, AlreadySettled: true,
			})
			fmt.Fprintf(r.out, "· %-22s already %s on the record\n", step.ID, verdict)
			continue
		}

		sr := &stepRun{step: step, startedAt: r.now()}
		fmt.Fprintf(r.out, "▸ %-22s %s\n", step.ID, step.Title)

		if reason, skip := r.preflight(step); skip {
			sr.skip(reason)
		} else {
			r.dispatch(ctx, step.ID, sr)
		}

		entry := r.entryFor(sr)
		if err := r.recorder.Append(entry); err != nil {
			return summary, err
		}
		r.written = append(r.written, entry)
		summary.Outcomes = append(summary.Outcomes, Outcome{
			Step: step.ID, Title: step.Title, Verdict: sr.verdict, Reason: sr.reason,
		})
		fmt.Fprintf(r.out, "  %s%s\n\n", sr.verdict, reasonSuffix(sr.reason))

		if sr.verdict == VerdictAwaitingRestart {
			summary.Halted = true
			summary.Halt = sr.reason
			summary.Outstanding = r.outstanding()
			return summary, nil
		}
		if errors.Is(sr.abort, ErrNotATerminal) {
			summary.Halted = true
			summary.Halt = ErrNotATerminal.Error()
			summary.Outstanding = r.outstanding()
			return summary, ErrNotATerminal
		}
		if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {
			summary.Halted = true
			summary.Halt = "interrupted; everything recorded so far stands"
			summary.Outstanding = r.outstanding()
			return summary, sr.abort
		}
	}

	summary.Outstanding = r.outstanding()
	if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {
		return summary, fmt.Errorf(
			"verify: the run finished but %d live object(s) this tool created are still on the account: %s — "+
				"cancel them before doing anything else", len(leftovers), describeArtifacts(leftovers))
	}
	return summary, nil
}

// settled reports whether the record already has a terminal verdict for a step.
//
// A fresh invocation that finds one does not run the step again: re-running a
// mutating step would place a second live order for a measurement that has
// already been made. --redo is the deliberate override.
func (r *Runner) settled(id StepID) (bool, Verdict) {
	if r.redo[id] {
		return false, ""
	}
	e, ok := LastEntry(r.prior, id)
	if !ok || !e.Verdict.Terminal() {
		return false, ""
	}
	return true, e.Verdict
}

// preflight applies the catalogue's own gates: opt-in, deferral, dependencies and
// the account's holdings. It returns the reason a step is skipped.
func (r *Runner) preflight(step Step) (string, bool) {
	if step.Deferred != "" {
		// Never skipped. A deferred step still runs, and its body records the
		// deferral as an explicit "unverified" — which is the whole reason task
		// 2.6 can produce a list of what automatic entry is not allowed on.
		return "", false
	}
	if step.OptIn != "" && !r.optedIn(step) {
		return fmt.Sprintf("not requested — pass %s to run it. %s", step.OptIn, step.Procedure[0]), true
	}
	for _, dep := range step.DependsOn {
		if !r.passed(dep) {
			return fmt.Sprintf("%s did not pass, so there is nothing for this step to observe", dep), true
		}
	}
	if step.NeedsHolding && r.holdingSymbol == "" {
		return "the account holds nothing this tool can use. This step needs an existing position — the tool " +
			"never buys one to create it. Hold at least one share of a KR symbol and re-run with --resume " +
			"(or --holding-symbol to name it)", true
	}
	if step.Mutates && MarketOf(r.mutationSymbol(step)) != MarketKR {
		return fmt.Sprintf("%s is not a KR symbol; the amend and price-band rules this step relies on are KR's, "+
			"so the US path is left unverified rather than probed with the wrong rules", r.mutationSymbol(step)), true
	}
	return "", false
}

func (r *Runner) mutationSymbol(step Step) string {
	if step.NeedsHolding {
		return r.holdingSymbol
	}
	return r.symbol
}

func (r *Runner) optedIn(step Step) bool {
	return step.ID == StepIdempotencyTTLEdge && r.includeTTLEdge
}

// passed looks in the record and in this run.
func (r *Runner) passed(id StepID) bool {
	if Passed(r.written, id) {
		return true
	}
	if _, ok := LastEntry(r.written, id); ok {
		return false
	}
	return Passed(r.prior, id)
}

func (r *Runner) allEntries() []Entry {
	out := append([]Entry(nil), r.prior...)
	return append(out, r.written...)
}

func (r *Runner) outstanding() []Artifact { return Outstanding(r.allEntries()) }

func (r *Runner) entryFor(sr *stepRun) Entry {
	return Entry{
		FormatVersion: RecordFormatVersion,
		Kind:          KindStep,
		RunID:         r.runID,
		Process:       r.process,
		StepID:        sr.step.ID,
		Title:         sr.step.Title,
		Tasks:         sr.step.Tasks,
		StartedAt:     sr.startedAt.UTC(),
		FinishedAt:    r.now().UTC(),
		AccountRef:    maskedAccount(r.accountRef),
		Mutating:      sr.step.Mutates,
		Verdict:       sr.verdict,
		Reason:        sr.reason,
		Calls:         sr.calls,
		Observations:  sr.observations,
		Artifacts:     sr.artifacts,
	}
}

// --- the step accumulator -----------------------------------------------------

// stepRun collects what one step learned, and becomes one line of the record.
type stepRun struct {
	step      Step
	startedAt time.Time

	calls        []Call
	observations []Observation
	artifacts    []Artifact

	verdict Verdict
	reason  string
	// abort is set when the failure is one the run cannot continue past.
	abort error
}

func (sr *stepRun) observe(key, value, detail string) {
	sr.observations = append(sr.observations, Observation{Key: key, Value: value, Detail: detail})
}

func (sr *stepRun) pass()              { sr.verdict = VerdictPass }
func (sr *stepRun) skip(reason string) { sr.verdict, sr.reason = VerdictSkipped, reason }
func (sr *stepRun) deferStep(reason string) {
	sr.verdict, sr.reason = VerdictDeferred, reason
}
func (sr *stepRun) refuse(reason string) { sr.verdict, sr.reason = VerdictRefused, reason }
func (sr *stepRun) awaitRestart(reason string) {
	sr.verdict, sr.reason = VerdictAwaitingRestart, reason
}

// fail records a measured negative. It is not a program error: "the broker did
// not behave the way the document says" is a result, and the run keeps going so
// the remaining independent properties still get measured.
func (sr *stepRun) fail(format string, a ...any) {
	sr.verdict, sr.reason = VerdictFail, fmt.Sprintf(format, a...)
}

// resolve turns an error from a step body into a verdict.
//
// A refusal is its own verdict and never a failure: the operator made a decision
// and the record says so. A missing terminal aborts the run, because every
// remaining mutating step would ask the same impossible question.
func (sr *stepRun) resolve(err error) {
	switch {
	case err == nil:
		if sr.verdict == "" {
			sr.pass()
		}
	case errors.Is(err, ErrRefused), errors.Is(err, ErrConfirmationExpired):
		sr.refuse(err.Error())
	case errors.Is(err, ErrNotATerminal):
		sr.refuse(err.Error())
		sr.abort = err
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		sr.fail("interrupted: %v", err)
		sr.abort = err
	default:
		sr.fail("%s", truncateError(err))
	}
}

func (sr *stepRun) logCall(endpoint string, req any, started, finished time.Time, resp any, err error) {
	c := Call{
		Endpoint:      endpoint,
		At:            started.UTC(),
		LatencyMS:     finished.Sub(started).Milliseconds(),
		RequestDigest: Digest(req),
	}
	if err != nil {
		c.Error = truncateError(err)
		c.ErrorCode = brokerErrorCode(err)
	} else {
		c.ResponseDigest = Digest(resp)
	}
	sr.calls = append(sr.calls, c)
}

func (sr *stepRun) created(kind, id, symbol string, at time.Time, note string) {
	sr.artifacts = append(sr.artifacts, Artifact{
		Kind: kind, ID: id, Symbol: symbol, CreatedAt: at.UTC(), Note: note,
	})
}

func (sr *stepRun) cancelled(kind, id, symbol string, at time.Time, note string) {
	sr.artifacts = append(sr.artifacts, Artifact{
		Kind: kind, ID: id, Symbol: symbol, CancelledAt: at.UTC(), Cancelled: true, Note: note,
	})
}

// markDeliberate flags the artifact that is *meant* to outlive the process, so
// the end-of-run check does not report the conditional-persistence design as a
// leak.
func (sr *stepRun) markDeliberate(kind, id, note string) {
	for i := range sr.artifacts {
		if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].Cancelled {
			sr.artifacts[i].Deliberate = true
			sr.artifacts[i].Note = note
		}
	}
}

// latencyStats summarises the round trips a step measured, which is the input to
// any safety margin on an idempotent replay.
func (sr *stepRun) latencyStats(endpoint string) (maxMS, count int64) {
	for _, c := range sr.calls {
		if c.Endpoint != endpoint {
			continue
		}
		count++
		if c.LatencyMS > maxMS {
			maxMS = c.LatencyMS
		}
	}
	return maxMS, count
}

// --- helpers ------------------------------------------------------------------

func undeliberate(as []Artifact) []Artifact {
	var out []Artifact
	for _, a := range as {
		if !a.Deliberate {
			out = append(out, a)
		}
	}
	return out
}

func describeArtifacts(as []Artifact) string {
	parts := make([]string, 0, len(as))
	for _, a := range as {
		parts = append(parts, fmt.Sprintf("%s %s (%s)", a.Kind, a.ID, a.Symbol))
	}
	return strings.Join(parts, ", ")
}

func reasonSuffix(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return ""
	}
	return " — " + reason
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}
