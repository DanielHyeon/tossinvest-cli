package main

// verify_reconcile_test.go — a121 tasks 2.2.2·2.4.1 의 cmd 쪽: design G3-2(자체 Accounts() 로 계좌 수·seq·기형 행,
// seq 명시 재결속)·G3-3(프로필·환경 변수·override)·G3-4(시장)·동시성(flock·lease)·명령 표면(mutating 주석, 금지 플래그).
//
// 대사 판정(G1·G2·G3-1)은 internal/verifylive/reconcile_*_test.go 에 있다. 전 경로를 끝까지 모는 시험(잠금 보유·GET 만·
// supervisedProofs 불변)은 Q1·Q3 seam 이 필요해 verify_reconcile_seams_test.go(tossos_testseams 태그)에 있다.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/ratebudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

func requireReconcileRefusal(t *testing.T, err error, want verifylive.ReconcileRefusalCode) {
	t.Helper()
	var refusal *verifylive.ReconcileRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("want refusal %q, got %v", want, err)
	}
	if refusal.Code != want {
		t.Fatalf("want refusal %q, got %q (%s)", want, refusal.Code, refusal.Detail)
	}
}

// --- G3-3 · G3-4 사전 검사 ------------------------------------------------------------

// TestReconcilePreflightRefusals 는 프로필 결속이 구성으로 성립하는 세 조건(--config-dir 필수 · 환경 변수 자격 거절 ·
// --record override 거절)과 --market 필수다. 환경 변수는 **하나라도** 설정돼 있으면 거절(로트 1 D4 · Manager 2026-10-05).
func TestReconcilePreflightRefusals(t *testing.T) {
	configDir := testenv.Isolate(t)
	unset := func(string) string { return "" }
	only := func(name string) func(string) string {
		return func(k string) string {
			if k == name {
				return "set"
			}
			return ""
		}
	}
	cases := map[string]struct {
		root   *rootOptions
		opts   verifyReconcileOptions
		getenv func(string) string
		want   verifylive.ReconcileRefusalCode
	}{
		"no-config-dir":   {&rootOptions{}, verifyReconcileOptions{market: verifylive.MarketKR}, unset, verifylive.RefuseNoConfigDir},
		"env-key-only":    {&rootOptions{configDir: configDir}, verifyReconcileOptions{market: verifylive.MarketKR}, only("TOSSCTL_OPENAPI_KEY"), verifylive.RefuseEnvCredentials},
		"env-secret-only": {&rootOptions{configDir: configDir}, verifyReconcileOptions{market: verifylive.MarketKR}, only("TOSSCTL_OPENAPI_SECRET"), verifylive.RefuseEnvCredentials},
		"env-both": {&rootOptions{configDir: configDir}, verifyReconcileOptions{market: verifylive.MarketKR},
			func(string) string { return "set" }, verifylive.RefuseEnvCredentials},
		"record-override": {&rootOptions{configDir: configDir},
			verifyReconcileOptions{market: verifylive.MarketKR, record: filepath.Join(configDir, "elsewhere.jsonl")}, unset, verifylive.RefuseRecordOverride},
		"no-market": {&rootOptions{configDir: configDir}, verifyReconcileOptions{}, unset, verifylive.RefuseNoMarket},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			opts := c.opts
			_, err := reconcilePreflight(c.root, &opts, c.getenv)
			requireReconcileRefusal(t, err, c.want)
		})
	}
}

// TestReconcilePreflightDerivesTheRecordFromTheProfile — 통과하면 기록은 --config-dir 아래 시장별 파일이다.
func TestReconcilePreflightDerivesTheRecordFromTheProfile(t *testing.T) {
	configDir := testenv.Isolate(t)
	for _, market := range []string{verifylive.MarketKR, verifylive.MarketUS} {
		path, err := reconcilePreflight(&rootOptions{configDir: configDir}, &verifyReconcileOptions{market: market},
			func(string) string { return "" })
		if err != nil {
			t.Fatalf("%s: %v", market, err)
		}
		if want := filepath.Join(configDir, verifylive.RecordFileName(market)); path != want {
			t.Fatalf("%s: record %q, want %q", market, path, want)
		}
	}
}

// --- G3-2 좁은 생성자 -------------------------------------------------------------------

