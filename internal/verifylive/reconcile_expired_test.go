package verifylive

// reconcile_expired_test.go — a164 tasks 2.2~2.5·2.8: 만료 대사(`--expired`)의 증거 규칙·순서·기록.
//
// a121 시험과 같은 방식이다: 수락 픽스처(newRCExpiredHarness 기본값 = 모든 조건 참)에서 **하나만** 비틀어, 거절이 그 가드의
// 코드로만 나올 수 있게 한다. 하네스는 Q1 보존 측정을 생산과 같은 nil 로 둔다 — 만료 모드는 Q1 을 부르지 않아야
// 수락이 성립한다(design D7).

import (
	"bytes"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// rcExpiredStatement 는 0.6 영수증 ① 모양의 by-id statement 다(대상 id·심볼·시장, SINGLE, EXPIRED/EXPIRED, 미발동, second 없음).
func rcExpiredStatement() official.ReconcileConditionalDetail {
	return official.ReconcileConditionalDetail{ID: rcTargetID, Type: "SINGLE", Status: "EXPIRED", Symbol: rcSymbol,
		Market: MarketKR, FirstStatus: "EXPIRED", ResultDigest: "sha256:00112233445566778899aabbccddeeff"}
}

// rcExpiredClosedRow 는 CLOSED 목록이 대상을 EXPIRED 로 이름한 행이다(영수증 ②).
func rcExpiredClosedRow() official.ReconcileConditionalRow {
	row := rcCondRow(rcTargetID, "EXPIRED")
	row.FirstStatus = "EXPIRED"
	return row
}

// newRCExpiredHarness 는 만료 모드 수락 픽스처다 — by-id 두 번 모두 statement, CLOSED 대상 행은 withClosedRow 일 때만.
func newRCExpiredHarness(t *testing.T, withClosedRow bool) *rcHarness {
	t.Helper()
	h := newRCHarness(t, rcA063Entries())
	h.expired = true
	rcSetPolicies(t, nil, rcFreshness) // 생산과 같은 Q1 부재
	h.reader.byID = func(int, string) (official.ReconcileConditionalDetail, error) { return rcExpiredStatement(), nil }
	if withClosedRow {
		h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcExpiredClosedRow()))
	}
	return h
}

// rcByID 는 n 번째 by-id 응답만 바꾼 스크립트다(나머지는 수락 statement).
func rcByID(h *rcHarness, at int, d official.ReconcileConditionalDetail, err error) {
	h.reader.byID = func(n int, _ string) (official.ReconcileConditionalDetail, error) {
		if n == at || at < 0 {
			return d, err
		}
		return rcExpiredStatement(), nil
	}
}

func rcStatement(edit func(*official.ReconcileConditionalDetail)) official.ReconcileConditionalDetail {
	d := rcExpiredStatement()
	edit(&d)
	return d
}

// rcRequireExpiredLine 는 design D5 의 줄 모양을 단언한다.
func rcRequireExpiredLine(t *testing.T, line map[string]any, raw string) {
	t.Helper()
	if line["kind"] != KindReconcile || line["step_id"] != string(StepReconcileExpired) || line["title"] != "reconciled expired" {
		t.Fatalf("line identity kind=%v step=%v title=%v", line["kind"], line["step_id"], line["title"])
	}
	if _, ok := line["calls"]; ok {
		t.Fatal("the expiry line carries calls — a by-id 200 would become endpoint evidence (P2-2)")
	}
	arts, _ := line["artifacts"].([]any)
	if len(arts) != 1 {
		t.Fatalf("want one artifact, got %v", line["artifacts"])
	}
	art := arts[0].(map[string]any)
	if art["id"] != rcTargetID || art["reconciled_expired"] != true || art["cancelled"] != false {
		t.Fatalf("artifact %v", art)
	}
	for _, k := range []string{"reconciled_absent", "filled", "held_until", "cancelled_at"} {
		if _, ok := art[k]; ok {
			t.Fatalf("expired artifact carries %q: %v", k, art)
		}
	}
	keys := []string{}
	obs, _ := line["observations"].([]any)
	for _, o := range obs {
		m := o.(map[string]any)
		keys = append(keys, m["key"].(string))
		if m["key"] == ObservationReconcileBasis && !strings.HasPrefix(m["value"].(string), ReconcileExpiredBasisDomain+":sha256:") {
			t.Fatalf("basis %v is not domain-tagged %s", m["value"], ReconcileExpiredBasisDomain)
		}
	}
	sort.Strings(keys)
	want := []string{ObservationReconcileBasis, ObservationReconcileExpiredByID, ObservationReconcileExpiredClosedRows}
	sort.Strings(want)
	if !reflect.DeepEqual(keys, want) {
		t.Fatalf("observation keys %v, want %v", keys, want)
	}
	if strings.Contains(raw, rcAccount) || line["account_ref"] != rcMask() {
		t.Fatalf("the line must carry only the record's mask: %s", raw)
	}
}

