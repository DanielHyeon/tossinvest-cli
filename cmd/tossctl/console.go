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
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
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
  signals     /signals — what the discovery sources have been saying over time,
              with what nobody has checked spelled out rather than left blank. It
              reads the discovery store and calls no source
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

	// The console's one live client, shared by every seam below that reads through
	// it. Building it resolves the account against the Open API, and that read is
	// the call which came back 429 three times on 2026-07-26 and cost a
	// verification run three steps (measurements.md M4) — so it happens once per
	// console process, not once per screen the operator opens. It is still lazy:
	// nothing is built until a screen asks.
	reads := newConsoleBroker(root)

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
		Holdings: newConsoleHoldings(reads),
		// The orders screen's read (change console-orders-screen). One method,
		// behind which this file makes the three calls a refresh costs.
		Orders:      consoleOrdersSeam(reads),
		JournalPath: journalPath,
		RunLockPath: verifyRunLockPath(verifyRecord),
		Settings:    consoleSettingsSeam(root),

		// The overview's read-only view of the Guardian's ceilings (change
		// console-operator-overview). Five numbers and a currency: this seam
		// displays what the file says and can change none of it.
		GateLimits: consoleGateLimitsSeam(root),

		// The settings screen's editor for those same ceilings (change
		// console-sets-guardian-limits). Separate from the reader above, and its
		// Save takes five numbers and a currency — there is no field on the wire
		// for `enabled`, so the console still cannot open the automation gate.
		Limits: consoleLimitSettingsSeam(root),

		// The discovery screen's read (change add-candidate-discovery, task 5.5).
		// It opens internal/candidate's store, runs its assessment and hands over
		// values — no source is called, so an open /signals tab spends none of the
		// account's rate budget and does not become a second discoverer.
		Signals: consoleSignalsSeam(root),

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

// consoleLimitSettingsSeam adapts the Guardian-limit editor (limitsettings.go).
// Same nil-on-the-concrete-pointer care as above: a typed-nil inside the
// interface would look wired and the screen would offer controls that panic
// instead of explaining themselves.
func consoleLimitSettingsSeam(root *rootOptions) console.LimitSettings {
	if s := newLimitSettingsSeam(root); s != nil {
		return s
	}
	return nil
}

// consoleGateLimitsSeam hands the overview screen the Guardian's ceilings, and
// nothing else (change console-operator-overview task 5.1).
//
// It reads, and it stays a reader. The writer is a separate seam
// (consoleLimitSettingsSeam, limitsettings.go) because the overview must not
// gain the ability to change a number it exists to display.
//
// Turning the automation gate ON remains a §0.7 human decision taken outside any
// browser and neither seam can do it; console-sets-guardian-limits separated the
// ceilings from the switch rather than opening both. What crosses this boundary
// is five float64s and a currency string — not the config service.
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

// --- the console's one account resolution -----------------------------------------

// consoleBroker is the live client every read seam on this console shares.
//
// Building that client resolves the account against the Open API: one
// /api/v1/accounts read inside buildVerifyBroker, which is the call that came back
// 429 three times on 2026-07-26 and cost a verification run three steps
// (measurements.md M4). Each seam used to build its own, so a session that opened
// the positions screen and then /orders paid for that resolution twice — per
// process, not per refresh, but twice. This type is where the "once" lives.
//
// It widens nothing. A seam is handed this resolver and nothing else, and what
// crosses into internal/console is still one bound method value per screen; the
// client those method values come from was always reachable from this file, since
// a bound method pins its receiver, and it is reachable from the console neither
// before nor after.
type consoleBroker struct {
	root *rootOptions

	mu     sync.Mutex
	client verifylive.Broker
}

// newConsoleBroker holds the console's live client, and builds nothing yet.
//
// Lazily because that build resolves the account, and doing it at `tossctl
// console` startup would make a screen the operator may never open into a
// precondition for the console coming up at all. The first render pays for it; a
// failure is a sentence on the page.
func newConsoleBroker(root *rootOptions) *consoleBroker {
	return &consoleBroker{root: root}
}

