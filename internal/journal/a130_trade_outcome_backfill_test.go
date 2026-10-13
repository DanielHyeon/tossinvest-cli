package journal

// a130_trade_outcome_backfill_test.go 는 a130 design D4 의 시험임 — 한정 모집단 · 묶음 목록 · 분봉 집합(도구가 브로커에서 읽는 것의 대역) ·
// 미리보기(쓰기 0) → 사람 확인 토큰 → 항목별 재대조 실행.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// a130BackfillFixture 는 비용 모델 없이 묶인 저널(체결 종결마다 cost_model_unconfigured 로 거절)에 다섯을 만듦:
//
//	grown      편입 3 → 대사 3→6 → 엔진 6주 매도(분봉 있음 — 69000)        모집단 · fill
//	unobserved 편입 3 → 대사 3→6 → 엔진 6주 매도(그 분의 분봉 없음)          모집단 · refused(no_intake_candle)
//	adjusted   편입 3 → 대사 3→0(엔진 매도 아님)                         모집단 밖(ADJUSTMENT_CLOSED)
//	entered    엔진 진입 왕복(행 없음)                                   모집단 밖(편입 아님)
//	frozen     편입 3 → 엔진 3주 매도 → 손으로 백필한 행                    모집단 밖(행 있음)
func a130BackfillFixture(t *testing.T) (j *Journal, path string, grown, unobserved string) {
	t.Helper()
	path = t.TempDir() + "/journal.db"
	j = openTestJournalAt(t, path)
	if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	fake := j.clk.(*clock.Fake)

	roundTrip(t, j, "10", "70000", "72000") // entered
	fake.Advance(time.Minute)

	grown = a130GrownTagged(t, j, fake, "g1")
	fake.Advance(time.Minute)
	unobserved = a130GrownTagged(t, j, fake, "g2")
	fake.Advance(time.Minute)

	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Adjust(t, j, "acct-1", AdjustmentExternal, "3", "0", "") // adjusted
	fake.Advance(time.Minute)

	a130Adjust(t, j, "acct-1", AdjustmentExternal, "0", "3", "70100")
	frozen := a130Adopt(t, j, "acct-1", "3", "70000", "66500")
	fake.Advance(time.Minute)
	a130Sell(t, j, "acct-1", frozen, "fz", "3", "72000")
	if _, err := j.BackfillTradeOutcome(ctx, frozen, costs.DefaultModel()); err != nil {
		t.Fatalf("seeding the frozen row: %v", err)
	}
	return j, path, grown, unobserved
}

