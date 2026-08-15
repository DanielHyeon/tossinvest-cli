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
// # The approval comes first
//
// Nothing is dispatched until the run-wide batch has been approved (approveBatch).
// That is the whole of the default model: the plan is built from read-only calls,
// shown in full, and answered once. A run that is not approved does not walk the
// catalogue at all — not even its read-only steps — because "anything else aborts
// the run" is what the operator was told when they were shown the list.
//
// # Abort, refuse, unapproved, resume
//
// Four different things, kept apart:
//
//	refuse       a person declined one confirmation under --confirm-each. The step
//	             records VerdictRefused and the run continues with the steps that
//	             do not depend on it. Refusing the idempotency check should not
//	             cost you the conditional-order measurements.
//	abort        the context was cancelled (Ctrl-C), or a confirmation was asked
//	             for with no terminal to type it at. Everything recorded so far
//	             stands, and Run returns.
//	unapproved   a step was about to send something the approved list does not
//	             carry. Nothing was sent, and the whole run stops: the alternative
//	             is adapting to conditions the operator never agreed to. A step
//	             with no approved lines at all is skipped instead, before it
//	             starts.
//	resume       the conditional-persistence step cannot pass inside the process
//	             that registered the conditional, so the run halts there and the
//	             operator starts a new one — which approves its own remaining
//	             batch. The record is how the new process knows what the old one
//	             did.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// Options configures a run.
type Options struct {
	// Broker is the live account. Required.
	Broker Broker
	// Recorder receives one entry per step. Required.
	Recorder *Recorder
	// Confirm gates one mutation at a time. It is used when ConfirmEach is set,
	// and it is required either way — a nil one is refused at construction rather
	// than defaulted to something permissive.
	Confirm Confirmer
	// ConfirmBatch asks for the run-wide approval, once, before anything is sent.
	// Required, for the same reason.
	ConfirmBatch BatchConfirmer
	// ConfirmEach opts out of the batch approval and back into a typed
	// confirmation immediately before every single mutation.
	ConfirmEach bool
	// Resume says this invocation continues the owner-only verification record.
	// M0 trigger evidence is intentionally valid only on a resumed run.
	Resume bool
	// Receipt is the durable, sanitized M0 causal evidence stream. Nil keeps a
	// caller out of the trigger-only measurement mode.
	Receipt *CausalReceipt
	// Market is the market this run sends orders in. The zero value is KR, so a
	// caller that says nothing behaves exactly as this tool did before it knew
	// about US (§0.2). A symbol from another market is skipped with a reason.
	Market string
	// ApprovalChannel is how the person answered, for the evidence record. The
	// zero value is the terminal's typed string, so a caller that does not set it
	// records what it always recorded (§0.2); the console sets
	// ApprovalChannelConsoleClick.
	ApprovalChannel string
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
	// IncludeTrigger opts into the trigger observation, which is the only step
	// that registers something the market is meant to reach. A successful run
	// sells one share and cannot be undone.
	IncludeTrigger bool
	// TriggerWindow is how long the trigger observation waits for the market to
	// come to its trigger before cancelling it and recording the attempt as
	// inconclusive. Zero is DefaultTriggerWindow.
	TriggerWindow time.Duration
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
	broker       Broker
	recorder     *Recorder
	confirm      Confirmer
	confirmBatch BatchConfirmer
	confirmEach  bool
	// market is the market this run may send orders in.
	market string
	// approvalChannel is what the record says about how the approval arrived.
	approvalChannel string
	out             io.Writer
	now             func() time.Time
	sleep           func(ctx context.Context, d time.Duration) error

	// plan is the batch the operator approved. Nil means nothing is approved, and
	// mutate.go's authorise treats that as "send nothing" rather than as
	// "unrestricted" — the failure direction has to be the safe one.
	plan *Plan

	accountRef       string
	symbol           string
	holdingSymbol    string
	offset           float64
	maxSellQuantity  float64
	includeTTLEdge   bool
	includeTrigger   bool
	m0Receipt        *CausalReceipt
	m0ReceiptLease   *CausalReceiptLease
	m0ReadSource     *official.M0ReadSource
	m0CriticalWindow bool
	m0Gap            bool
	m0ReceiptErr     error
	m0PendingClient  string
	m0Mu             sync.Mutex
	m0Attempts       map[string][]m0AttemptStamp
	m0ParentBarrier  M0ReceiptStamp
	m0ChildFill      M0ReceiptStamp
	// Private killpoints prove durable ordering; production leaves both nil.
	m0BeforeConditionalCreate func() error
	m0AfterConditionalCreate  func() error
	m0AfterChildCheckpoint    func() error
	m0AfterParentCausal       func() error
	m0AppendCheckpoint        func(Entry) error
	ttlWait                   time.Duration
	triggerWindow             time.Duration
	redo                      map[StepID]bool

	// readBackoffs counts how many times a read-only call has waited out a 429 in
	// this process. The trigger observation snapshots it around each poll: inside
	// the observation window the error bound on a timestamp is the backoff, not
	// the polling interval, and a record that did not say so would overstate its
	// own precision (measurements.md M4, M8, M10).
	readBackoffs int

	runID   string
	process Process
	prior   []Entry
	written []Entry
}

