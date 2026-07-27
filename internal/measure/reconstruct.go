// Package measure holds the measurement half of add-net-rr-measurement: the
// gap-closing job that rebuilds observations a crash lost, and the counterfactual
// harness that produces the boundary map a later threshold will be chosen from.
//
// The degradation counter is deliberately *not* here — it is
// internal/measure/degrade, a package with no journal dependency at all, because
// a counter of storage failures must not share a failure domain with the storage.
package measure

// reconstruct.go is the gap-closing half of design D6 (change
// add-net-rr-measurement tasks 2.5, 2.5b, 2.5c, 2.5d).
//
// # What this job is, and what it must never become
//
// Observations are written outside the issuance transaction, because inside it a
// measurement failure would roll a decision back and refuse a trade the chain
// allowed. The price of that is a crash window: a decision commits, the process
// dies, and the observation never lands. This job closes that window by finding
// those decisions and rebuilding the row from the preimage.
//
// It is emphatically **not** a supervised engine loop. The runtime's supervisor
// treats a loop's return as a reason to stop every other loop, so a fault in this
// measurement job would take the exit observer and the fill detector down with it
// (§0.3). Its consecutive-failure threshold escalates to ENTRY_BLOCKED, and its
// trigger vocabulary is a closed enumeration — so a measurement job registered
// there would have to *borrow* somebody else's trigger and misattribute its own
// failure as a reconciliation outage.
//
// Hence: a plain function, run from a separate process or schedule boundary, that
// returns a report instead of an error class anything escalates on.
// reconstruct_isolation_test.go is what keeps it that way.
//
// # Reconstructed is not restored
//
// The preimage carries the three prices, the venue, the size and the policy
// version, so those come back exactly. It does not carry the rate set the original
// break-even was computed under — nothing does — so the net ratio this job writes
// is a *new measurement under today's model* wearing an old decision's timestamp.
// That is why a rebuilt row is marked, carries both instants, and carries the
// fingerprint of the model used to rebuild it. A consumer that averages rebuilt
// and live rows without splitting on the marker is mixing two rate sets, and the
// marker is what makes that its mistake rather than an invisible one.
//
// Target price is a special case worth naming: RiskIntent.Canonical() treats
// target_price as optional (canonicalDecimal with required=false), so the preimage
// contract alone does not guarantee one is there. It is there in practice because
// the minimum-reward-risk rung refuses an entry whose ratio cannot be computed,
// and no target means no ratio. The reconstruction depends on that rung and says
// so — if the rung is ever relaxed, targets start arriving empty here.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/measure/degrade"
)

// Store is the journal surface this job uses. It is an interface so the job can be
// tested against a fake, and so that the dependency reads in one direction: the
// journal knows nothing about this package.
type Store interface {
	DetectMissingEntryObservations(ctx context.Context, now time.Time,
		opts journal.GapScanOptions) (journal.EntryObservationGap, error)
	RecordEntryObservation(ctx context.Context, obs journal.EntryObservation) error
}

// Ratios is what a caller computes for one rebuilt row. The arithmetic lives in
// internal/risk; this package does not do cost maths, so that there is exactly one
// implementation of "실질본전" in the tree.
type Ratios struct {
	BreakEvenPrice  string
	GrossRewardRisk string
	NetRewardRisk   string
	// Fingerprint identifies the rate set used *now*, not the one used then.
	Fingerprint string
}

// Recompute produces the ratios for one missing observation. It returns an error
// when the intent cannot be measured today; the row is still written, with the
// unmeasurable fields empty, because "this entry was issued and we cannot say what
// its net ratio was" is itself a fact worth having.
type Recompute func(missing journal.MissingObservation) (Ratios, error)

// RepairReport is one reconstruction run's outcome. It is a value, not an error
// class: nothing escalates on it, and the caller decides what to do with the
// numbers.
type RepairReport struct {
	// Found is how many gaps the scan offered.
	Found int
	// Rebuilt is how many rows were written.
	Rebuilt int
	// AlreadyPresent is how many lost the race to a real write that landed after
	// the deadline. Not a failure — the unique index doing its job (task 2.5c).
	AlreadyPresent int
	// Unmeasurable is how many were written with empty ratios because the cost
	// model could not price them today.
	Unmeasurable int
	// Failed is how many could not be written at all.
	Failed int
	// LapsedBeyondHorizon is what the scan reported as past rebuilding.
	LapsedBeyondHorizon int
}

