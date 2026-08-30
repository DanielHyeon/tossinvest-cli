package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	candidatepkg "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

// StrategyEntryLoopName is the single outer runtime loop which owns both market
// evaluation children. Market-local failures stay inside this loop so they do
// not cancel the engine's safety-class loops.
const StrategyEntryLoopName = "strategy-entry"

const (
	DefaultStrategyQueueDepth     = 1
	MaximumStrategyQueueDepth     = 1024
	DefaultStrategyCycleLimit     = 5 * time.Second
	MaximumStrategyCycleLimit     = 30 * time.Second
	MinimumStrategyPollInterval   = time.Second
	DefaultStrategyRestartStep    = 5 * time.Second
	MaximumStrategyRestartBackoff = 30 * time.Second
)

// StrategyMarket is deliberately a closed, per-market enumeration. There is no
// combined KR+US value because calendar, activation, evidence and budget
// authority are independently scoped.
type StrategyMarket string

const (
	StrategyMarketKR StrategyMarket = "KR"
	StrategyMarketUS StrategyMarket = "US"
)

func validStrategyMarket(market StrategyMarket) bool {
	return market == StrategyMarketKR || market == StrategyMarketUS
}

// StrategyCycle evaluates one immutable market snapshot. This leaf provides no
// Gateway, journal, activation writer, broker mutator, latch writer or recovery
// writer to the callback. A callback that ignores cancellation is abandoned at
// the watchdog boundary; its result is buffered and can no longer affect state.
type StrategyCycle func(context.Context) error

// StrategyMarketWorker is the construction-time evaluation authority for one
// child. Effective defaults to false. The exact generation, evidence and latch
// revision let the future durable owner reject stale or cross-market fault and
// recovery records without this leaf acquiring mutation capability.
type StrategyMarketWorker struct {
	Market              StrategyMarket
	Effective           bool
	Cycle               StrategyCycle
	AuthorityGeneration uint64
	AuthorityExpiresAt  time.Time
	EvidenceDigest      string
	LatchRevision       uint64
	FirstRefusal        StrategyWorkerRefusalCode
	RestartAttempt      uint64
	RestartNotBefore    time.Time
	PollInterval        time.Duration
	RefreshesAuthority  bool
}

// StrategyWorkerRefusalCode is the closed operational classification retained
// across market-local child restarts. It is evidence only and cannot restore
// entry authority.
type StrategyWorkerRefusalCode string

const (
	StrategyWorkerRefusalNone             StrategyWorkerRefusalCode = ""
	StrategyWorkerRefusalFailure          StrategyWorkerRefusalCode = "WORKER_FAILURE"
	StrategyWorkerRefusalAbnormal         StrategyWorkerRefusalCode = "WORKER_ABNORMAL"
	StrategyWorkerRefusalAuthorityExpired StrategyWorkerRefusalCode = "AUTHORITY_EXPIRED"
)

func validStrategyWorkerRefusal(code StrategyWorkerRefusalCode) bool {
	switch code {
	case StrategyWorkerRefusalNone, StrategyWorkerRefusalFailure, StrategyWorkerRefusalAbnormal, StrategyWorkerRefusalAuthorityExpired:
		return true
	default:
		return false
	}
}

// StrategyWorkerFault is the immutable handoff emitted when one market latches
// OFF. It is not a recovery receipt and grants no way to turn entry back ON.
type StrategyWorkerFault struct {
	Market              StrategyMarket
	LatchID             string
	ExpectedRevision    uint64
	NextRevision        uint64
	AuthorityGeneration uint64
	AuthorityExpiresAt  time.Time
	EvidenceDigest      string
	Reason              string
	FirstRefusal        StrategyWorkerRefusalCode
	Abnormal            bool
	ObservedAt          time.Time
	RestartAttempt      uint64
	RestartNotBefore    time.Time
}

// StrategyEntrySupervisorOptions wires the evaluation-only child supervisor.
// There is deliberately no latch or recovery callback: durable writers must be
// connected later through the fenced central owner, and no boolean result may
// restore in-memory entry authority.
type StrategyEntrySupervisorOptions struct {
	Workers    []StrategyMarketWorker
	QueueDepth int
	CycleLimit time.Duration
	Clock      clock.Clock
}

// ErrStrategyCentralIntegrity is the only non-cancellation failure the outer
// strategy-entry loop returns. The existing Runtime then performs its normal
// process-wide fail-closed cancellation and drain.
var ErrStrategyCentralIntegrity = errors.New("engine: central strategy integrity failure")

// ErrStrategyCycleDeadline is market-local. It latches only the affected market
// OFF and records that the evaluation callback crossed its watchdog boundary.
var ErrStrategyCycleDeadline = errors.New("engine: strategy evaluation deadline exceeded")

// ErrStrategyAuthorityExpired is market-local and is checked again in the
// market child immediately before invoking evaluation.
var ErrStrategyAuthorityExpired = errors.New("engine: strategy worker authority expired")

type centralStrategyIntegrityError struct{ cause error }

func (e centralStrategyIntegrityError) Error() string {
	if e.cause == nil {
		return ErrStrategyCentralIntegrity.Error()
	}
	return ErrStrategyCentralIntegrity.Error() + ": " + e.cause.Error()
}

