package engine_test

// a198 — 보호 청산의 길 치우기가 내는 취소가 **실제 주문 경로**(게이트웨이 + 원장)를 지나 브로커에 닿는지 잼.
//
// 가짜 제출자를 쓰지 않음: 엔진 조립(engine.New)이 만든 진짜 execgw.Gateway · 진짜 원장(임시 경로) · 운영과 같은
// RiskGuardian 위에 가짜 브로커(httptest)만 둠. 기존 진짜-게이트웨이 하네스(exit_e2e 의 buildE2EStack · entry · fill)를
// 무편집으로 재사용하고, 브로커만 취소 끝점을 답하도록 감쌈. 픽스처는 합성 값만.

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// a198Broker 는 e2eBroker 의 시세 · 발주 · 되읽기를 그대로 쓰고 취소 끝점(POST /api/v1/orders/{id}/cancel)만 더함.
type a198Broker struct {
	*e2eBroker
	cmu          sync.Mutex
	cancelled    []string // 취소 요청을 받은 주문 번호(순서대로)
	cancelStatus int      // 0 이면 200, 아니면 그 HTTP 상태로 답함
	// cancelBody 는 거절 본문임 — 비면 합성 본문. 계약 모양(`{"error":{…,"code":…}}`)을 시험이 넣음.
	cancelBody string
	// cancelPlan 은 주문 번호별 다음 취소 응답(상태, 본문)의 대기열임 — 비면 cancelStatus·cancelBody.
	cancelPlan map[string][]a198Reply
	// sellRefusals 는 다음 매도 발주를 그 수만큼 409 opposite-pending-order-exists 로 거절함(리뷰 탐침 P2).
	sellRefusals int
}

// a198Reply 는 취소 응답 하나임(status 0 = 접수).
type a198Reply struct {
	status int
	body   string
}

// a198AlreadyBody 는 계약(openapi 409 예시)과 같은 모양의 거절 본문임 — code 는 `error.code` 자리.
func a198AlreadyBody(code string) string {
	return `{"error":{"requestId":"r-a198","code":"` + code + `","message":"synthetic"}}`
}

func (b *a198Broker) planCancel(orderID string, replies ...a198Reply) {
	b.cmu.Lock()
	defer b.cmu.Unlock()
	if b.cancelPlan == nil {
		b.cancelPlan = map[string][]a198Reply{}
	}
	b.cancelPlan[orderID] = append(b.cancelPlan[orderID], replies...)
}

func (b *a198Broker) refuseSells(n int) {
	b.cmu.Lock()
	defer b.cmu.Unlock()
	b.sellRefusals = n
}

// serveCreateA198 는 매도 발주 거절 대기열이 있으면 409 opposite-pending 으로 답하고, 아니면 e2eBroker 의 발주로 넘김.
func (b *a198Broker) serveCreateA198(t *testing.T, w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	b.cmu.Lock()
	refuse := b.sellRefusals > 0 && strings.Contains(string(raw), `"side":"SELL"`)
	if refuse {
		b.sellRefusals--
	}
	b.cmu.Unlock()
	if refuse {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(a198AlreadyBody("opposite-pending-order-exists")))
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	b.serveCreate(t, w, r)
}

