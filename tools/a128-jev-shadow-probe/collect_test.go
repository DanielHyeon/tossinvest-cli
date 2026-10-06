package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestLedgerRowSchema 는 judgment·label 행의 열쇠 집합을 고정함(spec 「원장 행은 판단의 모집단을
// 복원 가능하게 적는다」 — 시각·시장·심볼·질문 원문 버전·state digest·p·기준가).
func TestLedgerRowSchema(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	rig.clock.Set(at(11, 0))
	if err := rig.probe.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(rig.dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	wantJudgment := []string{"at", "base_price", "collection_gaps", "currency", "horizon_minutes", "id", "input_tokens",
		"kind", "label_due_at", "market", "model_answered", "model_requested", "p", "question_digest", "question_id",
		"question_rev", "schema", "state_digest", "symbol"}
	wantLabel := []string{"at", "due_at", "judgment_id", "kind", "realized_price", "realized_return", "schema"}
	kinds := map[string]int{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("not JSON: %s", line)
		}
		kind, _ := row["kind"].(string)
		kinds[kind]++
		if row["schema"] != ledgerSchema {
			t.Fatalf("schema = %v", row["schema"])
		}
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		switch kind {
		case kindJudgment:
			if !reflect.DeepEqual(keys, wantJudgment) {
				t.Fatalf("judgment keys = %v", keys)
			}
			if !strings.HasPrefix(row["state_digest"].(string), "sha256:") || !strings.HasPrefix(row["question_digest"].(string), "sha256:") {
				t.Fatalf("digests missing: %s", line)
			}
			if row["question_rev"] != questionRev || row["model_answered"] != "jev-1.13.0" {
				t.Fatalf("population fields: %s", line)
			}
			if _, err := time.Parse(time.RFC3339, row["at"].(string)); err != nil {
				t.Fatalf("at is not RFC3339: %s", line)
			}
		case kindLabel:
			if !reflect.DeepEqual(keys, wantLabel) {
				t.Fatalf("label keys = %v", keys)
			}
		default:
			t.Fatalf("unexpected row kind %q", kind)
		}
	}
	if kinds[kindJudgment] != 4 || kinds[kindLabel] != 4 {
		t.Fatalf("rows = %v, want 4 judgments + 4 labels", kinds)
	}
	// state 원문은 digest 이름으로 보관돼 재검증 가능해야 함.
	contents := rig.contents(t)
	stateFile := filepath.Join(rig.dir, statesDirName, strings.TrimPrefix(contents.Judgments[0].StateDigest, "sha256:")+".json")
	if _, err := os.Stat(stateFile); err != nil {
		t.Fatalf("state file for the digest is missing: %v", err)
	}
}

func TestLabelComputesRealizedReturn(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	rig.clock.Set(at(10, 59))
	if err := rig.probe.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(rig.contents(t).Labels) != 0 {
		t.Fatal("labels written before they are due")
	}
	rig.official.mu.Lock()
	rig.official.prices["005930"] = "71400" // +2%
	rig.official.mu.Unlock()
	rig.clock.Set(at(11, 0))
	if err := rig.probe.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	for _, row := range contents.Judgments {
		label := contents.Labels[row.ID]
		if label.RealizedReturn == nil {
			t.Fatalf("no realized return for %s: %+v", row.ID, label)
		}
		if row.Symbol == "005930" && *label.RealizedReturn != 0.02 {
			t.Fatalf("005930 return = %v, want 0.02", *label.RealizedReturn)
		}
		if row.Symbol == "000660" && *label.RealizedReturn != 0 {
			t.Fatalf("000660 return = %v, want 0", *label.RealizedReturn)
		}
	}
	if rig.probe.pendingCount() != 0 {
		t.Fatal("pending labels left after labeling")
	}
}

// TestLabelGapReasons 는 실현값을 못 얻는 세 경로가 각자 사유를 남기는지 봄.
func TestLabelGapReasons(t *testing.T) {
	t.Parallel()
	t.Run("price missing", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
			t.Fatal(err)
		}
		rig.official.mu.Lock()
		delete(rig.official.prices, "000660")
		rig.official.mu.Unlock()
		rig.clock.Set(at(11, 0))
		if err := rig.probe.labelTick(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertLabelGap(t, rig.contents(t), "000660", "realized_price_missing")
	})
	t.Run("rate limit gave up", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
			t.Fatal(err)
		}
		rig.official.mu.Lock()
		rig.official.rateLimited["/api/v1/prices"] = 99
		rig.official.mu.Unlock()
		rig.clock.Set(at(11, 0))
		if err := rig.probe.labelTick(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertLabelGap(t, rig.contents(t), "005930", "price_read_failed: rate_limited_gave_up")
	})
	t.Run("missed due window", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
			t.Fatal(err)
		}
		before := rig.official.callCount("/api/v1/prices")
		rig.clock.Set(at(11, 10))
		if err := rig.probe.labelTick(context.Background()); err != nil {
			t.Fatal(err)
		}
		assertLabelGap(t, rig.contents(t), "005930", "missed_due_window")
		if rig.official.callCount("/api/v1/prices") != before {
			t.Fatal("a missed window must not read a price at the wrong horizon")
		}
	})
}

