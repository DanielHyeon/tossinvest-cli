package verifylive

// reconcile_expired.go 는 a164(verify-reconcile-closes-broker-expired-conditional) 만료 대사 판정이다.
//
// 부재 대사(a121)와 같은 사전 검사·계좌 결속·승인·두 번 읽기·행 검사·잠금·추가 경로를 쓰고(Reconcile 안의 p.Expired 분기),
// 판정 하나만 바꿈: 대상 id 의 종결이 이번 실행의 by-id 상세 GET 두 번으로 EXPIRED·미발동이라고 **양성** 확인될 때만 닫음.
// 설계 정본: openspec/changes/a164-verify-reconcile-closes-broker-expired-conditional/design.md (D1~D10).

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// ReconcileModeExpired 는 승인 화면·질문이 이름하는 만료 대사 모드임(design D1).
const ReconcileModeExpired = "expired"

// StepReconcileExpired 는 만료 대사 줄의 StepID 임 — Steps() 카탈로그 ∪ {cleanup, abort, reconcile-absent} 밖(design D5).
const StepReconcileExpired StepID = "reconcile-expired"

// ReconcileExpiredBasisDomain 은 만료 대사 근거 지문의 버전·도메인 태그임(design D5).
const ReconcileExpiredBasisDomain = "a164/reconcile-expired-basis/v1"

// 만료 대사 줄의 관측 키(design D5) — 근거 지문은 부재 대사와 같은 키 ObservationReconcileBasis 에 실음.
const (
	ObservationReconcileExpiredByID       = "reconcile.expired.by_id"
	ObservationReconcileExpiredClosedRows = "reconcile.expired.closed_rows"
)

// ReconcileGroupConditionalByID 는 근거 multiset 에서 by-id statement 행의 그룹 이름임(design D5).
const ReconcileGroupConditionalByID = "conditional-by-id"

// reconcileExpiryLegVocabulary 는 만료 leg(first.status) 어휘임 — H5 확정(2026-10-22): OpenAPI ConditionalOrderCondition.status
// enum ∩ 0.6 사람 실측 영수증(first.status EXPIRED). 영수증 사본 testdata/a164-expiry-probe-2026-10-22.json 과 골든 시험이 대조함.
// 설정·플래그 경로 없음 — 넓히려면 새 영수증과 리뷰된 커밋.
var reconcileExpiryLegVocabulary = map[string]bool{"EXPIRED": true}

// reconcileExpiryClosedVocabulary 는 만료 모드 CLOSED 행 status 허용 목록임(A6) — H5 확정: ConditionalOrderDetailResponse.status
// enum ∩ 영수증(top-level EXPIRED). 부재 모드의 reconcileClosedStatusAllowlist(비어 있음)는 건드리지 않음.
var reconcileExpiryClosedVocabulary = map[string]bool{"EXPIRED": true}

// checkReconcileExpiryVocabulary 는 만료 어휘 확정을 승인 앞에서 봄(D4 정적 거절) — 비면 모든 진술이 거절되므로 미리 멈춤.
func checkReconcileExpiryVocabulary() error {
	if len(reconcileExpiryLegVocabulary) == 0 || len(reconcileExpiryClosedVocabulary) == 0 {
		return refuse(RefuseExpiryVocabularyUnfixed, "the expiry vocabulary has no receipt-backed value (H5)")
	}
	return nil
}

// readReconcileByID 는 by-id 상세 읽기 한 번임 — 오류는 무엇이든 거절(429·5xx·인증·전송·스키마, 자동 재시도 없음 — D6).
// silent(404 conditional-order-not-found)는 오류가 아니라 분류로 돌아오고, 근거로 쓰일지는 A3 이 정함.
func readReconcileByID(ctx context.Context, r ReconcileReader, id string) (official.ReconcileConditionalDetail, error) {
	d, err := r.ReconcileConditionalOrder(ctx, id)
	if err != nil {
		return official.ReconcileConditionalDetail{}, refuse(RefuseExpiryReadError, "the by-id read failed: "+err.Error())
	}
	return d, nil
}

// checkReconcileByIDPair 는 A2 다 — 두 응답의 분류가 같아야 하고, 둘 다 statement 면 status · first.status ·
// triggeredOrderId · second 존재가 **필드마다** 같아야 함.
func checkReconcileByIDPair(byID [2]official.ReconcileConditionalDetail) error {
	a, b := byID[0], byID[1]
	if a.Silent != b.Silent {
		return refuse(RefuseExpiryReadsDiffer, "one by-id read is a statement and the other is silent (not found)")
	}
	if a.Status != b.Status || a.FirstStatus != b.FirstStatus || a.TriggeredOrderID != b.TriggeredOrderID || a.HasSecond != b.HasSecond {
		return refuse(RefuseExpiryReadsDiffer, "the two by-id statements disagree on status, first-leg status, triggered order or second leg")
	}
	return nil
}