// --- 2.2 수락 두 갈래 --------------------------------------------------------------------

func TestExpiryReconcileAcceptsTwoStatementsWhenTheClosedListEvictedTheTarget(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	res, line, raw := rcAccept(t, h)
	if !res.Artifact.ReconciledExpired || res.Artifact.ReconciledAbsent || !strings.HasPrefix(res.Basis, ReconcileExpiredBasisDomain+":") {
		t.Fatalf("result %+v basis %s", res.Artifact, res.Basis)
	}
	rcRequireExpiredLine(t, line, raw)
}

func TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	_, line, raw := rcAccept(t, h)
	rcRequireExpiredLine(t, line, raw)
}

// TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow 는 A6 의 양성 대조다 — 같은 심볼의 다른 EXPIRED 행은 어휘 안.
func TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	other := rcCondRow("CO-OTHER-EXPIRED", "EXPIRED")
	other.FirstStatus = "EXPIRED"
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcExpiredClosedRow(), other))
	_, _, raw := rcAccept(t, h)
	if strings.Contains(raw, "CO-OTHER-EXPIRED") {
		t.Fatal("the line names another broker identifier (Q2 — only the artifact's own keys)")
	}
}

// TestExpiryReconcileAcceptsATerminalChainSibling 는 A7 의 양성 대조다 — 같은 체인의 다른 artifact 가 종결이면 통과.
func TestExpiryReconcileAcceptsATerminalChainSibling(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	rcWriteEntries(t, h.path, []Entry{{Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictPass, AccountRef: rcMask(),
		Artifacts: []Artifact{{Kind: KindOrder, ID: "ORD-CHAIN", Symbol: rcSymbol, CreatedAt: rcCreated, ChainID: rcChainID, Cancelled: true}}}})
	rcAccept(t, h)
}

// --- 2.3 거절 — 가드마다 코드 하나 -----------------------------------------------------------

