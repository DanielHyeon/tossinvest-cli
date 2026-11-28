//go:build ignore

// decode_ambiguity 는 a168 tasks 0.3 의 측정기임 — 현재 HEAD 의 대사 읽기 네 개(Client.Reconcile*)가 모호한 응답을
// 어떻게 다루는지 httptest 서버로 잼(실 브로커·실계좌 호출 없음). 손으로 읽은 분기가 아니라 실행 결과를 영수증으로 남김.
//
// 실행(저장소 루트에서):
//
//	go run ./openspec/changes/a168-verify-read-decode-hardening/analysis/measure/decode_ambiguity.go
//
// 각 줄: 사례 id · 읽기 · 응답 상태 · 결과(ACCEPT + 읽힌 값 | REFUSE + 오류 분류). HOLE 은 「앞 값이 사라지고 뒤 값이
// 판정 입력이 된 채 받아들여짐」 또는 「200 아닌 2xx 를 받아들임」 을 뜻하는 측정기의 표시임(판정 아님 — 기대는 design 에 있음).
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type reply struct {
	status int
	body   string
}

const (
	emptyCond  = `{"conditionalOrders":[],"nextCursor":null,"hasNext":false}`
	firedRow   = `{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","type":"SINGLE","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"second":null}`
	firedCond  = `{"conditionalOrders":[` + firedRow + `],"nextCursor":null,"hasNext":false}`
	emptyOrder = `{"orders":[],"nextCursor":null,"hasNext":false}`
	openOrder  = `{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING"}],"nextCursor":null,"hasNext":false}`
	detail     = `{"conditionalOrderId":"CO-T","type":"SINGLE","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED","triggeredOrderId":null},"second":null}`
	detailFire = `{"conditionalOrderId":"CO-T","type":"SINGLE","status":"COMPLETED","symbol":"005930","market":"KR","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"second":null}`
)

func row(first string) string {
	return `{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"CANCELED","type":"SINGLE","first":` + first + `,"second":null}],"nextCursor":null,"hasNext":false}`
}

type kase struct {
	id, read string
	resp     reply
}