// checkReconcileExpiry 는 만료 판정임(design D4 순서): A5 생존 판정 → A3 진술 요구 → A4 진술 일관성 → A7 체인 → A6 CLOSED 어휘.
// A5 는 부재 모드와 같은 보조 함수이고 어떤 증거 조합으로도 건너뛰지 않음(D2 위계).
func checkReconcileExpiry(first, second reconcileRead, byID [2]official.ReconcileConditionalDetail, target Artifact, entries []Entry, market string) error {
	if err := checkReconcileNoLiveOrder(first); err != nil {
		return err
	}
	if err := checkReconcileNoFiredTrace(first); err != nil {
		return err
	}
	// A3 — by-id 두 응답이 둘 다 statement 여야 함(P1-1). 404 는 정정·취소·축출을 구분하지 못하므로 근거가 아님(D3).
	closed := reconcileTargetClosedRows(first, second, target.ID)
	if byID[0].Silent || byID[1].Silent {
		if len(closed) > 0 {
			return refuse(RefuseExpirySourcesDisagree, fmt.Sprintf(
				"the detail read cannot find the artifact (requestId %s) while the CLOSED list names it — a modify leaves exactly this signature", byID[0].RequestID))
		}
		return refuse(RefuseExpiryEvidenceAbsent, fmt.Sprintf(
			"the detail read cannot find the artifact (requestId %s) — a 404 or an absence is not evidence of an ending", byID[0].RequestID))
	}
	// A4 — by-id statement 는 대상 자신(id 바이트 · 기록 심볼 · 기록 시장 · SINGLE)이어야 하고, 진술 집합 S 의 모든 원소가 만료·미발동.
	for i, d := range byID {
		if d.ID != target.ID || d.Symbol != target.Symbol || d.Market != market || d.Type != "SINGLE" {
			return refuse(RefuseExpiryTargetStatus, fmt.Sprintf(
				"by-id read %d names id/symbol/market/type %q/%q/%q/%q — not this record's SINGLE artifact", i+1, d.ID, d.Symbol, d.Market, d.Type))
		}
		if err := checkReconcileExpiryStatement(fmt.Sprintf("by-id read %d", i+1), d.Status, d.FirstStatus, d.TriggeredOrderID, d.HasSecond); err != nil {
			return err
		}
	}
	for _, row := range closed {
		if err := checkReconcileExpiryStatement("the CLOSED row", row.Status, row.FirstStatus, row.TriggeredOrderID, row.HasSecond); err != nil {
			return err
		}
	}
	// A7 — 같은 체인의 다른 artifact(이 도구가 만든 정정 후속)가 남아 있으면, 대상의 만료는 그것에 대해 말하지 않음.
	if chain := ChainOf(entries, target.Kind, target.ID); chain != "" {
		for _, a := range Outstanding(entries) {
			if a.ChainID == chain && (a.Kind != target.Kind || a.ID != target.ID) {
				return refuse(RefuseExpiryChainOpen, "another artifact of the same chain is still outstanding — the target's expiry says nothing about it")
			}
		}
	}
	// A6 — CLOSED 행 status 허용 목록(마지막). 결측·미지·OPEN 계열은 거절.
	for _, row := range first.closed {
		if row.Status == "" || reconcileOpenFamily[row.Status] || !reconcileExpiryClosedVocabulary[row.Status] {
			return refuse(RefuseClosedStatus, fmt.Sprintf("CLOSED row status %q is outside the expiry-mode vocabulary", row.Status))
		}
	}
	return nil
}

// checkReconcileExpiryStatement 는 진술 하나의 A4 판정임 — 발동(COMPLETED·발동 id) → 살아 있음(OPEN 계열) → 그 밖(top-level
// EXPIRED 아님 · second 있음 · leg 어휘 밖).
func checkReconcileExpiryStatement(source, status, firstStatus, triggered string, hasSecond bool) error {
	switch {
	case status == "COMPLETED" || triggered != "":
		return refuse(RefuseExpiryTargetFired, source+" says the artifact fired (COMPLETED or a triggered order) — an expiry is not a fill")
	case reconcileOpenFamily[status]:
		return refuse(RefuseExpiryTargetLive, source+" says the artifact is live ("+status+")")
	case status != "EXPIRED" || hasSecond || !reconcileExpiryLegVocabulary[firstStatus]:
		return refuse(RefuseExpiryTargetStatus, fmt.Sprintf("%s reports status %q, first-leg %q, second leg %v — not an untriggered single-leg expiry",
			source, status, firstStatus, hasSecond))
	}
	return nil
}