func newA198Broker(t *testing.T) (*a198Broker, *httptest.Server) {
	t.Helper()
	b := &a198Broker{e2eBroker: &e2eBroker{last: map[string]string{}, symbols: map[string]string{}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case r.URL.Path == "/api/v1/prices":
			b.servePrices(w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodPost:
			b.serveCreateA198(t, w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":"","hasNext":false}}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/") && strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == http.MethodPost:
			b.serveCancel(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/") && r.Method == http.MethodGet:
			b.serveReadBack(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return b, srv
}

func (b *a198Broker) serveCancel(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/orders/"), "/cancel")
	b.cmu.Lock()
	b.cancelled = append(b.cancelled, id)
	status, body := b.cancelStatus, b.cancelBody
	if plan := b.cancelPlan[id]; len(plan) > 0 {
		status, body = plan[0].status, plan[0].body
		b.cancelPlan[id] = plan[1:]
	}
	b.cmu.Unlock()
	if status != 0 {
		if body == "" {
			body = `{"error":{"code":"a198-synthetic","message":"synthetic refusal"}}`
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"orderId": "C-" + id}})
}

func (b *a198Broker) cancels() []string {
	b.cmu.Lock()
	defer b.cmu.Unlock()
	return append([]string(nil), b.cancelled...)
}

func (b *a198Broker) refuseCancels(status int) {
	b.cmu.Lock()
	defer b.cmu.Unlock()
	b.cancelStatus = status
}

// a198Alerts 는 관측기의 운영자 알림(지연 경보)을 받음.
type a198Alerts struct {
	mu     sync.Mutex
	events []obs.Event
}

func (a *a198Alerts) Notify(_ context.Context, e obs.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, e)
	return nil
}

func (a *a198Alerts) keys() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	var out []string
	for _, e := range a.events {
		out = append(out, e.Key)
	}
	return out
}

// a198Critical 는 a094 의 명명 critical(연속 실패 경보)을 받음.
type a198Critical struct{ a198Alerts }

func (c *a198Critical) RecordCritical(ctx context.Context, e obs.Event, _ time.Duration) error {
	return c.Notify(ctx, e)
}

type a198Fixture struct {
	t        *testing.T
	broker   *a198Broker
	s        *e2eStack
	sink     *a129Sink
	alerts   *a198Alerts
	critical *a198Critical
	p        journal.Position
}

// newA198Fixture 는 진짜 게이트웨이 엔진(buildE2EStack) 위에 실패 기록 싱크 · 알림 수집기를 단 관측기를 세움.
func newA198Fixture(t *testing.T) *a198Fixture {
	t.Helper()
	broker, srv := newA198Broker(t)
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	s := buildE2EStack(t, broker.e2eBroker, srv, dir, false)
	f := &a198Fixture{t: t, broker: broker, s: s, sink: &a129Sink{}, alerts: &a198Alerts{}, critical: &a198Critical{}}
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: s.engine.Journal, Clock: s.clk, AccountRef: s.engine.AccountRef,
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "a198-test",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: s.engine.Journal, Prices: e2eClockDomainPrices{next: s.engine.Official, clk: s.clk}, Retrier: s.engine.Retrier,
		Issuer: guardian, Submit: s.engine.Gateway, Alerts: f.alerts, Critical: f.critical,
		ExitOrderLog: obs.NewLogger(obs.LogOptions{Writer: f.sink, JSON: true, Clock: s.clk}),
		Costs:        costs.DefaultModel(), AccountRef: s.engine.AccountRef, Clock: s.clk,
	})
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	s.observer = observer
	return f
}

// workingOrder 는 엔진 귀속 미체결 주문 하나를 원장에 남김 — 접수 확정된 PLACE attempt, 종결 체결 기록 없음.
// 되읽기(취소 전 사전 조회)가 답하도록 브로커에도 그 번호를 등록함.
func a198WorkingOrder(t *testing.T, s *e2eStack, b *a198Broker, symbol, side, orderID, quantity, price string) {
	t.Helper()
	ctx := context.Background()
	decisionID := "d-a198-" + orderID
	class := journal.SafetyClassRiskReducing
	var preimage journal.Preimage = journal.ReductionIntent{
		AccountRef: s.engine.AccountRef, Market: "kr", Symbol: symbol, Side: side,
		MaxQuantity: quantity, Reason: "a198 synthetic working order",
	}
	limits := ""
	if side == "BUY" {
		class = journal.SafetyClassExposureRaising
		preimage = journal.RiskIntent{
			AccountRef: s.engine.AccountRef, Market: "kr", Symbol: symbol, Side: "BUY",
			Quantity: quantity, EntryPrice: price, StopPrice: "60000", TargetPrice: "90000",
			PolicyVersion: "a198-test",
		}
		enc, err := execgw.EncodeLimits(execgw.Limits{
			MaxQuantity: execgw.Bound(100), MaxNotional: execgw.Bound(1_000_000),
			MaxTotalExposure: execgw.Bound(10_000_000), MaxDailyLossAmount: execgw.Bound(100_000),
			MaxDailyLossRatio: execgw.Bound(0.01), Currency: "KRW",
		})
		if err != nil {
			t.Fatalf("EncodeLimits: %v", err)
		}
		limits = enc
	}
	if _, err := s.engine.Journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: s.engine.AccountRef, SafetyClass: class, Kind: journal.KindPlace,
		Preimage: preimage, LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: s.clk.Now(), ExpiresAt: s.clk.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("RecordDecision(%s): %v", orderID, err)
	}
	attempt, err := s.engine.Journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "i-a198-" + orderID, Market: "kr", TradingDay: "2026-03-30",
			AccountRef: s.engine.AccountRef, Symbol: symbol, Side: side,
			OrderType: "LIMIT", TimeInForce: "DAY", Quantity: quantity, Price: price,
			Currency: "KRW", Source: "engine", Fingerprint: "fp-a198-" + orderID,
		},
		Kind: journal.KindPlace, AttemptID: "a-a198-" + orderID,
		AccountRef: s.engine.AccountRef, DecisionID: decisionID, SafetyClass: class,
		ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		t.Fatalf("Prepare(%s): %v", orderID, err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		t.Fatalf("Settle: %v", err)
	}
	b.e2eBroker.mu.Lock()
	b.e2eBroker.symbols[orderID] = symbol
	b.e2eBroker.mu.Unlock()
}

