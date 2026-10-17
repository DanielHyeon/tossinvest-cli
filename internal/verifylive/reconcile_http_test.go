package verifylive

// reconcile_http_test.go — a121 codex F2·R2-1 의 조합 사례: 기형 목록 응답이 **두 번 같게** 오고 종목 조회는 성공해도
// 빈 목록이 아니라 거절이 된다. 실제 official.Client(대사 읽기)를 httptest 에 붙여 해독 경로째 잰다 — 가짜 읽기는
// 이미 해독된 페이지를 주므로 이 구멍을 못 본다. 대조군(정상 빈 페이지)은 수락되어, 차이가 응답 모양뿐임을 보인다.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type rcHTTPServer struct {
	*httptest.Server
	mu   sync.Mutex
	seen []string
}

func newRCHTTPServer(t *testing.T, condBody, orderBody string) *rcHTTPServer {
	t.Helper()
	s := &rcHTTPServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.seen = append(s.seen, r.Method+" "+r.URL.Path)
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
			fmt.Fprintf(w, `{"result":[{"symbol":%q,"name":"삼성전자","market":"KOSPI"}]}`, r.URL.Query().Get("symbols"))
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
		}
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func (s *rcHTTPServer) client(t *testing.T) *official.Client {
	return official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		official.WithBaseURL(s.URL), official.WithHTTPClient(s.Client()), official.WithAccountSeq(7))
}

const (
	rcEmptyCond  = `{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`
	rcEmptyOrder = `{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`
)

// TestReconcileAcceptsWellFormedEmptyListsOverHTTP 는 아래 거절들의 대조군이다.
func TestReconcileAcceptsWellFormedEmptyListsOverHTTP(t *testing.T) {
	srv := newRCHTTPServer(t, rcEmptyCond, rcEmptyOrder)
	h := newRCHarness(t, rcA063Entries())
	p := h.params()
	p.Reader = srv.client(t)
	if _, err := Reconcile(t.Context(), p); err != nil {
		t.Fatalf("well-formed empty lists must reconcile: %v", err)
	}
}

// TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree 는 F2·R2-1 이다.
func TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree(t *testing.T) {
	cases := map[string]struct{ cond, order string }{
		"cond-result-null":         {`{"result":null}`, rcEmptyOrder},
		"cond-result-empty":        {`{"result":{}}`, rcEmptyOrder},
		"cond-collection-missing":  {`{"result":{"nextCursor":null,"hasNext":false}}`, rcEmptyOrder},
		"cond-collection-null":     {`{"result":{"conditionalOrders":null,"nextCursor":null,"hasNext":false}}`, rcEmptyOrder},
		"cond-has-next-null":       {`{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":null}}`, rcEmptyOrder},
		"cond-has-next-missing":    {`{"result":{"conditionalOrders":[],"nextCursor":null}}`, rcEmptyOrder},
		"cond-has-next-no-cursor":  {`{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":true}}`, rcEmptyOrder},
		"order-result-null":        {rcEmptyCond, `{"result":null}`},
		"order-result-empty":       {rcEmptyCond, `{"result":{}}`},
		"order-collection-null":    {rcEmptyCond, `{"result":{"orders":null,"nextCursor":null,"hasNext":false}}`},
		"order-has-next-null":      {rcEmptyCond, `{"result":{"orders":[],"nextCursor":null,"hasNext":null}}`},
		"order-has-next-no-cursor": {rcEmptyCond, `{"result":{"orders":[],"nextCursor":"","hasNext":true}}`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			srv := newRCHTTPServer(t, c.cond, c.order)
			h := newRCHarness(t, rcA063Entries())
			before := h.bytes()
			p := h.params()
			p.Reader = srv.client(t)
			_, err := Reconcile(t.Context(), p)
			rcRequireRefusal(t, err, RefuseReadError)
			rcRequireUnchanged(t, h, before)
			srv.mu.Lock()
			seen := strings.Join(srv.seen, ",")
			srv.mu.Unlock()
			if !strings.Contains(seen, "GET /api/v1/stocks") {
				t.Fatalf("the instrument control did not run, so this case does not exercise the combination: %s", seen)
			}
		})
	}
}
