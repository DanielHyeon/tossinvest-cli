//go:build tossos_testseams

package main

// verify_reconcile_seams_test.go — a121 tasks 2.4·2.4.1 의 전 경로 시험(tossos_testseams 태그 — `make test-seams`).
//
// Q1 보존 측정은 생산 빌드에서 nil 이라(design G1-4 P0-1) 명령은 거기서 거절만 한다. 명령을 끝까지 몰려면
// verifylive.SetReconcilePoliciesForTest 가 필요하고 그 함수는 이 태그 이진에만 있다.

import (
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

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/ratebudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

const seamTarget = "CO-A063-SEAM"

// seamA063Record 는 a063 형 KR 기록을 프로필 아래에 쓴다(등록 → 존속 → 취소 실패 404 → 정리 실패 404).
func seamA063Record(t *testing.T, configDir string, created time.Time) string {
	t.Helper()
	path := filepath.Join(configDir, verifylive.RecordFileName(verifylive.MarketKR))
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()
	acct := attest.Mask("123-45-678901")
	notFound := verifylive.Call{Endpoint: "DELETE /api/v1/conditional-orders/{id}", At: created.Add(30 * time.Minute),
		Error: "official: HTTP 404 conditional-order-not-found", ErrorCode: "conditional-order-not-found"}
	for _, e := range []verifylive.Entry{
		{Kind: verifylive.KindStep, StepID: verifylive.StepConditionalRegister, Verdict: verifylive.VerdictPass, AccountRef: acct,
			Calls: []verifylive.Call{{Endpoint: "POST /api/v1/conditional-orders", At: created}},
			Artifacts: []verifylive.Artifact{{Kind: verifylive.KindConditional, ID: seamTarget, Symbol: "005930",
				CreatedAt: created, Deliberate: true}}},
		{Kind: verifylive.KindStep, StepID: verifylive.StepConditionalPersist, Verdict: verifylive.VerdictPass, AccountRef: acct},
		{Kind: verifylive.KindStep, StepID: verifylive.StepConditionalCancel, Verdict: verifylive.VerdictFail, AccountRef: acct,
			Calls: []verifylive.Call{notFound}},
		{Kind: verifylive.KindCleanup, StepID: verifylive.StepCleanup, Verdict: verifylive.VerdictFail, AccountRef: acct,
			Calls: []verifylive.Call{notFound}},
	} {
		if err := rec.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func seamPolicies(t *testing.T) {
	t.Helper()
	t.Cleanup(verifylive.SetReconcilePoliciesForTest(48*time.Hour, "duration", time.Minute))
}

func seamApprove(t *testing.T) {
	t.Helper()
	previous := verifyReconcileApprove
	verifyReconcileApprove = func(context.Context, verifylive.ReconcileApproval) error { return nil }
	t.Cleanup(func() { verifyReconcileApprove = previous })
}

func seamRecordHasReconcileLine(t *testing.T, path string) bool {
	t.Helper()
	entries, err := verifylive.LoadEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Kind == verifylive.KindReconcile {
			return true
		}
	}
	return false
}

// --- 잠금 보유 --------------------------------------------------------------------------

type seamLockProbe struct {
	t         *testing.T
	configDir string
	checked   int
	lists     int
}

func (p *seamLockProbe) probe() {
	p.t.Helper()
	if lock, err := enginelock.Acquire(p.configDir); err == nil {
		lock.Release()
		p.t.Error("a list read ran without the journal execution flock held")
	}
	if lease, ok, err := ratebudget.TryAcquire(filepath.Join(p.configDir, ratebudget.FileName)); err != nil {
		p.t.Error(err)
	} else if ok {
		lease.Release()
		p.t.Error("a list read ran without the rate-budget lease held")
	}
	p.checked++
}

func (p *seamLockProbe) ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (official.ReconcileConditionalPage, error) {
	p.probe()
	p.lists++
	return official.ReconcileConditionalPage{Rows: []official.ReconcileConditionalRow{}}, nil
}

func (p *seamLockProbe) ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (official.ReconcileOrderPage, error) {
	p.probe()
	p.lists++
	return official.ReconcileOrderPage{Rows: []official.ReconcileOrderRow{}}, nil
}

func (p *seamLockProbe) ReconcileInstrument(ctx context.Context, symbol string) (string, error) {
	p.probe()
	return symbol, nil
}

// TestReconcileHoldsTheExecutionLockAndRateBudgetThroughTheReads 는 tasks 2.4.1 셋째 요구다 — 명령은 `verify run` 과
// 같은 배제(실행 flock + rate-budget lease)를 쥔 채 읽고 추가한다.
func TestReconcileHoldsTheExecutionLockAndRateBudgetThroughTheReads(t *testing.T) {
	configDir := testenv.Isolate(t)
	path := seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	seamApprove(t)
	probe := &seamLockProbe{t: t, configDir: configDir}
	previous := verifyReconcileReaderFactory
	verifyReconcileReaderFactory = func(context.Context, *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
		return probe, verifylive.ReconcileAccount{Ref: "123-45-678901", Seq: 7, Count: 1}, nil
	}
	t.Cleanup(func() { verifyReconcileReaderFactory = previous })

	err := runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR})
	if err != nil {
		t.Fatalf("every condition holds: %v", err)
	}
	if probe.checked < 7 { // 종목 조회 2 + 목록 3 × 2
		t.Fatalf("probed %d reads, want at least 7", probe.checked)
	}
	if !seamRecordHasReconcileLine(t, path) {
		t.Fatal("no reconcile line appended")
	}
}

