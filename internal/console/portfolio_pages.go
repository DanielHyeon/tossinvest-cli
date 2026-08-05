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
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

type positionsPage struct {
	chrome
	Snap positionsView
}

// brokerStateView is the broker-cache banner's data plus this screen's fold
// state.
//
// The "brokerstate" template is invoked with the holdings snapshot as its dot,
// so `$` inside it is that snapshot and not the page — and the fold in it needs
// the page's explain state. Embedding keeps every existing `.Wired`, `.Held`,
// `.TakenAt` reference working and adds the one field.
type brokerStateView struct {
	holdingsSnapshot
	Explain explainState
}

// BrokerState is what the banner template is given.
func (p positionsPage) BrokerState() brokerStateView {
	return brokerStateView{holdingsSnapshot: p.Snap.Holdings, Explain: p.Explain}
}

// RefreshSeconds is the reload period: exactly the holdings cache TTL, derived
// from it so the two cannot drift apart — a period under the TTL would be a
// reload that costs broker calls faster than the budget the spec fixes.
//
// Refresh is the embedded chrome's field, set by the handler below.
func (positionsPage) RefreshSeconds() int { return int(holdingsTTL / time.Second) }

func (c *Console) handlePositions(w http.ResponseWriter, r *http.Request) {
	snap := c.positions(r.Context())
	page := positionsPage{
		chrome: c.chromeFor("positions", verifylive.MarketKR, snap.Holdings.freshness()),
		Snap:   snap,
	}
	page.Refresh = true
	// Same reason as the orders screen: a reloading screen cannot keep a native
	// fold open (change a055 §6, explain.go).
	page.Explain = explainFrom(r)
	c.decoratePositionRows(r.Context(), page.Snap.Rows, c.now())
	c.render(w, "positions", page)
}

// decoratePositionRows attaches the desired/effective/lifecycle and persisted
// exit evidence that both holdings screens render. Keeping this at one boundary
// is what prevents /dashboard and /positions from giving the same holding two
// different management or protection answers.
//
// The helper mutates request-local display rows only. Runtime/settings failures
// remain unknown, and attachPositionExitLines continues to suppress stale or
// mismatched actionable values. No broker call, journal write, order, or config
// save is available from this path.
func (c *Console) decoratePositionRows(ctx context.Context, rows []positionRow, asOf time.Time) {
	var (
		runtime          positionpolicy.ManagementRuntime
		runtimeAttempted bool
		policyByID       map[string]positionpolicy.State
	)
	if c.opts.PositionPolicies != nil {
		runtimeAttempted = true
		// One reading of the engine, shared by both screens and by every render
		// inside each half's interval (change a081, position_policy_cache.go).
		// Reading here per render put the number of open screens into the interval
		// between stop-loss judgements, because the List behind it runs on the
		// engine's single write connection.
		//
		// A read failure intentionally leaves EffectiveKnown false, and a lifecycle
		// listing the engine could not serve leaves policyByID nil. Desired config
		// below remains useful display context, but is never substituted for the
		// running engine snapshot — which is why the cache serves a failed attempt
		// as a failure rather than reviving the success before it.
		//
		// The listing is also allowed to be older than the journal row it is joined
		// against, and always was: c.positions() reads on every render. That
		// disagreement can only withhold a verdict — an unmatched row renders
		// 관리 여부 불명 — never invent one.
		reading := c.enginePolicy.read(ctx, asOf)
		runtime = reading.Runtime
		if reading.StatesErr == nil {
			policyByID = make(map[string]positionpolicy.State, len(reading.States))
			for _, state := range reading.States {
				policyByID[strings.TrimSpace(state.PositionID)] = state
			}
		}
	}
	if c.opts.Settings != nil {
		if block, _, err := c.opts.Settings.Load(); err == nil {
			// One Load stamps both lists: two reads could return two different
			// snapshots and draw a row that is on neither or on both.
			for i := range rows {
				rows[i].Designated = block.Included(rows[i].Symbol)
				rows[i].Excluded = block.Excludes(rows[i].Symbol)
			}
		}
	}
	if runtimeAttempted {
		for i := range rows {
			row := &rows[i]
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
				row.ManagementBlock = newReconcileBlockView(*row.Management.Block, asOf)
			}
		}
	}
	attachPositionExitLines(rows, asOf, runtime, c.protectionLiveness(asOf))
}

// attachPositionExitLines selects the already-persisted effective snapshot for
// display. Freshness is evaluated at the screen boundary; no exit value is
// recomputed here or in the template.
func attachPositionExitLines(rows []positionRow, asOf time.Time, runtime positionpolicy.ManagementRuntime,
	live protectionLiveness) {
	for i := range rows {
		row := &rows[i]
		referenceStatus := row.Management.Status
		if referenceStatus == "" && row.PendingDesignation() {
			// The desired include list proves only that this holding is awaiting
			// management. If the engine-owned commander is unavailable, keep the
			// effective state explicitly unknown; never promote desired defaults.
			referenceStatus = positionpolicy.ManagementStatusUnknown
		}
		if !row.HasExit {
			row.ExitReference = operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
				Market: row.Market, ManagementStatus: string(referenceStatus),
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
		snapshot := exitFreshness(row.Exit.Snapshot, asOf, live)
		if row.Quarantined() {
			// The exit loop refuses a quarantined position outright
			// (ErrExitSnapshotQuarantined), so nothing is maintaining this line
			// no matter how healthy the engine or the evidence looks. This
			// overrides the freshness verdict rather than joining it: an
			// operator whose stop is not being evaluated needs that sentence,
			// not a remark about the evaluation's age.
			snapshot.Stale, snapshot.StaleReason = true, "snapshot_quarantined"
			if snapshot.Snapshot == nil {
				snapshot.UnknownReason = "snapshot_quarantined"
			}
		}
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
			ManagementStatus:           string(referenceStatus), EffectiveSettingsKnown: runtime.EffectiveKnown,
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
	chrome
	Snap historyView
}

func (c *Console) handleHistory(w http.ResponseWriter, r *http.Request) {
	c.render(w, "history", historyPage{
		chrome: c.chromeOnRequest("history"),
		Snap:   c.history(r.Context()),
	})
}
