package verifylive

// reconcile_check.go — a121 대사 판정의 보조 함수들. 순서는 Reconcile 이 정하고, 여기 함수는 각자 가드 하나만 판정함.

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// reconcileListLimit 는 페이지당 행 수임(design G1 「치르는 값」 — 10 페이지 × 100).
const reconcileListLimit = 100

// reconcileQ6Message 는 Q6 거절 문구임(design G1-2 · Q6 — 거절 메시지와 문서 양쪽에 적음).
const reconcileQ6Message = "만료된 artifact 는 이 경로로 영구히 대사되지 않는다"

// reconcileOpenFamily 는 OPEN 그룹의 조건주문 status 어휘임 — CLOSED 행에 나오면 그룹과 모순(codex F4).
var reconcileOpenFamily = map[string]bool{"WATCHING": true, "PAUSED": true, "ORDERING": true, "ORDERED": true}

func refuse(code ReconcileRefusalCode, detail string) error {
	return &ReconcileRefusal{Code: code, Detail: detail}
}

// --- 후보 선택(G2) ---------------------------------------------------------------

// selectReconcileCandidate 는 재개 정리 선택이 지금 내놓을 조건주문 중 M0 가 가리키지 않는 것을 정확히 하나 고름.
// 호출자는 id 를 주지 않음 — 보유가 풀리지 않은 것·M0 수동 대사 대상은 PendingCleanup 이 이미 뺌.
func selectReconcileCandidate(entries []Entry) (Artifact, error) {
	var conditionals []Artifact
	for _, a := range PendingCleanup(entries) {
		if a.Kind == KindConditional {
			conditionals = append(conditionals, a)
		}
	}
	candidates, err := withoutM0Unsettled(entries, conditionals)
	if err != nil {
		return Artifact{}, err
	}
	switch len(candidates) {
	case 0:
		return Artifact{}, refuse(RefuseNoCandidate, "no released outstanding conditional artifact to reconcile")
	case 1:
		return candidates[0], nil
	default:
		return Artifact{}, refuse(RefuseMultipleCandidates,
			fmt.Sprintf("%d released outstanding conditional artifacts — reconciliation needs exactly one", len(candidates)))
	}
}