type reconcileAccountsServer struct {
	*httptest.Server
	mu       sync.Mutex
	seen     []string
	headers  map[string][]string
	accounts atomic.Int32
}

// newReconcileAccountsServer 는 /accounts 응답을 호출 순서별로 준다(마지막 값 반복).
func newReconcileAccountsServer(t *testing.T, accountBodies ...string) *reconcileAccountsServer {
	t.Helper()
	s := &reconcileAccountsServer{headers: map[string][]string{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.seen = append(s.seen, r.Method+" "+r.URL.Path)
		s.headers[r.URL.Path] = append(s.headers[r.URL.Path], r.Header.Get("X-Tossinvest-Account"))
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case r.URL.Path == "/api/v1/accounts":
			n := int(s.accounts.Add(1)) - 1
			if n >= len(accountBodies) {
				n = len(accountBodies) - 1
			}
			fmt.Fprint(w, accountBodies[n])
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
	t.Cleanup(s.Server.Close)
	return s
}

func (s *reconcileAccountsServer) listReads() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, r := range s.seen {
		if strings.HasSuffix(r, "/api/v1/conditional-orders") || strings.HasSuffix(r, "/api/v1/orders") {
			out = append(out, r)
		}
	}
	return out
}

func (s *reconcileAccountsServer) newReader(t *testing.T) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
	t.Helper()
	return newVerifyReconcileReader(context.Background(), official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(s.URL), official.WithHTTPClient(s.Client()))
}

func accountsBody(rows ...string) string {
	return `{"result":[` + strings.Join(rows, ",") + `]}`
}

func accountRow(no string, seq int) string {
	return fmt.Sprintf(`{"accountNo":%q,"accountSeq":%d,"accountType":"BROKERAGE"}`, no, seq)
}

// TestReconcileNarrowReaderRefusesAccountShapes 는 design G3-2(P1-2·Q5·R2-2) 다 — 계좌 0, DisplayName 비공란 2+,
// 끝 4자리 충돌, seq 0, 기형 행(빈 번호 + 양수 seq — 계수에서 빼지 않고 그 자체로 거절). 모두 목록 읽기 전에.
func TestReconcileNarrowReaderRefusesAccountShapes(t *testing.T) {
	cases := map[string]struct {
		body string
		want verifylive.ReconcileRefusalCode
	}{
		"no-account":       {accountsBody(), verifylive.RefuseAccountCount},
		"two-accounts":     {accountsBody(accountRow("123-45-678901", 5), accountRow("555-55-550000", 6)), verifylive.RefuseAccountCount},
		"suffix-collision": {accountsBody(accountRow("123-45-678901", 5), accountRow("999-99-998901", 6)), verifylive.RefuseAccountCount},
		"seq-zero":         {accountsBody(accountRow("123-45-678901", 0)), verifylive.RefuseAccountSeqZero},
		// 기형 행이 첫 행이면 transport 가 그 seq(7)를 캐시한다 — 표시는 다른 계좌를, 부재는 이 seq 에서 읽힌다.
		"malformed-row": {accountsBody(accountRow("", 7), accountRow("123-45-678901", 5)), verifylive.RefuseAccountMalformedRow},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			srv := newReconcileAccountsServer(t, c.body)
			reader, _, err := srv.newReader(t)
			requireReconcileRefusal(t, err, c.want)
			if reader != nil {
				t.Fatal("a refused construction returned a reader")
			}
			if reads := srv.listReads(); len(reads) != 0 {
				t.Fatalf("list reads before the account checks: %v", reads)
			}
		})
	}
}

