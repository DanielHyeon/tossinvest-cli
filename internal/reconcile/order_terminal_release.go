package reconcile

// order_terminal_release.go 는 missing-order 종목 차단을 「그것을 만든 주문들의 종결 증명」으로 푸는
// 추적기 쪽 절반임(a192 design D2~D5).
//
// # 결속은 프로세스 기억이고, 삽입 증명으로만 생김
//
// 조정 크레딧처럼 결속도 프로세스 기억임 — 재기동하면 사라지고 남은 행은 운영자 해제로 돌아감(fail-closed).
// 결속은 이 프로세스의 missing-order 관측이 장부 행을 **실제로 삽입했다**는 증명이 있을 때만 생김:
// `EnterReconcile` 은 삽입한 경우에만 `entered = true` 와 그 행 id 를 돌려주므로 그 반환값이 증명임.
// 이미 있던 행(투영 거절이 먼저 만든 행, 이전 프로세스의 행)에 겹친 관측은 `entered = false` 라 결속이
// 없고, 그 행은 이 경로로 풀리지 않음. 처음 만들어진 관측의 첫 쓰기에서 삽입되지 않은(재시도로 삽입된)
// 블록도 결속하지 않음 — 그 사이의 관측이 결속에 안 들어 있기 때문(M14).
//
// # 판정은 하나
//
// `orderTerminalReleaseVerdict` 하나가 1차 판정(Observe 의 깨끗한 갈래)과 해제 트랜잭션 안의 재판정을 둘 다
// 맡음. 호출자는 얼린 비교 입력·결속·장부 증명을 인자로 넘길 뿐 규칙을 따로 갖지 않음.
//
// # 무장-후-해제
//
// `RecordFill` 의 `committed_at` 은 tx 시작 전 시각이라 「비교 as-of > 종결 기록 시각」만으로는 그 비교의 로컬
// 읽기가 종결을 반영했다는 보장이 없음. 그래서 처음 통과한 비교는 무장(as-of 와 증명 digest 기억)만 하고,
// 같은 digest 로 그보다 나중 as-of 의 비교가 다시 통과할 때 해제함. 주기는 직렬이고 각 주기는 수집 뒤에
// 로컬을 읽으므로, 무장 뒤 비교의 로컬 읽기는 무장 때 이미 읽은 종결 커밋을 반영함.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// OrderTerminalJournal 은 종결 증명 해제가 쓰는 장부 쪽 세 메서드임. production 은 *journal.Journal.
//
// 비어 있으면(nil) 추적기는 오늘과 같은 행동을 함 — 결속도, 판정도, 해제도 없음.
type OrderTerminalJournal interface {
	MaxFillEventID(ctx context.Context) (int64, error)
	ReadOrderTerminalProof(ctx context.Context, req journal.OrderTerminalProofRequest) journal.OrderTerminalProof
	ReleaseReconcileOnOrderTerminalProof(ctx context.Context, req journal.OrderTerminalProofRequest,
		evidence string, decide func(journal.OrderTerminalProof) bool) (journal.ReconcileState, bool, error)
}

// boundOrder 는 결속된 missing-order 하나 — 정규 신원과 원래(추적) 번호임.
type boundOrder struct {
	identity OrderIdentity
	original string
}

// terminalBinding 은 이 프로세스가 삽입한 행 하나의 결속임.
type terminalBinding struct {
	rowID         string
	symbol        string
	orders        map[boundOrder]struct{}
	tainted       bool
	fillWatermark int64
	armed         bool
	armAsOf       time.Time
	armDigest     string
	// reason 은 마지막 판정(1차·tx 재판정)의 사유 코드임 — 「유지 + 사유」(M5)를 운영자가 볼 수 있게 남김.
	reason string
}

// terminalState 는 추적기의 결속 기억 전부임. 영값이 「결속 없음」이고 `Restore` 가 영값으로 되돌림.
type terminalState struct {
	bindings map[string]*terminalBinding
	// seen 은 이번 Observe 가 시작될 때 이미 있던 블록 키임 — 이번 관측에서 처음 만들어진 블록만 결속함.
	seen map[string]bool
}

// terminalComparison 은 판정에 넘기는 얼린 비교 입력임.
type terminalComparison struct {
	trackerAccount string
	diffAccount    string
	asOf           string
	disputed       bool
	blocksEntry    bool
	externalIDs    map[string]bool
}