func main() {
	cases := []kase{
		// 봉투 층(unwrapAndDecode — apiEnvelope 구조체).
		{"E0-baseline-fired", "cond", reply{200, `{"result":` + firedCond + `}`}},
		{"E1-envelope-dup-result", "cond", reply{200, `{"result":` + firedCond + `,"result":` + emptyCond + `}`}},
		{"E2-envelope-fold-Result", "cond", reply{200, `{"result":` + firedCond + `,"Result":` + emptyCond + `}`}},
		{"E2b-envelope-lone-Result", "cond", reply{200, `{"Result":` + firedCond + `}`}},
		{"E3-envelope-trailing-object", "cond", reply{200, `{"result":` + emptyCond + `}{"x":1}`}},
		{"E3b-envelope-trailing-brace", "cond", reply{200, `{"result":` + emptyCond + `}}`}},
		{"E3c-envelope-trailing-bracket", "cond", reply{200, `{"result":` + emptyCond + `}]`}},
		// result 층(decodeReconcileEnvelope — map[string]json.RawMessage).
		{"R1-result-dup-collection", "cond", reply{200, `{"result":{"conditionalOrders":[` + firedRow + `],"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`}},
		{"R2-result-dup-hasNext", "cond", reply{200, `{"result":{"conditionalOrders":[],"nextCursor":"C2","hasNext":true,"hasNext":false}}`}},
		{"R3-result-fold-collection", "cond", reply{200, `{"result":{"conditionalOrders":[` + firedRow + `],"ConditionalOrders":[],"nextCursor":null,"hasNext":false}}`}},
		{"R4-order-result-dup-collection", "order", reply{200, `{"result":{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING"}],"orders":[],"nextCursor":null,"hasNext":false}}`}},
		// 행 층(DecodeReconcileConditionalPage — apiConditionalOrder 구조체).
		{"W1-row-first-dup-triggered", "cond", reply{200, `{"result":` + row(`{"status":"CANCELED","triggeredOrderId":"ORD-9","triggeredOrderId":null}`) + `}`}},
		{"W1b-row-first-dup-triggered-empty", "cond", reply{200, `{"result":` + row(`{"status":"CANCELED","triggeredOrderId":"ORD-9","triggeredOrderId":""}`) + `}`}},
		{"W1c-row-first-dup-triggered-nullfirst", "cond", reply{200, `{"result":` + row(`{"status":"CANCELED","triggeredOrderId":null,"triggeredOrderId":"ORD-9"}`) + `}`}},
		{"W2-row-first-fold-triggered", "cond", reply{200, `{"result":` + row(`{"status":"CANCELED","triggeredOrderId":"ORD-9","TriggeredOrderId":null}`) + `}`}},
		{"W2b-row-first-fold-triggered-empty", "cond", reply{200, `{"result":` + row(`{"status":"CANCELED","triggeredOrderId":"ORD-9","TriggeredOrderId":""}`) + `}`}},
		{"W3b-row-fold-status", "cond", reply{200, `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","Status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED"},"second":null}],"nextCursor":null,"hasNext":false}}`}},
		{"W8-row-deep-dup-ignored-field", "cond", reply{200, `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED","condition":{"price":"1","price":"2"}},"second":null}],"nextCursor":null,"hasNext":false}}`}},
		{"W3-row-dup-status", "cond", reply{200, `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED"},"second":null}],"nextCursor":null,"hasNext":false}}`}},
		{"W4-row-dup-second", "cond", reply{200, `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"EXPIRED","type":"OCO","first":{"status":"EXPIRED"},"second":{"status":"WATCHING"},"second":null}],"nextCursor":null,"hasNext":false}}`}},
		{"W5-row-dup-first", "cond", reply{200, `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"CANCELED","type":"SINGLE","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"first":{"status":"CANCELED"},"second":null}],"nextCursor":null,"hasNext":false}}`}},
		{"W6-order-row-dup-status", "order", reply{200, `{"result":{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING","status":"CANCELED"}],"nextCursor":null,"hasNext":false}}`}},
		{"W7-order-row-fold-symbol", "order", reply{200, `{"result":{"orders":[{"orderId":"O-1","symbol":"005930","Symbol":"000660","status":"PENDING"}],"nextCursor":null,"hasNext":false}}`}},
		// by-id 층(a164 — reconcileObject 는 result·first 만 봄).
		{"D0-detail-baseline", "detail", reply{200, `{"result":` + detail + `}`}},
		{"D1-detail-envelope-dup-result", "detail", reply{200, `{"result":` + detailFire + `,"result":` + detail + `}`}},
		{"D2-detail-envelope-fold-Result", "detail", reply{200, `{"result":` + detailFire + `,"Result":` + detail + `}`}},
		{"D3-detail-result-dup-status", "detail", reply{200, `{"result":{"conditionalOrderId":"CO-T","type":"SINGLE","status":"COMPLETED","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"second":null}}`}},
		{"D4-detail-result-fold-Status", "detail", reply{200, `{"result":{"conditionalOrderId":"CO-T","type":"SINGLE","Status":"COMPLETED","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"second":null}}`}},
		// 종목 양성 대조(unwrapAndDecode → []struct{Symbol}).
		{"I0-instrument-baseline", "inst", reply{200, `{"result":[{"symbol":"005930"}]}`}},
		{"I1-instrument-dup-symbol", "inst", reply{200, `{"result":[{"symbol":"000660","symbol":"005930"}]}`}},
		// 상태 코드(send 는 2xx 전부를 성공으로 받음).
		{"S1-cond-201", "cond", reply{201, `{"result":` + emptyCond + `}`}},
		{"S2-cond-203", "cond", reply{203, `{"result":` + emptyCond + `}`}},
		{"S3-cond-206", "cond", reply{206, `{"result":` + emptyCond + `}`}},
		{"S4-order-202", "order", reply{202, `{"result":` + emptyOrder + `}`}},
		{"S5-detail-201", "detail", reply{201, `{"result":` + detail + `}`}},
		{"S6-inst-201", "inst", reply{201, `{"result":[{"symbol":"005930"}]}`}},
		{"S7-cond-299", "cond", reply{299, `{"result":` + emptyCond + `}`}},
		{"S8-cond-204-empty", "cond", reply{204, ``}},
		{"S9-cond-200-baseline", "cond", reply{200, `{"result":` + emptyCond + `}`}},
		{"S10-order-200-open", "order", reply{200, `{"result":` + openOrder + `}`}},
	}
	tmp, err := os.MkdirTemp("", "a168-measure-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)
	for i, k := range cases {
		fmt.Println(run(k, filepath.Join(tmp, fmt.Sprintf("t%d.json", i))))
	}
	// reconcileObject 의 뒤따르는 데이터 검사(dec.More)를 공개 해독기로 직접 잼 — 생산 경로에서는 result 원문이
	// unwrapAndDecode 의 json.RawMessage(값 하나)라 도달하지 않음(E3 줄이 그 앞에서 거절됨을 보임).
	for _, tail := range []string{`}`, `]`, ` {"x":1}`, `"x"`} {
		_, err := official.DecodeReconcileConditionalDetail([]byte(detail + tail))
		fmt.Printf("T-direct-detail-trailing %-10q -> %s\n", tail, verdict(err))
	}
}

func verdict(err error) string {
	switch {
	case err == nil:
		return "ACCEPT"
	case errors.Is(err, official.ErrReconcileSchema):
		return "REFUSE ErrReconcileSchema"
	case errors.Is(err, official.ErrServer):
		return "REFUSE ErrServer"
	default:
		return "REFUSE other: " + strings.SplitN(err.Error(), "\n", 2)[0]
	}
}

func run(k kase, cache string) string {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		w.WriteHeader(k.resp.status)
		fmt.Fprint(w, k.resp.body)
	}))
	defer srv.Close()
	c := official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, cache,
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()), official.WithAccountSeq(7))
	ctx := context.Background()
	var got string
	var err error
	switch k.read {
	case "cond":
		var p official.ReconcileConditionalPage
		p, err = c.ReconcileConditionalOrdersPage(ctx, "CLOSED", "005930", "", 100)
		got = fmt.Sprintf("rows=%d hasNext=%v", len(p.Rows), p.HasNext)
		for _, r := range p.Rows {
			got += fmt.Sprintf(" [id=%s status=%s triggered=%q second=%v firstStatus=%s]", r.ID, r.Status, r.TriggeredOrderID, r.HasSecond, r.FirstStatus)
		}
	case "order":
		var p official.ReconcileOrderPage
		p, err = c.ReconcileOpenOrdersPage(ctx, "005930", "", 100)
		got = fmt.Sprintf("rows=%d", len(p.Rows))
		for _, r := range p.Rows {
			got += fmt.Sprintf(" [id=%s symbol=%s status=%s]", r.ID, r.Symbol, r.Status)
		}
	case "detail":
		var d official.ReconcileConditionalDetail
		d, err = c.ReconcileConditionalOrder(ctx, "CO-T")
		got = fmt.Sprintf("silent=%v status=%s firstStatus=%s triggered=%q", d.Silent, d.Status, d.FirstStatus, d.TriggeredOrderID)
	case "inst":
		got, err = c.ReconcileInstrument(ctx, "005930")
		got = "echo=" + got
	}
	if err != nil {
		return fmt.Sprintf("%-32s %-6s %d -> %s", k.id, k.read, k.resp.status, verdict(err))
	}
	return fmt.Sprintf("%-32s %-6s %d -> ACCEPT %s", k.id, k.read, k.resp.status, got)
}