type m0AttemptStamp struct {
	Receipt        M0ReceiptStamp
	RequestStartNS int64
	StatusCode     int
	Err            error
}

func (r *Runner) m0ReadContext(ctx context.Context, phase string) context.Context {
	if !r.m0ReceiptUsable() {
		return ctx
	}
	return official.WithAttemptObserver(ctx, func(a official.AttemptTrace) {
		if r.m0ReceiptLease == nil {
			r.m0Mu.Lock()
			r.m0ReceiptErr = errors.New("verify: M0 causal receipt run lease is unavailable")
			r.m0Gap = true
			r.m0Mu.Unlock()
			return
		}
		stamp, err := r.m0ReceiptLease.RecordAttempt(phase, a)
		r.m0Mu.Lock()
		defer r.m0Mu.Unlock()
		if err != nil {
			r.m0ReceiptErr = err
			r.m0Gap = true
			return
		}
		r.m0Attempts[phase] = append(r.m0Attempts[phase], m0AttemptStamp{
			Receipt: stamp, RequestStartNS: a.RequestStart.Sub(r.m0Receipt.start).Nanoseconds(), StatusCode: a.StatusCode, Err: a.Err,
		})
		if r.m0CriticalWindow && (a.Err != nil || a.StatusCode == 401 || a.StatusCode == 429) {
			r.m0Gap = true
		}
	})
}

func (r *Runner) m0ReceiptUsable() bool {
	return r.m0Receipt != nil && r.m0Receipt.usable()
}

func (r *Runner) recordM0Attempts(phase string, attempts []official.AttemptTrace) {
	for _, attempt := range attempts {
		if r.m0ReceiptLease == nil {
			r.m0ReceiptErr = errors.New("verify: M0 causal receipt run lease is unavailable")
			r.m0Gap = true
			continue
		}
		stamp, err := r.m0ReceiptLease.RecordAttempt(phase, attempt)
		r.m0Mu.Lock()
		if err != nil {
			r.m0ReceiptErr = err
			r.m0Gap = true
		} else {
			r.m0Attempts[phase] = append(r.m0Attempts[phase], m0AttemptStamp{Receipt: stamp, RequestStartNS: attempt.RequestStart.Sub(r.m0Receipt.start).Nanoseconds(), StatusCode: attempt.StatusCode, Err: attempt.Err})
			if r.m0CriticalWindow && (attempt.Err != nil || attempt.StatusCode == 401 || attempt.StatusCode == 429) {
				r.m0Gap = true
			}
		}
		r.m0Mu.Unlock()
	}
}

