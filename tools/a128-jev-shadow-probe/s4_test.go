package main

// s4_test.go 는 질문 추가 s4(design.md 「질문 추가 (s4)」, tasks 5.x)를 고정함:
// J3·J4 가 같은 요청에 원문대로 실리고, J1·J2 는 바이트 단위로 그대로이며, 선별은 J1·J2 만 읽고,
// 보고서는 J3·J4 적중 경계와 개장 후 경과 시간대를 따로 셈. 옛 원장은 예전 보고를 그대로 냄.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestJ1J2AreByteIdenticalAndPinned 는 J1·J2 의 id·rev·digest 를 s4 이전 값으로 못 박음.
// digest 값은 2026-10-20 공유 원장의 J1·J2 행 각 7,710 개가 가진 값과 같음(grep -c 실측) — 옛 행과 새 행이 같은 모집단.
func TestJ1J2AreByteIdenticalAndPinned(t *testing.T) {
	t.Parallel()
	want := []struct{ id, rev, digest string }{
		{"j1_up_within_horizon", "a128-r1", "sha256:a362ac64cef9d57c051b3aa695fa9350848ba6e356cbc0595fe1d0be82f49d10"},
		{"j2_hold_off_red_flag", "a128-r1", "sha256:ea5fc246294791ded355e8ea838661bf95df2cbd456b6ef098e640f8e5b16737"},
		{"j3_up_1pct_within_horizon", "a128-r2", ""},
		{"j4_down_1pct_within_horizon", "a128-r2", ""},
	}
	questions := probeQuestions()
	if len(questions) != len(want) {
		t.Fatalf("probe asks %d questions, want %d", len(questions), len(want))
	}
	for index, q := range questions {
		if q.ID != want[index].id || q.Rev != want[index].rev {
			t.Errorf("question %d = %s@%s, want %s@%s", index, q.ID, q.Rev, want[index].id, want[index].rev)
		}
		if want[index].digest != "" && q.digest() != want[index].digest {
			t.Errorf("%s digest = %s, want %s — J1·J2 text must stay byte-identical", q.ID, q.digest(), want[index].digest)
		}
	}
}

