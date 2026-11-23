package engine_test

// a199_exit_only_runtime_test.go — 청산 전용 기동의 **실제 런타임** 연결(Manager 결정 9 · 구현 리뷰 r1 처분 4 · 7(f) · 7(g)).
//
// 조립은 생산과 같은 길: engine.New(게이트 ON · 생산 Guardian · 실제 원장 · 실제 게이트웨이) → Context.Recovery →
// Context.ReconcileDriver / Context.ExitObserver → engine.NewRuntime{Recover, 루프}. 브로커만 httptest 이고, 브로커가 받는 주문이
// 관측 대상이다.
//
//   - 만료(계좌 일치) 파일: 기동 복구가 끝나고 exit 관측이 원장의 보유에 손절 매도를 제출한다.
//   - 7(f) endpoint 누락 파일(만료 아님 — 2026-11-23 운영에서 난 모양: POST 주문 · 취소가 빠진 유효 기간 안 파일): 같은 조립에서
//     게이트웨이로 낸 매수는 래치 사유로 거절(브로커 매수 0), 손절 매도는 제출.
//   - 7(g) 편입: 대사 주기가 원장에 없는 보유(재기동 뒤 사람이 산 보유)를 편입하고, 그 포지션에 exit 관측이 붙어 손절이 나간다 —
//     이 change 의 목적.
//
// 시계: 공식 클라이언트는 시세의 FetchedAt 을 벽시계로 찍고 exit 관측은 「아직 오지 않은 시세」와 15초보다 오래된 시세를 거름
// (exitloop observe · quoteUsable). 그래서 가짜 시계를 벽시계보다 2초 앞에 두고 벽시계를 따라가게 함.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a199Broker 는 청산 전용 엔진이 기동 복구 · 대사 · exit 관측에서 부르는 읽기와 주문 생성에 답함. 보유 10주(005930).
type a199Broker struct {
	mu     sync.Mutex
	price  string
	orders []map[string]any
	ids    map[string]string
	next   int
}

func newA199Broker(t *testing.T) (*a199Broker, *httptest.Server) {
	t.Helper()
	b := &a199Broker{price: "70100", ids: map[string]string{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case r.URL.Path == "/api/v1/prices":
			b.mu.Lock()
			price := b.price
			b.mu.Unlock()
			_, _ = w.Write([]byte(`{"result":[{"symbol":"005930","lastPrice":"` + price + `","currency":"KRW"}]}`))
		case r.URL.Path == "/api/v1/holdings":
			_, _ = w.Write([]byte(`{"result":{"items":[{"symbol":"005930","marketCountry":"KR","currency":"KRW",` +
				`"quantity":"10","averagePurchasePrice":"70000","lastPrice":"70000"}]}}`))
		case r.URL.Path == "/api/v1/sellable-quantity":
			_, _ = w.Write([]byte(`{"result":{"symbol":"005930","sellableQuantity":"10"}}`))
		case r.URL.Path == "/api/v1/buying-power":
			_, _ = w.Write([]byte(`{"result":{"cashBuyingPower":"100000000","currency":"KRW"}}`))
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			b.mu.Lock()
			b.next++
			id := fmt.Sprintf("BO-%d", b.next)
			b.orders = append(b.orders, body)
			symbol, _ := body["symbol"].(string)
			b.ids[id] = symbol
			b.mu.Unlock()
			key, _ := body["clientOrderId"].(string)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"orderId": id, "clientOrderId": key}})
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":"","hasNext":false}}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/") && r.Method == http.MethodGet:
			id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
			b.mu.Lock()
			symbol, ok := b.ids[id]
			b.mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"orderId": id, "symbol": symbol, "status": "OPEN"}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return b, srv
}

func (b *a199Broker) setPrice(p string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.price = p
}

func (b *a199Broker) side(side string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []map[string]any
	for _, o := range b.orders {
		if o["side"] == side {
			out = append(out, o)
		}
	}
	return out
}

