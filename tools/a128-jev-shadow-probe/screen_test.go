package main

// screen_test.go 는 1.6 Jev 선별(풀 전체 판단 → J2 악재 제외 → J1 내림차순 상위 K 결선)을 고정함.
// 실 API 는 부르지 않음 — httptest mock 만.

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestScreenFinalistsRules 는 순수 선별 함수: J2 ≥ 0.5 제외(경계 0.5 제외·0.4999 생존), J1 내림차순,
// 동률은 랭킹 순(입력 순서) 우선, 생존이 K 미만이면 생존 전부가 결선.
func TestScreenFinalistsRules(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		in   []screenInput
		k    int
		want []screenRecord
	}{
		"red flag boundary": {
			in: []screenInput{{J1: 0.9, J2: 0.5}, {J1: 0.8, J2: 0.4999}, {J1: 0.7, J2: 0.51}},
			k:  2,
			want: []screenRecord{
				{RedFlagExcluded: true},
				{Finalist: true, Rank: 1},
				{RedFlagExcluded: true},
			},
		},
		"J1 descending with a cut at K": {
			in: []screenInput{{J1: 0.2}, {J1: 0.9}, {J1: 0.5}, {J1: 0.7}},
			k:  2,
			want: []screenRecord{
				{Rank: 4},
				{Finalist: true, Rank: 1},
				{Rank: 3},
				{Finalist: true, Rank: 2},
			},
		},
		"tie goes to the earlier ranking": {
			in: []screenInput{{J1: 0.4}, {J1: 0.6}, {J1: 0.6}, {J1: 0.6}},
			k:  2,
			want: []screenRecord{
				{Rank: 4},
				{Finalist: true, Rank: 1},
				{Finalist: true, Rank: 2},
				{Rank: 3},
			},
		},
		"fewer survivors than K": {
			in: []screenInput{{J1: 0.3, J2: 0.1}, {J1: 0.9, J2: 0.7}, {J1: 0.6, J2: 0.2}},
			k:  10,
			want: []screenRecord{
				{Finalist: true, Rank: 2},
				{RedFlagExcluded: true},
				{Finalist: true, Rank: 1},
			},
		},
		"empty": {in: nil, k: 10, want: []screenRecord{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := screenFinalists(tc.in, tc.k); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("screenFinalists = %+v\nwant             %+v", got, tc.want)
			}
		})
	}
}

// TestScreenQuestionsAreAsked 는 선별이 읽는 두 질문 id 가 실제로 묻는 질문에 있음을 고정함 —
// 질문 집합이 바뀌면 선별이 0 값을 p 로 오인하지 않게(cycle 은 빠진 p 를 오류로 올림).
func TestScreenQuestionsAreAsked(t *testing.T) {
	t.Parallel()
	ids := map[string]bool{}
	for _, q := range probeQuestions() {
		ids[q.ID] = true
	}
	if !ids[screenUpQuestion] || !ids[screenRedFlagQuestion] {
		t.Fatalf("probe questions %v lack the screen inputs %q/%q", ids, screenUpQuestion, screenRedFlagQuestion)
	}
}

// screenAnswers 는 종목별 (J1, J2) — 정의 안 된 종목은 J1 0.10·J2 0.10.
func screenAnswers(table map[string][2]float64) func(symbol, questionID string) string {
	return func(symbol, questionID string) string {
		pair, found := table[symbol]
		if !found {
			pair = [2]float64{0.10, 0.10}
		}
		value := pair[0]
		if questionID == screenRedFlagQuestion {
			value = pair[1]
		}
		// %.4f 로 보내 같은 리터럴이 같은 float64 로 읽히게 함(동률 시험).
		return fmt.Sprintf(`{"type":"noul","noul":%.4f}`, value)
	}
}