// TestReconcileNarrowReaderBindsTheValidatedSequence 는 codex R2-2 다 — 생성자가 자체 Accounts() 로 검증한 seq 로 읽기를
// 묶는다. 목록 읽기 헤더는 검증한 5 이고, /accounts 는 정확히 두 번(계좌 판정 + 읽기용 클라이언트의 첫 범위 읽기 전
// 신원 재확인, codex CG-6)이다 — 지연 해석이 다시 일어나면 셋 이상이 된다. (CG-6 수리로 개정: 원래는 "한 번뿐" 이었고
// 서버가 둘째 호출에 다른 계좌(seq 8)를 주는 픽스처였다 — 그 모양은 이제 신원 재확인이 거절한다:
// TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead.)
// 한계(정직하게): 검증 행 = 첫 행인 한(기형 행 거절이 그것을 보장한다), 암묵 캐시 재사용과 명시 재결속은 같은 헤더를
// 낸다.
func TestReconcileNarrowReaderBindsTheValidatedSequence(t *testing.T) {
	srv := newReconcileAccountsServer(t,
		accountsBody(accountRow("123-45-678901", 5)),
		accountsBody(accountRow("123-45-678901", 5)),
		accountsBody(accountRow("999-99-990000", 8)))
	reader, account, err := srv.newReader(t)
	if err != nil {
		t.Fatalf("one well-formed account must construct a reader: %v", err)
	}
	if account.Ref != "123-45-678901" || account.Seq != 5 || account.Count != 1 {
		t.Fatalf("account facts %+v", account)
	}
	ctx := context.Background()
	if _, err := reader.ReconcileConditionalOrdersPage(ctx, "OPEN", "005930", "", 100); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.ReconcileOpenOrdersPage(ctx, "005930", "", 100); err != nil {
		t.Fatal(err)
	}
	if n := srv.accounts.Load(); n != 2 {
		t.Fatalf("/accounts read %d times; want 2 (validation + the reading client's identity recheck) — more means lazy re-resolution", n)
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	for _, path := range []string{"/api/v1/conditional-orders", "/api/v1/orders"} {
		hs := srv.headers[path]
		if len(hs) == 0 {
			t.Fatalf("no read of %s", path)
		}
		for _, h := range hs {
			if h != "5" {
				t.Fatalf("%s read with account header %q, want the validated sequence 5", path, h)
			}
		}
	}
}

// TestReconcileNarrowReaderNeverYieldsAWriteMethod 는 봉인의 행동 쪽이다 — 돌려받은 값에서 쓰기 7이름 어느 것도
// (임베딩 승격·type assertion 포함) 도달할 수 없다.
func TestReconcileNarrowReaderNeverYieldsAWriteMethod(t *testing.T) {
	srv := newReconcileAccountsServer(t, accountsBody(accountRow("123-45-678901", 5)))
	reader, _, err := srv.newReader(t)
	if err != nil || reader == nil {
		t.Fatalf("construction: %v", err)
	}
	if _, ok := reader.(verifylive.Broker); ok {
		t.Fatal("the reconcile reader is a verifylive.Broker")
	}
	if _, ok := reader.(*official.Client); ok {
		t.Fatal("the reconcile reader is the concrete *official.Client")
	}
	typ := reflect.TypeOf(reader)
	for _, name := range []string{"PlaceOrder", "CancelOrder", "ModifyOrder", "CreateConditionalOrder",
		"ModifyConditionalOrder", "ModifyConditionalOrderRef", "CancelConditionalOrder"} {
		if _, ok := typ.MethodByName(name); ok {
			t.Fatalf("the reconcile reader's dynamic type %v has %s", typ, name)
		}
	}
	for i := 0; i < typ.NumMethod(); i++ {
		if name := typ.Method(i).Name; !strings.HasPrefix(name, "Reconcile") {
			t.Fatalf("the reconcile reader's dynamic type exposes %s", name)
		}
	}
}

// --- 동시성 · 명령 표면 ------------------------------------------------------------------

func stubReconcileFactory(t *testing.T) *int {
	t.Helper()
	calls := 0
	previous := verifyReconcileReaderFactory
	verifyReconcileReaderFactory = func(context.Context, *rootOptions) (verifylive.ReconcileReader, verifylive.ReconcileAccount, error) {
		calls++
		return nil, verifylive.ReconcileAccount{}, errors.New("factory must not run")
	}
	t.Cleanup(func() { verifyReconcileReaderFactory = previous })
	return &calls
}

func reconcileCmdWithContext(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetContext(ctx)
	return cmd
}

// TestReconcileRefusesWhenTheExecutionLockIsHeld — 엔진·갱신·다른 검증이 실행 배제를 쥐고 있으면 계좌 읽기 전에 거절.
func TestReconcileRefusesWhenTheExecutionLockIsHeld(t *testing.T) {
	configDir := testenv.Isolate(t)
	lock, err := enginelock.Acquire(configDir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	calls := stubReconcileFactory(t)
	err = runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR})
	requireReconcileRefusal(t, err, verifylive.RefuseExecutionLock)
	if *calls != 0 {
		t.Fatal("the reader factory ran while the execution exclusion was held elsewhere")
	}
}

