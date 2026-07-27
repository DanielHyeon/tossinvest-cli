package main

// console.go is `tossctl console`: the local operator console for the one-off
// live-account verification (openspec change verify-execution-capability,
// task 1.6).
//
// # Why the console exists at all
//
// `tossctl verify run` already does the measurement, and it will keep doing it.
// The console exists because the user decided (사용자 결정 2026-07-26) to drive
// the verification from a screen rather than a terminal stream: the batch summary
// is a dozen live requests with prices and reversals, and reading that in a
// scrollback while deciding whether to type a confirmation is worse than reading
// it on a page.
//
// It is a stopgap. internal/console's package doc says so and means it: single
// user, loopback only, deleted when the Phase 4 daemon lands.
//
// # What this file is responsible for
//
// Exactly one thing that matters: it is the only place in the binary that hands
// internal/console a way to reach a live account, and it hands it the same runner
// `verify run` builds, gated on the console's web confirmer instead of the
// terminal's. Everything else — the approval, the session, the rendering — is
// internal/console's.
//
// # The absences, which are the point
//
// There is no flag that presets the session token, answers the approval, or moves
// the console off the loopback interface. `verify run` gains nothing: its
// confirmers are still terminalConfirmer and terminalBatchConfirmer, and
// console_test.go asserts that in the source, because "the CLI must not gain a
// non-interactive approval path" is the condition under which task 1.6 permits the
// web form to exist at all.
//
// --confirm-each is not offered here either. The console is batch-only, and the
// per-mutation confirmer it wires refuses — so the finer gate fails closed rather
// than being silently satisfied if a future edit ever reaches it.

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
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/handoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

// consoleProbeSymbolKR is the KR symbol the buy-side probes are placed against.
//
// It is `verify run --symbol`'s default, restated rather than shared because
// verify.go is a live-order path and this task's scope is additive. Drift is not
// left to review: console_test.go asserts the two agree.
const consoleProbeSymbolKR = "005930"

// consoleProbeSymbol picks the symbol the buy-side probes use in a market.
//
// KR has a default worth pinning: a large, liquid, always-quoted name. US has no
// equivalent constant here on purpose — which US symbols an account can afford to
// place a one-share order against differs per account, so the probe uses a symbol
// the account already holds. Buy and sell probes then share one symbol, which
// keeps the exposure to a single name and makes the opposite-pending rejection
// observable in the same run. No usable US holding means the US probes are
// skipped by the runner's own gates, with the reason recorded.
func consoleProbeSymbol(ctx context.Context, broker verifylive.Broker, market string) string {
	if verifylive.NormalizeMarket(market) != verifylive.MarketUS {
		return consoleProbeSymbolKR
	}
	return firstUsableHoldingIn(ctx, broker, verifylive.MarketUS)
}

type consoleOptions struct {
	port int
}

func newConsoleCmd(root *rootOptions) *cobra.Command {
	opts := &consoleOptions{}

	cmd := &cobra.Command{
		Use:   "console",
		Short: "Serve the local operator console for the live-account verification",
		Long: strings.TrimSpace(`
Serve a small web console on 127.0.0.1 for driving the one-off live-account
verification from a browser instead of a terminal.

It binds the loopback interface and nothing else. There is no flag to change
that. On start it prints a URL carrying this process's session token — it stays
valid until the console stops, and it is not single-use (the single-use one is
the restart handoff token). Opening that URL in this machine's browser is what
authenticates you, so possession of this terminal is the credential. Do not
paste the link anywhere else.

  overview    /dashboard — engine state, holdings, today's realised P&L, the
              leftovers and the Guardian limits, gathered per market. Read-only:
              no form, and it makes no broker call of its own
  console     / — soak progress, attestation state, verification progress
  positions   what the account holds, joined to the engine's exit lines
  history     completed round trips and the exit judgement stream
  verify      the step list, the batch summary, the typed approval, live progress
  report      the measured attributes and the ones still unverified, plus JSON

Everything is read-only except the verification approval. The console places no
order of its own, toggles no gate, edits no configuration and shows no
credential: the only requests that reach the account are the ones the verify
runner makes, under the same plan authorisation, the same exposure caps and the
same cancellation rails as ` + "`tossctl verify run`" + `.

Approving is a three-part act and all three are required: the session token, a
CSRF token the page carries, and the expiring confirmation string the page shows
you, typed back by hand. Anything missing or wrong sends nothing. There is no
flag, here or on ` + "`tossctl verify run`" + `, that answers it for you.

The conditional-order persistence check needs a NEW process, so this console runs
at most one verification per start: when it stops there, quit with Ctrl-C, start
the console again, and press 이어하기.

A step whose last verdict is fail or skipped can be attempted again from the
verify screen's 재측정 button — the market was closed, the broker throttled, the
account held nothing. The set is worked out from the evidence record, never from
the page: a step that passed is never re-measured, and a re-measurement asks for
its own batch approval with a new confirmation string like any other run.

While a verification is running this command marks the account busy so a
concurrent ` + "`tossctl soak run`" + ` delays its cycle rather than spending the same
rate limit.`),
		// official: the verification it drives reaches the Open API. mutating: it
		// can place live orders — through the verify runner, after a typed approval.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConsole(cmd, root, opts)
		},
	}

	cmd.Flags().IntVar(&opts.port, "port", 0,
		"Loopback port to serve on; 0 lets the OS pick a free one")
	return cmd
}