// resolve returns the console's live client, building it on the first call.
//
// The build happens under the lock, so two screens opened in the same second make
// one resolution and the second waits for it instead of starting a second
// /api/v1/accounts read. Waiting is the point: what is being serialised is exactly
// the call whose rate limit costs a run its steps.
//
// A failure is returned rather than remembered: credentials that were not there
// when the console started may be there after `tossctl openapi login`, and what
// stops a failing build from being retried on every render is the console's own
// cache (holdings.go bounds attempts by the TTL, not by successes).
func (c *consoleBroker) resolve() (verifylive.Broker, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		return c.client, nil
	}
	broker, _, err := verifyBrokerFactory(c.root)
	if err != nil {
		return nil, err
	}
	c.client = broker
	return c.client, nil
}

// newConsoleHoldings builds the console's holdings reader.
//
// It is handed the shared resolver rather than the root options, which is what
// makes the account resolution behind the positions screen the same one the orders
// screen uses. The laziness is unchanged — it now lives in consoleBroker.resolve.
func newConsoleHoldings(shared *consoleBroker) console.HoldingsReader {
	return &lazyHoldings{shared: shared}
}

// holdingsFunc is the single capability the console is handed.
type holdingsFunc func(ctx context.Context, symbol string) ([]domain.Position, error)

type lazyHoldings struct {
	shared *consoleBroker
}

// Holdings satisfies console.HoldingsReader.
//
// The method value is taken off the shared client per call and dropped again. This
// type holds no broker and no field at all beyond the resolver, so the Open API
// client's PlaceOrder / CancelOrder / ModifyOrder are not reachable from the value
// that crosses into internal/console.
func (l *lazyHoldings) Holdings(ctx context.Context, symbol string) ([]domain.Position, error) {
	broker, err := l.shared.resolve()
	if err != nil {
		return nil, err
	}
	var read holdingsFunc = broker.Holdings
	return read(ctx, symbol)
}

// --- the orders screen's read (change console-orders-screen, task 4.6) ------------
//
// The console declares one method and this is the only thing in the binary that
// satisfies it. Behind that one method are the three broker calls one refresh
// costs, which is where they belong: the console cannot then spend the budget
// three times over, and it cannot report one endpoint's silence as another's zero.

// The two values the `status` query parameter of GET /api/v1/orders takes. It is
// documented `required: true`, and the parameter is not a filter in the way the
// other five are — it selects between two differently-shaped answers:
//
//	OPEN    "모든 대기 중 주문을 전량 반환합니다. limit, cursor 는 무시되며"
//	CLOSED  "limit (기본 20, 최대 100), cursor, from/to 파라미터 모두 적용됩니다"
//
// The first implementation of this screen sent neither and asked for limit=100,
// which is either a rejected request (the live count permanently unmeasured) or
// one page over the account's entire history (a live order past row 100 missing
// from the table AND from the count, rendered as "0건 이상"). Both are the failure
// the screen exists to prevent. They are constants rather than literals so the
// call sites read as a choice between two groups and not as a filter somebody may
// tidy away.
const (
	orderGroupOpen   = "OPEN"
	orderGroupClosed = "CLOSED"
)

// consoleOrdersPageLimit bounds one page of the CLOSED list.
//
// It is a page and not the whole history on purpose — the screen is "what is
// alive and what happened today", and an unbounded walk would be a loop of broker
// calls behind one page load. When the broker says there is more, the console
// renders that count as a floor ("N건 이상") rather than as a number, which is why
// truncating here is honest rather than lossy.
//
// It is deliberately NOT sent with the OPEN request. The API ignores it there, so
// sending it would put a bound on the wire that the answer does not have — and the
// live count's whole claim is that it cannot be short.
const consoleOrdersPageLimit = 100