// TestReconcileRefusesWhenTheRateBudgetLeaseIsHeld — rate-budget lease 를 못 얻으면 계좌 읽기 전에 거절.
func TestReconcileRefusesWhenTheRateBudgetLeaseIsHeld(t *testing.T) {
	configDir := testenv.Isolate(t)
	lease, ok, err := ratebudget.TryAcquire(filepath.Join(configDir, ratebudget.FileName))
	if err != nil || !ok {
		t.Fatalf("holding the lease: %v %v", ok, err)
	}
	defer lease.Release()
	calls := stubReconcileFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err = runVerifyReconcile(reconcileCmdWithContext(ctx), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR})
	requireReconcileRefusal(t, err, verifylive.RefuseRateBudget)
	if *calls != 0 {
		t.Fatal("the reader factory ran without the rate-budget lease")
	}
}

// TestReconcileRefusesPreflightBeforeAnyLockOrRead — 사전 검사 거절은 잠금·읽기보다 앞선다.
func TestReconcileRefusesPreflightBeforeAnyLockOrRead(t *testing.T) {
	configDir := testenv.Isolate(t)
	t.Setenv("TOSSCTL_OPENAPI_KEY", "from-env")
	calls := stubReconcileFactory(t)
	err := runVerifyReconcile(reconcileCmdWithContext(context.Background()), &rootOptions{configDir: configDir},
		&verifyReconcileOptions{market: verifylive.MarketKR})
	requireReconcileRefusal(t, err, verifylive.RefuseEnvCredentials)
	if *calls != 0 {
		t.Fatal("the reader factory ran after a preflight refusal")
	}
}

// TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand 는 로트 1 S2 다 — mutating=true(기록에 영속 이벤트를 쓰고
// 사람 승인을 요구 — 대화형 에이전트 자동 실행 금지), source=official.
func TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"verify", "reconcile"})
	if err != nil || cmd == nil || cmd.Name() != "reconcile" {
		t.Fatalf("`tossctl verify reconcile` is not registered: %v", err)
	}
	if cmd.Annotations["mutating"] != "true" || cmd.Annotations["source"] != "official" {
		t.Fatalf("annotations %v, want mutating=true source=official", cmd.Annotations)
	}
}

// TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag 는 design 「Safety and attestation boundary」 다 — 임의 식별자·
// 심볼·변이·승인 우회·재시도·--redo 를 받지 않고, --market 은 받는다.
func TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag(t *testing.T) {
	cmd := newVerifyReconcileCmd(&rootOptions{})
	if cmd.Flags().Lookup("market") == nil {
		t.Fatal("--market is required by design G3-4 and must exist")
	}
	for _, name := range []string{"id", "artifact", "symbol", "redo", "resume", "yes", "force", "confirm",
		"cancel", "retry", "base-url", "account"} {
		if cmd.Flags().Lookup(name) != nil || cmd.PersistentFlags().Lookup(name) != nil {
			t.Fatalf("verify reconcile offers --%s", name)
		}
	}
}

// --- 사람 승인 기본값 · 이름만 있던 거절 모양 (A-RED 리뷰 P1-3 · P1-4) ------------------------

func setReconcileTerminal(t *testing.T, input string, interactive bool) *strings.Builder {
	t.Helper()
	out := &strings.Builder{}
	previous := verifyReconcileTerminal
	verifyReconcileTerminal = func() (io.Reader, io.Writer, bool) { return strings.NewReader(input), out, interactive }
	t.Cleanup(func() { verifyReconcileTerminal = previous })
	return out
}

func reconcileApprovalFixture() verifylive.ReconcileApproval {
	return verifylive.ReconcileApproval{RecordAccountMask: "*********8901", CurrentAccountMask: "*********8901",
		AccountCount: 1, Kind: verifylive.KindConditional, ID: "CO-A063", Symbol: "005930", Market: verifylive.MarketKR}
}

