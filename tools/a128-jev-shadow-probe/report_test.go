package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// syntheticLedger 는 질문 하나·시장 하나에 p 와 수익률을 지정해 원장 내용을 만듦.
func syntheticLedger(questionID string, rows []struct {
	p   float64
	ret *float64
}) ledgerContents {
	contents := ledgerContents{Labels: map[string]labelRow{}}
	for index, row := range rows {
		id := fmt.Sprintf("j-%d", index)
		contents.Judgments = append(contents.Judgments, judgmentRow{
			ID: id, Market: "KR", QuestionID: questionID, QuestionRev: questionRev, QuestionDigest: "sha256:q",
			ModelAnswered: "jev-1.13.0", P: row.p,
		})
		label := labelRow{JudgmentID: id, RealizedReturn: row.ret}
		if row.ret == nil {
			label.LabelGap = "realized_price_missing"
		}
		contents.Labels[id] = label
	}
	return contents
}

func ret(v float64) *float64 { return &v }

func rowsAt(p float64, hits, misses int) []struct {
	p   float64
	ret *float64
} {
	var rows []struct {
		p   float64
		ret *float64
	}
	for index := 0; index < hits; index++ {
		rows = append(rows, struct {
			p   float64
			ret *float64
		}{p, ret(0.01)})
	}
	for index := 0; index < misses; index++ {
		rows = append(rows, struct {
			p   float64
			ret *float64
		}{p, ret(-0.01)})
	}
	return rows
}

// TestReportHoldsBinsBelowThirtySamples 는 보류 규칙: 구간 n<30 이면 적중률을 싣지 않음.
func TestReportHoldsBinsBelowThirtySamples(t *testing.T) {
	t.Parallel()
	contents := syntheticLedger("j1_up_within_horizon", rowsAt(0.92, 20, 9)) // 29건
	report := renderCalibrationReport(contents, "ledger.jsonl")
	if !strings.Contains(report, "| 0.90–0.95 | 29 | 29 | 20 | — | 보류(n<30) |") {
		t.Fatalf("a 29-sample bin must be held without a rate:\n%s", report)
	}
	if !strings.Contains(report, "권고: **보류**") {
		t.Fatalf("the p ≥ 0.90 verdict must be held below 30 labels:\n%s", report)
	}
	if strings.Contains(report, "0.690") {
		t.Fatalf("a held bin leaked its hit rate:\n%s", report)
	}
}

func TestReportJudgesBinsAtThirtySamples(t *testing.T) {
	t.Parallel()
	contents := syntheticLedger("j1_up_within_horizon", rowsAt(0.95, 27, 3)) // 30건, 0.900
	report := renderCalibrationReport(contents, "ledger.jsonl")
	if !strings.Contains(report, "| 0.95–1.00 | 30 | 30 | 27 | 0.900 | 판정 가능 |") {
		t.Fatalf("a 30-sample bin must be judged:\n%s", report)
	}
	if !strings.Contains(report, "예비 채택 후보") {
		t.Fatalf("0.900 at p ≥ 0.90 is a preliminary adopt candidate:\n%s", report)
	}
	low := renderCalibrationReport(syntheticLedger("j1_up_within_horizon", rowsAt(0.97, 20, 10)), "l")
	if !strings.Contains(low, "예비 기각 후보") {
		t.Fatalf("0.667 at p ≥ 0.90 is a preliminary reject candidate:\n%s", low)
	}
}

func TestReportUsesTheQuestionsHitRuleAndExcludesGaps(t *testing.T) {
	t.Parallel()
	rows := rowsAt(0.91, 0, 30) // 수익률 전부 음수
	rows = append(rows, struct {
		p   float64
		ret *float64
	}{0.91, nil})
	j2 := renderCalibrationReport(syntheticLedger("j2_hold_off_red_flag", rows), "l")
	if !strings.Contains(j2, "| 0.90–0.95 | 31 | 30 | 30 | 1.000 | 판정 가능 |") {
		t.Fatalf("J2 counts negative returns as hits and leaves the label gap out of the denominator:\n%s", j2)
	}
	if !strings.Contains(j2, "라벨 결손 1") {
		t.Fatalf("the label gap must be counted in the header:\n%s", j2)
	}
}

func TestReportStatesTheTwoDayLimitAndBinEdges(t *testing.T) {
	t.Parallel()
	report := renderCalibrationReport(syntheticLedger("j1_up_within_horizon", rowsAt(0.5, 1, 0)), "l")
	for _, want := range []string{"2일 측정", "예비 판정", "| < 0.50 |", "| 0.50–0.55 | 1 |", "| 0.95–1.00 |"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report lacks %q:\n%s", want, report)
		}
	}
	for p, want := range map[float64]int{0.4999: -1, 0.5: 0, 0.55: 1, 0.8999: 7, 0.9: 8, 0.95: 9, 1.0: 9} {
		if got := binIndex(p); got != want {
			t.Errorf("binIndex(%v) = %d, want %d", p, got, want)
		}
	}
}

// TestReportCommandReadsARealLedger 는 수집 → 라벨 → report 하위 명령을 한 줄로 이어 봄.
func TestReportCommandReadsARealLedger(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	rig.clock.Set(at(11, 0))
	if err := rig.probe.labelTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	path := filepath.Join(rig.dir, ledgerFileName)
	if err := dispatch(context.Background(), []string{"report", "--ledger", path}, os.Getenv, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"judgment 행 4 · 실현 라벨 4", "## j1_up_within_horizon — KR", "## j2_hold_off_red_flag — KR"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("report lacks %q:\n%s", want, out.String())
		}
	}
}
