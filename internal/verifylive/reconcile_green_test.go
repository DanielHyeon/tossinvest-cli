package verifylive

// reconcile_green_test.go — a121 GREEN 로트가 변이 원장에서 찾은 빈칸을 닫는 시험(RED 시험은 고치지 않음).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"
)

// TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt — CLOSED 행의 OPEN 계열 status 는 그룹과
// 모순이므로, 허용 목록이 (잘못 전사해) 그 값을 담아도 거절한다(codex F4). 이 가드가 허용 목록 판정에 가려지지 않게
// 허용 목록에 WATCHING 을 주입해 잰다(변이 V-status-openfamily).
func TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt(t *testing.T) {
	prev := reconcileClosedStatusAllowlist
	reconcileClosedStatusAllowlist = map[string]bool{"EXPIRED": true, "WATCHING": true}
	t.Cleanup(func() { reconcileClosedStatusAllowlist = prev })

	h := newRCHarness(t, rcA063Entries())
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcCondRow("CO-OTHER-OPENFAMILY", "WATCHING")))
	rcRefuse(t, h, RefuseClosedStatus)
}

// TestReconciledProjectionIsMonotoneLikeOutstanding — 대사 줄 뒤에 같은 artifact 를 종결 없이 다시 적은 줄이 와도
// (구 바이너리 재개 등) 종결은 단조라 Outstanding 은 되살리지 않는다. 투영 ReconciledArtifacts 도 같은 규칙이어야
// 그 artifact 가 화면에서 사라지지 않는다(어느 쪽에도 안 보이는 구멍 — 변이 V-projection-monotone).
func TestReconciledProjectionIsMonotoneLikeOutstanding(t *testing.T) {
	path := t.TempDir() + "/" + RecordFileName(MarketKR)
	rcWriteEntries(t, path, rcA063Entries())
	rcAppendRaw(t, path, rcReconcileLineJSON(rcTarget(), rcNow, ReconcileBasisDomain+":sha256:00"))
	rcWriteEntries(t, path, []Entry{{Kind: KindStep, StepID: StepConditionalPersist, Verdict: VerdictPass,
		AccountRef: rcMask(), Artifacts: []Artifact{rcTarget()}}})
	entries, err := LoadEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	if rcHas(Outstanding(entries), rcTargetID) {
		t.Fatal("a later non-terminal mention resurrected a reconciled artifact")
	}
	if !rcHas(ReconciledArtifacts(entries), rcTargetID) {
		t.Fatal("the reconciled artifact vanished from the reconciled projection after a later non-terminal mention")
	}
}

// TestReconcileRechecksTheWindowAtBothContractPoints 는 변이 원장 생존 V-final-window·V-window-before-instr2 의 대체 가드
// 핀이다(Manager 판정 2026-10-05 — 동등 수용, 심층 방어 유지). 두 재검 사이에는 로컬 기록 읽기뿐이라 행동 시험은
// 서로를 가린다. 그래서 Reconcile 본문에서 재검 헬퍼 checkReconcileWindow 호출이 정확히 둘이고, 하나는 둘째 종목 조회
// 뒤·multiset 비교 앞(가드 순서 계약), 다른 하나는 추가 직전 기록 재읽기 뒤·추가 앞(codex F7)임을 AST 로 센다.
func TestReconcileRechecksTheWindowAtBothContractPoints(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reconcile.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, d := range file.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "Reconcile" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("reconcile.go declares no Reconcile — the pin has no subject")
	}
	var calls []string // 소스 순서의 호출 이름
	ast.Inspect(body, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			switch f := c.Fun.(type) {
			case *ast.Ident:
				calls = append(calls, f.Name)
			case *ast.SelectorExpr:
				calls = append(calls, f.Sel.Name)
			}
		}
		return true
	})
	positions := func(name string) []int {
		var out []int
		for i, c := range calls {
			if c == name {
				out = append(out, i)
			}
		}
		return out
	}
	window := positions("checkReconcileWindow")
	if len(window) != 2 {
		t.Fatalf("Reconcile calls checkReconcileWindow %d time(s), want exactly 2 (after the second instrument read, and immediately before the append)", len(window))
	}
	instrument := positions("checkReconcileInstrument")
	multiset := positions("sameReconcileMultiset")
	reread := positions("readRecordStrictNoTail")
	appendLine := positions("appendReconcileLine")
	if len(instrument) != 2 || len(multiset) != 1 || len(reread) != 1 || len(appendLine) != 1 {
		t.Fatalf("Reconcile's landmarks changed: instrument %v multiset %v reread %v append %v", instrument, multiset, reread, appendLine)
	}
	if !(instrument[1] < window[0] && window[0] < multiset[0]) {
		t.Fatalf("the first recheck must sit after the second instrument read and before the multiset comparison: calls %v", calls)
	}
	if !(reread[0] < window[1] && window[1] < appendLine[0]) {
		t.Fatalf("the second recheck must sit after the record re-read and before the append: calls %v", calls)
	}
}

