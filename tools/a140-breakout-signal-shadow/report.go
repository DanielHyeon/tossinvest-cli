package main

// report.go 는 원장에서 세션별 퍼널 · 거절 분포 · 반사실 · 제안 · 결과 · 호출 예산을 셈(design §5). 판정은 사람이 함.

import (
	"fmt"
	"sort"
	"strings"
)

// minVerdictSamples 는 결과 판정(승률 · 평균 R)을 내는 최소 표본 — spec 「표본이 30 미만인 결과 구간의 판정을 보류」.
const minVerdictSamples = 30

// phaseRank 는 퍼널에서 「어디까지 갔나」의 순서.
var phaseRank = map[string]int{"RANGE_LOCKED": 1, "BREAKOUT_CLOSED": 2, "RETEST_WAIT": 3, "ARMED": 4, "PROPOSED": 5}

type sessionTally struct {
	universe   int
	furthest   map[string]string
	terminal   map[string]int
	refusals   map[string]int
	gaps       map[string]int
	rvol       [3]int
	rvolSeen   map[string]bool
	proposals  int
	attempts   int
	outcomes   map[string]int
	rSum       int64
	labels     int
	reads      map[string]int
	collectMax float64
}

func newTally() *sessionTally {
	return &sessionTally{furthest: map[string]string{}, terminal: map[string]int{}, refusals: map[string]int{}, gaps: map[string]int{},
		rvolSeen: map[string]bool{}, outcomes: map[string]int{}, reads: map[string]int{}}
}

func renderReport(rows []ledgerRow) string {
	tallies := map[string]*sessionTally{}
	var order []string
	get := func(session string) *sessionTally {
		if t, ok := tallies[session]; ok {
			return t
		}
		t := newTally()
		tallies[session] = t
		order = append(order, session)
		return t
	}
	for _, row := range rows {
		t := get(row.Session)
		switch row.Kind {
		case kindUniverse:
			t.universe = len(row.Symbols)
		case kindSignal:
			if phaseRank[row.Phase] > phaseRank[t.furthest[row.Symbol]] {
				t.furthest[row.Symbol] = row.Phase
			}
			if row.Refusal != "" {
				t.refusals[row.Refusal]++
			}
			if row.Attempt > 0 {
				t.attempts++
			}
			if row.Done != "" {
				t.terminal[row.Done]++
			}
			if row.Phase == "PROPOSED" {
				t.proposals++
			}
			if (row.RVOL12 || row.RVOL20 || row.RVOL25) && !t.rvolSeen[row.Symbol] {
				t.rvolSeen[row.Symbol] = true
				for i, hit := range []bool{row.RVOL12, row.RVOL20, row.RVOL25} {
					if hit {
						t.rvol[i]++
					}
				}
			}
		case kindLabel:
			t.labels++
			t.outcomes[row.Outcome]++
			t.rSum += row.RPPM
		case kindGap:
			t.gaps[row.Stage+" "+firstWords(row.Reason)]++
		case kindBudget:
			for group, n := range row.Reads {
				t.reads[group] += n
			}
			t.collectMax = max(t.collectMax, row.CollectSeconds)
		}
	}
	var b strings.Builder
	b.WriteString("# a140 breakout 그림자 신호 보고서\n\n")
	b.WriteString("한계: 가상 사이징(손절 = opening range 저점 · 목표 = 순 RR 1.5)이라 평가기의 RR 거절은 **구성상 공허**하다. 체결 가능성(호가 잔량)은\n")
	b.WriteString("재지 않았다. 결과 판정(승률 · 평균 R)은 표본 n<30 이면 **보류**한다. 판정은 사람이 한다.\n")
	for _, session := range order {
		t := tallies[session]
		fmt.Fprintf(&b, "\n## %s\n\n", session)
		funnel := map[string]int{}
		for _, phase := range t.furthest {
			funnel[phase]++
		}
		fmt.Fprintf(&b, "- 종목 %d · 퍼널(가장 멀리 간 단계): RANGE_LOCKED %d · RETEST_WAIT %d · ARMED %d · PROPOSED %d\n", t.universe,
			funnel["RANGE_LOCKED"]+funnel["BREAKOUT_CLOSED"], funnel["RETEST_WAIT"], funnel["ARMED"], funnel["PROPOSED"])
		fmt.Fprintf(&b, "- 종결: %s\n", joinCounts(t.terminal))
		fmt.Fprintf(&b, "- 거절 코드(행 기준): %s\n", joinCounts(t.refusals))
		fmt.Fprintf(&b, "- RVOL 반사실(돌파 이상 종목): 1.2× %d · 2.0× %d · 2.5× %d\n", t.rvol[0], t.rvol[1], t.rvol[2])
		fmt.Fprintf(&b, "- 제안 %d(ARMED 시도 %d) · 라벨 %d: %s\n", t.proposals, t.attempts, t.labels, joinCounts(t.outcomes))
		if t.labels < minVerdictSamples {
			fmt.Fprintf(&b, "- 결과 판정: 보류(n<30 — n=%d)\n", t.labels)
		} else {
			fmt.Fprintf(&b, "- 승률 %.1f%% · 평균 R %.3f (n=%d)\n", 100*float64(t.outcomes["target"])/float64(t.labels),
				float64(t.rSum)/float64(t.labels)/ppmScale, t.labels)
		}
		fmt.Fprintf(&b, "- 결손: %s\n", joinCounts(t.gaps))
		fmt.Fprintf(&b, "- 호출 예산(그룹별 GET 합): %s · 10 분 창 수집 시간 최대 %.1f 초\n", joinCounts(t.reads), t.collectMax)
	}
	return b.String()
}

func joinCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return "없음"
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = fmt.Sprintf("%s %d", key, counts[key])
	}
	return strings.Join(parts, " · ")
}
