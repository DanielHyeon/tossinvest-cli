package filldetect

// closed_lookup.go 는 a191 — 미체결(OPEN) 목록을 떠났는데 단건 조회가 「살아 있음」을 돌려주는 추적 주문의 종결을
// 종결(CLOSED) 목록으로 확인함.
//
// # 왜 필요한가
//
// 토스 단건 조회는 KR 주문을 앱에서 취소해도 옛 주문번호를 계속 PENDING(취소 시각 없음)으로 돌려줌
// (verify-execution-capability measurements M32). 그러면 collect 는 그 주문을 매 주기 「열림·미체결」로 읽고 종결을
// 쓰지 않아, 추적 집합과 대사의 열린 주문에 영원히 남음(missing-order → 계좌 영구 차단).
//
// # 받아들이는 경우는 하나뿐임 (design D3)
//
// 같은 주문번호(바이트 동일)·같은 정규 범위 행이 정확히 하나이고, 그 행에 체결 수량이 실제로 있고(M6), 그 행의 파생이
// 깨끗한 종결이며, 주문·체결 수량이 단건과 같을 때만. 그때도 바꾸는 것은 **상태 필드뿐**임 — 수량·평균가·금액·체결 시각은
// 단건 것을 유지함(종결 목록 혼자서 체결을 만들거나 지우지 않음).
//
// # 그 밖은 오늘과 같음 (design D4)
//
// 못 찾음·읽기 실패·모호 등 어느 경우든 단건 스냅숏을 그대로 쓰고, 주기를 실패시키지 않으며, 새 refusal·래치를 만들지
// 않음. 사유 코드만 남기고 30s 이상 이어지면 한 번 알림(D6).
//
// # 조회량과 시간 (design D5, Manager M8)
//
// 주기당 조회 최대 1건(안 물은 것 먼저, 그다음 lastLookup 이 가장 오래된 것), 같은 주문은 30s 간격, 조회 전체에 2s
// 취소 시한. 간격·보고 판정은 주입 Clock 으로 함 — context 시한은 매달린 전송을 끊는 용도로만 씀.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// Alerter 는 운영자 알림 통로임. 운영 배선은 *obs.Notifier.
type Alerter interface {
	Notify(ctx context.Context, e obs.Event) error
}

// EventFillOrderUnresolved 는 목록을 떠난 추적 주문이 종결 목록으로 확인되지 않은 채 이어짐을 알림.
// a129 선례대로 internal/obs 를 고치지 않고 지역 리터럴로 둠 — 등급표에 없으므로 normal(전달 보장 없음).
const EventFillOrderUnresolved obs.EventType = "fill.order_unresolved"

// UnresolvedReason 은 종결 목록이 그 주기에 종결을 확인하지 못한 사유임. 안정 문자열(경보에 실림).
type UnresolvedReason string

const (
	UnresolvedClosedListNotFound          UnresolvedReason = "closed_list_not_found"
	UnresolvedClosedListReadFailed        UnresolvedReason = "closed_list_read_failed"
	UnresolvedClosedListTruncated         UnresolvedReason = "closed_list_truncated"
	UnresolvedClosedListRowUnreadable     UnresolvedReason = "closed_list_row_unreadable"
	UnresolvedClosedListAmbiguous         UnresolvedReason = "closed_list_ambiguous"
	UnresolvedClosedListScopeMismatch     UnresolvedReason = "closed_list_scope_mismatch"
	UnresolvedClosedListExecutionMissing  UnresolvedReason = "closed_list_execution_missing"
	UnresolvedClosedListNotCleanTerminal  UnresolvedReason = "closed_list_not_clean_terminal"
	UnresolvedClosedListQuantityDisagrees UnresolvedReason = "closed_list_quantity_disagrees"
	UnresolvedClosedLookupWindowInvalid   UnresolvedReason = "closed_lookup_window_invalid"
	UnresolvedClosedLookupDeferred        UnresolvedReason = "closed_lookup_deferred"
)

// UnresolvedOrder 는 그 주기에 종결 목록으로 확인되지 않은 후보 하나임(주기 결과용 — 경보에는 OrderID 를 싣지 않음).
type UnresolvedOrder struct {
	OrderID  string
	IntentID string
	Market   string
	Reason   UnresolvedReason
}

