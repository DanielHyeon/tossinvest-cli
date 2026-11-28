package official

// reconcile_ambiguity_test.go — a168 2.2: 대사 읽기 네 개가 모호한 200 본문과 200 아닌 성공 상태를 거절하는지 클라이언트 경로째 잼.
//
// 공개·기존 API(Client.Reconcile*)만 부르므로 편집 전에도 컴파일됨 — 편집 전 실행이 RED 를 보임(analysis/red-green/).
// 판정 함수 직접 시험과 계약 키 목록 핀은 reconcile_response_test.go(GREEN 로트)에 있음.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// ambReply 는 시험 서버의 응답 한 번임.
type ambReply struct {
	status int
	body   string
}

// ambServer 는 토큰 요청 밖의 요청마다 replies 를 차례로 내고, 다 쓰면 마지막 응답을 되풀이함.
type ambServer struct {
	*httptest.Server
	mu   sync.Mutex
	next int
}

func newAmbServer(t *testing.T, replies ...ambReply) *ambServer {
	t.Helper()
	s := &ambServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		s.mu.Lock()
		reply := replies[min(s.next, len(replies)-1)]
		s.next++
		s.mu.Unlock()
		w.WriteHeader(reply.status)
		fmt.Fprint(w, reply.body)
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func (s *ambServer) client(t *testing.T) *Client {
	t.Helper()
	return New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(s.URL), WithHTTPClient(s.Client()), WithAccountSeq(7))
}

// ambRead 는 읽기 종류 하나를 부르고 (요약, 오류)를 돌려줌.
func ambRead(t *testing.T, c *Client, read string) (string, error) {
	t.Helper()
	ctx := context.Background()
	switch read {
	case "cond":
		p, err := c.ReconcileConditionalOrdersPage(ctx, "CLOSED", "005930", "", 100)
		return fmt.Sprintf("%+v", p), err
	case "order":
		p, err := c.ReconcileOpenOrdersPage(ctx, "005930", "", 100)
		return fmt.Sprintf("%+v", p), err
	case "inst":
		return c.ReconcileInstrument(ctx, "005930")
	case "detail":
		d, err := c.ReconcileConditionalOrder(ctx, "CO-T")
		if err == nil && d.Silent {
			return "silent", errors.New("classified silent")
		}
		return fmt.Sprintf("%+v", d), err
	}
	t.Fatalf("unknown read %q", read)
	return "", nil
}

const (
	ambEmptyCond  = `{"conditionalOrders":[],"nextCursor":null,"hasNext":false}`
	ambEmptyOrder = `{"orders":[],"nextCursor":null,"hasNext":false}`
	ambFiredRow   = `{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","type":"SINGLE","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"second":null}`
	ambDetail     = `{"conditionalOrderId":"CO-T","type":"SINGLE","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED","triggeredOrderId":null},"second":null}`
	ambDetailFire = `{"conditionalOrderId":"CO-T","type":"SINGLE","status":"COMPLETED","symbol":"005930","market":"KR","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"second":null}`
)

// ambRow 는 CLOSED 목록 한 행짜리 페이지임 — first 원문만 바꿈.
func ambRow(status, first string) string {
	return `{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"` + status +
		`","type":"SINGLE","first":` + first + `,"second":null}],"nextCursor":null,"hasNext":false}`
}

// ambPage 는 행 하나(원문)짜리 조건주문 페이지임.
func ambPage(row string) string {
	return `{"conditionalOrders":[` + row + `],"nextCursor":null,"hasNext":false}`
}