// Run scans for gaps and rebuilds what it can.
//
// It returns an error only for a scan that could not run (a misconfigured
// schedule, a storage failure on the read). Individual rebuild failures are
// counted in the report and recorded on the degradation counter, because one
// unwritable row is not a reason to abandon the rest.
func Run(
	ctx context.Context,
	store Store,
	now time.Time,
	opts journal.GapScanOptions,
	recompute Recompute,
	idFor func(journal.MissingObservation) string,
	counter *degrade.Counter,
) (RepairReport, error) {
	gap, err := store.DetectMissingEntryObservations(ctx, now, opts)
	if err != nil {
		return RepairReport{}, fmt.Errorf("measure: scanning for missing entry observations: %w", err)
	}
	report := RepairReport{Found: len(gap.Missing), LapsedBeyondHorizon: gap.LapsedBeyondHorizon}

	for i := 0; i < gap.LapsedBeyondHorizon; i++ {
		counter.Record(degrade.LossLapsedBeyondHorizon)
	}

	for _, missing := range gap.Missing {
		ratios, rerr := recompute(missing)
		if rerr != nil {
			report.Unmeasurable++
			ratios.BreakEvenPrice, ratios.GrossRewardRisk, ratios.NetRewardRisk = "", "", ""
		}
		obs := rebuild(missing, ratios, now, idFor(missing))
		switch err := store.RecordEntryObservation(ctx, obs); {
		case err == nil:
			report.Rebuilt++
		case errors.Is(err, journal.ErrObservationExists):
			// The real write landed after the deadline. Exactly what the unique
			// index is for: one decision keeps one row, and the distribution a
			// threshold gets drawn from is not double-counted.
			report.AlreadyPresent++
			if rerr != nil {
				report.Unmeasurable--
			}
		default:
			report.Failed++
			if rerr != nil {
				report.Unmeasurable--
			}
			counter.Record(degrade.LossObservationWrite,
				"decision_id", missing.DecisionID, "phase", "reconstruction", "error", err.Error())
		}
	}
	return report, nil
}

// rebuild assembles the row. Everything restored comes from the preimage;
// everything recomputed is stamped with today's fingerprint and marked.
func rebuild(missing journal.MissingObservation, ratios Ratios, now time.Time,
	id string) journal.EntryObservation {
	return journal.EntryObservation{
		ID:         id,
		AccountRef: missing.Preimage.AccountRef,
		Market:     missing.Preimage.Market,
		Symbol:     missing.Preimage.Symbol,
		// Restored exactly: the preimage is the hashed text the gateway
		// re-verifies against, so these are the values the decision was made on.
		EntryPrice:  missing.Preimage.EntryPrice,
		StopPrice:   missing.Preimage.StopPrice,
		TargetPrice: missing.Preimage.TargetPrice,
		// Recomputed, not restored. See the file header.
		BreakEvenPrice:       ratios.BreakEvenPrice,
		GrossRewardRisk:      ratios.GrossRewardRisk,
		NetRewardRisk:        ratios.NetRewardRisk,
		CostScope:            journal.CostScopeFeeTaxOnly,
		CostModelFingerprint: ratios.Fingerprint,
		// A committed EXPOSURE_RAISING decision exists, so the chain allowed and
		// the ledger issued. There is no other way for that row to be there.
		Outcome:    journal.OutcomeAllowedIssued,
		DecisionID: missing.DecisionID,
		// Both instants. The observation is "as of" the rebuild, and the verdict
		// it describes happened at issuance; a row carrying only one of them would
		// put this run's fingerprint next to that day's timestamp.
		ObservedAt:      now,
		IssuedAt:        missing.IssuedAt,
		Reconstructed:   true,
		ReconstructedAt: now,
	}
}