// terminalRelease 는 1차 판정이 「해제」라고 답한 블록에 다는 표시 — 해제 tx 가 같은 입력으로 재판정함.
type terminalRelease struct {
	binding      terminalBinding
	comparison   terminalComparison
	request      journal.OrderTerminalProofRequest
	boundOrders  int
	latestCommit string
}

type terminalOutcome int

const (
	terminalHold terminalOutcome = iota
	terminalArm
	terminalReleaseOutcome
)

// terminalVerdict 는 판정 결과임. reason 은 유지·무장 사유 코드(시험·진단용).
type terminalVerdict struct {
	outcome      terminalOutcome
	reason       string
	digest       string
	asOf         time.Time
	boundOrders  int
	latestCommit string
}

// ReleasedByOrderTerminal 은 이 블록이 종결 증명으로 해제된 것인지와 결속 주문 수를 알려 줌(로그용).
func (b Block) ReleasedByOrderTerminal() (int, bool) {
	if b.terminalRelease == nil {
		return 0, false
	}
	return b.terminalRelease.boundOrders, true
}

// sortedBoundOrders 는 결속 주문을 정해진 순서로 늘어놓음 — 요청과 판정이 같은 순서로 짝지어야 함.
func sortedBoundOrders(b terminalBinding) []boundOrder {
	out := make([]boundOrder, 0, len(b.orders))
	for order := range b.orders {
		out = append(out, order)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].identity != out[j].identity {
			return out[i].identity.less(out[j].identity)
		}
		return out[i].original < out[j].original
	})
	return out
}

// orderTerminalRequest 는 결속으로 장부 증명 요청을 만듦. 스냅숏은 감지기가 쓴 원래 번호로 찾음.
func orderTerminalRequest(b terminalBinding) journal.OrderTerminalProofRequest {
	orders := sortedBoundOrders(b)
	req := journal.OrderTerminalProofRequest{
		StateID: b.rowID, Symbol: b.symbol, FillWatermark: b.fillWatermark,
		Orders: make([]journal.FillSnapshotScope, 0, len(orders)),
	}
	for _, order := range orders {
		req.Orders = append(req.Orders, journal.FillSnapshotScope{
			OrderID: order.original, AccountRef: order.identity.AccountRef, Market: order.identity.Market,
			TradingDay: order.identity.TradingDay, Symbol: order.identity.Symbol, Side: order.identity.Side,
		})
	}
	return req
}

// orderTerminalComparison 은 이 비교에서 판정이 쓰는 값만 얼림.
func orderTerminalComparison(trackerAccount string, diff Diff, disputed bool) terminalComparison {
	external := make(map[string]bool, len(diff.ExternalOrd))
	for _, order := range diff.ExternalOrd {
		// 바이트 그대로 — 주문번호는 불투명 식별자라 정규화하지 않음.
		external[order.OrderID] = true
	}
	return terminalComparison{
		trackerAccount: strings.TrimSpace(trackerAccount), diffAccount: strings.TrimSpace(diff.AccountRef),
		asOf: diff.AsOf, disputed: disputed, blocksEntry: diff.BlocksEntry(), externalIDs: external,
	}
}

