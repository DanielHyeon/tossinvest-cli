package main

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func applyExitLineReference(out *httpapi.Position, stored *journal.PositionExit,
	state positionpolicy.State, lifecycleKnown bool, runtime positionpolicy.ManagementRuntime) {
	var (
		raw              *operatorview.StoredExitReference
		canonicalPresent bool
		evidencePresent  bool
		evidenceGen      int64
		unknownReason    string
	)
	if stored != nil && stored.HasExit {
		evidencePresent = true
		evidenceGen = stored.Exit.LifecycleGeneration
		canonicalPresent = stored.Exit.Snapshot.Snapshot != nil
		unknownReason = stored.Exit.Snapshot.UnknownReason
		if hasStoredExitEvidence(stored.Exit) {
			raw = &operatorview.StoredExitReference{
				EntryPrice: stored.Exit.EntryPrice, InitialStop: stored.Exit.InitialStop,
				Baseline: stored.Exit.Baseline, HighWater: stored.Exit.HighWater,
				LifecycleGeneration: stored.Exit.LifecycleGeneration,
			}
		}
	}
	reference := operatorview.BuildExitLineReference(operatorview.ExitLineReferenceSource{
		Market: out.Market, CanonicalSnapshotPresent: canonicalPresent,
		EvidencePresent: evidencePresent, EvidenceLifecycleGeneration: evidenceGen, Raw: raw,
		LifecycleKnown: lifecycleKnown, LifecycleProofRequired: evidencePresent,
		CurrentLifecycleGeneration: state.AdoptionGeneration,
		ManagementStatus:           string(out.AdoptionStatus), EffectiveSettingsKnown: runtime.EffectiveKnown,
		EffectiveStopPct: runtime.Effective.DefaultStopPct,
		Released:         lifecycleKnown && state.Status == positionpolicy.StatusReleased,
		UnknownReason:    unknownReason,
	})
	if reference.GenerationMismatch() || reference.LifecycleUnknown() {
		out.ExitLine = httpapi.ExitLineFrom(operatorview.BuildExitLine(operatorview.Source{
			UnknownReason: "lifecycle_generation_unverified",
		}))
		out.StoredExitEvidence = nil
	}
	if evidencePresent && raw != nil && !canonicalPresent && !reference.LegacyRaw() {
		out.StoredExitEvidence = nil
	}
	out.ExitLineReference = httpapi.ExitLineReferenceFrom(reference)
}
