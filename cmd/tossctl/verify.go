package main

// verify.go is `tossctl verify`: the guided one-off live verification that the
// automation gate's remaining measurements need a person to run (openspec change
// verify-execution-capability, task 1.5).
//
// # Why this is a command and not a test
//
// The things it measures — that an order can be placed and cancelled on this
// account, that a conditional order survives the process exiting, that the
// broker's idempotency key does what the document says — cannot be established by
// anything that does not place a real order. WORKFLOW's 불변 규칙 forbid an
// automated test from doing that, so it is an operator tool: it runs at a
// terminal, and nothing is sent until a person has typed an expiring string.
//
//	tossctl verify run --list      print the whole procedure, touch nothing
//	tossctl verify run             walk it
//	tossctl verify run --resume    continue an interrupted or halted run
//	tossctl verify status          how far it got, and what is still live
//	tossctl verify report          the attributes tasks 2.6 and 1.4 consume
//
// # One approval for the run, or one per mutation
//
// The default is one typed confirmation for the whole run: it prints every live
// request it plans to make — action, symbol, side, quantity, how the price is
// derived, how the exposure ends — and waits for a single expiring string that
// approves exactly that list (tasks.md 1.5, 사용자 결정 2026-07-26). A request the
// list does not carry a line for is never sent, and a step that would have to send
// one stops the run rather than adapting to it. --confirm-each is the finer gate —
// one prompt immediately before each mutation — and it is kept for the operator who
// wants to be able to stop halfway through a boundary probe.
//
// # No --yes, and no --base-url
//
// Two absences, for two different reasons. There is no automation flag because
// the spec forbids one outright ("자동화 플래그 금지") and because an agent
// answering these prompts would be an agent trading somebody's account. There is
// no way to point the tool at another server because its output is the evidence
// an attestation is written from, and evidence produced against a server of one's
// own choosing is not evidence. The tests replace verifyBrokerFactory, which is
// not reachable from the command line — the same arrangement `tossctl soak` uses.
//
// # What it will refuse to do
//
// Run without official credentials. Run against a US symbol (the amend and
// price-band rules the steps rely on are KR's). Place more than one order at a
// time. Finish while something it created is still live. Buy anything to create a
// holding a step needs — those steps are skipped with a reason instead.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/ratebudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/tui"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

type verifyOptions struct {
	record string
	market string

	symbol          string
	holdingSymbol   string
	offsetPct       float64
	maxSellQuantity float64
	includeTTLEdge  bool
	includeTrigger  bool
	confirmEach     bool
	ttlWait         time.Duration
	triggerWindow   time.Duration
	resume          bool
	redo            []string
	list            bool

	// why is the reason an abort records against the chains it closes.
	why string
}

func newVerifyCmd(root *rootOptions) *cobra.Command {
	opts := &verifyOptions{}

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Run and report the supervised live-account capability verification",
		Long: strings.TrimSpace(`
Measure, on a real account, the things a read-only survey cannot prove: that
orders can be placed, amended and cancelled; that a conditional order survives
this process exiting; what the broker's idempotency key actually does; what a
resting sell does to the sellable quantity.

It places real orders. Each one is a single share, limit-only, priced far enough
from the market that it cannot fill, and cancelled inside the step that placed it.
Before any of them is sent, the run lists every request it plans to make and waits
for one expiring string, typed at a terminal. There is no flag that answers that
prompt.

  tossctl verify run --list    print the whole procedure and touch nothing
  tossctl verify run           walk it
  tossctl verify status        how far it got, and what is still live
  tossctl verify report        the measured attributes, and what is still unverified

Read ` + "`tossctl verify run --list`" + ` before the first real run.`),
	}

	cmd.AddCommand(
		newVerifyRunCmd(root, opts),
		newVerifyStatusCmd(root, opts),
		newVerifyReportCmd(root, opts),
		newVerifyAbortCmd(root, opts),
	)
	return cmd
}

