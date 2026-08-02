package console

// portfolio_pages.go is the two dashboard handlers (change
// add-operator-dashboard, task 2.3).
//
// Both are GET readings and neither is registered behind the CSRF gate, because
// there is nothing on either page to submit. That is asserted from both
// directions in static_test.go: a state-changing route without the gate fails,
// and a read route behind it fails too — a screen that only answered POSTs would
// be a screen nobody can open.
//
// The positions screen reloads itself at the holdings cache TTL — no faster
// (change refresh-positions-screen; spec "rate budget 보호"). Each reload is an
// ordinary lazy request, so an open tab costs at most one holdings call per TTL
// and a verification in progress still suspends the broker call entirely. The
// history screen renders frozen values and stays manual.

import (
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

type positionsPage struct {
	Nav  string
	Snap positionsView
}

// Refresh is what the head template reads: the positions screen asks the
// browser to reload, at the period RefreshSeconds names.
func (positionsPage) Refresh() bool { return true }

// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	page := positionsPage{Nav: "positions", Snap: c.positions(r.Context())}
	var (
		runtime          positionpolicy.ManagementRuntime
		runtimeAttempted bool
		policyByID       map[string]positionpolicy.State
	)
	if c.opts.PositionPolicies != nil {
		runtimeAttempted = true
		// A read failure intentionally leaves EffectiveKnown false. Desired config
		// below remains useful display context, but is never substituted for the
		// running engine snapshot.
		runtime, _ = c.opts.PositionPolicies.Runtime(r.Context())
		if states, err := c.opts.PositionPolicies.List(r.Context()); err == nil {
			policyByID = make(map[string]positionpolicy.State, len(states))
			for _, state := range states {
				policyByID[strings.TrimSpace(state.PositionID)] = state
			}
		}
	}
	if c.opts.Settings != nil {
		if block, _, err := c.opts.Settings.Load(); err == nil {
			// One Load stamps both lists: two reads could return two different
			// snapshots and draw a row that is on neither or on both.
			for i := range page.Snap.Rows {
				page.Snap.Rows[i].Designated = block.Included(page.Snap.Rows[i].Symbol)
				page.Snap.Rows[i].Excluded = block.Excludes(page.Snap.Rows[i].Symbol)
			}
		}
	}
	if runtimeAttempted {
		for i := range page.Snap.Rows {
			row := &page.Snap.Rows[i]
			journalKnown := row.JournalReadable
			released := false
			if row.InJournal {
				row.LifecycleProofRequired = true
				state, ok := policyByID[strings.TrimSpace(row.PositionID)]
				if !ok {
					journalKnown = false
				} else {
					row.LifecycleKnown, row.LifecycleStatus = true, state.Status
					row.LifecycleGeneration = state.AdoptionGeneration
					released = state.Status == positionpolicy.StatusReleased && state.Version > 0
				}
			}
			row.Management = positionpolicy.ProjectManagement(positionpolicy.ManagementInput{
				Market: row.Market, Symbol: row.Symbol, JournalKnown: journalKnown,
				Managed: row.Managed(), Released: released, Runtime: runtime,
			})
			if row.Management.Block != nil {
				row.ManagementBlock = newReconcileBlockView(*row.Management.Block, c.now())
			}
		}
	}
	attachPositionExitLines(page.Snap.Rows, c.now(), runtime)
	c.render(w, "positions", page)
}