// TestReconcileStaticRefusalsComeBeforeTheHumanApproval 는 A-GREEN P2-a 의 순서 핀이다 — Q1(측정 부재·축출 모형·영
// CreatedAt·나이 초과)과 Q3(미확정)는 로컬 판정이라 사람 승인 seam 이 불리기 **전에** 거절해야 한다. 운영자가 y 를
// 누른 뒤에 retention-unmeasured 로 거절당하는 모양을 막는다(리뷰어 변이 GO2).
func TestReconcileStaticRefusalsComeBeforeTheHumanApproval(t *testing.T) {
	cases := map[string]struct {
		retention *retentionMeasurement
		freshness time.Duration
		mutate    func([]Entry)
		want      ReconcileRefusalCode
	}{
		"q1-unmeasured": {nil, rcFreshness, nil, RefuseRetentionUnmeasured},
		"q1-eviction":   {&retentionMeasurement{Bound: rcRetentionBound, Eviction: evictionCount}, rcFreshness, nil, RefuseRetentionEviction},
		"q1-created-zero": {&retentionMeasurement{Bound: rcRetentionBound, Eviction: evictionDuration}, rcFreshness,
			func(e []Entry) { e[1].Artifacts[0].CreatedAt = time.Time{} }, RefuseCreatedAtZero},
		"q1-too-old": {&retentionMeasurement{Bound: time.Hour, Eviction: evictionDuration}, rcFreshness, nil, RefuseArtifactTooOld},
		"q3-unfixed": {&retentionMeasurement{Bound: rcRetentionBound, Eviction: evictionDuration}, 0, nil, RefuseFreshnessUnfixed},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			entries := rcA063Entries()
			if c.mutate != nil {
				c.mutate(entries)
			}
			h := newRCHarness(t, entries)
			rcSetPolicies(t, c.retention, c.freshness)
			rcRefuse(t, h, c.want)
			if len(h.approvals) != 0 {
				t.Fatalf("the operator was asked to approve before a local refusal (%s)", c.want)
			}
			if h.out.Len() != 0 {
				t.Fatalf("the approval text was shown before a local refusal:\n%s", h.out.String())
			}
		})
	}
}

// TestReconcileLineCarriesNoHoldEvenWhenTheOutstandingLineHadOne — 대상 줄이 명시 보유(HeldUntil)를 실었어도(해제된
// 뒤) 대사 줄은 보유를 싣지 않는다(design P2-6 「HeldUntil 영」). 대사된 artifact 의 직렬화 확정형이 그 필드를 빼는 것이
// 유일한 보장이므로 그 자리를 잰다(A-GREEN P2-b — 중복 대입 삭제 뒤 MJ10 확인).
func TestReconcileLineCarriesNoHoldEvenWhenTheOutstandingLineHadOne(t *testing.T) {
	entries := rcA063Entries()
	entries[1].Artifacts[0].HeldUntil = StepConditionalCancel
	h := newRCHarness(t, entries)
	_, line, raw := rcAccept(t, h)
	a := line["artifacts"].([]any)[0].(map[string]any)
	if _, ok := a["held_until"]; ok {
		t.Fatalf("the reconcile line carries held_until: %s", raw)
	}
}
