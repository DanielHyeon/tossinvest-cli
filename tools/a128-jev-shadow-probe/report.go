package main

// report.go 는 원장을 p 구간표로 바꿈(design.md 「임계 판정 보고서」).
//
// 규칙:
//   - 구간: p < 0.50 한 줄 + 0.50~1.00 을 0.05 폭 10 칸(마지막 칸은 1.00 포함).
//   - 적중: J1 = 실현 수익률 > 0, J2 = 실현 수익률 < 0(악재 판단 뒤 h분 수익률이 음수).
//   - 구간의 실현 라벨 n < 30 이면 그 구간은 「보류」 — 적중률 숫자를 싣지 않음
//     (작은 표본의 비율이 임계 근거로 인용되는 일을 양식에서 막음).
//   - 표본 선택 로트(selection.rule, 없는 옛 행 = top10)마다 표를 따로 만듦 — 상위 10 고정과
//     상한 이하 무작위 10 은 모집단이 다르므로 한 표에 섞지 않음(tasks.md 1.5).
//   - 선별 기록(screen, 1.6)이 있는 로트는 J1 에 대해 결선 vs 비결선(대조군) 비교 표를 추가로 냄 —
//     같은 적중 규칙·같은 n<30 보류 규칙. 비교는 적중 빈도만 싣고 수익성 결론은 내지 않음.
//   - 결정은 사람이 함. 보고서는 예비 권고만 냄.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	minBinSamples   = 30
	thresholdP      = 0.90
	binCount        = 10
	allMarketsLabel = "ALL"
)

// hitRules 는 질문 id → 적중 판정. 여기 없는 질문은 보고서가 표를 만들지 않고 이름만 적음.
var hitRules = map[string]struct {
	rule string
	hit  func(realizedReturn float64) bool
}{
	"j1_up_within_horizon": {"realized return > 0", func(r float64) bool { return r > 0 }},
	"j2_hold_off_red_flag": {"realized return < 0", func(r float64) bool { return r < 0 }},
	// s4: 경계 ±0.01 은 참에 포함 — 라벨러가 수익률을 소수 8 자리로 반올림해 적으므로 정확히 1% 는 0.01 로 저장됨.
	"j3_up_1pct_within_horizon":   {"realized return ≥ +0.01", func(r float64) bool { return r >= 0.01 }},
	"j4_down_1pct_within_horizon": {"realized return ≤ -0.01", func(r float64) bool { return r <= -0.01 }},
}

// sessionBuckets 는 개장 후 경과 분(state 의 minutes_since_open) 구간 — 개장 30분(0–29)을 따로 보이기 위함(s4).
// 각 구간은 upTo 이하를 담음. 마지막 줄 「알 수 없음」 = state 파일 없음·해시 불일치·값 없음·음수.
var sessionBuckets = []struct {
	label string
	upTo  int
}{{"0–29", 29}, {"30–60", 60}, {"61–120", 120}, {"121+", math.MaxInt}}

const unknownSessionBucket = "알 수 없음"

// sessionBucket 은 구간 번호를 돌려줌. 알 수 없으면 len(sessionBuckets).
func sessionBucket(minutes int, known bool) int {
	if !known || minutes < 0 {
		return len(sessionBuckets)
	}
	for index, bucket := range sessionBuckets {
		if minutes <= bucket.upTo {
			return index
		}
	}
	return len(sessionBuckets)
}

const twoDayLimit = "이 보고서는 사용자 지시(2026-10-07)에 따른 **2일 측정**(KR 2세션 + US 2세션)의 산출물이다. " +
	"구간당 표본이 작아 어떤 판정도 임계의 **확정이 아니라 예비 판정**이다. 같은 날·같은 종목군의 표본은 " +
	"서로 독립이 아니므로(시장 공통 움직임) 실제 불확실성은 n 이 말하는 것보다 크다. 원장 포맷은 측정 연장을 " +
	"전제로 하며, 임계의 생산 배선은 이 보고서가 아니라 사람의 결정과 별도 change 를 거친다."

type binStats struct {
	judgments int
	labeled   int
	hits      int
}