func TestExpiryReconcileRefusesEachGuardWithItsOwnCode(t *testing.T) {
	silent := official.ReconcileConditionalDetail{Silent: true, RequestID: "req-1"}
	readErr := errors.New("official: rate limited")
	fired := rcCondRow("CO-OTHER-FIRED", "EXPIRED")
	fired.TriggeredOrderID = "ORD-CHILD"
	cases := map[string]struct {
		closed bool
		setup  func(h *rcHarness)
		want   ReconcileRefusalCode
	}{
		// A3 — by-id 두 응답 모두 statement(P1-1)
		"silent-twice-closed-silent-too": {false, func(h *rcHarness) { rcByID(h, -1, silent, nil) }, RefuseExpiryEvidenceAbsent},
		"silent-twice-closed-names-it":   {true, func(h *rcHarness) { rcByID(h, -1, silent, nil) }, RefuseExpirySourcesDisagree},
		// A4 — by-id statement
		"by-id-watching": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status, d.FirstStatus = "WATCHING", "WATCHING" }), nil)
		}, RefuseExpiryTargetLive},
		"by-id-paused": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "PAUSED" }), nil)
		}, RefuseExpiryTargetLive},
		"by-id-ordering": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "ORDERING" }), nil)
		}, RefuseExpiryTargetLive},
		"by-id-ordered": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "ORDERED" }), nil)
		}, RefuseExpiryTargetLive},
		"by-id-completed": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status, d.FirstStatus = "COMPLETED", "COMPLETED" }), nil)
		}, RefuseExpiryTargetFired},
		"by-id-triggered": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.TriggeredOrderID = "ORD-X" }), nil)
		}, RefuseExpiryTargetFired},
		"by-id-unknown": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "CANCELLED" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-empty": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-leg-out": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.FirstStatus = "ORDERED" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-leg-empty": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.FirstStatus = "" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-second": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.HasSecond = true }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-type-oco": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Type = "OCO" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-type-none": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Type = "" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-other-id": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.ID = "CO-OTHER" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-other-sym": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Symbol = "000660" }), nil)
		}, RefuseExpiryTargetStatus},
		"by-id-other-mkt": {false, func(h *rcHarness) {
			rcByID(h, -1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Market = MarketUS }), nil)
		}, RefuseExpiryTargetStatus},
		// A4 — CLOSED 대상 행(P2-3)
		"closed-target-leg-out": {false, func(h *rcHarness) {
			row := rcExpiredClosedRow()
			row.FirstStatus = "ORDERED"
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
		}, RefuseExpiryTargetStatus},
		"closed-target-leg-empty": {false, func(h *rcHarness) {
			row := rcExpiredClosedRow()
			row.FirstStatus = ""
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
		}, RefuseExpiryTargetStatus},
		"closed-target-unknown-status": {false, func(h *rcHarness) {
			row := rcExpiredClosedRow()
			row.Status = "WEIRD"
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
		}, RefuseExpiryTargetStatus},
		"closed-target-watching": {false, func(h *rcHarness) {
			row := rcExpiredClosedRow()
			row.Status = "WATCHING"
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
		}, RefuseExpiryTargetLive},
		// A2 — 읽기 실패·두 응답 불일치
		"by-id-first-fails":   {false, func(h *rcHarness) { rcByID(h, 0, official.ReconcileConditionalDetail{}, readErr) }, RefuseExpiryReadError},
		"by-id-second-fails":  {false, func(h *rcHarness) { rcByID(h, 1, official.ReconcileConditionalDetail{}, readErr) }, RefuseExpiryReadError},
		"by-id-first-silent":  {false, func(h *rcHarness) { rcByID(h, 0, silent, nil) }, RefuseExpiryReadsDiffer},
		"by-id-second-silent": {false, func(h *rcHarness) { rcByID(h, 1, silent, nil) }, RefuseExpiryReadsDiffer},
		"by-id-status-differs": {false, func(h *rcHarness) {
			rcByID(h, 1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.Status = "COMPLETED" }), nil)
		}, RefuseExpiryReadsDiffer},
		"by-id-leg-differs": {false, func(h *rcHarness) {
			rcByID(h, 1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.FirstStatus = "COMPLETED" }), nil)
		}, RefuseExpiryReadsDiffer},
		"by-id-trigger-differs": {false, func(h *rcHarness) {
			rcByID(h, 1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.TriggeredOrderID = "ORD-Y" }), nil)
		}, RefuseExpiryReadsDiffer},
		"by-id-second-differs": {false, func(h *rcHarness) {
			rcByID(h, 1, rcStatement(func(d *official.ReconcileConditionalDetail) { d.HasSecond = true }), nil)
		}, RefuseExpiryReadsDiffer},
		// A5 — 생존 판정은 증거로 완화하지 않음(P2-6 사례 a · P3-5 사례 b)
		"open-conditional-despite-evidence": {true, func(h *rcHarness) {
			h.reader.set(-1, ReconcileGroupConditionalOpen, rcCondPage("", false, rcCondRow("CO-PROTECT", "WATCHING")))
		}, RefuseOpenConditional},
		"open-plain-order": {true, func(h *rcHarness) {
			h.reader.set(-1, ReconcileGroupPlainOpen, rcPlainPage("", false, official.ReconcileOrderRow{ID: "ORD-REST", Symbol: rcSymbol, Status: "PENDING"}))
		}, RefuseOpenPlainOrder},
		"closed-target-completed": {false, func(h *rcHarness) {
			row := rcExpiredClosedRow()
			row.Status, row.FirstStatus = "COMPLETED", "COMPLETED"
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, row))
		}, RefuseClosedFiredTrace},
		"closed-other-fired": {true, func(h *rcHarness) {
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcExpiredClosedRow(), fired))
		}, RefuseClosedFiredTrace},
		// A6 — CLOSED 어휘(다른 id)
		"closed-other-unknown": {true, func(h *rcHarness) {
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcExpiredClosedRow(), rcCondRow("CO-OTHER", "WEIRD")))
		}, RefuseClosedStatus},
		"closed-other-empty": {true, func(h *rcHarness) {
			h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcExpiredClosedRow(), rcCondRow("CO-OTHER", "")))
		}, RefuseClosedStatus},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := newRCExpiredHarness(t, c.closed)
			c.setup(h)
			rcRefuse(t, h, c.want)
		})
	}
}

