package engine

// a092 26라운드 codex P0 · P1 #4 — 배달 실행자에도 델타 「모든 발송자」 규칙:
// 반납이 행 없음 · 모르는 결과로 돌아오면 선점이 아님 → 원칙 E 조건부 차단(승격 없음), 이미 선 한도 판정은 보존.
// 시도 기록이 선점(승인 · 남의 임차)으로 돌아오면 그 사실을 기록.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a092ReleaseLedger 는 반납 결과만 바꿔 치는 원장 — 나머지는 a124 시험 원장 그대로.
type a092ReleaseLedger struct {
	*a124Ledger
	releaseOutcome journal.SettleOutcome
	releases       int
}

func (l *a092ReleaseLedger) ReleaseAlertClaim(ctx context.Context, id int64, token string) (journal.SettleResult, error) {
	l.releases++
	return journal.SettleResult{Outcome: l.releaseOutcome}, nil
}

func a092DelivererReleaseFixture(t *testing.T, outcome journal.SettleOutcome) (*a124Fixture, *a092ReleaseLedger) {
	t.Helper()
	f := a124Setup(t, a124Failing())
	led := &a092ReleaseLedger{a124Ledger: f.led, releaseOutcome: outcome}
	f.d.ledger = led
	return f, led
}

func a092Undelivered(g *execgw.EntryGate) bool {
	_, ok := g.Blocks()[execgw.ReasonAlertUndelivered]
	return ok
}

// 한도 미만 실패 + 반납 행 없음 → 차단(승격 없음).
func TestA092TheDelivererLatchesWhenTheReleaseFindsNoRow(t *testing.T) {
	for name, outcome := range map[string]journal.SettleOutcome{"not-found": journal.SettleNotFound, "unknown": journal.SettleOutcome(99)} {
		t.Run(name, func(t *testing.T) {
			f, led := a092DelivererReleaseFixture(t, outcome)
			f.row(t, "a092-release-"+name)
			_ = f.d.cycle(context.Background())
			if led.releases == 0 {
				t.Fatal("arrangement: no release happened")
			}
			if !a092Undelivered(f.gate) {
				t.Error("a release that found no row was read as a preemption — nothing latched")
			}
			if f.led.escalationCount() != 0 {
				t.Errorf("escalations = %d — a lost row latches but does not escalate (N6)", f.led.escalationCount())
			}
		})
	}
}

// 세대 읽기 뒤의 해제는 존중(원칙 E) — 그 자리의 차단을 다시 세우지 않음.
func TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite(t *testing.T) {
	f, _ := a092DelivererReleaseFixture(t, journal.SettleNotFound)
	f.row(t, "a092-release-epoch")
	a124AtStage(f.d, alertStageEpochRead, func() { f.gate.Clear(execgw.ReasonAlertUndelivered) })
	_ = f.d.cycle(context.Background())
	if a092Undelivered(f.gate) {
		t.Error("the release-site latch was re-applied over an operator release after the epoch read")
	}
}

// 이미 선 한도 판정은 보존 — 한도에 이른 실패의 차단 + 승격은 반납 결과와 무관하게 섬.
func TestA092TheAttemptLimitVerdictSurvivesAMissingRelease(t *testing.T) {
	f, _ := a092DelivererReleaseFixture(t, journal.SettleNotFound)
	id := f.row(t, "a092-release-limit")
	f.exhaust(t, id, alertAttemptLimit-1)
	_ = f.d.cycle(context.Background())
	if !a092Undelivered(f.gate) {
		t.Error("no latch at the attempt limit")
	}
	if f.led.escalationCount() != 1 {
		t.Errorf("escalations = %d, want the limit verdict's one", f.led.escalationCount())
	}
}

// P1 #4: 시도 기록이 선점으로 돌아오면 그 사실을 기록(행 id · 결과), 잠그지도 승격하지도 않음.
func TestA092TheDelivererRecordsAPreemptedAttempt(t *testing.T) {
	for name, outcome := range map[string]journal.SettleOutcome{"acknowledged": journal.SettleAlreadySettled, "lease-lost": journal.SettleLeaseLost} {
		t.Run(name, func(t *testing.T) {
			f := a124Setup(t, a124Failing())
			o := outcome
			f.led.recordOutcome = &o
			f.row(t, "a092-preempted-"+name)
			_ = f.d.cycle(context.Background())
			logs := f.logs.String()
			if !strings.Contains(logs, "preempted") || !strings.Contains(logs, outcome.String()) {
				t.Errorf("the preemption was not recorded:\n%s", logs)
			}
			if a092Undelivered(f.gate) || f.led.escalationCount() != 0 {
				t.Error("a preemption latched or escalated")
			}
		})
	}
}

// 26라운드 codex 재확인 R1: 시도 기록은 Applied(한도 미만)인데 반납에서야 선점(승인 · 남의 임차)이 드러나도 그 사실을 기록함
// (델타 「선점이 일어났다는 사실은 기록되어야 한다」 — 모든 발송자). 잠그지도 승격하지도 않음.
func TestA092TheDelivererRecordsAPreemptionSeenOnlyAtRelease(t *testing.T) {
	for name, outcome := range map[string]journal.SettleOutcome{"acknowledged": journal.SettleAlreadySettled, "lease-lost": journal.SettleLeaseLost} {
		t.Run(name, func(t *testing.T) {
			f, led := a092DelivererReleaseFixture(t, outcome)
			f.row(t, "a092-release-preempted-"+name)
			_ = f.d.cycle(context.Background())
			if led.releases == 0 {
				t.Fatal("arrangement: no release happened")
			}
			logs := f.logs.String()
			if !strings.Contains(logs, "before it could be handed back") || !strings.Contains(logs, outcome.String()) {
				t.Errorf("a preemption seen at release was not recorded:\n%s", logs)
			}
			if a092Undelivered(f.gate) || f.led.escalationCount() != 0 {
				t.Error("a preemption seen at release latched or escalated")
			}
		})
	}
}