// TestReconcileReadsRefuseAmbiguousBodies 는 design §8 R1~R5 · R10 · R10b · R12(official 층)와 Manager 결정 4 의 단독 변형임.
// 각 본문은 편집 전 측정(decode_ambiguity-output.txt)에서 ACCEPT 였고, 뒤 값이나 접힌 키가 판정 입력을 바꿨음.
func TestReconcileReadsRefuseAmbiguousBodies(t *testing.T) {
	cases := map[string]struct{ read, body string }{
		// 봉투(R2 · R12)
		"E1-envelope-dup-result":   {"cond", `{"result":` + ambPage(ambFiredRow) + `,"result":` + ambEmptyCond + `}`},
		"E2-envelope-fold-Result":  {"cond", `{"result":` + ambPage(ambFiredRow) + `,"Result":` + ambEmptyCond + `}`},
		"E2b-envelope-lone-Result": {"cond", `{"Result":` + ambEmptyCond + `}`},
		// 목록 result(R2 · R3)
		"R1-result-dup-collection": {"cond", wrapResult(`{"conditionalOrders":[` + ambFiredRow + `],"conditionalOrders":[],"nextCursor":null,"hasNext":false}`)},
		"R2-result-dup-hasNext":    {"cond", wrapResult(`{"conditionalOrders":[],"nextCursor":"C2","hasNext":true,"hasNext":false}`)},
		"R4-order-dup-collection":  {"order", wrapResult(`{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING"}],"orders":[],"nextCursor":null,"hasNext":false}`)},
		// 행 · first(R1)
		"W1-first-dup-triggered-null":   {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","triggeredOrderId":"ORD-9","triggeredOrderId":null}`))},
		"W1b-first-dup-triggered-empty": {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","triggeredOrderId":"ORD-9","triggeredOrderId":""}`))},
		"W2b-first-fold-triggered":      {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","triggeredOrderId":"ORD-9","TriggeredOrderId":""}`))},
		"W2c-first-escaped-fold":        {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","triggeredOrderId":"ORD-9","\u0054riggeredOrderId":""}`))},
		"W3-row-dup-status":             {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED"},"second":null}`))},
		"W3b-row-fold-status":           {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"COMPLETED","Status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED"},"second":null}`))},
		"W4-row-dup-second":             {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"EXPIRED","type":"OCO","first":{"status":"EXPIRED"},"second":{"status":"WATCHING"},"second":null}`))},
		"W4b-row-long-s-second":         {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"EXPIRED","type":"OCO","first":{"status":"EXPIRED"},"second":{"status":"WATCHING"},"ſecond":null}`))},
		"W4c-row-kelvin-market":         {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","mar` + "\u212a" + `et":"US","status":"EXPIRED","type":"SINGLE","first":{"status":"EXPIRED"},"second":null}`))},
		"W5-row-dup-first":              {"cond", wrapResult(ambPage(`{"conditionalOrderId":"CO-OTHER","symbol":"005930","market":"KR","status":"CANCELED","type":"SINGLE","first":{"status":"COMPLETED","triggeredOrderId":"ORD-9"},"first":{"status":"CANCELED"},"second":null}`))},
		"W9-row-lone-triggered-variant": {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","TriggeredOrderId":"ORD-9"}`))},
		"W6-order-dup-status":           {"order", wrapResult(`{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING","status":"CANCELED"}],"nextCursor":null,"hasNext":false}`)},
		"W7-order-fold-symbol":          {"order", wrapResult(`{"orders":[{"orderId":"O-1","symbol":"005930","Symbol":"000660","status":"PENDING"}],"nextCursor":null,"hasNext":false}`)},
		// 판정에 안 쓰는 깊은 필드(R10)와 계약 밖 접힘 짝(R10b)
		"W8-deep-dup-ignored":          {"cond", wrapResult(ambRow("EXPIRED", `{"status":"EXPIRED","condition":{"price":"1","price":"2"}}`))},
		"R10-order-metadata-dup":       {"order", wrapResult(`{"orders":[],"nextCursor":null,"hasNext":false,"metadata":{"x":1,"x":2}}`)},
		"R10b-noncontract-fold-pair":   {"order", wrapResult(`{"orders":[],"nextCursor":null,"hasNext":false,"metadata":{"datetime":1,"dateTime":2}}`)},
		"R10b-noncontract-long-s-pair": {"order", wrapResult(`{"orders":[],"nextCursor":null,"hasNext":false,"metadata":{"size":1,"ſize":2}}`)},
		// 종목(R4)
		"I1-instrument-dup-symbol": {"inst", `{"result":[{"symbol":"000660","symbol":"005930"}]}`},
		// by-id(R5 · Manager 결정 4)
		"D1-detail-envelope-dup-result":   {"detail", `{"result":` + ambDetailFire + `,"result":` + ambDetail + `}`},
		"D2-detail-envelope-fold-Result":  {"detail", `{"result":` + ambDetailFire + `,"Result":` + ambDetail + `}`},
		"D4-detail-fold-Status":           {"detail", wrapResult(`{"conditionalOrderId":"CO-T","type":"SINGLE","Status":"COMPLETED","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"second":null}`)},
		"D5-detail-lone-TriggeredOrderId": {"detail", wrapResult(`{"conditionalOrderId":"CO-T","type":"SINGLE","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED","TriggeredOrderId":"ORD-9"}}`)},
		"D6-detail-lone-Second":           {"detail", wrapResult(`{"conditionalOrderId":"CO-T","type":"SINGLE","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"Second":{"status":"WATCHING"}}`)},
		"D7-detail-lone-long-s-second":    {"detail", wrapResult(`{"conditionalOrderId":"CO-T","type":"SINGLE","status":"EXPIRED","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"ſecond":{"status":"WATCHING"}}`)},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ambRead(t, newAmbServer(t, ambReply{http.StatusOK, c.body}).client(t), c.read)
			if err == nil {
				t.Fatalf("ambiguous 200 body was accepted as %s — body %s", got, c.body)
			}
			if !errors.Is(err, ErrReconcileSchema) {
				t.Fatalf("want ErrReconcileSchema, got %v", err)
			}
		})
	}
}

