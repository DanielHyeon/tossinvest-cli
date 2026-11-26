package main

// a203_verdict.go 는 판정 사전등록(design §7)과 부트스트랩 — 결과를 보기 전에 커밋한 숫자 그대로.

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

const (
	a203Seed            = 203
	a203BootstrapRounds = 2000
	a203MinLabeled      = minVerdictSamples // 30 — a140 report.go:12
	a203MinSessions     = 20                // 결정 M5
	a203MaxGapPercent   = 20                // a184 unobserved 와 같은 상한(초과면 INSUFFICIENT)

	a203Pass         = "PASS"
	a203Fail         = "FAIL"
	a203Insufficient = "INSUFFICIENT"
)

// a203Judgement 는 한 가정의 판정 지표.
type a203Judgement struct {
	Verdict          string
	Reasons          []string
	Labeled          int
	Unlabeled        int
	ProposalSessions int
	Gaps, Total      int
	Mean, Lower      float64
	ExcludingBest    float64
	Daily            []float64 // 날짜 오름차순 일 평균(제안 0 건 세션 제외)
}

// a203Judge 는 §7 — trades 는 한 가정의 제안 전부, gaps/total 은 결손 종목-세션 / 전체 종목-세션.
func a203Judge(trades []a203Trade, gaps, total int) a203Judgement {
	j := a203Judgement{Gaps: gaps, Total: total}
	bySession := map[string][]float64{}
	for _, t := range trades {
		if !t.Labeled {
			j.Unlabeled++
			continue
		}
		j.Labeled++
		bySession[t.Session] = append(bySession[t.Session], t.Return)
	}
	sessions := make([]string, 0, len(bySession))
	for s := range bySession {
		sessions = append(sessions, s) // "KRX:YYYY-MM-DD" — 사전순 = 날짜순
	}
	sort.Strings(sessions)
	for _, s := range sessions {
		j.Daily = append(j.Daily, a203Mean(bySession[s]))
	}
	j.ProposalSessions = len(j.Daily)
	if len(j.Daily) > 0 {
		j.Mean = a203Mean(j.Daily)
		j.Lower = a203BootstrapLower(j.Daily, a203Seed)
		j.ExcludingBest = a203ExcludingBest(j.Daily)
	}
	if j.Labeled < a203MinLabeled {
		j.Reasons = append(j.Reasons, "labeled proposals < 30")
	}
	if j.ProposalSessions < a203MinSessions {
		j.Reasons = append(j.Reasons, "sessions with a proposal < 20")
	}
	if gaps*100 > total*a203MaxGapPercent {
		j.Reasons = append(j.Reasons, "collection gaps > 20% of symbol-sessions")
	}
	if j.Unlabeled > 0 {
		j.Reasons = append(j.Reasons, "a proposal has no label")
	}
	switch {
	case len(j.Reasons) > 0:
		j.Verdict = a203Insufficient
	case j.Mean > 0 && j.Lower > 0 && j.ExcludingBest > 0:
		j.Verdict = a203Pass
	default:
		j.Verdict = a203Fail
	}
	return j
}

func a203Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// a203ExcludingBest 는 가장 좋은 하루 하나를 뺀 평균(같은 값이 여럿이면 하나만 뺌). 하루뿐이면 0.
func a203ExcludingBest(days []float64) float64 {
	if len(days) < 2 {
		return 0
	}
	best := 0
	for i, v := range days {
		if v > days[best] {
			best = i
		}
	}
	rest := append(append([]float64(nil), days[:best]...), days[best+1:]...)
	return a203Mean(rest)
}

// a203Percentile 과 a203BootstrapLower 는 a182 report.go(브랜치 feat/a182-kr-crash-day-rule-replay e0890b870)의
// percentile · bootstrapLower 와 같은 정의(design §7 — a182 는 main 에 없어 같은 식을 여기 다시 적음, 시험이 같은 값을 고정).
func a203Percentile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if q < 0 {
		q = 0
	}
	if q > 1 {
		q = 1
	}
	p := q * float64(len(sorted)-1)
	i := int(p)
	if i == len(sorted)-1 {
		return sorted[i]
	}
	return sorted[i] + (sorted[i+1]-sorted[i])*(p-float64(i))
}