const (
	statusClosed          = "CLOSED"
	closedLookupLimit     = 100
	closedLookupMaxPages  = 3
	closedLookupDeadline  = 2 * time.Second
	closedLookupInterval  = 30 * time.Second
	unresolvedReportAfter = 30 * time.Second
	closedTradingDay      = "2006-01-02"
)

// leftOpenTrack 은 후보 키 하나의 간격·보고 상태임. poll 뮤텍스 아래에서만 읽고 씀.
type leftOpenTrack struct {
	firstSeen  time.Time
	lastLookup time.Time
	lastReason UnresolvedReason
	reported   bool
}

// leftOpenCandidate 는 한 주기의 후보임. cycle 포인터가 그 주기의 신원이라, 중간에 실패한 주기의 후보는 다음 주기에 버려짐.
type leftOpenCandidate struct {
	cycle *Cycle
	key   canonicalOrderKey
	order TrackedOrder
	index int
}

// noteLeftOpen 은 collect 가 OPEN 에서 못 본 추적 주문의 단건 스냅숏을 붙이기 직전에 부름.
// 후보는 단건 파생이 비종결이면서 fail-closed 가 아닐 때뿐임(M2 = 제외 — UNKNOWN 은 기존 refusal 경로 그대로).
func (d *Detector) noteLeftOpen(cycle *Cycle, key canonicalOrderKey, order TrackedOrder, index int, snap Snapshot) {
	if snap.Derived.Terminal || snap.Derived.FailClosed {
		return
	}
	// 앞 주기가 settle 전에 오류로 끝났으면 그 후보가 남아 있음 — 이번 주기 것만 남겨 쌓이지 않게 함(코드 리뷰 1 #1).
	if len(d.leftOpenCycle) > 0 && d.leftOpenCycle[0].cycle != cycle {
		d.leftOpenCycle = d.leftOpenCycle[:0]
	}
	d.leftOpenCycle = append(d.leftOpenCycle, leftOpenCandidate{cycle: cycle, key: key, order: order, index: index})
}

// settleLeftOpen 은 collect 의 끝에서 부름 — 이번 주기 후보 가운데 하나만 종결 목록에 묻고, 받아들이면 그 스냅숏의 상태
// 필드를 바꾸며, 간격 상태를 정리하고, 오래 이어진 미해결을 한 번 알림.
func (d *Detector) settleLeftOpen(ctx context.Context, cycle *Cycle, snaps []Snapshot) []Snapshot {
	candidates := make([]leftOpenCandidate, 0, len(d.leftOpenCycle))
	for _, c := range d.leftOpenCycle {
		if c.cycle == cycle {
			candidates = append(candidates, c)
		}
	}
	d.leftOpenCycle = nil

	now := d.clk().Now()
	if d.leftOpen == nil {
		d.leftOpen = map[canonicalOrderKey]*leftOpenTrack{}
	}
	current := make(map[canonicalOrderKey]bool, len(candidates))
	for _, c := range candidates {
		current[c.key] = true
		if _, known := d.leftOpen[c.key]; !known {
			d.leftOpen[c.key] = &leftOpenTrack{firstSeen: now}
		}
	}
	// 이번 주기 후보가 아닌 키는 해결된 것임(OPEN 재등장·종결·추적 종료) — 상태를 지워 다시 생기면 처음부터 셈.
	for key := range d.leftOpen {
		if !current[key] {
			delete(d.leftOpen, key)
		}
	}

	chosen := pickClosedLookup(candidates, d.leftOpen, now)
	for i, c := range candidates {
		track := d.leftOpen[c.key]
		reason := UnresolvedClosedLookupDeferred
		if i == chosen {
			// 결과와 상관없이 물은 시각을 남김 — 실패해도 같은 주기 재시도는 없고 간격 뒤에 다시 물음.
			track.lastLookup = now
			cycle.ClosedReads++
			accepted, why := d.lookupClosed(ctx, c, snaps[c.index])
			if why == "" {
				// 상태는 지우지 않음 — 이 스냅숏이 원장에 적용되기 전에 주기가 실패하면 다음 주기에도 후보로 남는데,
				// 그때 30s 간격을 지켜야 함(코드 리뷰 1 #2). 다음 주기에 후보가 아니면 위 정리가 지움.
				snaps[c.index] = accepted
				continue
			}
			reason = why
			track.lastReason = why
		}
		cycle.Unresolved = append(cycle.Unresolved, UnresolvedOrder{
			OrderID: c.order.OrderID, IntentID: c.order.IntentID, Market: c.key.Market, Reason: reason,
		})
		if !track.reported && now.Sub(track.firstSeen) >= unresolvedReportAfter {
			track.reported = true
			reported := track.lastReason
			if reported == "" {
				reported = reason
			}
			d.reportUnresolved(ctx, c, reported, now.Sub(track.firstSeen))
		}
	}
	return snaps
}