// orderTerminalReleaseVerdict 는 종결 증명 해제의 유일한 판정임(design D3 C1~C13).
//
// 처음 실패한 조건의 사유로 「유지」를 돌려줌. 모두 통과하면 무장이 없거나 증명이 바뀐 경우 「무장」,
// 같은 증명으로 무장보다 나중 비교면 「해제」.
func orderTerminalReleaseVerdict(b terminalBinding, cmp terminalComparison, proof journal.OrderTerminalProof) terminalVerdict {
	hold := func(reason string) terminalVerdict { return terminalVerdict{outcome: terminalHold, reason: reason} }

	// C1 — 이 프로세스의 삽입 증명이 만든 결속인가.
	if strings.TrimSpace(b.rowID) == "" {
		return hold("not_inserted_by_this_process")
	}
	// 읽기 실패는 판정 불가 — 유지(M5).
	if proof.ReadErr != nil {
		return hold("terminal_read_failed")
	}
	// C2 — 활성 행이 그 id 그대로이고 추적기 종목 범위 행인가.
	state := proof.State
	if !proof.StateFound || !state.Active() || state.ID != b.rowID ||
		strings.ToUpper(strings.TrimSpace(state.Symbol)) != b.symbol ||
		state.Cause != journal.ReconcileCauseQuantityMismatch || strings.TrimSpace(state.ScopeMarket) != "" ||
		cmp.trackerAccount == "" || strings.TrimSpace(state.AccountRef) != cmp.trackerAccount {
		return hold("row_changed")
	}
	// C3 — 수량 불일치·외부 포지션 흔적이 있으면 그 행은 조정 규칙만 풂(I2).
	if b.tainted {
		return hold("quantity_dispute_seen")
	}
	// C4 — 결속 주문이 있고 신원이 완전하며 계좌·종목이 모두 같은가.
	orders := sortedBoundOrders(b)
	if len(orders) == 0 {
		// 빈 표본의 전칭 판정은 자동 참이 되므로 따로 막음.
		return hold("identity_incomplete")
	}
	for _, order := range orders {
		id := order.identity
		if id.AccountRef == "" || id.Market == "" || id.TradingDay == "" || id.Symbol == "" || id.Side == "" ||
			strings.TrimSpace(id.OrderID) == "" {
			return hold("identity_incomplete")
		}
		if id.AccountRef != cmp.trackerAccount || cmp.diffAccount != cmp.trackerAccount || id.Symbol != b.symbol {
			return hold("account_mismatch")
		}
	}
	// 증명이 요청과 같은 순서·범위로 왔는지 — 짝이 어긋나면 판정하지 않음.
	if len(proof.Snapshots) != len(orders) {
		return hold("proof_mismatch")
	}
	latest := time.Time{}
	latestText := ""
	for i, order := range orders {
		snapshot := proof.Snapshots[i]
		if snapshot.Scope.OrderID != order.original {
			return hold("proof_mismatch")
		}
		// C5 — 정정 lineage 이동 없음(현재 번호 = 원래 번호, 대체 흔적 없음).
		if order.identity.OrderID != order.original || snapshot.Amended {
			return hold("lineage_moved")
		}
		// C6 — 스냅숏이 있는가.
		if !snapshot.Found {
			return hold("terminal_not_found")
		}
		// C7 — 깨끗한 종결인가(실패 갱신은 terminal 을 두고 fail_closed 만 세움).
		if !snapshot.Record.Terminal || snapshot.Record.FailClosed {
			return hold("not_clean_terminal")
		}
		committed, err := time.Parse(time.RFC3339, strings.TrimSpace(snapshot.Record.CommittedAt))
		if err != nil {
			return hold("comparison_not_after_terminal")
		}
		if committed.After(latest) {
			latest, latestText = committed, snapshot.Record.CommittedAt
		}
	}
	// C8 — 비교 as-of 가 모든 종결 기록보다 엄격히 나중인가.
	asOf, err := time.Parse(time.RFC3339, strings.TrimSpace(cmp.asOf))
	if err != nil || !asOf.After(latest) {
		return hold("comparison_not_after_terminal")
	}
	// C9 — 이 비교가 그 종목에 차이를 보지 않는가.
	if cmp.disputed {
		return hold("symbol_disputed")
	}
	// C10 — 장부는 종결인데 브로커 open 목록에 같은 번호가 남아 있으면 모순 — 유지(M4).
	for _, order := range orders {
		if cmp.externalIDs[order.identity.OrderID] || cmp.externalIDs[order.original] {
			return hold("terminal_order_still_open")
		}
	}
	// C11 — 비교 전체가 진입을 막지 않는가(호출 자리가 보장하지만 판정도 확인함).
	if cmp.blocksEntry {
		return hold("comparison_blocks_entry")
	}
	// C12 — 행 삽입 전 워터마크 뒤에 그 종목 체결이 커밋됐으면 유지(행 뒤의 투영 거절을 가리는 길 — M11).
	if proof.FillAfterWatermark {
		return hold("fill_after_block")
	}
	// C13 — 무장-후-해제(M12). digest 가 바뀌면 다시 무장함.
	digest := orderTerminalDigest(b.rowID, orders, proof.Snapshots)
	verdict := terminalVerdict{digest: digest, asOf: asOf, boundOrders: len(orders), latestCommit: latestText}
	if !b.armed || b.armDigest != digest {
		verdict.outcome, verdict.reason = terminalArm, "armed"
		return verdict
	}
	if !asOf.After(b.armAsOf) {
		verdict.outcome, verdict.reason = terminalHold, "arm_not_later"
		return verdict
	}
	verdict.outcome, verdict.reason = terminalReleaseOutcome, "released"
	return verdict
}