func a203BootstrapLower(days []float64, seed int64) float64 {
	if len(days) == 0 {
		return 0
	}
	r := rand.New(rand.NewSource(seed))
	out := make([]float64, a203BootstrapRounds)
	for i := range out {
		sum := 0.0
		for range days {
			sum += days[r.Intn(len(days))]
		}
		out[i] = sum / float64(len(days))
	}
	sort.Float64s(out)
	return a203Percentile(out, .025)
}

// ---- 보고서 (design §4 첫 줄 · §5 한계 · §6 대조 · §7 판정 · §11) — 종목 코드 · 가격은 적지 않음 ----

// a203ReportInput 은 보고서 한 벌의 재료.
type a203ReportInput struct {
	sample                 string
	in                     a203Inputs
	outcomes               []a203Outcome
	lot, cacheSHA          string
	cacheFiles             int
	ledgerPath, ledgerSHA  string
	ledgerLines, ledgerLen int
	runs                   []a203RunRecord
	costPPMNow             uint64
	costErr                error
}

const (
	a203QuoteLimitLine    = "과거 호가가 없어 호가 관문(스프레드 · 호가 나이)은 재생하지 않았고, 괴리와 체결가는 그 분 시가를 ask 로 대신함(P0). 체결 가능성(호가 잔량)도 재지 않음."
	a203UniverseLimitLine = "종목군 U-C 한계: P(a) 는 11월 엔진 관측 · P(b) 는 지금 6mo 랭킹이라 6 ~ 10월 세션에 모집단 미래 정보 · 생존 편향이 있음, 541 은 하루 33 시점 랭킹의 합이라 엔진 진입 검토 집합보다 넓음, B 몫 R 규칙은 재생 전용 근사(엔진 장중 순위와 다름), 세션별 선정은 d−1 까지의 일봉만 씀."
)

type a203Key struct{ session, symbol string }

// a203A140Terminal 은 a140 원장의 종목-세션별 「ARMED 전 종결」(재생 a203Outcome.Terminal 과 같은 분류).
func a203A140Terminal(rows []ledgerRow) map[a203Key]string {
	out := map[a203Key]string{}
	for _, row := range rows {
		if row.Kind != kindSignal {
			continue
		}
		key := a203Key{row.Session, row.Symbol}
		switch {
		case out[key] == "ARMED":
		case row.Phase == "ARMED" || row.Phase == "PROPOSED" || row.Attempt > 0:
			out[key] = "ARMED"
		default:
			out[key] = row.Phase
		}
	}
	return out
}

type a203A140Proposal struct {
	ask, worst, roundTrip, exit uint64
	labeled                     bool
}

func a203A140Proposals(rows []ledgerRow) map[a203Key]*a203A140Proposal {
	out := map[a203Key]*a203A140Proposal{}
	bySetup := map[string]*a203A140Proposal{}
	for _, row := range rows {
		if row.Kind == kindSignal && row.Phase == "PROPOSED" {
			p := &a203A140Proposal{ask: row.Ask, worst: row.Worst, roundTrip: row.RoundTrip}
			out[a203Key{row.Session, row.Symbol}], bySetup[row.SetupID] = p, p
		}
	}
	for _, row := range rows {
		if p := bySetup[row.SetupID]; row.Kind == kindLabel && p != nil {
			p.exit, p.labeled = row.ExitMinor, true
		}
	}
	return out
}

// a203ControlMismatches 는 대조 1 의 불일치 줄(종목 코드 포함 — 저장소 밖 파일로만).
func a203ControlMismatches(in a203Inputs, outcomes []a203Outcome) []string {
	a140 := a203A140Terminal(in.a140Rows)
	var out []string
	for _, o := range outcomes {
		if o.Universe == a203UniverseA140 && o.Terminal != a140[a203Key{o.Session, o.Symbol}] {
			out = append(out, fmt.Sprintf("%s %s replay=%q a140=%q gap=%q", o.Session, o.Symbol, o.Terminal, a140[a203Key{o.Session, o.Symbol}], o.Gap))
		}
	}
	return out
}

func a203Counts(m map[string]int) string { return joinCounts(m) }

func a203Pct(n, d int) string {
	if d == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", 100*float64(n)/float64(d))
}

