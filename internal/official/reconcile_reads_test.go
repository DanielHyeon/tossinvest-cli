package official

// reconcile_reads_test.go — a121 codex F2·R2-1·F6: 대사 읽기의 스키마 존재·형 검증과 second 다리 노출.
//
// 기존 경로(unwrapAndDecode → 구조체)는 `{"result":null}`·`{}`·null 값 필드를 무오류 빈 페이지로 접는다
// (TestTheExistingConditionalReadFoldsANullResultIntoAnEmptyPage 가 그 사실을 고정한다 — 대사 읽기가 따로 있어야 하는 이유).

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

// TestTheExistingConditionalReadFoldsANullResultIntoAnEmptyPage 는 design G1-6 F2 의 전제(client.go:213-227)를 고정한다.
// 이 시험이 실패하면(기존 경로가 거절하게 바뀌면) 대사 읽기의 존재 이유가 바뀐 것이다.
func TestTheExistingConditionalReadFoldsANullResultIntoAnEmptyPage(t *testing.T) {
	srv := reconcileTestServer(t, `{"result":null}`, `{"result":null}`)
	got, err := reconcileTestClient(t, srv).ProtectionConditionalOrdersRaw(context.Background(), "OPEN", "005930", "", 100)
	if err != nil || len(got.Orders) != 0 || got.HasNext {
		t.Fatalf("premise changed: existing read on {\"result\":null} = %+v, %v", got, err)
	}
}

func TestDecodeReconcileConditionalPageRefusesSchemaHoles(t *testing.T) {
	for name, body := range map[string]string{
		"result-null":           `null`,
		"result-empty-object":   `{}`,
		"result-array":          `[]`,
		"collection-missing":    `{"nextCursor":null,"hasNext":false}`,
		"collection-null":       `{"conditionalOrders":null,"nextCursor":null,"hasNext":false}`,
		"collection-object":     `{"conditionalOrders":{},"nextCursor":null,"hasNext":false}`,
		"has-next-missing":      `{"conditionalOrders":[],"nextCursor":null}`,
		"has-next-null":         `{"conditionalOrders":[],"nextCursor":null,"hasNext":null}`,
		"has-next-string":       `{"conditionalOrders":[],"nextCursor":null,"hasNext":"false"}`,
		"cursor-missing":        `{"conditionalOrders":[],"hasNext":false}`,
		"cursor-number":         `{"conditionalOrders":[],"nextCursor":5,"hasNext":false}`,
		"has-next-null-cursor":  `{"conditionalOrders":[],"nextCursor":null,"hasNext":true}`,
		"has-next-empty-cursor": `{"conditionalOrders":[],"nextCursor":"","hasNext":true}`,
		"row-null":              `{"conditionalOrders":[null],"nextCursor":null,"hasNext":false}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeReconcileConditionalPage([]byte(body))
			if !errors.Is(err, ErrReconcileSchema) {
				t.Fatalf("body %s: want ErrReconcileSchema, got %v", body, err)
			}
		})
	}
}

func TestDecodeReconcileOrderPageRefusesSchemaHoles(t *testing.T) {
	for name, body := range map[string]string{
		"result-null":           `null`,
		"result-empty-object":   `{}`,
		"collection-missing":    `{"nextCursor":null,"hasNext":false}`,
		"collection-null":       `{"orders":null,"nextCursor":null,"hasNext":false}`,
		"has-next-missing":      `{"orders":[],"nextCursor":null}`,
		"has-next-null":         `{"orders":[],"nextCursor":null,"hasNext":null}`,
		"cursor-missing":        `{"orders":[],"hasNext":false}`,
		"has-next-null-cursor":  `{"orders":[],"nextCursor":null,"hasNext":true}`,
		"has-next-empty-cursor": `{"orders":[],"nextCursor":"","hasNext":true}`,
		"row-null":              `{"orders":[null],"nextCursor":null,"hasNext":false}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeReconcileOrderPage([]byte(body))
			if !errors.Is(err, ErrReconcileSchema) {
				t.Fatalf("body %s: want ErrReconcileSchema, got %v", body, err)
			}
		})
	}
}

// TestDecodeReconcilePagesAcceptWellFormedPages 는 위 거절들의 대조군이다.
func TestDecodeReconcilePagesAcceptWellFormedPages(t *testing.T) {
	for _, body := range []string{
		`{"conditionalOrders":[],"nextCursor":null,"hasNext":false}`,
		`{"conditionalOrders":[],"nextCursor":"c2","hasNext":true}`,
	} {
		if _, err := DecodeReconcileConditionalPage([]byte(body)); err != nil {
			t.Fatalf("well-formed conditional page %s refused: %v", body, err)
		}
	}
	for _, body := range []string{
		`{"orders":[],"nextCursor":null,"hasNext":false}`,
		`{"orders":[],"nextCursor":"c2","hasNext":true}`,
	} {
		if _, err := DecodeReconcileOrderPage([]byte(body)); err != nil {
			t.Fatalf("well-formed order page %s refused: %v", body, err)
		}
	}
}

