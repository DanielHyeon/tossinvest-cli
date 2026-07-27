package engine

// runtime_wiring.go is the production assembly the runtime needs and nothing had
// a constructor for (openspec change add-engine-runtime, tasks 1.1 and 1.4).
//
// Three things live here:
//
//	AccountSweep     the holdings/buying-power reader both the reconciliation
//	                 snapshot and the fill detector's account sweep take
//	SnapshotCollector / Recovery
//	                 the two consumers of that sweep, wired from the engine's own
//	                 journal, resolver, gateway and entry gate
//	UnmetInterlockClauses
//	                 the startup refusal, rendered as the enumerated list the
//	                 command prints
//
// # Why the refusal is rendered and not re-derived
//
// engine-safety asks the runtime to refuse "인터록이 반환한 미충족 항목을
// 열거하며". The interlock stops at its first failing clause (verifyGate: "It
// stops at the first failure"), so the unmet set *is* what the returned error
// says — and this function reads it rather than re-running any check. A second
// evaluation here would be a second opinion about whether the gate may open,
// which is precisely the thing one interlock exists to be.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// DefaultLimitCurrency is the currency the account snapshot reads a balance in
// when the automation gate names none.
//
// The gate's own LimitCurrency is authoritative and the interlock refuses a gate
// without one, so this is reached only by a gate-off engine — where it decides
// which single balance read a snapshot makes and nothing else.
const DefaultLimitCurrency = "KRW"

// AccountSweep is the engine's account read, in the two shapes its consumers
// declare.
//
// One type rather than two adapters because it is one endpoint: reconciliation's
// snapshot and the fill detector's positions sweep both mean `GET
// /api/v1/holdings`, and two adapters would invite two calls out of the §0.4
// budget for one fact.
type AccountSweep struct{ reads OfficialReads }

// Positions satisfies reconcile.PositionsReader and filldetect.PositionsReader.
// An empty symbol filter is the whole-account sweep.
func (a AccountSweep) Positions(ctx context.Context) ([]domain.Position, error) {
	return a.reads.Holdings(ctx, "")
}

// PositionsRaw satisfies reconcile.RawPositionsReader, which is the path the
// collector prefers: it is what preserves the broker's own cost-basis decimal
// for the adoption record.
func (a AccountSweep) PositionsRaw(ctx context.Context) ([]reconcile.RawHolding, error) {
	items, err := a.reads.HoldingsRaw(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]reconcile.RawHolding, 0, len(items))
	for _, h := range items {
		out = append(out, reconcile.RawHolding{
			Symbol:       h.Symbol,
			Market:       h.MarketCountry,
			Quantity:     h.Quantity,
			AveragePrice: h.AveragePurchasePrice,
		})
	}
	return out, nil
}

// BuyingPower satisfies reconcile.BuyingPowerReader and
// filldetect.BuyingPowerReader.
func (a AccountSweep) BuyingPower(ctx context.Context, currency string) (float64, error) {
	bp, err := a.reads.BuyingPower(ctx, currency)
	if err != nil {
		return 0, err
	}
	return bp.CashBuyingPower, nil
}

// Compile-time proof that the sweep satisfies what its consumers declare.
var (
	_ reconcile.PositionsReader    = AccountSweep{}
	_ reconcile.RawPositionsReader = AccountSweep{}
	_ reconcile.BuyingPowerReader  = AccountSweep{}
)

// AccountSweep returns the account reader for this engine.
func (c *Context) AccountSweep() AccountSweep { return AccountSweep{reads: c.Official} }

// SnapshotCurrencies is the balance set an account snapshot reads.
//
// One currency: the one the gate's limits are expressed in. A snapshot that read
// every currency the account could hold would spend the rate budget on numbers no
// limit is compared against.
func (c *Context) SnapshotCurrencies() []string {
	currency := strings.ToUpper(strings.TrimSpace(c.Config.Engine.AutomationGate.LimitCurrency))
	if currency == "" {
		currency = DefaultLimitCurrency
	}
	return []string{currency}
}

// SnapshotCollector builds the account snapshot reader.
//
// It is shared by the reconciliation driver and the restart recovery on purpose:
// the two compare the account against the same projection, and a second collector
// with different pagination or a different currency set would make "the account"
// mean two things inside one process.
//
// clk is the caller's, defaulting to the engine's. It is a parameter because the
// snapshot's as-of stamp is what stabilisation measures its interval against: a
// collector on one clock driven by a loop on another would offer two snapshots
// that look simultaneous and never stabilise.
func (c *Context) SnapshotCollector(clk clock.Clock) *reconcile.Collector {
	if clk == nil {
		clk = c.clock()
	}
	return &reconcile.Collector{
		Orders:     execgw.OfficialOrders{Client: c.Official},
		Positions:  c.AccountSweep(),
		Balance:    c.AccountSweep(),
		Currencies: c.SnapshotCurrencies(),
		AccountRef: c.AccountRef,
		Clock:      clk,
	}
}

// ErrRecoveryUnavailable means the restart recovery could not be assembled from
// this engine's wiring.
var ErrRecoveryUnavailable = errors.New("engine: the restart recovery sequence could not be assembled")