// syncBuffer 는 런타임 goroutine 들이 함께 쓰는 로그 버퍼(경합 없이 읽게).
type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

type a199Stack struct {
	broker *a199Broker
	eng    *engine.Context
	clk    *clock.Fake
	logs   *syncBuffer
	logger *obs.Logger
}

// a199OpenExitOnly 는 낡은 attestation(mutate 가 정함)과 설정(cfg 가 고침)으로 생산 조립을 엶 — 청산 전용이어야 함.
// 원장에는 이 계좌로 엔진 주문이 확정된 기록이 있음(a199 G2 — a199IsolateWithEngineOrder).
func a199OpenExitOnly(t *testing.T, mutate func(a *attest.Attestation, now time.Time), cfg func(*config.File), protectionReady bool) a199Stack {
	t.Helper()
	broker, srv := newA199Broker(t)
	s, err := a199Boot(t, a199IsolateWithEngineOrder(t), broker, srv, mutate, cfg, protectionReady)
	if err != nil {
		t.Fatalf("a stale attestation must start exit-only: %v", err)
	}
	if s.eng.Automation.Verified || !s.eng.Automation.ExitOnly {
		t.Fatalf("automation = %+v, want an exit-only start", s.eng.Automation)
	}
	return s
}

// a199Boot 은 dir 의 설정 · 원장으로 생산 조립을 한 번 엶(같은 dir 로 다시 부르면 재기동).
func a199Boot(t *testing.T, dir string, broker *a199Broker, srv *httptest.Server, mutate func(a *attest.Attestation, now time.Time),
	cfg func(*config.File), protectionReady bool,
) (a199Stack, error) {
	t.Helper()
	file := config.DefaultFile()
	file.Trading = openTradingPolicy()
	file.Engine.AutomationGate = smallLiveGate()
	if cfg != nil {
		cfg(&file)
	}
	data, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	clk := clock.NewFake(time.Now().UTC().Add(2 * time.Second))
	writeAttestation(t, dir, func(a *attest.Attestation) { mutate(a, clk.Now()) })

	logs := &syncBuffer{}
	logger := obs.NewLogger(obs.LogOptions{Writer: logs, JSON: true, Clock: clk})
	opts := engine.Options{
		ConfigDir: dir, Clock: clk, Logger: logger,
		OfficialOptions: []official.Option{official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client())},
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}))
	if protectionReady {
		opts.SetProtectionReadyForTest()
	}
	eng, err := engine.New(opts)
	if err != nil {
		return a199Stack{}, err
	}
	t.Cleanup(func() { _ = eng.Close() })
	return a199Stack{broker: broker, eng: eng, clk: clk, logs: logs, logger: logger}, nil
}

// run 은 기동 복구 뒤 loops 를 돌리고 cond 가 참이 되거나 30초가 지나면 멈춤. 런타임이 먼저 끝나면 그 오류로 실패(hang 없음 — 처분 4).
func (s a199Stack) run(t *testing.T, loops []engine.SupervisedLoop, cond func() bool) {
	t.Helper()
	recovery, err := s.eng.Recovery(reconcile.Options{Clock: s.clk})
	if err != nil {
		t.Fatalf("Recovery: %v", err)
	}
	recovered := make(chan struct{})
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: s.eng.AccountRef, Clock: s.clk, Log: s.logger,
		Recover: func(ctx context.Context) error {
			_, err := recovery.Run(ctx)
			if err == nil {
				close(recovered)
			}
			return err
		},
		Loops: loops,
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	finished := false
	defer func() {
		cancel()
		if !finished {
			<-done
		}
	}()
	deadline := time.Now().Add(30 * time.Second)
	for !cond() && time.Now().Before(deadline) {
		select {
		case err := <-done:
			finished = true
			t.Fatalf("runtime ended early: %v; engine log:\n%s", err, s.logs.String())
		default:
		}
		// 가짜 시계를 벽시계 + 2초로 따라가게 함 — 시세 증거(벽시계 FetchedAt)가 15초 창 안에 들게(exitloop observe).
		s.clk.Set(time.Now().UTC().Add(2 * time.Second))
		time.Sleep(50 * time.Millisecond)
	}
	select {
	case <-recovered:
	default:
		t.Fatal("the boot recovery did not complete")
	}
}