func newVerifyRunCmd(root *rootOptions, opts *verifyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Walk the live verification procedure under one typed batch approval",
		Long: strings.TrimSpace(`
Walk the verification steps in order, appending the evidence for each one to a
durable local record.

Before anything is sent, the run prints a numbered list of every live request it
plans to make — action, symbol, side, quantity, how the limit price is derived and
how each exposure ends — and waits for ONE expiring confirmation string to be typed
at a terminal. Typing it approves that list. Anything else aborts the run before a
single request goes out.

The list is the boundary, not a preview. A request it does not carry a line for is
never sent: if conditions change what a step would have to send — a different
symbol, a different side, more than the approved quantity — the run stops and asks
you to start again with the new list in front of you. Prices are the exception that
proves it, and the list says so: each order is re-quoted by the stated rule (so
many percent from the last trade, snapped to the tick grid, clamped inside the day's
band) at the moment its step runs.

--confirm-each opts out of the batch and back into a separate typed confirmation
immediately before every single mutation. Refusing one of those refuses that step
only, and the run continues with the steps that do not depend on it.

Safety rules the command enforces for you:

  one share      every order is the minimum quantity, LIMIT only
  far away       prices are set a configurable distance from the last trade and
                 clamped inside the day's price band, so they cannot fill
  one at a time  never more than one live order from this tool
  cleaned up     an order is cancelled inside the step that placed it, and the
                 command fails loudly if anything is left behind

Two steps are special. The conditional-order persistence check cannot pass inside
the process that registered the conditional, so the run stops there and asks you
to start a new one with --resume — and that resumed run approves its remaining
batch from scratch, because the earlier approval covered earlier requests. The
idempotency validity-window check deliberately creates a second live order and is
therefore skipped unless you pass --include-ttl-edge.

Steps that need an existing holding are skipped with a reason when the account
has none. The tool never buys anything to create one.`),
		// official: every request goes to the Open API. mutating: it places live
		// orders, one confirmed share at a time.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyRun(cmd, root, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.list, "list", false,
		"Print the procedure and exit; no credentials are read and no request is made")
	cmd.Flags().BoolVar(&opts.resume, "resume", false,
		"Continue a verification that is already on the record")
	cmd.Flags().StringVar(&opts.market, "market", verifylive.MarketKR,
		"Market this run measures: KR or US. A run sends orders only for symbols in its own market")
	cmd.Flags().StringVar(&opts.symbol, "symbol", "005930",
		"Symbol the buy-side probes are placed against (KR default; a US run uses a held US symbol)")
	cmd.Flags().StringVar(&opts.holdingSymbol, "holding-symbol", "",
		"Held symbol the sell-side and conditional steps use (default: the first suitable holding in this market)")
	cmd.Flags().Float64Var(&opts.offsetPct, "offset-pct", verifylive.DefaultOffset*100,
		"How far from the last trade orders are priced, in percent")
	cmd.Flags().Float64Var(&opts.maxSellQuantity, "max-sell-quantity", verifylive.DefaultMaxSellQuantity,
		"Largest whole-holding sell the boundary step may place; above this the boundary is left unverified")
	cmd.Flags().BoolVar(&opts.includeTTLEdge, "include-ttl-edge", false,
		"Also probe the idempotency validity window — this deliberately creates a SECOND live order")
	cmd.Flags().BoolVar(&opts.includeTrigger, "include-trigger", false,
		"Also observe a conditional order FIRING. This registers a stop the market is meant to reach: "+
			"if it does, one share is sold at market and it cannot be undone")
	cmd.Flags().DurationVar(&opts.triggerWindow, "trigger-window", verifylive.DefaultTriggerWindow,
		"How long the trigger observation waits for the market to reach its trigger before cancelling it")
	cmd.Flags().BoolVar(&opts.confirmEach, "confirm-each", false,
		"Ask for a separate typed confirmation immediately before every mutation, instead of one for the run")
	cmd.Flags().DurationVar(&opts.ttlWait, "ttl-wait", verifylive.DefaultTTLWait,
		"How long the validity-window probe waits before replaying the key")
	cmd.Flags().StringSliceVar(&opts.redo, "redo", nil,
		"Run these steps again even though the record already has a verdict for them")
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the evidence record path")

	return cmd
}