// binIndex 는 p 를 구간 번호로 바꿈: -1 = p < 0.50, 0..9 = 0.05 폭 구간.
// 정수 퍼센트로 내려 계산함 — 0.55 같은 경계값이 부동소수 오차로 아래 칸에 떨어지지 않게.
func binIndex(p float64) int {
	percent := int(math.Floor(p*100 + 1e-9))
	if percent < 50 {
		return -1
	}
	return min((percent-50)/5, binCount-1)
}

func binLabel(index int) string {
	if index < 0 {
		return "< 0.50"
	}
	low := 0.50 + 0.05*float64(index)
	if index == binCount-1 {
		return fmt.Sprintf("%.2f–1.00", low)
	}
	return fmt.Sprintf("%.2f–%.2f", low, low+0.05)
}

// renderCalibrationReport 는 원장 내용을 Markdown 보고서로 만듦.
func renderCalibrationReport(contents ledgerContents, ledgerPath string) string {
	var out strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&out, format+"\n", args...) }

	realized, labelGaps, unlabeled := 0, 0, 0
	revs, models, markets, questionIDs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	lots := map[string]int{}
	for _, row := range contents.Judgments {
		lots[selectionLot(row)]++
		revs[row.QuestionRev+" "+row.QuestionDigest] = true
		models[row.ModelAnswered] = true
		markets[row.Market] = true
		questionIDs[row.QuestionID] = true
		label, found := contents.Labels[row.ID]
		switch {
		case !found:
			unlabeled++
		case label.RealizedReturn == nil:
			labelGaps++
		default:
			realized++
		}
	}
	gapStages := map[string]int{}
	for _, row := range contents.Gaps {
		gapStages[row.Market+" "+row.Stage]++
	}

	line("# a128 Jev shadow calibration report")
	line("")
	line("- 원장: `%s`", ledgerPath)
	line("- judgment 행 %d · 실현 라벨 %d · 라벨 결손 %d · 미라벨 %d · gap 행 %d",
		len(contents.Judgments), realized, labelGaps, unlabeled, len(contents.Gaps))
	if len(gapStages) > 0 {
		line("- gap 단계별: %s", joinCounts(gapStages))
	}
	if len(lots) > 0 {
		line("- 표본 선택 로트: %s — 로트마다 표를 따로 둠(섞지 않음)", joinCounts(lots))
	}
	for _, budget := range cycleBudgets(contents) {
		line("%s", budget)
	}
	line("- 질문 rev·digest: %s", joinKeys(revs))
	line("- 답한 모델: %s", joinKeys(models))
	if len(revs) > len(questionIDs) || len(models) > 1 {
		line("- **경고: 질문 원문 또는 모델 버전이 섞여 있음 — 모집단이 하나가 아니다. 구간표를 rev·모델별로 나눠 읽을 것.**")
	}
	line("")
	line("## 한계")
	line("")
	line("%s", twoDayLimit)
	line("")
	line("구간의 실현 라벨이 %d 미만이면 그 구간은 「보류」이며 적중률을 싣지 않는다. 라벨 결손·미라벨 판단은 "+
		"적중률의 분모에서 빠지고 위 행 수로만 드러난다.", minBinSamples)

	marketList := sortedKeys(markets)
	if len(marketList) > 1 {
		marketList = append(marketList, allMarketsLabel)
	}
	for _, questionID := range sortedKeys(questionIDs) {
		rule, known := hitRules[questionID]
		if !known {
			line("")
			line("## %s", questionID)
			line("")
			line("적중 규칙이 정의되지 않은 질문 — 표를 만들지 않음.")
			continue
		}
		for _, lot := range sortedKeys(keysOf(lots)) {
			for _, market := range marketList {
				renderQuestionSection(&out, contents, lot, questionID, market, rule.rule, rule.hit)
			}
		}
	}
	if questionIDs[screenUpQuestion] {
		rule := hitRules[screenUpQuestion]
		for _, lot := range sortedKeys(screenedLots(contents)) {
			for _, market := range marketList {
				renderScreenComparison(&out, contents, lot, market, rule.rule, rule.hit)
			}
		}
	}
	// s4 시간대 표는 맨 뒤에 덧붙임 — 그 앞의 기존 출력은 바이트 그대로 남음(골든 접두 시험).
	for _, questionID := range sortedKeys(questionIDs) {
		rule, known := hitRules[questionID]
		if !known {
			continue
		}
		for _, lot := range sortedKeys(keysOf(lots)) {
			for _, market := range marketList {
				renderSessionTimeSection(&out, contents, lot, questionID, market, rule.rule, rule.hit)
			}
		}
	}
	return out.String()
}