func (e centralStrategyIntegrityError) Unwrap() error { return e.cause }

// StrategyCentralIntegrityFailure marks an evaluator-discovered invariant as
// process-wide. It grants no authority and only narrows behavior to fail-closed.
func StrategyCentralIntegrityFailure(cause error) error {
	if cause == nil {
		cause = ErrStrategyCentralIntegrity
	}
	return centralStrategyIntegrityError{cause: cause}
}

func isCentralStrategyIntegrity(err error) bool {
	var target centralStrategyIntegrityError
	return errors.As(err, &target) || errors.Is(err, ErrStrategyCentralIntegrity)
}

type strategyMarketRuntime struct {
	descriptor       StrategyMarketWorker
	queue            chan struct{}
	effective        bool
	latched          bool
	firstFailure     string
	firstRefusal     StrategyWorkerRefusalCode
	firstAbnormal    bool
	latchID          string
	latchRevision    uint64
	abandoned        bool
	restartAttempt   uint64
	restartNotBefore time.Time
}

// StrategyWorkerSnapshot is a read-only, market-keyed operational view. It
// carries no activation, release or mutation method.
type StrategyWorkerSnapshot struct {
	Market              StrategyMarket
	Effective           bool
	Latched             bool
	FirstFailure        string
	FirstRefusal        StrategyWorkerRefusalCode
	FirstAbnormal       bool
	LatchID             string
	LatchRevision       uint64
	AuthorityGeneration uint64
	AuthorityExpiresAt  time.Time
	EvidenceDigest      string
	AbandonedEvaluation bool
	RestartAttempt      uint64
	RestartNotBefore    time.Time
	QueueDepth          int
	QueueCapacity       int
}

type StrategyTriggerResult string

const (
	StrategyTriggerEnqueued StrategyTriggerResult = "ENQUEUED"
	StrategyTriggerDisabled StrategyTriggerResult = "DISABLED"
	StrategyTriggerFull     StrategyTriggerResult = "FULL"
	StrategyTriggerInvalid  StrategyTriggerResult = "INVALID_MARKET"
)

// StrategyEntrySupervisor owns exactly two independent market children and one
// outer SupervisedLoop.
type StrategyEntrySupervisor struct {
	clk        clock.Clock
	cycleLimit time.Duration

	mu        sync.RWMutex
	workers   map[StrategyMarket]*strategyMarketRuntime
	accepting bool
	run       bool
	ready     chan struct{}
	faults    chan StrategyWorkerFault
}

// StrategyEntryProductionSnapshot is the complete command-to-engine handoff
// for this dormant production checkpoint. These are observations, not entry
// authorities: schedule contains no opaque activation/calendar capability and
// there is deliberately no candidate reader, journal, Gateway, broker,
// Guardian, trigger or worker callback in the shape. Ordinary automation or
// schedule readiness can therefore never promote either market.
type StrategyEntryProductionSnapshot struct {
	Clock              clock.Clock
	AutomationVerified bool
	EntryPermitted     bool
	ProtectionWired    bool
	Schedule           PairedStrategyScheduleSnapshot
}

// StrategyEntryProductionAssembly is the read-only result returned across the
// command boundary. Schedule contains scalar verification observations only;
// the opaque signed activation and adapted calendar values never leave the
// engine package.
type StrategyEntryProductionAssembly struct {
	Supervisor *StrategyEntrySupervisor
	Schedule   PairedStrategyScheduleSnapshot
	Candidate  PairedStrategyCandidateSnapshot
	Route      PairedStrategyRouteSnapshot
	FX         PairedStrategyFXSnapshot
	Proposal   PairedStrategyProposalSnapshot
	Risk       PairedStrategyRiskSnapshot
	Account    PairedStrategyAccountSnapshot
	firstLeg   *strategyFirstLegAdmissionBridge
	dispatch   *strategyDispatchCycle
	proposals  strategyProposalAuthorityPair
}