func newVerifyStatusCmd(root *rootOptions, opts *verifyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report how far the live verification got and what is still live on the account",
		Long: strings.TrimSpace(`
Read the evidence record and report each step's verdict, what is still pending,
and — the part that matters after an interrupted run — any order or conditional
order this tool created and has not cancelled.

Reads only the local record. It makes no network call.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyStatus(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.market, "market", verifylive.MarketKR,
		"Market whose evidence record to read: KR or US")
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the evidence record path")
	return cmd
}

func newVerifyReportCmd(root *rootOptions, opts *verifyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Render the measured capability attributes and the ones still unverified",
		Long: strings.TrimSpace(`
Turn the evidence record into the attribute set the rest of the change consumes:
the ProtectiveCapability properties, the idempotency-key behaviour, the
sellable-quantity semantics and the realised costs.

Anything that was not measured is printed as "unverified" rather than omitted.
That list is what decides which markets and order types automatic entry stays
forbidden on.

Idempotent replay is reported as DISABLED unless the record positively shows a
repeated key returning the first order and creating nothing.

Reads only the local record. It makes no network call.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyReport(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.market, "market", verifylive.MarketKR,
		"Market whose evidence record to read: KR or US")
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the evidence record path")
	return cmd
}

// --- run ----------------------------------------------------------------------

func runVerifyRun(cmd *cobra.Command, root *rootOptions, opts *verifyOptions) error {
	out := cmd.OutOrStdout()

	// --list is the read-only preview. It reads no credentials and makes no call,
	// so it is safe to run anywhere, which is the point: an operator should be
	// able to see the whole procedure before deciding to run it.
	if opts.list {
		verifylive.WriteSteps(out, opts.includeTTLEdge, opts.includeTrigger)
		return nil
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// Ctrl-C has to leave the record intact and, more importantly, has to let the
	// run report what is still live rather than being killed mid-cancel.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	executionLock, err := acquireVerifyExecutionLock(root)
	if err != nil {
		return err
	}
	defer executionLock.Release()
	fmt.Fprintf(out, "execution lock   %s (engine · update · verification exclusion)\n",
		executionLock.Path())

	market := verifylive.NormalizeMarket(opts.market)
	recordPath, err := resolveVerifyRecordFor(root, opts.record, market)
	if err != nil {
		return err
	}
	prior, err := verifylive.LoadEntries(recordPath)
	if err != nil {
		return err
	}
	// Steps, not lines: a run whose batch approval was declined leaves the refusal
	// on the record and nothing else, and that must not stand between the operator
	// and a second attempt.
	if steps := verifylive.StepCount(prior); steps > 0 && !opts.resume && len(opts.redo) == 0 {
		return fmt.Errorf(
			"verify: %s already holds %d step(s) of a verification. Continue it with `tossctl verify run "+
				"--resume`, inspect it with `tossctl verify status`, or start a separate one with --record. "+
				"Starting over silently would place live orders for measurements already made",
			recordPath, steps)
	}
	releaseIntent, err := holdVerifyRateBudgetIntent(ctx, out, root)
	if err != nil {
		return err
	}
	defer releaseIntent()
	budgetLease, err := acquireVerifyRateBudget(ctx, out, root)
	if err != nil {
		return err
	}
	defer budgetLease.Release()

	broker, accountRef, err := verifyBrokerFactory(root)
	if err != nil {
		return err
	}

	holding := strings.TrimSpace(opts.holdingSymbol)
	if holding == "" {
		holding = firstUsableHoldingIn(ctx, broker, market)
	}
	symbol := strings.TrimSpace(opts.symbol)
	if market == verifylive.MarketUS && !verifylive.SameMarket(verifylive.MarketOf(symbol), market) {
		// The --symbol default is KR's. A US run that kept it would plan orders
		// this run is not allowed to send, so it falls back to the held US symbol
		// and the runner's own gates skip the buy-side probes if there is none.
		symbol = holding
	}

	recorder, err := verifylive.OpenRecorder(recordPath)
	if err != nil {
		return err
	}
	defer recorder.Close()

	runner, err := verifylive.New(verifylive.Options{
		Broker:          broker,
		Recorder:        recorder,
		Confirm:         terminalConfirmer(cmd),
		ConfirmBatch:    terminalBatchConfirmer(cmd),
		ConfirmEach:     opts.confirmEach,
		Out:             out,
		AccountRef:      accountRef,
		Market:          market,
		Symbol:          symbol,
		HoldingSymbol:   holding,
		Offset:          opts.offsetPct / 100,
		MaxSellQuantity: opts.maxSellQuantity,
		IncludeTTLEdge:  opts.includeTTLEdge,
		IncludeTrigger:  opts.includeTrigger,
		TTLWait:         opts.ttlWait,
		TriggerWindow:   opts.triggerWindow,
		Redo:            toStepIDs(opts.redo),
		Prior:           prior,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "evidence record  %s\n", recordPath)

	summary, runErr := runner.Run(ctx)
	writeVerifySummary(out, recordPath, summary)

	if runErr != nil && (errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded)) {
		fmt.Fprintln(out, "\ninterrupted — everything recorded so far is kept.")
		return nil
	}
	return runErr
}

// acquireVerifyExecutionLock closes the cross-process start race among the
// engine, executable replacement, and either verification entry point.
//
// The runlock marker remains advisory because its consumer is the soak. This is
// the actual exclusion: a crash-released kernel flock in the journal directory,
// acquired before account resolution and held through complete runner cleanup.
func acquireVerifyExecutionLock(root *rootOptions) (*enginelock.Lock, error) {
	dir, err := engineJournalDir(root)
	if err != nil {
		return nil, fmt.Errorf("verify: resolving execution exclusion: %w", err)
	}
	lock, err := enginelock.Acquire(dir)
	if err != nil {
		return nil, fmt.Errorf(
			"verify: engine, system update, or another verification owns the execution exclusion: %w",
			err)
	}
	return lock, nil
}

// writeVerifySummary prints the end-of-run state, with the outstanding objects
// first because that is what an operator has to act on.
func writeVerifySummary(w interface{ Write([]byte) (int, error) }, recordPath string, s verifylive.Summary) {
	fmt.Fprintf(w, "\nrun %s\n", s.RunID)
	for _, o := range s.Outcomes {
		note := o.Reason
		if o.AlreadySettled {
			note = "already on the record"
		}
		fmt.Fprintf(w, "  %-22s %-16s %s\n", o.Step, o.Verdict, note)
	}
	if len(s.Outstanding) > 0 {
		fmt.Fprintf(w, "\nstill live on the account:\n")
		for _, a := range s.Outstanding {
			why := "NOT CANCELLED — deal with this"
			if a.Deliberate {
				why = "left on purpose; `tossctl verify run --resume` cancels it"
			}
			fmt.Fprintf(w, "  %s %s (%s) — %s\n", a.Kind, a.ID, a.Symbol, why)
		}
	}
	if s.Halted {
		fmt.Fprintf(w, "\nhalted: %s\n", s.Halt)
	}
	fmt.Fprintf(w, "\n`tossctl verify report --record %s` renders the measured attributes.\n", recordPath)
}

// terminalConfirmer binds the confirmation to the real terminal.
//
// tui.IsInteractive is evaluated per prompt against os.Stdin and os.Stdout rather
// than the cobra streams: a caller that redirected the command's streams must not
// thereby satisfy the terminal check, since redirection is exactly the case the
// check exists to catch.
func terminalConfirmer(cmd *cobra.Command) verifylive.Confirmer {
	return func(m verifylive.Mutation) error {
		return verifylive.Confirm(
			cmd.InOrStdin(), cmd.OutOrStdout(), m,
			tui.IsInteractive(os.Stdin, os.Stdout), time.Now(),
		)
	}
}

// terminalBatchConfirmer binds the run-wide approval to the same real terminal,
// under the same rule: a redirected stdin does not satisfy it.
func terminalBatchConfirmer(cmd *cobra.Command) verifylive.BatchConfirmer {
	return func(b verifylive.Batch) error {
		return verifylive.ConfirmBatch(
			cmd.InOrStdin(), cmd.OutOrStdout(), b,
			tui.IsInteractive(os.Stdin, os.Stdout), time.Now(),
		)
	}
}

// firstUsableHolding picks the held KR symbol the sell-side and conditional steps
// will use.
//
// Best effort: a failure here means those steps are skipped with a reason, which
// is the same outcome as an account with nothing in it and a better one than
// aborting a run whose read-only steps would have worked.
func firstUsableHolding(ctx context.Context, broker verifylive.Broker) string {
	return firstUsableHoldingIn(ctx, broker, verifylive.MarketKR)
}

// firstUsableHoldingIn picks a holding the sell-side and conditional steps can
// use in one market.
//
// Usable means whole shares: this tool places no fractional orders, so a holding
// below one share (a US fractional position) is passed over rather than probed
// with an order the broker would refuse.
func firstUsableHoldingIn(ctx context.Context, broker verifylive.Broker, market string) string {
	positions, err := broker.Holdings(ctx, "")
	if err != nil {
		return ""
	}
	for _, p := range positions {
		if p.Quantity < verifylive.MinQuantity {
			continue
		}
		if !verifylive.SameMarket(verifylive.MarketOf(p.Symbol), market) {
			continue
		}
		return p.Symbol
	}
	return ""
}

func toStepIDs(names []string) []verifylive.StepID {
	out := make([]verifylive.StepID, 0, len(names))
	for _, n := range names {
		if trimmed := strings.TrimSpace(n); trimmed != "" {
			out = append(out, verifylive.StepID(trimmed))
		}
	}
	return out
}

// --- abort ----------------------------------------------------------------------

func newVerifyAbortCmd(root *rootOptions, opts *verifyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort",
		Short: "End a measurement chain this tool is still holding, and cancel what it holds",
		Long: strings.TrimSpace(`
Cancel every order and conditional order the evidence record says this tool created
and has not finished with, and record that the measurement they belonged to is over.

This exists because objects are released by a verdict and never by a clock. A
conditional order waits for the cancel step; a child order waits for the trigger
observation to decide. That is deliberate — a long wait means the market has not
come to the price yet, not that anything failed, and a lease that expired halfway
through would cancel the very thing being measured. The cost is that a measurement
which never reaches a verdict holds its objects indefinitely, and the cleanup
prologue will not offer them because it is doing what it was told.

So this is the operator saying "I am ending this".

  tossctl verify abort --list    show exactly what would be cancelled, send nothing
  tossctl verify abort           cancel it, under one approval

It lists the targets first and waits for the same single expiring confirmation
` + "`verify run`" + ` uses. There is no separate prompt and no extra phrase to type.
Nothing off that list is ever sent, and the targets come from this tool's own record
— whatever else is on the account is not its business.`),
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVerifyAbort(cmd, root, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.list, "list", false,
		"Print what would be cancelled and exit; no request is made")
	cmd.Flags().StringVar(&opts.market, "market", verifylive.MarketKR,
		"Market whose evidence record to end: KR or US")
	cmd.Flags().StringVar(&opts.why, "why", "",
		"Why the measurement is being ended, recorded against the chains it closes")
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the evidence record path")
	return cmd
}

func runVerifyAbort(cmd *cobra.Command, root *rootOptions, opts *verifyOptions) error {
	out := cmd.OutOrStdout()
	// --list is the read-only half, and it is listed first because "what would this
	// cancel" has to be answerable without credentials or a network call.
	if opts.list {
		recordPath, prior, err := loadVerifyRecord(root, opts)
		if err != nil {
			return err
		}
		writeVerifyAbortTargets(out, recordPath, verifylive.AbortTargets(prior))
		fmt.Fprintln(out, "\n--list: 아무것도 전송되지 않았다.")
		return nil
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()
	executionLock, err := acquireVerifyExecutionLock(root)
	if err != nil {
		return err
	}
	defer executionLock.Release()
	fmt.Fprintf(out, "execution lock   %s (engine · update · verification exclusion)\n",
		executionLock.Path())

	releaseIntent, err := holdVerifyRateBudgetIntent(ctx, out, root)
	if err != nil {
		return err
	}
	defer releaseIntent()
	budgetLease, err := acquireVerifyRateBudget(ctx, out, root)
	if err != nil {
		return err
	}
	defer budgetLease.Release()

	// Admission can wait for one bounded metadata read. Read the record only after
	// exclusivity so the cancellation plan cannot omit an artifact appended while
	// this command was waiting.
	recordPath, prior, err := loadVerifyRecord(root, opts)
	if err != nil {
		return err
	}
	targets := verifylive.AbortTargets(prior)
	writeVerifyAbortTargets(out, recordPath, targets)
	if len(targets) == 0 {
		return nil
	}

	broker, accountRef, err := verifyBrokerFactory(root)
	if err != nil {
		return err
	}
	recorder, err := verifylive.OpenRecorder(recordPath)
	if err != nil {
		return err
	}
	defer recorder.Close()

	runner, err := verifylive.New(verifylive.Options{
		Broker:       broker,
		Recorder:     recorder,
		Confirm:      terminalConfirmer(cmd),
		ConfirmBatch: terminalBatchConfirmer(cmd),
		Out:          out,
		AccountRef:   accountRef,
		Market:       verifylive.NormalizeMarket(opts.market),
		Prior:        prior,
	})
	if err != nil {
		return err
	}
	result, abortErr := runner.Abort(ctx, opts.why)
	if !result.Approved && result.Reason != "" {
		fmt.Fprintf(out, "\n%s\n", result.Reason)
	}
	if len(result.Remaining) == 0 && result.Approved {
		fmt.Fprintln(out, "\n계좌에 이 도구가 만든 살아 있는 객체가 남아 있지 않다.")
	}
	if abortErr != nil && (errors.Is(abortErr, context.Canceled) || errors.Is(abortErr, context.DeadlineExceeded)) {
		fmt.Fprintln(out, "\ninterrupted — everything recorded so far is kept.")
		return nil
	}
	return abortErr
}

func writeVerifyAbortTargets(out io.Writer, recordPath string, targets []verifylive.Artifact) {
	fmt.Fprintf(out, "evidence record  %s\n", recordPath)
	if len(targets) == 0 {
		fmt.Fprintln(out, "이 도구의 기록에 살아 있는 객체가 없다 — 끝낼 사슬이 없다.")
		return
	}
	fmt.Fprintf(out, "\n끝낼 대상 %d건:\n", len(targets))
	for _, a := range targets {
		held := ""
		if a.HeldUntil != "" {
			held = fmt.Sprintf(" — %s의 판정을 기다리며 붙잡혀 있다", a.HeldUntil)
		}
		fmt.Fprintf(out, "  %s %s (%s)%s\n", a.Kind, a.ID, a.Symbol, held)
	}
}

// --- status and report ----------------------------------------------------------

func runVerifyStatus(cmd *cobra.Command, root *rootOptions, opts *verifyOptions) error {
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}
	recordPath, entries, err := loadVerifyRecord(root, opts)
	if err != nil {
		return err
	}
	progress := verifylive.BuildProgress(recordPath, entries)

	out := cmd.OutOrStdout()
	if format == output.FormatJSON {
		return output.WriteJSON(out, progress)
	}
	progress.WriteText(out)
	return nil
}

func runVerifyReport(cmd *cobra.Command, root *rootOptions, opts *verifyOptions) error {
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}
	recordPath, entries, err := loadVerifyRecord(root, opts)
	if err != nil {
		return err
	}
	report := verifylive.BuildReport(recordPath, entries, time.Now().UTC())

	out := cmd.OutOrStdout()
	if format == output.FormatJSON {
		return output.WriteJSON(out, report)
	}
	report.WriteText(out)
	return nil
}

func loadVerifyRecord(root *rootOptions, opts *verifyOptions) (string, []verifylive.Entry, error) {
	recordPath, err := resolveVerifyRecordFor(root, opts.record, opts.market)
	if err != nil {
		return "", nil, err
	}
	entries, err := verifylive.LoadEntries(recordPath)
	if err != nil {
		return recordPath, nil, err
	}
	return recordPath, entries, nil
}

// resolveVerifyRecord decides where the evidence lives.
//
// It follows internal/soak's rule exactly, so the two records sit side by side:
// an explicit --config-dir means an isolated profile and the record follows it,
// and otherwise the record is durable state and belongs in the data directory
// next to the journal, not among the configuration.
func resolveVerifyRecord(root *rootOptions, override string) (string, error) {
	return resolveVerifyRecordFor(root, override, verifylive.MarketKR)
}

// resolveVerifyRecordFor is resolveVerifyRecord for one market.
//
// A capability verdict belongs to an account *and* a market, so each market keeps
// its own file (verifylive.RecordFileName). An explicit --record still wins, and
// an unspecified market resolves to the KR file every existing record lives in.
func resolveVerifyRecordFor(root *rootOptions, override, market string) (string, error) {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed, nil
	}
	name := verifylive.RecordFileName(market)
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		return filepath.Join(root.configDir, name), nil
	}
	dir, err := journal.DataDir()
	if err != nil {
		return "", fmt.Errorf("verify: resolving the record location: %w", err)
	}
	return filepath.Join(dir, name), nil
}

