package execgw_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// observation_test.go is change add-net-rr-measurement tasks 4.1, 4.1b, 4.2 and
// 4.5: the Guardian records what each entry verdict measured, and doing so cannot
// cost a trade.

// recordingObserver captures rows synchronously. Used where the assertion is
// *what* was recorded; the asynchronous contract is tested separately with the
// real AsyncObserver.
type recordingObserver struct {
	mu   sync.Mutex
	rows []journal.EntryObservation
}

func (o *recordingObserver) ObserveEntry(row journal.EntryObservation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.rows = append(o.rows, row)
}

func (o *recordingObserver) all() []journal.EntryObservation {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]journal.EntryObservation(nil), o.rows...)
}

func (o *recordingObserver) only(t *testing.T) journal.EntryObservation {
	t.Helper()
	rows := o.all()
	if len(rows) != 1 {
		t.Fatalf("observations = %d, want exactly 1: %+v", len(rows), rows)
	}
	return rows[0]
}

// TestARefusalIsRecordedForTheFirstTime is task 4.1's larger half. Before this
// change a chain refusal was an IssueRefusal built in memory and returned; nothing
// wrote it down, so the question "what would a tighter gate have refused?" had no
// answer on disk.
func TestARefusalIsRecordedForTheFirstTime(t *testing.T) {
	observer := &recordingObserver{}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })

	intent := guardianIntent()
	intent.TargetPrice = "70500" // gross RR 0.5, under the 2.0 minimum

	_, err := rig.guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
		Intent:  intent,
		Account: guardianAccount(),
		Collect: rig.collect,
	})
	if err == nil {
		t.Fatal("the fixture must be refused")
	}

	row := observer.only(t)
	if row.Outcome != journal.OutcomeRefusedChain {
		t.Errorf("outcome = %q, want REFUSED_CHAIN", row.Outcome)
	}
	if row.StoppedStep != "min_reward_risk" {
		t.Errorf("stopped step = %q, want the rung that refused", row.StoppedStep)
	}
	if row.ReasonCode != string(risk.ReasonMinRRNotMet) {
		t.Errorf("reason = %q, want MIN_RR_NOT_MET", row.ReasonCode)
	}
	if row.DecisionID != "" {
		t.Errorf("a chain refusal writes no decision, so it references none: %q", row.DecisionID)
	}

	// The measurement itself: both ratios, the break-even, and the geometry that
	// produced them. This row is the whole record of a judgement that used to
	// vanish.
	if row.GrossRewardRisk != "0.5" {
		t.Errorf("gross = %q, want 0.5", row.GrossRewardRisk)
	}
	if row.NetRewardRisk == "" || row.BreakEvenPrice == "" {
		t.Errorf("the refusal must still carry what was measurable: %+v", row)
	}
	if row.EntryPrice != "70000" || row.StopPrice != "69000" || row.TargetPrice != "70500" {
		t.Errorf("the geometry must be recorded verbatim: %+v", row)
	}
	if row.Symbol != "005930" || row.Market != "kr" || row.AccountRef != "acct-7" {
		t.Errorf("the venue must be recorded: %+v", row)
	}
}

// TestAnIssuedVerdictIsRecordedWithItsDecision is task 4.1's other half, and the
// ordering claim: the observation carries the committed decision's id and its
// issued_at, which only exist after the transaction committed.
func TestAnIssuedVerdictIsRecordedWithItsDecision(t *testing.T) {
	observer := &recordingObserver{}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Observer = observer
		o.NewID = fixedIDs("decision-1", "nonce-1", "obs-1")
	})

	issued, err := rig.issue(context.Background())
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}

	row := observer.only(t)
	if row.Outcome != journal.OutcomeAllowedIssued {
		t.Errorf("outcome = %q, want ALLOWED_ISSUED", row.Outcome)
	}
	if row.DecisionID != issued.Decision.ID {
		t.Errorf("decision reference = %q, want %q", row.DecisionID, issued.Decision.ID)
	}
	if row.IssuedAt.IsZero() {
		t.Error("an issued observation carries the decision's issued_at")
	}
	if row.StoppedStep != "" || row.ReasonCode != "" {
		t.Errorf("an allowed verdict stopped at no rung: %+v", row)
	}
	if row.Reconstructed {
		t.Error("a row the verdict wrote is not a reconstruction")
	}
	// It really is a decision the gateway can read — so the observation is
	// referencing a committed row, not one that rolled back.
	if !decisionOnDisk(t, rig.journal, issued.Decision.ID) {
		t.Error("the referenced decision must be on disk")
	}
}

// TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome is task 4.2 (R2-5). The
// chain allowed; the ledger refused. Recording that as a chain refusal would
// attribute a full account to the intent's geometry, and the resulting distribution
// would say the geometry is worse than it is.
func TestChainAllowFollowedByIssuanceRefusalIsItsOwnOutcome(t *testing.T) {
	observer := &recordingObserver{}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })
	ctx := context.Background()

	// Fill the exposure limit with a first entry, so the second is refused by the
	// reservation rather than by the chain.
	if _, err := rig.issue(ctx); err != nil {
		t.Fatalf("the first issuance must succeed: %v", err)
	}
	observer.mu.Lock()
	observer.rows = nil
	observer.mu.Unlock()

	// A collector reporting the account already at its exposure limit.
	atLimit := func(ctx context.Context, attempt int) (execgw.ExposureSnapshot, error) {
		snapshot, err := rig.collect(ctx, attempt)
		if err != nil {
			return snapshot, err
		}
		snapshot.OpenExposure = guardianPolicy().MaxOpenExposure
		return snapshot, nil
	}
	_, err := rig.guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent:  guardianIntent(),
		Account: guardianAccount(),
		Collect: atLimit,
	})
	if err == nil {
		t.Fatal("the second issuance must be refused by the ledger")
	}
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) || refusal.Stage != execgw.StageIssuance {
		t.Fatalf("the fixture must refuse at the issuance stage, got %v", err)
	}

	row := observer.only(t)
	if row.Outcome != journal.OutcomeAllowedIssuanceRefused {
		t.Fatalf("outcome = %q, want ALLOWED_ISSUANCE_REFUSED", row.Outcome)
	}
	if row.IssuanceReasonCode != refusal.Reason {
		t.Errorf("issuance reason = %q, want the ledger's own %q",
			row.IssuanceReasonCode, refusal.Reason)
	}
	if row.IssuanceReasonCode != journal.IssueReasonLimitReached {
		t.Errorf("issuance reason = %q, want LIMIT_REACHED for a full account",
			row.IssuanceReasonCode)
	}
	// The distinction that matters: not filed as a geometry refusal.
	if row.StoppedStep != "" || row.ReasonCode != "" {
		t.Errorf("the chain allowed, so no rung stopped it: %+v", row)
	}
	if row.DecisionID != "" {
		t.Error("a refused issuance rolled its decision back; there is none to reference")
	}
	// And it was still measured — the geometry is exactly what a threshold study
	// wants from an entry the account had no room for.
	if row.GrossRewardRisk == "" || row.NetRewardRisk == "" {
		t.Errorf("a refused issuance is still a measured intent: %+v", row)
	}
}

// TestTheCostBasisIsRecordedWithEveryRow is task 4.5 and 4.4's consumer side: the
// scope and the fingerprint travel with the numbers, so a ratio computed under the
// `[미검증]` placeholders can never be read as a measured one.
func TestTheCostBasisIsRecordedWithEveryRow(t *testing.T) {
	observer := &recordingObserver{}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })

	if _, err := rig.issue(context.Background()); err != nil {
		t.Fatal(err)
	}
	row := observer.only(t)

	if row.CostScope != journal.CostScopeFeeTaxOnly {
		t.Errorf("cost scope = %q, want FEE_TAX_ONLY — slippage is not in this number, "+
			"which is why the metric is 수수료·세금 차감 후 RR", row.CostScope)
	}
	if want := costs.DefaultModel().Fingerprint(); row.CostModelFingerprint != want {
		t.Errorf("fingerprint = %q, want the model the Guardian was built with (%q)",
			row.CostModelFingerprint, want)
	}
	if row.CostModelFingerprint == costs.FingerprintUnconfigured {
		t.Error("a Guardian cannot be built without a cost model, so its rows are never " +
			"unfingerprinted")
	}
}