// NewPairedStrategyEntryProductionAssembly loads KR and US from one frozen
// observation and then assembles both markets in the same wave. At this
// checkpoint schedule, candidate and FX authority are retained only long enough
// to build read-only observations and immutable candidate snapshots. Risk
// buckets, lane input and first-leg execution authority are not complete, so
// neither worker is promoted to Effective.
func (c *Context) NewPairedStrategyEntryProductionAssembly(ctx context.Context, clk clock.Clock) (StrategyEntryProductionAssembly, error) {
	if c == nil {
		return StrategyEntryProductionAssembly{}, errors.New("engine: strategy production context is unavailable")
	}
	scheduleAuthority := newStrategyScheduleAuthorityLoader(c.Paths.ConfigDir, clk, c.official, os.Getenv).collect(ctx)
	candidateStorePath := ""
	journalPath := ""
	if c.Journal != nil && strings.TrimSpace(c.Journal.Path()) != "" {
		journalPath = c.Journal.Path()
		candidateStorePath = filepath.Join(filepath.Dir(journalPath), candidatepkg.DBFileName)
	}
	candidateAuthority := newStrategyCandidateAuthorityLoader(c.Paths.ConfigDir, candidateStorePath, os.Getenv).collect(ctx, scheduleAuthority)
	routeAuthority := newStrategyRouteAuthorityLoader(c.Paths.ConfigDir, journalPath, c.AccountRef, os.Getenv).
		collect(ctx, scheduleAuthority, candidateAuthority)
	accountCurrency := strings.ToUpper(strings.TrimSpace(c.Config.Engine.AutomationGate.LimitCurrency))
	fxAuthority := newStrategyFXAuthorityLoader(c.Paths.ConfigDir, c.AccountRef, accountCurrency,
		scheduleAuthority.observedAt, c.official, os.Getenv).collect(ctx, candidateAuthority)
	evidencePath := ""
	if journalPath != "" {
		evidencePath = filepath.Join(filepath.Dir(journalPath), "evidence.db")
	}
	proposalAuthority := newStrategyProposalAuthorityLoader(c.Paths.ConfigDir, evidencePath, journalPath, c.AccountRef, os.Getenv).
		collect(ctx, scheduleAuthority, routeAuthority, fxAuthority)
	resultAuthority := proposalAuthority.ResultAuthority()
	riskAuthority := newStrategyRiskAuthorityLoader(c.Paths.ConfigDir, journalPath, c.AccountRef, accountCurrency,
		scheduleAuthority.observedAt, os.Getenv).collect(ctx, resultAuthority, fxAuthority)
	accountAuthority := newStrategyAccountAuthorityLoader(c.Paths.ConfigDir, c.AccountRef, accountCurrency,
		scheduleAuthority.observedAt, os.Getenv).collect(ctx, proposalAuthority)
	guardian, _ := c.Guardian.(*execgw.RiskGuardian)
	firstLegLoader := newProductionStrategyFirstLegAuthorityLoader(clk, c.Journal, guardian, scheduleAuthority,
		proposalAuthority, riskAuthority, fxAuthority, accountAuthority)
	firstLegBridge := newStrategyFirstLegAdmissionBridge(guardian, firstLegLoader)
	dispatchCycle := newStrategyDispatchCycle(c.Journal, c.Gateway, firstLegBridge, scheduleAuthority, fxAuthority, riskAuthority, c.strategyDispatchOwner)
	dispatchCycle.revalidateSchedule = func(checkCtx context.Context, market StrategyMarket, expected strategyScheduleMarketAuthority) error {
		fresh := newStrategyScheduleAuthorityLoader(c.Paths.ConfigDir, clk, c.official, os.Getenv).collectMarket(checkCtx, market)
		if !fresh.snapshot.Ready || fresh.restore.Activation == nil || expected.restore.Activation == nil ||
			fresh.desired.Revision != expected.desired.Revision || fresh.calendar.Version != expected.calendar.Version ||
			fresh.snapshot.ActivationManifestDigest != expected.snapshot.ActivationManifestDigest ||
			fresh.restore.Activation.Generation() != expected.restore.Activation.Generation() ||
			!fresh.restore.Activation.ExpiresAt().Equal(expected.restore.Activation.ExpiresAt()) {
			return errors.New("engine: signed scheduler activation no longer matches dispatch admission")
		}
		return nil
	}
	snapshot := StrategyEntryProductionSnapshot{
		Clock:              clk,
		AutomationVerified: c.Automation.Verified,
		EntryPermitted:     c.Automation.EntryPermitted,
		ProtectionWired:    c.Automation.Protection == ProtectionWired,
		Schedule:           scheduleAuthority.Snapshot(),
	}
	workers := make([]StrategyMarketWorker, 0, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		workers = append(workers, c.productionStrategyWorker(ctx, clk, market, scheduleAuthority, candidateAuthority,
			routeAuthority, fxAuthority, proposalAuthority, riskAuthority, accountAuthority))
	}
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: clk, CycleLimit: MaximumStrategyCycleLimit, Workers: workers})
	if err != nil {
		return StrategyEntryProductionAssembly{}, err
	}
	assembly := StrategyEntryProductionAssembly{Supervisor: supervisor, Schedule: snapshot.Schedule,
		Candidate: candidateAuthority.Snapshot(), Route: routeAuthority.Snapshot(), FX: fxAuthority.Snapshot(), Proposal: proposalAuthority.Snapshot(),
		Risk: riskAuthority.Snapshot(), Account: accountAuthority.Snapshot(), firstLeg: firstLegBridge, dispatch: dispatchCycle,
		proposals: proposalAuthority}
	if err := c.publishStrategyRuntime(assembly); err != nil {
		return StrategyEntryProductionAssembly{}, err
	}
	return assembly, nil
}

// NewRefreshingPairedStrategyEntrySupervisor builds the production runtime's
// non-blocking bootstrap. It performs no authority, broker, filesystem or
// journal read. KR and US start together as dormant refresh workers; their
// first internal poll collects one coalesced paired authority wave after the
// safety loops have started. Public Trigger remains disabled until construction
// of a future supervisor from complete opaque authority.
func (c *Context) NewRefreshingPairedStrategyEntrySupervisor(clk clock.Clock) (*StrategyEntrySupervisor, error) {
	if c == nil || clk == nil {
		return nil, errors.New("engine: paired strategy refresh supervisor unavailable")
	}
	workers := make([]StrategyMarketWorker, 0, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		workers = append(workers, StrategyMarketWorker{
			Market: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true,
			Cycle: func(cycleCtx context.Context) error {
				return c.runProductionStrategyMarketCycle(cycleCtx, clk, market)
			},
		})
	}
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{
		Clock: clk, CycleLimit: MaximumStrategyCycleLimit, Workers: workers,
	})
	if err != nil {
		return nil, err
	}
	c.strategyProjectionMu.Lock()
	c.strategySupervisor = supervisor
	c.strategyProjectionMu.Unlock()
	return supervisor, nil
}