// TestS4QuestionsRideInTheSameRequest 는 J3·J4 가 J1·J2 와 한 요청에 원문대로 실리고(요청 수 불변),
// 원장 행이 question_id 로 가려지며 같은 state·시각을 공유함을 봄.
func TestS4QuestionsRideInTheSameRequest(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	// 전송 본문의 질문 맵을 원문 리터럴과 바이트 단위로 대조 — 원문이 몰래 바뀌면 여기서 멈춤.
	wantBodies := map[string]string{
		"j3_up_1pct_within_horizon": `{"type":"noul","instructions":"Will this stock's last traded price, exactly ` +
			"`horizon_minutes` minutes after `observed_at`, be at least 1% above `base_price`?" + `","criteria":{"true":` +
			`"The last traded price at ` + "`observed_at` plus `horizon_minutes` is at least 1.01 times `base_price`." +
			`","false":"That price is below 1.01 times ` + "`base_price`." + `"}}`,
		"j4_down_1pct_within_horizon": `{"type":"noul","instructions":"Will this stock's last traded price, exactly ` +
			"`horizon_minutes` minutes after `observed_at`, be at least 1% below `base_price`?" + `","criteria":{"true":` +
			`"The last traded price at ` + "`observed_at` plus `horizon_minutes` is at most 0.99 times `base_price`." +
			`","false":"That price is above 0.99 times ` + "`base_price`." + `"}}`,
	}
	requests := rig.typeSafe.requests()
	if len(requests) != 2 {
		t.Fatalf("requests = %d, want 2 (one per symbol — the new questions add no request)", len(requests))
	}
	for _, body := range requests {
		var request struct {
			Questions map[string]json.RawMessage `json:"questions"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatal(err)
		}
		if len(request.Questions) != 4 {
			t.Fatalf("request carries %d questions, want 4", len(request.Questions))
		}
		for id, want := range wantBodies {
			var compact bytes.Buffer
			if err := json.Compact(&compact, request.Questions[id]); err != nil {
				t.Fatalf("%s missing or not JSON: %v", id, err)
			}
			if compact.String() != want {
				t.Fatalf("%s body =\n%s\nwant\n%s", id, compact.String(), want)
			}
		}
	}
	type key struct{ at, symbol, digest string }
	perSymbol := map[key]map[string]float64{}
	for _, row := range rig.contents(t).Judgments {
		k := key{row.At, row.Symbol, row.StateDigest}
		if perSymbol[k] == nil {
			perSymbol[k] = map[string]float64{}
		}
		perSymbol[k][row.QuestionID] = row.P
	}
	if len(perSymbol) != 2 {
		t.Fatalf("rows group into %d (at, symbol, state) keys, want 2", len(perSymbol))
	}
	for k, byQuestion := range perSymbol {
		want := map[string]float64{"j1_up_within_horizon": 0.93, "j2_hold_off_red_flag": 0.12,
			"j3_up_1pct_within_horizon": 0.31, "j4_down_1pct_within_horizon": 0.22}
		if !reflect.DeepEqual(byQuestion, want) {
			t.Fatalf("%v rows by question_id = %v, want %v", k, byQuestion, want)
		}
	}
}

// TestScreeningIgnoresJ3J4 는 선별(결선·순위·악재 제외)이 J1·J2 만 읽음을 봄: J1·J2 가 같고 J3·J4 만 서로 반대
// 순서인 두 사이클의 선별 기록이 종목마다 같고, 결선이 J1·J2 규칙으로 손계산한 집합과 같아야 함.
// J3·J4 순서를 J1 과 반대(A)·같은 방향(B)으로 두어, 선별이 J3·J4 를 읽으면 적어도 한쪽이 손계산과 어긋나게 함.
func TestScreeningIgnoresJ3J4(t *testing.T) {
	t.Parallel()
	if fields := reflect.TypeOf(screenInput{}).NumField(); fields != 2 {
		t.Fatalf("screenInput has %d fields, want 2 (J1, J2)", fields)
	}
	var rows []rankingRow
	for index := 0; index < 15; index++ {
		rows = append(rows, rankingRow{fmt.Sprintf("S%02d", index), "1000", "KRW"})
	}
	screensWith := func(j3, j4 func(index int) float64) map[string]screenRecord {
		rig := newTestRig(t)
		rig.official.setRanking(rows...)
		rig.typeSafe.answerFor = func(symbol, questionID string) string {
			var index int
			fmt.Sscanf(symbol, "S%02d", &index)
			j2 := 0.10
			if index%5 == 0 {
				j2 = 0.80 // S00·S05·S10 악재 제외
			}
			value := map[string]float64{
				screenUpQuestion:              0.10 + 0.03*float64(index), // 번호가 클수록 J1 높음
				screenRedFlagQuestion:         j2,
				"j3_up_1pct_within_horizon":   j3(index),
				"j4_down_1pct_within_horizon": j4(index),
			}[questionID]
			return fmt.Sprintf(`{"type":"noul","noul":%.4f}`, value)
		}
		if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
			t.Fatal(err)
		}
		out := map[string]screenRecord{}
		for _, row := range rig.contents(t).Judgments {
			if previous, seen := out[row.Symbol]; seen && previous != *row.Screen {
				t.Fatalf("%s rows disagree on the screen record: %+v vs %+v", row.Symbol, previous, *row.Screen)
			}
			out[row.Symbol] = *row.Screen
		}
		return out
	}
	against := screensWith(func(i int) float64 { return 0.90 - 0.05*float64(i) }, func(i int) float64 { return 0.05 * float64(i) })
	along := screensWith(func(i int) float64 { return 0.05 * float64(i) }, func(i int) float64 { return 0.90 - 0.05*float64(i) })
	if len(against) != 15 || !reflect.DeepEqual(against, along) {
		t.Fatalf("screening changed with J3/J4 only:\n%+v\n%+v", against, along)
	}
	// 손계산: 생존 12(S00·S05·S10 제외) 중 J1 내림차순 상위 10 → S01·S02 가 순위 밖(11·12위).
	wantFinalists := []string{"S03", "S04", "S06", "S07", "S08", "S09", "S11", "S12", "S13", "S14"}
	for _, screens := range []map[string]screenRecord{against, along} {
		var finalists []string
		for symbol, record := range screens {
			if record.Finalist {
				finalists = append(finalists, symbol)
			}
		}
		sort.Strings(finalists)
		if !reflect.DeepEqual(finalists, wantFinalists) || screens["S14"].Rank != 1 || screens["S01"].Rank != 12 ||
			!screens["S05"].RedFlagExcluded {
			t.Fatalf("finalists = %v (S14 rank %d, S01 rank %d), want %v from J1/J2 alone",
				finalists, screens["S14"].Rank, screens["S01"].Rank, wantFinalists)
		}
	}
}

// TestS4HitRuleBoundaries 는 J3 ⇔ ≥ +0.01, J4 ⇔ ≤ −0.01 경계(경계 자체는 참)를 규칙 함수와 보고서 표 양쪽에서 봄.
func TestS4HitRuleBoundaries(t *testing.T) {
	t.Parallel()
	j3, j4 := hitRules["j3_up_1pct_within_horizon"], hitRules["j4_down_1pct_within_horizon"]
	if j3.hit == nil || j4.hit == nil {
		t.Fatal("J3/J4 have no hit rule — the report would skip their tables")
	}
	for r, want := range map[float64]bool{0.01: true, 0.0099: false, 0.02: true, 0: false, -0.01: false} {
		if got := j3.hit(r); got != want {
			t.Errorf("J3 hit(%v) = %v, want %v", r, got, want)
		}
	}
	for r, want := range map[float64]bool{-0.01: true, -0.0099: false, -0.02: true, 0: false, 0.01: false} {
		if got := j4.hit(r); got != want {
			t.Errorf("J4 hit(%v) = %v, want %v", r, got, want)
		}
	}
	// 보고서 경로: 정확히 경계인 15 건은 적중, 경계 바로 안쪽 15 건은 빗나감 → 0.500.
	boundary := func(questionID string, onEdge, inside float64) string {
		var rows []struct {
			p   float64
			ret *float64
		}
		for index := 0; index < 15; index++ {
			rows = append(rows, struct {
				p   float64
				ret *float64
			}{0.95, ret(onEdge)}, struct {
				p   float64
				ret *float64
			}{0.95, ret(inside)})
		}
		return renderCalibrationReport(syntheticLedger(questionID, rows), "l")
	}
	for questionID, report := range map[string]string{
		"j3_up_1pct_within_horizon":   boundary("j3_up_1pct_within_horizon", 0.01, 0.0099),
		"j4_down_1pct_within_horizon": boundary("j4_down_1pct_within_horizon", -0.01, -0.0099),
	} {
		if !strings.Contains(report, "## "+questionID+" — KR") || !strings.Contains(report, "| 0.95–1.00 | 30 | 30 | 15 | 0.500 | 판정 가능 |") {
			t.Fatalf("%s boundary table is wrong:\n%s", questionID, report)
		}
	}
}

// TestLabelerStoresAnExactOnePercentMoveAsTheBoundary 는 경계 판정의 전제(라벨러가 소수 8 자리로 반올림)를 실제
// writeLabels 로 확인함: 3.00 → 3.03 처럼 부동소수 나눗셈이 0.00999… 를 내는 경우도 0.01 로 저장돼 J3 적중이 됨.
func TestLabelerStoresAnExactOnePercentMoveAsTheBoundary(t *testing.T) {
	t.Parallel()
	// 변수로 계산해야 함 — 상수식은 Go 가 정확한 산술로 접어 0.01 을 냄.
	base, up := 3.00, 3.03
	if raw := (up - base) / base; raw >= 0.01 {
		t.Fatalf("fixture no longer exercises float error: raw = %v", raw)
	}
	rig := newTestRig(t)
	cases := map[string]struct{ base, realized float64 }{"up": {3.00, 3.03}, "down": {3.00, 2.97}, "kr": {10000, 10100}}
	for name, c := range cases {
		realized := c.realized
		item := pendingLabel{Market: "US", Symbol: name, BasePrice: c.base, Due: at(11, 0), JudgmentIDs: []string{name}}
		if err := rig.probe.writeLabels(item, at(11, 0), &realized, ""); err != nil {
			t.Fatal(err)
		}
	}
	labels := rig.contents(t).Labels
	if r := *labels["up"].RealizedReturn; !hitRules["j3_up_1pct_within_horizon"].hit(r) {
		t.Fatalf("+1%% move stored as %v is not a J3 hit", r)
	}
	if r := *labels["kr"].RealizedReturn; !hitRules["j3_up_1pct_within_horizon"].hit(r) {
		t.Fatalf("+1%% KR move stored as %v is not a J3 hit", r)
	}
	if r := *labels["down"].RealizedReturn; !hitRules["j4_down_1pct_within_horizon"].hit(r) {
		t.Fatalf("−1%% move stored as %v is not a J4 hit", r)
	}
}

// TestSessionBucketBoundaries 는 개장 후 경과 분 구간 경계(0–29 · 30–60 · 61–120 · 121+ · 알 수 없음).
func TestSessionBucketBoundaries(t *testing.T) {
	t.Parallel()
	for minutes, want := range map[int]int{0: 0, 29: 0, 30: 1, 60: 1, 61: 2, 120: 2, 121: 3, 390: 3, -1: 4} {
		if got := sessionBucket(minutes, true); got != want {
			t.Errorf("sessionBucket(%d) = %d, want %d", minutes, got, want)
		}
	}
	if got := sessionBucket(10, false); got != len(sessionBuckets) {
		t.Errorf("unknown minutes went to bucket %d", got)
	}
}

// reportSection 은 제목 줄부터 다음 「## 」 제목 전까지를 돌려줌.
func reportSection(t *testing.T, report, heading string) string {
	t.Helper()
	start := strings.Index(report, heading+"\n")
	if start < 0 {
		t.Fatalf("report lacks %q:\n%s", heading, report)
	}
	rest := report[start+len(heading)+1:]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		rest = rest[:end]
	}
	return rest
}

// TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay 는 s4 이전 모양의 원장(testdata/pre-s4 — top10 옛 행·s3 선별 행·
// 라벨 결손·미라벨·gap 행)이 그대로 읽히고, report 출력이 base 코드(bbe4ceaa6)로 만든 골든을 **접두로 바이트 그대로**
// 포함하며 그 뒤에는 시간대 표만 덧붙음을 봄. state 파일 없음·해시 불일치는 「알 수 없음」 줄로 셈.
func TestPreS4LedgerStillReportsTheSameAndAddsTimeOfDay(t *testing.T) {
	t.Parallel()
	golden, err := os.ReadFile(filepath.Join("testdata", "pre-s4", "report-base.golden.md"))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := reportCommand([]string{"--ledger", "testdata/pre-s4/ledger.jsonl"}, &out); err != nil {
		t.Fatalf("an old ledger must still parse: %v", err)
	}
	report := out.String()
	if !strings.HasPrefix(report, string(golden)) {
		t.Fatalf("the pre-s4 report changed — new output must keep the base output as its exact prefix:\n%s", report)
	}
	added := strings.TrimPrefix(report, string(golden))
	for _, line := range strings.Split(added, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "## 개장 후 경과 시간대 — ") {
			t.Fatalf("the addition carries a non time-of-day section %q", line)
		}
	}
	const s3Lot = "jev_screen10_of_random30_of_top100_le_order_cap@a128-s3"
	for heading, want := range map[string][]string{
		"## 개장 후 경과 시간대 — j1_up_within_horizon · KR · 로트 " + s3Lot: {
			"| 0–29 | 0 | 0 | 0 | — | — | 보류(n<30) |",
			"| 30–60 | 1 | 1 | 0 | — | 0.520 | 보류(n<30) |",
			"| 61–120 | 0 | 0 | 0 | — | — | 보류(n<30) |",
			"| 121+ | 0 | 0 | 0 | — | — | 보류(n<30) |",
			"| 알 수 없음 | 2 | 2 | 2 | — | 0.555 | 보류(n<30) |", // state 파일 없음 1 + 해시 불일치 1
		},
		"## 개장 후 경과 시간대 — j1_up_within_horizon · US · 로트 " + s3Lot: {
			"| 61–120 | 1 | 0 | 0 | — | — | 보류(n<30) |", // 라벨 결손
			"| 121+ | 1 | 0 | 0 | — | — | 보류(n<30) |",   // 미라벨
			"| 알 수 없음 | 0 | 0 | 0 | — | — | 보류(n<30) |",
		},
		"## 개장 후 경과 시간대 — j1_up_within_horizon · KR · 로트 top10": {
			"| 0–29 | 1 | 1 | 1 | — | 0.550 | 보류(n<30) |",
		},
		"## 개장 후 경과 시간대 — j2_hold_off_red_flag · ALL · 로트 " + s3Lot: {
			"| 30–60 | 1 | 1 | 1 | — | 0.300 | 보류(n<30) |",
			"| 알 수 없음 | 2 | 2 | 0 | — | 0.225 | 보류(n<30) |",
		},
	} {
		section := reportSection(t, report, heading)
		for _, line := range want {
			if !strings.Contains(section, line) {
				t.Fatalf("%s lacks %q:\n%s", heading, line, section)
			}
		}
	}
}

// TestUnknownQuestionGetsNoTimeOfDayTable 는 적중 규칙이 없는 질문이 구간표처럼 시간대 표도 만들지 않음을 봄(FLM B20).
func TestUnknownQuestionGetsNoTimeOfDayTable(t *testing.T) {
	t.Parallel()
	report := renderCalibrationReport(syntheticLedger("zz_unknown_question", rowsAt(0.6, 1, 0)), "l")
	if !strings.Contains(report, "적중 규칙이 정의되지 않은 질문") || strings.Contains(report, "## 개장 후 경과 시간대 — zz_unknown_question") {
		t.Fatalf("an unknown question must get neither a p table nor a time-of-day table:\n%s", report)
	}
}

// TestTimeOfDayTableJudgesAtThirtyLabels 는 시간대 표도 같은 n<30 보류 규칙을 쓰고, 구간별로 섞이지 않음을 봄.
func TestTimeOfDayTableJudgesAtThirtyLabels(t *testing.T) {
	t.Parallel()
	contents := syntheticLedger("j3_up_1pct_within_horizon", append(rowsAt(0.6, 30, 0), rowsAt(0.4, 0, 29)...))
	contents.StateMinutes = map[string]int{}
	for index := range contents.Judgments {
		digest := fmt.Sprintf("sha256:%d", index)
		contents.Judgments[index].StateDigest = digest
		contents.StateMinutes[digest] = 10 // 앞 30 = 개장 30분 안
		if index >= 30 {
			contents.StateMinutes[digest] = 200
		}
	}
	section := reportSection(t, renderCalibrationReport(contents, "l"), "## 개장 후 경과 시간대 — j3_up_1pct_within_horizon · KR · 로트 top10")
	for _, want := range []string{
		"| 0–29 | 30 | 30 | 30 | 1.000 | 0.600 | 판정 가능 |",
		"| 121+ | 29 | 29 | 0 | — | 0.400 | 보류(n<30) |",
		"| 알 수 없음 | 0 | 0 | 0 | — | — | 보류(n<30) |",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("time-of-day table lacks %q:\n%s", want, section)
		}
	}
}