// openWithWorkingBuy 는 005930 10주(진입 70000, 손절 68000) 위에 엔진의 매수 미체결 W-BUY-1 을 두고 상태를 엶.
func (f *a198Fixture) openWithWorkingBuy() {
	f.t.Helper()
	f.p = f.s.entry("005930", "10", "70000", "68000")
	a198WorkingOrder(f.t, f.s, f.broker, "005930", "BUY", "W-BUY-1", "3", "69000")
	f.broker.quote("005930", "70100")
	f.s.observe()
}

// breach 는 손절선 아래를 관측시켜 보호 청산을 발의하게 함(한 주기).
func (f *a198Fixture) breach() {
	f.t.Helper()
	f.broker.quote("005930", "67900")
	f.s.observe()
}

func (f *a198Fixture) db() *sql.DB {
	f.t.Helper()
	db, err := sql.Open("sqlite", "file:"+f.s.engine.Journal.Path())
	if err != nil {
		f.t.Fatalf("open journal db: %v", err)
	}
	f.t.Cleanup(func() { _ = db.Close() })
	return db
}

// cancelAttempts 는 원장의 CANCEL attempt 를 「상태 | 대상 주문 | 결정 키 유무 | 결정 preimage」 로 돌려줌.
func (f *a198Fixture) cancelAttempts() []string {
	f.t.Helper()
	rows, err := f.db().Query(`SELECT a.state, a.target_order_id, coalesce(d.client_order_id,''), coalesce(d.risk_preimage,'')
		FROM mutation_attempts a LEFT JOIN decisions d ON d.id = a.decision_id
		WHERE a.kind = 'CANCEL' ORDER BY a.rowid`)
	if err != nil {
		f.t.Fatalf("query cancel attempts: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var state, target, key, preimage string
		if err := rows.Scan(&state, &target, &key, &preimage); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, fmt.Sprintf("%s|%s|key=%t|%s", state, target, key != "", preimage))
	}
	return out
}

// sells 는 브로커가 받은 매도 발주 본문임.
func (f *a198Fixture) sells() []map[string]any {
	var out []map[string]any
	for _, body := range f.broker.placed() {
		if body["side"] == "SELL" {
			out = append(out, body)
		}
	}
	return out
}

// failureLines 는 실패 기록 싱크(ExitOrderLog)의 취소 줄임.
func (f *a198Fixture) failureLines() []map[string]any {
	f.t.Helper()
	var out []map[string]any
	for _, l := range a129Parse(f.t, f.sink.String()) {
		if l["kind"] == "cancel" {
			out = append(out, l)
		}
	}
	return out
}

// --- §4-1 · 4-2 · 4-8: 엔진 매수 미체결 → 취소가 나가고 같은 주기에 손절 -----------------------------------------

func TestA198AWorkingEngineBuyIsCancelledAndTheStopFollowsInTheSameCycle(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.breach()

	if got := f.broker.cancels(); len(got) != 1 || got[0] != "W-BUY-1" {
		t.Fatalf("broker cancels = %v, want exactly one for W-BUY-1 — the clearing cancel never reached the broker; "+
			"cancel attempts in the ledger: %v", got, f.cancelAttempts())
	}
	attempts := f.cancelAttempts()
	if len(attempts) != 1 || !strings.HasPrefix(attempts[0], "CONFIRMED|W-BUY-1|key=false|") ||
		!strings.Contains(attempts[0], `"side":"BUY"`) {
		t.Fatalf("cancel attempts = %v, want one CONFIRMED on W-BUY-1 whose decision carries no idempotency key "+
			"and the order's side (BUY)", attempts)
	}
	sells := f.sells()
	if len(sells) != 1 || fmt.Sprint(sells[0]["quantity"]) != "10" {
		t.Fatalf("broker sells = %v, want the protective stop of 10 in the same cycle", sells)
	}
	// §4-8: 손절 제출의 결정은 종전대로 주문 생성 결정(멱등 키 있음).
	var key string
	if err := f.db().QueryRow(`SELECT coalesce(d.client_order_id,'') FROM mutation_attempts a
		JOIN decisions d ON d.id = a.decision_id JOIN intents i ON i.id = a.intent_id
		WHERE a.kind = 'PLACE' AND i.side = 'SELL' ORDER BY a.rowid DESC LIMIT 1`).Scan(&key); err != nil {
		t.Fatalf("stop decision: %v", err)
	}
	if key == "" {
		t.Fatal("the stop's decision carries no idempotency key — the submit path's issuance changed")
	}
	if lines := f.failureLines(); len(lines) != 0 {
		t.Fatalf("failure lines = %v on a clean cancel, want none", lines)
	}
}