func (s a199Stack) exitLoop(t *testing.T) engine.SupervisedLoop {
	t.Helper()
	observer, err := s.eng.ExitObserver(engine.ExitObserverOptions{Clock: s.clk, Costs: costs.DefaultModel(), Log: s.logger})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	return engine.SupervisedLoop{Name: "exit", Run: observer.Run}
}

func a199ExpiredFile(a *attest.Attestation, now time.Time) {
	a.IssuedAt = now.Add(-31 * 24 * time.Hour)
	a.ExpiresAt = now.Add(-time.Hour)
}

// a199ScopeLostFile 은 2026-11-23 운영 모양 — 유효 기간 안, POST 주문 · 취소가 빠짐.
func a199ScopeLostFile(a *attest.Attestation, now time.Time) {
	a.IssuedAt = now.Add(-24 * time.Hour)
	a.ExpiresAt = now.Add(29 * 24 * time.Hour)
	var kept []string
	for _, e := range engine.RequiredEndpoints() {
		if !strings.HasPrefix(e, "POST ") {
			kept = append(kept, e)
		}
	}
	a.Endpoints = kept
}

func TestA199AnExitOnlyRuntimeRecoversAndSubmitsTheStopLoss(t *testing.T) {
	s := a199OpenExitOnly(t, a199ExpiredFile, nil, false)
	stack := &e2eStack{t: t, engine: s.eng, clk: s.clk}
	stack.entry("005930", "10", "70000", "68000") // 보유: 원장의 진입 체결(손절 68000)
	s.broker.setPrice("67000")                    // 손절선 아래 — 첫 관측 주기가 손절을 낸다
	s.run(t, []engine.SupervisedLoop{s.exitLoop(t)}, func() bool { return len(s.broker.side("SELL")) > 0 })
	sells := s.broker.side("SELL")
	if len(sells) == 0 {
		t.Fatalf("no stop-loss sell reached the broker from the exit-only runtime; engine log:\n%s", s.logs.String())
	}
	if fmt.Sprint(sells[0]["symbol"]) != "005930" {
		t.Fatalf("sell = %+v, want 005930", sells[0])
	}
	if _, latched := s.eng.Entry.Blocks()[a199StaleReason]; !latched {
		t.Error("the stale-attestation entry latch is gone while the runtime runs")
	}
}