// a130FileState 는 원장 파일 셋의 상태임 — db · -wal 은 존재 · 바이트 · 수정 시각, -shm 은 존재만. -shm(wal-index)은 WAL 원장을
// 읽는 모든 연결이 읽기 표시를 갱신하는 공유 메모리라, operator-console spec 이 읽기 전용 열기의 문서화된 예외로 둠(readonly.go 머리 주석).
func a130FileState(t *testing.T, path string) [3]string {
	t.Helper()
	var out [3]string
	for i, p := range []string{path, path + "-wal", path + "-shm"} {
		info, err := os.Stat(p)
		if errors.Is(err, os.ErrNotExist) {
			out[i] = "absent"
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if i == 2 {
			out[i] = "present"
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		out[i] = info.ModTime().String() + "|" + hex.EncodeToString(sum[:])
	}
	return out
}

// a130BrokerBars 는 묶음 목록을 읽고 첫 묶음(grown)에만 그 분의 봉을 매김 — 둘째(unobserved)는 그 분에 봉이 없는 것으로 둠.
func a130BrokerBars(t *testing.T, path string) IntakePrices {
	t.Helper()
	ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer ro.Close()
	lots, err := IntakeLotsForBackfill(context.Background(), ro)
	if err != nil {
		t.Fatalf("IntakeLotsForBackfill: %v", err)
	}
	if len(lots) != 2 || !lots[0].CreatedAt.Before(lots[1].CreatedAt) {
		t.Fatalf("lots = %+v, want the grown and the unobserved lots in order", lots)
	}
	return IntakePrices{lots[0].Key(): {Label: lots[0].BarLabel(), Close: "69000", Source: IntakeSourceBrokerCandle1m}}
}

func a130Preview(t *testing.T, path string, now time.Time, prices IntakePrices) TradeOutcomeBackfillPlan {
	t.Helper()
	ro, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer ro.Close()
	plan, err := PreviewTradeOutcomeBackfill(context.Background(), ro, costs.DefaultModel(), now, prices)
	if err != nil {
		t.Fatalf("PreviewTradeOutcomeBackfill: %v", err)
	}
	return plan
}

func a130CountOutcomes(t *testing.T, j *Journal) int {
	t.Helper()
	var n int
	if err := j.db.QueryRow(`SELECT count(*) FROM trade_outcomes`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestA130ThePreviewNamesOnlyTheAdoptedEngineClosesAndWritesNothing: 모집단은 엔진 매도로 닫힌 편입 포지션 중 행 없는 둘뿐이고,
// 미리보기는 원장 파일(db · -wal)의 바이트와 수정 시각을 바꾸지 않고 새 파일을 만들지 않음.
func TestA130ThePreviewNamesOnlyTheAdoptedEngineClosesAndWritesNothing(t *testing.T) {
	j, path, _, _ := a130BackfillFixture(t)
	now := j.clk.Now()
	before := a130FileState(t, path)
	rows := a130CountOutcomes(t, j)

	plan := a130Preview(t, path, now, a130BrokerBars(t, path))

	if after := a130FileState(t, path); after != before {
		t.Errorf("the preview changed the journal files:\nbefore %v\nafter  %v", before, after)
	}
	if got := a130CountOutcomes(t, j); got != rows {
		t.Errorf("trade_outcomes rows %d → %d across a preview", rows, got)
	}
	if len(plan.Items) != 2 {
		t.Fatalf("items = %+v, want the grown and the unobserved adopted closes only", plan.Items)
	}
	fill, refused := plan.Items[0], plan.Items[1]
	if fill.Action != BackfillActionFill || fill.Outcome.InitialQuantity != "6" || len(fill.IntakeCandles) != 1 ||
		fill.IntakeCandles[0].Close != "69000" || fill.IntakeCandles[0].Source != IntakeSourceBrokerCandle1m {
		t.Errorf("fill item = %+v, want quantity 6 priced by one broker 1m bar at 69000", fill)
	}
	if _, wantR := a130ExpectedR(t, "417000", "432000", "3500", "6"); fill.Outcome.RealizedR != wantR {
		t.Errorf("fill R = %s, want %s", fill.Outcome.RealizedR, wantR)
	}
	if refused.Action != BackfillActionRefused || refused.Reason != OutcomeRefusalNoIntakeCandle {
		t.Errorf("refused item = %+v, want no_intake_candle", refused)
	}
	if plan.Token == "" {
		t.Error("the preview must issue a confirmation token")
	}
}

// TestA130TheBackfillWritesOnlyWithTheConfirmedToken: 다른 토큰 → 0 행, 미리보기의 토큰 → 채움 대상만, 같은 토큰 재사용 → 0 행.
func TestA130TheBackfillWritesOnlyWithTheConfirmedToken(t *testing.T) {
	j, path, grown, unobserved := a130BackfillFixture(t)
	ctx := context.Background()
	now := j.clk.Now()
	prices := a130BrokerBars(t, path)
	plan := a130Preview(t, path, now, prices)
	rows := a130CountOutcomes(t, j)

	if w, err := j.BackfillTradeOutcomes(ctx, costs.DefaultModel(), now, prices, "0000000000000000"); !errors.Is(err, ErrBackfillTokenMismatch) || len(w) != 0 {
		t.Fatalf("wrong token: written=%v err=%v", w, err)
	}
	moved := IntakePrices{}
	for k, c := range prices {
		c.Close = "69100"
		moved[k] = c
	}
	if w, err := j.BackfillTradeOutcomes(ctx, costs.DefaultModel(), now, moved, plan.Token); !errors.Is(err, ErrBackfillTokenMismatch) || len(w) != 0 {
		t.Fatalf("a bar that changed since the preview: written=%v err=%v, want the token to refuse", w, err)
	}
	if got := a130CountOutcomes(t, j); got != rows {
		t.Fatalf("a refused execution wrote %d rows", got-rows)
	}
	written, err := j.BackfillTradeOutcomes(ctx, costs.DefaultModel(), now, prices, plan.Token)
	if err != nil || len(written) != 1 || written[0].InstanceSeq != 2 {
		t.Fatalf("confirmed execution: written=%+v err=%v, want the grown position only", written, err)
	}
	if got := outcomeOf(t, j, grown); !sameOutcome(got, plan.Items[0].Outcome) {
		t.Errorf("written = %+v, previewed = %+v — the row must be the previewed one", got, plan.Items[0].Outcome)
	}
	a130NoRow(t, j, unobserved)
	if w, err := j.BackfillTradeOutcomes(ctx, costs.DefaultModel(), now, prices, plan.Token); !errors.Is(err, ErrBackfillTokenMismatch) || len(w) != 0 {
		t.Errorf("replayed token: written=%v err=%v", w, err)
	}
}

// TestA130AMovedItemIsNotWritten: 쓰는 트랜잭션 안에서 다시 계산한 값이 계획 값과 다르면 쓰지 않음(F-7).
func TestA130AMovedItemIsNotWritten(t *testing.T) {
	j, path, grown, _ := a130BackfillFixture(t)
	prices := a130BrokerBars(t, path)
	plan := a130Preview(t, path, j.clk.Now(), prices)
	item := plan.Items[0]
	item.Outcome.RealizedR = "999"
	if _, err := j.backfillPlannedOutcome(context.Background(), item, costs.DefaultModel(), prices); !errors.Is(err, ErrBackfillItemMoved) {
		t.Fatalf("err = %v, want ErrBackfillItemMoved", err)
	}
	a130NoRow(t, j, grown)
}

// TestA130TheBackfillDoesNotRecreateWhatRetentionPruned: 보존 기간(180일) 밖에서 종결된 포지션은 채움 대상이 아님.
func TestA130TheBackfillDoesNotRecreateWhatRetentionPruned(t *testing.T) {
	j, path, _, _ := a130BackfillFixture(t)
	later := j.clk.Now().Add(TradeOutcomeRetention + 24*time.Hour)
	prices := a130BrokerBars(t, path)
	plan := a130Preview(t, path, later, prices)
	for _, it := range plan.Items {
		if it.Action == BackfillActionFill {
			t.Errorf("item %+v is past the retention horizon and must not be filled", it)
		}
	}
	rows := a130CountOutcomes(t, j)
	if w, err := j.BackfillTradeOutcomes(context.Background(), costs.DefaultModel(), later, prices, plan.Token); err != nil || len(w) != 0 {
		t.Errorf("written=%v err=%v, want nothing written", w, err)
	}
	if got := a130CountOutcomes(t, j); got != rows {
		t.Errorf("rows %d → %d", rows, got)
	}
}