// --- the soak/verify rate-budget marker -------------------------------------------

// verifyRunLockPath is where the advisory marker lives: beside the evidence
// record, so an isolated --config-dir profile gets its own and the default one
// sits in the data directory with everything else this change writes.
func verifyRunLockPath(recordPath string) string {
	return filepath.Join(filepath.Dir(recordPath), runlock.FileName)
}

func verifyRateBudgetPath(root *rootOptions) (string, error) {
	dir, err := engineJournalDir(root)
	if err != nil {
		return "", fmt.Errorf("verify: resolving the Open API rate-budget directory: %w", err)
	}
	return filepath.Join(dir, ratebudget.FileName), nil
}

func acquireVerifyRateBudget(ctx context.Context, out io.Writer, root *rootOptions) (*ratebudget.Lease, error) {
	path, err := verifyRateBudgetPath(root)
	if err != nil {
		return nil, err
	}
	lease, err := ratebudget.Acquire(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("verify: reserving the Open API rate budget: %w", err)
	}
	fmt.Fprintf(out, "rate budget lock %s (verification excludes optional metadata reads)\n", lease.Path())
	return lease, nil
}

// holdVerifyRateBudgetIntent gives supervised verification priority over
// optional metadata before it waits for the kernel lease. The execution flock is
// already held by every caller, so this profile marker has exactly one owner and
// its remove-on-release semantics cannot erase another live operation's intent.
func holdVerifyRateBudgetIntent(ctx context.Context, out io.Writer, root *rootOptions) (func(), error) {
	budgetPath, err := verifyRateBudgetPath(root)
	if err != nil {
		return nil, err
	}
	profileRecord := filepath.Join(filepath.Dir(budgetPath), verifylive.FileName)
	path := verifyRunLockPath(profileRecord)
	release, err := runlock.Hold(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("verify: publishing the Open API rate-budget intent: %w", err)
	}
	fmt.Fprintf(out, "soak pause       %s (required admission intent; optional metadata yields while verification starts)\n", path)
	return release, nil
}

