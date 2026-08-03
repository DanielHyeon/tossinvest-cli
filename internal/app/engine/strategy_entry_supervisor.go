package engine

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// StrategyEntryLoopName is the single outer runtime loop which owns both market
// evaluation children. Market-local failures stay inside this loop so they do
// not cancel the engine's safety-class loops.
const StrategyEntryLoopName = "strategy-entry"

const (
	DefaultStrategyQueueDepth = 1
	MaximumStrategyQueueDepth = 1024
	DefaultStrategyCycleLimit = 5 * time.Second
	MaximumStrategyCycleLimit = 30 * time.Second
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
	Abnormal            bool
	ObservedAt          time.Time
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
	descriptor    StrategyMarketWorker
	queue         chan struct{}
	effective     bool
	latched       bool
	firstFailure  string
	firstAbnormal bool
	latchID       string
	latchRevision uint64
	abandoned     bool
}

// StrategyWorkerSnapshot is a read-only, market-keyed operational view. It
// carries no activation, release or mutation method.
type StrategyWorkerSnapshot struct {
	Market              StrategyMarket
	Effective           bool
	Latched             bool
	FirstFailure        string
	FirstAbnormal       bool
	LatchID             string
	LatchRevision       uint64
	AuthorityGeneration uint64
	AuthorityExpiresAt  time.Time
	EvidenceDigest      string
	AbandonedEvaluation bool
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

// NewDormantStrategyEntrySupervisor constructs the deploy-safe baseline: both
// markets exist, both are OFF and neither can evaluate or mutate anything.
func NewDormantStrategyEntrySupervisor() (*StrategyEntrySupervisor, error) {
	return NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Workers: []StrategyMarketWorker{
		{Market: StrategyMarketKR},
		{Market: StrategyMarketUS},
	}})
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
		if !descriptor.Effective && (descriptor.AuthorityGeneration != 0 || !descriptor.AuthorityExpiresAt.IsZero() || descriptor.EvidenceDigest != "" || descriptor.LatchRevision != 0) {
			return nil, fmt.Errorf("engine: dormant %s strategy worker carries authority", descriptor.Market)
		}
		workers[descriptor.Market] = &strategyMarketRuntime{
			descriptor:    descriptor,
			queue:         make(chan struct{}, depth),
			effective:     descriptor.Effective,
			latchRevision: descriptor.LatchRevision,
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
		FirstFailure: worker.firstFailure, FirstAbnormal: worker.firstAbnormal,
		LatchID: worker.latchID, LatchRevision: worker.latchRevision,
		AuthorityGeneration: worker.descriptor.AuthorityGeneration, AuthorityExpiresAt: worker.descriptor.AuthorityExpiresAt,
		EvidenceDigest:      worker.descriptor.EvidenceDigest,
		AbandonedEvaluation: worker.abandoned,
		QueueDepth:          len(worker.queue), QueueCapacity: cap(worker.queue),
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
			allowed, expired := s.evaluationState(worker)
			if expired {
				if err := s.latchMarket(worker, ErrStrategyAuthorityExpired, false); err != nil {
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
			if isCentralStrategyIntegrity(err) {
				s.signalCentral(central, err)
				return
			}
			if err := s.latchMarket(worker, err, abnormal); err != nil {
				s.signalCentral(central, err)
				return
			}
		}
	}
}

func (s *StrategyEntrySupervisor) evaluationState(worker *strategyMarketRuntime) (bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.accepting || worker == nil || !worker.effective || worker.latched || worker.descriptor.Cycle == nil {
		return false, false
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

func (s *StrategyEntrySupervisor) latchMarket(worker *strategyMarketRuntime, failure error, abnormal bool) error {
	reason := strings.TrimSpace(failure.Error())
	if reason == "" {
		reason = "strategy evaluation failed without a reason"
	}
	s.mu.Lock()
	if worker.latchRevision == math.MaxUint64 {
		s.mu.Unlock()
		return errors.New("strategy latch revision exhausted")
	}
	worker.effective = false
	worker.latched = true
	if worker.firstFailure == "" {
		worker.firstFailure = reason
		worker.firstAbnormal = abnormal
		worker.latchID = fmt.Sprintf("strategy-latch:%s:%d:%d", worker.descriptor.Market, worker.descriptor.AuthorityGeneration, worker.latchRevision+1)
		worker.latchRevision++
	}
	observedAt := s.clk.Now()
	if observedAt.IsZero() {
		s.mu.Unlock()
		return errors.New("strategy fault observation time is unavailable")
	}
	fault := StrategyWorkerFault{
		Market: worker.descriptor.Market, LatchID: worker.latchID,
		ExpectedRevision: worker.latchRevision - 1, NextRevision: worker.latchRevision,
		AuthorityGeneration: worker.descriptor.AuthorityGeneration, AuthorityExpiresAt: worker.descriptor.AuthorityExpiresAt,
		EvidenceDigest: worker.descriptor.EvidenceDigest,
		Reason:         worker.firstFailure, Abnormal: worker.firstAbnormal, ObservedAt: observedAt.UTC(),
	}
	s.mu.Unlock()
	select {
	case s.faults <- fault:
		return nil
	default:
		return errors.New("strategy fault handoff saturated before durable owner acknowledgement")
	}
}

func (s *StrategyEntrySupervisor) signalCentral(central chan<- error, err error) {
	select {
	case central <- err:
	default:
	}
}