// --- §4-3: 무장된 발의의 매도 미체결 → 취소 1, 종결 증거 대기, 종결 뒤 손절 --------------------------------------

func TestA198AnArmedSellIsCancelledOnceAndTheStopWaitsForItsClosingRecord(t *testing.T) {
	f := newA198Fixture(t)
	f.p = f.s.entry("005930", "10", "70000", "68000")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.quote("005930", "71000")
	f.s.observe()
	f.broker.quote("005930", "72000")
	f.s.observe() // +1.0R — 익절 매도 4 를 무장하고 브로커에 냄
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v, want the armed partial", f.sells())
	}
	partial := f.s.lastBrokerOrder()

	f.broker.quote("005930", "67000")
	f.s.observe()
	if got := f.broker.cancels(); len(got) != 1 || got[0] != partial {
		t.Fatalf("broker cancels = %v, want one for the armed sell %s; cancel attempts: %v", got, partial, f.cancelAttempts())
	}
	if a := f.cancelAttempts(); len(a) != 1 || !strings.HasPrefix(a[0], "CONFIRMED|"+partial+"|key=false|") || !strings.Contains(a[0], `"side":"SELL"`) {
		t.Fatalf("cancel attempts = %v, want one CONFIRMED on %s with a key-less SELL decision", a, partial)
	}
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v — a stop went out on top of a cancelled sell with no closing record", f.sells())
	}
	f.s.observe()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("broker cancels = %v after another cycle, want still one — an accepted cancel is not re-sent", got)
	}
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v, the stop must still wait for the closing record", f.sells())
	}

	// 종결 체결 기록(무체결 취소) — 체결 감지가 쓰는 입구 그대로.
	if _, err := f.s.engine.Journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: partial, AccountRef: f.s.engine.AccountRef, Symbol: "005930", Market: "kr",
		TradingDay: "2026-03-30", Side: "SELL", State: "CANCELLED", Terminal: true,
		Quantity: "4", FilledQuantity: "0", AveragePrice: "0",
		ObservedAt: f.s.clk.Now().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("RecordFill(closing): %v", err)
	}
	f.s.observe()
	sells := f.sells()
	if len(sells) != 2 || fmt.Sprint(sells[1]["quantity"]) != "10" {
		t.Fatalf("broker sells = %v, want the stop of 10 after the closing record", sells)
	}
}

// --- §4-4: 브로커가 취소를 거절 → 손절 없음, 실패 줄 1, 반복에도 1, 결과가 바뀌면 2 --------------------------------

func TestA198ARefusedCancelIsLoggedOnceWithoutAccountQuantityPriceOrDetail(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.refuseCancels(http.StatusBadRequest)
	f.breach()
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v, a stop went out over an uncleared buy", f.sells())
	}
	lines := f.failureLines()
	if len(lines) != 1 {
		t.Fatalf("failure lines = %v (sink %q), want exactly one", lines, f.sink.String())
	}
	if lines[0]["order_id"] != "W-BUY-1" || lines[0]["side"] != "BUY" || lines[0]["to_state"] != string(journal.StateFailedConfirmed) ||
		lines[0]["event"] != string(obs.EventOrderRejected) {
		t.Fatalf("failure line = %v, want W-BUY-1 · BUY · FAILED_CONFIRMED · order.rejected", lines[0])
	}
	f.breach()
	if got := f.broker.cancels(); len(got) != 2 {
		t.Fatalf("broker cancels = %v, want the next cycle to try again", got)
	}
	if lines := f.failureLines(); len(lines) != 1 {
		t.Fatalf("failure lines = %d after the same refusal repeated, want still one", len(lines))
	}
	f.broker.refuseCancels(http.StatusInternalServerError)
	f.breach()
	lines = f.failureLines()
	if len(lines) != 2 || lines[1]["to_state"] != string(journal.StateInDoubt) {
		t.Fatalf("failure lines = %v, want a second line once the outcome changed to IN_DOUBT", lines)
	}
	raw := f.sink.String()
	t.Logf("failure lines (raw):\n%s", raw)
	for _, forbidden := range []string{f.s.engine.AccountRef, "123-45", `"quantity"`, `"price"`, `"detail"`, "synthetic refusal"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("failure lines carry %q: %s", forbidden, raw)
		}
	}
}