// consoleOrdersReader is the part of the live client the orders screen needs.
//
// verifyBrokerFactory returns a verifylive.Broker, which does not declare the
// raw order reads — it declares OrdersPageRaw, whose orders are undecoded JSON.
// The concrete client has both, so the path is chosen by asserting for exactly
// the two reads on the client the console has already resolved, rather than by
// building a second *official.Client. That second client would resolve the account
// sequence again, and that resolution is the call that came back 429 three times
// on 2026-07-26 and cost a run three steps (measurements.md M4) — the same reason
// this seam takes consoleBroker rather than rootOptions.
type consoleOrdersReader interface {
	OrdersRaw(ctx context.Context, filter official.OrdersFilter) (official.RawOrderList, error)
	ConditionalOrdersRaw(ctx context.Context, status, symbol, cursor string,
		limit int) (official.RawConditionalOrderList, error)
}

// consoleOrdersSeam builds the console's orders reader.
//
// It is handed the same shared resolver newConsoleHoldings gets, so opening this
// screen after the positions screen costs no second account resolution. The
// laziness is unchanged and now lives in one place — consoleBroker.resolve.
func consoleOrdersSeam(shared *consoleBroker) console.OrdersReader {
	return &lazyOrders{shared: shared}
}

type lazyOrders struct {
	shared *consoleBroker
}

