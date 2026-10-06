package main

// helpers_test.go 는 mock 두 개(official·TypeSafe)와 시험용 probe 조립을 제공함.
// 실 API 는 부르지 않음 — 모든 시험은 httptest 서버만 침.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// 계좌·시크릿 모양의 표식 — mock official 응답에 섞어 넣고, 외부 전송·원장에 새지 않음을 확인함.
// 실제 값이 아닌 시험용 문자열임.
const (
	sentinelAccount  = "ACCT-SENTINEL-9911"
	sentinelHolding  = "HOLD-SENTINEL-4242"
	sentinelBalance  = "BAL-SENTINEL-5150"
	sentinelOfficial = "OFFICIAL-TOKEN-SENTINEL-7733"
	testTypeSafeKey  = "test-typesafe-key-not-a-secret"
)

var sentinels = []string{sentinelAccount, sentinelHolding, sentinelBalance, sentinelOfficial, testTypeSafeKey}

// kst 는 시험 시각의 기준 시간대.
var kst = time.FixedZone("KST", 9*3600)

func at(hour, minute int) time.Time { return time.Date(2026, 10, 7, hour, minute, 0, 0, kst) }

// officialMock 은 official Open API 의 공개 읽기 다섯 경로 + 토큰 경로를 흉내냄.
type officialMock struct {
	t      *testing.T
	server *httptest.Server

	mu          sync.Mutex
	prices      map[string]string // symbol → lastPrice
	rateLimited map[string]int    // path → 앞으로 돌려줄 429 횟수
	failStatus  map[string]int    // path → 고정 실패 상태
	calls       map[string]int
	methods     map[string]bool
	noRegular   bool
}

func newOfficialMock(t *testing.T) *officialMock {
	m := &officialMock{
		t:           t,
		prices:      map[string]string{"005930": "70000", "000660": "120000"},
		rateLimited: map[string]int{},
		failStatus:  map[string]int{},
		calls:       map[string]int{},
		methods:     map[string]bool{},
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(m.server.Close)
	return m
}

func (m *officialMock) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.calls[r.URL.Path]++
	m.methods[r.Method+" "+r.URL.Path] = true
	if n := m.rateLimited[r.URL.Path]; n > 0 {
		m.rateLimited[r.URL.Path] = n - 1
		m.mu.Unlock()
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	if status := m.failStatus[r.URL.Path]; status != 0 {
		m.mu.Unlock()
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"error":"mock failure"}`)
		return
	}
	prices := map[string]string{}
	for symbol, price := range m.prices {
		prices[symbol] = price
	}
	noRegular := m.noRegular
	m.mu.Unlock()

	write := func(body string) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}
	switch r.URL.Path {
	case "/oauth2/token":
		write(`{"access_token":"` + sentinelOfficial + `","expires_in":86400,"token_type":"Bearer"}`)
	case "/api/v1/market-calendar/KR", "/api/v1/market-calendar/US":
		regular := `"regularMarket":{"startTime":"2026-10-07T09:00:00+09:00","endTime":"2026-10-07T15:30:00+09:00"}`
		if noRegular {
			regular = `"regularMarket":null`
		}
		write(`{"result":{"previousBusinessDay":{"date":"2026-10-06"},"today":{"date":"2026-10-07",` + regular +
			`},"nextBusinessDay":{"date":"2026-10-08"}}}`)
	case "/api/v1/rankings":
		// 계좌 모양 필드를 일부러 섞음 — 공개 어댑터가 버려야 함.
		write(`{"result":{"rankedAt":"2026-10-07T10:00:00+09:00","accountNo":"` + sentinelAccount + `","rankings":[` +
			`{"rank":1,"symbol":"005930","currency":"KRW","price":{"lastPrice":"70000","basePrice":"69000","changeRate":"1.45"},` +
			`"tradingVolume":"1000","tradingAmount":"70000000","holdingQuantity":"` + sentinelHolding + `"},` +
			`{"rank":2,"symbol":"000660","currency":"KRW","price":{"lastPrice":"120000","basePrice":"121000","changeRate":"-0.83"},` +
			`"tradingVolume":"500","tradingAmount":"60000000","accountBalance":"` + sentinelBalance + `"}]}}`)
	case "/api/v1/stocks":
		write(`{"result":[` +
			`{"symbol":"005930","name":"삼성전자","market":"KOSPI","currency":"KRW","status":"ACTIVE","accountNo":"` + sentinelAccount + `"},` +
			`{"symbol":"000660","name":"SK하이닉스","market":"KOSPI","currency":"KRW","status":"ACTIVE","holdingQuantity":"` + sentinelHolding + `"}]}`)
	case "/api/v1/prices":
		var rows []string
		for _, symbol := range strings.Split(r.URL.Query().Get("symbols"), ",") {
			if price, found := prices[symbol]; found {
				rows = append(rows, `{"symbol":"`+symbol+`","lastPrice":"`+price+`","currency":"KRW","timestamp":"2026-10-07T10:00:00+09:00",`+
					`"accountBalance":"`+sentinelBalance+`"}`)
			}
		}
		write(`{"result":[` + strings.Join(rows, ",") + `]}`)
	case "/api/v1/orderbook":
		write(`{"result":{"timestamp":"2026-10-07T09:59:58+09:00","currency":"KRW",` +
			`"asks":[{"price":"70100","volume":"12"}],"bids":[{"price":"70000","volume":"30"}]}}`)
	default:
		m.t.Errorf("unexpected official call %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *officialMock) client(t *testing.T) *official.Client {
	return official.New(official.Credentials{APIKey: "mock-key", SecretKey: "mock-secret"},
		filepath.Join(t.TempDir(), "openapi-token.json"), official.WithBaseURL(m.server.URL))
}

func (m *officialMock) callCount(path string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[path]
}

// typeSafeMock 은 /v1/systemone 을 흉내내고 받은 본문·헤더를 모두 보관함.
type typeSafeMock struct {
	server *httptest.Server

	mu       sync.Mutex
	bodies   [][]byte
	auth     []string
	statuses []int // 앞에서부터 하나씩 소비, 비면 200
	answer   func(questionID string) string
}

func newTypeSafeMock(t *testing.T) *typeSafeMock {
	m := &typeSafeMock{answer: func(id string) string {
		return fmt.Sprintf(`{"type":"noul","noul":%s}`, map[string]string{
			"j1_up_within_horizon": "0.93", "j2_hold_off_red_flag": "0.12",
		}[id])
	}}
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.bodies = append(m.bodies, body)
		m.auth = append(m.auth, r.Header.Get("Authorization"))
		status := http.StatusOK
		if len(m.statuses) > 0 {
			status, m.statuses = m.statuses[0], m.statuses[1:]
		}
		answer := m.answer
		m.mu.Unlock()
		if r.Method != http.MethodPost || r.URL.Path != systemOnePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			_, _ = io.WriteString(w, `{"detail":"mock refusal"}`)
			return
		}
		var request struct {
			Questions map[string]json.RawMessage `json:"questions"`
		}
		_ = json.Unmarshal(body, &request)
		var answers []string
		for id := range request.Questions {
			answers = append(answers, `"`+id+`":`+answer(id))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"model":"jev-1.13.0","answers":{`+strings.Join(answers, ",")+
			`},"usage":{"input_tokens":321,"output_tokens":20}}`)
	}))
	t.Cleanup(m.server.Close)
	return m
}