// withoutM0Unsettled 는 M0Unsettled 가 가리키는 부모 조건주문을 후보에서 뺌. M0 판정 자체가 실패(복수 owner)하면
// 무엇을 빼야 할지 모르므로 거절(fail-closed).
func withoutM0Unsettled(entries []Entry, candidates []Artifact) ([]Artifact, error) {
	owner, ok, err := M0Unsettled(entries)
	if err != nil {
		return nil, refuse(RefuseM0Unresolved, "an unresolved M0 owner needs manual reconciliation first: "+err.Error())
	}
	if !ok || owner.ParentConditionalID == "" {
		return candidates, nil
	}
	out := make([]Artifact, 0, len(candidates))
	for _, a := range candidates {
		if a.Kind == KindConditional && a.ID == owner.ParentConditionalID {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

// --- 계좌 결속(G3-1) --------------------------------------------------------------

// bindReconcileAccount 는 대상을 언급한 모든 줄의 마스크가 하나이고 현재 자격의 마스크와 같은지 봄.
// 끝 4자리 일치는 신원이 아님 — 그래서 승인 출력이 두 마스크를 보임(F1).
func bindReconcileAccount(entries []Entry, target Artifact, currentRef string) (string, string, error) {
	refs := map[string]bool{}
	for _, e := range entries {
		for _, a := range e.Artifacts {
			if a.Kind == target.Kind && a.ID == target.ID {
				refs[strings.TrimSpace(e.AccountRef)] = true
			}
		}
	}
	if len(refs) != 1 {
		return "", "", refuse(RefuseAccountMixed,
			fmt.Sprintf("the lines naming the artifact carry %d different account references", len(refs)))
	}
	var recordMask string
	for ref := range refs {
		recordMask = ref
	}
	currentMask := attest.Mask(strings.TrimSpace(currentRef))
	if unusableMask(recordMask) || unusableMask(currentMask) {
		return "", "", refuse(RefuseAccountRefUnusable,
			"an account reference retains no digits (empty, \"(none)\" or four characters or fewer) — the comparison would degrade to a length check")
	}
	if currentMask != recordMask {
		return "", "", refuse(RefuseAccountMismatch, "the current credentials' masked account differs from the record's")
	}
	return recordMask, currentMask, nil
}

// unusableMask 는 남은 자리가 하나도 없는 마스크를 가려냄.
func unusableMask(mask string) bool {
	return mask == "" || mask == "(none)" || strings.Trim(mask, "*") == ""
}

// --- Q1 · Q3 ------------------------------------------------------------------------

// checkReconcileRetention 은 Q1(G1-4) — 측정한 기간 기반 보존 한도 안의 artifact 만 통과.
func checkReconcileRetention(target Artifact, now time.Time) error {
	m := reconcileRetention
	if m == nil || m.Bound <= 0 {
		return refuse(RefuseRetentionUnmeasured, "the fired-conditional retention of the CLOSED list has not been measured (Q1)")
	}
	if m.Eviction != evictionDuration {
		return refuse(RefuseRetentionEviction,
			fmt.Sprintf("the measured CLOSED eviction model is %q — only a duration-based model can be covered", string(m.Eviction)))
	}
	if target.CreatedAt.IsZero() {
		return refuse(RefuseCreatedAtZero, "the outstanding line carries no creation time, so its age cannot be measured")
	}
	age := now.Sub(target.CreatedAt)
	if age < 0 || age > m.Bound {
		return refuse(RefuseArtifactTooOld, fmt.Sprintf("artifact age %s is outside the measured retention bound %s", age, m.Bound))
	}
	return nil
}

// checkReconcileWindow 는 Q3 창(첫 목록 읽기 시작부터 지금까지)과 Q1 나이를 다시 잼(codex F7).
func checkReconcileWindow(target Artifact, start, now time.Time) error {
	elapsed := now.Sub(start)
	if elapsed < 0 || elapsed > reconcileFreshnessBound {
		return refuse(RefuseFreshnessExceeded,
			fmt.Sprintf("%s elapsed since the first list read — the freshness bound is %s", elapsed, reconcileFreshnessBound))
	}
	return checkReconcileRetention(target, now)
}

// --- 양성 대조(P1-3) -----------------------------------------------------------------

func checkReconcileInstrument(ctx context.Context, r ReconcileReader, symbol string) error {
	echo, err := r.ReconcileInstrument(ctx, symbol)
	if err != nil {
		return refuse(RefuseInstrumentControl, "the instrument read for the artifact's symbol failed: "+err.Error())
	}
	if echo != symbol {
		return refuse(RefuseInstrumentControl, "the instrument read echoed a different symbol")
	}
	return nil
}

// --- 두 번 읽기(G1-6) ----------------------------------------------------------------

// reconcileRead 는 한 번의 읽기 집합임 — 세 그룹 전 페이지.
type reconcileRead struct {
	condOpen  []official.ReconcileConditionalRow
	plainOpen []official.ReconcileOrderRow
	closed    []official.ReconcileConditionalRow
}

// basis 는 (그룹, id, status, triggeredOrderId) multiset 원소들임.
func (r reconcileRead) basis() []ReconcileBasisRow {
	var out []ReconcileBasisRow
	for _, row := range r.condOpen {
		out = append(out, ReconcileBasisRow{Group: ReconcileGroupConditionalOpen, ID: row.ID, Status: row.Status, TriggeredOrderID: row.TriggeredOrderID})
	}
	for _, row := range r.plainOpen {
		out = append(out, ReconcileBasisRow{Group: ReconcileGroupPlainOpen, ID: row.ID, Status: row.Status})
	}
	for _, row := range r.closed {
		out = append(out, ReconcileBasisRow{Group: ReconcileGroupConditionalClosed, ID: row.ID, Status: row.Status, TriggeredOrderID: row.TriggeredOrderID})
	}
	return out
}

// readReconcileSet 은 고정 순서(조건주문 OPEN → 일반 OPEN → CLOSED)로 세 그룹을 끝까지 읽음. 한 읽기 안의
// (그룹, id) 중복은 그 자체로 거절(P1-4).
func readReconcileSet(ctx context.Context, r ReconcileReader, symbol string) (reconcileRead, error) {
	var out reconcileRead
	var err error
	if out.condOpen, err = readReconcileConditionalGroup(ctx, r, "OPEN", ReconcileGroupConditionalOpen, symbol); err != nil {
		return reconcileRead{}, err
	}
	if out.plainOpen, err = readReconcilePlainGroup(ctx, r, symbol); err != nil {
		return reconcileRead{}, err
	}
	if out.closed, err = readReconcileConditionalGroup(ctx, r, "CLOSED", ReconcileGroupConditionalClosed, symbol); err != nil {
		return reconcileRead{}, err
	}
	return out, nil
}

// reconcilePager 는 m0RecoverPending 과 같은 모양의 페이지 검사임 — 반복 커서·hasNext 인데 빈 커서·상한·오류.
type reconcilePager struct {
	group  string
	cursor string
	seen   map[string]bool
	ids    map[string]bool
}

func newReconcilePager(group string) *reconcilePager {
	return &reconcilePager{group: group, seen: map[string]bool{}, ids: map[string]bool{}}
}

// before 는 페이지를 읽기 전 반복 커서를 거절함.
func (p *reconcilePager) before() error {
	if p.cursor != "" && p.seen[p.cursor] {
		return refuse(RefuseRepeatedCursor, p.group+": the list repeated a cursor")
	}
	if p.cursor != "" {
		p.seen[p.cursor] = true
	}
	return nil
}

// row 는 한 읽기 안의 (그룹, id) 중복을 거절함.
func (p *reconcilePager) row(id string) error {
	if p.ids[id] {
		return refuse(RefuseDuplicateRow, p.group+": an identifier appeared twice within one read")
	}
	p.ids[id] = true
	return nil
}

// after 는 페이지 경계를 판정함 — done 이면 그룹 끝. 상한(maxFixturePages)을 넘는 hasNext 는 호출 루프의 끝이
// RefusePageCap 으로 거절함(한 곳).
func (p *reconcilePager) after(hasNext bool, next string) (bool, error) {
	if !hasNext {
		return true, nil
	}
	if strings.TrimSpace(next) == "" {
		return false, refuse(RefuseEmptyCursor, p.group+": hasNext with an empty cursor")
	}
	p.cursor = next
	return false, nil
}

func readReconcileConditionalGroup(ctx context.Context, r ReconcileReader, status, group, symbol string) ([]official.ReconcileConditionalRow, error) {
	p := newReconcilePager(group)
	var rows []official.ReconcileConditionalRow
	for page := 0; page < maxFixturePages; page++ {
		if err := p.before(); err != nil {
			return nil, err
		}
		got, err := r.ReconcileConditionalOrdersPage(ctx, status, symbol, p.cursor, reconcileListLimit)
		if err != nil {
			return nil, refuse(RefuseReadError, group+": "+err.Error())
		}
		for _, row := range got.Rows {
			if err := p.row(row.ID); err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		done, err := p.after(got.HasNext, got.NextCursor)
		if err != nil {
			return nil, err
		}
		if done {
			return rows, nil
		}
	}
	return nil, refuse(RefusePageCap, fmt.Sprintf("%s: more than %d pages", group, maxFixturePages))
}

func readReconcilePlainGroup(ctx context.Context, r ReconcileReader, symbol string) ([]official.ReconcileOrderRow, error) {
	p := newReconcilePager(ReconcileGroupPlainOpen)
	var rows []official.ReconcileOrderRow
	for page := 0; page < maxFixturePages; page++ {
		if err := p.before(); err != nil {
			return nil, err
		}
		got, err := r.ReconcileOpenOrdersPage(ctx, symbol, p.cursor, reconcileListLimit)
		if err != nil {
			return nil, refuse(RefuseReadError, ReconcileGroupPlainOpen+": "+err.Error())
		}
		for _, row := range got.Rows {
			if err := p.row(row.ID); err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
		done, err := p.after(got.HasNext, got.NextCursor)
		if err != nil {
			return nil, err
		}
		if done {
			return rows, nil
		}
	}
	return nil, refuse(RefusePageCap, fmt.Sprintf("%s: more than %d pages", ReconcileGroupPlainOpen, maxFixturePages))
}

// sameReconcileMultiset 은 두 읽기를 multiset 으로 비교함(집합으로 접지 않음).
func sameReconcileMultiset(a, b []ReconcileBasisRow) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := sortedReconcileRows(a), sortedReconcileRows(b)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func sortedReconcileRows(rows []ReconcileBasisRow) []ReconcileBasisRow {
	out := append([]ReconcileBasisRow{}, rows...)
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Group != b.Group {
			return a.Group < b.Group
		}
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		if a.Status != b.Status {
			return a.Status < b.Status
		}
		return a.TriggeredOrderID < b.TriggeredOrderID
	})
	return out
}

// --- 행 검사(OCO → 필드 결측 → 심볼 → 시장) -----------------------------------------

func checkReconcileRows(r reconcileRead, symbol, market string) error {
	conditional := append(append([]official.ReconcileConditionalRow{}, r.condOpen...), r.closed...)
	for _, row := range conditional {
		if row.HasSecond {
			return refuse(RefuseOCO, "a row carries a second (OCO) leg — OCO conditionals are not reconciled through this path")
		}
	}
	for _, row := range conditional {
		if row.ID == "" || row.Symbol == "" || row.Market == "" {
			return refuse(RefuseRowIncomplete, "a conditional row lacks its id, symbol or market")
		}
	}
	for _, row := range r.plainOpen {
		if row.ID == "" || row.Symbol == "" {
			return refuse(RefuseRowIncomplete, "a plain order row lacks its id or symbol")
		}
	}
	for _, row := range conditional {
		if row.Symbol != symbol {
			return refuse(RefuseRowSymbol, "a conditional row carries another symbol — the symbol filter was not honoured")
		}
	}
	for _, row := range r.plainOpen {
		if row.Symbol != symbol {
			return refuse(RefuseRowSymbol, "a plain order row carries another symbol — the symbol filter was not honoured")
		}
	}
	for _, row := range conditional {
		if row.Market != market {
			return refuse(RefuseRowMarket, "a conditional row reports another market")
		}
	}
	return nil
}

// --- 부재 검사(OPEN 조건 → 일반 OPEN → CLOSED target → 발동 흔적) --------------------

func checkReconcileAbsence(r reconcileRead, target Artifact) error {
	if len(r.condOpen) > 0 {
		return refuse(RefuseOpenConditional,
			fmt.Sprintf("%d open conditional order(s) on the symbol — a live successor may exist under a new identifier", len(r.condOpen)))
	}
	if len(r.plainOpen) > 0 {
		return refuse(RefuseOpenPlainOrder,
			fmt.Sprintf("%d open plain order(s) on the symbol — a fired child may be resting", len(r.plainOpen)))
	}
	for _, row := range r.closed {
		if row.ID != target.ID {
			continue
		}
		if row.Status == "EXPIRED" {
			return refuse(RefuseClosedTargetExpired, "the artifact is listed CLOSED as EXPIRED: "+reconcileQ6Message+
				" — that is an ending, not an absence (Q6)")
		}
		return refuse(RefuseClosedTargetPresent, "the artifact's identifier is listed in the CLOSED group — it is not absent")
	}
	for _, row := range r.closed {
		if row.TriggeredOrderID != "" || row.Status == "COMPLETED" {
			return refuse(RefuseClosedFiredTrace, "the symbol's CLOSED group holds a fired conditional — a successor may have fired")
		}
	}
	return nil
}

// checkReconcileClosedStatus 는 CLOSED 행 status 를 허용 목록으로 판정함(F4 — 마지막). 결측·미지·OPEN 계열은 거절.
func checkReconcileClosedStatus(r reconcileRead) error {
	for _, row := range r.closed {
		if row.Status == "" || reconcileOpenFamily[row.Status] || !reconcileClosedStatusAllowlist[row.Status] {
			return refuse(RefuseClosedStatus,
				fmt.Sprintf("CLOSED row status %q is missing, unknown or inconsistent with its group", row.Status))
		}
	}
	return nil
}

// --- 승인 출력 ------------------------------------------------------------------------

// WriteText 는 승인 내용을 운영자 화면에 씀 — 두 마스크와 계좌 수를 반드시 보임(F1 처분 ①). 원문 계좌는 없음.
func (a ReconcileApproval) WriteText(w io.Writer) {
	fmt.Fprintf(w, "대사 대상 (공식 목록 조회 전용 · 기록에 사건 한 줄 추가 · 주문 변경 없음)\n")
	fmt.Fprintf(w, "  artifact         %s %s (%s, %s)\n", a.Kind, a.ID, a.Symbol, a.Market)
	fmt.Fprintf(w, "  기록의 계좌       %s\n", a.RecordAccountMask)
	fmt.Fprintf(w, "  현재 자격의 계좌  %s\n", a.CurrentAccountMask)
	fmt.Fprintf(w, "  자격의 계좌 수    %d\n", a.AccountCount)
	fmt.Fprintln(w, "  끝 4자리는 신원이 아니다 — 두 마스크가 같아도 다른 계좌일 수 있다. 이 프로필의 자격을 바꾼 적이 있으면 진행하지 말 것.")
}