// TestReconcileReadsRefuseANon200Success 는 R6 — 네 읽기 × 200 아닌 2xx. 계약(openapi)의 성공 응답은 200 하나뿐임.
// 204 는 빈 본문이라 편집 전에도 해독 실패(ErrServer)로 거절됨 — 특성 고정으로 둠.
func TestReconcileReadsRefuseANon200Success(t *testing.T) {
	bodies := map[string]string{
		"cond": wrapResult(ambEmptyCond), "order": wrapResult(ambEmptyOrder),
		"inst": `{"result":[{"symbol":"005930"}]}`, "detail": wrapResult(ambDetail),
	}
	for read, body := range bodies {
		for _, status := range []int{201, 202, 203, 206, 299} {
			t.Run(fmt.Sprintf("%s-%d", read, status), func(t *testing.T) {
				got, err := ambRead(t, newAmbServer(t, ambReply{status, body}).client(t), read)
				if err == nil {
					t.Fatalf("HTTP %d was accepted as %s — the reconcile contract names 200 only", status, got)
				}
				// 상태 판정이 거절했음을 문구로 묶음 — 다른 갈래의 우연한 거절로 통과하지 않게 함(리뷰 1회차 Testing)
				if !errors.Is(err, ErrReconcileSchema) || !strings.Contains(err.Error(), "the reconcile contract names 200 only") {
					t.Fatalf("HTTP %d: want the status refusal, got %v", status, err)
				}
			})
		}
		t.Run(read+"-204-empty", func(t *testing.T) {
			if _, err := ambRead(t, newAmbServer(t, ambReply{http.StatusNoContent, ``}).client(t), read); err == nil {
				t.Fatal("HTTP 204 with an empty body was accepted")
			}
		})
	}
}

// TestReconcileReadsJudgeTheLastAttempt 는 R7 · C2 — 401 갱신 재시도 뒤의 마지막 시도가 판정 대상임.
// C2(401 본문에 중복 키 → 200 정상 수락)는 「첫 시도」 변이(M2)를 잡는 양성 대조임.
func TestReconcileReadsJudgeTheLastAttempt(t *testing.T) {
	dupAuth := ambReply{http.StatusUnauthorized, `{"error":{"code":"unauthorized","code":"expired"}}`}
	for _, read := range []string{"cond", "order", "inst", "detail"} {
		ok := map[string]string{"cond": wrapResult(ambEmptyCond), "order": wrapResult(ambEmptyOrder),
			"inst": `{"result":[{"symbol":"005930"}]}`, "detail": wrapResult(ambDetail)}[read]
		t.Run(read+"-401-then-200", func(t *testing.T) {
			if got, err := ambRead(t, newAmbServer(t, dupAuth, ambReply{http.StatusOK, ok}).client(t), read); err != nil {
				t.Fatalf("401 then a clean 200 must be accepted (judge the last attempt, not the first): %v / %s", err, got)
			}
		})
		t.Run(read+"-401-then-201", func(t *testing.T) {
			if got, err := ambRead(t, newAmbServer(t, dupAuth, ambReply{http.StatusCreated, ok}).client(t), read); err == nil {
				t.Fatalf("a retry ending on 201 was accepted as %s", got)
			}
		})
	}
}

// TestReconcileReadsKeepTheirAcceptance 는 C1 — 정상 200 응답은 편집 전과 같은 값으로 읽힘(양성 대조).
func TestReconcileReadsKeepTheirAcceptance(t *testing.T) {
	cases := map[string]struct{ read, body, want string }{
		"cond-fired-row": {"cond", wrapResult(ambPage(ambFiredRow)),
			"{Rows:[{ID:CO-OTHER Symbol:005930 Market:KR Status:COMPLETED TriggeredOrderID:ORD-9 HasSecond:false FirstStatus:COMPLETED}] NextCursor: HasNext:false}"},
		"order-open-row": {"order", wrapResult(`{"orders":[{"orderId":"O-1","symbol":"005930","status":"PENDING"}],"nextCursor":"N","hasNext":true}`),
			"{Rows:[{ID:O-1 Symbol:005930 Status:PENDING}] NextCursor:N HasNext:true}"},
		"inst":   {"inst", `{"result":[{"symbol":"005930","name":"x","market":"KOSPI"}]}`, "005930"},
		"detail": {"detail", wrapResult(ambDetail), ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ambRead(t, newAmbServer(t, ambReply{http.StatusOK, c.body}).client(t), c.read)
			if err != nil {
				t.Fatalf("a well-formed 200 must be accepted: %v", err)
			}
			if c.want != "" && got != c.want {
				t.Fatalf("decoded as %s, want %s", got, c.want)
			}
		})
	}
}