// orderTerminalDigest 는 무장이 묶이는 증명의 지문 — 행 id, 결속 신원, 결속 주문의 스냅숏 상태.
func orderTerminalDigest(rowID string, orders []boundOrder, snapshots []journal.OrderTerminalSnapshot) string {
	h := sha256.New()
	write := func(parts ...string) {
		for _, part := range parts {
			fmt.Fprintf(h, "%d:%s|", len(part), part)
		}
	}
	write(rowID)
	for i, order := range orders {
		id := order.identity
		write(id.AccountRef, id.Market, id.TradingDay, id.Symbol, id.Side, id.OrderID, order.original)
		record := snapshots[i].Record
		write(record.State, fmt.Sprint(record.Terminal), fmt.Sprint(record.FailClosed),
			record.FilledQuantity, record.CommittedAt)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// beginOrderTerminalObservation 은 Observe 가 갈래를 나누기 전에 부름: 이번 관측 시작 때 있던 블록 키를
// 적고(이번에 처음 만들어진 블록만 결속), 기존 결속에 이번 비교의 신원·흔적을 더함. 호출자가 t.mu 를 쥠.
//
// 갈래를 나누기 전인 이유: 깨끗한 갈래(외부 포지션만 있는 비교)와 blocking 갈래 둘 다 영구 보류 권위 읽기
// 실패로 일찍 돌아갈 수 있고, 그 전에 흔적을 남겨야 실격이 사라지지 않음.
func (t *Tracker) beginOrderTerminalObservation(diff Diff) {
	if t.Terminal == nil {
		return
	}
	t.terminal.seen = make(map[string]bool, len(t.blocks))
	for key := range t.blocks {
		t.terminal.seen[key] = true
	}
	t.updateOrderTerminalBindings(diff)
}

// updateOrderTerminalBindings 는 결속이 있는 종목마다 이번 비교의 missing-order 신원을 더하고, 수량
// 불일치·외부 포지션이 보이면 실격 흔적을 세움. 멱등 — persist 뒤에 한 번 더 불러 같은 비교에서 처음 생긴
// 결속에도 같은 종목의 다른 missing-order 를 더함. 호출자가 t.mu 를 쥠.
func (t *Tracker) updateOrderTerminalBindings(diff Diff) {
	if t.Terminal == nil || len(t.terminal.bindings) == 0 {
		return
	}
	// 그 종목의 블록이 다른 길(조정 해제·운영자 해제 등)로 사라졌으면 결속도 버림 — 남은 사유가 운영자를 속이지
	// 않게 하고, 같은 종목에 새 행이 생겨도 옛 결속을 쓰지 않게 함(코드 리뷰 2 #1).
	for symbol := range t.terminal.bindings {
		if !hasBlockFor(t.blocks, symbol) {
			delete(t.terminal.bindings, symbol)
		}
	}
	for _, missing := range diff.MissingOrders {
		if b := t.terminal.bindings[strings.ToUpper(strings.TrimSpace(missing.Symbol))]; b != nil {
			b.orders[boundOrder{identity: missing.Identity(), original: missing.OriginalOrderID}] = struct{}{}
		}
	}
	for _, mismatch := range diff.Quantities {
		if b := t.terminal.bindings[strings.ToUpper(strings.TrimSpace(mismatch.Symbol))]; b != nil {
			b.tainted = true
		}
	}
	for _, external := range diff.ExternalPos {
		if b := t.terminal.bindings[strings.ToUpper(strings.TrimSpace(external.Symbol))]; b != nil {
			b.tainted = true
		}
	}
}

// enterReconcileWithBinding 은 persist 의 `EnterReconcile` 호출을 감쌈: 이번 관측에서 처음 만들어진
// missing-order 블록이면 삽입 전에 체결 워터마크를 읽고, 장부가 「삽입했다」(entered)고 답할 때만 결속을
// 만듦. 장부 쓰기 자체는 그대로 — Terminal 이 nil 이면 `EnterReconcile` 한 번과 같음. 호출자가 t.mu 를 쥠.
func (t *Tracker) enterReconcileWithBinding(ctx context.Context, block Block,
	req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	bindable := t.Terminal != nil && block.missing != nil && !t.terminal.seen[block.Key()]
	var watermark int64
	if bindable {
		w, err := t.Terminal.MaxFillEventID(ctx)
		if err != nil {
			// 워터마크가 없으면 C12 를 못 따지므로 결속하지 않음 — 유지 쪽.
			bindable = false
		}
		watermark = w
	}
	state, entered, err := t.Journal.EnterReconcile(ctx, req)
	if err == nil && entered && bindable && strings.TrimSpace(state.ID) != "" {
		if t.terminal.bindings == nil {
			t.terminal.bindings = map[string]*terminalBinding{}
		}
		symbol := strings.ToUpper(strings.TrimSpace(block.Symbol))
		t.terminal.bindings[symbol] = &terminalBinding{
			rowID: state.ID, symbol: symbol, fillWatermark: watermark,
			orders: map[boundOrder]struct{}{*block.missing: {}},
		}
	}
	return state, entered, err
}

// orderTerminalCandidate 는 Observe 깨끗한 갈래에서 크레딧이 없는 블록에 1차 판정을 물음. 「무장」이면 무장만
// 기억하고 거짓, 「해제」면 블록에 재판정 입력을 달고 참. 호출자가 t.mu 를 쥠.
func (t *Tracker) orderTerminalCandidate(ctx context.Context, block *Block, diff Diff, disputed map[string]bool) bool {
	if t.Terminal == nil || t.Journal == nil {
		return false
	}
	symbol := strings.ToUpper(strings.TrimSpace(block.Symbol))
	b := t.terminal.bindings[symbol]
	if b == nil {
		return false
	}
	request := orderTerminalRequest(*b)
	comparison := orderTerminalComparison(t.AccountRef, diff, disputed[symbol])
	verdict := orderTerminalReleaseVerdict(*b, comparison, t.Terminal.ReadOrderTerminalProof(ctx, request))
	b.reason = verdict.reason
	switch verdict.outcome {
	case terminalArm:
		b.armed, b.armAsOf, b.armDigest = true, verdict.asOf, verdict.digest
		return false
	case terminalReleaseOutcome:
		block.terminalRelease = &terminalRelease{
			binding: *b, comparison: comparison, request: request,
			boundOrders: verdict.boundOrders, latestCommit: verdict.latestCommit,
		}
		return true
	default:
		if verdict.reason == "row_changed" {
			// 그 행이 다른 길로 풀렸거나 바뀜 — 이 결속은 더 쓸 데가 없음.
			delete(t.terminal.bindings, symbol)
		}
		return false
	}
}

// releaseOnOrderTerminal 은 persist 가 종결 증명 블록을 풀 때 부름: 같은 판정을 해제 tx 안에서 다시 통과할
// 때만 장부가 해제함. 해제되면 결속을 지움. 호출자가 t.mu 를 쥠 — 콜백은 순수 함수라 잠금을 다시 잡지 않음.
func (t *Tracker) releaseOnOrderTerminal(ctx context.Context, block Block) (bool, error) {
	release := block.terminalRelease
	evidence := fmt.Sprintf(
		"missing-order terminal proof — %d bound order(s) terminal, latest terminal commit %s, comparison as-of %s",
		release.boundOrders, release.latestCommit, release.comparison.asOf)
	binding, comparison := release.binding, release.comparison
	// 콜백은 순수 함수 — 판정 결과만 지역 변수에 남기고 잠금·장부를 다시 부르지 않음.
	reason := ""
	_, released, err := t.Terminal.ReleaseReconcileOnOrderTerminalProof(ctx, release.request, evidence,
		func(proof journal.OrderTerminalProof) bool {
			verdict := orderTerminalReleaseVerdict(binding, comparison, proof)
			reason = verdict.reason
			return verdict.outcome == terminalReleaseOutcome
		})
	if err != nil {
		return false, err
	}
	if released {
		delete(t.terminal.bindings, binding.symbol)
	} else if b := t.terminal.bindings[binding.symbol]; b != nil {
		// tx 안 재판정이 유지 — 그 사유를 남김(M5).
		b.reason = "tx_" + reason
	}
	return released, nil
}

// OrderTerminalHoldReasons 는 지금 결속이 있는 종목마다 마지막 판정 사유를 돌려줌(M5 「유지 + 사유」 — 로그용).
// 계좌·주문번호·수량을 담지 않음.
func (t *Tracker) OrderTerminalHoldReasons() map[string]string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make(map[string]string, len(t.terminal.bindings))
	for symbol, b := range t.terminal.bindings {
		if b.reason != "" {
			out[symbol] = b.reason
		}
	}
	return out
}
