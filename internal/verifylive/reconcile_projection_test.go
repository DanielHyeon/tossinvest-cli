package verifylive

// reconcile_projection_test.go — a121 tasks 2.1·2.3·2.4·2.4.1 과 로트 1 S1(표시 라벨).
//
// 투영 시험은 Reconcile 구현과 독립적으로, 설계 모양의 대사 줄(rcReconcileLineJSON)을 a063 기록 뒤에 붙여 잰다 —
// 그 줄과 실제 줄의 동형은 TestReconcileLineMatchesTheDesignShape 가 묶는다.

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

// --- 2.1 DELETE 404 · 일반 관측은 대사가 아니다 ------------------------------------

// TestADeleteNotFoundLeavesTheArtifactOutstanding 은 spec 「DELETE 404 remains nonterminal」 이다(현 동작 고정).
func TestADeleteNotFoundLeavesTheArtifactOutstanding(t *testing.T) {
	e := rcA063Entries()
	if !rcHas(Outstanding(e), rcTargetID) {
		t.Fatal("a DELETE 404 made the artifact non-outstanding")
	}
	if !rcHas(PendingCleanup(e), rcTargetID) {
		t.Fatal("a released artifact whose DELETE returned 404 must still be planned for cleanup until reconciled")
	}
}

// TestAGenericOperatorObservationDoesNotReconcile 은 tasks 2.1 의 "generic operator observation" 이다 — 운영자가 본
// "활성 주문 없음" 관측 줄도, reconciled_absent 필드 없는 reconcile 종류 줄도 artifact 를 종결하지 않는다.
func TestAGenericOperatorObservationDoesNotReconcile(t *testing.T) {
	e := rcA063Entries()
	e = append(e,
		Entry{Kind: "operator-note", AccountRef: rcMask(),
			Observations: []Observation{{Key: "operator.no_active_orders", Value: "true"}}},
		Entry{Kind: KindReconcile, StepID: StepReconcile, AccountRef: rcMask(),
			Artifacts: []Artifact{rcTarget()}},
	)
	if !rcHas(Outstanding(e), rcTargetID) {
		t.Fatal("an observation without the reconciled-absent fact ended the artifact")
	}
}

// TestADeleteNotFoundFromEveryListReadDoesNotReconcile — 목록 읽기 자체가 404 류 오류를 내면 그것은 부재가 아니라
// 읽기 실패다.
func TestADeleteNotFoundFromEveryListReadDoesNotReconcile(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	notFound := rcPage{err: rcErr("official: HTTP 404 conditional-order-not-found")}
	h.reader.set(-1, ReconcileGroupConditionalOpen, notFound)
	h.reader.set(-1, ReconcileGroupPlainOpen, notFound)
	h.reader.set(-1, ReconcileGroupConditionalClosed, notFound)
	rcRefuse(t, h, RefuseReadError)
}

// --- 2.3 재개 계획 · 실패 증거 · 판정 보존 · RedoSet 핀 ------------------------------

// TestReconcileOnlyRemovesItsExactArtifactFromCleanupPlanning — 같은 기록의 다른 남은 객체(다른 종목의 일반 주문)는
// 그대로 정리 대상이고, 실패한 cleanup 줄과 모든 단계 판정은 그대로다.
func TestReconcileOnlyRemovesItsExactArtifactFromCleanupPlanning(t *testing.T) {
	e := rcA063Entries()
	leftover := Artifact{Kind: KindOrder, ID: "ORD-LEFTOVER", Symbol: "000660", CreatedAt: rcCreated}
	e = append(e, Entry{Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictFail, AccountRef: rcMask(),
		Artifacts: []Artifact{leftover}})
	before := rcLoad(t, e)
	after := rcEntriesWithReconcileLine(t, e)

	if got := rcIDs(PendingCleanup(before)); !reflect.DeepEqual(got, []string{rcTargetID, "ORD-LEFTOVER"}) {
		t.Fatalf("fixture: cleanup before reconciliation %v", got)
	}
	if got := rcIDs(PendingCleanup(after)); !reflect.DeepEqual(got, []string{"ORD-LEFTOVER"}) {
		t.Fatalf("cleanup after reconciliation %v, want only the unrelated leftover", got)
	}
	if got := rcIDs(Outstanding(after)); !reflect.DeepEqual(got, []string{"ORD-LEFTOVER"}) {
		t.Fatalf("outstanding after reconciliation %v, want only the unrelated leftover", got)
	}
	// 원래 줄들은 바이트 단위로 그대로(append-only) — 실패한 DELETE 기록 보존.
	for i := range before {
		if !reflect.DeepEqual(rcCanon(t, before[i]), rcCanon(t, after[i])) {
			t.Fatalf("line %d changed under reconciliation", i)
		}
	}
	// 모든 단계 판정 불변.
	for _, s := range Steps() {
		b, okB := LastEntry(before, s.ID)
		a, okA := LastEntry(after, s.ID)
		if okB != okA || b.Verdict != a.Verdict || Settled(before, s.ID) != Settled(after, s.ID) ||
			Passed(before, s.ID) != Passed(after, s.ID) {
			t.Fatalf("step %s verdict changed: %v/%v → %v/%v", s.ID, okB, b.Verdict, okA, a.Verdict)
		}
	}
	if !Settled(after, StepConditionalCancel) || Passed(after, StepConditionalCancel) {
		t.Fatal("the failed conditional-cancel must stay a settled failure")
	}
}