func (c *Context) productionStrategyWorker(ctx context.Context, clk clock.Clock, market StrategyMarket,
	schedule strategyScheduleAuthorityPair, candidate strategyCandidateAuthorityPair, route strategyRouteAuthorityPair,
	fx strategyFXAuthorityPair, proposal strategyProposalAuthorityPair, riskAuthority strategyRiskAuthorityPair,
	account strategyAccountAuthorityPair,
) StrategyMarketWorker {
	if c == nil {
		return StrategyMarketWorker{Market: market}
	}
	return buildProductionStrategyMarketWorker(ctx, clk, market,
		c.Journal != nil && c.Gateway != nil && c.Guardian != nil && c.Automation.Verified,
		c.Gateway, schedule, candidate, route, fx, proposal, riskAuthority, account,
		func(cycleCtx context.Context) error { return c.runProductionStrategyMarketCycle(cycleCtx, clk, market) })
}

func buildProductionStrategyMarketWorker(ctx context.Context, clk clock.Clock, market StrategyMarket, wiringReady bool,
	gateway strategyDispatchGateway, schedule strategyScheduleAuthorityPair, candidate strategyCandidateAuthorityPair,
	route strategyRouteAuthorityPair, fx strategyFXAuthorityPair, proposal strategyProposalAuthorityPair,
	riskAuthority strategyRiskAuthorityPair, account strategyAccountAuthorityPair, cycle StrategyCycle,
) StrategyMarketWorker {
	dormant := StrategyMarketWorker{Market: market}
	if ctx == nil || clk == nil || !wiringReady || gateway == nil || cycle == nil {
		return dormant
	}
	// A dormant worker may refresh authority, but its public Trigger remains
	// disabled and it carries no entry capability. This lets an already-running
	// engine discover a later candidate/evidence wave without a process restart.
	dormant.Cycle = cycle
	dormant.PollInterval = DefaultStrategyCycleLimit
	dormant.RefreshesAuthority = true
	s, ca, ro, f, p, r, a := schedule.forMarket(market), candidate.forMarket(market), route.forMarket(market),
		fx.forMarket(market), proposal.forMarket(market), riskAuthority.forMarket(market), account.forMarket(market)
	// 제안 쪽 준비 상태와 개수는 경계가 혼자 판단한다. 여기서 다시 세면
	// 승격 규칙과 dispatch 규칙이 따로 놀 수 있고, 그 차이는 아무도 보고하지 않는다.
	result, handedOff := p.dispatchHandoff().Single()
	if !s.snapshot.Ready || s.restore.Activation == nil || !ca.snapshot.Ready || !ro.snapshot.Ready || !f.snapshot.Ready ||
		!handedOff || !r.snapshot.Ready || !a.snapshot.Ready {
		return dormant
	}
	if !result.ValidProposal() {
		return dormant
	}
	if _, err := gateway.ObserveStrategyProtection(ctx, strings.ToLower(string(market)), result.Quantity); err != nil {
		return dormant
	}
	if _, err := gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), result.Lineage.Symbol); err != nil {
		return dormant
	}
	digest := strategyWorkerEvidenceDigest(s.snapshot.ActivationManifestDigest, s.calendar.Version,
		ca.snapshot.ThresholdSetDigest, ca.snapshot.EvidenceDigest, ro.snapshot.OwnerSetDigest,
		f.snapshot.Digest, p.snapshot.ProposalSetDigest, r.snapshot.BundleDigest, a.snapshot.Identity)
	if !validStrategyDigest(digest) || s.desired.Revision == 0 || a.authority.FreshUntil().IsZero() {
		return dormant
	}
	return StrategyMarketWorker{Market: market, Effective: true, AuthorityGeneration: s.desired.Revision,
		AuthorityExpiresAt: a.authority.FreshUntil(), EvidenceDigest: digest, LatchRevision: 1,
		PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: cycle}
}

func (c *Context) runProductionStrategyMarketCycle(ctx context.Context, clk clock.Clock, market StrategyMarket) error {
	fresh, err := c.refreshPairedStrategyEntryProductionAssembly(ctx, clk)
	if err != nil {
		return err
	}
	if fresh.dispatch == nil {
		return nil
	}
	// 조정자가 고른 것은 이 한 값으로만 공유 dispatch 에 닿는다.
	//
	// 여기서 `Single()` 이 아니라 `Deliver` 를 쓰는 이유는 하나다: **무시할 수
	// 있는 답을 이 함수에 두지 않기 위해서**. 앞선 판본은 `result, handedOff :=`
	// 로 받아 `if !handedOff || …` 로 막았고, 적대 리뷰가 그 관문을 지우면서
	// `if !handedOff { }` 같은 빈 조건만 남기는 편집으로 두 스위트를 모두
	// 통과시켰다 — 답을 "썼는지" 보는 검사는 답이 무엇을 막는지 보지 못한다.
	// Deliver 에서는 몸통이 도는지 마는지를 이 함수가 정하지 않는다.
	//
	// 몸통이 받는 result 는 경계가 건넨 값이다. 그 값을 다른 것으로 바꿔치는
	// 편집은 `entries` 를 다시 읽어야 하고, 그 철자는
	// TestSeamConsumersCannotReadTheRawEntryListAgain 이 막는다.
	return fresh.proposals.forMarket(market).dispatchHandoff().Deliver(func(result strategyflow.Result) error {
		lineage := result.Lineage
		cas, err := c.Journal.CurrentPositionCampaignCAS(ctx, lineage.AccountRef, string(lineage.Market), lineage.Symbol)
		if err != nil {
			return err
		}
		if cas.Claimed || cas.State != "FLAT" && cas.State != "CLOSED" {
			return nil
		}
		_, err = fresh.dispatch.dispatch(ctx, result)
		if errors.Is(err, journal.ErrStrategyDispatchLeaseConsumed) {
			return nil
		}
		return err
	})
}

