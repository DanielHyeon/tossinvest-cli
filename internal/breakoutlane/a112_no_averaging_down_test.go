package breakoutlane

// a112 breakout 덮개 2차(Manager 승인 2026-10-01 — 우선순위 높음): 정본 손절 불변식의 breakout 적용.
// spec「breakout invalidation과 first-touch 금지는 fail-closed다」: "Failed reclaim 이후 더 낮은 가격으로 평균단가를 낮추거나 stop을
// entry에서 멀어지게 하는 proposal을 만들면 안 되며 (MUST NOT)". 레인이 그 둘을 만들 수 있는 길은 셋뿐이다:
//   (1) 실패(INVALIDATED · TIMED_OUT) 뒤 같은 setup 에서 다시 제안 — 더 낮은 재진입(물타기 다리).
//   (2) PROPOSED 뒤 같은 setup 에서 다른 사이징(더 낮은 entry · 더 먼 stop)으로 다시 제안 — 둘째 다리 또는 손절 후퇴.
//   (3) 보호가 아닌 stop(0 · entry 이상)으로 처음부터 제안.
// 각 길에 대조군을 둔다: prior 없이 같은 스냅숏을 평가하면 **실제로 제안이 나오는** 경우에만 금지가 prior 덕분임을 보인다
// (대조군 없는 「제안 0」은 픽스처가 원래 제안을 못 만들어서일 수 있다).
// 범위: 이 금지는 호출자가 prior 를 넘길 때의 레인 판정이다. 생산에서 BreakoutRequest.Prior 를 채우는 생산자는 현재 0 이다(측정:
// 시험 밖 `BreakoutRequest{` 생성 0 곳) — prior 를 넘기는 배선은 6.4 이후 생산자의 몫이고, 진입 뒤 stop 관리는 레인이 아니라 보호 경로다.

import "testing"

// a112ExtendedBars 는 픽스처 18 봉 뒤에 다시 retest(99) · reclaim(101) 두 봉을 잇는다 — 실패 뒤 「더 낮은 재진입」 모양.
func a112ExtendedBars(t *testing.T, bars []ClosedBar) []ClosedBar {
	t.Helper()
	out := append([]ClosedBar(nil), bars...)
	return append(out, fixtureBar(t, 19, 100, 98, 99, 1_000_000, 100_000), fixtureBar(t, 20, 102, 99, 101, 1_000_000, 100_000))
}

// a112LowerLeg 는 더 낮은 entry 와 entry 에서 더 먼 stop 의 사이징이다(픽스처 entry 101 · stop 95 대비).
func a112LowerLeg(s SizingInput) SizingInput {
	s.ProposedEntryMinor, s.StopMinor = 95, 50
	return s
}

func TestNoAveragingDownLegAfterAFailedSetup(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bar18 func(ClosedBarInput) ClosedBarInput
		want  string
	}{
		{"volume-expanded failed reclaim", func(b ClosedBarInput) ClosedBarInput {
			b.CloseMinor, b.HighMinor, b.VolumeExpanded = 99, 100, true
			return b
		}, "INVALIDATED"},
		{"close below the range low", func(b ClosedBarInput) ClosedBarInput {
			b.CloseMinor, b.HighMinor, b.LowMinor = 89, 100, 85
			return b
		}, "INVALIDATED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			failed := fixtureInput(t)
			failed.Bars[17] = ClosedBar{value: tc.bar18(failed.Bars[17].value)}
			prior := Evaluate(snapshot(t, failed), nil)
			if prior.Phase() != tc.want || prior.ProposalID() != "" {
				t.Fatalf("arrangement: the failed setup is %s proposal=%q, want %s with no proposal", prior.Phase(), prior.ProposalID(), tc.want)
			}
			// 실패 뒤 다시 retest · reclaim 이 이어지고 호출자는 더 낮은 entry · 더 먼 stop 을 들고 온다.
			later := failed
			later.Bars = a112ExtendedBars(t, failed.Bars)
			later.Sizing = a112LowerLeg(failed.Sizing)
			// prior 없이: 걷기는 첫 종단에서 돌아온다 — 뒤의 reclaim 은 보이지 않는다.
			if fresh := Evaluate(snapshot(t, later), nil); fresh.Phase() != tc.want || fresh.ProposalID() != "" || fresh.FinalQuantity() != 0 || fresh.CandidateQuantity() != 0 {
				t.Errorf("fresh walk past the failure: phase=%s proposal=%q final=%d candidate=%d, want %s and no leg",
					fresh.Phase(), fresh.ProposalID(), fresh.FinalQuantity(), fresh.CandidateQuantity(), tc.want)
			}
			// prior 와 함께: 종단은 보존된다 — 같은 스냅숏 digest · 진단, 제안 없음.
			held := Evaluate(snapshot(t, later), &prior)
			if held.Phase() != tc.want || held.ProposalID() != "" || held.FinalQuantity() != 0 || held.SnapshotDigest() != prior.SnapshotDigest() ||
				held.Diagnostic() != DiagnosticCorrectionAfterProposal {
				t.Errorf("with the failed prior: %+v, want the %s prior preserved with the correction diagnostic and no leg", held, tc.want)
			}
		})
	}
}