// renderSessionTimeSection 은 한 질문·한 로트·한 시장의 행을 개장 후 경과 시간대로 나눠 셈.
// 평균 p 는 실현 라벨 행만으로 냄 — 같은 행의 적중률과 나란히 읽게(보정 대조).
func renderSessionTimeSection(out *strings.Builder, contents ledgerContents, lot, questionID, market, rule string, hit func(float64) bool) {
	line := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	stats := make([]binStats, len(sessionBuckets)+1)
	sumP := make([]float64, len(sessionBuckets)+1)
	for _, row := range contents.Judgments {
		if selectionLot(row) != lot || row.QuestionID != questionID || (market != allMarketsLabel && row.Market != market) {
			continue
		}
		minutes, known := contents.StateMinutes[row.StateDigest]
		index := sessionBucket(minutes, known)
		stats[index].judgments++
		label, found := contents.Labels[row.ID]
		if !found || label.RealizedReturn == nil {
			continue
		}
		stats[index].labeled++
		sumP[index] += row.P
		if hit(*label.RealizedReturn) {
			stats[index].hits++
		}
	}

	line("")
	line("## 개장 후 경과 시간대 — %s · %s · 로트 %s", questionID, market, lot)
	line("")
	line("적중 규칙: %s. 구간 = state 파일의 minutes_since_open(개장 30분 = 0–29). 「%s」 = state 파일 없음·해시 불일치·값 없음.",
		rule, unknownSessionBucket)
	line("")
	line("| 개장 후 분 | 판단 n | 실현 라벨 n | 적중 | 적중률 | 평균 p | 판정 |")
	line("|---|---:|---:|---:|---:|---:|---|")
	for index, s := range stats {
		name := unknownSessionBucket
		if index < len(sessionBuckets) {
			name = sessionBuckets[index].label
		}
		rate, meanP, verdict := "—", "—", fmt.Sprintf("보류(n<%d)", minBinSamples)
		if s.labeled > 0 {
			meanP = fmt.Sprintf("%.3f", sumP[index]/float64(s.labeled))
		}
		if s.labeled >= minBinSamples {
			rate = fmt.Sprintf("%.3f", float64(s.hits)/float64(s.labeled))
			verdict = "판정 가능"
		}
		line("| %s | %d | %d | %d | %s | %s | %s |", name, s.judgments, s.labeled, s.hits, rate, meanP, verdict)
	}
}

// readStateMinutes 는 원장이 참조하는 state 파일에서 minutes_since_open 만 읽음(읽기 전용 — 쓰기·이동 없음).
// 파일 내용(끝 줄바꿈 제외)의 sha256 이 행의 state_digest 와 같을 때만 씀 — 판단한 그 state 의 값만 구간에 들어가게.
// 파일 없음·읽기 실패·해시 불일치·값 없음은 맵에 넣지 않음 → 보고서의 「알 수 없음」 줄로 셈(조용히 빼지 않음).
func readStateMinutes(statesDir string, judgments []judgmentRow) map[string]int {
	minutes := map[string]int{}
	seen := map[string]bool{}
	for _, row := range judgments {
		digest := row.StateDigest
		if seen[digest] {
			continue
		}
		seen[digest] = true
		// 순수 16진 64 자만 파일 이름으로 씀 — 원장 값이 states/ 밖 경로를 가리킬 수 없게.
		hexDigest, ok := strings.CutPrefix(digest, "sha256:")
		if decoded, err := hex.DecodeString(hexDigest); !ok || err != nil || len(decoded) != sha256.Size {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(statesDir, hexDigest+".json"))
		if err != nil {
			continue
		}
		raw = bytes.TrimSuffix(raw, []byte("\n"))
		if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != hexDigest {
			continue
		}
		var state struct {
			MinutesSinceOpen *int `json:"minutes_since_open"`
		}
		if json.Unmarshal(raw, &state) != nil || state.MinutesSinceOpen == nil {
			continue
		}
		minutes[digest] = *state.MinutesSinceOpen
	}
	return minutes
}