// TestCycleJudgesTheWholePoolAndFlagsOnlyTheFinalists 는 사이클 단위 측정 설계를 고정함:
// 풀 30 전부를 판단해 행으로 남기고(비결선 = 대조군) 전부 라벨 대기에 올리며, 결선 표시는 정확히
// 선별 규칙의 10 종목에만 붙음. 판단 실패 종목은 gap 행이고 선별 대상이 아님.
func TestCycleJudgesTheWholePoolAndFlagsOnlyTheFinalists(t *testing.T) {
	t.Parallel()
	var rows []rankingRow
	answers := map[string][2]float64{}
	for index := 0; index < 30; index++ {
		symbol := fmt.Sprintf("S%02d", index)
		rows = append(rows, rankingRow{symbol, "1000", "KRW"})
		answers[symbol] = [2]float64{0.10 + 0.01*float64(index), 0.10}
	}
	answers["S00"] = [2]float64{1.00, 0.50}   // J1 최고지만 J2 경계 0.5 → 제외
	answers["S01"] = [2]float64{0.99, 0.4999} // J2 경계 바로 아래 → 생존·1위
	answers["S29"] = [2]float64{0.39, 0.90}   // 악재 → 제외
	answers["S03"] = [2]float64{0.29, 0.10}   // S19(0.29)와 동률 — 랭킹이 앞선 S03 이 10위
	rig := newTestRig(t)
	rig.official.setRanking(rows...)
	rig.typeSafe.answerFor = screenAnswers(answers)
	rig.typeSafe.failSymbols = map[string]bool{"S28": true} // 판단 실패(생존했다면 2위)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)

	// 풀 30 중 판단 성공 29 종목 × 질문 2 = 58 행, 실패 1 종목은 gap(judgment).
	if len(contents.Judgments) != 58 {
		t.Fatalf("judgment rows = %d, want 58 (29 judged pool symbols × 2 questions)", len(contents.Judgments))
	}
	if len(contents.Gaps) != 1 || contents.Gaps[0].Stage != "judgment" || contents.Gaps[0].Symbol != "S28" {
		t.Fatalf("gaps = %+v, want one judgment gap for S28", contents.Gaps)
	}
	if got := rig.probe.pendingCount(); got != 29 {
		t.Fatalf("pending label groups = %d, want 29 (every judged pool symbol, finalist or not)", got)
	}

	wantFinalists := []string{"S01", "S03", "S20", "S21", "S22", "S23", "S24", "S25", "S26", "S27"}
	wantRank := map[string]int{"S01": 1, "S27": 2, "S20": 9, "S03": 10, "S19": 11, "S02": 27, "S00": 0, "S29": 0}
	finalists := map[string]bool{}
	for _, row := range contents.Judgments {
		if row.Screen == nil {
			t.Fatalf("judgment row %s lacks its screen record", row.ID)
		}
		if row.Screen.Finalist {
			finalists[row.Symbol] = true
		}
		if rank, pinned := wantRank[row.Symbol]; pinned && row.Screen.Rank != rank {
			t.Errorf("%s screen_rank = %d, want %d", row.Symbol, row.Screen.Rank, rank)
		}
		redFlag := row.Symbol == "S00" || row.Symbol == "S29"
		if row.Screen.RedFlagExcluded != redFlag {
			t.Errorf("%s red_flag_excluded = %v, want %v", row.Symbol, row.Screen.RedFlagExcluded, redFlag)
		}
		if row.Screen.Finalist && (row.Screen.Rank < 1 || row.Screen.Rank > 10) {
			t.Errorf("%s is a finalist with rank %d", row.Symbol, row.Screen.Rank)
		}
	}
	if got := sortedKeys(finalists); !reflect.DeepEqual(got, wantFinalists) {
		t.Fatalf("finalists = %v\nwant        %v", got, wantFinalists)
	}

	got := onlySelection(t, contents)
	want := selectionRecord{
		Rule: selectionRule, Seed: testSeed, CapValue: 500000, CapCurrency: "KRW", PriceCurrency: "KRW",
		PriceSource: selectionPriceSource, RankingRequested: 100, RankingReceived: 30, Candidates: 30,
		Pool: 30, Selected: 30, Judged: 29, RedFlagExcluded: 2, FinalistsRequested: 10, Finalists: 10,
	}
	if got != want {
		t.Fatalf("selection = %+v\nwant        %+v", got, want)
	}

	// 대조군까지 전부 라벨링됨.
	rig.clock.Set(at(11, 0))
	if err := rig.probe.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := len(rig.contents(t).Labels); got != 58 {
		t.Fatalf("labels = %d, want 58 (finalists and non-finalists alike)", got)
	}
}

