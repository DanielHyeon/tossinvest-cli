//go:build tossos_testseams

package main

// verify_reconcile_expired_seams_test.go — a164 의 전 경로 시험(tossos_testseams — supervisedProofs 가 이 태그의 seam 을 씀).
// 실제 좁은 생성자 + 실제 official 클라이언트 + httptest 로 `verify reconcile --expired` 를 끝까지 몬다. Q1 정책 seam 은
// 부르지 않는다(생산과 같은 nil) — 만료 모드는 Q1 없이 닫혀야 한다(design D7).

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// TestExpiryReconcileCommandAppendsWithOnlyGetReadsAndLeavesProofsUnchanged 는 tasks 2.5·2.7 의 명령 쪽이다 — 요청은 GET 과
// 토큰 POST 뿐(by-id 2 · 목록 6), 줄 하나가 reconciled_expired 로 붙고, 감독 증명(soak attest·엔진 인터록 입력)은 그대로다.
func TestExpiryReconcileCommandAppendsWithOnlyGetReadsAndLeavesProofsUnchanged(t *testing.T) {
	configDir := testenv.Isolate(t)
	path := seamA063Record(t, configDir, time.Now().Add(-90*24*time.Hour))
	seamApprove(t)
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Path+"?"+r.URL.Query().Get("status"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders/"+seamTarget:
			fmt.Fprint(w, expiredDetailBody(seamTarget, "005930"))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders" && r.URL.Query().Get("status") == "CLOSED":
			fmt.Fprint(w, `{"result":{"conditionalOrders":[{"conditionalOrderId":"`+seamTarget+`","type":"SINGLE","status":"EXPIRED",`+
				`"symbol":"005930","market":"KR","quantity":"1","orderType":"MARKET","expireDate":"2026-08-07",`+
				`"first":{"type":"STOP","status":"EXPIRED","orderSide":"SELL","triggerPrice":"3200","targetProfitRate":null,"orderPrice":null,"triggeredOrderId":null},`+
				`"second":null,"createdAt":"2026-07-31T18:49:05+09:00"}],"nextCursor":null,"hasNext":false}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/conditional-orders":
			fmt.Fprint(w, `{"result":{"conditionalOrders":[],"nextCursor":null,"hasNext":false}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orders":
			fmt.Fprint(w, `{"result":{"orders":[],"nextCursor":null,"hasNext":false}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/stocks":
			fmt.Fprintf(w, `{"result":[{"symbol":%q,"name":"x","market":"KOSPI"}]}`, r.URL.Query().Get("symbols"))
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":"not-found"}}`)
		}
	}))
	t.Cleanup(srv.Close)
	previous := verifyReconcileReaderFactory
	verifyReconcileReaderFactory = func(ctx context.Context, root *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
		return newVerifyReconcileReader(ctx, official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()))
	}
	t.Cleanup(func() { verifyReconcileReaderFactory = previous })

	root := &rootOptions{configDir: configDir}
	now := time.Now()
	proofsBefore, err := supervisedProofs(root, &soakOptions{}, now, 365*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cmd := reconcileCmdWithContext(context.Background())
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := runVerifyReconcile(cmd, root, &verifyReconcileOptions{market: verifylive.MarketKR, expired: true}); err != nil {
		t.Fatalf("every expiry condition holds: %v", err)
	}
	mu.Lock()
	byID, lists := 0, 0
	for _, s := range seen {
		switch {
		case strings.HasPrefix(s, "POST /oauth2/token"):
		case strings.HasPrefix(s, "GET /api/v1/conditional-orders/"):
			byID++
		case strings.HasPrefix(s, "GET /api/v1/conditional-orders?") || strings.HasPrefix(s, "GET /api/v1/orders?"):
			lists++
		case strings.HasPrefix(s, "GET "):
		default:
			t.Fatalf("expiry reconciliation sent %s — only official GET reads and the token POST are inside the boundary", s)
		}
	}
	mu.Unlock()
	if byID != 2 || lists != 6 {
		t.Fatalf("by-id %d (want 2), list %d (want 6): %v", byID, lists, seen)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(after, before) || bytes.Count(after[len(before):], []byte("\n")) != 1 ||
		!bytes.Contains(after[len(before):], []byte(`"reconciled_expired":true`)) || bytes.Contains(after[len(before):], []byte(`"calls"`)) {
		t.Fatalf("want exactly one appended reconciled-expired line without calls:\n%s", after[len(before):])
	}
	if !strings.Contains(out.String(), "reconciled expired:") {
		t.Fatalf("result line does not name the ending:\n%s", out.String())
	}
	proofsAfter, err := supervisedProofs(root, &soakOptions{}, now, 365*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(proofsBefore, proofsAfter) {
		t.Fatalf("expiry reconciliation changed the supervised proofs\nbefore %+v\nafter  %+v", proofsBefore, proofsAfter)
	}
}

// ReconcileConditionalOrder 는 a164 넷째 리더 메서드다 — 부재 모드 seam 시험은 부르지 않아야 하므로 오류로 답한다.
func (p *seamLockProbe) ReconcileConditionalOrder(ctx context.Context, id string) (official.ReconcileConditionalDetail, error) {
	p.probe()
	return official.ReconcileConditionalDetail{}, fmt.Errorf("seam probe: no by-id read in absence mode")
}