// --- GET 만 -------------------------------------------------------------------------------

// TestReconcileSucceedsWithOnlyGetRequestsAndTheTokenPost 는 tasks 2.4·2.4.1 의 「no live mutation is reachable」 행동
// 쪽이다 — 실제 좁은 생성자 + 실제 official 클라이언트로 끝까지 몰고, 서버가 본 요청이 GET 과 토큰 갱신 POST 하나뿐이다.
func TestReconcileSucceedsWithOnlyGetRequestsAndTheTokenPost(t *testing.T) {
	configDir := testenv.Isolate(t)
	path := seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	seamApprove(t)
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`)
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

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR}); err != nil {
		t.Fatalf("every condition holds: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	lists := 0
	for _, s := range seen {
		switch {
		case s == "POST /oauth2/token":
		case strings.HasPrefix(s, "GET "):
			if strings.HasSuffix(s, "/conditional-orders") || strings.HasSuffix(s, "/orders") {
				lists++
			}
		default:
			t.Fatalf("reconciliation sent %s — only official GET reads and the token refresh POST are inside the boundary", s)
		}
	}
	if lists != 6 {
		t.Fatalf("want 6 list reads (three groups twice), got %d: %v", lists, seen)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(after), string(before)) || strings.Count(string(after[len(before):]), "\n") != 1 {
		t.Fatal("want exactly one local append and no rewrite")
	}
}

// --- supervisedProofs (soak attest · 엔진 인터록 입력) 불변 ---------------------------------

// TestReconcileLineDoesNotChangeSupervisedProofs 는 tasks 2.4 다 — soak attest 가 엔진 인터록 증명으로 빌려 가는 감독 증거
// (supervisedProofs → SucceededEndpoints)가 대사 전후로 같다.
func TestReconcileLineDoesNotChangeSupervisedProofs(t *testing.T) {
	configDir := testenv.Isolate(t)
	seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	seamApprove(t)
	root := &rootOptions{configDir: configDir}
	now := time.Now()
	before, err := supervisedProofs(root, &soakOptions{}, now, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	probe := &seamLockProbe{t: t, configDir: configDir}
	previous := verifyReconcileReaderFactory
	verifyReconcileReaderFactory = func(context.Context, *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
		return probe, verifylive.ReconcileAccount{Ref: "123-45-678901", Seq: 7, Count: 1}, nil
	}
	t.Cleanup(func() { verifyReconcileReaderFactory = previous })
	if err := runVerifyReconcile(reconcileCmdWithContext(context.Background()), root,
		&verifyReconcileOptions{market: verifylive.MarketKR}); err != nil {
		t.Fatalf("every condition holds: %v", err)
	}
	after, err := supervisedProofs(root, &soakOptions{}, now, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("reconciliation changed the supervised proofs\nbefore %+v\nafter  %+v", before, after)
	}
}

// --- 기본 승인이 명령에 배선됐는가 (A-RED 리뷰 P1-3) ------------------------------------------

func seamProbeFactory(t *testing.T, configDir string) *seamLockProbe {
	t.Helper()
	probe := &seamLockProbe{t: t, configDir: configDir}
	previous := verifyReconcileReaderFactory
	verifyReconcileReaderFactory = func(context.Context, *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
		return probe, verifylive.ReconcileAccount{Ref: "123-45-678901", Seq: 7, Count: 1}, nil
	}
	t.Cleanup(func() { verifyReconcileReaderFactory = previous })
	return probe
}

// TestReconcileDefaultApprovalStopsTheCommandBeforeAnyListRead — 승인 seam 을 바꾸지 않은 명령은 비대화형 stdin 에서
// 목록을 하나도 읽지 않고 RefuseNotApproved 로 멈춘다(명령이 기본 승인을 실제로 쓴다는 배선 핀).
func TestReconcileDefaultApprovalStopsTheCommandBeforeAnyListRead(t *testing.T) {
	configDir := testenv.Isolate(t)
	path := seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	setReconcileTerminal(t, "y\n", false)
	probe := seamProbeFactory(t, configDir)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR})
	requireReconcileRefusal(t, err, verifylive.RefuseNotApproved)
	if probe.lists != 0 {
		t.Fatalf("%d list read(s) before a refused approval", probe.lists)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("a refused approval changed the record")
	}
}

// TestReconcileInteractiveYesLetsTheCommandAppend — 같은 배선에서 대화형 y 는 진행한다(위 거절이 우연이 아님을 보이는 짝).
func TestReconcileInteractiveYesLetsTheCommandAppend(t *testing.T) {
	configDir := testenv.Isolate(t)
	path := seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	setReconcileTerminal(t, "y\n", true)
	seamProbeFactory(t, configDir)
	if err := runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR}); err != nil {
		t.Fatalf("an explicit interactive y with every condition holding must append: %v", err)
	}
	if !seamRecordHasReconcileLine(t, path) {
		t.Fatal("no reconcile line appended")
	}
}