// TestExpiryReconcileRefusesAnOpenChainSibling 는 A7 이다 — 같은 체인의 다른 artifact 가 아직 outstanding 이면 거절.
func TestExpiryReconcileRefusesAnOpenChainSibling(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	rcWriteEntries(t, h.path, []Entry{{Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictFail, AccountRef: rcMask(),
		Artifacts: []Artifact{{Kind: KindOrder, ID: "ORD-CHAIN", Symbol: rcSymbol, CreatedAt: rcCreated, ChainID: rcChainID}}}})
	rcRefuse(t, h, RefuseExpiryChainOpen)
}

// TestExpiryReconcileRefusesWithoutAReceiptBackedVocabulary 는 spec 「the expiry vocabulary has no receipt yet」 이다 —
// 어휘가 비면 승인 앞에서 거절하고 아무것도 읽지 않는다.
func TestExpiryReconcileRefusesWithoutAReceiptBackedVocabulary(t *testing.T) {
	for name, which := range map[string]*map[string]bool{"leg": &reconcileExpiryLegVocabulary, "closed": &reconcileExpiryClosedVocabulary} {
		t.Run(name, func(t *testing.T) {
			h := newRCExpiredHarness(t, true)
			prev := *which
			*which = map[string]bool{}
			t.Cleanup(func() { *which = prev })
			rcRefuse(t, h, RefuseExpiryVocabularyUnfixed)
			if len(h.approvals) != 0 || len(h.reader.allCalls()) != 0 {
				t.Fatalf("approvals %d, reads %v — the vocabulary refusal is static and comes before approval", len(h.approvals), h.reader.allCalls())
			}
		})
	}
}