// TestReconcilePinsRedoSetBeforeAndAfter 는 freeze P1-6 이다 — 대사 뒤 subjectLost 가 conditional-register 를 되살리는
// 현 동작을 바꾸지 않고 값으로 고정한다(설치 실행은 여전히 사람 일괄 승인 뒤).
//
// 기대값은 고정 사본 2c6ef1ef 에서 같은 기록을 Cancelled 로 종결한 등가 기록으로 RedoSet 을 계산해 얻었다(대사는 투영에서
// 셋째 종결이므로 같은 결과여야 한다).
func TestReconcilePinsRedoSetBeforeAndAfter(t *testing.T) {
	e := rcA063Entries()
	if got := RedoSet(e); !reflect.DeepEqual(got, []StepID{StepConditionalCancel}) {
		t.Fatalf("RedoSet before reconciliation %v", got)
	}
	after := rcEntriesWithReconcileLine(t, e)
	if got := RedoSet(after); !reflect.DeepEqual(got, []StepID{StepConditionalRegister, StepConditionalCancel}) {
		t.Fatalf("RedoSet after reconciliation %v, want [conditional-register conditional-cancel] (subjectLost reopens the register)", got)
	}
}

// --- 2.4 · 2.4.1 성공 endpoint · 증명 · 보고 경계 ------------------------------------

// TestReconcileLineNeverBecomesEndpointEvidence — 대사 줄은 Calls 를 싣지 않으므로 SucceededEndpoints 결과가 같다.
// 손으로 만든 줄과 실제 줄 양쪽에서 잰다(실제 줄은 Reconcile 구현이 있어야 생긴다).
func TestReconcileLineNeverBecomesEndpointEvidence(t *testing.T) {
	now := rcNow.Add(time.Minute)
	for _, window := range []time.Duration{time.Hour, 24 * time.Hour, 365 * 24 * time.Hour} {
		before := SucceededEndpoints(rcA063Entries(), now, window)
		after := SucceededEndpoints(rcEntriesWithReconcileLine(t, rcA063Entries()), now, window)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("window %v: endpoint evidence changed\nbefore %+v\nafter  %+v", window, before, after)
		}
	}
	h := newRCHarness(t, rcA063Entries())
	rcAccept(t, h)
	real, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	last := real[len(real)-1]
	if len(last.Calls) != 0 {
		t.Fatalf("the reconcile line carries %d call(s)", len(last.Calls))
	}
	for _, window := range []time.Duration{time.Hour, 365 * 24 * time.Hour} {
		if !reflect.DeepEqual(SucceededEndpoints(real[:len(real)-1], now, window), SucceededEndpoints(real, now, window)) {
			t.Fatalf("window %v: the real reconcile line changed endpoint evidence", window)
		}
	}
}

// TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged — 측정 속성·미검증 목록·단계·멱등 재생은 그대로이고
// Outstanding 에서 대상만 빠진다(attestation 은 이 줄을 보지 않는다).
func TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged(t *testing.T) {
	before := BuildReport("r", rcA063Entries(), rcNow)
	after := BuildReport("r", rcEntriesWithReconcileLine(t, rcA063Entries()), rcNow)
	if !reflect.DeepEqual(before.Groups, after.Groups) || !reflect.DeepEqual(before.Unverified, after.Unverified) ||
		!reflect.DeepEqual(before.Steps, after.Steps) || before.ReplayEnabled != after.ReplayEnabled ||
		before.AccountRef != after.AccountRef {
		t.Fatal("the reconcile line changed a measured attribute, a step verdict or replay")
	}
	if rcHas(after.Outstanding, rcTargetID) {
		t.Fatal("report still lists the reconciled artifact as outstanding")
	}
}

// TestReconcileTextLabelsReconciledAbsentInReportAndStatus 는 로트 1 S1 이다 — status·report 텍스트가 대사된 artifact 를
// "reconciled absent" 로 보이고, "아직 살아 있다" 절(그리고 그 취소 권고)에 남기지 않으며, 취소·체결로 쓰지 않는다.
func TestReconcileTextLabelsReconciledAbsentInReportAndStatus(t *testing.T) {
	entries := rcEntriesWithReconcileLine(t, rcA063Entries())
	var status, report bytes.Buffer
	BuildProgress("r", entries).WriteText(&status)
	BuildReport("r", entries, rcNow).WriteText(&report)
	for name, text := range map[string]string{"status": status.String(), "report": report.String()} {
		if strings.Contains(text, "아직 계좌에 살아 있다") {
			t.Fatalf("%s still lists a live object after reconciliation:\n%s", name, text)
		}
		var labelled bool
		for _, line := range strings.Split(text, "\n") {
			if !strings.Contains(line, rcTargetID) {
				continue
			}
			lower := strings.ToLower(line)
			if strings.Contains(lower, "reconciled absent") {
				labelled = true
			}
			for _, wrong := range []string{"cancelled", "filled", "취소", "체결"} {
				if strings.Contains(lower, wrong) {
					t.Fatalf("%s labels the reconciled artifact as %q: %q", name, wrong, line)
				}
			}
		}
		if !labelled {
			t.Fatalf("%s does not show %s as \"reconciled absent\":\n%s", name, rcTargetID, text)
		}
	}
}

// --- helpers ---------------------------------------------------------------------

func rcHas(arts []Artifact, id string) bool {
	for _, a := range arts {
		if a.ID == id {
			return true
		}
	}
	return false
}

func rcIDs(arts []Artifact) []string {
	out := []string{}
	for _, a := range arts {
		out = append(out, a.ID)
	}
	return out
}

func rcCanon(t *testing.T, e Entry) string {
	t.Helper()
	return Digest(e)
}

type rcErr string

func (e rcErr) Error() string { return string(e) }
