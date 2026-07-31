package scheduler

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	budgetObservationMaxAge           = 2 * time.Minute
	deltaResetTolerance               = time.Second
	maxIssuedCommitmentsPerGeneration = 256
	maxObservationCyclesPerGeneration = 1024
)

type PollClass string

const (
	PollCandidate     PollClass = "candidate"
	PollEntry         PollClass = "entry"
	PollAnalytics     PollClass = "analytics"
	PollEmergencyExit PollClass = "emergency-exit"
	PollReconcile     PollClass = "reconcile"
	PollFillDetection PollClass = "fill-detection"
	PollProtection    PollClass = "protection"
)

// Priority is a total ordering for a047's future cadence dispatcher. Higher
// values run first; candidate and entry intentionally share one tier.
func (p PollClass) Priority() int {
	switch p {
	case PollEmergencyExit:
		return 6
	case PollReconcile:
		return 5
	case PollFillDetection:
		return 4
	case PollProtection:
		return 3
	case PollCandidate, PollEntry:
		return 2
	case PollAnalytics:
		return 1
	default:
		return 0
	}
}

type BudgetReason string

const (
	BudgetGranted               BudgetReason = "GRANTED"
	BudgetReserved              BudgetReason = "SAFETY_RESERVE"
	BudgetEntryPriority         BudgetReason = "ENTRY_CANDIDATE_PRIORITY"
	BudgetSafetyPriority        BudgetReason = "SAFETY_PRIORITY"
	BudgetMissing               BudgetReason = "MISSING_PROVENANCE"
	BudgetStale                 BudgetReason = "STALE_PROVENANCE"
	BudgetClockSkew             BudgetReason = "CLOCK_SKEW"
	BudgetUnknownClass          BudgetReason = "UNKNOWN_POLL_CLASS"
	BudgetInvalidReset          BudgetReason = "INVALID_RESET_KIND"
	BudgetConflictingProvenance BudgetReason = "CONFLICTING_PROVENANCE"
	BudgetInvalidBounds         BudgetReason = "INVALID_BUDGET_BOUNDS"
	BudgetTokenUnavailable      BudgetReason = "COMMITMENT_TOKEN_UNAVAILABLE"
)

// CommitmentToken is an opaque, unguessable capability for completing one
// low-priority request. Its authority is bound to the issuing coordinator,
// endpoint key, poll class, and reset generation.
type CommitmentToken struct {
	coordinator [16]byte
	capability  [32]byte
	keyDigest   [sha256.Size]byte
	class       PollClass
	generation  uint64
}

// ObservationCycle is an opaque, one-shot proof that one rate-budget request
// began at a specific point in the coordinator's causal history. Wall-clock
// timestamps cannot provide that proof because clocks can roll back and a held
// response can be processed long after the request began.
type ObservationCycle struct {
	coordinator [16]byte
	capability  [32]byte
	keyDigest   [sha256.Size]byte
	generation  uint64
}

type BudgetGrant struct {
	Allowed    bool
	Reason     BudgetReason
	Remaining  int
	Reserve    int
	Available  int
	Reset      time.Time
	ObservedAt time.Time
	// Commitment must be completed after the request finishes, regardless of
	// success, error, or cancellation. Completion records the outcome but does
	// not restore capacity without a causally later authoritative request cycle.
	// Safety-class grants deliberately have no commitment.
	Commitment CommitmentToken
}

type budgetCommitment struct {
	class              PollClass
	generation         uint64
	completed          bool
	completedAt        time.Time
	completionSequence uint64
}

type observationCycleRecord struct {
	generation          uint64
	completionWatermark uint64
}

type endpointBudget struct {
	observation             official.RateBudget
	commitments             map[[32]byte]budgetCommitment
	issued                  map[[32]byte]struct{}
	observationCycles       map[[32]byte]observationCycleRecord
	issuedObservationCycles map[[32]byte]struct{}
	provenanceConflict      bool
	trustedReset            time.Time
	trustedResetAnchor      time.Time
	trustedResetKind        official.ResetKind
	hasTrustedReset         bool
	generation              uint64
	generationExhausted     bool
}