// TestExpiryVocabularyIsTheReceiptIntersectedWithTheContract 는 H5 다 — 어휘는 0.6 영수증(testdata 사본)에 나온 값이면서
// OpenAPI enum 안에 있는 값과 정확히 같다(영수증 없이 넓히면 실패).
func TestExpiryVocabularyIsTheReceiptIntersectedWithTheContract(t *testing.T) {
	var receipt struct {
		GetByID struct {
			Status string `json:"status"`
			First  struct {
				Status string `json:"status"`
			} `json:"first"`
		} `json:"get_by_id"`
		Closed struct {
			Orders []struct {
				Status string `json:"status"`
				First  struct {
					Status string `json:"status"`
				} `json:"first"`
			} `json:"orders"`
		} `json:"list_closed_symbol_333430"`
	}
	raw, err := os.ReadFile("testdata/a164-expiry-probe-2026-10-22.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	spec, err := os.ReadFile("../../docs/migration/openapi.latest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(spec, &contract); err != nil {
		t.Fatal(err)
	}
	enum := func(schema string) map[string]bool {
		out := map[string]bool{}
		for _, v := range contract.Components.Schemas[schema].Properties["status"].Enum {
			out[v] = true
		}
		if len(out) == 0 {
			t.Fatalf("contract %s.status enum is empty", schema)
		}
		return out
	}
	top, leg := enum("ConditionalOrderDetailResponse"), enum("ConditionalOrderCondition")
	wantTop, wantLeg := map[string]bool{}, map[string]bool{}
	add := func(dst, in map[string]bool, v string) {
		if in[v] {
			dst[v] = true
		}
	}
	add(wantTop, top, receipt.GetByID.Status)
	add(wantLeg, leg, receipt.GetByID.First.Status)
	for _, o := range receipt.Closed.Orders {
		add(wantTop, top, o.Status)
		add(wantLeg, leg, o.First.Status)
	}
	if len(wantTop) == 0 || len(wantLeg) == 0 {
		t.Fatal("the receipt names no contract status — the vocabulary would have no source")
	}
	if !reflect.DeepEqual(reconcileExpiryClosedVocabulary, wantTop) {
		t.Fatalf("closed vocabulary %v, want receipt ∩ contract %v", reconcileExpiryClosedVocabulary, wantTop)
	}
	if !reflect.DeepEqual(reconcileExpiryLegVocabulary, wantLeg) {
		t.Fatalf("leg vocabulary %v, want receipt ∩ contract %v", reconcileExpiryLegVocabulary, wantLeg)
	}
}

// --- 2.4 순서 ------------------------------------------------------------------------------

// TestExpiryReconcileStaticRefusalsComeBeforeApproval — 후보·계좌·Q3 는 만료 모드에서도 승인 앞.
func TestExpiryReconcileStaticRefusalsComeBeforeApproval(t *testing.T) {
	cases := map[string]struct {
		setup func(t *testing.T, h *rcHarness)
		want  ReconcileRefusalCode
	}{
		"freshness-unfixed": {func(t *testing.T, h *rcHarness) { rcSetPolicies(t, nil, 0) }, RefuseFreshnessUnfixed},
		"account-mismatch":  {func(t *testing.T, h *rcHarness) { h.account.Ref = "999-99-990000" }, RefuseAccountMismatch},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			h := newRCExpiredHarness(t, true)
			c.setup(t, h)
			rcRefuse(t, h, c.want)
			if len(h.approvals) != 0 || len(h.reader.allCalls()) != 0 {
				t.Fatalf("%s: approvals %d reads %v", name, len(h.approvals), h.reader.allCalls())
			}
		})
	}
}

// TestExpiryReconcileDeclinedApprovalReadsNothing 는 P2-1 이다 — 승인 거절이면 by-id 를 포함해 어떤 읽기도 없다.
func TestExpiryReconcileDeclinedApprovalReadsNothing(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	h.approveErr = errors.New("operator said no")
	rcRefuse(t, h, RefuseNotApproved)
	if calls := h.reader.allCalls(); len(calls) != 0 {
		t.Fatalf("a declined approval still read: %+v", calls)
	}
}

// TestExpiryReconcileDoesNotNeedTheQ1Measurement 은 D7 이다 — 같은 기록·같은 nil Q1 에서 만료 모드는 수락하고 부재 모드는
// retention-unmeasured 로 거절한다(Q1 을 부르는 쪽이 어느 모드인지를 행동으로 가름).
func TestExpiryReconcileDoesNotNeedTheQ1Measurement(t *testing.T) {
	rcAccept(t, newRCExpiredHarness(t, false))
	h := newRCExpiredHarness(t, false)
	h.expired = false
	rcRefuse(t, h, RefuseRetentionUnmeasured)
}

