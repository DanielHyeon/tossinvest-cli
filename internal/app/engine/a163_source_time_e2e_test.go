package engine_test

// a163 tasks 2.1~2.4 — 브로커 데이터 시각은 판정 자격을 바꾸지 않는다(편집 전 GREEN 인 고정 시험).
//
// 공식 어댑터를 httptest 브로커로 물려(exit_e2e_test.go 의 하네스 재사용) /api/v1/prices 본문의 `timestamp` 만 바꿔 가며
// 같은 보유 · 같은 가격 · 같은 받은 시각에서 판정 · 발의 · 제출이 대조군과 같은지 잰다. 어댑터가 그 필드 때문에 행을 버리거나
// 오류를 내면(변이 ⑦⑧) 판정이 0 이 되어 여기서 잡힘. 받은 시각(FetchedAt)은 하네스가 가짜 시계로 덮으므로 15 s 안이다.
//
// 기준 시각: 하네스의 e2eNow = 2026-03-30T01:00:00Z = 10:00 KST.

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// 브로커 데이터 시각 원문 조각들. 값은 `"timestamp":…` 이하 JSON 조각 그대로(키 없음은 빈 조각).
const (
	a163FreshStamp  = `"timestamp":"2026-03-30T09:59:58.000+09:00"` // 받은 시각보다 2초 앞 — 대조군
	a163StaleStamp  = `"timestamp":"2026-03-30T00:30:00.000+09:00"` // 9시간 30분 묵음(a112 결정 46 의 모양)
	a163FutureStamp = `"timestamp":"2026-03-30T10:05:00.000+09:00"` // 받은 시각보다 5분 미래(시계 차)
)

// a163PriceBroker 는 e2eBroker 의 주문 · 되읽기 응답을 그대로 쓰고 /api/v1/prices 만 원문 본문으로 답함.
type a163PriceBroker struct {
	*e2eBroker
	mu    sync.Mutex
	stamp string            // 모든 행에 붙는 timestamp 조각
	last  map[string]string // 종목 → lastPrice
}

