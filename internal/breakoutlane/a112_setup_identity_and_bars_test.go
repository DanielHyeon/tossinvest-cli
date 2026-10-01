package breakoutlane

// a112 breakout 덮개 2차 — 태스크 2.2 · 2.2.1 · 2.3 의 빈칸(Manager 승인 2026-10-01).
//   2.2   중복 · 순서 바뀐 봉(같은 ID · 같은 순번 · 자리 바꿈)은 스냅숏 봉인 전에 거절된다.
//   2.2.1 setup ID 는 같은 세션의 정정(종단 전 · PROPOSED 뒤 · 종단 뒤) 내내 같고, 세션이 바뀌면 다르며 이전 종단을 상속하지 않는다.
//   2.3   돌파 봉 RVOL 의 입장 경계(1.5)와 반사실 2.0 · 2.5 의 경계(한 ppm 아래 · 정확히).
// 반사실 1.2 는 여기서 단언하지 않는다 — 입장(1.5) 아래 봉에는 반사실이 기록되지 않고 입장한 봉에서는 1.2 가 항상 참이라, 스펙의
// 「1.2 반사실 결과」 를 어떻게 읽을지가 열려 있다(Manager 보고 — 임의 해석 금지).

import "testing"

func TestDuplicateOrReorderedBarsAreRefusedBeforeTheSnapshotSeals(t *testing.T) {
	for name, mutate := range map[string]func([]ClosedBar){
		"same id on two bars": func(b []ClosedBar) {
			v := b[16].value
			v.ID = b[15].value.ID
			b[16] = ClosedBar{value: v}
		},
		"same sequence twice": func(b []ClosedBar) {
			v := b[16].value
			v.Sequence = b[15].value.Sequence
			b[16] = ClosedBar{value: v}
		},
		"two bars swapped":           func(b []ClosedBar) { b[15], b[16] = b[16], b[15] },
		"opening-range bars swapped": func(b []ClosedBar) { b[0], b[1] = b[1], b[0] },
	} {
		i := fixtureInput(t)
		i.Bars = append([]ClosedBar(nil), i.Bars...)
		mutate(i.Bars)
		if _, err := NewEvidenceSnapshot(i); err == nil {
			t.Errorf("%s: the snapshot sealed", name)
		}
	}
}

func TestTheSetupIDIsStableAcrossSameSessionCorrectionsAndChangesWithTheSession(t *testing.T) {
	base := fixtureInput(t)
	proposed := Evaluate(snapshot(t, base), nil)
	setup := proposed.SetupID()
	if proposed.Phase() != "PROPOSED" || setup == "" {
		t.Fatalf("arrangement: %+v", proposed)
	}
	corrected := func(in EvidenceInput, n int, set func(*ClosedBarInput)) EvidenceInput {
		in.Bars = append([]ClosedBar(nil), in.Bars...)
		b := in.Bars[n].value
		b.Revision++
		set(&b)
		in.Bars[n] = ClosedBar{value: b}
		return in
	}
	// 종단 전 정정: RETEST_WAIT 의 돌파 봉을 정정 → 새 스냅숏 · 같은 setup.
	wait := base
	wait.Bars = base.Bars[:16]
	before := Evaluate(snapshot(t, wait), nil)
	after := Evaluate(snapshot(t, corrected(wait, 15, func(b *ClosedBarInput) { b.CloseMinor = 100 })), &before)
	// PROPOSED 뒤 정정 · 종단 뒤 정정.
	held := Evaluate(snapshot(t, corrected(base, 17, func(b *ClosedBarInput) {})), &proposed)
	failed := corrected(base, 17, func(b *ClosedBarInput) { b.CloseMinor, b.HighMinor, b.VolumeExpanded = 99, 100, true })
	terminal := Evaluate(snapshot(t, failed), nil)
	afterTerminal := Evaluate(snapshot(t, corrected(failed, 16, func(b *ClosedBarInput) {})), &terminal)
	for name, d := range map[string]Decision{"pre-terminal": before, "pre-terminal corrected": after, "after PROPOSED": held, "terminal": terminal, "after terminal": afterTerminal} {
		if d.SetupID() != setup {
			t.Errorf("%s (%s): setup %s, want the session's %s", name, d.Phase(), d.SetupID(), setup)
		}
	}
	if after.SnapshotDigest() == before.SnapshotDigest() || after.Phase() != "RANGE_LOCKED" {
		t.Errorf("pre-terminal correction did not replay on a new snapshot: %s %s", after.Phase(), after.SnapshotDigest())
	}
	if afterTerminal.Phase() != "INVALIDATED" || afterTerminal.ProposalID() != "" {
		t.Errorf("after-terminal correction resurrected: %+v", afterTerminal)
	}
	// 세션 교체: 다음 정규 세션의 같은 종목 · 같은 모양 → 다른 setup, 이전 세션의 INVALIDATED prior 를 상속하지 않는다.
	next := base
	next.SessionID = "KRX:2026-08-19"
	next.Bars = append([]ClosedBar(nil), base.Bars...)
	for n := range next.Bars {
		b := next.Bars[n].value
		b.SessionID, b.ID = next.SessionID, "next-"+b.ID
		next.Bars[n] = ClosedBar{value: b}
	}
	rolled := Evaluate(snapshot(t, next), &terminal)
	if rolled.SetupID() == setup || rolled.Phase() != "PROPOSED" || rolled.Diagnostic() != "" {
		t.Errorf("session rollover: setup equal=%v phase=%s diagnostic=%q, want a new setup evaluated fresh", rolled.SetupID() == setup, rolled.Phase(), rolled.Diagnostic())
	}
	// 세션만 다르고 봉 ID 가 같아도(ID 에 날짜가 없는 원천) setup 은 다르다 — 세션 자체가 setup 신원의 일부다.
	sameIDs := next
	sameIDs.Bars = append([]ClosedBar(nil), base.Bars...)
	for n := range sameIDs.Bars {
		b := sameIDs.Bars[n].value
		b.SessionID = next.SessionID
		sameIDs.Bars[n] = ClosedBar{value: b}
	}
	if d := Evaluate(snapshot(t, sameIDs), &terminal); d.SetupID() == setup || d.Phase() != "PROPOSED" {
		t.Errorf("session rollover with the same bar ids: setup equal=%v phase=%s, want a new setup", d.SetupID() == setup, d.Phase())
	}
}