func a203WriteJudgement(b *strings.Builder, name string, j a203Judgement, primary bool) {
	label := "판정"
	if !primary {
		label = "지표(판정 아님)"
	}
	fmt.Fprintf(b, "- %s %s: %s", name, label, j.Verdict)
	if len(j.Reasons) > 0 {
		fmt.Fprintf(b, " (%s)", strings.Join(j.Reasons, "; "))
	}
	fmt.Fprintf(b, "\n  - 라벨 제안 %d · 라벨 없음 %d · 제안 세션 %d · 결손 %d/%d (%s)\n", j.Labeled, j.Unlabeled, j.ProposalSessions, j.Gaps, j.Total, a203Pct(j.Gaps, j.Total))
	fmt.Fprintf(b, "  - 일 평균의 평균 %+.4f%% · 부트스트랩 0.025 하한(seed %d · %d 회) %+.4f%% · 최고일 제외 평균 %+.4f%%\n",
		100*j.Mean, a203Seed, a203BootstrapRounds, 100*j.Lower, 100*j.ExcludingBest)
}

// a203RenderReport 는 보고서 본문. 첫 줄 호가 한계, 둘째 줄 종목군 한계(design §4 · §5).
func a203RenderReport(r a203ReportInput) string {
	var b strings.Builder
	b.WriteString(a203QuoteLimitLine + "\n")
	b.WriteString(a203UniverseLimitLine + "\n\n")
	fmt.Fprintf(&b, "# a203 돌파 KR 과거 재생 보고서 — 표본 %s\n\n", r.sample)
	fmt.Fprintf(&b, "- 판: %s\n- 세션 목록 sha256 %s · P(a) %d · P(b) %d · P %d\n", r.lot, r.in.sessionsSHA, len(r.in.pa), len(r.in.pb), len(r.in.population))
	var uc []a203Outcome
	var c1 []a203Outcome
	for _, o := range r.outcomes {
		if o.Universe == a203UniverseA140 {
			c1 = append(c1, o)
		} else {
			uc = append(uc, o)
		}
	}
	gaps := 0
	trades := map[string][]a203Trade{}
	for _, o := range uc {
		if o.Gap != "" {
			gaps++
		}
		for _, a := range a203Assumptions {
			if t := o.Trades[a]; t != nil {
				trades[a] = append(trades[a], *t)
			}
		}
	}
	if r.sample == a203SampleControl {
		a203WriteControls(&b, r, c1, uc)
	}
	fmt.Fprintf(&b, "\n## 판정(design §7, 주 가정 P0 — 종목군 U-C %d 종목-세션)\n\n", len(uc))
	if r.sample == a203SampleControl {
		b.WriteString("- 대조 표본(5 세션)은 판정에 쓰지 않음 — 아래 지표는 참고.\n")
	}
	for _, a := range a203Assumptions {
		a203WriteJudgement(&b, a, a203Judge(trades[a], gaps, len(uc)), a == a203AssumptionP0 && r.sample == a203SampleMain)
	}
	a203WriteFunnel(&b, uc)
	a203WriteShares(&b, uc)
	a203WriteRotation(&b, r)
	b.WriteString("\n## 캐시 · 원장 · 호출\n\n")
	fmt.Fprintf(&b, "- 캐시 디렉터리 지문 sha256 %s · 파일 %d\n", r.cacheSHA, r.cacheFiles)
	fmt.Fprintf(&b, "- 원장 %s · sha256 %s · %d 줄 · %d 바이트(커밋 안 함 — 결정 M4)\n", r.ledgerPath, r.ledgerSHA, r.ledgerLines, r.ledgerLen)
	total, byKind, stops := 0, map[string]int{}, 0
	for _, run := range r.runs {
		total += run.Gets
		for k, v := range run.ByKind {
			byKind[k] += v
		}
		if run.Stop != "" {
			stops++
		}
	}
	fmt.Fprintf(&b, "- 수집 실행 %d 회(정지 %d) · GET 합 %d: %s · 이 보고서 실행 GET 0\n", len(r.runs), stops, total, a203Counts(byKind))
	if r.costErr != nil {
		fmt.Fprintf(&b, "- 비용: a140 판 CostPPM %d · 현재 costs.DefaultModel 읽기 실패 %v\n", r.in.params.CostPPM, r.costErr)
	} else {
		fmt.Fprintf(&b, "- 비용: a140 판 CostPPM %d(판정에 씀) · 현재 costs.DefaultModel %d ppm(참고)\n", r.in.params.CostPPM, r.costPPMNow)
	}
	b.WriteString("\n## 한계(design §11)\n\n")
	for i, line := range []string{"호가 관문 미재생, 체결 가능성 미측정.", "종목군 U-C 한계(둘째 줄).", "일봉은 표준 해독기(엄격 리더 없음).",
		"비용은 a140 판 CostPPM 하나(세율 변경 시점 미반영).", "VI · 단일가 구간이 만든 빈 분은 GAP 로 셈.", "한 시장 국면(2026 하반기)의 결과일 수 있음.",
		"원장은 커밋하지 않으므로 재계산은 로컬 원장에 의존(커밋된 것은 지문뿐)."} {
		fmt.Fprintf(&b, "%d. %s\n", i+1, line)
	}
	return b.String()
}

