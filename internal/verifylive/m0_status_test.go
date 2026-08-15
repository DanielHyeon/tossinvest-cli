package verifylive

import (
	"bytes"
	"strings"
	"testing"
)

func TestM0CheckpointsAreVisibleInStatusButNeverAbortTargets(t *testing.T) {
	entries := []Entry{{Kind: KindM0Checkpoint, M0Checkpoint: &M0Checkpoint{Kind: "pending-create", ClientOrderID: "client-pending", Symbol: "005930"}},
		{Kind: KindM0Checkpoint, M0Checkpoint: &M0Checkpoint{Kind: "child-observed", ChildOrderID: "child-1", Symbol: "005930"}}}
	progress := BuildProgress("record.jsonl", entries)
	if len(progress.M0Checkpoints) != 2 {
		t.Fatalf("status checkpoints = %+v, want pending and child", progress.M0Checkpoints)
	}
	if got := Outstanding(entries); len(got) != 0 {
		t.Fatalf("checkpoint became cleanup outstanding: %+v", got)
	}
	if got := AbortTargets(entries); len(got) != 0 {
		t.Fatalf("checkpoint became abort target: %+v", got)
	}
	var text bytes.Buffer
	progress.WriteText(&text)
	if !strings.Contains(text.String(), "child-observed") || !strings.Contains(text.String(), "child-1") {
		t.Fatalf("status text hid checkpoint: %s", text.String())
	}
}
