package engine

// a124 tasks 4.4 · design D6 — 래치까지의 시간을 **잰다**(가짜 시계). 전제 H(동질 두절: 모든 발행이 같은 T 로
// 실패, 새 행 없음, 임차 경합 없음, 해제 없음, 정산 Applied) 아래 D6 의 조건부 식
//
//	래치 ≈ (L−1)·C + (T + S + M)   (첫 사이클 시작 기준, C = B·(T+S+M+E) + I_list + I)
//
// 이 가짜 시계에서는 원장 · 게이트 비용(S · M · E · I_list)이 0 이므로 식은 (L−1)·(B·T + I) + T 가 된다.
// 발행 실패 시간 T 는 transport 가 소유하므로, 발행기가 가짜 시계를 T 만큼 전진시킨 뒤 실패를 돌려준다.
// 큐 대기 Q(두절 뒤 첫 사이클 시작까지)는 이 측정에 넣지 않는다 — 첫 사이클 시작에서 잰다(D6 표의 첫 열).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type a124SlowFailingPublisher struct {
	clk *clock.Fake
	t   time.Duration
}

func (p *a124SlowFailingPublisher) Publish(context.Context, obs.Notification) error {
	p.clk.Advance(p.t)
	return errors.New("timed out")
}

func TestTheLatchTimeMatchesDesignD6UnderPremiseH(t *testing.T) {
	for _, tc := range []struct {
		name string
		rows int
		t    time.Duration
		want time.Duration
	}{
		{"one row, immediate failure", 1, 0, 4 * time.Second},
		{"one row, 10 s timeout", 1, 10 * time.Second, 34 * time.Second},
		{"a batch of 10, all 10 s timeout", 10, 10 * time.Second, 214 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := a124Setup(t, nil)
			f.d.Publisher = &a124SlowFailingPublisher{clk: f.clk, t: tc.t}
			for i := 0; i < tc.rows; i++ {
				f.row(t, "a124-4.4-"+tc.name+"-"+string(rune('a'+i)))
			}
			start := f.clk.Now()
			var latchedAt time.Time
			f.d.judgeHook = func(stage alertJudgeStage, _ int64) {
				if stage == alertStageLatched && latchedAt.IsZero() && f.latched() {
					latchedAt = f.clk.Now()
				}
			}
			for cycle := 0; cycle < alertAttemptLimit+1 && latchedAt.IsZero(); cycle++ {
				if err := f.d.cycle(context.Background()); err != nil {
					t.Fatalf("cycle: %v", err)
				}
				if latchedAt.IsZero() {
					f.clk.Advance(f.d.interval()) // Run 의 사이클 사이 대기
				}
			}
			if latchedAt.IsZero() {
				t.Fatal("never latched")
			}
			got := latchedAt.Sub(start)
			t.Logf("%s: latched %v after the first cycle started (design D6: %v)", tc.name, got, tc.want)
			if got != tc.want {
				t.Fatalf("latched after %v, design D6 says %v under premise H", got, tc.want)
			}
		})
	}
}

// 큐 대기 Q 를 넣은 판(tasks 4.4 「큐 대기 포함」, codex 3회차 T4). 두절이 **사이클 사이의 대기가 막 시작된 순간**에
// 시작하면 Q = I(2 s) 다 — 두절 뒤 첫 사이클은 그 대기가 끝나야 돈다. 두절 순간을 기준으로 잰다.
// Q 의 상한 C(진행 중인 배치 사이클의 나머지 + 대기)는 이 시험이 흉내 내지 않는다 — D6 는 Q ≤ C 를 식의 항으로만 둔다.
func TestTheLatchTimeIncludesTheQueueWait(t *testing.T) {
	for _, tc := range []struct {
		name string
		rows int
		t    time.Duration
		want time.Duration
	}{
		{"one row, 10 s timeout, outage at the start of the wait", 1, 10 * time.Second, 2*time.Second + 34*time.Second},
		{"a batch of 10, 10 s timeout, outage at the start of the wait", 10, 10 * time.Second, 2*time.Second + 214*time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := a124Setup(t, &a098RecordingPublisher{}) // 두절 전: 전송이 된다
			// 두절 전 한 사이클 — 선행 행을 보내고 대기에 든다. 그 대기가 시작되는 순간이 두절 순간이다.
			f.row(t, "a124-4.4q-before")
			if err := f.d.cycle(context.Background()); err != nil {
				t.Fatalf("cycle: %v", err)
			}
			outage := f.clk.Now()
			f.d.Publisher = &a124SlowFailingPublisher{clk: f.clk, t: tc.t}
			for i := 0; i < tc.rows; i++ {
				f.row(t, "a124-4.4q-"+tc.name+"-"+string(rune('a'+i)))
			}
			var latchedAt time.Time
			f.d.judgeHook = func(stage alertJudgeStage, _ int64) {
				if stage == alertStageLatched && latchedAt.IsZero() && f.latched() {
					latchedAt = f.clk.Now()
				}
			}
			f.clk.Advance(f.d.interval()) // Q = I: 두절 순간에 시작된 대기가 끝난다
			for cycle := 0; cycle < alertAttemptLimit+1 && latchedAt.IsZero(); cycle++ {
				if err := f.d.cycle(context.Background()); err != nil {
					t.Fatalf("cycle: %v", err)
				}
				if latchedAt.IsZero() {
					f.clk.Advance(f.d.interval())
				}
			}
			if latchedAt.IsZero() {
				t.Fatal("never latched")
			}
			got := latchedAt.Sub(outage)
			t.Logf("%s: latched %v after the outage began (Q = %v)", tc.name, got, f.d.interval())
			if got != tc.want {
				t.Fatalf("latched after %v, want %v (Q + D6)", got, tc.want)
			}
		})
	}
}