// --- §4-5: 취소 결정 발행 실패 → 취소 없음, 손절 없음, 줄 1 ----------------------------------------------------------

func TestA198AnUnrecordableCancelDecisionIsLoggedOnce(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	if _, err := f.db().Exec(`CREATE TRIGGER a198_refuse_reductions BEFORE INSERT ON decisions
		WHEN NEW.safety_class = 'RISK_REDUCING' BEGIN SELECT RAISE(ABORT, 'a198 injected decision write failure'); END`); err != nil {
		t.Fatalf("install trigger: %v", err)
	}
	f.breach()
	f.breach()
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v, a cancel went out without a recorded decision", got)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v, a stop went out over an uncleared buy", f.sells())
	}
	lines := f.failureLines()
	if len(lines) != 1 || lines[0]["reason"] != "cancel_decision_failed" || lines[0]["to_state"] != "" {
		t.Fatalf("failure lines = %v, want one cancel_decision_failed line with no state", lines)
	}
	t.Logf("failure lines (raw):\n%s", f.sink.String())
	if strings.Contains(f.sink.String(), "injected") {
		t.Fatalf("the raw error reached the line: %s", f.sink.String())
	}
}

// --- §4-6: 모호(IN_DOUBT) 취소와 ACKED 고착 → 손절 보류, 지연 경보가 실제 관측 주기에서 나감 -------------------------

func TestA198AnInDoubtCancelHoldsTheStopAndBothAlertsFire(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.refuseCancels(http.StatusInternalServerError)
	for i := 0; i < 3; i++ {
		f.breach()
		f.s.clk.Advance(11 * time.Second)
	}
	f.breach()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("broker cancels = %v, want one — an ambiguous cancel is not re-sent in session", got)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v, a stop went out over an IN_DOUBT cancel", f.sells())
	}
	delayKey := string(obs.EventExitLiquidationDelayed) + "|" + f.p.ID
	if keys := f.alerts.keys(); !strings.Contains(strings.Join(keys, ","), delayKey) {
		t.Fatalf("operator alerts = %v, want the liquidation-delay alert %s", keys, delayKey)
	}
	if keys := f.critical.keys(); !strings.Contains(strings.Join(keys, ","), "|streak:") {
		t.Fatalf("critical records = %v, want the consecutive-failure alert", keys)
	}
}

func TestA198AnAckedCancelHoldsTheStopAndTheDelayAlertStillFires(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	// 접수 뒤 최종 쓰기 전 중단의 원장 모양: W-BUY-1 을 겨눈 CANCEL attempt 가 ACKED 에 남음.
	attempt, err := f.s.engine.Journal.Prepare(context.Background(), journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "i-a198-acked-cancel", Market: "kr", TradingDay: "2026-03-30",
			AccountRef: f.s.engine.AccountRef, Symbol: "005930", Side: "BUY",
			OrderType: "LIMIT", Quantity: "3", Price: "69000", Currency: "KRW", Source: "engine",
			Fingerprint: "fp-a198-acked-cancel",
		},
		Kind: journal.KindCancel, AttemptID: "a-a198-acked-cancel", TargetOrderID: "W-BUY-1",
	})
	if err != nil {
		t.Fatalf("Prepare(cancel): %v", err)
	}
	if err := attempt.MarkDispatchStarted(context.Background()); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(context.Background(), "C-W-BUY-1"); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	for i := 0; i < 4; i++ {
		f.breach()
		f.s.clk.Advance(11 * time.Second)
	}
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v, want none while the ACKED cancel is unsettled", got)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v, a stop went out over an ACKED cancel", f.sells())
	}
	delayKey := string(obs.EventExitLiquidationDelayed) + "|" + f.p.ID
	if keys := f.alerts.keys(); !strings.Contains(strings.Join(keys, ","), delayKey) {
		t.Fatalf("operator alerts = %v, want the liquidation-delay alert %s — an ACKED cancel must not hold the stop silently", keys, delayKey)
	}
	if keys := f.critical.keys(); strings.Contains(strings.Join(keys, ","), "|streak:") {
		t.Fatalf("critical records = %v — an in-flight engine cancel is not counted (a094 D−2.7)", keys)
	}
}

// --- §4-7: 운영 조립(Context.ExitObserver)이 지도자인 실행 ----------------------------------------------------------

