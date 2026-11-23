package engine_test

// a199_exit_only_runtime_test.go — Manager 결정 9: 「생성자가 성공한다」가 아니라, 낡은 attestation 으로 청산 전용 기동한
// 엔진에서 **기동 복구(Recover)가 끝나고 exit 관측 주기가 돌아 손절 매도가 실제 게이트웨이로 브로커에 제출된다**를 본다.
//
// 조립은 생산과 같은 길: engine.New(게이트 ON · 생산 Guardian · 실제 원장 · 실제 게이트웨이) → Context.Recovery →
// Context.ExitObserver → engine.NewRuntime{Recover, exit 루프}. 브로커만 httptest. 브로커가 받는 매도 한 건이 관측 대상이다.
//
// 시계: 공식 클라이언트는 시세의 FetchedAt 을 벽시계로 찍고 exit 관측은 「아직 오지 않은 시세」와 15초보다 오래된 시세를 거름
// (exitloop observe · quoteUsable). 그래서 가짜 시계를 벽시계보다 2초 앞에 두고 벽시계를 따라가게 함.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a199Broker 는 청산 전용 엔진이 기동 복구와 exit 관측에서 부르는 읽기와 주문 생성에 답함. 보유 10주(005930).
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

func (b *a199Broker) sells() []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []map[string]any
	for _, o := range b.orders {
		if o["side"] == "SELL" {
			out = append(out, o)
		}
	}
	return out
}

func TestA199AnExitOnlyRuntimeRecoversAndSubmitsTheStopLoss(t *testing.T) {
	broker, srv := newA199Broker(t)
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	clk := clock.NewFake(time.Now().UTC().Add(2 * time.Second))
	writeAttestation(t, dir, func(a *attest.Attestation) {
		a.IssuedAt = clk.Now().Add(-31 * 24 * time.Hour)
		a.ExpiresAt = clk.Now().Add(-time.Hour) // 낡음: 만료(계좌 일치)
	})

	var logs syncBuffer
	logger := obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true, Clock: clk})
	opts := engine.Options{
		ConfigDir: dir, Clock: clk,
		OfficialOptions: []official.Option{official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client())},
		Logger:          logger,
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}))
	eng, err := engine.New(opts)
	if err != nil {
		t.Fatalf("a stale attestation must start exit-only: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })
	if eng.Automation.Verified || !eng.Automation.ExitOnly {
		t.Fatalf("automation = %+v, want an exit-only start", eng.Automation)
	}

	// 보유: 진입 체결이 원장에 있음(손절 68000). e2e 시험의 진입 고정물과 같은 원장 입구.
	stack := &e2eStack{t: t, engine: eng, clk: clk}
	stack.entry("005930", "10", "70000", "68000")

	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Clock: clk, Costs: costs.DefaultModel(), Log: logger})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	recovery, err := eng.Recovery(reconcile.Options{Clock: clk})
	if err != nil {
		t.Fatalf("Recovery: %v", err)
	}
	recovered := make(chan struct{})
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		AccountRef: eng.AccountRef, Clock: clk,
		Log: logger,
		Recover: func(ctx context.Context) error {
			_, err := recovery.Run(ctx)
			if err == nil {
				close(recovered)
			}
			return err
		},
		Loops: []engine.SupervisedLoop{{Name: "exit", Run: observer.Run}},
	})
	if err != nil {
		t.Fatalf("NewRuntime: %v", err)
	}

	broker.mu.Lock()
	broker.price = "67000" // 손절선 아래 — 첫 관측 주기가 손절을 낸다
	broker.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- rt.Run(ctx) }()
	defer func() {
		cancel()
		<-done
	}()

	deadline := time.Now().Add(20 * time.Second)
	for len(broker.sells()) == 0 && time.Now().Before(deadline) {
		select {
		case err := <-done:
			t.Fatalf("runtime ended before the stop was submitted: %v", err)
		default:
		}
		// 가짜 시계를 벽시계 + 2초로 따라가게 함 — 시세 증거(벽시계 FetchedAt)가 15초 창 안에 들게(exitloop observe).
		clk.Set(time.Now().UTC().Add(2 * time.Second))
		time.Sleep(50 * time.Millisecond)
	}
	select {
	case <-recovered:
	default:
		t.Fatal("the boot recovery did not complete")
	}
	sells := broker.sells()
	if len(sells) == 0 {
		t.Fatalf("no stop-loss sell reached the broker from the exit-only runtime; engine log:\n%s", logs.String())
	}
	if fmt.Sprint(sells[0]["symbol"]) != "005930" {
		t.Fatalf("sell = %+v, want 005930", sells[0])
	}
	if _, latched := eng.Entry.Blocks()[a199StaleReason]; !latched {
		t.Error("the stale-attestation entry latch is gone while the runtime runs")
	}
}