// holdVerifyRunLock marks the account as busy for the duration of a verification.
//
// It is advisory in both directions and the return type says so: the release is
// always callable, and a failure to take the marker is a line of output rather
// than a reason to refuse. The separate journal-directory execution flock has
// already excluded engine/update/other-verification starts; this marker only
// asks the soak to yield its rate budget.
func holdVerifyRunLock(ctx context.Context, out io.Writer, recordPath string) func() {
	path := verifyRunLockPath(recordPath)
	release, err := runlock.Hold(ctx, path)
	if err != nil {
		fmt.Fprintf(out, "note: the soak pause marker could not be written (%v). "+
			"A soak running against this account will keep going and may spend the rate budget "+
			"this run needs.\n", err)
		return release
	}
	fmt.Fprintf(out, "soak pause       %s (advisory; `tossctl soak run` delays its cycle while this run "+
		"is live)\n", path)
	return release
}

// verifyRunLockPause is the reader half, handed to the soak.
//
// It answers a question and never blocks: whether a live verification touched the
// marker inside runlock.StaleAfter. A crashed verification therefore costs one
// paused cycle rather than a wedged survey.
func verifyRunLockPause(lockPath string) func() (bool, string) {
	return func() (bool, string) {
		fresh, at := runlock.Fresh(lockPath, time.Now(), runlock.StaleAfter)
		if !fresh {
			return false, ""
		}
		return true, fmt.Sprintf(
			"a live verification is running (%s, touched %s). Yielding the rate budget: a step lost to a "+
				"429 costs a real order and a person's attention, and this cycle does not",
			lockPath, at.UTC().Format(time.RFC3339))
	}
}