// TestCycleRecordsFewerFinalistsWhenFewSurvive 는 생존이 K 미만이면 결선도 그만큼만이고 그 수가 기록됨.
func TestCycleRecordsFewerFinalistsWhenFewSurvive(t *testing.T) {
	t.Parallel()
	var rows []rankingRow
	answers := map[string][2]float64{}
	for index := 0; index < 12; index++ {
		symbol := fmt.Sprintf("S%02d", index)
		rows = append(rows, rankingRow{symbol, "1000", "KRW"})
		answers[symbol] = [2]float64{0.60, 0.80} // 전부 악재
	}
	answers["S04"] = [2]float64{0.55, 0.20}
	answers["S09"] = [2]float64{0.65, 0.20}
	rig := newTestRig(t)
	rig.official.setRanking(rows...)
	rig.typeSafe.answerFor = screenAnswers(answers)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if len(contents.Judgments) != 24 {
		t.Fatalf("judgment rows = %d, want 24 — red-flagged pool symbols stay in the ledger", len(contents.Judgments))
	}
	finalists := map[string]int{}
	for _, row := range contents.Judgments {
		if row.Screen.Finalist {
			finalists[row.Symbol] = row.Screen.Rank
		}
	}
	if !reflect.DeepEqual(finalists, map[string]int{"S09": 1, "S04": 2}) {
		t.Fatalf("finalists = %v, want S09=1 S04=2", finalists)
	}
	got := onlySelection(t, contents)
	if got.Selected != 12 || got.Judged != 12 || got.RedFlagExcluded != 10 || got.Finalists != 2 || got.FinalistsRequested != 10 {
		t.Fatalf("selection = %+v, want 12 drawn/judged, 10 red-flagged, 2 of 10 finalists", got)
	}
}

// TestPoolSamplingDrawsThePoolNotTheTop 는 후보가 풀보다 많으면 정확히 풀 크기를 무작위로(랭킹 상위가 아님),
// 이하면 전부를 판단하고, 결선은 --top 만큼임.
func TestPoolSamplingDrawsThePoolNotTheTop(t *testing.T) {
	t.Parallel()
	var rows []rankingRow
	for index := 0; index < 40; index++ {
		rows = append(rows, rankingRow{fmt.Sprintf("S%02d", index), "1000", "KRW"})
	}
	rig := newTestRig(t)
	rig.official.setRanking(rows...)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	symbols := judgedSymbols(contents)
	if len(symbols) != 30 {
		t.Fatalf("judged %d symbols, want the pool of 30 out of 40 candidates", len(symbols))
	}
	top30 := make([]string, 30)
	for index := range top30 {
		top30[index] = rows[index].symbol
	}
	if reflect.DeepEqual(symbols, top30) {
		t.Fatalf("the pool is the ranking top 30 — sampling is not random: %v", symbols)
	}
	if got := onlySelection(t, contents); got.Candidates != 40 || got.Pool != 30 || got.Selected != 30 || got.Finalists != 10 {
		t.Fatalf("selection = %+v, want 40 candidates / pool 30 drawn / 10 finalists", got)
	}

	few := newTestRig(t)
	few.official.setRanking(rows[:7]...)
	if err := few.probe.cycle(context.Background(), "KR", few.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if got := judgedSymbols(few.contents(t)); len(got) != 7 {
		t.Fatalf("below the pool size every candidate is judged; judged %v", got)
	}
	if got := onlySelection(t, few.contents(t)); got.Candidates != 7 || got.Selected != 7 || got.Finalists != 7 {
		t.Fatalf("selection = %+v, want 7 drawn and 7 finalists", got)
	}
}

// TestRunFlagsValidatePoolAgainstTop 는 --pool 이 --top 이상·100 이하여야 함을 봄.
func TestRunFlagsValidatePoolAgainstTop(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		top, pool int
		ok        bool
	}{
		{10, 30, true}, {10, 10, true}, {1, 100, true},
		{10, 9, false}, {10, 101, false}, {0, 30, false}, {101, 101, false},
	} {
		err := validateRunFlags(tc.top, tc.pool, 60, defaultInterval)
		if (err == nil) != tc.ok {
			t.Errorf("top %d pool %d: err = %v, want ok=%v", tc.top, tc.pool, err, tc.ok)
		}
		if err != nil && !strings.Contains(err.Error(), "--pool") {
			t.Errorf("the refusal must name --pool: %v", err)
		}
	}
	// run 하위 명령이 실제로 그 검사를 거침(자격증명 읽기 전에 거부).
	getenv := func(name string) string {
		if name == typeSafeAPIKeyEnv {
			return testTypeSafeKey
		}
		return ""
	}
	err := dispatch(context.Background(), []string{"run", "--market", "KR", "--top", "10", "--pool", "5",
		"--config-dir", t.TempDir()}, getenv, nil)
	if err == nil || !strings.Contains(err.Error(), "--pool") {
		t.Fatalf("run --pool 5 --top 10: err = %v", err)
	}
}
