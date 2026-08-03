package journal

import "time"

// WithIntegrity applies only the freshness checks that do not depend on an age
// bound (change a061).
//
// WithFreshness answers "is this reading too old", which presumes the stamp is a
// reading. exit_states.last_observed_at is not: it is written from record() and
// the observation loop returns before record() whenever the judgement moved
// nothing, so an untouched stamp means the line did not change — not that nobody
// looked. A caller that has separate, positive proof the engine is maintaining
// the line therefore has no use for the age, and the operator console asks this
// instead.
//
// What no proof of liveness excuses is corrupt evidence. A stamp that cannot be
// parsed and a stamp from the future are wrong under every staleness policy, so
// they stay here.
func (v ExitSnapshotView) WithIntegrity(asOf time.Time) ExitSnapshotView {
	if v.Snapshot == nil || asOf.IsZero() {
		return v
	}
	observed, err := time.Parse(time.RFC3339Nano, v.Snapshot.ObservedAt)
	switch {
	case err != nil:
		v.Stale, v.StaleReason = true, "invalid_observed_at"
	case observed.After(asOf):
		v.Stale, v.StaleReason = true, "observation_in_future"
	}
	return v
}