func (r *Runner) appendM0Checkpoint(checkpoint M0Checkpoint) error {
	if r.m0Receipt == nil {
		return nil
	}
	if !r.m0ReceiptUsable() {
		return errors.New("verify: M0 causal receipt is closed or unusable")
	}
	now := r.now().UTC()
	e := Entry{
		FormatVersion: RecordFormatVersion,
		Kind:          KindM0Checkpoint,
		RunID:         r.runID,
		Process:       r.process,
		StartedAt:     now,
		FinishedAt:    now,
		AccountRef:    maskedAccount(r.accountRef),
		M0Checkpoint:  &checkpoint,
	}
	appendCheckpoint := r.m0AppendCheckpoint
	if appendCheckpoint == nil {
		appendCheckpoint = r.recorder.Append
	}
	if err := appendCheckpoint(e); err != nil {
		return err
	}
	r.written = append(r.written, e)
	return nil
}

func (r *Runner) m0ChildAttemptAfterParent() bool {
	r.m0Mu.Lock()
	defer r.m0Mu.Unlock()
	if r.m0ParentBarrier.RunID != r.runID || r.m0ParentBarrier.Sequence == 0 {
		return false
	}
	for _, attempt := range r.m0Attempts["child-order-read"] {
		if attempt.Receipt.RunID == r.runID && attempt.RequestStartNS > r.m0ParentBarrier.FsyncDoneElapsedNS {
			return true
		}
	}
	return false
}