// TestReconcileReadsRefuseTrailingData 는 R-T — 꼬리 데이터는 편집 전에도 unwrapAndDecode 가 거절함(design P9). 새 판정이 그
// 의존 위에 있음을 고정함(꼬리를 두 번 판정하지 않음).
func TestReconcileReadsRefuseTrailingData(t *testing.T) {
	for _, tail := range []string{`{"x":1}`, `}`, `]`} {
		t.Run(tail, func(t *testing.T) {
			if _, err := ambRead(t, newAmbServer(t, ambReply{http.StatusOK, wrapResult(ambEmptyCond) + tail}).client(t), "cond"); err == nil {
				t.Fatalf("trailing %q was accepted", tail)
			}
		})
	}
}

// TestReconcileReadsKeepTheExistingSchemaRefusalFirst 는 C5 — 스키마 구멍과 중복 키를 함께 가진 본문은 기존 해독의 문구로 거절됨.
// 새 판정이 기존 해독 **뒤**에 있음을 고정함(앞으로 옮기면 기존 스키마 시험이 새 가드를 재게 됨).
// 네 리더 모두 고정함(리뷰 1회차 Testing — 처음에는 조건주문 목록만 고정했음).
func TestReconcileReadsKeepTheExistingSchemaRefusalFirst(t *testing.T) {
	for read, c := range map[string]struct{ body, want string }{
		"cond":   {wrapResult(`{"conditionalOrders":[],"nextCursor":null,"x":1,"x":2}`), "hasNext is missing"},
		"order":  {wrapResult(`{"orders":[],"nextCursor":null,"x":1,"x":2}`), "hasNext is missing"},
		"inst":   {`{"result":[{"symbol":"005930","x":1,"x":2},{"symbol":"005930"}]}`, "returned 2 items"},
		"detail": {wrapResult(`{"conditionalOrderId":"CO-T","type":"SINGLE","symbol":"005930","market":"KR","first":{"status":"EXPIRED"},"meta":{"x":1,"x":2}}`), "status is missing"},
	} {
		t.Run(read, func(t *testing.T) {
			_, err := ambRead(t, newAmbServer(t, ambReply{http.StatusOK, c.body}).client(t), read)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want the existing refusal (%s) before the ambiguity check, got %v", c.want, err)
			}
		})
	}
}

// TestReconcileReadsKeepTheirInputChecks 는 기존 입력 검사 갈래의 특성 고정임(FLM B1 · 종목 B2) — 편집 전후 같음.
func TestReconcileReadsKeepTheirInputChecks(t *testing.T) {
	c := newAmbServer(t, ambReply{http.StatusOK, `{"result":[{"symbol":"005930"},{"symbol":"005930"}]}`}).client(t)
	ctx := context.Background()
	if _, err := c.ReconcileConditionalOrdersPage(ctx, " ", "005930", "", 100); !errors.Is(err, ErrOrderStatusRequired) {
		t.Fatalf("empty status: %v", err)
	}
	if _, err := c.ReconcileConditionalOrder(ctx, " "); !errors.Is(err, ErrReconcileSchema) {
		t.Fatalf("empty id: %v", err)
	}
	if _, err := c.ReconcileInstrument(ctx, "005930"); !errors.Is(err, ErrReconcileSchema) {
		t.Fatalf("two instrument items: %v", err)
	}
}

// TestReconcileByIDSilentClassificationIsUnchanged 는 C4 — 404 conditional-order-not-found 는 편집 뒤에도 silent(a164 분류).
func TestReconcileByIDSilentClassificationIsUnchanged(t *testing.T) {
	c := newAmbServer(t, ambReply{http.StatusNotFound, `{"error":{"requestId":"r","code":"conditional-order-not-found","message":"x"}}`}).client(t)
	d, err := c.ReconcileConditionalOrder(context.Background(), "CO-T")
	if err != nil || !d.Silent {
		t.Fatalf("measured not-found must stay silent: %+v %v", d, err)
	}
}
