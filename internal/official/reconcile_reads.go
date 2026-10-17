package official

// reconcile_reads.go 는 a121(reconcile-stale-verification-artifacts) 대사 전용 읽기의 **골격**이다.
//
// 같은 endpoint·같은 getAcct/get 경로를 쓰되 result 원문을 받아 스키마 존재·형을 따로 검증한다(tasks 3.2).
//
// # 왜 기존 읽기(ProtectionConditionalOrdersRaw · OrdersPageRaw)를 그대로 쓰지 않는가
//
// design G1-6 codex F2·R2-1 — 기존 경로는 unwrapAndDecode 로 구조체에 바로 풀어서 `{"result":null}`·`{}`·
// null 값 필드를 무오류 빈 페이지로 접는다(client.go:213-227). 대사는 "빈 목록" 을 부재의 근거로 쓰므로
// 결측·null·형 불일치를 따로 거절해야 하고, 그러려면 result 원문 바이트를 받아야 한다. 또 codex F6 — 채택
// 리더 adaptProtectionConditional 은 Second 를 버린다(protection_reads.go:49-57). 같은 endpoint·같은
// getAcct 경로를 쓰되 해독만 새로 한다. 기존 함수는 편집하지 않는다.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ErrReconcileSchema 는 대사 읽기 응답이 스키마 존재·형 검증을 통과하지 못했음을 뜻함.
// 빈 목록으로 접지 않고 거절로 다루기 위한 표지임.
var ErrReconcileSchema = errors.New("official: reconcile read failed schema validation")

// ReconcileConditionalRow 는 대사 판정에 쓰는 조건주문 한 행임.
// HasSecond 는 응답의 second 가 null 이 아닌지(OCO/OTO 다리 존재)를 그대로 노출함(codex F6).
type ReconcileConditionalRow struct {
	ID               string
	Symbol           string
	Market           string
	Status           string
	TriggeredOrderID string
	HasSecond        bool
}

// ReconcileConditionalPage 는 조건주문 목록 한 페이지임.
type ReconcileConditionalPage struct {
	Rows       []ReconcileConditionalRow
	NextCursor string
	HasNext    bool
}

// ReconcileOrderRow 는 대사 판정에 쓰는 일반 주문 한 행임(일반 주문 응답에는 market 이 없음).
type ReconcileOrderRow struct {
	ID     string
	Symbol string
	Status string
}

// ReconcileOrderPage 는 일반 주문 목록 한 페이지임.
type ReconcileOrderPage struct {
	Rows       []ReconcileOrderRow
	NextCursor string
	HasNext    bool
}

// reconcilePageEnvelope 는 목록 페이지의 공통 검증 결과임 — 컬렉션 원문 행들과 페이지 경계.
type reconcilePageEnvelope struct {
	rows       []json.RawMessage
	nextCursor string
	hasNext    bool
}

func reconcileSchema(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrReconcileSchema}, args...)...)
}

// decodeReconcileEnvelope 는 result 가 객체이고, collection 키가 non-null 배열이며, hasNext 가 non-null 불리언이고,
// nextCursor 가 존재하며 null 또는 문자열이고, hasNext 인데 커서가 비지 않았음을 검증함(codex F2·R2-1).
// null 값은 Go 제로값으로 접혀 "빈 목록" 이 되므로 해독 전에 원문에서 가른다.
func decodeReconcileEnvelope(result []byte, collection string) (reconcilePageEnvelope, error) {
	trimmed := bytes.TrimSpace(result)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return reconcilePageEnvelope{}, reconcileSchema("result is not an object (%.20s)", trimmed)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return reconcilePageEnvelope{}, reconcileSchema("result does not decode: %v", err)
	}
	rawRows, ok := fields[collection]
	if !ok {
		return reconcilePageEnvelope{}, reconcileSchema("%s is missing", collection)
	}
	if t := bytes.TrimSpace(rawRows); len(t) == 0 || t[0] != '[' {
		return reconcilePageEnvelope{}, reconcileSchema("%s is not an array", collection)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(rawRows, &rows); err != nil {
		return reconcilePageEnvelope{}, reconcileSchema("%s does not decode: %v", collection, err)
	}
	for i, row := range rows {
		if t := bytes.TrimSpace(row); len(t) == 0 || t[0] != '{' {
			return reconcilePageEnvelope{}, reconcileSchema("%s[%d] is not an object", collection, i)
		}
	}
	rawHasNext, ok := fields["hasNext"]
	if !ok {
		return reconcilePageEnvelope{}, reconcileSchema("hasNext is missing")
	}
	var hasNext bool
	switch string(bytes.TrimSpace(rawHasNext)) {
	case "true":
		hasNext = true
	case "false":
	default:
		return reconcilePageEnvelope{}, reconcileSchema("hasNext is not a boolean (%.20s)", rawHasNext)
	}
	rawCursor, ok := fields["nextCursor"]
	if !ok {
		return reconcilePageEnvelope{}, reconcileSchema("nextCursor is missing")
	}
	var cursor string
	if t := bytes.TrimSpace(rawCursor); string(t) != "null" {
		if len(t) == 0 || t[0] != '"' {
			return reconcilePageEnvelope{}, reconcileSchema("nextCursor is neither null nor a string")
		}
		if err := json.Unmarshal(t, &cursor); err != nil {
			return reconcilePageEnvelope{}, reconcileSchema("nextCursor does not decode: %v", err)
		}
	}
	if hasNext && strings.TrimSpace(cursor) == "" {
		return reconcilePageEnvelope{}, reconcileSchema("hasNext is true but nextCursor is empty")
	}
	return reconcilePageEnvelope{rows: rows, nextCursor: cursor, hasNext: hasNext}, nil
}