// TestReconcileDefaultApprovalRefusesANonInteractiveStdin 은 P1-3 의 앞 절반이다 — 기본 승인은 파이프·리다이렉트 stdin
// 에서 "y" 가 들어와도 거절한다(자동 승인 경로 없음). seam 시험들이 승인을 바꿔 끼우므로 기본값은 여기서만 잰다.
func TestReconcileDefaultApprovalRefusesANonInteractiveStdin(t *testing.T) {
	setReconcileTerminal(t, "y\n", false)
	if err := verifyReconcileApprove(context.Background(), reconcileApprovalFixture()); err == nil {
		t.Fatal("the default approval accepted a non-interactive stdin")
	}
}

// TestReconcileDefaultApprovalAcceptsOnlyAnExplicitInteractiveYes 는 P1-3 의 뒤 절반이다 — 대화형 단말에서 명시 y 만
// 진행하고 기본은 N 이다(타이핑 확인 문구 없음 — 학습 「확인 문구 금지」 와 같은 결). 질문은 y/N 을 보인다.
func TestReconcileDefaultApprovalAcceptsOnlyAnExplicitInteractiveYes(t *testing.T) {
	for input, accept := range map[string]bool{
		"y\n": true, "Y\n": true,
		"\n": false, "n\n": false, "N\n": false, "": false, "maybe\n": false,
	} {
		t.Run(strings.TrimSpace(input)+"/"+fmt.Sprint(accept), func(t *testing.T) {
			out := setReconcileTerminal(t, input, true)
			err := verifyReconcileApprove(context.Background(), reconcileApprovalFixture())
			if accept && err != nil {
				t.Fatalf("an explicit interactive %q must proceed: %v", input, err)
			}
			if !accept && err == nil {
				t.Fatalf("interactive %q must not proceed (default is N)", input)
			}
			if !strings.Contains(strings.ToLower(out.String()), "y/n") {
				t.Fatalf("the prompt does not show y/N:\n%s", out.String())
			}
		})
	}
}

// TestReconcileRefusesWithoutProfileCredentials 는 RefuseNoCredentials 의 시험이다 — 프로필에 자격 파일이 없으면 아무
// 요청도 없이 거절(환경 변수 경로는 preflight 가 이미 막는다).
func TestReconcileRefusesWithoutProfileCredentials(t *testing.T) {
	configDir := testenv.Isolate(t)
	reader, _, err := buildVerifyReconcileReader(context.Background(), &rootOptions{configDir: configDir})
	requireReconcileRefusal(t, err, verifylive.RefuseNoCredentials)
	if reader != nil {
		t.Fatal("a refused construction returned a reader")
	}
}

// TestReconcileNarrowReaderRefusesAnAccountsReadFailure 는 RefuseAccountsRead 의 시험이다 — 계좌 목록 읽기가 실패하면
// 계좌를 판정할 수 없으므로 목록 읽기 없이 거절.
func TestReconcileNarrowReaderRefusesAnAccountsReadFailure(t *testing.T) {
	var lists atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error":{"code":"internal-error"}}`)
		default:
			lists.Add(1)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	reader, _, err := newVerifyReconcileReader(context.Background(), official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()))
	requireReconcileRefusal(t, err, verifylive.RefuseAccountsRead)
	if reader != nil || lists.Load() != 0 {
		t.Fatalf("reader %v, other requests %d", reader, lists.Load())
	}
}

// TestReconcileDefaultTerminalIsNotInteractiveOnAPipeOrDevNull 은 A-RED 재검 P1-3 마감이다 — 위 승인 시험들은 단말 seam
// 을 바꿔 끼우므로, 실제 기본 verifyReconcileTerminal 이 대화형 판정을 맞게 하는지는 여기서만 잰다. stdin 이 파이프
// (`echo y | tossctl verify reconcile` 자가 승인 구멍)이거나 /dev/null 이면 interactive 는 거짓이어야 한다.
// /dev/null 은 문자 장치라 ModeCharDevice 판정(auth.go isTerminal)으로는 대화형으로 오판된다 — 실제 tty 판정이 필요하다.
func TestReconcileDefaultTerminalIsNotInteractiveOnAPipeOrDevNull(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reader.Close(); writer.Close() })
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { devNull.Close() })
	original := os.Stdin
	t.Cleanup(func() { os.Stdin = original })
	for name, stdin := range map[string]*os.File{"pipe": reader, "dev-null": devNull} {
		os.Stdin = stdin
		if _, _, interactive := verifyReconcileTerminal(); interactive {
			t.Fatalf("the default terminal reports an interactive stdin for %s", name)
		}
	}
}
