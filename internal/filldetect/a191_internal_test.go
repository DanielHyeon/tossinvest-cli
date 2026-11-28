package filldetect

import "testing"

// TestA191StaleCandidatesDoNotAccumulate (코드 리뷰 1 #1) — settle 전에 실패한 주기의 후보는 다음 주기 등록에서 버려져 쌓이지 않음.
func TestA191StaleCandidatesDoNotAccumulate(t *testing.T) {
	d := &Detector{}
	key := canonicalOrderKey{OrderID: "o-a191-s"}
	for i := 0; i < 5; i++ {
		c := &Cycle{}
		d.noteLeftOpen(c, key, TrackedOrder{OrderID: "o-a191-s"}, 0, Snapshot{})
		if len(d.leftOpenCycle) != 1 || d.leftOpenCycle[0].cycle != c {
			t.Fatalf("after cycle %d: %d candidates, want only this cycle's 1", i, len(d.leftOpenCycle))
		}
	}
}
