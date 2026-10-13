package journal

// a130_journal_copy_census_test.go 는 a130 design D4 의 측정임 — 운영 저널의 **읽기 전용 사본**에서 ① 이미 동결된 행이 a130 판정으로
// 같은 숫자를 다시 내는지(회귀), ② 행 없는 CLOSED 의 사유 분포, ③ backfill 모집단(엔진 매도로 닫힌 편입 포지션 중 행 없는 것)의 판정
// 분포를 셈.
//
// 기본은 건너뜀. 사람이 사본 경로를 줄 때만 돎:
//
//	A130_JOURNAL_COPY=/path/to/copy/journal.db go test -count=1 -run TestA130JournalCopyCensus -v ./internal/journal/
//
// 운영 저널 자체에는 돌리지 않음(OpenReadOnly 라 쓰지는 않지만 측정은 사본으로 함). 출력은 수 · 사유만 — 종목 · 수량 · 손익 · 계좌 없음(F-6).

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

type a130CensusReader struct{ ro *ReadOnly }

func (r a130CensusReader) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.ro.db.QueryContext(ctx, query, args...)
}

func TestA130JournalCopyCensus(t *testing.T) {
	path := os.Getenv("A130_JOURNAL_COPY")
	if path == "" {
		t.Skip("A130_JOURNAL_COPY not set — the census runs only against a read-only copy a person names")
	}
	ctx := context.Background()
	ro, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer ro.Close()
	r := a130CensusReader{ro: ro}

	rows, err := ro.db.QueryContext(ctx, `
		SELECT p.id, o.position_id IS NOT NULL
		  FROM positions p LEFT JOIN trade_outcomes o ON o.position_id = p.id
		 WHERE p.state = 'CLOSED'`)
	if err != nil {
		t.Fatal(err)
	}
	type closed struct {
		id     string
		frozen bool
	}
	var all []closed
	for rows.Next() {
		var c closed
		if err := rows.Scan(&c.id, &c.frozen); err != nil {
			t.Fatal(err)
		}
		all = append(all, c)
	}
	_ = rows.Close()

	missing := map[TradeOutcomeRefusalReason]int{}
	frozenRefused := map[TradeOutcomeRefusalReason]int{}
	var frozen, frozenEqual, recoverable int
	var durations []time.Duration
	for _, c := range all {
		start := time.Now()
		computed, reason := computeTradeOutcome(ctx, r, c.id, costs.DefaultModel(), nil)
		durations = append(durations, time.Since(start))
		if !c.frozen {
			if reason == "" {
				recoverable++
			}
			missing[reason]++
			continue
		}
		frozen++
		if reason != "" {
			frozenRefused[reason]++
			continue
		}
		stored, err := ro.db.QueryContext(ctx, `
			SELECT position_id, realized_pnl_after_costs, realized_r, initial_risk, initial_quantity,
			       coalesce(held_seconds, 0), coalesce(exit_ratchet_level, ''), exit_rung, closed_at, cost_total
			  FROM trade_outcomes WHERE position_id = ?`, c.id)
		if err != nil {
			t.Fatal(err)
		}
		if stored.Next() {
			o, err := scanTradeOutcome(stored)
			if err == nil && o.RealizedPnLAfterCosts == computed.RealizedPnLAfterCosts && o.RealizedR == computed.RealizedR &&
				o.InitialRisk == computed.InitialRisk && o.InitialQuantity == computed.InitialQuantity {
				frozenEqual++
			}
		}
		_ = stored.Close()
	}
	t.Logf("schema=%d closed=%d frozen=%d missing=%d", ro.SchemaVersion(), len(all), frozen, len(all)-frozen)
	// 체결 트랜잭션 안에서 도는 것과 같은 판정(prices = nil)의 소요 — 리뷰 4차 P2-4(쓰기 잠금 아래 추가 시간).
	sort.Slice(durations, func(i, k int) bool { return durations[i] < durations[k] })
	if len(durations) > 0 {
		t.Logf("judgement duration (prices=nil): median %s · max %s over %d closed positions",
			durations[len(durations)/2].Round(time.Microsecond), durations[len(durations)-1].Round(time.Microsecond), len(durations))
	}
	t.Logf("missing rows that the a130 judgement would freeze (backfill targets): %d", recoverable)
	for _, line := range a130Histogram(missing) {
		t.Logf("missing · reason %s", line)
	}
	t.Logf("frozen rows recomputed to the same numbers: %d/%d", frozenEqual, frozen)
	for _, line := range a130Histogram(frozenRefused) {
		t.Logf("frozen but the a130 judgement now refuses · reason %s", line)
	}

	plan, err := PreviewTradeOutcomeBackfill(ctx, ro, costs.DefaultModel(), time.Now(), nil)
	if err != nil {
		t.Fatalf("PreviewTradeOutcomeBackfill: %v", err)
	}
	actions := map[TradeOutcomeRefusalReason]int{}
	for _, it := range plan.Items {
		actions[TradeOutcomeRefusalReason(it.Action+":"+string(it.Reason))]++
	}
	t.Logf("backfill population (engine-closed adopted, no row): %d", len(plan.Items))
	for _, line := range a130Histogram(actions) {
		t.Logf("backfill · %s", line)
	}
}

func a130Histogram(m map[TradeOutcomeRefusalReason]int) []string {
	var out []string
	for k, v := range m {
		name := string(k)
		if name == "" {
			name = "(none — computes)"
		}
		out = append(out, name+" = "+itoa(v))
	}
	sort.Strings(out)
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
