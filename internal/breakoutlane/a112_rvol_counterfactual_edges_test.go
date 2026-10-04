package breakoutlane

// a112 8.5 응답 로트(Manager 최종 판정 2026-10-04) — 2.3 의 1.2 반사실(B7) 시험이 못 보던 축 셋.
//   - ⑨ 보이스 3 P2-3: 커밋된 반사실 픽스처는 ATR=10 이라 close buffer 가 1 틱으로 접혀 「buffer 를 넘은 종가」와 「저항선을 넘은 종가」가 같은
//     입력에서 갈리지 않았고(변이 A4 생존), 다섯 경우 전부 첫 post-range 봉만 바꿔 「첫 봉만 본다」 변이(A5)도 살았다.
//   - ⑩ codex r2 P2 · 보이스 2 P2-3 · 보이스 3 P2-4: 골든 `thresholds.rvol_counterfactual_ppm[0]` 은 어떤 시험도 코드와 대조하지 않았다. 결속은
//     B7 자리만 한다 — 입장 경로의 `>= 1_200_000` 은 입장 봉이 RVOL ≥ 1.5M 을 함의해 1.5M 이하 어떤 값이든 동등이라 지킬 행동이 없다(FLM 기록).

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// a112CounterfactualDecision 은 fixtureInput 을 바꿔(입력 전체 · 봉 하나) 평가한다.
func a112CounterfactualDecision(t *testing.T, input func(*EvidenceInput), index int, set func(*ClosedBarInput)) Decision {
	t.Helper()
	i := fixtureInput(t)
	if input != nil {
		input(&i)
	}
	i.Bars = append([]ClosedBar(nil), i.Bars...)
	b := i.Bars[index].value
	set(&b)
	i.Bars[index] = ClosedBar{value: b}
	return Evaluate(snapshot(t, i), nil)
}

// ⑨-1: buffer 가 한 틱보다 넓을 때 반사실은 저항선이 아니라 **buffer 위** 종가만 센다.
func TestTheOnePointTwoCounterfactualCountsOnlyACloseBeyondTheATRBuffer(t *testing.T) {
	const resistance, atr = 100, 50 // 저항선 = opening range 고가(fixture 첫 15 봉 고가 100)
	config := fixtureInput(t).Config
	buffer := breakoutBufferMinor(config.value.TickMinor, atr, config.value.BreakoutBufferPPM)
	// 배치: buffer 가 한 틱보다 넓어야 두 식(「buffer 위」 · 「저항선 위」)이 갈린다 — 접히면 이 시험은 아무것도 재지 않는다.
	if buffer <= config.value.TickMinor || BreakoutCloseQualifies(resistance+buffer-1, resistance, atr, config) || !BreakoutCloseQualifies(resistance+buffer, resistance, atr, config) {
		t.Fatalf("arrangement: buffer=%d (tick %d) does not separate the two closes", buffer, config.value.TickMinor)
	}
	for _, tc := range []struct {
		name  string
		close uint64
		want  bool
	}{
		{"one minor unit inside the buffer (above resistance)", resistance + buffer - 1, false},
		{"exactly at the buffer", resistance + buffer, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := a112CounterfactualDecision(t, func(i *EvidenceInput) { i.ATRMinor = atr }, 15, func(b *ClosedBarInput) {
				b.HighMinor, b.CloseMinor, b.RVOLPPM, b.UpperWickRangePPM = resistance+buffer+5, tc.close, 1_300_000, 100_000
			})
			if p := d.Provenance(); p.RVOLAt1200000 != tc.want || p.RVOLAdmission {
				t.Fatalf("close=%d buffer=%d: provenance=%+v, want at1.2=%v and no admission", tc.close, buffer, p, tc.want)
			}
		})
	}
}

// ⑨-2: 반사실은 첫 post-range 봉만이 아니라 입장 못 한 **모든** post-range 봉을 본다.
func TestTheOnePointTwoCounterfactualSeesALaterBarNotOnlyTheFirst(t *testing.T) {
	// 첫 post-range 봉(색인 15)은 돌파 아님(저항선 아래 종가 · RVOL 1.0), 둘째(색인 16)가 1.2 만 통과하는 돌파 종가.
	quiet := func(i *EvidenceInput) {
		b := i.Bars[15].value
		b.HighMinor, b.LowMinor, b.CloseMinor, b.RVOLPPM, b.UpperWickRangePPM = 99, 90, 95, 1_000_000, 100_000
		i.Bars = append([]ClosedBar(nil), i.Bars...)
		i.Bars[15] = ClosedBar{value: b}
	}
	for _, tc := range []struct {
		name string
		rvol uint64
		want bool
	}{
		{"second bar RVOL 1.3", 1_300_000, true},
		{"second bar RVOL 1.0 (control)", 1_000_000, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := a112CounterfactualDecision(t, quiet, 16, func(b *ClosedBarInput) {
				b.HighMinor, b.CloseMinor, b.RVOLPPM, b.UpperWickRangePPM = 102, 101, tc.rvol, 100_000
			})
			if p := d.Provenance(); p.RVOLAt1200000 != tc.want || p.RVOLAdmission {
				t.Fatalf("provenance=%+v, want at1.2=%v and no admission", p, tc.want)
			}
		})
	}
}

// ⑩: B7 의 문턱이 골든 `thresholds.rvol_counterfactual_ppm[0]` 과 행동으로 같다 — 그 값에서 참, 1 ppm 아래에서 거짓.
func TestTheOnePointTwoCounterfactualThresholdIsTheGoldensFirstEntry(t *testing.T) {
	path := filepath.Join(repoRoot(t), "openspec/changes/a112-run-four-strategy-families-independently/analysis/goldens/breakout-evidence-and-sizing-v1.json")
	var golden struct {
		Thresholds struct {
			Counterfactual []uint64 `json:"rvol_counterfactual_ppm"`
			Admission      uint64   `json:"rvol_min_ppm"`
		} `json:"thresholds"`
	}
	if err := json.Unmarshal(mustRead(t, path), &golden); err != nil {
		t.Fatal(err)
	}
	if len(golden.Thresholds.Counterfactual) == 0 || golden.Thresholds.Counterfactual[0] == 0 || golden.Thresholds.Counterfactual[0] >= golden.Thresholds.Admission {
		t.Fatalf("golden counterfactual %v is not a threshold under admission %d — this test would measure nothing", golden.Thresholds.Counterfactual, golden.Thresholds.Admission)
	}
	threshold := golden.Thresholds.Counterfactual[0]
	for rvol, want := range map[uint64]bool{threshold: true, threshold - 1: false} {
		d := a112CounterfactualDecision(t, nil, 15, func(b *ClosedBarInput) { b.RVOLPPM = rvol })
		if p := d.Provenance(); p.RVOLAt1200000 != want || p.RVOLAdmission {
			t.Fatalf("RVOL %d (golden %d): provenance=%+v, want at1.2=%v and no admission", rvol, threshold, p, want)
		}
	}
}