func runConsole(cmd *cobra.Command, root *rootOptions, opts *consoleOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// Ctrl-C has to let a verification in progress finish cancelling whatever it
	// has resting; internal/console waits for that before it shuts the socket.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	verifyRecord, err := resolveVerifyRecord(root, "")
	if err != nil {
		return err
	}
	verifyRecordUS, err := resolveVerifyRecordFor(root, "", verifylive.MarketUS)
	if err != nil {
		return err
	}
	soakRecord, err := resolveSoakRecord(root, "")
	if err != nil {
		return err
	}
	attestation, err := resolveSoakAttestationPath(root, "")
	if err != nil {
		return err
	}

	journalPath, err := journal.DefaultPath()
	if err != nil {
		// Not fatal. The dashboard's journal half reports "배선되지 않았다" and the
		// verification console — which is what this command is for — is unaffected.
		fmt.Fprintf(cmd.ErrOrStderr(), "원장 경로를 해석할 수 없다 (%v). 포지션·이력 화면은 원장 없이 뜬다.\n", err)
		journalPath = ""
	}

	// The engine's advisory marker lives beside its journal. A resolution failure
	// is not fatal for the same reason the journal path's is not: the dashboard
	// reports the engine section as unwired and the verification console — which
	// is what this command is for — is unaffected.
	engineMarkerPath := ""
	if dir, derr := engineJournalDir(root); derr == nil {
		engineMarkerPath = enginelock.MarkerPath(dir)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "엔진 마커 경로를 해석할 수 없다 (%v). 엔진 상태 표시는 비어 있다.\n", derr)
	}

	out := cmd.OutOrStdout()
	return console.ListenAndServe(ctx, console.Options{
		Port:              opts.port,
		StartVerify:       consoleVerifyStarter(root),
		SoakRecord:        soakRecord,
		VerifyRecord:      verifyRecord,
		VerifyRecordUS:    verifyRecordUS,
		Attestation:       attestation,
		MinSoakDays:       soak.DefaultCriteria().MinConsecutiveDays,
		RequiredEndpoints: engine.RequiredEndpoints(),
		Out:               out,

		// The dashboard's three seams (change add-operator-dashboard). The console
		// reads the journal itself — read-only, per request — and is handed a
		// broker that can do exactly one thing.
		Holdings: newConsoleHoldings(root),
		// The orders screen's read (change console-orders-screen). One method,
		// behind which this file makes the two calls a refresh costs.
		Orders:      consoleOrdersSeam(root),
		JournalPath: journalPath,
		RunLockPath: verifyRunLockPath(verifyRecord),
		Settings:    consoleSettingsSeam(root),

		// The overview's read-only view of the Guardian's ceilings (change
		// console-operator-overview). Five numbers and a currency: the console
		// displays what the file says and can change none of it.
		GateLimits: consoleGateLimitsSeam(root),

		// The three seams task 1.8 puts behind the console's two restart buttons.
		// internal/console executes nothing: it decides whether the person asking
		// has cleared the session and CSRF gates, and then calls one of these.
		Relaunch:    consoleRelaunch(out),
		Handoff:     handoff.New(consoleHandoffPath(verifyRecord)),
		RestartSoak: func() (string, error) { return restartSoak(soakRecord) },

		// The engine's status and its two buttons (change add-engine-runtime,
		// task 2.1). Same arrangement as the two restarts above: internal/console
		// decides whether the person asking cleared both gates, and these do the
		// work. The marker is read-only for the console — the exclusion is the
		// flock the engine holds — and neither button can make the engine able to
		// trade: `engine run` re-checks the §0.7-approved gate and the whole
		// startup interlock every time it comes up, and a refusal comes back here
		// as the engine's own words.
		EngineMarker: engineMarkerPath,
		StartEngine:  func() (string, error) { return startEngine(root) },
		StopEngine:   func() (string, error) { return stopEngine(root) },
	})
}