// TestADifferentRateSetIsADifferentFingerprint: the field earns its place only if
// two Guardians on different rates produce rows an aggregate can tell apart.
func TestADifferentRateSetIsADifferentFingerprint(t *testing.T) {
	cheaper, err := costs.NewModel(map[string]string{costs.KeyKRSellTaxRate: "0.0018"})
	if err != nil {
		t.Fatal(err)
	}

	defaults := &recordingObserver{}
	rigA := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = defaults })
	if _, err := rigA.issue(context.Background()); err != nil {
		t.Fatal(err)
	}

	measured := &recordingObserver{}
	rigB := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Observer = measured
		o.Costs = cheaper
	})
	if _, err := rigB.issue(context.Background()); err != nil {
		t.Fatal(err)
	}

	a, b := defaults.only(t), measured.only(t)
	if a.CostModelFingerprint == b.CostModelFingerprint {
		t.Fatal("two rate sets produced one fingerprint; observations from before and after " +
			"the 2b rate measurement would aggregate as if they were comparable")
	}
	if a.NetRewardRisk == b.NetRewardRisk {
		t.Errorf("the two models must also produce different net ratios: %q and %q",
			a.NetRewardRisk, b.NetRewardRisk)
	}
}

// --- 4.1b: the observation is not in front of the submission -----------------

// slowObserver blocks for as long as it is told to, which is how "the recording
// must not be a synchronous wait point" is made into a measurable claim.
type slowObserver struct {
	delay   time.Duration
	entered chan struct{}
	once    sync.Once
}

func (o *slowObserver) ObserveEntry(journal.EntryObservation) {
	o.once.Do(func() { close(o.entered) })
	time.Sleep(o.delay)
}

// TestASlowObservationDoesNotDelayTheIssuance is task 4.1b (round 3's P1).
//
// The Guardian returns and the caller submits. If the observation were awaited
// here, a slow journal would sit in front of an authorised order and eat the
// 60-second decision TTL — losing an entry the chain allowed, for a measurement.
//
// The claim is timed rather than described: an observer that blocks far longer
// than the issuance takes must not lengthen the issuance.
func TestASlowObservationDoesNotDelayTheIssuance(t *testing.T) {
	const observerDelay = 2 * time.Second

	baseline := newGuardian(t, nil)
	start := time.Now()
	if _, err := baseline.issue(context.Background()); err != nil {
		t.Fatal(err)
	}
	unobserved := time.Since(start)

	slow := &slowObserver{delay: observerDelay, entered: make(chan struct{})}
	observed := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Observer = execgw.NewAsyncObserver(execgw.AsyncObserverOptions{
			Sink: blockingSink{observer: slow},
		})
	})
	start = time.Now()
	issued, err := observed.issue(context.Background())
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}

	// The write is genuinely in flight, so the test is not passing because nothing
	// happened.
	select {
	case <-slow.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the observation write never started; the fixture proves nothing")
	}

	if elapsed >= observerDelay {
		t.Fatalf("the issuance took %s with an observer that blocks for %s (%s without one). "+
			"The recording is a synchronous wait point, and a slow journal would eat the "+
			"decision TTL in front of an authorised order", elapsed, observerDelay, unobserved)
	}

	// And the decision is fully usable while the observation is still being
	// written: it is on disk, unexpired, with its reservation held.
	if !decisionOnDisk(t, observed.journal, issued.Decision.ID) {
		t.Error("the decision must be readable by the gateway immediately")
	}
	if len(issued.Reservations) != 1 {
		t.Errorf("the HELD reservation must be returned with the issuance, got %d",
			len(issued.Reservations))
	}
	if remaining := issued.ExpiresAt.Sub(observed.clock.Now()); remaining <= 0 {
		t.Errorf("the decision has %s of TTL left; the observation must not have consumed it",
			remaining)
	}
}