// DecodeReconcileConditionalPage 는 조건주문 목록의 result 원문을 스키마 검증과 함께 해독함.
// second 키가 없거나 null 이면 비-OCO(RED 로트 처분 ④), 객체면 HasSecond.
func DecodeReconcileConditionalPage(result []byte) (ReconcileConditionalPage, error) {
	env, err := decodeReconcileEnvelope(result, "conditionalOrders")
	if err != nil {
		return ReconcileConditionalPage{}, err
	}
	page := ReconcileConditionalPage{Rows: make([]ReconcileConditionalRow, 0, len(env.rows)), NextCursor: env.nextCursor, HasNext: env.hasNext}
	for i, raw := range env.rows {
		var o apiConditionalOrder
		if err := json.Unmarshal(raw, &o); err != nil {
			return ReconcileConditionalPage{}, reconcileSchema("conditionalOrders[%d] does not decode: %v", i, err)
		}
		page.Rows = append(page.Rows, ReconcileConditionalRow{
			ID: o.ConditionalOrderID, Symbol: o.Symbol, Market: o.Market, Status: o.Status,
			TriggeredOrderID: o.First.TriggeredOrderID, HasSecond: o.Second != nil,
		})
	}
	return page, nil
}

// DecodeReconcileOrderPage 는 일반 주문 목록의 result 원문을 스키마 검증과 함께 해독함.
func DecodeReconcileOrderPage(result []byte) (ReconcileOrderPage, error) {
	env, err := decodeReconcileEnvelope(result, "orders")
	if err != nil {
		return ReconcileOrderPage{}, err
	}
	page := ReconcileOrderPage{Rows: make([]ReconcileOrderRow, 0, len(env.rows)), NextCursor: env.nextCursor, HasNext: env.hasNext}
	for i, raw := range env.rows {
		var o struct {
			OrderID string `json:"orderId"`
			Symbol  string `json:"symbol"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal(raw, &o); err != nil {
			return ReconcileOrderPage{}, reconcileSchema("orders[%d] does not decode: %v", i, err)
		}
		page.Rows = append(page.Rows, ReconcileOrderRow{ID: o.OrderID, Symbol: o.Symbol, Status: o.Status})
	}
	return page, nil
}

func reconcileListQuery(status, symbol, cursor string, limit int) url.Values {
	q := url.Values{}
	if status != "" {
		q.Set("status", status)
	}
	if symbol != "" {
		q.Set("symbol", symbol)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	return q
}

// ReconcileConditionalOrdersPage 는 GET /api/v1/conditional-orders 한 페이지를 대사용으로 읽음(계좌 헤더 포함).
func (c *Client) ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (ReconcileConditionalPage, error) {
	if strings.TrimSpace(status) == "" {
		return ReconcileConditionalPage{}, fmt.Errorf("%w: the reconcile read names its group", ErrOrderStatusRequired)
	}
	var raw json.RawMessage
	if err := c.getAcct(ctx, "/api/v1/conditional-orders", reconcileListQuery(status, symbol, cursor, limit), &raw); err != nil {
		return ReconcileConditionalPage{}, err
	}
	return DecodeReconcileConditionalPage(raw)
}

// ReconcileOpenOrdersPage 는 GET /api/v1/orders?status=OPEN 한 페이지를 대사용으로 읽음(계좌 헤더 포함).
func (c *Client) ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (ReconcileOrderPage, error) {
	var raw json.RawMessage
	if err := c.getAcct(ctx, "/api/v1/orders", reconcileListQuery("OPEN", symbol, cursor, limit), &raw); err != nil {
		return ReconcileOrderPage{}, err
	}
	return DecodeReconcileOrderPage(raw)
}

// ReconcileInstrument 는 종목 조회 GET(양성 대조, design G1 P1-3)으로 응답이 되돌린 심볼을 돌려줌 — 정확히 한 항목.
func (c *Client) ReconcileInstrument(ctx context.Context, symbol string) (string, error) {
	q := url.Values{}
	q.Set("symbols", symbol)
	var raw []struct {
		Symbol string `json:"symbol"`
	}
	if err := c.get(ctx, "/api/v1/stocks", q, &raw); err != nil {
		return "", err
	}
	if len(raw) != 1 {
		return "", fmt.Errorf("%w: the instrument read returned %d items, want exactly one", ErrReconcileSchema, len(raw))
	}
	return raw[0].Symbol, nil
}