// --- the dashboard's broker (change add-operator-dashboard, task 1.3) ------------
//
// The console declares a one-method interface (console.HoldingsReader) and this
// is the only place in the binary that satisfies it. What crosses the boundary is
// a *bound method value*, not a broker: the Open API client that owns
// PlaceOrder, CancelOrder and the conditional-order mutations never becomes
// reachable from internal/console, so "the dashboard cannot send an order" is a
// fact about the wiring rather than about the handlers.

// consoleSettingsSeam adapts the adoption-settings seam
// (adoptionsettings.go) to the console's interface. The nil check happens on
// the concrete pointer HERE: a typed-nil inside the interface would defeat the
// console's own `Settings != nil` wiring test.
func consoleSettingsSeam(root *rootOptions) console.AdoptionSettings {
	if s := newAdoptionSettingsSeam(root); s != nil {
		return s
	}
	return nil
}

// consoleGateLimitsSeam hands the overview screen the Guardian's ceilings, and
// nothing else (change console-operator-overview task 5.1).
//
// It reads. There is no writer here and there is not going to be one: turning the
// automation gate on or moving a limit is a §0.7 human decision taken outside any
// browser, and the console's spec says in writing that it cannot edit them. What
// crosses the boundary is five float64s and a currency string — not the config
// service, and not the gate's own type, which internal/console is forbidden to
// name at all (its TestTheConsoleDecidesNothingAboutTheGate).
//
// A console with no resolvable config file gets no seam, and the overview renders
// the limits as seam 미배선 rather than as zero. The same nil-on-the-concrete-type
// care as consoleSettingsSeam: a typed-nil inside the interface would look wired.
func consoleGateLimitsSeam(root *rootOptions) console.GateLimitsReader {
	svc := configServiceFor(root)
	if svc == nil {
		return nil
	}
	return consoleGateLimits{svc: svc}
}

type consoleGateLimits struct{ svc *config.Service }

// consoleGateLimitsTimeout bounds one read of the config file.
//
// The seam takes no context — it is one method and the console holds nothing else
// of the config service — so the deadline is set here. It matters because of what
// is on the other end: the overview reloads itself every 30 seconds and an
// operator leaves it open all day, and a config read on a wedged filesystem with
// context.Background() behind it would hold an HTTP handler open with no bound at
// all. The overview's whole design is that every failure is a sentence on the
// page; a render that never returns is the one failure it has no words for.
const consoleGateLimitsTimeout = 5 * time.Second

// GateLimits satisfies console.GateLimitsReader.
//
// A read failure is returned rather than swallowed: the overview renders an
// unreadable limit as unmeasured with the error beside it, which is neither zero
// nor unlimited — and both of those would be a screen telling an operator
// something nobody read. A timeout arrives as one of those failures.
func (s consoleGateLimits) GateLimits() (console.GateLimits, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consoleGateLimitsTimeout)
	defer cancel()

	cfg, err := s.svc.Load(ctx)
	if err != nil {
		return console.GateLimits{}, err
	}
	gate := cfg.Engine.AutomationGate
	return console.GateLimits{
		MaxOrderQuantity:   gate.MaxOrderQuantity,
		MaxOrderNotional:   gate.MaxOrderNotional,
		MaxTotalExposure:   gate.MaxTotalExposure,
		MaxDailyLossAmount: gate.MaxDailyLossAmount,
		MaxDailyLossRatio:  gate.MaxDailyLossRatio,
		Currency:           gate.LimitCurrency,
	}, nil
}