// --- the broker adapter ---------------------------------------------------------

// verifyBrokerFactory builds the live client. It is a package variable so this
// package's own tests can point the verification at an httptest server; see the
// note at the top of this file about why there is no flag for it.
var verifyBrokerFactory = buildVerifyBroker

func buildVerifyBroker(root *rootOptions) (verifylive.Broker, string, error) {
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, "", err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return nil, "", fmt.Errorf("verify: reading the Open API credentials: %w", err)
	}
	if creds == nil {
		return nil, "", fmt.Errorf(
			"verify: no Open API credentials — orders are placed through the official API only, never the " +
				"web session. Run `tossctl openapi login` or set TOSSCTL_OPENAPI_KEY/TOSSCTL_OPENAPI_SECRET")
	}
	client := official.New(*creds, tokenFile)

	ref, seq, err := resolveVerifyAccount(context.Background(), client, sleepFor)
	if err != nil {
		return nil, "", err
	}
	if seq == 0 {
		// The reference is usable and the sequence is not. The client falls back to
		// its own lazy resolution, which is what every other command does.
		return client, ref, nil
	}
	// Rebuilt with the sequence already known, so no step triggers the lazy
	// /api/v1/accounts resolution on its first account-scoped GET. That resolution
	// is the call that came back 429 three times on 2026-07-26 and cost the run
	// three steps (measurements.md M4); it was also redundant, because the read
	// above already had the answer. The token is cached on disk, so the second
	// client re-uses it rather than exchanging again.
	return official.New(*creds, tokenFile, official.WithAccountSeq(seq)), ref, nil
}

