package journal

import (
	"context"
	"errors"
	"testing"
	"time"
)

// 이 파일은 a112 태스크 5.3.3 의 durable entry latch 계약을 값으로 잰다.
//
// 네 문장이 계약이다.
//  1. 잠금은 프로세스보다 오래 산다.
//  2. 열려 있는 잠금은 레인마다 하나이고, 그것은 **첫** 원인이다.
//  3. 복구는 더 큰 **서명 활성화 세대**가 있어야 한다. 성공한 사이클로는 못 연다.
//  4. 잠금도 복구도 지우거나 고칠 수 없다.

func laneLatchFixture(t *testing.T) StrategyLaneLatch {
	t.Helper()
	return StrategyLaneLatch{
		AccountRef: "acct-lane", Market: "KR", Family: "BREAKOUT_RETEST",
		LaneID: "kr_short_breakout_retest_v1", LaneVersion: "v1",
		LatchID:       "lane-latch:KR:BREAKOUT_RETEST:kr_short_breakout_retest_v1:v1:1",
		LatchRevision: 1, Reason: "breakout evidence store unavailable", Abnormal: true,
		ActivationGeneration: 7, ObservedAt: time.Date(2026, 9, 3, 5, 0, 0, 0, time.UTC),
	}
}

func TestAStrategyLaneLatchOutlivesTheProcessThatWroteIt(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/journal.db"
	first := openJournalAtSchema(t, path, SchemaVersion)
	stored, err := first.RecordStrategyLaneLatch(ctx, laneLatchFixture(t))
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if stored.Seq <= 0 {
		t.Fatalf("원장이 잠금에 신원을 주지 않았다: %+v", stored)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	// 프로세스가 다시 선다.
	second := openJournalAtSchema(t, path, SchemaVersion)
	defer second.Close()
	open, err := second.OpenStrategyLaneLatches(ctx, "acct-lane")
	if err != nil {
		t.Fatalf("open latches: %v", err)
	}
	if len(open) != 1 {
		t.Fatalf("재시작 뒤 열린 잠금 %d 개 — 잠금이 프로세스와 함께 사라졌다", len(open))
	}
	if open[0].Reason != "breakout evidence store unavailable" || !open[0].Abnormal ||
		open[0].ActivationGeneration != 7 || !open[0].ObservedAt.Equal(laneLatchFixture(t).ObservedAt) {
		t.Fatalf("잠금이 다른 값으로 돌아왔다: %+v", open[0])
	}
}

func TestTheOpenLatchIsAlwaysTheFirstCauseNotTheLast(t *testing.T) {
	ctx := context.Background()
	j := openJournalAtSchema(t, t.TempDir()+"/journal.db", SchemaVersion)
	defer j.Close()
	first, err := j.RecordStrategyLaneLatch(ctx, laneLatchFixture(t))
	if err != nil {
		t.Fatalf("first: %v", err)
	}

	later := laneLatchFixture(t)
	later.Reason = "a later and less interesting failure"
	later.LatchRevision = 2
	later.ObservedAt = later.ObservedAt.Add(time.Hour)
	second, err := j.RecordStrategyLaneLatch(ctx, later)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second.Seq != first.Seq || second.Reason != first.Reason {
		t.Fatalf("두 번째 실패가 첫 원인을 덮었다: %+v", second)
	}
	open, err := j.OpenStrategyLaneLatches(ctx, "acct-lane")
	if err != nil || len(open) != 1 {
		t.Fatalf("열린 잠금 %d 개 err=%v — 레인 하나에 열린 잠금은 하나여야 한다", len(open), err)
	}
}

func TestALatchOpensOnlyForAStrictlyNewerSignedActivation(t *testing.T) {
	ctx := context.Background()
	j := openJournalAtSchema(t, t.TempDir()+"/journal.db", SchemaVersion)
	defer j.Close()
	stored, err := j.RecordStrategyLaneLatch(ctx, laneLatchFixture(t))
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	for _, generation := range []uint64{1, 6, 7} {
		if err := j.RecoverStrategyLaneLatch(ctx, stored.Seq, generation); !errors.Is(err, ErrStrategyLaneLatchRecoveryEvidence) {
			t.Fatalf("세대 %d 로 잠금이 열렸다(err=%v) — 잠긴 그 상태가 증거가 되면 안 된다", generation, err)
		}
	}
	open, err := j.OpenStrategyLaneLatches(ctx, "acct-lane")
	if err != nil || len(open) != 1 {
		t.Fatalf("거절된 복구 뒤 열린 잠금 %d 개 err=%v", len(open), err)
	}

	if err := j.RecoverStrategyLaneLatch(ctx, stored.Seq, 8); err != nil {
		t.Fatalf("더 큰 세대인데 복구가 거절됐다: %v", err)
	}
	open, err = j.OpenStrategyLaneLatches(ctx, "acct-lane")
	if err != nil || len(open) != 0 {
		t.Fatalf("복구 뒤에도 열린 잠금 %d 개 err=%v", len(open), err)
	}

	// 복구된 레인이 다시 잠길 수 있어야 한다. 못 잠기면 한 번 복구한 레인이
	// 영원히 잠기지 않는 레인이 된다.
	again := laneLatchFixture(t)
	again.Reason = "the same fault came back"
	again.ActivationGeneration = 8
	again.ObservedAt = again.ObservedAt.Add(2 * time.Hour)
	reopened, err := j.RecordStrategyLaneLatch(ctx, again)
	if err != nil {
		t.Fatalf("복구된 레인을 다시 잠그지 못했다: %v", err)
	}
	if reopened.Seq == stored.Seq || reopened.Reason != "the same fault came back" {
		t.Fatalf("새 잠금이 옛 기록을 되쓴다: %+v", reopened)
	}
}

func TestAStrategyLaneLatchAndItsRecoveryAreHistoryNotState(t *testing.T) {
	ctx := context.Background()
	j := openJournalAtSchema(t, t.TempDir()+"/journal.db", SchemaVersion)
	defer j.Close()
	stored, err := j.RecordStrategyLaneLatch(ctx, laneLatchFixture(t))
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if err := j.RecoverStrategyLaneLatch(ctx, stored.Seq, 9); err != nil {
		t.Fatalf("recover: %v", err)
	}
	for _, statement := range []string{
		`UPDATE strategy_lane_latches SET reason='rewritten'`,
		`DELETE FROM strategy_lane_latches`,
		`UPDATE strategy_lane_latch_recoveries SET activation_generation=99`,
		`DELETE FROM strategy_lane_latch_recoveries`,
	} {
		if _, err := j.db.ExecContext(ctx, statement); err == nil {
			t.Fatalf("%q 가 통과했다 — 잠금과 복구는 기록이지 상태가 아니다", statement)
		}
	}
}

func TestTheLatchTableRefusesARecordItCannotClassify(t *testing.T) {
	ctx := context.Background()
	j := openJournalAtSchema(t, t.TempDir()+"/journal.db", SchemaVersion)
	defer j.Close()
	for name, mutate := range map[string]func(*StrategyLaneLatch){
		"계좌 없음":      func(l *StrategyLaneLatch) { l.AccountRef = "" },
		"모르는 시장":     func(l *StrategyLaneLatch) { l.Market = "JP" },
		"모르는 가족":     func(l *StrategyLaneLatch) { l.Family = "MEAN_REVERSION" },
		"이유 없음":      func(l *StrategyLaneLatch) { l.Reason = "" },
		"revision 0": func(l *StrategyLaneLatch) { l.LatchRevision = 0 },
		"관측 시각 없음":   func(l *StrategyLaneLatch) { l.ObservedAt = time.Time{} },
	} {
		latch := laneLatchFixture(t)
		mutate(&latch)
		if _, err := j.RecordStrategyLaneLatch(ctx, latch); !errors.Is(err, ErrStrategyLaneLatchInvalid) {
			t.Fatalf("%s: err=%v — 분류할 수 없는 기록이 들어갔다", name, err)
		}
	}
}