func TestA198TheProductionAssemblyCancelsTheWorkingBuyAndPlacesTheStop(t *testing.T) {
	broker, srv := newA198Broker(t)
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	sink := &a129Sink{}
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{
		Costs: costs.DefaultModel(), Clock: clock.System(),
		ExitOrderLog: obs.NewLogger(obs.LogOptions{Writer: sink, JSON: true}),
	})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	if observer.OptionsForTest().ExitOrderLog == nil {
		t.Fatal("the production assembly dropped ExitOrderLog — clearing failures would leave no line")
	}
	s := &e2eStack{t: t, broker: broker.e2eBroker, engine: eng, observer: observer, clk: clock.NewFake(time.Now().UTC())}
	s.entry("005930", "10", "70000", "68000")
	a198WorkingOrder(t, s, broker, "005930", "BUY", "W-BUY-1", "3", "69000")
	broker.quote("005930", "70100")
	if cycle := observer.ObserveOnce(context.Background()); cycle.Err != nil {
		t.Fatalf("assembled cycle: %v", cycle.Err)
	}
	broker.quote("005930", "67900")
	if cycle := observer.ObserveOnce(context.Background()); cycle.Err != nil {
		t.Fatalf("assembled cycle: %v", cycle.Err)
	}
	if got := broker.cancels(); len(got) != 1 || got[0] != "W-BUY-1" {
		t.Fatalf("broker cancels = %v under the production assembly, want one for W-BUY-1 (lines %q)", got, sink.String())
	}
	var sells int
	for _, body := range broker.placed() {
		if body["side"] == "SELL" {
			sells++
		}
	}
	if sells != 1 {
		t.Fatalf("broker sells = %d under the production assembly, want the stop", sells)
	}
}

// --- 보강(4b, D7·D8·D4a) -------------------------------------------------------------------------------------

// attemptRows 는 원장의 모든 attempt 를 「종류|상태|대상|사유」 로 돌려줌.
func (f *a198Fixture) attemptRows() []string {
	f.t.Helper()
	rows, err := f.db().Query(`SELECT kind, state, coalesce(target_order_id,''), coalesce(reason_code,'')
		FROM mutation_attempts ORDER BY rowid`)
	if err != nil {
		f.t.Fatalf("query attempts: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var k, s, tgt, r string
		if err := rows.Scan(&k, &s, &tgt, &r); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, k+"|"+s+"|"+tgt+"|"+r)
	}
	return out
}

// cancelStates 는 CANCEL attempt 의 「상태|사유」 목록임.
func (f *a198Fixture) cancelStates() []string {
	var out []string
	for _, row := range f.attemptRows() {
		parts := strings.Split(row, "|")
		if parts[0] == "CANCEL" {
			out = append(out, parts[1]+"|"+parts[3])
		}
	}
	return out
}

// terminal 은 체결 감지가 쓰는 입구로 그 주문의 종결 체결 기록을 남김.
func (f *a198Fixture) terminal(orderID, side, state, ordered, filled string) {
	f.t.Helper()
	if _, err := f.s.engine.Journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, AccountRef: f.s.engine.AccountRef, Symbol: "005930", Market: "kr",
		TradingDay: "2026-03-30", Side: side, State: state, Terminal: true,
		Quantity: ordered, FilledQuantity: filled, AveragePrice: "69000",
		ObservedAt: f.s.clk.Now().Format(time.RFC3339),
	}); err != nil {
		f.t.Fatalf("RecordFill(%s): %v", orderID, err)
	}
}

// next 는 다음 관측 주기(시계를 조금 움직이고 손절선 아래 관측)임.
func (f *a198Fixture) next() {
	f.t.Helper()
	f.s.clk.Advance(6 * time.Second)
	f.breach()
}

// §4-10 (리뷰 탐침 P1): 손절 발의 때 그 매수가 이미 체결 → 409 already-filled → 확정 거절, 종결 기록 뒤 손절.
func TestA198AnAlreadyFilledBuyIsADefinitiveRefusalAndTheStopFollowsItsClosingRecord(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.planCancel("W-BUY-1", a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-filled")})
	f.breach()
	if got := f.cancelStates(); len(got) != 1 || got[0] != string(journal.StateFailedConfirmed)+"|"+journal.ReasonDispatchRejected {
		t.Fatalf("cancel attempts = %v (all %v), want one FAILED_CONFIRMED|%s — an already-filled cancel is not in doubt",
			got, f.attemptRows(), journal.ReasonDispatchRejected)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v before the buy's closing record", f.sells())
	}
	f.terminal("W-BUY-1", "BUY", "FILLED", "3", "3")
	f.next()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v after the buy's closing record, want the stop — attempts %v", f.sells(), f.attemptRows())
	}
	t.Logf("P1 attempts=%v cancels=%v sells=%d", f.attemptRows(), f.broker.cancels(), len(f.sells()))
}