// Recovery assembles the landed restart recovery sequence (task 1.4).
//
// # No new recovery code
//
// Everything this sequence does is internal/reconcile's and internal/journal's:
// RecoverPending applies the RECORDED/DISPATCH_STARTED rules, the replay entry
// point recovers an identity from the broker's idempotent answer, the resolver
// settles what is left, and the account is re-read until it stops moving. This
// function supplies the four handles and nothing else — engine-safety: 재기동
// 복구는 landed 계약을 소비하며 새 복구 경로를 만들지 않는다(SHALL NOT).
//
// The adoption half of the recovery is not here and does not need to be: an
// adoption that committed before a crash is completed by the exit observer's next
// cycle, from the record already on disk (exitloop.go openAdoptedState). Starting
// the loops IS the recovery for that one.
//
// # The gate is shut from here on
//
// reconcile.New latches the entry gate in its constructor. A caller that builds
// this and never runs it therefore gets an engine that refuses to open positions,
// which is the failure mode everybody can live with.
//
// # The options
//
// The signature takes options for the reason [Context.ExitObserver] and
// [Context.ReconcileDriver] do: the tuning (the stabilisation bounds, the clock) is
// the caller's, and the handles that decide what the sequence is allowed to touch
// are not — every one of them is overwritten here, so a caller cannot point the
// recovery at a journal or a gate the engine did not assemble.
func (c *Context) Recovery(opts reconcile.Options) (*reconcile.Recovery, error) {
	if c == nil || c.Journal == nil || c.Resolver == nil || c.Entry == nil {
		return nil, fmt.Errorf("%w: the engine wiring is incomplete", ErrRecoveryUnavailable)
	}
	if opts.Clock == nil {
		opts.Clock = c.clock()
	}
	opts.Journal = c.Journal
	opts.Resolver = c.Resolver
	opts.Replayer = c.Gateway
	opts.Collector = c.SnapshotCollector(opts.Clock)
	opts.Gate = c.Entry
	opts.AccountRef = c.AccountRef
	rec, err := reconcile.New(opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRecoveryUnavailable, err)
	}
	return rec, nil
}

// --- the startup refusal, enumerated ---------------------------------------------

// interlockClauses is the enumeration, in the order verifyGate checks it.
//
// It is a table so the rendering cannot drift into naming a clause the interlock
// does not have: every sentinel here is one the interlock wraps, and a clause
// added there with no line here falls through to the raw message rather than
// being silently dropped.
var interlockClauses = []struct {
	sentinel error
	clause   string
}{
	{ErrGuardianRequired, "1. Guardian 주입 — 주문을 승인할 위험 권한이 없다"},
	{ErrLimitsRequired, "2. 위험 한도 전 항목 — engine.automation_gate의 한도가 완전하지 않다"},
	{ErrTradingPolicyRefused, "3. 거래 정책 — 매도·라이브 주문 동작이 모두 허용되어야 한다"},
	{attest.ErrMissing, "4. capability attestation — 파일이 없다"},
	{attest.ErrMalformed, "4. capability attestation — 파일을 해석할 수 없다"},
	{attest.ErrFormatTooNew, "4. capability attestation — 이 빌드보다 새로운 형식이다"},
	{attest.ErrIncomplete, "4. capability attestation — 내용이 불완전하다"},
	{attest.ErrExpired, "4. capability attestation — 만료되었다"},
	{attest.ErrAccountMismatch, "4. capability attestation — 다른 계좌의 것이다"},
	{attest.ErrEndpointNotAttested, "4. capability attestation — 엔진이 호출하는 endpoint를 다 덮지 못한다"},
	{ErrAccountUnresolved, "5. 계좌 확인 — 자격증명이 어느 계좌인지 해석되지 않았다"},
	{ErrGuardianLimitsMismatch, "6. 한도 단일 출처 — Guardian이 감사된 한도와 다른 값으로 승인한다"},
	{ErrGatewayRequired, "7. ExecutionGateway — 완전히 배선된 게이트웨이가 없다"},
	{ErrKeylessTransport, "8. idempotency key — 주문 전송 경로가 키를 실을 수 없다"},
	{ErrProtectionNotWired, "9. 브로커측 보호 실행(ProtectionReady) — 이 빌드에는 없다 [2c 소관]"},
}

// UnmetInterlockClauses renders a startup refusal as the enumerated list of
// clauses the interlock reported unmet.
//
// It returns nil for anything that is not an interlock refusal, which is how the
// caller tells "the gate said no" from "the journal would not open".
//
// The list is usually one line long, and that is the interlock's design rather
// than a limitation of this function: verifyGate stops at its first failing
// clause because the clauses are ordered cheapest-first, and reporting all of
// them would mean reading a file and interrogating a Guardian to tell an operator
// about a mistake they can see in their config. The raw message is always
// appended, because it carries the specific field or endpoint.
func UnmetInterlockClauses(err error) []string {
	if err == nil || !errors.Is(err, ErrAutomationGateRefused) {
		return nil
	}
	var out []string
	for _, c := range interlockClauses {
		if errors.Is(err, c.sentinel) {
			out = append(out, c.clause)
		}
	}
	if len(out) == 0 {
		out = append(out, "미상 조항 — 인터록이 이름 없는 사유로 거부했다")
	}
	return append(out, "상세: "+err.Error())
}

// clock is the engine's clock, defaulting to the system one. It is a method
// rather than a bare field read so a Context built by a test that did not set it
// behaves the same as one that did.
func (c *Context) clock() clock.Clock {
	if c.clk != nil {
		return c.clk
	}
	return clock.System()
}

// Compile-time proof that the raw holdings read the sweep needs is on the
// concrete client. A signature drift fails the build here.
var _ interface {
	HoldingsRaw(ctx context.Context, symbol string) ([]official.RawHolding, error)
} = (*official.Client)(nil)