// TestExpiryReconcileReadsInTheContractOrder 는 D4·P2-5 ⑤ 다 — 종목 → by-id ① → (조건 OPEN → 일반 OPEN → CLOSED) ×2 →
// by-id ② → 종목. 일반 OPEN 읽기는 두 번이다(만료 모드도 readReconcileSet 을 그대로 씀).
func TestExpiryReconcileReadsInTheContractOrder(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	rcAccept(t, h)
	var got []string
	for _, c := range h.reader.allCalls() {
		got = append(got, c.Kind+":"+c.Group)
	}
	want := []string{"instrument:", "by-id:",
		"list:" + ReconcileGroupConditionalOpen, "list:" + ReconcileGroupPlainOpen, "list:" + ReconcileGroupConditionalClosed,
		"list:" + ReconcileGroupConditionalOpen, "list:" + ReconcileGroupPlainOpen, "list:" + ReconcileGroupConditionalClosed,
		"by-id:", "instrument:"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("read order\n got %v\nwant %v", got, want)
	}
	for _, c := range h.reader.allCalls() {
		if c.Kind == "by-id" && c.Symbol != rcTargetID {
			t.Fatalf("by-id read asked for %q, want the record-selected %q", c.Symbol, rcTargetID)
		}
	}
}

// TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead 는 P3-1 이다 — by-id ① 뒤의 지연도 Q3 창 안에서 센다.
func TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	first := true
	h.reader.hook = func(c rcCall) {
		if c.Kind == "by-id" && first {
			first = false
			h.clock.advance(rcFreshness + time.Second)
		}
	}
	rcRefuse(t, h, RefuseFreshnessExceeded)
}

// TestExpiryReconcileApprovalNamesTheModeAndTheByIDReads 는 P3-6 이다 — 만료 모드 승인 화면은 모드와 「상세 조회 2회」 를
// 말하고, 부재 모드 화면은 a121 그대로다.
func TestExpiryReconcileApprovalNamesTheModeAndTheByIDReads(t *testing.T) {
	h := newRCExpiredHarness(t, false)
	rcAccept(t, h)
	if len(h.approvals) != 1 || h.approvals[0].Mode != ReconcileModeExpired {
		t.Fatalf("approvals %+v", h.approvals)
	}
	for _, want := range []string{"reconciled expired", "상세 조회 2회", rcTargetID} {
		if !strings.Contains(h.outAtApproval, want) {
			t.Fatalf("expiry approval screen lacks %q:\n%s", want, h.outAtApproval)
		}
	}
	var absent bytes.Buffer
	a := h.approvals[0]
	a.Mode = ""
	if err := a.WriteText(&absent); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(absent.String(), "상세 조회") || !strings.Contains(absent.String(), "공식 목록 조회 전용") {
		t.Fatalf("absence approval screen changed:\n%s", absent.String())
	}
}

// TestExpiryReconcileByIDReadsSitAfterApprovalInsideTheWindow 는 구조 핀이다 — Reconcile 본문에서 by-id 읽기는 정확히 두 번,
// 승인 뒤, 첫 목록 읽기 앞 하나와 둘째 목록 읽기 뒤·둘째 종목 대조 앞 하나. 만료 줄 직렬화는 잠금 뒤·마지막 게이트 앞.
func TestExpiryReconcileByIDReadsSitAfterApprovalInsideTheWindow(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reconcile.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, d := range file.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "Reconcile" {
			ast.Inspect(fn.Body, func(n ast.Node) bool {
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
		}
	}
	at := func(name string) []int {
		var out []int
		for i, c := range calls {
			if c == name {
				out = append(out, i)
			}
		}
		return out
	}
	byID, sets, inst, approve := at("readReconcileByID"), at("readReconcileSet"), at("checkReconcileInstrument"), at("Approve")
	if len(byID) != 2 || len(sets) != 2 || len(inst) != 2 || len(approve) != 1 {
		t.Fatalf("landmarks byID %v sets %v instrument %v approve %v in %v", byID, sets, inst, approve, calls)
	}
	if !(approve[0] < inst[0] && inst[0] < byID[0] && byID[0] < sets[0] && sets[1] < byID[1] && byID[1] < inst[1]) {
		t.Fatalf("want approve → instrument → by-id → list ×2 → by-id → instrument; calls %v", calls)
	}
	lock, encode, gates := at("lockRecordForAppend"), at("encodeReconcileExpiredLine"), at("checkReconcileWindow")
	if len(lock) != 1 || len(encode) != 1 || len(gates) != 2 || !(lock[0] < encode[0] && encode[0] < gates[1]) {
		t.Fatalf("the expiry line must be serialised after the lock and before the final gate; calls %v", calls)
	}
}

// --- 2.5 기록 · 2.8 표시 ---------------------------------------------------------------------

// TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader — Outstanding · PendingCleanup · 보고 투영이 넷째 종결을 따르고,
// 실패한 정리 줄과 단계 판정은 그대로다.
func TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	before, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	rcAccept(t, h)
	after, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	if !rcHas(Outstanding(before), rcTargetID) || !rcHas(PendingCleanup(before), rcTargetID) {
		t.Fatal("fixture: the target must start outstanding and planned for cleanup")
	}
	if rcHas(Outstanding(after), rcTargetID) || rcHas(PendingCleanup(after), rcTargetID) {
		t.Fatal("a reconciled-expired artifact is still outstanding or planned for cleanup")
	}
	rec := ReconciledArtifacts(after)
	if len(rec) != 1 || rec[0].ID != rcTargetID || !rec[0].ReconciledExpired {
		t.Fatalf("reconciled projection %+v", rec)
	}
	for _, s := range Steps() {
		b, okB := LastEntry(before, s.ID)
		a, okA := LastEntry(after, s.ID)
		if okB != okA || b.Verdict != a.Verdict || Settled(before, s.ID) != Settled(after, s.ID) || Passed(before, s.ID) != Passed(after, s.ID) {
			t.Fatalf("step %s verdict changed", s.ID)
		}
	}
}