func (c *Context) refreshPairedStrategyEntryProductionAssembly(ctx context.Context, clk clock.Clock) (StrategyEntryProductionAssembly, error) {
	if c == nil || clk == nil {
		return StrategyEntryProductionAssembly{}, errors.New("engine: paired strategy refresh unavailable")
	}
	c.strategyRefreshMu.Lock()
	defer c.strategyRefreshMu.Unlock()
	now := clk.Now().UTC()
	if c.strategyRefresh != nil && !now.Before(c.strategyRefreshAt) && now.Sub(c.strategyRefreshAt) < time.Second {
		return *c.strategyRefresh, nil
	}
	fresh, err := c.NewPairedStrategyEntryProductionAssembly(ctx, clk)
	if err != nil {
		return StrategyEntryProductionAssembly{}, err
	}
	c.strategyRefreshAt = now
	c.strategyRefresh = &fresh
	return fresh, nil
}

func strategyWorkerEvidenceDigest(values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		_, _ = hash.Write([]byte(fmt.Sprintf("%d:%s", len(value), value)))
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

// NewPairedStrategyEntryProductionAssembly constructs KR and US together from
// the read-only production snapshot. The context-owned path may already have
// verified signed per-market activation, but this value-only constructor never
// receives that opaque authority. Risk-snapshot, lane-input and first-leg cycle
// authority are not complete yet, so the only valid result is two
// explicit dormant workers. The observations are intentionally not converted
// into authority.
func NewPairedStrategyEntryProductionAssembly(snapshot StrategyEntryProductionSnapshot) (*StrategyEntrySupervisor, error) {
	return NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: snapshot.Clock, Workers: []StrategyMarketWorker{
		{Market: StrategyMarketKR},
		{Market: StrategyMarketUS},
	}})
}

// NewDormantStrategyEntrySupervisor constructs the deploy-safe baseline: both
// markets exist, both are OFF and neither can evaluate or mutate anything.
func NewDormantStrategyEntrySupervisor() (*StrategyEntrySupervisor, error) {
	return NewPairedStrategyEntryProductionAssembly(StrategyEntryProductionSnapshot{})
}