// attachPositionExitLines selects the already-persisted effective snapshot for
// display. Freshness is evaluated at the screen boundary; no exit value is
// recomputed here or in the template.
func attachPositionExitLines(rows []positionRow, asOf time.Time, runtime positionpolicy.ManagementRuntime) {
	for i := range rows {
		row := &rows[i]
		if !row.HasExit {
			row.ExitReference = operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
				Market: row.Market, ManagementStatus: string(row.Management.Status),
				EffectiveSettingsKnown: runtime.EffectiveKnown,
				EffectiveStopPct:       runtime.Effective.DefaultStopPct,
			})
			continue
		}
		raw := storedExitReference(row.Exit)
		if row.LifecycleKnown && row.Exit.LifecycleGeneration > 0 && row.LifecycleGeneration > 0 &&
			row.Exit.LifecycleGeneration != row.LifecycleGeneration {
			row.StoredExit = storedExitEvidenceView{}
			row.ExitLine = operatorview.BuildExitLine(operatorview.Source{
				UnknownReason: "lifecycle_generation_mismatch",
			})
			row.ExitReference = operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
				Market: row.Market, EvidencePresent: true,
				EvidenceLifecycleGeneration: row.Exit.LifecycleGeneration, Raw: raw,
				LifecycleKnown: true, LifecycleProofRequired: row.LifecycleProofRequired,
				CurrentLifecycleGeneration: row.LifecycleGeneration,
			})
			continue
		}
		if row.LifecycleKnown && row.LifecycleStatus == positionpolicy.StatusReleased {
			row.ExitLine = operatorview.BuildExitLine(operatorview.Source{
				UnknownReason: "operator_released_lifecycle",
			})
			row.ExitReference = operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
				Market: row.Market, EvidencePresent: true,
				EvidenceLifecycleGeneration: row.Exit.LifecycleGeneration, Raw: raw,
				LifecycleKnown: row.LifecycleKnown, LifecycleProofRequired: row.LifecycleProofRequired,
				CurrentLifecycleGeneration: row.LifecycleGeneration,
				Released:                   true, UnknownReason: row.Exit.Snapshot.UnknownReason,
			})
			if row.ExitReference.LegacyRaw() {
				row.StoredExit = storedExitEvidence(row.Exit)
			}
			continue
		}
		snapshot := row.Exit.Snapshot.WithFreshness(asOf, holdingsTTL)
		source := operatorview.Source{
			UnknownReason:     snapshot.UnknownReason,
			StaleReason:       snapshot.StaleReason,
			RemainingQuantity: row.JournalQuantity,
			EffectiveSource:   "persisted effective snapshot",
		}
		if snapshot.Snapshot != nil {
			source.Snapshot = &snapshot.Snapshot.Line
			source.ObservationSource = snapshot.Snapshot.ObservationSource
			source.ObservedAt = snapshot.Snapshot.ObservedAt
		}
		row.ExitLine = operatorview.BuildExitLine(source)
		row.ExitReference = operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
			Market: row.Market, CanonicalSnapshotPresent: snapshot.Snapshot != nil,
			CanonicalSnapshotStale: snapshot.Stale, EvidencePresent: true,
			EvidenceLifecycleGeneration: row.Exit.LifecycleGeneration, Raw: raw,
			LifecycleKnown: row.LifecycleKnown, LifecycleProofRequired: row.LifecycleProofRequired,
			CurrentLifecycleGeneration: row.LifecycleGeneration,
			ManagementStatus:           string(row.Management.Status), EffectiveSettingsKnown: runtime.EffectiveKnown,
			EffectiveStopPct: runtime.Effective.DefaultStopPct, UnknownReason: snapshot.UnknownReason,
		})
		if row.ExitReference.LifecycleUnknown() {
			row.StoredExit = storedExitEvidenceView{}
			row.ExitLine = operatorview.BuildExitLine(operatorview.Source{
				UnknownReason: "lifecycle_generation_unverified",
			})
			continue
		}
		if row.ExitReference.LegacyRaw() {
			row.StoredExit = storedExitEvidence(row.Exit)
		}
	}
}

func hasStoredExitEvidence(state journal.ExitState) bool {
	return strings.TrimSpace(state.EntryPrice) != "" || strings.TrimSpace(state.InitialStop) != "" ||
		strings.TrimSpace(state.Baseline) != "" || strings.TrimSpace(state.HighWater) != ""
}

func storedExitValue(value string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return "—"
}

type historyPage struct {
	Nav  string
	Snap historyView
}

func (historyPage) Refresh() bool { return false }

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{Nav: "history", Snap: c.history(r.Context())})
}
