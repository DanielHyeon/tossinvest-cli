package operatorview

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// ExitLiveness is the transport-neutral engine state relevant to a persisted
// exit line. Only a positively known stop overrides the evidence clock.
type ExitLiveness string

const (
	ExitLivenessRunning     ExitLiveness = "running"
	ExitLivenessStopped     ExitLiveness = "stopped"
	ExitLivenessUnavailable ExitLiveness = "unavailable"
	ExitLivenessUnwired     ExitLiveness = "unwired"
)

// ApplyExitFreshness is the single operator-facing freshness verdict. A111's
// durable observation heartbeat makes the persisted 30-second bound truthful
// for every state except a positively observed stopped engine.
func ApplyExitFreshness(view journal.ExitSnapshotView, asOf time.Time, liveness ExitLiveness) journal.ExitSnapshotView {
	if liveness == ExitLivenessStopped {
		view.Stale, view.StaleReason = true, "engine_not_running"
		return view
	}
	return view.WithFreshness(asOf, 30*time.Second)
}