// TestDecodeReconcileConditionalPageKeepsTheSecondLegAndTheTrigger 는 codex F6 이다 — 채택 리더가 버리는 second 를
// 노출하고, 첫 다리의 triggeredOrderId 를 싣는다.
func TestDecodeReconcileConditionalPageKeepsTheSecondLegAndTheTrigger(t *testing.T) {
	body := `{"conditionalOrders":[` +
		`{"conditionalOrderId":"CO-1","clientOrderId":"c1","type":"SINGLE","status":"COMPLETED","symbol":"005930","market":"KR",` +
		`"quantity":"1","orderType":"MARKET","expireDate":"2026-10-30","createdAt":"2026-10-01T09:00:00+09:00",` +
		`"first":{"orderSide":"SELL","type":"STOP","status":"COMPLETED","triggerPrice":"70000","targetProfitRate":null,` +
		`"orderPrice":null,"triggeredOrderId":"ORD-9"},"second":null},` +
		`{"conditionalOrderId":"CO-2","clientOrderId":"c2","type":"OCO","status":"WATCHING","symbol":"005930","market":"KR",` +
		`"quantity":"1","orderType":"LIMIT","expireDate":"2026-10-30","createdAt":"2026-10-01T09:00:00+09:00",` +
		`"first":{"orderSide":"SELL","type":"STOP","status":"WATCHING","triggerPrice":"70000","targetProfitRate":null,` +
		`"orderPrice":"69900","triggeredOrderId":null},` +
		`"second":{"orderSide":"SELL","type":"PROFIT_RATE","status":"WATCHING","triggerPrice":null,"targetProfitRate":"10.5",` +
		`"orderPrice":"80000","triggeredOrderId":null}}` +
		`],"nextCursor":null,"hasNext":false}`
	page, err := DecodeReconcileConditionalPage([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(page.Rows))
	}
	first, second := page.Rows[0], page.Rows[1]
	if first.ID != "CO-1" || first.Symbol != "005930" || first.Market != "KR" || first.Status != "COMPLETED" ||
		first.TriggeredOrderID != "ORD-9" || first.HasSecond {
		t.Fatalf("SINGLE row decoded as %+v", first)
	}
	if second.ID != "CO-2" || !second.HasSecond || second.TriggeredOrderID != "" {
		t.Fatalf("OCO row decoded as %+v — the second leg must be visible", second)
	}
	if page.HasNext || page.NextCursor != "" {
		t.Fatalf("page boundary decoded as %+v", page)
	}
}

// --- 클라이언트 경로 ----------------------------------------------------------------

type reconcileServer struct {
	*httptest.Server
	mu      sync.Mutex
	seen    []string
	queries []string
	headers []string
}

func reconcileTestServer(t *testing.T, condBody, orderBody string) *reconcileServer {
	t.Helper()
	s := &reconcileServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.seen = append(s.seen, r.Method+" "+r.URL.Path)
		s.queries = append(s.queries, r.URL.Path+"?"+r.URL.RawQuery)
		s.headers = append(s.headers, r.URL.Path+" "+r.Header.Get("X-Tossinvest-Account"))
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders":
			fmt.Fprint(w, condBody)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orders":
			fmt.Fprint(w, orderBody)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/stocks":
			fmt.Fprintf(w, `{"result":[{"symbol":%q,"name":"x","market":"KOSPI"}]}`, r.URL.Query().Get("symbols"))
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func reconcileTestClient(t *testing.T, s *reconcileServer) *Client {
	t.Helper()
	return New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(s.URL), WithHTTPClient(s.Client()), WithAccountSeq(7))
}

// TestReconcileClientReadsRefuseANullResult 는 클라이언트 경로가 기존 구조체 해독을 거치지 않음을 잰다.
func TestReconcileClientReadsRefuseANullResult(t *testing.T) {
	srv := reconcileTestServer(t, `{"result":null}`, `{"result":null}`)
	c := reconcileTestClient(t, srv)
	if _, err := c.ReconcileConditionalOrdersPage(context.Background(), "OPEN", "005930", "", 100); !errors.Is(err, ErrReconcileSchema) {
		t.Fatalf("conditional page on {\"result\":null}: want ErrReconcileSchema, got %v", err)
	}
	if _, err := c.ReconcileOpenOrdersPage(context.Background(), "005930", "", 100); !errors.Is(err, ErrReconcileSchema) {
		t.Fatalf("order page on {\"result\":null}: want ErrReconcileSchema, got %v", err)
	}
}

// TestReconcileClientReadsSendTheDocumentedGetQueries — 세 읽기 모두 GET, 계좌 헤더(목록 둘), status·symbol·cursor·limit.
func TestReconcileClientReadsSendTheDocumentedGetQueries(t *testing.T) {
	srv := reconcileTestServer(t, `{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`,
		`{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`)
	c := reconcileTestClient(t, srv)
	ctx := context.Background()
	if _, err := c.ReconcileConditionalOrdersPage(ctx, "CLOSED", "005930", "cur-1", 100); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ReconcileOpenOrdersPage(ctx, "005930", "cur-2", 100); err != nil {
		t.Fatal(err)
	}
	echo, err := c.ReconcileInstrument(ctx, "005930")
	if err != nil || echo != "005930" {
		t.Fatalf("instrument echo %q, %v", echo, err)
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for _, s := range srv.seen {
		if s != "POST /oauth2/token" && !strings.HasPrefix(s, "GET ") {
			t.Fatalf("reconcile read sent %s", s)
		}
	}
	joined := strings.Join(srv.queries, "\n")
	for _, want := range []string{
		"/api/v1/conditional-orders?", "status=CLOSED", "symbol=005930", "cursor=cur-1", "limit=100",
		"/api/v1/orders?", "status=OPEN", "cursor=cur-2", "/api/v1/stocks?symbols=005930",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("queries lack %q:\n%s", want, joined)
		}
	}
	for _, h := range srv.headers {
		if (strings.HasPrefix(h, "/api/v1/conditional-orders ") || strings.HasPrefix(h, "/api/v1/orders ")) && h[strings.LastIndex(h, " ")+1:] != "7" {
			t.Fatalf("account-scoped read without the bound account header: %q", h)
		}
	}
}