// newConsoleHoldings builds the console's holdings reader, lazily.
//
// Lazily because constructing the client resolves the account against the Open
// API, and doing that at `tossctl console` startup would make a screen the
// operator may never open into a precondition for the console coming up at all.
// The first render pays for it; a failure is a sentence on the page.
func newConsoleHoldings(root *rootOptions) console.HoldingsReader {
	return &lazyHoldings{root: root}
}

// holdingsFunc is the single capability the console is handed.
type holdingsFunc func(ctx context.Context, symbol string) ([]domain.Position, error)

type lazyHoldings struct {
	root *rootOptions

	mu   sync.Mutex
	read holdingsFunc
}

// Holdings satisfies console.HoldingsReader.
//
// A build failure is returned rather than remembered: credentials that were not
// there when the console started may be there after `tossctl openapi login`, and
// the console's own cache is what stops a failing build from being retried on
// every render (holdings.go bounds attempts by the TTL, not successes).
func (l *lazyHoldings) Holdings(ctx context.Context, symbol string) ([]domain.Position, error) {
	l.mu.Lock()
	read := l.read
	if read == nil {
		broker, _, err := verifyBrokerFactory(l.root)
		if err != nil {
			l.mu.Unlock()
			return nil, err
		}
		// Only the method value is kept. Nothing here or afterwards holds the
		// wide broker interface.
		read = broker.Holdings
		l.read = read
	}
	l.mu.Unlock()
	return read(ctx, symbol)
}

// --- the orders screen's read (change console-orders-screen, task 4.6) ------------
//
// The console declares one method and this is the only thing in the binary that
// satisfies it. Behind that one method are the two broker calls one refresh
// costs, which is where they belong: the console cannot then spend the budget
// twice, and it cannot report one endpoint's silence as the other's zero.

// consoleOrdersPageLimit bounds one page of each list.
//
// It is a page and not the whole history on purpose — the screen is "what is
// alive and what happened today", and an unbounded walk would be a loop of broker
// calls behind one page load. When the broker says there is more, the console
// renders the count as a floor ("N건 이상") rather than as a number, which is why
// truncating here is honest rather than lossy.
const consoleOrdersPageLimit = 100

// consoleOrdersReader is the part of the live client the orders screen needs.
//
// verifyBrokerFactory returns a verifylive.Broker, which does not declare the
// raw order reads — it declares OrdersPageRaw, whose orders are undecoded JSON.
// The concrete client has both, so the path is chosen by asserting for exactly
// the two reads rather than by building a second *official.Client: that second
// client would resolve the account sequence a second time, and that resolution is
// the call that came back 429 three times on 2026-07-26 and cost a run three
// steps (measurements.md M4).
type consoleOrdersReader interface {
	OrdersRaw(ctx context.Context, filter official.OrdersFilter) (official.RawOrderList, error)
	ConditionalOrdersRaw(ctx context.Context, status, symbol, cursor string,
		limit int) (official.RawConditionalOrderList, error)
}

// consoleOrdersSeam builds the console's orders reader, lazily.
//
// Lazily for newConsoleHoldings' reason: constructing the client resolves the
// account against the Open API, and doing that at `tossctl console` startup would
// make a screen the operator may never open into a precondition for the console
// coming up.
func consoleOrdersSeam(root *rootOptions) console.OrdersReader {
	return &lazyOrders{root: root}
}

type lazyOrders struct {
	root *rootOptions

	mu sync.Mutex
	// Only the two method values are kept, never the broker they came from.
	// Nothing here or afterwards holds an interface with PlaceOrder on it.
	plain       func(context.Context, official.OrdersFilter) (official.RawOrderList, error)
	conditional func(context.Context, string, string, string, int) (official.RawConditionalOrderList, error)
}