// NewStrategyEntrySupervisor validates the closed paired-market assembly.
func NewStrategyEntrySupervisor(opts StrategyEntrySupervisorOptions) (*StrategyEntrySupervisor, error) {
	depth := opts.QueueDepth
	if depth == 0 {
		depth = DefaultStrategyQueueDepth
	}
	if depth < 1 || depth > MaximumStrategyQueueDepth {
		return nil, fmt.Errorf("engine: strategy queue depth %d is outside 1..%d", depth, MaximumStrategyQueueDepth)
	}
	cycleLimit := opts.CycleLimit
	if cycleLimit == 0 {
		cycleLimit = DefaultStrategyCycleLimit
	}
	if cycleLimit < 0 || cycleLimit > MaximumStrategyCycleLimit {
		return nil, fmt.Errorf("engine: strategy cycle limit %s is outside 0..%s", cycleLimit, MaximumStrategyCycleLimit)
	}
	if len(opts.Workers) != 2 {
		return nil, errors.New("engine: strategy supervisor requires exact independent KR and US workers")
	}
	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	now := clk.Now()
	if now.IsZero() {
		return nil, errors.New("engine: strategy supervisor clock is unavailable")
	}

	workers := make(map[StrategyMarket]*strategyMarketRuntime, 2)
	for _, descriptor := range opts.Workers {
		if !validStrategyMarket(descriptor.Market) {
			return nil, fmt.Errorf("engine: unsupported strategy market %q", descriptor.Market)
		}
		if _, duplicate := workers[descriptor.Market]; duplicate {
			return nil, fmt.Errorf("engine: duplicate strategy market %q", descriptor.Market)
		}
		if descriptor.Effective && descriptor.Cycle == nil {
			return nil, fmt.Errorf("engine: effective %s strategy worker has no evaluation cycle", descriptor.Market)
		}
		if descriptor.Effective && (descriptor.AuthorityGeneration == 0 || descriptor.AuthorityExpiresAt.IsZero() ||
			!now.Before(descriptor.AuthorityExpiresAt) || !validStrategyDigest(descriptor.EvidenceDigest) || descriptor.LatchRevision == 0) {
			return nil, fmt.Errorf("engine: effective %s strategy worker has incomplete authority", descriptor.Market)
		}
		if descriptor.Effective && ((descriptor.RestartAttempt == 0 && (descriptor.FirstRefusal != StrategyWorkerRefusalNone || !descriptor.RestartNotBefore.IsZero())) ||
			(descriptor.RestartAttempt > 0 && (descriptor.FirstRefusal == StrategyWorkerRefusalNone || !validStrategyWorkerRefusal(descriptor.FirstRefusal) || descriptor.RestartNotBefore.IsZero()))) {
			return nil, fmt.Errorf("engine: effective %s strategy worker has invalid restart authority", descriptor.Market)
		}
		if descriptor.PollInterval < 0 || descriptor.PollInterval > MaximumStrategyCycleLimit ||
			(descriptor.PollInterval > 0 && descriptor.PollInterval < MinimumStrategyPollInterval) ||
			(descriptor.PollInterval > 0 && !descriptor.Effective && !descriptor.RefreshesAuthority) ||
			(descriptor.RefreshesAuthority && (descriptor.PollInterval == 0 || descriptor.Cycle == nil)) {
			return nil, fmt.Errorf("engine: %s strategy worker has invalid production polling authority", descriptor.Market)
		}
		if !descriptor.Effective && (descriptor.AuthorityGeneration != 0 || !descriptor.AuthorityExpiresAt.IsZero() || descriptor.EvidenceDigest != "" || descriptor.LatchRevision != 0) {
			return nil, fmt.Errorf("engine: dormant %s strategy worker carries authority", descriptor.Market)
		}
		if !descriptor.Effective && (descriptor.FirstRefusal != StrategyWorkerRefusalNone || descriptor.RestartAttempt != 0 || !descriptor.RestartNotBefore.IsZero()) {
			return nil, fmt.Errorf("engine: dormant %s strategy worker carries restart state", descriptor.Market)
		}
		workers[descriptor.Market] = &strategyMarketRuntime{
			descriptor:       descriptor,
			queue:            make(chan struct{}, depth),
			effective:        descriptor.Effective,
			latchRevision:    descriptor.LatchRevision,
			firstRefusal:     descriptor.FirstRefusal,
			restartAttempt:   descriptor.RestartAttempt,
			restartNotBefore: descriptor.RestartNotBefore,
		}
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		if workers[market] == nil {
			return nil, fmt.Errorf("engine: strategy supervisor is missing %s", market)
		}
	}

	return &StrategyEntrySupervisor{
		clk: clk, cycleLimit: cycleLimit, workers: workers, ready: make(chan struct{}), faults: make(chan StrategyWorkerFault, 2),
	}, nil
}

func validStrategyDigest(value string) bool {
	if strings.TrimSpace(value) != value || len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

// SupervisedLoop returns the single outer loop added to RuntimeOptions.Loops.
func (s *StrategyEntrySupervisor) SupervisedLoop() SupervisedLoop {
	if s == nil {
		return SupervisedLoop{Name: StrategyEntryLoopName}
	}
	return SupervisedLoop{Name: StrategyEntryLoopName, Run: s.Run}
}

// Faults is a read-only bounded stream. Reading a fault does not acknowledge,
// release or recover its latch.
func (s *StrategyEntrySupervisor) Faults() <-chan StrategyWorkerFault {
	if s == nil {
		return nil
	}
	return s.faults
}

// Ready closes only after both market children exist and Trigger acceptance has
// opened. It is an observation barrier, not an activation API.
func (s *StrategyEntrySupervisor) Ready() <-chan struct{} {
	if s == nil {
		return nil
	}
	return s.ready
}

// Trigger schedules one non-blocking evaluation cycle on exactly one market.
// The accepting check and queue send share one lock with shutdown, so no enqueue
// can win after shutdown begins.
func (s *StrategyEntrySupervisor) Trigger(market StrategyMarket) StrategyTriggerResult {
	if s == nil || !validStrategyMarket(market) {
		return StrategyTriggerInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	worker := s.workers[market]
	if !s.accepting || worker == nil || !worker.effective || worker.latched || worker.descriptor.Cycle == nil {
		return StrategyTriggerDisabled
	}
	select {
	case worker.queue <- struct{}{}:
		return StrategyTriggerEnqueued
	default:
		return StrategyTriggerFull
	}
}

func (s *StrategyEntrySupervisor) Snapshot(market StrategyMarket) (StrategyWorkerSnapshot, bool) {
	if s == nil || !validStrategyMarket(market) {
		return StrategyWorkerSnapshot{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	worker := s.workers[market]
	if worker == nil {
		return StrategyWorkerSnapshot{}, false
	}
	return StrategyWorkerSnapshot{
		Market: market, Effective: worker.effective, Latched: worker.latched,
		FirstFailure: worker.firstFailure, FirstRefusal: worker.firstRefusal, FirstAbnormal: worker.firstAbnormal,
		LatchID: worker.latchID, LatchRevision: worker.latchRevision,
		AuthorityGeneration: worker.descriptor.AuthorityGeneration, AuthorityExpiresAt: worker.descriptor.AuthorityExpiresAt,
		EvidenceDigest:      worker.descriptor.EvidenceDigest,
		AbandonedEvaluation: worker.abandoned,
		RestartAttempt:      worker.restartAttempt, RestartNotBefore: worker.restartNotBefore,
		QueueDepth: len(worker.queue), QueueCapacity: cap(worker.queue),
	}, true
}

// Run starts KR and US children behind one start barrier and remains alive across
// every market-local failure. It returns only on cancellation or central fault.
func (s *StrategyEntrySupervisor) Run(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("%w: nil strategy supervisor", ErrStrategyCentralIntegrity)
	}
	s.mu.Lock()
	if s.run {
		s.mu.Unlock()
		return fmt.Errorf("%w: strategy supervisor Run called more than once", ErrStrategyCentralIntegrity)
	}
	s.run = true
	s.mu.Unlock()

	childCtx, cancel := context.WithCancel(ctx)
	start := make(chan struct{})
	central := make(chan error, 1)
	var children sync.WaitGroup
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		worker := s.workers[market]
		children.Add(1)
		go func(worker *strategyMarketRuntime) {
			defer children.Done()
			s.runMarket(childCtx, start, worker, central)
		}(worker)
	}
	s.mu.Lock()
	s.accepting = true
	close(s.ready)
	s.mu.Unlock()
	close(start)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		worker := s.workers[market]
		if worker.descriptor.PollInterval <= 0 {
			continue
		}
		children.Add(1)
		go func(market StrategyMarket, interval time.Duration) {
			defer children.Done()
			s.runStrategyPoller(childCtx, market, interval)
		}(market, worker.descriptor.PollInterval)
	}

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case cause := <-central:
		err = fmt.Errorf("%w: %v", ErrStrategyCentralIntegrity, cause)
	}
	s.stopAcceptingAndDrain(cancel)
	children.Wait()
	return err
}