// §4-11 (P2) · §4-14 (D8): 매수 취소 접수 → 손절이 반대 미체결로 거절 → 다음 주기 재취소 0 · 손절 재제출.
func TestA198AnAcceptedBuyCancelIsNotSentAgainAndTheStopIsResubmitted(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.refuseSells(1)
	f.breach()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("broker cancels = %v, want the first cancel", got)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v, the refused stop must not be recorded as placed", f.sells())
	}
	f.broker.planCancel("W-BUY-1", a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-canceled")})
	f.next()
	if got := f.broker.cancels(); len(got) != 1 {
		t.Fatalf("broker cancels = %v on the next cycle, want still one — an accepted buy cancel is not sent again (D8)", got)
	}
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v, want the stop resubmitted on the next cycle — attempts %v", f.sells(), f.attemptRows())
	}
	t.Logf("P2 attempts=%v cancels=%v sells=%d", f.attemptRows(), f.broker.cancels(), len(f.sells()))
}

// §4-12 (P3): 매수 둘 — 하나 400 · 하나 접수 → 다음 주기 거절된 쪽만 재시도, 접수된 쪽 재취소 0.
func TestA198OnlyTheRefusedBuyIsCancelledAgain(t *testing.T) {
	f := newA198Fixture(t)
	f.p = f.s.entry("005930", "10", "70000", "68000")
	a198WorkingOrder(t, f.s, f.broker, "005930", "BUY", "W-BUY-1", "3", "69000")
	a198WorkingOrder(t, f.s, f.broker, "005930", "BUY", "W-BUY-2", "2", "68500")
	f.broker.quote("005930", "70100")
	f.s.observe()
	f.broker.planCancel("W-BUY-1", a198Reply{status: http.StatusBadRequest})
	f.broker.planCancel("W-BUY-2", a198Reply{status: 0}, a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-canceled")})
	f.breach()
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v with W-BUY-1 still live", f.sells())
	}
	f.next()
	if got := f.broker.cancels(); strings.Join(got, ",") != "W-BUY-1,W-BUY-2,W-BUY-1" {
		t.Fatalf("broker cancels = %v, want W-BUY-1,W-BUY-2,W-BUY-1 — the accepted W-BUY-2 is not cancelled again", got)
	}
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v, want the stop once both buys are cleared — attempts %v", f.sells(), f.attemptRows())
	}
	t.Logf("P3 attempts=%v cancels=%v sells=%d", f.attemptRows(), f.broker.cancels(), len(f.sells()))
}

// §4-13b: 엔진 취소 already-processing → 다음 주기 재취소 already-canceled → 확정 거절 → 종결 기록 뒤 손절.
func TestA198AnAlreadyCanceledAfterAlreadyProcessingIsDefinitive(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.planCancel("W-BUY-1",
		a198Reply{status: http.StatusConflict, body: `{"error":{"code":"already-processing","data":{"retryAfterSeconds":1}}}`},
		a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-canceled")})
	f.breach()
	f.next()
	got := f.cancelStates()
	if len(got) != 2 || !strings.HasPrefix(got[0], string(journal.StateFailedConfirmed)+"|") ||
		got[1] != string(journal.StateFailedConfirmed)+"|"+journal.ReasonDispatchRejected {
		t.Fatalf("cancel attempts = %v, want already-processing then already-canceled both FAILED_CONFIRMED (the second %s)",
			got, journal.ReasonDispatchRejected)
	}
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v before the buy's closing record", f.sells())
	}
	f.terminal("W-BUY-1", "BUY", "CANCELLED", "3", "0")
	f.next()
	if len(f.sells()) != 1 {
		t.Fatalf("broker sells = %v after the closing record, want the stop", f.sells())
	}
}

// §4-6/13: already-modified 는 종전대로 결과 불명 — 정정의 후계 주문은 원장이 모름.
func TestA198AnAlreadyModifiedCancelStaysInDoubt(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.planCancel("W-BUY-1", a198Reply{status: http.StatusConflict, body: a198AlreadyBody("already-modified")})
	f.breach()
	if got := f.cancelStates(); len(got) != 1 || !strings.HasPrefix(got[0], string(journal.StateInDoubt)+"|") {
		t.Fatalf("cancel attempts = %v, want IN_DOUBT for already-modified", got)
	}
}