// Orders satisfies console.OrdersReader.
//
// The two calls are sequential and each one's outcome is carried separately. A
// failure on one is NOT returned as an error for the pair: the console renders
// that half as unmeasured and refuses to add the counts, and collapsing them into
// one error would take the measured half down with the missing one.
//
// An error is returned only when there is no reading at all to describe — no
// credentials, or a client that does not have these reads.
func (l *lazyOrders) Orders(ctx context.Context) (console.OrdersReading, error) {
	l.mu.Lock()
	plain, conditional := l.plain, l.conditional
	if plain == nil || conditional == nil {
		broker, _, err := verifyBrokerFactory(l.root)
		if err != nil {
			l.mu.Unlock()
			return console.OrdersReading{}, err
		}
		reads, ok := broker.(consoleOrdersReader)
		if !ok {
			l.mu.Unlock()
			return console.OrdersReading{}, fmt.Errorf(
				"console: this build's broker (%T) has no raw order reads; the orders screen needs "+
					"the broker's own decimal strings, because a value that has been through float64 "+
					"cannot say whether the broker sent one at all", broker)
		}
		plain, conditional = reads.OrdersRaw, reads.ConditionalOrdersRaw
		l.plain, l.conditional = plain, conditional
	}
	l.mu.Unlock()

	var out console.OrdersReading

	page, err := plain(ctx, official.OrdersFilter{Limit: consoleOrdersPageLimit})
	if err != nil {
		out.PlainError = err.Error()
	} else {
		out.PlainTruncated = page.HasNext
		out.Plain = make([]console.OrderRecord, 0, len(page.Orders))
		for _, o := range page.Orders {
			out.Plain = append(out.Plain, console.OrderRecord{
				ID: o.ID, Symbol: o.Symbol, Side: o.Side, Kind: o.OrderType, Status: o.Status,
				Market: o.Market, Currency: o.Currency,
				Quantity: o.Quantity, Price: o.Price,
				FilledQuantity: o.FilledQuantity, AverageFilledPrice: o.AverageFilledPrice,
				OrderedAt: o.OrderedAt, CanceledAt: o.CanceledAt,
			})
		}
	}

	// The OPEN group, not a per-order status: openapi says the two vocabularies
	// differ, and this filter's job is to fetch exactly the conditionals that are
	// still watching — the ones filling the live-exposure cap.
	conds, err := conditional(ctx, verifylive.ConditionalStatusOpen, "", "", consoleOrdersPageLimit)
	if err != nil {
		out.ConditionalError = err.Error()
	} else {
		out.ConditionalTruncated = conds.HasNext
		out.Conditional = make([]console.ConditionalRecord, 0, len(conds.Orders))
		for _, o := range conds.Orders {
			out.Conditional = append(out.Conditional, console.ConditionalRecord{
				ID: o.ID, Symbol: o.Symbol, Market: o.Market, Kind: o.Type, Status: o.Status,
				Quantity: o.Quantity, TriggerPrice: o.TriggerPrice, OrderPrice: o.OrderPrice,
				ConditionKind: o.ConditionType, Triggered: o.TriggeredOrderID,
				ExpireDate: o.ExpireDate, CreatedAt: o.CreatedAt,
			})
		}
	}

	return out, nil
}

// consoleHandoffPath is where the single-use restart token lives: beside the
// evidence record and the soak pause marker, so an isolated --config-dir profile
// gets its own and the default sits in the data directory with the rest of this
// change's state.
func consoleHandoffPath(recordPath string) string {
	return filepath.Join(filepath.Dir(recordPath), handoff.FileName)
}

// consoleRelaunch re-executes this binary so a NEW process instance starts.
//
// That is the whole product: internal/verifylive's conditional-persistence step
// refuses to certify a conditional from the process that registered it, and it
// judges that on process.instance_id — minted fresh at every startup — rather than
// on the PID, which an exec preserves. The restart is therefore a real process
// boundary for the measurement, and the record proves it.
//
// The port is pinned so the browser comes back to the address it is already on.
// Everything else about the command line is kept: the same subcommand, the same
// --config-dir, the same flags the operator typed.
func consoleRelaunch(out io.Writer) console.Relaunch {
	return func(port int) error {
		path, err := binstamp.SelfPath()
		if err != nil {
			return err
		}
		argv := argvWithPort(os.Args, port)
		fmt.Fprintf(out, "  %s\n\n", strings.Join(argv, " "))
		return reexecSelf(path, argv)
	}
}

// consolePortFlag is the flag argvWithPort rewrites. It is named once.
const consolePortFlag = "--port"