func a203WriteFunnel(b *strings.Builder, outcomes []a203Outcome) {
	furthest, terminal, refusals, gaps := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	evalErrors := 0
	for _, o := range outcomes {
		if o.Furthest != "" {
			furthest[o.Furthest]++
		}
		if o.Terminal != "" {
			terminal[o.Terminal]++
		}
		for k, v := range o.Refusals {
			refusals[k] += v
		}
		for _, a := range o.Attempts[a203AssumptionP0] {
			if a.Eval.Refusal != "" {
				refusals["P0 attempt "+a.Eval.Refusal]++
			}
		}
		if o.Gap != "" {
			gaps[firstWords(o.Gap)]++
		}
		evalErrors += o.EvalErrors
	}
	b.WriteString("\n## 퍼널 · 거절 · 결손\n\n")
	fmt.Fprintf(b, "- 가장 멀리 간 단계: %s\n- ARMED 전 종결: %s\n- 거절 코드(바뀐 순간 · P0 시도): %s\n- 결손 사유: %s · 평가 오류 %d\n",
		a203Counts(furthest), a203Counts(terminal), a203Counts(refusals), a203Counts(gaps), evalErrors)
}

func a203WriteShares(b *strings.Builder, outcomes []a203Outcome) {
	b.WriteString("\n## 몫별(P0 — 판정은 30 개 전체로만)\n\n")
	for _, share := range []string{a203ShareA, a203ShareB, a203ShareFill} {
		n, rs := 0, []float64(nil)
		for _, o := range outcomes {
			if o.Share != share {
				continue
			}
			n++
			if t := o.Trades[a203AssumptionP0]; t != nil && t.Labeled {
				rs = append(rs, t.Return)
			}
		}
		fmt.Fprintf(b, "- %s: 종목-세션 %d · 제안 %d · 건당 평균 순 %+.4f%%\n", share, n, len(rs), 100*a203Mean(rs))
	}
}

// a203WriteRotation 은 세션별 종목군 회전율(전 표본 세션 대비 겹침 = 교집합 / 30) · 고유 종목 수 · 채움 수 · 자격 수.
func a203WriteRotation(b *strings.Builder, r a203ReportInput) {
	b.WriteString("\n## 종목군 회전율 · 고유 종목(엔진 실측 겹침: 당일 분모 중앙 71.3%, 58.8 ~ 77.4% — design §5)\n\n")
	var prev map[string]bool
	unique := map[string]bool{}
	var overlaps []float64
	var lines []string
	for _, s := range r.in.sessions {
		picks, ok := r.in.universes[s.Date]
		if !ok || s.Role != r.sample {
			continue
		}
		cur, fill := map[string]bool{}, 0
		for _, p := range picks {
			cur[p.Symbol], unique[p.Symbol] = true, true
			if p.Share == a203ShareFill {
				fill++
			}
		}
		overlap := "—"
		if prev != nil {
			both := 0
			for code := range cur {
				if prev[code] {
					both++
				}
			}
			v := float64(both) / float64(2*a203PerShare)
			overlaps = append(overlaps, v)
			overlap = fmt.Sprintf("%.1f%%", 100*v)
		}
		lines = append(lines, fmt.Sprintf("%s 선정 %d(자격 %d · 채움 %d) 겹침 %s", s.Date, len(picks), r.in.eligible[s.Date], fill, overlap))
		prev = cur
	}
	fmt.Fprintf(b, "- 세션 %d · 고유 종목 %d · 겹침 평균 %.1f%% · τ(회전율 평균) %.3f\n", len(lines), len(unique), 100*a203Mean(overlaps), 1-a203Mean(overlaps))
	for _, line := range lines {
		b.WriteString("  - " + line + "\n")
	}
}