func (m *typeSafeMock) requests() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]byte(nil), m.bodies...)
}

func (m *typeSafeMock) judgeClient(t *testing.T) *judgeClient {
	client, err := newJudgeClient(testTypeSafeKey, m.server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	client.sleep = func(context.Context, time.Duration) error { return nil }
	return client
}

// fakeClock 은 sleep 이 시계를 앞으로 미는 결정적 시계.
type fakeClock struct {
	mu    sync.Mutex
	now   time.Time
	slept []time.Duration
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.slept = append(c.slept, d)
	if d > 0 {
		c.now = c.now.Add(d)
	}
	return ctx.Err()
}

func (c *fakeClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

type testRig struct {
	official *officialMock
	typeSafe *typeSafeMock
	clock    *fakeClock
	probe    *probe
	dir      string
}

func newTestRig(t *testing.T) *testRig {
	t.Helper()
	officialServer := newOfficialMock(t)
	typeSafeServer := newTypeSafeMock(t)
	dir := t.TempDir()
	l, err := openLedger(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	clock := &fakeClock{now: at(10, 0)}
	p := newProbe(officialServer.client(t), typeSafeServer.judgeClient(t), l, config{
		Interval: 10 * time.Minute, TopK: 10, Horizon: 60 * time.Minute, LabelTick: 30 * time.Second,
		RetryDelays: []time.Duration{time.Second, 2 * time.Second},
	})
	p.now, p.sleep = clock.Now, clock.Sleep
	p.logf = func(string, ...any) {}
	return &testRig{official: officialServer, typeSafe: typeSafeServer, clock: clock, probe: p, dir: dir}
}

func (r *testRig) contents(t *testing.T) ledgerContents {
	t.Helper()
	contents, err := readLedger(filepath.Join(r.dir, ledgerFileName))
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func (r *testRig) session() session { return session{Open: at(9, 0), Close: at(15, 30)} }