// blockingSink turns an EntryObserver into an ObservationSink, so the slow
// observer can sit where the journal write goes.
type blockingSink struct{ observer execgw.EntryObserver }

func (s blockingSink) RecordEntryObservation(_ context.Context, row journal.EntryObservation) error {
	s.observer.ObserveEntry(row)
	return nil
}

// countingLosses is the loss counter's interface, captured.
type countingLosses struct {
	mu    sync.Mutex
	kinds []string
}

func (c *countingLosses) Record(kind string, _ ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.kinds = append(c.kinds, kind)
}

func (c *countingLosses) all() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.kinds...)
}

// failingSink refuses every write, the way a full disk would.
type failingSink struct{}

func (failingSink) RecordEntryObservation(context.Context, journal.EntryObservation) error {
	return errors.New("the observation store is full")
}

// TestAFailedObservationDoesNotChangeTheVerdict is task 4.1's safety half and 2.7
// seen from the caller: the write fails, the issuance succeeds, and the loss is
// counted rather than reported as a refusal.
func TestAFailedObservationDoesNotChangeTheVerdict(t *testing.T) {
	losses := &countingLosses{}
	observer := execgw.NewAsyncObserver(execgw.AsyncObserverOptions{
		Sink:   failingSink{},
		Losses: losses,
	})
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })

	issued, err := rig.issue(context.Background())
	if err != nil {
		t.Fatalf("a failed observation must not refuse the issuance: %v", err)
	}
	observer.Close() // drains, so the write has definitely been attempted

	if !decisionOnDisk(t, rig.journal, issued.Decision.ID) {
		t.Error("the decision must survive its observation failing")
	}
	kinds := losses.all()
	if len(kinds) != 1 {
		t.Fatalf("losses = %v, want exactly one", kinds)
	}
	if kinds[0] != "observation_write_failed" {
		t.Errorf("an issued verdict's lost observation is rebuildable from the preimage and "+
			"must be counted as such, got %q", kinds[0])
	}
}

// TestALostRefusalIsCountedAsUnrecoverable is design D6's asymmetry at the wiring
// level. A refusal has no decision and therefore no preimage: nothing will ever
// rebuild it, and the counter has to say so.
func TestALostRefusalIsCountedAsUnrecoverable(t *testing.T) {
	losses := &countingLosses{}
	observer := execgw.NewAsyncObserver(execgw.AsyncObserverOptions{
		Sink:   failingSink{},
		Losses: losses,
	})
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.Observer = observer })

	intent := guardianIntent()
	intent.TargetPrice = "70500"
	if _, err := rig.guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
		Intent:  intent,
		Account: guardianAccount(),
		Collect: rig.collect,
	}); err == nil {
		t.Fatal("the fixture must be refused")
	}
	observer.Close()

	kinds := losses.all()
	if len(kinds) != 1 || kinds[0] != "refusal_observation_lost" {
		t.Fatalf("losses = %v, want one refusal_observation_lost — a refusal has no preimage "+
			"to rebuild from, and counting it as rebuildable would overstate what is recoverable",
			kinds)
	}
}