// a203WriteControls 는 대조 1 · 2(design §6 · tasks 3.1 ~ 3.3).
func a203WriteControls(b *strings.Builder, r a203ReportInput, c1, uc []a203Outcome) {
	a140Terminal := a203A140Terminal(r.in.a140Rows)
	a140Props := a203A140Proposals(r.in.a140Rows)
	match := 0
	for _, o := range c1 {
		if o.Terminal == a140Terminal[a203Key{o.Session, o.Symbol}] {
			match++
		}
	}
	b.WriteString("\n## 대조 1 — a140 종목군 그대로(파이프라인 차이만)\n\n")
	fmt.Fprintf(b, "- ARMED 전 종결 일치 %d / %d · 불일치 %d", match, len(c1), len(c1)-match)
	if match != len(c1) {
		b.WriteString(" — **fail-stop: 봉 차이를 설명하기 전에는 주 표본을 수집하지 않음**(목록은 상태 디렉터리 control1-mismatches.txt)")
	}
	b.WriteString("\n")
	both, a140Only, replayOnly := 0, 0, 0
	ticks := map[string]int{}
	var diffs []float64
	for _, o := range c1 {
		p, t := a140Props[a203Key{o.Session, o.Symbol}], o.Trades[a203AssumptionP0]
		switch {
		case p != nil && t != nil:
			both++
			d := (int64(t.Plan.Ask) - int64(p.ask)) / int64(krTickMinor(p.ask))
			ticks[fmt.Sprintf("%+d", d)]++
			if p.labeled && t.Labeled && p.worst > 0 {
				diffs = append(diffs, t.Return-(float64(p.exit)-float64(p.worst)-float64(p.roundTrip))/float64(p.worst))
			}
		case p != nil:
			a140Only++
		case t != nil:
			replayOnly++
		}
	}
	fmt.Fprintf(b, "- PROPOSED: a140 %d 건 · 둘 다 %d · a140 만 %d · 재생(P0)만 %d\n", len(a140Props), both, a140Only, replayOnly)
	fmt.Fprintf(b, "- 대리 ask − a140 ask(틱, 둘 다 제안): %s · 건당 순 수익률 차(재생 − a140) 평균 %+.4f%% (n=%d) — P0 가정의 실측 오차\n",
		a203Counts(ticks), 100*a203Mean(diffs), len(diffs))
	b.WriteString("\n## 대조 2 — U-C 종목군(세션별 a140 종목군과 나란히)\n\n")
	for _, date := range a203ControlDates {
		a140Set := map[string]bool{}
		for _, code := range r.in.a140U[date] {
			a140Set[code] = true
		}
		overlap := 0
		for _, p := range r.in.universes[date] {
			if a140Set[p.Symbol] {
				overlap++
			}
		}
		shares := map[string][]float64{}
		proposals := 0
		for _, o := range uc {
			if o.Session != "KRX:"+date {
				continue
			}
			if t := o.Trades[a203AssumptionP0]; t != nil {
				proposals++
				if t.Labeled {
					shares[o.Share] = append(shares[o.Share], t.Return)
				}
			}
		}
		var a140R []float64
		for key, p := range a140Props {
			if key.session == "KRX:"+date && p.labeled && p.worst > 0 {
				a140R = append(a140R, (float64(p.exit)-float64(p.worst)-float64(p.roundTrip))/float64(p.worst))
			}
		}
		fmt.Fprintf(b, "- %s: 겹침 %d/30 · U-C 제안 %d (A %d 평균 %+.4f%% · B %d 평균 %+.4f%% · 채움 %d 평균 %+.4f%%) · a140 라벨 %d 평균 %+.4f%%\n",
			date, overlap, proposals, len(shares[a203ShareA]), 100*a203Mean(shares[a203ShareA]), len(shares[a203ShareB]), 100*a203Mean(shares[a203ShareB]),
			len(shares[a203ShareFill]), 100*a203Mean(shares[a203ShareFill]), len(a140R), 100*a203Mean(a140R))
	}
}
