package main

// report.go 는 원장을 p 구간표로 바꿈(design.md 「임계 판정 보고서」).
//
// 규칙:
//   - 구간: p < 0.50 한 줄 + 0.50~1.00 을 0.05 폭 10 칸(마지막 칸은 1.00 포함).
//   - 적중: J1 = 실현 수익률 > 0, J2 = 실현 수익률 < 0(악재 판단 뒤 h분 수익률이 음수).
//   - 구간의 실현 라벨 n < 30 이면 그 구간은 「보류」 — 적중률 숫자를 싣지 않음
//     (작은 표본의 비율이 임계 근거로 인용되는 일을 양식에서 막음).
//   - 결정은 사람이 함. 보고서는 예비 권고만 냄.

import (
	"fmt"
	"math"
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
	for _, row := range contents.Judgments {
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
		for _, market := range marketList {
			renderQuestionSection(&out, contents, questionID, market, rule.rule, rule.hit)
		}
	}
	return out.String()
}

func renderQuestionSection(out *strings.Builder, contents ledgerContents, questionID, market, rule string, hit func(float64) bool) {
	line := func(format string, args ...any) { fmt.Fprintf(out, format+"\n", args...) }
	bins := make([]binStats, binCount+1) // [0] = p < 0.50, [1..] = 구간 0..9
	var above binStats
	for _, row := range contents.Judgments {
		if row.QuestionID != questionID || (market != allMarketsLabel && row.Market != market) {
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
	line("## %s — %s", questionID, market)
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
