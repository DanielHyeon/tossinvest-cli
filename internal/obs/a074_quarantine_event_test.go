package obs_test

// a074 task 3.1: the grading of the quarantine-creation event.
//
// The rule this file pins is the one event.go states for the whole exit group —
// "does the condition mean a position is *not being protected*?" A quarantined
// generation is dropped from the judgement set entirely, stop evaluation
// included, so the answer is yes and the grade is critical.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestAQuarantineCreationIsCritical(t *testing.T) {
	if got := obs.SeverityOf(obs.EventExitSnapshotQuarantined); got != obs.SeverityCritical {
		t.Fatalf("severity of %s = %s, want critical", obs.EventExitSnapshotQuarantined, got)
	}
}

func TestTheQuarantineEventIsNamedUnderTheExitSubject(t *testing.T) {
	// The strings are a contract: the log's `event` field, the ntfy tag and the
	// outbox's event_type column. A dashboard keyed on `exit.` has to see it.
	if got := obs.EventExitSnapshotQuarantined.Subject(); got != "exit" {
		t.Fatalf("subject = %q, want exit", got)
	}
	if string(obs.EventExitSnapshotQuarantined) != "exit.snapshot_quarantined" {
		t.Fatalf("event name = %q", obs.EventExitSnapshotQuarantined)
	}
}

func TestTheQuarantineEventIsListedAmongTheCriticalOnes(t *testing.T) {
	for _, e := range obs.CriticalEvents() {
		if e == obs.EventExitSnapshotQuarantined {
			return
		}
	}
	t.Fatal("the quarantine creation event is graded critical but is not in CriticalEvents()")
}