// cycleBudgets 는 예산 관측 필드(1.7)가 있는 로트마다 사이클(시장·시각) 단위의 GET 수·수집 시간 요약 한 줄.
// 간격 절반을 넘은 사이클은 다음 사이클이 밀리기 시작한 신호(공유 Open API 예산을 사람이 판단할 재료).
func cycleBudgets(contents ledgerContents) []string {
	type cycle struct{ reads, seconds, interval, wts float64 }
	byLot := map[string]map[string]cycle{}
	for _, row := range contents.Judgments {
		if row.Selection == nil || row.Selection.IntervalSeconds == 0 {
			continue
		}
		lot := selectionLot(row)
		if byLot[lot] == nil {
			byLot[lot] = map[string]cycle{}
		}
		byLot[lot][row.Market+"|"+row.At] = cycle{float64(row.Selection.BrokerReads), row.Selection.CollectSeconds,
			float64(row.Selection.IntervalSeconds), float64(row.Selection.WTSRequests)}
	}
	lots := make([]string, 0, len(byLot))
	for lot := range byLot {
		lots = append(lots, lot)
	}
	sort.Strings(lots)
	var lines []string
	for _, lot := range lots {
		var reads, seconds, wts []float64
		over := 0
		for _, c := range byLot[lot] {
			reads, seconds, wts = append(reads, c.reads), append(seconds, c.seconds), append(wts, c.wts)
			if c.seconds > c.interval/2 {
				over++
			}
		}
		line := fmt.Sprintf("- 사이클 예산 %s: 사이클 %d · broker GET 중앙 %.0f · 최대 %.0f · 수집 중앙 %.0fs · 최대 %.0fs · 간격 절반 초과 %d",
			lot, len(reads), median(reads), maxOf(reads), median(seconds), maxOf(seconds), over)
		// 1.8(s3) 로트만 WTS 요청이 있음 — 공유 Open API 예산과 다른 서버라 따로 적음.
		if maxOf(wts) > 0 {
			line += fmt.Sprintf(" · WTS 요청 중앙 %.0f · 최대 %.0f", median(wts), maxOf(wts))
		}
		lines = append(lines, line)
	}
	return lines
}

func median(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted)%2 == 1 {
		return sorted[len(sorted)/2]
	}
	return (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
}

func maxOf(values []float64) float64 {
	out := math.Inf(-1)
	for _, value := range values {
		out = math.Max(out, value)
	}
	return out
}

// screenedLots 는 선별 기록이 있는 행을 가진 로트 — 옛 로트(top10·random10)는 비교 대상이 아님.
func screenedLots(contents ledgerContents) map[string]bool {
	lots := map[string]bool{}
	for _, row := range contents.Judgments {
		if row.Screen != nil {
			lots[selectionLot(row)] = true
		}
	}
	return lots
}

// renderScreenComparison 은 한 로트·한 시장의 J1 행을 결선 / 비결선(악재 제외 + 생존·순위 밖)으로 나눠 셈.
// 비결선은 같은 사이클에서 같은 질문으로 판단·라벨된 대조군임.
func renderScreenComparison(out *strings.Builder, contents ledgerContents, lot, market, rule string, hit func(float64) bool) {
	line := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	var finalist, redFlag, outranked binStats
	for _, row := range contents.Judgments {
		if row.Screen == nil || selectionLot(row) != lot || row.QuestionID != screenUpQuestion ||
			(market != allMarketsLabel && row.Market != market) {
			continue
		}
		stats := &outranked
		switch {
		case row.Screen.Finalist:
			stats = &finalist
		case row.Screen.RedFlagExcluded:
			stats = &redFlag
		}
		stats.judgments++
		label, found := contents.Labels[row.ID]
		if !found || label.RealizedReturn == nil {
			continue
		}
		stats.labeled++
		if hit(*label.RealizedReturn) {
			stats.hits++
		}
	}
	control := binStats{
		judgments: redFlag.judgments + outranked.judgments,
		labeled:   redFlag.labeled + outranked.labeled,
		hits:      redFlag.hits + outranked.hits,
	}

	line("")
	line("## 결선 vs 비결선 (%s) — %s · 로트 %s", screenUpQuestion, market, lot)
	line("")
	line("적중 규칙: %s. 결선 = J2 p < %.1f 생존 중 J1 p 상위; 비결선 = 같은 풀의 나머지(대조군). "+
		"적중 빈도 비교일 뿐 수익성 판단이 아니다.", rule, redFlagExclusionP)
	line("")
	line("| 집단 | 판단 n | 실현 라벨 n | 적중 | 적중률 | 판정 |")
	line("|---|---:|---:|---:|---:|---|")
	for _, group := range []struct {
		name  string
		stats binStats
	}{{"결선", finalist}, {"비결선 전체", control}, {"└ J2 악재 제외", redFlag}, {"└ 생존·순위 밖", outranked}} {
		rate, verdict := "—", fmt.Sprintf("보류(n<%d)", minBinSamples)
		if group.stats.labeled >= minBinSamples {
			rate = fmt.Sprintf("%.3f", float64(group.stats.hits)/float64(group.stats.labeled))
			verdict = "판정 가능"
		}
		line("| %s | %d | %d | %d | %s | %s |", group.name, group.stats.judgments, group.stats.labeled, group.stats.hits, rate, verdict)
	}
}