func assertLabelGap(t *testing.T, contents ledgerContents, symbol, reasonPrefix string) {
	t.Helper()
	found := 0
	for _, row := range contents.Judgments {
		if row.Symbol != symbol {
			continue
		}
		label, ok := contents.Labels[row.ID]
		if !ok || label.RealizedReturn != nil || label.RealizedPrice != nil || !strings.HasPrefix(label.LabelGap, reasonPrefix) {
			t.Fatalf("%s label = %+v, want gap %q", row.ID, label, reasonPrefix)
		}
		found++
	}
	if found == 0 {
		t.Fatalf("no judgments for %s", symbol)
	}
}

// TestRateLimitBackoffThenGapRow 는 429 지수 백오프와, 포기 시 원장 결손 기록을 봄.
func TestRateLimitBackoffThenGapRow(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.mu.Lock()
	rig.official.rateLimited["/api/v1/rankings"] = 2 // 재시도 두 번이면 회복
	rig.official.mu.Unlock()
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if got := len(rig.contents(t).Judgments); got != 4 {
		t.Fatalf("recovered cycle wrote %d judgments, want 4", got)
	}
	if !reflect.DeepEqual(rig.clock.slept[:2], []time.Duration{time.Second, 2 * time.Second}) {
		t.Fatalf("backoff = %v, want the configured exponential delays", rig.clock.slept)
	}

	rig.official.mu.Lock()
	rig.official.rateLimited["/api/v1/rankings"] = 99
	rig.official.mu.Unlock()
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 10)); err != nil {
		t.Fatal(err)
	}
	gaps := rig.contents(t).Gaps
	if len(gaps) != 1 || gaps[0].Stage != "rankings" || !strings.HasPrefix(gaps[0].Reason, "rate_limited_gave_up") {
		t.Fatalf("gave-up cycle must leave one rankings gap row: %+v", gaps)
	}
}

// TestOrderbookGapIsCarriedOnTheJudgmentRow 는 종목 단위 결손이 판단 행과 state 양쪽에 남는지 봄.
func TestOrderbookGapIsCarriedOnTheJudgmentRow(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.mu.Lock()
	rig.official.failStatus["/api/v1/orderbook"] = 503
	rig.official.mu.Unlock()
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	// 429 가 아닌 실패는 재시도하지 않음 — 종목당 정확히 한 번.
	if got := rig.official.callCount("/api/v1/orderbook"); got != 2 {
		t.Fatalf("orderbook calls = %d, want 2 (a 503 is not a rate limit and is not retried)", got)
	}
	for _, row := range rig.contents(t).Judgments {
		if hasPrefix(row.CollectionGaps, "top_of_book: rate_limited") {
			t.Fatalf("a 503 must not be reported as a rate-limit give-up: %v", row.CollectionGaps)
		}
		if !hasPrefix(row.CollectionGaps, "top_of_book: ") {
			t.Fatalf("row %s does not carry the orderbook gap: %v", row.ID, row.CollectionGaps)
		}
	}
	var request struct {
		State PublicState `json:"state"`
	}
	_ = json.Unmarshal(rig.typeSafe.requests()[0], &request)
	if request.State.TopOfBook != nil || !hasPrefix(request.State.MissingInputs, "top_of_book: ") {
		t.Fatalf("outbound state does not name the orderbook gap: %+v", request.State)
	}
}

