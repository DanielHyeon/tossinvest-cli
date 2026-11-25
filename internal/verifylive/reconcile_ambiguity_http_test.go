package verifylive

// reconcile_ambiguity_http_test.go — a168 2.2: 실제 official.Client(대사 읽기)를 httptest 에 붙여, 모호한 브로커 응답이 두 모드에서
// 기록 줄을 쓰지 않고 그 읽기의 거절 코드로 끝나는지 잼(design §8 R8a·R8b·R8c·R9·R11·R12). 편집 전 실행은 「줄 씀」 을 보임.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// rcAmbServer 는 조건주문 목록을 status 질의(OPEN·CLOSED)로 가르고, by-id 상세를 따로 내는 시험 브로커임.
type rcAmbServer struct {
	*httptest.Server
	mu   sync.Mutex
	seen []string
}

type rcAmbBodies struct {
	open, closed, orders string
	byIDStatus           int
	byID                 string
}

func newRCAmbServer(t *testing.T, b rcAmbBodies) *rcAmbServer {
	t.Helper()
	if b.byIDStatus == 0 {
		b.byIDStatus = http.StatusOK
	}
	s := &rcAmbServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.seen = append(s.seen, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders" && r.URL.Query().Get("status") == "OPEN":
			fmt.Fprint(w, b.open)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders" && r.URL.Query().Get("status") == "CLOSED":
			fmt.Fprint(w, b.closed)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/conditional-orders/"):
			w.WriteHeader(b.byIDStatus)
			fmt.Fprint(w, b.byID)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orders":
			fmt.Fprint(w, b.orders)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/stocks":
			fmt.Fprintf(w, `{"result":[{"symbol":%q,"name":"삼성전자","market":"KOSPI"}]}`, r.URL.Query().Get("symbols"))
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

const (
	rcAmbEmptyCond = `{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`
	// rcAmbLiveRow 는 같은 심볼의 살아 있는 다른 조건주문임 — 보이면 open-conditional 거절.
	rcAmbLiveRow = `{"conditionalOrderId":"CO-LIVE","symbol":"` + rcSymbol + `","market":"KR","status":"WATCHING","type":"SINGLE","first":{"status":"WATCHING"},"second":null}`
)

// rcAmbStatement 는 대상의 by-id statement 원문임 — first·leg 원문만 바꿈.
func rcAmbStatement(extra, first string) string {
	return `{"result":{"conditionalOrderId":"` + rcTargetID + `","type":"SINGLE","status":"EXPIRED","symbol":"` + rcSymbol +
		`","market":"KR"` + extra + `,"first":` + first + `}}`
}

// rcAmbClosedRow 는 다른 id 의 EXPIRED CLOSED 행 하나짜리 페이지임(만료 모드 A6 어휘 안) — first 원문만 바꿈.
func rcAmbClosedRow(first string) string {
	return `{"result":{"conditionalOrders":[{"conditionalOrderId":"CO-OTHER-EXPIRED","symbol":"` + rcSymbol +
		`","market":"KR","status":"EXPIRED","type":"SINGLE","first":` + first + `,"second":null}],"nextCursor":null,"hasNext":false}}`
}

// rcAmbRun 은 한 모드로 대사를 돌리고 (덧붙은 줄 수, 오류)를 돌려줌.
func rcAmbRun(t *testing.T, expired bool, b rcAmbBodies) (int, error) {
	t.Helper()
	srv := newRCAmbServer(t, b)
	h := newRCHarness(t, rcA063Entries())
	if expired {
		h.expired = true
		rcSetPolicies(t, nil, rcFreshness)
	}
	before := h.bytes()
	p := h.params()
	p.Reader = srv.client(t)
	_, err := Reconcile(t.Context(), p)
	return len(rcAppendedLines(t, before, h.bytes())), err
}

func (s *rcAmbServer) client(t *testing.T) ReconcileReader {
	t.Helper()
	return (&rcHTTPServer{Server: s.Server}).client(t)
}

func rcAmbDefaults(b rcAmbBodies) rcAmbBodies {
	if b.open == "" {
		b.open = rcAmbEmptyCond
	}
	if b.closed == "" {
		b.closed = rcAmbEmptyCond
	}
	if b.orders == "" {
		b.orders = rcEmptyOrder
	}
	if b.byID == "" {
		b.byID = rcAmbStatement("", `{"status":"EXPIRED"}`)
	}
	return b
}

// TestReconcileAmbiguityControlsAppendOverHTTP 는 아래 거절들의 양성 대조임 — 같은 서버·같은 정상 본문이면 두 모드 모두 줄 하나를 씀.
func TestReconcileAmbiguityControlsAppendOverHTTP(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(fmt.Sprintf("expired=%v", expired), func(t *testing.T) {
			lines, err := rcAmbRun(t, expired, rcAmbDefaults(rcAmbBodies{}))
			if err != nil || lines != 1 {
				t.Fatalf("well-formed responses must append exactly one line: err=%v lines=%d", err, lines)
			}
		})
	}
}

// TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP 는 R8a·R8b·R8c·R9·R11·R12 임.
func TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP(t *testing.T) {
	hideOpen := `{"result":{"conditionalOrders":[` + rcAmbLiveRow + `],"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`
	hideOpenEnvelope := `{"result":{"conditionalOrders":[` + rcAmbLiveRow + `],"nextCursor":null,"hasNext":false},"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`
	loneResult := `{"Result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`
	cases := []struct {
		name    string
		expired bool
		b       rcAmbBodies
		want    ReconcileRefusalCode
	}{
		{"R8a-expiry-other-expired-row-emptied-trigger", true,
			rcAmbBodies{closed: rcAmbClosedRow(`{"status":"EXPIRED","triggeredOrderId":"ORD-9","triggeredOrderId":""}`)}, RefuseReadError},
		{"R8b-absence-open-row-hidden-by-collection", false, rcAmbBodies{open: hideOpen}, RefuseReadError},
		{"R8b-absence-open-row-hidden-by-envelope", false, rcAmbBodies{open: hideOpenEnvelope}, RefuseReadError},
		{"R8c-expiry-open-row-hidden-by-collection", true, rcAmbBodies{open: hideOpen}, RefuseReadError},
		{"R9-expiry-by-id-201", true, rcAmbBodies{byIDStatus: http.StatusCreated}, RefuseExpiryReadError},
		{"R11-expiry-by-id-lone-TriggeredOrderId", true,
			rcAmbBodies{byID: rcAmbStatement("", `{"status":"EXPIRED","TriggeredOrderId":"ORD-9"}`)}, RefuseExpiryReadError},
		{"R11-expiry-by-id-lone-Second", true,
			rcAmbBodies{byID: rcAmbStatement(`,"Second":{"status":"WATCHING"}`, `{"status":"EXPIRED"}`)}, RefuseExpiryReadError},
		{"R11-expiry-by-id-fold-Status", true,
			rcAmbBodies{byID: rcAmbStatement(`,"Status":"COMPLETED"`, `{"status":"EXPIRED"}`)}, RefuseExpiryReadError},
		{"R12-absence-lone-Result", false, rcAmbBodies{open: loneResult, closed: loneResult}, RefuseReadError},
		{"R12-expiry-row-lone-TriggeredOrderId", true,
			rcAmbBodies{closed: rcAmbClosedRow(`{"status":"EXPIRED","TriggeredOrderId":"ORD-9"}`)}, RefuseReadError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lines, err := rcAmbRun(t, c.expired, rcAmbDefaults(c.b))
			if err == nil {
				t.Fatalf("the ambiguous response was reconciled — %d line(s) appended", lines)
			}
			if lines != 0 {
				t.Fatalf("a refusal appended %d line(s): %v", lines, err)
			}
			rcRequireRefusal(t, err, c.want)
		})
	}
}