// §4-15 (리뷰 P1-2): 취소 결정만 실패(손절의 결정은 기록됨) → 손절 미제출.
func TestA198OnlyTheCancelDecisionFailingStillWithholdsTheStop(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	// 취소 결정은 키가 없고 손절 결정은 키가 있음 — 키 없는 위험 감소 결정만 막음.
	if _, err := f.db().Exec(`CREATE TRIGGER a198_refuse_cancel_decisions BEFORE INSERT ON decisions
		WHEN NEW.safety_class = 'RISK_REDUCING' AND NEW.client_order_id IS NULL
		BEGIN SELECT RAISE(ABORT, 'a198 injected cancel decision failure'); END`); err != nil {
		t.Fatalf("install trigger: %v", err)
	}
	f.breach()
	if len(f.sells()) != 0 {
		t.Fatalf("broker sells = %v — a stop went out although the working buy was never cancelled", f.sells())
	}
	if got := f.broker.cancels(); len(got) != 0 {
		t.Fatalf("broker cancels = %v without a recorded cancel decision", got)
	}
}

// §4-16: 결정의 최대 수량은 정확히 그 주문의 수량.
func TestA198TheCancelDecisionCeilingIsTheOrderQuantity(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.breach()
	if a := f.cancelAttempts(); len(a) != 1 || !strings.Contains(a[0], `"max_quantity":"3"`) {
		t.Fatalf("cancel attempts = %v, want the decision's max_quantity to be the order's 3", a)
	}
}

// §4-16 · D4a: 수량·가격 해석 실패 갈래도 실패 줄 한 번.
func TestA198AnUnreadableWorkingOrderIsLoggedOnce(t *testing.T) {
	for _, column := range []string{"quantity", "price"} {
		t.Run(column, func(t *testing.T) {
			f := newA198Fixture(t)
			f.openWithWorkingBuy()
			if _, err := f.db().Exec(`UPDATE intents SET ` + column + ` = 'not-a-number' WHERE id = 'i-a198-W-BUY-1'`); err != nil {
				t.Fatalf("corrupt %s: %v", column, err)
			}
			f.breach()
			f.next()
			if len(f.sells()) != 0 {
				t.Fatalf("broker sells = %v over an unreadable working buy", f.sells())
			}
			lines := f.failureLines()
			if len(lines) != 1 || lines[0]["order_id"] != "W-BUY-1" || lines[0]["reason"] == "" {
				t.Fatalf("failure lines = %v, want one line with a fixed reason for W-BUY-1", lines)
			}
		})
	}
}

// §4-16 · D4a: 첫 로그 쓰기가 패닉하면 표시가 남지 않아 다음 주기에 다시 씀.
func TestA198AFailureLineWhoseWritePanickedIsWrittenNextCycle(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.refuseCancels(http.StatusBadRequest)
	f.sink.mu.Lock()
	f.sink.panics = true
	f.sink.mu.Unlock()
	f.breach()
	f.sink.mu.Lock()
	f.sink.panics = false
	f.sink.mu.Unlock()
	f.next()
	if lines := f.failureLines(); len(lines) != 1 {
		t.Fatalf("failure lines = %d after the panicked write, want the failure written on the next cycle", len(lines))
	}
}

// §4-16: 접수 확정 뒤 표시 정리 — 같은 주문이 나중에 다시 실패하면 줄이 다시 나감.
func TestA198AConfirmedCancelClearsTheFailureMark(t *testing.T) {
	f := newA198Fixture(t)
	f.openWithWorkingBuy()
	f.broker.planCancel("W-BUY-1", a198Reply{status: http.StatusBadRequest}, a198Reply{status: 0}, a198Reply{status: http.StatusBadRequest})
	f.broker.refuseSells(1)
	f.breach() // 400 → 줄 1
	f.next()   // 접수 → 표시 삭제, 손절은 반대 미체결로 거절
	// 접수 확정 증거를 다른 거래일로 옮겨 D8 이 보지 못하게 함(전일 결속 잔여의 모양) — 다음 주기가 다시 취소하고 400.
	if _, err := f.db().Exec(`UPDATE intents SET trading_day = '2026-03-27'
		WHERE id IN (SELECT intent_id FROM mutation_attempts WHERE kind = 'CANCEL' AND state = 'CONFIRMED')`); err != nil {
		t.Fatalf("move the confirmed cancel: %v", err)
	}
	f.next()
	if got := f.broker.cancels(); len(got) != 3 {
		t.Fatalf("broker cancels = %v, want the third cancel once the confirmation is out of D8's scope", got)
	}
	if lines := f.failureLines(); len(lines) != 2 {
		t.Fatalf("failure lines = %d, want a second line after the mark was cleared by the confirmed cancel", len(lines))
	}
}
