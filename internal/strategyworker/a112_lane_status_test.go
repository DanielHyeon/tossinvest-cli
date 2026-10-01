package strategyworker

// a112 태스크 7.5 D2 — Lane.Status 는 레인 한 행을 **한 잠금**으로 읽는다. 값은 개별 접근자가 같은 순간에 읽었을 값과 같아야 한다
// (조용한 상태 · 실패 뒤 · 잠긴 뒤 세 시점에서 대조). 건강 판정은 Health 와 한 함수(healthLocked)를 쓴다.

import (
	"testing"
	"time"
)

func TestTheLaneStatusIsTheRowItsAccessorsRead(t *testing.T) {
	lane, fake := boundedLane(t, threeStrikes())
	check := func(stage string) {
		t.Helper()
		got := lane.Status()
		want := LaneStatus{Health: lane.Health(), ConsecutiveFailures: lane.ConsecutiveFailures(), LatchRevision: lane.LatchRevision(),
			FirstFailure: lane.FirstFailure(), Pending: lane.Pending(), Dropped: lane.Dropped(), Abandoned: lane.Abandoned(),
			NextDue: lane.NextDue(), RestartNotBefore: lane.RestartNotBefore()}
		if got != want {
			t.Fatalf("%s: Status()=%+v, accessors=%+v", stage, got, want)
		}
	}
	check("quiet")
	lane.Offer()
	lane.Offer()
	lane.Fail("first", false)
	check("degraded")
	if lane.Status().Health != LaneDegraded || lane.Status().Dropped != 1 {
		t.Fatalf("degraded row=%+v", lane.Status())
	}
	fake.Advance(time.Minute)
	lane.Fail("second", false)
	lane.Fail("third", false)
	check("latched")
	if row := lane.Status(); row.Health != LaneLatched || row.LatchRevision != 1 || row.FirstFailure != "third" {
		t.Fatalf("latched row=%+v", row)
	}
}
