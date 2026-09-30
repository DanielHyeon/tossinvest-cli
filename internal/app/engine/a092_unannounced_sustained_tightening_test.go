package engine_test

// a092 26라운드 보이스 B #7 잔재: 지속 실패 승격이 커밋되고 통지 기록만 실패하면(ErrModeAnnouncementFailed) 「원장에 닿지 않아
// 재시작이 차단을 푼다」가 아님 — 단위 ⑤ 의 exit · 자격 증명 두 자리와 같은 규칙.

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA092AnUnannouncedSustainedTighteningIsReportedAsTightened(t *testing.T) {
	buf := &bytes.Buffer{}
	escalator := &recordingEscalator{err: fmt.Errorf("journal: %w", journal.ErrModeAnnouncementFailed)}
	health := &countingHealth{}
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: runtimeAccount,
		Alerts:     &recordingAlerts{},
		Escalate:   escalator,
		Log:        obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true}),
		Loops: []engine.SupervisedLoop{{
			Name:    "reconcile",
			Run:     func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
			Health:  health,
			Trigger: journal.ModeTriggerReconcileCycleFailure,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	health.set(engine.DefaultDegradationThreshold)
	rt.CheckHealth(context.Background())
	if len(escalator.seen()) != 1 {
		t.Fatalf("arrangement: escalations = %v", escalator.seen())
	}
	logs := buf.String()
	if strings.Contains(logs, "did not reach the operating mode") {
		t.Errorf("a committed tightening whose notice failed was reported as not persisted:\n%s", logs)
	}
	if !strings.Contains(logs, "notice could not be recorded") {
		t.Errorf("the notice failure was not named:\n%s", logs)
	}
}