func (r *Runner) m0PassReady() bool {
	if !r.m0ReceiptUsable() {
		return false
	}
	if r.m0Gap || r.m0CriticalWindow || r.m0ParentBarrier.RunID != r.runID || r.m0ChildFill.RunID != r.runID {
		return false
	}
	if r.m0ParentBarrier.Sequence == 0 || r.m0ChildFill.Sequence == 0 || r.m0ParentBarrier.Sequence >= r.m0ChildFill.Sequence {
		return false
	}
	return r.m0ChildAttemptAfterParent()
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
	if o.ConfirmBatch == nil {
		return nil, errors.New("verifylive: a batch confirmer is required; the run-wide approval is the default " +
			"gate and there is no mode that skips it")
	}
	if strings.TrimSpace(o.AccountRef) == "" {
		return nil, errors.New("verifylive: the account could not be identified, so nothing can be attested about it")
	}
	offset, err := ValidateOffset(o.Offset)
	if err != nil {
		return nil, err
	}
	if o.IncludeTrigger {
		if !o.ConfirmEach || !o.Resume || o.IncludeTTLEdge || o.Receipt == nil || !o.Receipt.usable() ||
			len(o.Redo) != 1 || o.Redo[0] != StepConditionalTrigger {
			return nil, errors.New("verifylive: --include-trigger requires M0 receipt, --confirm-each, --resume, and --redo conditional-trigger only")
		}
		if len(Outstanding(o.Prior)) != 0 {
			return nil, errors.New("verifylive: M0 trigger receipt holds before cleanup because the prior record has outstanding artifacts")
		}
		if checkpoint, ok, checkpointErr := M0Unsettled(o.Prior); checkpointErr != nil {
			return nil, checkpointErr
		} else if ok && checkpoint.Kind != "pending-create" {
			return nil, fmt.Errorf("verifylive: M0 causal receipt HOLD: unsettled %s owner requires manual reconciliation", checkpoint.Kind)
		} else if !ok {
			// A pending-create checkpoint is the one response-lost state that may
			// enter this process: Run performs its read-only all-page recovery and
			// always terminates in HOLD.  It must happen before prerequisite checks
			// so a crash can be reconciled even if the older record predates M0.
			if err := M0ExactPrerequisites(o.Prior); err != nil {
				return nil, err
			}
		}
		if _, ok := official.M0ReadSourceFor(o.Broker); !ok {
			return nil, errors.New("verifylive: M0 causal receipt requires an official-client raw read source")
		}
	}
	r := &Runner{
		broker:          o.Broker,
		recorder:        o.Recorder,
		confirm:         o.Confirm,
		confirmBatch:    o.ConfirmBatch,
		confirmEach:     o.ConfirmEach,
		market:          NormalizeMarket(o.Market),
		approvalChannel: strings.TrimSpace(o.ApprovalChannel),
		out:             o.Out,
		now:             o.Now,
		sleep:           o.Sleep,
		accountRef:      strings.TrimSpace(o.AccountRef),
		symbol:          strings.TrimSpace(o.Symbol),
		holdingSymbol:   strings.TrimSpace(o.HoldingSymbol),
		offset:          offset,
		maxSellQuantity: o.MaxSellQuantity,
		includeTTLEdge:  o.IncludeTTLEdge,
		includeTrigger:  o.IncludeTrigger,
		m0Receipt:       o.Receipt,
		m0Attempts:      make(map[string][]m0AttemptStamp),
		ttlWait:         o.TTLWait,
		triggerWindow:   o.TriggerWindow,
		redo:            map[StepID]bool{},
		prior:           append([]Entry(nil), o.Prior...),
		process:         o.Process,
	}
	if r.approvalChannel == "" {
		r.approvalChannel = ApprovalChannelTyped
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
	if r.triggerWindow <= 0 {
		r.triggerWindow = DefaultTriggerWindow
	}
	if r.process.InstanceID == "" {
		r.process = NewProcess(r.now())
	}
	for _, id := range o.Redo {
		r.redo[id] = true
	}
	r.runID = newToken("run")
	if r.m0Receipt != nil {
		r.runID = r.m0Receipt.RunID()
		r.m0ReadSource, _ = official.M0ReadSourceFor(o.Broker)
	}
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
	if r.includeTrigger {
		lease, err := r.m0Receipt.AcquireRunLease()
		if err != nil {
			summary.Halted, summary.Halt = true, truncateError(err)
			return summary, fmt.Errorf("verify: M0 causal receipt HOLD: %w", err)
		}
		r.m0ReceiptLease = lease
		defer lease.Release()
	}
	r.writeBanner()
	if r.includeTrigger && !r.m0ReceiptUsable() {
		err := errors.New("verify: M0 causal receipt HOLD: receipt was closed before the trigger run")
		summary.Halted, summary.Halt, summary.Outstanding = true, truncateError(err), r.outstanding()
		return summary, err
	}
	if pending, ok, pendingErr := r.m0PendingCheckpoint(); pendingErr != nil {
		summary.Halted = true
		summary.Halt = truncateError(pendingErr)
		summary.Outstanding = r.outstanding()
		return summary, pendingErr
	} else if ok {
		err := r.m0RecoverPending(ctx, pending)
		summary.Halted = true
		summary.Halt = truncateError(err)
		summary.Outstanding = r.outstanding()
		return summary, err
	}

	if halt, err := r.approveBatch(ctx); err != nil || halt != "" {
		summary.Halted = true
		summary.Halt = halt
		summary.Outstanding = r.outstanding()
		return summary, err
	}

	// Before the catalogue: whatever an earlier run left resting. It is on the list
	// the operator just approved, and until it is gone the exposure cap refuses
	// every mutating step below (cleanup.go).
	if outcome, err, stop := r.cleanup(ctx); outcome.Step != "" || err != nil {
		if outcome.Step != "" {
			summary.Outcomes = append(summary.Outcomes, outcome)
		}
		if stop {
			summary.Halted = true
			summary.Halt = outcome.Reason
			if outcome.Reason == "" {
				summary.Halt = "정리를 기록하지 못했다"
			}
			summary.Outstanding = r.outstanding()
			return summary, err
		}
	}

	for _, step := range Steps() {
		if err := ctx.Err(); err != nil {
			summary.Halted = true
			summary.Halt = "중단됨. 여기까지 기록된 것은 모두 유효하다"
			summary.Outstanding = r.outstanding()
			return summary, err
		}

		if settled, verdict := r.settled(step.ID); settled {
			summary.Outcomes = append(summary.Outcomes, Outcome{
				Step: step.ID, Title: step.Title, Verdict: verdict, AlreadySettled: true,
			})
			fmt.Fprintf(r.out, "· %-22s %s — 기록에 이미 판정이 있어 건너뛴다\n", step.ID, verdict)
			continue
		}

		sr := &stepRun{step: step, startedAt: r.now()}
		fmt.Fprintf(r.out, "▸ %-22s %s\n", step.ID, StepLabel(step.ID))

		if reason, skip := r.preflight(step); skip {
			sr.skip(reason)
		} else {
			r.dispatch(ctx, step.ID, sr)
			r.sweepStep(ctx, sr)
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
		if errors.Is(sr.abort, ErrOutsidePlan) {
			// The run would have had to send something the approved list does not
			// cover. Nothing was sent, and carrying on would mean the remaining
			// steps run against conditions the approval did not describe.
			summary.Halted = true
			summary.Halt = sr.reason
			summary.Outstanding = r.outstanding()
			return summary, sr.abort
		}
		if errors.Is(sr.abort, ErrM0TerminalHold) {
			summary.Halted = true
			summary.Halt = sr.reason
			summary.Outstanding = r.outstanding()
			return summary, sr.abort
		}
		if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {
			summary.Halted = true
			summary.Halt = "중단됨. 여기까지 기록된 것은 모두 유효하다"
			summary.Outstanding = r.outstanding()
			return summary, sr.abort
		}
	}

	summary.Outstanding = r.outstanding()
	if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {
		return summary, fmt.Errorf(
			"verify: 실행은 끝났지만 이 도구가 만든 객체 %d건이 아직 계좌에 살아 있다: %s — "+
				"다른 일을 하기 전에 먼저 취소하라", len(leftovers), describeArtifacts(leftovers))
	}
	return summary, nil
}

// writeBanner says which approval model this run is using before it uses it.
func (r *Runner) writeBanner() {
	fmt.Fprintf(r.out, "실계좌 검증 — run %s, process %d\n", r.runID, r.process.PID)
	fmt.Fprintf(r.out, "  계좌 %s, 매수측 종목 %s, 보유 종목 %s\n",
		maskedAccount(r.accountRef), orNone(r.symbol), orNone(r.holdingSymbol))
	if r.confirmEach {
		fmt.Fprintf(r.out, "  --confirm-each: 모든 라이브 mutation이 전송 직전에 표시되는 자기 몫의 "+
			"만료 확인 문자열을\n  기다린다. 여기에 --yes는 없다.\n\n")
		return
	}
	// The banner names the gate the operator is actually about to meet: a typed
	// string at a terminal, a click on the console's approval screen. Describing
	// the other one would be the tool telling them to do something that does not
	// work here.
	answer := "만료되는 확인 문자열 1개"
	if r.approvalChannel == ApprovalChannelConsoleClick {
		answer = "만료되는 승인 1건(화면의 클릭)"
	}
	fmt.Fprintf(r.out, "  승인: 실행 전체에 대해 %s. 이 실행이 보낼 수 있는 모든 "+
		"라이브 요청을\n  먼저 나열하고, 목록에 없는 것은 절대 전송되지 않으며, 보내야 하는 상황이 되면 "+
		"실행이 멈춘다.\n  여기에 --yes는 없다. 단계별 확인은 `--confirm-each`다.\n\n", answer)
}

// approveBatch is the run-wide gate.
//
// It returns the halt reason when the run must not continue. An empty reason and a
// nil error mean the run may proceed — either because a person approved the list,
// or because there is no list to approve and therefore nothing to send.
func (r *Runner) approveBatch(ctx context.Context) (string, error) {
	if r.confirmEach {
		return "", nil
	}

	startedAt := r.now()
	if r.plan == nil {
		// Computed once. The package's own tests preset it, which is how "the run
		// and the approved list disagree" is exercised without waiting for a broker
		// to change its mind mid-procedure; nothing outside this package can, since
		// New never sets it and there is no option that does.
		p := r.Plan(ctx)
		r.plan = &p
	}
	plan := *r.plan

	if len(plan.Mutations) == 0 {
		fmt.Fprintf(r.out, "이 실행은 라이브 mutation을 계획하지 않는다 — 승인할 것이 없다.\n")
		plan.WriteLines(r.out)
		fmt.Fprintln(r.out)
		return "", nil
	}

	batch := NewBatch(plan, StepCount(r.prior) > 0, r.now())
	err := r.confirmBatch(batch)

	verdict, reason := VerdictPass, ""
	if err != nil {
		verdict, reason = VerdictRefused, err.Error()
		r.plan = nil
	}
	if recErr := r.recordApproval(plan, batch, verdict, reason, startedAt); recErr != nil {
		return "승인을 기록할 수 없어 아무것도 전송되지 않았다", recErr
	}

	switch {
	case err == nil:
		fmt.Fprintf(r.out, "\n승인됨: 라이브 요청 %d건, 그 외에는 아무것도 나가지 않는다.\n\n", len(plan.Mutations))
		return "", nil
	case errors.Is(err, ErrNotATerminal):
		return ErrNotATerminal.Error(), ErrNotATerminal
	default:
		fmt.Fprintf(r.out, "\n%s 아무것도 전송되지 않았고 어떤 단계도 실행되지 않았다.\n", err.Error())
		fmt.Fprintf(r.out, "`tossctl verify run --list`는 계좌를 건드리지 않고 같은 절차를 출력한다.\n")
		return "배치가 승인되지 않았다. 아무것도 전송되지 않았고 어떤 단계도 실행되지 않았다", nil
	}
}

// recordApproval puts the approval — granted or declined — on the evidence record.
//
// The plan is stored as a digest and a count rather than as a copy: the record is
// read by people and pasted into threads, and the list itself is reproducible from
// the catalogue plus the flags. What has to be provable later is that a person was
// shown a list of exactly this shape and answered.
func (r *Runner) recordApproval(plan Plan, batch Batch, verdict Verdict, reason string, startedAt time.Time) error {
	return r.recorder.Append(Entry{
		FormatVersion: RecordFormatVersion,
		Kind:          KindApproval,
		RunID:         r.runID,
		Process:       r.process,
		Title:         "batch approval",
		StartedAt:     startedAt.UTC(),
		FinishedAt:    r.now().UTC(),
		AccountRef:    maskedAccount(r.accountRef),
		Verdict:       verdict,
		Reason:        reason,
		Observations: []Observation{
			{Key: "approval.model", Value: "batch", Detail: r.approvalChannel},
			{Key: "approval.requests_listed", Value: strconv.Itoa(len(plan.Mutations))},
			{Key: "approval.steps_listed", Value: joinSteps(plan.MutatingSteps())},
			{Key: "approval.plan_digest", Value: plan.Digest()},
			{Key: "approval.resumed", Value: strconv.FormatBool(batch.Resumed)},
		},
	})
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

// deferredForm reports that this run will drive the step in its deferred shape:
// it runs, and its body records an explicit "unverified" instead of measuring
// anything.
//
// It exists because one step is now both deferred *and* opted in, and the two
// gates disagree about what that means. A deferred step is never skipped — its
// body has to run so the deferral lands on the record, which is what task 2.6
// reads to decide where automatic entry stays forbidden. An opt-in step that was
// not asked for is skipped outright. Getting a step that is both wrong in either
// direction costs something real:
//
//	skip it              the unverified observations disappear from the record, and
//	                     the report silently stops saying the property is unmeasured
//	run it as mutating   the approval list carries a live request for a step the
//	                     operator did not opt into
//
// So the deferral is lifted by the opt-in rather than overridden by it, and
// mutatesNow is what the plan and the record consult instead of Step.Mutates.
func (r *Runner) deferredForm(step Step) bool {
	return step.Deferred != "" && (step.OptIn == "" || !r.optedIn(step))
}

// mutatesNow reports that the step can send a live request in this run. A step in
// its deferred form declares mutations it will not make.
func (r *Runner) mutatesNow(step Step) bool {
	return step.Mutates && !r.deferredForm(step)
}

// preflight decides whether a step runs at all.
//
// It is the catalogue's own gates plus one more: a mutating step that is not on the
// approved list is skipped here, before it can reach a mutation call. That is the
// difference between the two ways a plan can say no — a step with no approved lines
// never starts, while a step that starts and then tries to send something outside
// its lines stops the whole run (mutate.go's authorise).
func (r *Runner) preflight(step Step) (string, bool) {
	if reason, skip := r.preflightStatic(step, r.passed); skip {
		return reason, true
	}
	if r.mutatesNow(step) && !r.approvedStep(step.ID) {
		return "이 실행에서 승인된 배치에 없다 — " + r.unapprovedReason(step.ID), true
	}
	return "", false
}

// approvedStep reports whether this step may mutate at all.
func (r *Runner) approvedStep(id StepID) bool {
	if r.confirmEach {
		return true
	}
	if r.plan == nil {
		return false
	}
	return r.plan.Covers(id)
}

func (r *Runner) unapprovedReason(id StepID) string {
	if r.plan == nil {
		return "이 실행에는 승인된 배치가 없다"
	}
	return r.plan.ExclusionReason(id)
}

// preflightStatic applies the catalogue's own gates: opt-in, deferral, dependencies
// and the account's holdings. It returns the reason a step is skipped.
//
// passed is how it decides a dependency held. At run time that is the record; while
// the plan is still being built it also counts the steps this run is about to run,
// because a plan that left out conditional-modify on the grounds that
// conditional-register has not passed *yet* would exclude half the procedure.
func (r *Runner) preflightStatic(step Step, passed func(StepID) bool) (string, bool) {
	if r.deferredForm(step) {
		// Never skipped. A deferred step still runs, and its body records the
		// deferral as an explicit "unverified" — which is the whole reason task
		// 2.6 can produce a list of what automatic entry is not allowed on.
		//
		// deferredForm rather than Step.Deferred: once the opt-in that unlocks a
		// deferred step is present the deferral is lifted, and the gates below —
		// the holding, the market, the dependencies — apply as they do to any
		// other step. Returning here regardless would let an opted-in trigger
		// observation run against an account with nothing to sell.
		return "", false
	}
	if step.OptIn != "" && !r.optedIn(step) {
		return fmt.Sprintf("요청하지 않았다 — 실행하려면 %s를 붙여라. %s", step.OptIn, step.Procedure[0]), true
	}
	for _, dep := range step.DependsOn {
		if !passed(dep) {
			return fmt.Sprintf("%s가 통과하지 않아 이 단계가 관측할 대상이 없다", dep), true
		}
	}
	if step.NeedsHolding && r.holdingSymbol == "" {
		return "이 도구가 쓸 수 있는 보유 종목이 계좌에 없다. 이 단계는 기존 포지션이 필요하고, 도구는 " +
			"그것을 만들려고 매수하지 않는다. KR 종목을 최소 1주 보유한 뒤 --resume으로 다시 실행하라 " +
			"(--holding-symbol로 종목을 지정할 수도 있다)", true
	}
	if symbol := r.mutationSymbol(step); step.Mutates && !SameMarket(MarketOf(symbol), r.market) {
		return fmt.Sprintf("%s는 %s 종목인데 이 실행의 시장은 %s다. 한 실행은 자기 시장의 종목에만 주문을 "+
			"내며, 다른 시장은 그 시장으로 실행할 때 측정한다", symbol, MarketOf(symbol), r.market), true
	}
	return "", false
}

// mutationSymbol is what a step's live requests are about.
//
// It is the single answer to that question: the plan reads it to write the line an
// operator approves (plan.go), and preflight reads it to keep a run inside its own
// market. Those two have to agree with the step body, because the plan's symbol is
// compared exactly against the request's — a step resolving its target one way while
// the list said another does not send the wrong thing, it sends nothing and stops
// the run (2026-07-29, conditional-modify, twice).
//
// The conditional steps are the case that made this explicit. They act on the
// conditional order the record already carries, whose symbol is a property of the
// account rather than of this run's flags. Before it is registered — a full run that
// starts from the beginning — the answer is the holding, because that is what
// conditional-register is about to register against.
func (r *Runner) mutationSymbol(step Step) string {
	if step.ActsOnConditional {
		if _, symbol, ok := r.liveConditional(); ok && strings.TrimSpace(symbol) != "" {
			return symbol
		}
		return r.holdingSymbol
	}
	if step.NeedsHolding {
		return r.holdingSymbol
	}
	return r.symbol
}

// optedIn reports that the operator asked for a step that does not run by
// default. The default is false: a new opt-in step whose flag was never wired up
// stays off rather than on.
func (r *Runner) optedIn(step Step) bool {
	switch step.ID {
	case StepIdempotencyTTLEdge:
		return r.includeTTLEdge
	case StepConditionalTrigger:
		return r.includeTrigger
	default:
		return false
	}
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
		// The catalogue's flag as this run resolved it. A step whose deferral was
		// not lifted declares mutations it does not make, and a line claiming it
		// mutated would be a leftover this tool never created.
		Mutating:     r.mutatesNow(sr.step),
		Verdict:      sr.verdict,
		Reason:       sr.reason,
		Calls:        sr.calls,
		Observations: sr.observations,
		Artifacts:    sr.artifacts,
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

// outsidePlan records the one condition that stops the whole run without anything
// having been sent: the step would have had to make a request the operator's
// approval does not cover.
func (sr *stepRun) outsidePlan(err error) {
	sr.verdict, sr.reason = VerdictFail, truncateError(err)
	sr.abort = err
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
	case errors.Is(err, ErrOutsidePlan):
		sr.outsidePlan(err)
	case errors.Is(err, ErrM0TerminalHold):
		sr.fail("%s", truncateError(err))
		sr.abort = err
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

// filled records the other ending: the object did what it was for and stopped
// existing. An order filled; a conditional fired and became one.
//
// It carries the chain the object belonged to, so the record can still say the
// conditional and the order it produced were one measurement after both of them
// are terminal and neither is outstanding any more.
func (sr *stepRun) filled(kind, id, symbol string, at time.Time, chain, note string) {
	sr.artifacts = append(sr.artifacts, Artifact{
		Kind: kind, ID: id, Symbol: symbol, FilledAt: at.UTC(), Filled: true,
		ChainID: chain, Note: note,
	})
}

// markHeld records that this object stays live on purpose and says who is
// waiting on it.
//
// Two readers, two fields. Deliberate is for the person: the end-of-run check and
// every screen use it to say "this is not a leak". HeldUntil is for the cleanup
// rule: the object is not offered for cancellation until gate reaches a terminal
// verdict recorded after this line (cleanup.go holdGate).
//
// It only reaches an artifact this step actually recorded. A step that reads an
// object back without creating one — conditional-persist is the case today — has
// nothing to mark, and calling this is harmless there rather than a second way to
// declare a hold.
func (sr *stepRun) markHeld(kind, id string, gate StepID, chain, note string) {
	for i := range sr.artifacts {
		if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].Cancelled {
			sr.artifacts[i].Deliberate = true
			sr.artifacts[i].HeldUntil = gate
			sr.artifacts[i].ChainID = chain
			sr.artifacts[i].Note = note
		}
	}
}

// joinChain records which measurement an object belongs to, without claiming
// anything about whether it is allowed to stay alive.
//
// markHeld says three things at once — this is deliberate, this step's verdict
// releases it, and it is part of this chain — and the three are not always true
// together. An object that is *supposed* to end inside the step that made it
// belongs to a chain and nothing more: exempting it from the end-of-run leftover
// check would hide exactly the case where the step did not get to finish.
func (sr *stepRun) joinChain(kind, id, chain, note string) {
	for i := range sr.artifacts {
		if sr.artifacts[i].Kind == kind && sr.artifacts[i].ID == id && !sr.artifacts[i].terminal() {
			sr.artifacts[i].ChainID = chain
			sr.artifacts[i].Note = note
		}
	}
}

// chainOf returns the chain identifier the record already carries for an object,
// newest mention first: this step, then what this run has written, then the
// record it started from.
//
// A modify creates a new identifier and invalidates the old one at once
// (measurements.md M19, M40), so the successor has to inherit its predecessor's
// chain here — after the fact the broker cannot say the two were one protection.
func (r *Runner) chainOf(sr *stepRun, kind, id string) string {
	for _, entries := range [][]Entry{{{Artifacts: sr.artifacts}}, r.written, r.prior} {
		if c := ChainOf(entries, kind, id); c != "" {
			return c
		}
	}
	return ""
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
