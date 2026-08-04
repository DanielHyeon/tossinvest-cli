package operatorview

import "strings"

// staleStatusText is the short verdict beside a row whose actionable values are
// hidden (change a077).
//
// Every stale reason hides the same five values, but they are not the same
// problem and the operator's next move differs: an old evaluation is waited on,
// a stopped engine is started, and a quarantined position needs a person to look
// at why the engine refused it. Before a077 all three read 오래된 평가, which is
// actively wrong for the last two — a quarantined line is not old, it is
// unmaintained.
//
// The status code stays "stale" in all three cases so no transport has to learn
// a new state; only the words change.
func staleStatusText(reason string) string {
	switch strings.TrimSpace(reason) {
	case "snapshot_quarantined":
		return "판정 격리"
	case "engine_not_running":
		return "엔진 정지"
	default:
		return "오래된 평가"
	}
}