// argvWithPort preserves the command line and pins the loopback port.
//
// A console started without --port took whatever the OS offered, and the browser is
// sitting on it. Re-executing with the original argument list would take another
// free port and strand the tab, so the port this process ended up on is written in
// explicitly — which is the one thing about the restart the operator did not choose
// and should not have to.
func argvWithPort(args []string, port int) []string {
	out := make([]string, 0, len(args)+2)
	skipNext := false
	for i, a := range args {
		switch {
		case skipNext:
			skipNext = false
		case i == 0:
			out = append(out, a)
		case a == consolePortFlag:
			skipNext = true // and its value
		case strings.HasPrefix(a, consolePortFlag+"="):
		default:
			out = append(out, a)
		}
	}
	return append(out, consolePortFlag, strconv.Itoa(port))
}

// consoleVerifyStarter builds the runner the console drives.
//
// It is `runVerifyRun`'s wiring with two differences and no others: the batch
// confirmer is the console's instead of the terminal's, and the run is always a
// resumption. The second follows from the first — a console has no --resume flag
// to forget, and the runner already refuses to re-measure a settled step (it skips
// it, and the plan excludes it with the reason on the page), so continuing the
// record is both the safe default and the only sensible one.
//
// redo is `verify run --redo`'s field reached from a button instead of a flag
// (task 1.7). The console computes the set from the evidence record — never from
// the request — and it changes only which steps the runner will walk: the plan is
// rebuilt and a new expiring string still has to be typed before anything is sent.
func consoleVerifyStarter(root *rootOptions) console.StartVerify {
	return func(
		ctx context.Context,
		confirm verifylive.BatchConfirmer,
		out io.Writer,
		market string,
		redo []verifylive.StepID,
	) (verifylive.Summary, []verifylive.Entry, error) {
		var empty verifylive.Summary

		market = verifylive.NormalizeMarket(market)
		// The record is resolved per run rather than captured once, because it is
		// the market that decides which file this run's verdicts belong in.
		recordPath, err := resolveVerifyRecordFor(root, "", market)
		if err != nil {
			return empty, nil, err
		}
		prior, err := verifylive.LoadEntries(recordPath)
		if err != nil {
			return empty, nil, err
		}
		broker, accountRef, err := verifyBrokerFactory(root)
		if err != nil {
			return empty, nil, err
		}

		recorder, err := verifylive.OpenRecorder(recordPath)
		if err != nil {
			return empty, nil, err
		}
		defer recorder.Close()

		runner, err := verifylive.New(verifylive.Options{
			Broker:          broker,
			Recorder:        recorder,
			Confirm:         consoleMutationConfirmer(),
			ConfirmBatch:    confirm,
			ApprovalChannel: verifylive.ApprovalChannelConsoleClick,
			Out:             out,
			AccountRef:      accountRef,
			Market:          market,
			Symbol:          consoleProbeSymbol(ctx, broker, market),
			HoldingSymbol:   firstUsableHoldingIn(ctx, broker, market),
			Offset:          verifylive.DefaultOffset,
			MaxSellQuantity: verifylive.DefaultMaxSellQuantity,
			TTLWait:         verifylive.DefaultTTLWait,
			Redo:            redo,
			Prior:           prior,
		})
		if err != nil {
			return empty, nil, err
		}

		fmt.Fprintf(out, "evidence record  %s\n", recordPath)
		releaseLock := holdVerifyRunLock(ctx, out, recordPath)
		defer releaseLock()

		summary, runErr := runner.Run(ctx)
		if runErr != nil && (errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded)) {
			// Ctrl-C on the console. Everything recorded stands, and the summary
			// already names whatever is still live.
			runErr = nil
		}
		return summary, runner.Entries(), runErr
	}
}

// consoleMutationConfirmer is the per-mutation gate the console does not offer.
//
// verifylive.Options requires a non-nil Confirmer and refuses a nil one rather
// than defaulting to something permissive; this is the value that satisfies it.
// The console never sets ConfirmEach, so nothing calls it — and if a later edit
// ever did, it refuses, which is the direction a mistake here has to fail in.
func consoleMutationConfirmer() verifylive.Confirmer {
	return func(verifylive.Mutation) error { return verifylive.ErrNotATerminal }
}