func newA163Broker(t *testing.T) (*a163PriceBroker, *httptest.Server) {
	t.Helper()
	b := &a163PriceBroker{
		e2eBroker: &e2eBroker{last: map[string]string{}, symbols: map[string]string{}},
		last:      map[string]string{},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/accounts":
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case r.URL.Path == "/api/v1/prices":
			b.servePricesRaw(w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodPost:
			b.serveCreate(t, w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":"","hasNext":false}}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/") && r.Method == http.MethodGet:
			b.serveReadBack(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return b, srv
}

func (b *a163PriceBroker) servePricesRaw(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.e2eBroker.mu.Lock()
	b.e2eBroker.priceCalls++
	b.e2eBroker.mu.Unlock()
	var rows []string
	for _, symbol := range strings.Split(r.URL.Query().Get("symbols"), ",") {
		symbol = strings.TrimSpace(symbol)
		price, ok := b.last[symbol]
		if !ok {
			continue
		}
		row := fmt.Sprintf(`{"symbol":%q,"lastPrice":%q,"currency":"KRW"`, symbol, price)
		if b.stamp != "" {
			row += "," + b.stamp
		}
		rows = append(rows, row+"}")
	}
	_, _ = w.Write([]byte(`{"result":[` + strings.Join(rows, ",") + `]}`))
}

func (b *a163PriceBroker) set(symbol, price, stamp string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.last[symbol] = price
	b.stamp = stamp
}

func newA163Stack(t *testing.T) (*e2eStack, *a163PriceBroker) {
	t.Helper()
	broker, srv := newA163Broker(t)
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	return buildE2EStack(t, broker.e2eBroker, srv, dir, false), broker
}

// a163Outcome 은 한 주기의 판정 · 발의 · 제출을 비교 가능한 값으로 모음.
type a163Outcome struct {
	Opened, Judged, Proposed, Observed int
	Orders                             []string // side/quantity/orderType
}

func a163Breach(t *testing.T, stamp string) a163Outcome {
	t.Helper()
	s, broker := newA163Stack(t)
	s.entry("005930", "10", "70000", "68000")
	broker.set("005930", "67000", stamp) // 손절선(68000) 아래
	cycle := s.observe()
	out := a163Outcome{Opened: cycle.Opened, Judged: cycle.Judged, Proposed: cycle.Proposed, Observed: cycle.Observed}
	for _, o := range broker.placed() {
		out.Orders = append(out.Orders, fmt.Sprintf("%v/%v/%v", o["side"], o["quantity"], o["orderType"]))
	}
	return out
}

// 2.1 — 9시간 30분 묵은 체결가(받은 시각은 15 s 안)라도 손절선 아래면 같은 주기에 보호 청산이 제출됨.
func TestA163AStaleBrokerTimeStillLiquidatesInTheSameCycle(t *testing.T) {
	control := a163Breach(t, a163FreshStamp)
	if control.Judged != 1 || control.Proposed != 1 || len(control.Orders) != 1 || control.Orders[0] != "SELL/10/LIMIT" {
		t.Fatalf("control = %+v, want the breach judged and the whole position sold in the same cycle", control)
	}
	stale := a163Breach(t, a163StaleStamp)
	if fmt.Sprint(stale) != fmt.Sprint(control) {
		t.Fatalf("stale broker time = %+v, control = %+v — the broker data time must not delay the stop (safety 4)", stale, control)
	}
}

// 2.2 — 시각이 null · 키 없음 · 빈 문자열 · 해독 불가 · `+0900` 형식이어도 판정 · 제출이 대조군과 같음(행 탈락 · 오류 0).
func TestA163AMissingOrMalformedBrokerTimeJudgesTheSame(t *testing.T) {
	control := a163Breach(t, a163FreshStamp)
	for name, stamp := range map[string]string{
		"null":    `"timestamp":null`,
		"absent":  ``,
		"empty":   `"timestamp":""`,
		"garbage": `"timestamp":"garbage"`,
		"+0900":   `"timestamp":"2026-03-30T09:59:58.000+0900"`,
	} {
		t.Run(name, func(t *testing.T) {
			got := a163Breach(t, stamp)
			if fmt.Sprint(got) != fmt.Sprint(control) {
				t.Fatalf("%s broker time = %+v, control = %+v — the adapter dropped the row or failed the read", name, got, control)
			}
		})
	}
}

// 2.3 — 브로커 시계가 앞서도(미래 시각) 판정에서 빠지지 않음.
func TestA163AFutureBrokerTimeJudgesTheSame(t *testing.T) {
	control := a163Breach(t, a163FreshStamp)
	if got := a163Breach(t, a163FutureStamp); fmt.Sprint(got) != fmt.Sprint(control) {
		t.Fatalf("future broker time = %+v, control = %+v", got, control)
	}
}

// 2.4 — 같은 체결 시각 · 같은 가격을 다른 두 주기에 받으면 관측 identity 와 원장 observed-at 은 각자의 받은 시각으로 만들어져 다름.
func TestA163IdentityAndObservedAtComeFromTheReceiptNotTheBrokerTime(t *testing.T) {
	s, broker := newA163Stack(t)
	p := s.entry("005930", "10", "70000", "68000")
	broker.set("005930", "70100", a163StaleStamp) // 손절선 위 · 첫 익절 아래 — 주문 없이 관측만 갱신

	first := s.observe()
	if first.Judged != 1 {
		t.Fatalf("first cycle judged %d, want 1", first.Judged)
	}
	firstAt := s.clk.Now().UTC()
	before := a163StoredSnapshot(t, s, p.ID)

	s.clk.Advance(20 * time.Second)
	if second := s.observe(); second.Judged != 1 {
		t.Fatalf("second cycle judged %d, want 1", second.Judged)
	}
	secondAt := s.clk.Now().UTC()
	after := a163StoredSnapshot(t, s, p.ID)

	if before.Line.ObservationID == after.Line.ObservationID {
		t.Fatalf("one broker time across two receipts produced one observation id %q — identity must follow the receipt", after.Line.ObservationID)
	}
	for _, c := range []struct {
		raw  string
		want time.Time
	}{{before.ObservedAt, firstAt}, {after.ObservedAt, secondAt}} {
		got, err := time.Parse(time.RFC3339Nano, c.raw)
		if err != nil || !got.Equal(c.want) {
			t.Fatalf("ledger observed_at = %q (%v), want the receipt %s — never the broker time", c.raw, err, c.want.Format(time.RFC3339Nano))
		}
	}
	if len(broker.placed()) != 0 {
		t.Fatalf("broker saw %d orders for an in-band quote, want none", len(broker.placed()))
	}
}

func a163StoredSnapshot(t *testing.T, s *e2eStack, positionID string) journal.StoredExitSnapshot {
	t.Helper()
	state := s.state(positionID)
	if state.Snapshot.Snapshot == nil {
		t.Fatal("no evaluated snapshot was stored")
	}
	return *state.Snapshot.Snapshot
}
