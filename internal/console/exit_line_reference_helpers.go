package console

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"
)

func storedExitReference(state journal.ExitState) *operatorview.StoredExitReference {
	if !hasStoredExitEvidence(state) {
		return nil
	}
	return &operatorview.StoredExitReference{
		EntryPrice: state.EntryPrice, InitialStop: state.InitialStop,
		Baseline: state.Baseline, HighWater: state.HighWater,
		LifecycleGeneration: state.LifecycleGeneration,
	}
}

func storedExitEvidence(state journal.ExitState) storedExitEvidenceView {
	if !hasStoredExitEvidence(state) {
		return storedExitEvidenceView{}
	}
	return storedExitEvidenceView{
		Present: true, EntryPrice: storedExitValue(state.EntryPrice),
		InitialStop: storedExitValue(state.InitialStop), Baseline: storedExitValue(state.Baseline),
		HighWater: storedExitValue(state.HighWater),
	}
}