type BudgetCoordinator struct {
	mu                 sync.Mutex
	endpoints          map[string]endpointBudget
	entropy            io.Reader
	now                func() time.Time
	coordinatorID      [16]byte
	entropyOK          bool
	completionSequence uint64
}

func NewBudgetCoordinator() *BudgetCoordinator {
	return newBudgetCoordinatorWithEntropy(rand.Reader)
}

// newBudgetCoordinatorWithEntropy is the deterministic entropy test seam for
// token issuance. Production callers always use crypto/rand.Reader and the
// system clock; the clock-aware variant supports completion-order race tests.
func newBudgetCoordinatorWithEntropy(entropy io.Reader) *BudgetCoordinator {
	return newBudgetCoordinatorWithEntropyAndClock(entropy, time.Now)
}

func newBudgetCoordinatorWithEntropyAndClock(entropy io.Reader, now func() time.Time) *BudgetCoordinator {
	c := &BudgetCoordinator{endpoints: make(map[string]endpointBudget), entropy: entropy, now: now}
	if entropy != nil {
		_, err := io.ReadFull(entropy, c.coordinatorID[:])
		c.entropyOK = err == nil
	}
	return c
}

func SafetyReserve(remaining int) int {
	reserve := remaining/2 + remaining%2
	if reserve < 5 {
		return 5
	}
	return reserve
}

