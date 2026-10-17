package strategyworker

// a112 태스크 2.7 의 빈칸 둘(2.x 대조 감사): 레인 하나의 cadence 창 · 비행 상태가 **이웃 레인의 사이클 열기**를 막지 않는다는 교차 시험이
// 없었다(cadence 는 한 레인에서만, 「Flight」 는 시험 이름에만 있었다). 시계는 가짜라 마감 시한 감시견이 스스로 흐르지 않는다.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func quickStep(context.Context, Input) (Cycle, error) { return Cycle{Outcome: OutcomeDormant}, nil }

// 레인 0 이 방금 사이클을 열어 cadence 창 안(TOO_SOON)이어도 같은 순간 레인 1 은 연다.
func TestOneLaneInsideItsCadenceWindowDoesNotGateItsPeer(t *testing.T) {
	lanes := ProductionLanes(clock.NewFake(laneNow))
	first, peer := lanes[0], lanes[1]
	admit(t, first)
	if _, start := first.RunBounded(context.Background(), Input{}, quickStep); start != StartAdmitted {
		t.Fatalf("arrangement: first cycle start=%s", start)
	}
	admit(t, first)
	if _, start := first.RunBounded(context.Background(), Input{}, quickStep); start != StartTooSoon {
		t.Fatalf("arrangement: the first lane is not inside its cadence window: %s", start)
	}
	admit(t, peer)
	if _, start := peer.RunBounded(context.Background(), Input{}, quickStep); start != StartAdmitted {
		t.Fatalf("a peer was gated by another lane's cadence window: %s", start)
	}
}

// 레인 0 의 사이클이 비행 중(step 이 아직 안 돌아옴)이어도 레인 1 은 사이클을 연다. 레인 0 의 둘째 시도만 IN_FLIGHT 다.
func TestOneLaneInFlightDoesNotMakeItsPeerInFlight(t *testing.T) {
	lanes := ProductionLanes(clock.NewFake(laneNow))
	first, peer := lanes[0], lanes[1]
	started, release := make(chan struct{}), make(chan struct{})
	done := make(chan Start, 1)
	admit(t, first)
	go func() {
		_, start := first.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			close(started)
			<-release
			return Cycle{Outcome: OutcomeDormant}, nil
		})
		done <- start
	}()
	<-started
	admit(t, peer)
	if _, start := peer.RunBounded(context.Background(), Input{}, quickStep); start != StartAdmitted {
		t.Fatalf("a peer reported %s while another lane was in flight", start)
	}
	admit(t, first)
	if _, start := first.RunBounded(context.Background(), Input{}, quickStep); start != StartInFlight {
		t.Fatalf("the in-flight lane opened a second cycle: %s", start)
	}
	close(release)
	if start := <-done; start != StartAdmitted {
		t.Fatalf("the held cycle start=%s", start)
	}
}