func (s *StrategyEntrySupervisor) runStrategyPoller(ctx context.Context, market StrategyMarket, interval time.Duration) {
	for {
		if result := s.enqueueStrategyPoll(market); result == StrategyTriggerInvalid || result == StrategyTriggerDisabled {
			return
		}
		if err := s.clk.Sleep(ctx, interval); err != nil {
			return
		}
	}
}

func (s *StrategyEntrySupervisor) enqueueStrategyPoll(market StrategyMarket) StrategyTriggerResult {
	if s == nil || !validStrategyMarket(market) {
		return StrategyTriggerInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	worker := s.workers[market]
	if !s.accepting || worker == nil || worker.latched || worker.descriptor.Cycle == nil ||
		(!worker.effective && !worker.descriptor.RefreshesAuthority) {
		return StrategyTriggerDisabled
	}
	select {
	case worker.queue <- struct{}{}:
		return StrategyTriggerEnqueued
	default:
		return StrategyTriggerFull
	}
}

func (s *StrategyEntrySupervisor) stopAcceptingAndDrain(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accepting = false
	cancel()
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		queue := s.workers[market].queue
		for {
			select {
			case <-queue:
				continue
			default:
				break
			}
			break
		}
	}
}

func (s *StrategyEntrySupervisor) runMarket(ctx context.Context, start <-chan struct{}, worker *strategyMarketRuntime, central chan<- error) {
	select {
	case <-ctx.Done():
		return
	case <-start:
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-worker.queue:
			refreshOnly := false
			s.mu.RLock()
			refreshOnly = !worker.effective && worker.descriptor.RefreshesAuthority
			s.mu.RUnlock()
			allowed, expired := s.evaluationState(worker)
			if expired {
				restartNotBefore, err := s.latchMarket(worker, ErrStrategyAuthorityExpired, false)
				if err != nil {
					s.signalCentral(central, err)
					return
				}
				if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {
					if ctx.Err() != nil {
						return
					}
					s.signalCentral(central, err)
					return
				}
				continue
			}
			if !allowed {
				continue
			}
			err, abnormal, cancelled, abandoned := invokeBoundedStrategyCycle(ctx, s.clk, s.cycleLimit, worker.descriptor.Cycle)
			if abandoned {
				s.markAbandoned(worker)
			}
			if cancelled {
				return
			}
			if err == nil {
				continue
			}
			if refreshOnly {
				continue
			}
			if isCentralStrategyIntegrity(err) {
				s.signalCentral(central, err)
				return
			}
			restartNotBefore, err := s.latchMarket(worker, err, abnormal)
			if err != nil {
				s.signalCentral(central, err)
				return
			}
			if err := s.waitMarketRestart(ctx, restartNotBefore); err != nil {
				if ctx.Err() != nil {
					return
				}
				s.signalCentral(central, err)
				return
			}
		}
	}
}

func (s *StrategyEntrySupervisor) waitMarketRestart(ctx context.Context, notBefore time.Time) error {
	if notBefore.IsZero() {
		return errors.New("strategy market restart deadline is unavailable")
	}
	now := s.clk.Now()
	if now.IsZero() {
		return errors.New("strategy market restart clock is unavailable")
	}
	delay := notBefore.Sub(now)
	if delay > MaximumStrategyRestartBackoff {
		return errors.New("strategy market restart delay is outside the bounded contract")
	}
	if delay <= 0 {
		return nil
	}
	return s.clk.Sleep(ctx, delay)
}