// reconcileTargetClosedRows 는 두 읽기의 CLOSED 에서 대상 id 를 이름한 행들임 — first.status 는 multiset 비교 tuple 밖이라 두
// 읽기를 각각 봄(codex CG-1 과 같은 이유).
func reconcileTargetClosedRows(first, second reconcileRead, id string) []official.ReconcileConditionalRow {
	var out []official.ReconcileConditionalRow
	for _, r := range []reconcileRead{first, second} {
		for _, row := range r.closed {
			if row.ID == id {
				out = append(out, row)
			}
		}
	}
	return out
}

// ReconcileExpiredBasisDigest 는 만료 근거 지문임 — 목록 multiset 에 by-id statement 행(그룹 conditional-by-id)을 더해 버전·
// 도메인 태그와 함께 정규 순서로 지문화함(design D5).
func ReconcileExpiredBasisDigest(rows []ReconcileBasisRow, byID [2]official.ReconcileConditionalDetail) string {
	all := append([]ReconcileBasisRow{}, rows...)
	for _, d := range byID {
		all = append(all, ReconcileBasisRow{Group: ReconcileGroupConditionalByID, ID: d.ID, Status: d.Status, TriggeredOrderID: d.TriggeredOrderID})
	}
	return ReconcileExpiredBasisDomain + ":" + Digest(struct {
		Domain string              `json:"domain"`
		Rows   []ReconcileBasisRow `json:"rows"`
	}{Domain: ReconcileExpiredBasisDomain, Rows: sortedReconcileRows(all)})
}

// encodeReconcileExpiredLine 은 만료 대사 줄 하나를 개행까지 직렬화함 — Calls 없음(P2-2), 계좌는 기록의 마스크, 관측에는
// 판정한 원 status 와 응답 분류·결과 지문만(원 응답·다른 broker id 없음 — Q2).
func encodeReconcileExpiredLine(accountMask string, art Artifact, basis string, first, second reconcileRead,
	byID [2]official.ReconcileConditionalDetail, start, at time.Time) ([]byte, error) {
	runID, err := reconcileToken("reconcile")
	if err != nil {
		return nil, err
	}
	var reads, classes []string
	for i, d := range byID {
		// 분류는 응답에서 유도함(gstack 리뷰 F4) — 상수로 박으면 A3 가 바뀔 때 기록이 조용히 거짓이 됨.
		class := "statement"
		if d.Silent {
			class = "silent"
		}
		classes = append(classes, class)
		reads = append(reads, fmt.Sprintf("%d: status=%s first=%s triggered=%q second=%v result=%s", i+1, d.Status, d.FirstStatus,
			d.TriggeredOrderID, d.HasSecond, d.ResultDigest))
	}
	closedStatuses, closedLegs := []string{}, []string{}
	for _, row := range reconcileTargetClosedRows(first, second, art.ID) {
		closedStatuses = append(closedStatuses, row.Status)
		closedLegs = append(closedLegs, row.FirstStatus)
	}
	closedValue := strings.Join(closedStatuses, ",")
	if closedValue == "" {
		closedValue = "none"
	}
	b, err := json.Marshal(Entry{
		FormatVersion: RecordFormatVersion,
		Kind:          KindReconcile,
		RunID:         runID,
		StepID:        StepReconcileExpired,
		Title:         "reconciled expired",
		StartedAt:     start.UTC(),
		FinishedAt:    at.UTC(),
		AccountRef:    accountMask,
		Observations: []Observation{
			{Key: ObservationReconcileBasis, Value: basis,
				Detail: "two by-id reads named the artifact EXPIRED and untriggered and two complete list reads found no live or fired order on the symbol; not a cancel, fill or endpoint proof"},
			{Key: ObservationReconcileExpiredByID, Value: strings.Join(classes, ","), Detail: strings.Join(reads, "; ")},
			{Key: ObservationReconcileExpiredClosedRows, Value: closedValue, Detail: "first-leg " + strings.Join(closedLegs, ",")},
		},
		Artifacts: []Artifact{art},
	})
	if err != nil {
		return nil, fmt.Errorf("verify reconcile: encoding the expiry line: %w", err)
	}
	return append(b, '\n'), nil
}