func TestJudgmentFailureLeavesAGapRow(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.typeSafe.statuses = []int{422, 422}
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if len(contents.Judgments) != 0 || len(contents.Gaps) != 2 || contents.Gaps[0].Stage != "judgment" {
		t.Fatalf("rejected judgments must become gap rows: %+v", contents)
	}
	if rig.probe.pendingCount() != 0 {
		t.Fatal("no label may be pending for a judgment that was never made")
	}
}

func hasPrefix(list []string, prefix string) bool {
	for _, item := range list {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}
	return false
}

// TestSessionBoundaries 는 세션 경계: 마감 뒤 거부, 개장 전 거부(대기 플래그 없을 때),
// 휴장 거부, 그리고 판단 시각 + horizon 이 마감을 넘기 전까지만 표본을 뜸.
func TestSessionBoundaries(t *testing.T) {
	t.Parallel()
	t.Run("closed", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		rig.clock.Set(at(16, 0))
		if err := rig.probe.runMarket(context.Background(), "KR"); err == nil || !strings.Contains(err.Error(), "already closed") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("before open without wait", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		rig.clock.Set(at(8, 0))
		if err := rig.probe.runMarket(context.Background(), "KR"); err == nil || !strings.Contains(err.Error(), "--wait-open") {
			t.Fatalf("err = %v", err)
		}
		if rig.official.callCount("/api/v1/rankings") != 0 {
			t.Fatal("no sampling may happen before the open")
		}
	})
	t.Run("holiday", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		rig.official.noRegular = true
		if err := rig.probe.runMarket(context.Background(), "KR"); err == nil || !strings.Contains(err.Error(), "no regular session") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("sampling window", func(t *testing.T) {
		t.Parallel()
		rig := newTestRig(t)
		rig.probe.cfg.WaitOpen = true
		rig.clock.Set(at(8, 30))
		if err := rig.probe.runMarket(context.Background(), "KR"); err != nil {
			t.Fatal(err)
		}
		// 09:00 부터 10분 간격, 마지막 표본은 14:30(+60분 = 15:30 마감). 09:00..14:30 = 34 사이클.
		if got := rig.official.callCount("/api/v1/rankings"); got != 34 {
			t.Fatalf("cycles = %d, want 34", got)
		}
		latest := ""
		for _, row := range rig.contents(t).Judgments {
			if row.At > latest {
				latest = row.At
			}
		}
		if latest != stamp(at(14, 30)) {
			t.Fatalf("last judgment at %s, want %s", latest, stamp(at(14, 30)))
		}
	})
}

// TestRestorePendingAfterRestart 는 재시작 시 라벨 없는 판단을 다시 대기열에 올리는지 봄.
func TestRestorePendingAfterRestart(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	restarted := newProbe(rig.probe.src, rig.probe.judge, rig.probe.ledger, rig.probe.cfg)
	restarted.now, restarted.sleep, restarted.logf = rig.clock.Now, rig.clock.Sleep, func(string, ...any) {}
	restarted.restorePending(rig.contents(t))
	if restarted.pendingCount() != 2 {
		t.Fatalf("restored %d pending groups, want 2 (one per symbol)", restarted.pendingCount())
	}
	rig.clock.Set(at(11, 0))
	if err := restarted.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := len(rig.contents(t).Labels); got != 4 {
		t.Fatalf("labels after restart = %d, want 4", got)
	}
}

func TestLedgerIsAppendOnlyAcrossOpens(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for index := 0; index < 2; index++ {
		l, err := openLedger(dir)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.gap(gapRow{At: stamp(at(10, index)), Market: "KR", Stage: "rankings", Reason: "x"}); err != nil {
			t.Fatal(err)
		}
		_ = l.Close()
	}
	contents, err := readLedger(filepath.Join(dir, ledgerFileName))
	if err != nil || len(contents.Gaps) != 2 {
		t.Fatalf("second open must append, not truncate: %+v %v", contents, err)
	}
	l, _ := openLedger(dir)
	defer l.Close()
	if err := l.label(labelRow{JudgmentID: "x"}); err == nil {
		t.Fatal("a label row with neither a return nor a gap reason must be refused")
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.jsonl"), []byte(`{"kind":"mystery"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readLedger(filepath.Join(dir, "bad.jsonl")); err == nil {
		t.Fatal("an unknown row kind must stop the reader")
	}
}
