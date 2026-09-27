package engine_test

// a124 tasks 2.1 — 생산 생성자(`Context.AlertDeliverer`)로 세운 실행자가, 동기 알림 경로가 한 번도
// 시도하지 않는 구성에서 재시도 한도에 이르면 진입을 막고 운영 모드를 승격한다.
//
// 이 파일은 HEAD 의 공개 표면만 쓴다 — 그래서 HEAD 에서 컴파일되고 **행동으로** 빨갛다(a098 의 실행자는
// 게이트도 모드도 부르지 않는다). 내부 판정 순서는 `a124_the_deliverer_judges_internal_test.go` 가 잰다.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type a124DeadTransport struct{}

func (a124DeadTransport) Publish(context.Context, obs.Notification) error {
	return errors.New("transport is down")
}

func TestTheProductionExecutorLatchesWithoutTheSynchronousPath(t *testing.T) {
	for _, tc := range []struct {
		name string
		pub  obs.Publisher
	}{{"transport fails", a124DeadTransport{}}, {"no publisher", nil}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clk := clock.NewFake(time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))
			j, err := journal.Open(ctx, journal.Options{
				Path:     filepath.Join(t.TempDir(), journal.DBFileName),
				Clock:    clk,
				FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
			})
			if err != nil {
				t.Fatalf("journal.Open: %v", err)
			}
			t.Cleanup(func() { _ = j.Close() })
			gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
			// Notifier 는 전송 수단의 출처일 뿐 — 이 시험에서 Notify 는 한 번도 불리지 않는다.
			n := &obs.Notifier{Publisher: tc.pub, Journal: j, Gate: gate, AccountRef: exitAccount, Clock: clk}
			if _, err := j.EnqueueAlert(ctx, journal.Alert{
				EventKey: "a124-production", Type: "execgw.order_unresolved", Severity: "critical", Title: "t",
			}); err != nil {
				t.Fatalf("EnqueueAlert: %v", err)
			}
			aux, err := (&engine.Context{Journal: j, Entry: gate, Notifier: n, AccountRef: exitAccount}).AlertDeliverer(clk)
			if err != nil {
				t.Fatalf("AlertDeliverer: %v", err)
			}
			runCtx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})
			go func() { defer close(done); _ = aux.Run(runCtx) }()
			t.Cleanup(func() { cancel(); <-done })

			const limit = obs.DefaultCriticalAttempts
			for cycle := 0; cycle < limit; cycle++ {
				if !clk.WaitForSleepers(1, 10*time.Second) {
					t.Fatalf("cycle %d: the executor never slept", cycle)
				}
				if cycle < limit-1 {
					clk.Advance(2 * time.Second)
				}
			}
			if _, latched := gate.Blocks()[execgw.ReasonAlertUndelivered]; !latched {
				t.Fatalf("after %d failed cycles the gate is open — the executor does not own sustained failure", limit)
			}
			snap, err := j.CurrentOperatingMode(ctx, exitAccount)
			if err != nil {
				t.Fatalf("CurrentOperatingMode: %v", err)
			}
			if snap.Mode != journal.ModeEntryBlocked {
				t.Fatalf("mode = %s, want %s", snap.Mode, journal.ModeEntryBlocked)
			}
		})
	}
}