// Orders satisfies console.OrdersReader.
//
// The two read method values are taken off the shared client per call and dropped
// again: this type holds no broker, so nothing reachable from the value the
// console was handed has PlaceOrder on it.
//
// The three calls are sequential and each one's outcome is carried separately. A
// failure on one is NOT returned as an error for the set: the console renders that
// part as unmeasured and refuses to add the counts, and collapsing them into one
// error would take the measured parts down with the missing one.
//
// An error is returned only when there is no reading at all to describe — no
// credentials, or a client that does not have these reads.
func (l *lazyOrders) Orders(ctx context.Context) (console.OrdersReading, error) {
	broker, err := l.shared.resolve()
	if err != nil {
		return console.OrdersReading{}, err
	}
	reads, ok := broker.(consoleOrdersReader)
	if !ok {
		return console.OrdersReading{}, fmt.Errorf(
			"console: this build's broker (%T) has no raw order reads; the orders screen needs "+
				"the broker's own decimal strings, because a value that has been through float64 "+
				"cannot say whether the broker sent one at all", broker)
	}
	plain, conditional := reads.OrdersRaw, reads.ConditionalOrdersRaw

	var out console.OrdersReading

	// The live half. status=OPEN and nothing else: the broker returns every
	// pending order for it, so this list cannot be short and the count taken from
	// it is a number rather than a floor. That is the only call on this screen
	// which structurally cannot miss a leftover, and missing a leftover is what
	// the screen is for.
	open, err := plain(ctx, official.OrdersFilter{Status: orderGroupOpen})
	if err != nil {
		out.OpenError = err.Error()
	} else {
		out.OpenTruncated = open.HasNext
		out.Open = consoleOrderRecords(open.Orders)
	}

	// The finished half. It is a call of its own rather than a saving, because a
	// cancelled or a rejected order never becomes a trade — so /history, which is
	// built from round trips, never shows it. This is the only place that
	// information exists at all.
	closed, err := plain(ctx, official.OrdersFilter{
		Status: orderGroupClosed,
		Limit:  consoleOrdersPageLimit,
	})
	if err != nil {
		out.ClosedError = err.Error()
	} else {
		out.ClosedTruncated = closed.HasNext
		out.Closed = consoleOrderRecords(closed.Orders)
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

// consoleOrderRecords carries one page of raw orders across the console boundary.
//
// Every value stays the string the broker sent. parseDecimal anywhere on this path
// would put back the zero the whole raw read exists to keep out: the API sends
// price as null for a market order and the entire execution object as null while
// an order is alive, so a converted reading says every live order filled at
// nothing.
func consoleOrderRecords(orders []official.RawOrder) []console.OrderRecord {
	out := make([]console.OrderRecord, 0, len(orders))
	for _, o := range orders {
		out = append(out, console.OrderRecord{
			ID: o.ID, Symbol: o.Symbol, Side: o.Side, Kind: o.OrderType, Status: o.Status,
			Market: o.Market, Currency: o.Currency,
			Quantity: o.Quantity, Price: o.Price,
			FilledQuantity: o.FilledQuantity, AverageFilledPrice: o.AverageFilledPrice,
			OrderedAt: o.OrderedAt, CanceledAt: o.CanceledAt,
		})
	}
	return out
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
//
// This is the one place in the console that does NOT take the shared read client,
// and that is deliberate. It builds its own through verifyBrokerFactory on every
// run, which costs one more /api/v1/accounts read than reusing the console's,
// because:
//
//   - The run is about to place real orders, and the account it names in the
//     evidence record is the one it resolved at run start. Reusing a resolution
//     performed whenever the operator happened to open a read screen — possibly
//     hours earlier, possibly before `tossctl openapi login` replaced the
//     credentials — would produce a record naming an account that this run never
//     confirmed. An unconfirmed account on a record of real orders is worse than
//     one extra read.
//   - The resolution's failure contract differs. A read screen renders it as a
//     sentence and stays up; here it is fatal before a person is asked anything,
//     under verifylive's read retry policy (resolveVerifyAccount).
//   - Reads and a verification are different trust contexts. The read seams are
//     handed a client stripped to bound method values; the runner needs the whole
//     broker, and that asymmetry should not be resolved by giving the read path's
//     client the wider job.
//
// The cost is bounded: one resolution per run started, not per screen and not per
// refresh.
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

// --- the discovery screen (change add-candidate-discovery, task 5.5) ---------------
//
// The seam below hands /signals a read of the discovery store, and nothing else.
// It lives in this file because console_test.go's
// TestOnlyConsoleGoReachesTheConsolePackage allows exactly one importer of
// internal/console: a second one would be a second place the web confirmer could
// be wired, and reviewing one file is the only way that claim stays checkable.
//
// # It reads and it never scans
//
// candidate.Assess loads what is already stored and calls no source. That is the
// whole arrangement: spec Requirement 7 ranks the account's rate budget engine >
// verification > discovery, and a browser tab reloading every fifteen seconds into
// a second discoverer would compete with `tossctl candidate watch` for the
// undocumented RANKING limit — where one 429 costs the entire source, because the
// rankings have no WTS fallback (D14 decision 2). It would also write a second
// near-duplicate observation per symbol per tick into the series the acceleration
// is differenced from.
//
// # And it opens the store per read rather than holding it
//
// Store.Checkpoint's own documentation names this screen as the long-lived reader
// that stops `wal_checkpoint(TRUNCATE)` from reclaiming anything — and the write-
// ahead log grows beside the table on the filesystem the order ledger writes to
// (D16). A console that held the store open would quietly disable the cleanup the
// watch loop runs every turn. Opening for the length of one render costs a file
// open every refresh and keeps that sweep working.

// consoleSignalsMarkets is what the screen reports on, in the contract's order
// (design.md 결정된 계약값: markets.KR and markets.US, both enabled from the start).
var consoleSignalsMarkets = []string{candidate.MarketKR, candidate.MarketUS}

// consoleSignalsPanelWhy is why the screen cannot name the sources a scan lost.
//
// ScanResult.Missing is the only thing that carries a source id together with the
// reason it did not answer, and it lives for the length of one cycle in the
// process that ran it. The console is not that process and the store persists no
// scan record, so this reading has no names to give — and saying so is the whole
// of the rule task 5.7 states: a degradation nobody can attribute is a display
// nobody can act on, and inventing an attribution would be worse than admitting
// the gap. See openspec/changes/add-candidate-discovery/issues.md.
//
// What IS on the screen is the store's own per-candidate completeness — attempted,
// answered and degraded, as recorded by the last scan that touched each row — so
// the page is not silent about degradation, only about which source caused it.
const consoleSignalsPanelWhy = "이 콘솔 프로세스는 스캔을 돌리지 않는다. " +
	"빠진 원천의 이름과 사유는 스캔 결과에만 있고 저장소에 남지 않으므로, " +
	"`tossctl candidate scan`이나 `tossctl candidate watch`의 출력에서 확인한다. " +
	"아래 후보별 완전성은 저장소가 기록한 마지막 스캔의 값이다."

// consoleSignals is the seam's implementation.
type consoleSignals struct {
	// open resolves the discovery store and hands back the release for it. It is a
	// field rather than a direct call so this package's tests can point it at a
	// temporary one.
	//
	// It takes the caller's context because Open migrates: the schema check, the
	// ladder behind it and every query afterwards belong to the request that asked
	// for the page, and this page reloads itself every fifteen seconds. A request
	// the browser has already abandoned used to keep all of it running under
	// context.Background().
	open func(ctx context.Context) (*candidate.Store, func(), error)
}

// consoleSignalsSeam wires the discovery screen.
//
// What crosses the boundary is a value: verdicts, tallies and a panel report.
// internal/candidate's Store — which can promote, cool and prune — never becomes
// reachable from internal/console, so "the discovery screen changes nothing" is a
// fact about the wiring rather than about the handlers.
func consoleSignalsSeam(root *rootOptions) console.SignalsReader {
	return &consoleSignals{
		open: func(ctx context.Context) (*candidate.Store, func(), error) {
			return candidateStoreFactory(ctx, root)
		},
	}
}

// Signals reads every contract market's assessment as of one instant.
func (s *consoleSignals) Signals(ctx context.Context) (console.SignalsReading, error) {
	store, release, err := s.open(ctx)
	if err != nil {
		return console.SignalsReading{}, fmt.Errorf("candidate: opening the discovery store: %w", err)
	}
	defer release()

	// One instant for every market, taken from the store's clock rather than this
	// process's: every input age on the page is measured against it, and two
	// markets read at two instants would answer the same question differently.
	at := store.Now()
	out := console.SignalsReading{At: at}
	for _, market := range consoleSignalsMarkets {
		out.Markets = append(out.Markets, consoleSignalsMarket(ctx, store, market, at))
	}
	return out, nil
}

// consoleSignalsMarket assesses one market.
//
// A market that could not be read comes back present with Why set rather than
// omitted: a market missing from the page is indistinguishable from a market with
// nothing in it, which is the confusion this whole screen exists to remove, one
// level up from the veto.
func consoleSignalsMarket(ctx context.Context, store *candidate.Store, market string,
	at time.Time) console.SignalsMarket {

	out := console.SignalsMarket{
		Market: market,
		Panel:  console.SignalsPanel{Why: consoleSignalsPanelWhy},
	}
	verdicts, err := candidate.Assess(ctx, store, candidate.AssessOptions{
		Market: market,
		At:     at,
		// The same thresholds `tossctl candidate scan` applies, from the same
		// function. This used to be its own literal, and a literal on each side of
		// the seam is how the two surfaces come to disagree about the same store at
		// the same instant — each one internally consistent, neither one failing.
		// See vetothresholds.go.
		Thresholds: candidateVetoThresholds(),
	})
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Verdicts = verdicts

	// The same reducer the scan output uses, rather than the same three
	// constructors called again here. This block used to assemble the tallies by
	// hand: the two agreed, and the whole cost sat in the future — a fourth shadow
	// band added to internal/candidate would have appeared on `tossctl candidate
	// scan` and not on /signals, and a band missing from a page renders as a band
	// nobody crossed rather than as one nobody counted.
	out.Vetoes, out.Crossings, out.Bands = candidate.TallyVerdicts(verdicts)
	// Same reducer as the scan output's, for the same reason as the line above it.
	out.Sightings = candidate.TallySightingSources(verdicts)
	return out
}