// TestExpiryReconcileLiftsTheOutstandingHoldOfVerifyNew — runner.go 의 미결 HOLD(New B9)가 넷째 종결을 따른다(cmd 쪽은
// cmd/tossctl 의 짝 시험).
func TestExpiryReconcileLiftsTheOutstandingHoldOfVerifyNew(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	build := func(prior []Entry) error {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		receipt, err := OpenCausalReceipt(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer receipt.Close()
		recorder, err := OpenRecorder(t.TempDir() + "/record.jsonl")
		if err != nil {
			t.Fatal(err)
		}
		defer recorder.Close()
		op := alwaysConfirm()
		for _, step := range Steps() {
			if step.ID != StepConditionalTrigger {
				prior = append(prior, Entry{Kind: KindStep, StepID: step.ID, Verdict: VerdictPass})
			}
		}
		_, err = New(Options{Broker: official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token",
			official.WithBaseURL("http://127.0.0.1")), Recorder: recorder, Confirm: op.confirmer(), ConfirmBatch: op.batchConfirmer(),
			AccountRef: "1", IncludeTrigger: true, ConfirmEach: true, Resume: true, Redo: []StepID{StepConditionalTrigger},
			Receipt: receipt, Prior: prior})
		return err
	}
	register := []Entry{{Kind: KindStep, StepID: StepConditionalRegister, Verdict: VerdictPass, Artifacts: []Artifact{rcTarget()}}}
	if err := build(register); err == nil || !strings.Contains(err.Error(), "outstanding") {
		t.Fatalf("control: an outstanding conditional must hold New; got %v", err)
	}
	rcAccept(t, h)
	all, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	ended := append(append([]Entry{}, register...), all[len(all)-1])
	if err := build(ended); err != nil && strings.Contains(err.Error(), "outstanding") {
		t.Fatalf("a reconciled-expired artifact still holds New: %v", err)
	}
}

// TestExpiryReconcileIsIdempotent — 같은 artifact 를 다시 대사하면 후보가 없어 아무것도 쓰지 않는다.
func TestExpiryReconcileIsIdempotent(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	rcAccept(t, h)
	h.reader = newRCFakeReader()
	h.reader.byID = func(int, string) (official.ReconcileConditionalDetail, error) { return rcExpiredStatement(), nil }
	rcRefuse(t, h, RefuseNoCandidate)
	if n := len(h.reader.allCalls()); n != 0 {
		t.Fatalf("the repeat read the broker %d time(s)", n)
	}
}