// pickClosedLookup 은 이번 주기에 물을 후보 하나의 위치를 고름(없으면 -1).
// 간격이 지난 것만 대상이고, 한 번도 안 물은 것이 먼저, 그다음 lastLookup 이 가장 오래된 것, 같으면 추적 순서.
func pickClosedLookup(candidates []leftOpenCandidate, tracks map[canonicalOrderKey]*leftOpenTrack, now time.Time) int {
	chosen := -1
	var chosenAt time.Time
	for i, c := range candidates {
		last := tracks[c.key].lastLookup
		if !last.IsZero() && now.Sub(last) < closedLookupInterval {
			continue
		}
		if chosen == -1 || (!chosenAt.IsZero() && (last.IsZero() || last.Before(chosenAt))) {
			chosen, chosenAt = i, last
		}
	}
	return chosen
}

// lookupClosed 는 후보 하나를 종결 목록에서 찾아 수용 조건(design D3)을 판정함. 받아들이면 상태 필드만 바꾼 단건
// 스냅숏과 빈 사유를, 아니면 사유를 돌려줌.
func (d *Detector) lookupClosed(ctx context.Context, c leftOpenCandidate, single Snapshot) (Snapshot, UnresolvedReason) {
	day, err := time.Parse(closedTradingDay, c.key.TradingDay)
	if err != nil {
		return Snapshot{}, UnresolvedClosedLookupWindowInvalid
	}
	// from/to 는 KST 주문 생성일이고 거래일은 시장 현지 날짜라, 앞뒤 하루로 둘 다 덮음.
	query := execgw.OrderQuery{
		Status: statusClosed,
		Symbol: c.key.Symbol,
		From:   day.AddDate(0, 0, -1).Format(closedTradingDay),
		To:     day.AddDate(0, 0, 1).Format(closedTradingDay),
		Limit:  closedLookupLimit,
	}
	lookupCtx, cancel := context.WithTimeout(ctx, closedLookupDeadline)
	raws, err := execgw.ScanOrders(lookupCtx, d.Orders, query, closedLookupMaxPages)
	cancel()
	if err != nil {
		if errors.Is(err, execgw.ErrTooManyPages) || errors.Is(err, execgw.ErrCursorLoop) {
			return Snapshot{}, UnresolvedClosedListTruncated
		}
		return Snapshot{}, UnresolvedClosedListReadFailed
	}

	row, why := matchClosedRow(raws, c.order.OrderID)
	if why != "" {
		return Snapshot{}, why
	}
	rowSnap, err := parseSnapshot(row, single.ObservedAt, single.BrokerVisibleAt)
	if err != nil {
		// 체결 수량이 없거나 못 읽혀 파싱이 실패한 것이면 M6 사유로 가름(코드 리뷰 1 #3).
		if !closedRowCarriesFill(row) {
			return Snapshot{}, UnresolvedClosedListExecutionMissing
		}
		return Snapshot{}, UnresolvedClosedListRowUnreadable
	}
	rowSnap.AccountRef = single.AccountRef
	if actual, complete := snapshotOrderKey(rowSnap); !complete || actual != c.key {
		return Snapshot{}, UnresolvedClosedListScopeMismatch
	}
	if !closedRowCarriesFill(row) {
		return Snapshot{}, UnresolvedClosedListExecutionMissing
	}
	derived := brokerstate.Derive(viewOf(rowSnap, c.order.Lineage))
	if !derived.Terminal || derived.FailClosed {
		return Snapshot{}, UnresolvedClosedListNotCleanTerminal
	}
	if !closedQuantityEqual(rowSnap.Quantity, single.Quantity) ||
		!closedQuantityEqual(rowSnap.FilledQuantity, single.FilledQuantity) {
		return Snapshot{}, UnresolvedClosedListQuantityDisagrees
	}

	// 상태 필드만 교체 — 수량·평균가·금액·체결 시각은 단건 것(원장 delta 는 직전 스냅숏 대비라 가격 출처가 중요함).
	accepted := single
	accepted.RawStatus = rowSnap.RawStatus
	accepted.Canceled = rowSnap.Canceled
	accepted.CanceledAt = rowSnap.CanceledAt
	accepted.Derived = derived
	return accepted, ""
}