// resolveVerifyAccount names the account every prompt and every record line will
// carry (masked), and returns the sequence number the same entry is keyed by.
//
// Both come from one entry on purpose. The reference is what the evidence says
// was measured and the sequence is what the X-Tossinvest-Account header selects;
// taking them from different accounts would produce a record that names one
// account and measured another. seq is 0 when the entry's identifier is not a
// number the header can carry, which is not fatal — the client resolves it
// lazily, exactly as it did before.
//
// Without a reference there is nothing to attest about, so that half is a hard
// failure rather than a blank field. A 429 is retried under verifylive's read
// policy: this is the read that was rate limited, and failing here costs the whole
// run before a person has been asked anything.
func resolveVerifyAccount(ctx context.Context, client interface {
	Accounts(context.Context) ([]domain.Account, error)
}, sleep func(context.Context, time.Duration) error) (string, int, error) {
	var (
		accounts []domain.Account
		err      error
	)
	for extra := 0; ; extra++ {
		accounts, err = client.Accounts(ctx)
		if err == nil || !errors.Is(err, official.ErrRateLimited) || extra >= verifylive.ReadRetryExtraAttempts {
			break
		}
		if sleepErr := sleep(ctx, verifylive.ReadRetryBackoff(extra)); sleepErr != nil {
			break
		}
	}
	if err != nil {
		return "", 0, fmt.Errorf("verify: the account could not be identified: %w", err)
	}
	for _, a := range accounts {
		ref := strings.TrimSpace(a.DisplayName)
		if ref == "" {
			continue
		}
		seq, convErr := strconv.Atoi(strings.TrimSpace(a.ID))
		if convErr != nil || seq <= 0 {
			seq = 0
		}
		return ref, seq, nil
	}
	return "", 0, errors.New("verify: the broker returned no account number")
}

// sleepFor is the retry wait, cancellable. It is a variable so a test can drive
// the backoff without spending it.
var sleepFor = func(ctx context.Context, d time.Duration) error {
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