// TestExpiryReconcileLineNeverBecomesEndpointEvidence 는 P2-2 다 — 줄에 calls 가 없고 SucceededEndpoints 가 같다.
func TestExpiryReconcileLineNeverBecomesEndpointEvidence(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	rcAccept(t, h)
	all, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	last := all[len(all)-1]
	if len(last.Calls) != 0 || last.Verdict != "" || last.Mutating {
		t.Fatalf("the expiry line carries calls %d / verdict %q / mutating %v", len(last.Calls), last.Verdict, last.Mutating)
	}
	now := rcNow.Add(time.Minute)
	for _, window := range []time.Duration{time.Hour, 365 * 24 * time.Hour} {
		if !reflect.DeepEqual(SucceededEndpoints(all[:len(all)-1], now, window), SucceededEndpoints(all, now, window)) {
			t.Fatalf("window %v: the expiry line changed endpoint evidence", window)
		}
	}
}

// TestExpiryReconcileTextLabelsReconciledExpired 는 2.8 이다 — status·report 가 「reconciled expired」 로 보이고 부재·취소·
// 체결로 쓰지 않는다.
func TestExpiryReconcileTextLabelsReconciledExpired(t *testing.T) {
	h := newRCExpiredHarness(t, true)
	rcAccept(t, h)
	entries, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	var status, report bytes.Buffer
	BuildProgress("r", entries).WriteText(&status)
	BuildReport("r", entries, rcNow).WriteText(&report)
	for name, text := range map[string]string{"status": status.String(), "report": report.String()} {
		labelled := false
		for _, line := range strings.Split(text, "\n") {
			if !strings.Contains(line, rcTargetID) {
				continue
			}
			lower := strings.ToLower(line)
			if strings.Contains(lower, "reconciled expired") {
				labelled = true
			}
			for _, wrong := range []string{"reconciled absent", "cancelled", "filled", "취소", "체결"} {
				if strings.Contains(lower, wrong) {
					t.Fatalf("%s labels the expired artifact %q: %q", name, wrong, line)
				}
			}
		}
		if !labelled {
			t.Fatalf("%s does not show %s as reconciled expired:\n%s", name, rcTargetID, text)
		}
	}
}

// TestExpiryReconciledArtifactSerialisesTheSettledShape — 넷째 종결도 확정형(보유·취소 시각 없음)으로 쓰고, 부재 확정형은
// 그대로다.
func TestExpiryReconciledArtifactSerialisesTheSettledShape(t *testing.T) {
	a := rcTarget()
	a.HeldUntil = StepConditionalCancel
	a.ReconciledExpired = true
	a.ReconciledAt = rcNow
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["reconciled_expired"] != true || m["cancelled"] != false {
		t.Fatalf("expired shape %s", b)
	}
	for _, k := range []string{"held_until", "reconciled_absent", "cancelled_at", "filled"} {
		if _, ok := m[k]; ok {
			t.Fatalf("expired shape carries %q: %s", k, b)
		}
	}
	absent := rcTarget()
	absent.ReconciledAbsent = true
	absent.ReconciledAt = rcNow
	got, err := json.Marshal(absent)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"kind":"conditional-order","id":"` + rcTargetID + `","symbol":"` + rcSymbol + `","created_at":"` +
		rcCreated.Format(time.RFC3339) + `","cancelled":false,"deliberate":true,"chain_id":"` + rcChainID +
		`","reconciled_absent":true,"reconciled_at":"` + rcNow.Format(time.RFC3339) + `"}`; string(got) != want {
		t.Fatalf("absence shape changed\n got %s\nwant %s", got, want)
	}
}

// TestReconcileExpiredStepIDIsOutsideTheCatalogueAndTheAbsenceStep — StepID 만 비교하는 소비자가 이 줄을 단계 판정으로 읽지
// 않게(design D5).
func TestReconcileExpiredStepIDIsOutsideTheCatalogueAndTheAbsenceStep(t *testing.T) {
	taken := map[StepID]bool{StepCleanup: true, StepAbort: true, StepReconcile: true, "": true}
	for _, s := range Steps() {
		taken[s.ID] = true
	}
	if taken[StepReconcileExpired] {
		t.Fatalf("StepReconcileExpired %q collides", StepReconcileExpired)
	}
}