// matchClosedRow 는 주문번호가 바이트 동일한 행 하나를 고름. raw 바이트가 같은 중복(페이지 경계)은 하나로 세고, 바이트가
// 다른 같은 번호 행이 있으면 모호함. 번호를 못 읽는 행이 하나라도 있으면 결론 불가.
func matchClosedRow(raws []json.RawMessage, orderID string) (json.RawMessage, UnresolvedReason) {
	var match json.RawMessage
	for _, raw := range raws {
		var head struct {
			OrderID *string `json:"orderId"`
		}
		if err := json.Unmarshal(unwrapResult(raw), &head); err != nil || head.OrderID == nil {
			return nil, UnresolvedClosedListRowUnreadable
		}
		if *head.OrderID != orderID {
			continue
		}
		if match == nil {
			match = raw
			continue
		}
		if !bytes.Equal(match, raw) {
			return nil, UnresolvedClosedListAmbiguous
		}
	}
	if match == nil {
		return nil, UnresolvedClosedListNotFound
	}
	return match, ""
}

// closedRowCarriesFill 은 M6 — 행에 execution.filledQuantity 가 있고 유한한 십진수로 읽히는지 봄(값 0 허용).
// 없으면 parseSnapshot 이 체결 0 으로 읽어 수량 일치가 표본 없이 참이 되므로 받지 않음. Q9 결과로 뒤집힐 수 있어 따로 둠.
func closedRowCarriesFill(raw json.RawMessage) bool {
	var row struct {
		Execution *struct {
			FilledQuantity *string `json:"filledQuantity"`
		} `json:"execution"`
	}
	if err := json.Unmarshal(unwrapResult(raw), &row); err != nil || row.Execution == nil || row.Execution.FilledQuantity == nil {
		return false
	}
	text := strings.TrimSpace(*row.Execution.FilledQuantity)
	if text == "" {
		return false
	}
	v, err := strconv.ParseFloat(text, 64)
	return err == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}

// closedQuantityEqual 은 brokerstate 와 같은 상대 허용(1e-9)으로 두 수량을 비교함.
func closedQuantityEqual(a, b float64) bool {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= 1e-9*scale
}

// reportUnresolved 는 미해결이 이어진 후보를 한 번 알림. 브로커 주문번호·계좌·종목·수량·가격은 싣지 않음(Manager M1).
// 조회와 별개의 2s context 로 감쌈(조회 시한이 지난 context 를 재사용하지 않음).
func (d *Detector) reportUnresolved(ctx context.Context, c leftOpenCandidate, reason UnresolvedReason, lasted time.Duration) {
	if d.Alerts == nil {
		return
	}
	notifyCtx, cancel := context.WithTimeout(ctx, closedLookupDeadline)
	defer cancel()
	_ = d.Alerts.Notify(notifyCtx, obs.Event{
		Type:  EventFillOrderUnresolved,
		Title: "체결 감지: 미체결 목록을 떠난 주문의 종결을 확인하지 못함",
		Body:  "단건 조회는 살아 있다고 하고 종결 목록으로도 확인되지 않음 — 대사 차단이 이어질 수 있음",
		Fields: map[string]any{
			obs.FieldIntentID:    c.order.IntentID,
			obs.FieldMarket:      c.key.Market,
			obs.FieldReason:      string(reason),
			"unresolved_seconds": int64(lasted / time.Second),
		},
	})
}