// 정정이 실패한 봉을 성공 reclaim 으로 바꿔도 종단은 되살아나지 않는다 — 대조군(prior 없음)은 같은 스냅숏에서 실제로 제안한다.
func TestACorrectionCannotResurrectAFailedSetupIntoALowerLeg(t *testing.T) {
	failed := fixtureInput(t)
	bar := failed.Bars[17].value
	bar.CloseMinor, bar.HighMinor, bar.VolumeExpanded = 99, 100, true
	failed.Bars[17] = ClosedBar{value: bar}
	prior := Evaluate(snapshot(t, failed), nil)
	if prior.Phase() != "INVALIDATED" {
		t.Fatalf("arrangement: %s", prior.Phase())
	}
	corrected := failed
	corrected.Bars = append([]ClosedBar(nil), failed.Bars...)
	bar = fixtureInput(t).Bars[17].value
	bar.Revision = 2
	corrected.Bars[17] = ClosedBar{value: bar}
	corrected.Sizing = a112LowerLeg(failed.Sizing)
	if control := Evaluate(snapshot(t, corrected), nil); control.Phase() != "PROPOSED" || control.FinalQuantity() == 0 {
		t.Fatalf("control: without the prior the corrected snapshot is %s final=%d — it must propose, or this test proves nothing", control.Phase(), control.FinalQuantity())
	}
	held := Evaluate(snapshot(t, corrected), &prior)
	if held.Phase() != "INVALIDATED" || held.ProposalID() != "" || held.FinalQuantity() != 0 || held.Diagnostic() != DiagnosticCorrectionAfterProposal {
		t.Fatalf("a correction resurrected the failed setup: %+v", held)
	}
}

// PROPOSED 뒤 같은 setup 에서 더 낮은 entry · 더 먼 stop 의 둘째 사이징은 새 제안이 되지 않는다.
func TestAProposedSetupNeverReSizesIntoALowerLegOrARetreatedStop(t *testing.T) {
	proposed := Evaluate(snapshot(t, fixtureInput(t)), nil)
	if proposed.Phase() != "PROPOSED" || proposed.ProposalID() == "" {
		t.Fatalf("arrangement: %s", proposed.Phase())
	}
	for _, tc := range []struct {
		name   string
		mutate func(*EvidenceInput)
		refuse bool // 증거 계보가 그대로면(봉 추가 · 정정 없음) 정정이 아니다 — 거절
	}{
		{"stop retreated on the same bars", func(i *EvidenceInput) { i.Sizing.StopMinor = 50 }, true},
		{"lower entry and retreated stop on the same bars", func(i *EvidenceInput) { i.Sizing = a112LowerLeg(i.Sizing) }, true},
		{"lower leg after one more bar", func(i *EvidenceInput) {
			i.Bars = append(append([]ClosedBar(nil), i.Bars...), fixtureBar(t, 19, 100, 98, 99, 1_000_000, 100_000))
			i.Sizing = a112LowerLeg(i.Sizing)
		}, false},
		{"retreated stop riding a bar correction", func(i *EvidenceInput) {
			i.Bars = append([]ClosedBar(nil), i.Bars...)
			b := i.Bars[17].value
			b.Revision = 2
			i.Bars[17] = ClosedBar{value: b}
			i.Sizing.StopMinor = 50
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next := fixtureInput(t)
			tc.mutate(&next)
			// 대조군: prior 없이 이 스냅숏은 다른 수량의 새 제안을 낸다 — 금지가 없다면 실제로 둘째 다리가 생긴다.
			control := Evaluate(snapshot(t, next), nil)
			if control.Phase() != "PROPOSED" || control.ProposalID() == proposed.ProposalID() || control.CandidateQuantity() == proposed.CandidateQuantity() {
				t.Fatalf("control: %s proposal-equal=%v candidate %d vs %d — the snapshot must re-size without the prior",
					control.Phase(), control.ProposalID() == proposed.ProposalID(), control.CandidateQuantity(), proposed.CandidateQuantity())
			}
			got := Evaluate(snapshot(t, next), &proposed)
			if tc.refuse {
				if got.Refusal() != RefusalEvidenceInvalid || got.ProposalID() != "" || got.FinalQuantity() != 0 || got.Phase() == "PROPOSED" {
					t.Fatalf("a re-sizing without new evidence: %+v, want EVIDENCE_INVALID with no proposal", got)
				}
				return
			}
			if got.Phase() != "PROPOSED" || got.ProposalID() != proposed.ProposalID() || got.CandidateQuantity() != proposed.CandidateQuantity() ||
				got.FinalQuantity() != proposed.FinalQuantity() || got.SnapshotDigest() != proposed.SnapshotDigest() || got.Diagnostic() != DiagnosticCorrectionAfterProposal {
				t.Fatalf("the proposal moved: %+v, want the original proposal (candidate %d, final %d) preserved with the correction diagnostic",
					got, proposed.CandidateQuantity(), proposed.FinalQuantity())
			}
		})
	}
}