func (s *StrategyEntrySupervisor) evaluationState(worker *strategyMarketRuntime) (bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.accepting || worker == nil || worker.latched || worker.descriptor.Cycle == nil ||
		(!worker.effective && !worker.descriptor.RefreshesAuthority) {
		return false, false
	}
	if worker.descriptor.RefreshesAuthority {
		return true, false
	}
	fresh := s.clk.Now().Before(worker.descriptor.AuthorityExpiresAt)
	return fresh, !fresh
}

func (s *StrategyEntrySupervisor) markAbandoned(worker *strategyMarketRuntime) {
	s.mu.Lock()
	worker.abandoned = true
	s.mu.Unlock()
}

type strategyCycleResult struct {
	err      error
	abnormal bool
}

func invokeBoundedStrategyCycle(ctx context.Context, clk clock.Clock, limit time.Duration, cycle StrategyCycle) (error, bool, bool, bool) {
	result := make(chan strategyCycleResult, 1)
	go func() {
		result <- invokeStrategyCycle(ctx, cycle)
	}()
	watchdogCtx, cancelWatchdog := context.WithCancel(ctx)
	defer cancelWatchdog()
	deadline := make(chan error, 1)
	go func() { deadline <- clk.Sleep(watchdogCtx, limit) }()
	select {
	case <-ctx.Done():
		return ctx.Err(), false, true, true
	case outcome := <-result:
		return outcome.err, outcome.abnormal, false, false
	case <-deadline:
		if ctx.Err() != nil {
			return ctx.Err(), false, true, true
		}
		return ErrStrategyCycleDeadline, true, false, true
	}
}

func invokeStrategyCycle(ctx context.Context, cycle StrategyCycle) (outcome strategyCycleResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			outcome.abnormal = true
			outcome.err = fmt.Errorf("strategy evaluation panic: %v", recovered)
			if recoveredErr, ok := recovered.(error); ok && isCentralStrategyIntegrity(recoveredErr) {
				outcome.err = recoveredErr
			}
		}
	}()
	outcome.err = cycle(ctx)
	return outcome
}

func (s *StrategyEntrySupervisor) latchMarket(worker *strategyMarketRuntime, failure error, abnormal bool) (time.Time, error) {
	reason := strings.TrimSpace(failure.Error())
	if reason == "" {
		reason = "strategy evaluation failed without a reason"
	}
	refusal := StrategyWorkerRefusalFailure
	if abnormal {
		refusal = StrategyWorkerRefusalAbnormal
	}
	if errors.Is(failure, ErrStrategyAuthorityExpired) {
		refusal = StrategyWorkerRefusalAuthorityExpired
	}
	observedAt := s.clk.Now()
	if observedAt.IsZero() {
		return time.Time{}, errors.New("strategy fault observation time is unavailable")
	}
	s.mu.Lock()
	if worker.latchRevision == math.MaxUint64 {
		s.mu.Unlock()
		return time.Time{}, errors.New("strategy latch revision exhausted")
	}
	worker.effective = false
	worker.latched = true
	if worker.firstRefusal == StrategyWorkerRefusalNone {
		worker.firstRefusal = refusal
	}
	if worker.firstFailure == "" {
		worker.firstFailure = reason
		worker.firstAbnormal = abnormal
		worker.latchID = fmt.Sprintf("strategy-latch:%s:%d:%d", worker.descriptor.Market, worker.descriptor.AuthorityGeneration, worker.latchRevision+1)
		worker.latchRevision++
	}
	if worker.restartAttempt < math.MaxUint64 {
		worker.restartAttempt++
	}
	restartDelay := strategyRestartBackoff(worker.restartAttempt)
	worker.restartNotBefore = strategyRestartNotBefore(observedAt, restartDelay)
	fault := StrategyWorkerFault{
		Market: worker.descriptor.Market, LatchID: worker.latchID,
		ExpectedRevision: worker.latchRevision - 1, NextRevision: worker.latchRevision,
		AuthorityGeneration: worker.descriptor.AuthorityGeneration, AuthorityExpiresAt: worker.descriptor.AuthorityExpiresAt,
		EvidenceDigest: worker.descriptor.EvidenceDigest,
		Reason:         worker.firstFailure, FirstRefusal: worker.firstRefusal, Abnormal: worker.firstAbnormal, ObservedAt: observedAt.UTC(),
		RestartAttempt: worker.restartAttempt, RestartNotBefore: worker.restartNotBefore,
	}
	s.mu.Unlock()
	select {
	case s.faults <- fault:
		return fault.RestartNotBefore, nil
	default:
		return time.Time{}, errors.New("strategy fault handoff saturated before durable owner acknowledgement")
	}
}

func strategyRestartBackoff(attempt uint64) time.Duration {
	maximumSteps := uint64(MaximumStrategyRestartBackoff / DefaultStrategyRestartStep)
	if attempt >= maximumSteps {
		return MaximumStrategyRestartBackoff
	}
	return time.Duration(attempt) * DefaultStrategyRestartStep
}

func strategyRestartNotBefore(observed time.Time, backoff time.Duration) time.Time {
	maximumTime := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
	if observed.After(maximumTime.Add(-backoff)) {
		return maximumTime
	}
	return observed.Add(backoff)
}

func (s *StrategyEntrySupervisor) signalCentral(central chan<- error, err error) {
	select {
	case central <- err:
	default:
	}
}