// TestAFullQueueDropsRatherThanBlocks: the non-blocking contract has to hold under
// back-pressure too, or an analysis backlog becomes an execution delay.
func TestAFullQueueDropsRatherThanBlocks(t *testing.T) {
	losses := &countingLosses{}
	release := make(chan struct{})
	observer := execgw.NewAsyncObserver(execgw.AsyncObserverOptions{
		Sink:   gatedSink{release: release},
		Losses: losses,
		Depth:  1,
	})
	t.Cleanup(func() {
		close(release)
		observer.Close()
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			observer.ObserveEntry(journal.EntryObservation{
				ID: "obs", Outcome: journal.OutcomeAllowedIssued, DecisionID: "d",
			})
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ObserveEntry blocked on a full queue; the entry path would be waiting " +
			"on the analysis path")
	}
	if len(losses.all()) == 0 {
		t.Error("a dropped observation is still a lost measurement and must be counted")
	}
}

// gatedSink blocks until released, so the queue behind it fills.
type gatedSink struct{ release chan struct{} }

func (s gatedSink) RecordEntryObservation(context.Context, journal.EntryObservation) error {
	<-s.release
	return nil
}

// TestAGuardianWithNoObserverIssuesAsBefore: the recording is optional wiring, so
// a Guardian built without one behaves exactly as it did before this change. That
// is what makes the measurement non-load-bearing rather than merely non-blocking.
func TestAGuardianWithNoObserverIssuesAsBefore(t *testing.T) {
	rig := newGuardian(t, nil)
	issued, err := rig.issue(context.Background())
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}
	if !decisionOnDisk(t, rig.journal, issued.Decision.ID) {
		t.Error("the decision must be on disk")
	}

	intent := guardianIntent()
	intent.TargetPrice = "70500"
	_, err = rig.guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
		Intent:  intent,
		Account: guardianAccount(),
		Collect: rig.collect,
	})
	if err == nil || !strings.Contains(err.Error(), string(risk.ReasonMinRRNotMet)) {
		t.Errorf("the refusal must be unchanged: %v", err)
	}
}

// TestObservationsReachTheJournalEndToEnd wires the real sink, because everything
// above uses a substitute for one half or the other and none of them shows a row
// arriving in the table this change added.
func TestObservationsReachTheJournalEndToEnd(t *testing.T) {
	rig := newGuardian(t, nil)
	ctx := context.Background()
	observer := execgw.NewAsyncObserver(execgw.AsyncObserverOptions{Sink: rig.journal})
	observed := newGuardianOn(t, rig, observer)

	if _, err := observed.IssueEntry(ctx, execgw.EntryIssuance{
		Intent:  guardianIntent(),
		Account: guardianAccount(),
		Collect: rig.collect,
	}); err != nil {
		t.Fatalf("the allowed issuance: %v", err)
	}
	refused := guardianIntent()
	refused.TargetPrice = "70500"
	if _, err := observed.IssueEntry(ctx, execgw.EntryIssuance{
		Intent:  refused,
		Account: guardianAccount(),
		Collect: rig.collect,
	}); err == nil {
		t.Fatal("the second intent must be refused")
	}
	observer.Close()

	rows, err := rig.journal.EntryObservations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want one allowed and one refused: %+v", len(rows), rows)
	}
	outcomes := map[string]bool{}
	for _, row := range rows {
		outcomes[row.Outcome] = true
		if row.CostScope != journal.CostScopeFeeTaxOnly || row.CostModelFingerprint == "" {
			t.Errorf("every stored row carries its cost basis: %+v", row)
		}
	}
	if !outcomes[journal.OutcomeAllowedIssued] || !outcomes[journal.OutcomeRefusedChain] {
		t.Errorf("both populations must be on disk, got %v", outcomes)
	}
}

// newGuardianOn builds a second Guardian over the same journal and clock, so a
// test can add an observer without rebuilding the rig.
func newGuardianOn(t *testing.T, rig *guardianRig, observer execgw.EntryObserver) *execgw.RiskGuardian {
	t.Helper()
	g, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal:       rig.journal,
		Clock:         rig.clock,
		AccountRef:    "acct-7",
		Policy:        guardianPolicy(),
		Costs:         costs.DefaultModel(),
		PolicyVersion: "add-core-domain/4.1",
		Observer:      observer,
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	return g
}