// 보호가 아닌 stop 은 처음부터 제안이 되지 않는다. 경계 대조: entry 바로 아래 1 minor 의 stop 은 제안된다.
func TestANonProtectiveStopNeverProposes(t *testing.T) {
	for _, tc := range []struct {
		name string
		stop uint64
		want RefusalCode
	}{
		{"stop zero", 0, RefusalNonProtectiveStop},
		{"stop equals entry", 101, RefusalNonProtectiveStop},
		{"stop above entry", 102, RefusalNonProtectiveStop},
		{"stop one minor below entry", 100, RefusalNone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := fixtureInput(t)
			i.Sizing.StopMinor = tc.stop
			d := Evaluate(snapshot(t, i), nil)
			if tc.want == RefusalNone {
				if d.Phase() != "PROPOSED" || d.FinalQuantity() == 0 {
					t.Fatalf("boundary control: %+v, want a proposal", d)
				}
				return
			}
			if d.Phase() != "ARMED" || d.Refusal() != tc.want || d.ProposalID() != "" || d.FinalQuantity() != 0 || d.CandidateQuantity() != 0 {
				t.Fatalf("%+v, want ARMED with %s and no quantity", d, tc.want)
			}
		})
	}
	// entry 0 · ask 0 은 봉인 · quote 검사가 먼저 막는다 — 사이징 가드는 그 뒤의 이중 방어다. 각 층을 따로 잰다.
	i := fixtureInput(t)
	i.Sizing.ProposedEntryMinor = 0
	if d := Evaluate(snapshot(t, i), nil); d.Refusal() != RefusalEvidenceInvalid || d.ProposalID() != "" {
		t.Errorf("entry zero through Evaluate: %+v, want EVIDENCE_INVALID from the quote check", d)
	}
	zeroAsk := QuoteSealInput{BidMinor: 100, AskMinor: 0, LastMinor: 101, Currency: "USD", SourceObservedAtMS: 5, ReceivedAtMS: 6}
	zeroAsk.Digest = QuoteSealDigest(zeroAsk)
	if _, err := NewQuoteSeal(zeroAsk); err == nil {
		t.Error("a zero-ask quote was sealed")
	}
	base := fixtureInput(t).Sizing
	for name, c := range map[string]struct {
		in SizingInput
		q  QuoteSeal
	}{
		"entry zero":          {func() SizingInput { s := base; s.ProposedEntryMinor = 0; return s }(), fixtureQuote(t, 5, 6)},
		"ask zero (unsealed)": {base, QuoteSeal{value: zeroAsk}},
		"stop zero":           {func() SizingInput { s := base; s.StopMinor = 0; return s }(), fixtureQuote(t, 5, 6)},
		"stop equals entry":   {func() SizingInput { s := base; s.StopMinor = s.ProposedEntryMinor; return s }(), fixtureQuote(t, 5, 6)},
	} {
		if r := size(c.in, c.q, fixtureFX(t, 10), 10); r.Refusal != RefusalNonProtectiveStop || r.FinalQuantity != 0 || r.CandidateQuantity != 0 {
			t.Errorf("size %s: %+v, want NON_PROTECTIVE_STOP with no quantity", name, r)
		}
	}
}