func keysOf(counts map[string]int) map[string]bool {
	set := make(map[string]bool, len(counts))
	for key := range counts {
		set[key] = true
	}
	return set
}

// renderQuestionSection 은 한 로트·한 질문·한 시장의 표 하나. lot 이 다른 행은 세지 않음.
func renderQuestionSection(out *strings.Builder, contents ledgerContents, lot, questionID, market, rule string, hit func(float64) bool) {
	line := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	bins := make([]binStats, binCount+1) // [0] = p < 0.50, [1..] = 구간 0..9
	var above binStats
	for _, row := range contents.Judgments {
		if selectionLot(row) != lot || row.QuestionID != questionID || (market != allMarketsLabel && row.Market != market) {
			continue
		}
		stats := &bins[binIndex(row.P)+1]
		stats.judgments++
		if row.P >= thresholdP {
			above.judgments++
		}
		label, found := contents.Labels[row.ID]
		if !found || label.RealizedReturn == nil {
			continue
		}
		stats.labeled++
		isHit := hit(*label.RealizedReturn)
		if isHit {
			stats.hits++
		}
		if row.P >= thresholdP {
			above.labeled++
			if isHit {
				above.hits++
			}
		}
	}

	line("")
	line("## %s — %s · 로트 %s", questionID, market, lot)
	line("")
	line("적중 규칙: %s", rule)
	line("")
	line("| p 구간 | 판단 n | 실현 라벨 n | 적중 | 적중률 | 판정 |")
	line("|---|---:|---:|---:|---:|---|")
	for index := -1; index < binCount; index++ {
		stats := bins[index+1]
		rate, verdict := "—", fmt.Sprintf("보류(n<%d)", minBinSamples)
		if stats.labeled >= minBinSamples {
			rate = fmt.Sprintf("%.3f", float64(stats.hits)/float64(stats.labeled))
			verdict = "판정 가능"
		}
		line("| %s | %d | %d | %d | %s | %s |", binLabel(index), stats.judgments, stats.labeled, stats.hits, rate, verdict)
	}
	line("")
	line("### p ≥ %.2f 임계 (예비)", thresholdP)
	line("")
	line("- 판단 n %d · 실현 라벨 n %d · 적중 %d", above.judgments, above.labeled, above.hits)
	switch {
	case above.labeled < minBinSamples:
		line("- 권고: **보류** — 실현 라벨 %d < %d. 2일 표본으로는 이 임계를 판정할 수 없다.", above.labeled, minBinSamples)
	default:
		rate := float64(above.hits) / float64(above.labeled)
		if rate >= thresholdP {
			line("- 권고: **예비 채택 후보** — 적중률 %.3f ≥ %.2f. 확정이 아니며 결정은 사람이 한다.", rate, thresholdP)
		} else {
			line("- 권고: **예비 기각 후보** — 적중률 %.3f < %.2f (p 가 실현 빈도를 과대평가). 결정은 사람이 한다.", rate, thresholdP)
		}
	}
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func joinKeys(set map[string]bool) string {
	if len(set) == 0 {
		return "(없음)"
	}
	return "`" + strings.Join(sortedKeys(set), "`, `") + "`"
}

func joinCounts(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, ", ")
}