// BeginObservation starts one endpoint request/observation cycle. The returned
// value is deliberately opaque and one-shot. A zero value means the coordinator
// could not mint bounded, unguessable chronology evidence, so the caller may
// still Observe the response but it cannot reconcile commitments.
func (c *BudgetCoordinator) BeginObservation(key string) ObservationCycle {
	if c == nil {
		return ObservationCycle{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, ok := c.endpoints[key]
	if !ok {
		state = newEndpointBudget()
	}
	if !c.entropyOK || state.generationExhausted || len(state.issuedObservationCycles) >= maxObservationCyclesPerGeneration {
		return ObservationCycle{}
	}
	var capability [32]byte
	if _, err := io.ReadFull(c.entropy, capability[:]); err != nil {
		c.entropyOK = false
		return ObservationCycle{}
	}
	if _, collision := state.issuedObservationCycles[capability]; collision {
		return ObservationCycle{}
	}
	state.observationCycles[capability] = observationCycleRecord{
		generation:          state.generation,
		completionWatermark: c.completionSequence,
	}
	state.issuedObservationCycles[capability] = struct{}{}
	c.endpoints[key] = state
	return ObservationCycle{
		coordinator: c.coordinatorID,
		capability:  capability,
		keyDigest:   sha256.Sum256([]byte(key)),
		generation:  state.generation,
	}
}

// Observe ingests budget evidence without granting reconciliation authority.
// It is suitable for initial/manual observations. A request that may reconcile
// commitments must use BeginObservation before it starts and ObserveCycle when
// its response completes.
func (c *BudgetCoordinator) Observe(observation official.RateBudget) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.observeLocked(observation, nil)
}

// ObserveCycle ingests one response with the opaque cycle minted before its
// request started. False means the cycle was zero, forged, replayed, cross-key,
// cross-coordinator, or cross-generation; no observation is ingested then.
func (c *BudgetCoordinator) ObserveCycle(observation official.RateBudget, cycle ObservationCycle) bool {
	if c == nil || cycle == (ObservationCycle{}) {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, ok := c.endpoints[observation.Path]
	if !ok || cycle.coordinator != c.coordinatorID || cycle.keyDigest != sha256.Sum256([]byte(observation.Path)) ||
		cycle.generation != state.generation {
		return false
	}
	record, ok := state.observationCycles[cycle.capability]
	if !ok || record.generation != cycle.generation {
		return false
	}
	delete(state.observationCycles, cycle.capability)
	c.endpoints[observation.Path] = state
	c.observeLocked(observation, &record)
	return true
}

func (c *BudgetCoordinator) observeLocked(observation official.RateBudget, cycle *observationCycleRecord) {
	previous, ok := c.endpoints[observation.Path]
	if ok {
		hasPreviousObservation := !previous.observation.ObservedAt.IsZero()
		switch {
		case hasPreviousObservation && observation.ObservedAt.Before(previous.observation.ObservedAt):
			return
		case hasPreviousObservation && observation.ObservedAt.Equal(previous.observation.ObservedAt):
			// Equal-time reports are only unambiguous when every provenance field
			// agrees. A contradictory correction invalidates low-priority grants
			// until a strictly newer response supplies fresh evidence.
			if !sameBudgetProvenance(previous.observation, observation) {
				previous.provenanceConflict = true
				c.endpoints[observation.Path] = previous
				return
			}
			if observation.Remaining < previous.observation.Remaining {
				previous.observation.Remaining = observation.Remaining
			}
			if trustedBudgetWindow(observation) && sameTrustedWindow(previous, observation) {
				reconcileCompleted(previous.commitments, cycle)
			}
			c.endpoints[observation.Path] = previous
			return
		}
		// A reset relation is trusted only through exact epoch identity or bounded
		// delta derivation around a fixed anchor. Reconciliation itself additionally
		// requires the request-start cycle; response processing time has no authority.
		if trustedBudgetWindow(observation) {
			switch classifyBudgetWindow(previous, observation) {
			case budgetWindowInitial:
				setTrustedReset(&previous, observation)
			case budgetWindowSame:
				reconcileCompleted(previous.commitments, cycle)
				retainEarliestDeltaReset(&previous, observation)
			case budgetWindowNext:
				if cycle != nil {
					reconcileCompleted(previous.commitments, cycle)
				}
				if cycle != nil && len(previous.commitments) == 0 {
					advanceBudgetGeneration(&previous, observation)
				}
			case budgetWindowConflict:
				previous.provenanceConflict = true
				c.endpoints[observation.Path] = previous
				return
			}
		}
		previous.observation = observation
		previous.provenanceConflict = false
		c.endpoints[observation.Path] = previous
		return
	}
	state := newEndpointBudget()
	state.observation = observation
	if trustedBudgetWindow(observation) {
		setTrustedReset(&state, observation)
	}
	c.endpoints[observation.Path] = state
}

func (c *BudgetCoordinator) TryAcquire(key string, class PollClass, now time.Time) BudgetGrant {
	if !isKnownPollClass(class) {
		return BudgetGrant{Reason: BudgetUnknownClass}
	}
	if isSafetyClass(class) {
		return BudgetGrant{Allowed: true, Reason: BudgetSafetyPriority}
	}
	if c == nil {
		return BudgetGrant{Reason: BudgetMissing}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, ok := c.endpoints[key]
	if ok && state.provenanceConflict {
		return BudgetGrant{Reason: BudgetConflictingProvenance}
	}
	if !ok || !state.observation.Reported || state.observation.Reset.IsZero() || state.observation.ObservedAt.IsZero() {
		return BudgetGrant{Reason: BudgetMissing}
	}
	obs := state.observation
	effectiveReset := obs.Reset
	if state.hasTrustedReset {
		effectiveReset = state.trustedReset
	}
	grant := BudgetGrant{Remaining: obs.Remaining, Reset: effectiveReset, ObservedAt: obs.ObservedAt}
	if !validResetSemantics(obs) {
		grant.Reason = BudgetInvalidReset
		return grant
	}
	if now.Before(obs.ObservedAt) {
		grant.Reason = BudgetClockSkew
		return grant
	}
	if !now.Before(effectiveReset) {
		grant.Reason = BudgetStale
		return grant
	}
	if now.Sub(obs.ObservedAt) >= budgetObservationMaxAge {
		grant.Reason = BudgetStale
		return grant
	}
	if obs.Limit < 0 || obs.Remaining < 0 || obs.Remaining > obs.Limit {
		grant.Reason = BudgetInvalidBounds
		return grant
	}
	grant.Reserve = SafetyReserve(obs.Remaining)
	discretionary := obs.Remaining - grant.Reserve
	if discretionary <= 0 || len(state.commitments) >= discretionary {
		grant.Reason = BudgetReserved
		return grant
	}
	grant.Available = discretionary - len(state.commitments)
	if class == PollAnalytics {
		analyticsLimit := discretionary / 2
		if countCommitments(state.commitments, PollAnalytics) >= analyticsLimit {
			grant.Reason = BudgetEntryPriority
			return grant
		}
	}
	if !c.entropyOK || state.generationExhausted || len(state.issued) >= maxIssuedCommitmentsPerGeneration {
		grant.Reason = BudgetTokenUnavailable
		return grant
	}
	var capability [32]byte
	if _, err := io.ReadFull(c.entropy, capability[:]); err != nil {
		c.entropyOK = false
		grant.Reason = BudgetTokenUnavailable
		return grant
	}
	if _, collision := state.issued[capability]; collision {
		grant.Reason = BudgetTokenUnavailable
		return grant
	}
	state.commitments[capability] = budgetCommitment{class: class, generation: state.generation}
	state.issued[capability] = struct{}{}
	c.endpoints[key] = state
	grant.Allowed = true
	grant.Reason = BudgetGranted
	grant.Commitment = CommitmentToken{
		coordinator: c.coordinatorID,
		capability:  capability,
		keyDigest:   sha256.Sum256([]byte(key)),
		class:       class,
		generation:  state.generation,
	}
	return grant
}

// Complete marks one request finished, but conservatively keeps it counted as
// consumed/unreconciled. An authoritative same-window request cycle begun after
// completion may reconcile it; a causally covered next trusted reset clears the
// generation.
// Callers use this for success, error, and cancellation alike. False means the
// token is forged, cross-scoped, already completed, or already reconciled.
func (c *BudgetCoordinator) Complete(key string, token CommitmentToken) bool {
	if c == nil || token == (CommitmentToken{}) {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, ok := c.endpoints[key]
	if !ok {
		return false
	}
	if token.coordinator != c.coordinatorID || token.keyDigest != sha256.Sum256([]byte(key)) ||
		token.generation != state.generation {
		return false
	}
	record, ok := state.commitments[token.capability]
	if !ok || record.completed || record.class != token.class || record.generation != token.generation {
		return false
	}
	if c.now == nil {
		return false
	}
	completedAt := c.now()
	if completedAt.IsZero() {
		return false
	}
	if c.completionSequence == ^uint64(0) {
		return false
	}
	c.completionSequence++
	record.completed = true
	record.completedAt = completedAt
	record.completionSequence = c.completionSequence
	state.commitments[token.capability] = record
	c.endpoints[key] = state
	return true
}

func sameBudgetProvenance(a, b official.RateBudget) bool {
	return a.Path == b.Path && a.Limit == b.Limit && a.Reset.Equal(b.Reset) && a.ResetRaw == b.ResetRaw &&
		a.ResetKind == b.ResetKind && a.ObservedAt.Equal(b.ObservedAt) && a.Reported == b.Reported
}

func trustedBudgetWindow(b official.RateBudget) bool {
	return b.Reported && !b.Reset.IsZero() && !b.ObservedAt.IsZero() && b.Reset.After(b.ObservedAt) && b.Limit >= 0 && b.Remaining >= 0 &&
		b.Remaining <= b.Limit && validResetSemantics(b)
}

func reconcileCompleted(commitments map[[32]byte]budgetCommitment, cycle *observationCycleRecord) {
	if cycle == nil {
		return
	}
	for capability, commitment := range commitments {
		if commitment.completed && commitment.completionSequence != 0 && commitment.completionSequence <= cycle.completionWatermark {
			delete(commitments, capability)
		}
	}
}

type budgetWindowRelation uint8

const (
	budgetWindowInitial budgetWindowRelation = iota
	budgetWindowSame
	budgetWindowNext
	budgetWindowConflict
)

func newEndpointBudget() endpointBudget {
	return endpointBudget{
		commitments:             make(map[[32]byte]budgetCommitment),
		issued:                  make(map[[32]byte]struct{}),
		observationCycles:       make(map[[32]byte]observationCycleRecord),
		issuedObservationCycles: make(map[[32]byte]struct{}),
		generation:              1,
	}
}

func validResetSemantics(b official.RateBudget) bool {
	raw, reset, kind := official.ParseRateBudgetReset(b.ResetRaw, b.ObservedAt)
	return raw == b.ResetRaw && kind == b.ResetKind &&
		(kind == official.ResetEpoch || kind == official.ResetDelta) && reset.Equal(b.Reset)
}

func classifyBudgetWindow(state endpointBudget, observation official.RateBudget) budgetWindowRelation {
	if !state.hasTrustedReset {
		return budgetWindowInitial
	}
	if observation.ResetKind != state.trustedResetKind {
		return budgetWindowConflict
	}
	switch observation.ResetKind {
	case official.ResetEpoch:
		if observation.Reset.Equal(state.trustedResetAnchor) {
			return budgetWindowSame
		}
		if !observation.ObservedAt.Before(state.trustedResetAnchor) && observation.Reset.After(state.trustedResetAnchor) {
			return budgetWindowNext
		}
	case official.ResetDelta:
		latestPlausibleBoundary := state.trustedResetAnchor.Add(deltaResetTolerance)
		if !observation.ObservedAt.Before(latestPlausibleBoundary) && observation.Reset.After(latestPlausibleBoundary) {
			return budgetWindowNext
		}
		earliestPlausibleBoundary := state.trustedResetAnchor.Add(-deltaResetTolerance)
		if observation.ObservedAt.Before(latestPlausibleBoundary) &&
			!observation.Reset.Before(earliestPlausibleBoundary) && !observation.Reset.After(latestPlausibleBoundary) {
			return budgetWindowSame
		}
	}
	return budgetWindowConflict
}

func sameTrustedWindow(state endpointBudget, observation official.RateBudget) bool {
	return classifyBudgetWindow(state, observation) == budgetWindowSame
}

func setTrustedReset(state *endpointBudget, observation official.RateBudget) {
	state.trustedReset = observation.Reset
	state.trustedResetAnchor = observation.Reset
	state.trustedResetKind = observation.ResetKind
	state.hasTrustedReset = true
}

func retainEarliestDeltaReset(state *endpointBudget, observation official.RateBudget) {
	if observation.ResetKind != official.ResetDelta {
		return
	}
	if observation.Reset.Before(state.trustedReset) {
		state.trustedReset = observation.Reset
	}
}

func advanceBudgetGeneration(state *endpointBudget, observation official.RateBudget) {
	state.commitments = make(map[[32]byte]budgetCommitment)
	state.issued = make(map[[32]byte]struct{})
	state.observationCycles = make(map[[32]byte]observationCycleRecord)
	state.issuedObservationCycles = make(map[[32]byte]struct{})
	setTrustedReset(state, observation)
	if state.generation == ^uint64(0) {
		state.generationExhausted = true
		return
	}
	state.generation++
}

func countCommitments(commitments map[[32]byte]budgetCommitment, class PollClass) int {
	count := 0
	for _, current := range commitments {
		if current.class == class {
			count++
		}
	}
	return count
}

func isKnownPollClass(class PollClass) bool {
	switch class {
	case PollCandidate, PollEntry, PollAnalytics, PollEmergencyExit, PollReconcile, PollFillDetection, PollProtection:
		return true
	default:
		return false
	}
}

func isSafetyClass(class PollClass) bool {
	switch class {
	case PollEmergencyExit, PollReconcile, PollFillDetection, PollProtection:
		return true
	default:
		return false
	}
}