// 처분 7(f): endpoint 누락(운영 모양) 파일 — 같은 조립에서 매수는 거절(브로커 매수 0), 손절 매도는 제출.
// 생산 조립의 게이트웨이는 보호 readiness 를 생산 공급자에서 받으므로(SetProtectionReadyForTest 는 보고 칸만 바꿈) 이 조립의 매수
// 거절 사유는 보호 검사(`broker_protection_not_wired`)가 먼저다 — 그 뒤에 서 있는 래치는 같은 게이트웨이가 묻는 EntryGate 로 단언하고,
// 래치 사유가 매수 거절 사유가 되는 것은 보호 WIRED 게이트웨이 시험(execgw TestA199AWiredGateway…)이 잰다.
func TestA199AScopeLostRuntimeRefusesABuyAndSubmitsTheStopLoss(t *testing.T) {
	s := a199OpenExitOnly(t, a199ScopeLostFile, nil, false)
	if s.eng.Automation.AttestationStale != "scope_lost" {
		t.Fatalf("stale = %q, want scope_lost", s.eng.Automation.AttestationStale)
	}
	issuer := &execgw.Issuer{Journal: s.eng.Journal, Clock: s.clk, AccountRef: s.eng.AccountRef, TTL: 30 * time.Second}
	decision, err := issuer.IssueEntry(context.Background(), execgw.EntryRequest{
		Market: "kr", Symbol: "000660", Side: "buy", Quantity: 1, EntryPrice: 100000, StopPrice: 95000,
		PolicyVersion: "a199-r1", Limits: s.eng.Automation.Limits,
	})
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}
	out, err := s.eng.Gateway.Place(context.Background(), execgw.PlaceRequest{
		Intent: orderintent.PlaceIntent{Symbol: "000660", Market: "kr", Side: "buy", OrderType: "limit",
			Quantity: 1, Price: 100000, CurrencyMode: "KRW"},
		Decision: decision,
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || (rejected.Reason != execgw.ReasonProtectionNotWired && rejected.Reason != a199StaleReason) {
		t.Fatalf("buy = %+v, %v, want a gateway refusal", out, err)
	}
	if r := s.eng.Entry.CheckEntryFor("kr", "000660"); r == nil || r.Reason != a199StaleReason {
		t.Fatalf("the gateway's entry gate = %v, want the stale-attestation latch behind the protection check", r)
	}
	if buys := s.broker.side("BUY"); len(buys) != 0 {
		t.Fatalf("broker saw %d buy(s) from an exit-only engine", len(buys))
	}

	stack := &e2eStack{t: t, engine: s.eng, clk: s.clk}
	stack.entry("005930", "10", "70000", "68000")
	s.broker.setPrice("67000")
	s.run(t, []engine.SupervisedLoop{s.exitLoop(t)}, func() bool { return len(s.broker.side("SELL")) > 0 })
	if len(s.broker.side("SELL")) == 0 {
		t.Fatalf("no stop-loss sell from the scope-lost exit-only runtime; engine log:\n%s", s.logs.String())
	}
	if buys := s.broker.side("BUY"); len(buys) != 0 {
		t.Fatalf("broker saw %d buy(s) from an exit-only engine", len(buys))
	}
}

// 처분 7(g): 청산 전용 런타임에서 편입이 돈다 — 대사 주기가 원장에 없는 보유를 편입하고, 그 포지션에 exit 관측이 붙어 손절이 나감.
func TestA199AnExitOnlyRuntimeAdoptsAHoldingAndProtectsIt(t *testing.T) {
	s := a199OpenExitOnly(t, a199ExpiredFile, func(f *config.File) {
		f.Engine.Adoption = config.Adoption{Enabled: true, DefaultStopPct: 0.05}
	}, false)
	driver, err := s.eng.ReconcileDriver(engine.ReconcileDriverOptions{
		Collector: s.eng.SnapshotCollector(s.clk), Clock: s.clk, DefaultMarket: "kr",
		Period: time.Second, StabilisationInterval: time.Second,
	})
	if err != nil {
		t.Fatalf("ReconcileDriver: %v", err)
	}
	adopted := func() bool {
		p, err := s.eng.Journal.CurrentPosition(context.Background(), s.eng.AccountRef, "kr", "005930")
		if err != nil || !p.Adopted() {
			return false
		}
		s.broker.setPrice("60000") // 편입 뒤 손절선(관측가의 95%) 아래로
		return len(s.broker.side("SELL")) > 0
	}
	s.run(t, []engine.SupervisedLoop{{Name: "reconcile", Run: driver.Run}, s.exitLoop(t)}, adopted)
	p, err := s.eng.Journal.CurrentPosition(context.Background(), s.eng.AccountRef, "kr", "005930")
	if err != nil || !p.Adopted() {
		t.Fatalf("the exit-only runtime did not adopt the unmanaged holding: %+v, %v; engine log:\n%s", p, err, s.logs.String())
	}
	if len(s.broker.side("SELL")) == 0 {
		t.Fatalf("the adopted holding got no stop-loss sell; engine log:\n%s", s.logs.String())
	}
}