func TestRVOLAdmissionAndCounterfactualBoundaries(t *testing.T) {
	at := func(rvol uint64) Decision {
		i := fixtureInput(t)
		i.Bars = append([]ClosedBar(nil), i.Bars...)
		b := i.Bars[15].value
		b.RVOLPPM = rvol
		i.Bars[15] = ClosedBar{value: b}
		return Evaluate(snapshot(t, i), nil)
	}
	if d := at(1_499_999); d.Phase() != "RANGE_LOCKED" || d.Provenance().RVOLAdmission {
		t.Errorf("RVOL one ppm under admission: %s admission=%v, want no breakout", d.Phase(), d.Provenance().RVOLAdmission)
	}
	for _, tc := range []struct {
		rvol           uint64
		at2000, at2500 bool
	}{
		{1_500_000, false, false}, {1_999_999, false, false}, {2_000_000, true, false}, {2_499_999, true, false}, {2_500_000, true, true},
	} {
		p := at(tc.rvol).Provenance()
		if !p.RVOLAdmission || p.RVOLAt2000000 != tc.at2000 || p.RVOLAt2500000 != tc.at2500 {
			t.Errorf("RVOL %d: admission=%v at2.0=%v at2.5=%v, want true %v %v", tc.rvol, p.RVOLAdmission, p.RVOLAt2000000, p.RVOLAt2500000, tc.at2000, tc.at2500)
		}
	}
}

// CONSUMED 는 v1 평가기에 생산자가 없다(B1 — 6.4 의 몫). 6.4 가 그것을 만들면 평가기는 정정 뒤에도 그대로 보존해야 한다 — 그 계약을
// 패키지 안에서 봉인한 CONSUMED prior 로 미리 잰다(생산자가 아니라 소비 쪽 계약). 대조: prior 없이는 같은 스냅숏이 새 제안을 낸다.
func TestACorrectionAfterConsumedNeitherResurrectsNorReissues(t *testing.T) {
	proposed := Evaluate(snapshot(t, fixtureInput(t)), nil)
	consumed := proposed
	consumed.phase = phaseConsumed
	consumed.seal = decisionSeal(consumed)
	next := fixtureInput(t)
	next.Bars = append([]ClosedBar(nil), next.Bars...)
	b := next.Bars[17].value
	b.Revision = 2
	next.Bars[17] = ClosedBar{value: b}
	next.Sizing.StopMinor = 50
	if control := Evaluate(snapshot(t, next), nil); control.Phase() != "PROPOSED" || control.ProposalID() == proposed.ProposalID() {
		t.Fatalf("control: %s — without the prior the corrected snapshot must issue a new proposal", control.Phase())
	}
	got := Evaluate(snapshot(t, next), &consumed)
	if got.Phase() != "CONSUMED" || got.ProposalID() != proposed.ProposalID() || got.SetupID() != proposed.SetupID() ||
		got.SnapshotDigest() != proposed.SnapshotDigest() || got.Diagnostic() != DiagnosticCorrectionAfterProposal {
		t.Fatalf("after CONSUMED: %+v, want the consumed decision preserved with the correction diagnostic only", got)
	}
}
