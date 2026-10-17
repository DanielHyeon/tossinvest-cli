package main

// helpers_test.go 는 official Open API mock 과 run() 조립을 제공함.
// 실 API 는 부르지 않음 — 모든 시험은 httptest 서버만 침.

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// 시크릿·계좌 모양의 표식 — mock 에 섞어 넣고 영수증·출력에 새지 않음을 확인함. 실제 값 아님.
const (
	sentinelAPIKey    = "APIKEY-SENTINEL-1357"
	sentinelSecret    = "SECRET-SENTINEL-2468"
	sentinelToken     = "ACCESS-TOKEN-SENTINEL-7733"
	sentinelAccountNo = "ACCTNO-SENTINEL-9911"
	sentinelSeq       = "914273"
)

var sentinels = []string{sentinelAPIKey, sentinelSecret, sentinelToken, sentinelAccountNo, sentinelSeq}

// fixedNow 는 시험 기준 시각(평일 장중 KST).
var fixedNow = time.Date(2026, 10, 12, 10, 15, 0, 0, kst)

// orderResponder 는 주문 endpoint 의 응답을 정함. nil 이면 200 접수.
type orderResponder func(w http.ResponseWriter, r *http.Request, body []byte)

type officialMock struct {
	t      *testing.T
	server *httptest.Server

	mu         sync.Mutex
	calls      map[string]int
	orderBody  [][]byte
	orderHdrs  []http.Header
	respond    orderResponder
	onOrderHit func()
}

func newOfficialMock(t *testing.T) *officialMock {
	m := &officialMock{t: t, calls: map[string]int{}}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(m.server.Close)
	return m
}

func (m *officialMock) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	m.mu.Lock()
	m.calls[r.Method+" "+r.URL.Path]++
	respond := m.respond
	hook := m.onOrderHit
	if r.URL.Path == orderPath {
		m.orderBody = append(m.orderBody, body)
		m.orderHdrs = append(m.orderHdrs, r.Header.Clone())
	}
	m.mu.Unlock()

	write := func(status int, payload string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, payload)
	}
	switch r.Method + " " + r.URL.Path {
	case "POST /oauth2/token":
		write(200, `{"access_token":"`+sentinelToken+`","expires_in":86400,"token_type":"Bearer"}`)
	case "GET /api/v1/accounts":
		write(200, `{"result":[{"accountNo":"`+sentinelAccountNo+`","accountSeq":`+sentinelSeq+`,"accountType":"GENERAL"}]}`)
	case "POST " + orderPath:
		if hook != nil {
			hook()
		}
		if respond != nil {
			respond(w, r, body)
			return
		}
		write(200, `{"result":{"orderId":"ORDER-MOCK-1","clientOrderId":`+echoKey(body)+`}}`)
	default:
		write(404, `{"error":{"code":"not-found"}}`)
	}
}

// echoKey 는 요청 본문의 clientOrderId 를 JSON 문자열로 되돌림(브로커의 "요청 시 전달한 값 그대로").
func echoKey(body []byte) string {
	match := regexp.MustCompile(`"clientOrderId":("[^"]*")`).FindSubmatch(body)
	if match == nil {
		return "null"
	}
	return string(match[1])
}

func (m *officialMock) count(key string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[key]
}

func (m *officialMock) totalCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for _, n := range m.calls {
		total += n
	}
	return total
}

func (m *officialMock) setHook(fn func()) {
	m.mu.Lock()
	m.onOrderHit = fn
	m.mu.Unlock()
}

// orderRequest 는 주문 endpoint 가 i 번째로 받은 본문·헤더임.
func (m *officialMock) orderRequest(i int) ([]byte, http.Header) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orderBody[i], m.orderHdrs[i]
}

func (m *officialMock) setRespond(fn orderResponder) {
	m.mu.Lock()
	m.respond = fn
	m.mu.Unlock()
}

// harness 는 run() 에 넣을 deps 와 결과를 묶음.
type harness struct {
	t        *testing.T
	mock     *officialMock
	out      string
	now      time.Time
	terminal bool
	timeout  time.Duration
	opened   int
}

func newHarness(t *testing.T) *harness {
	return &harness{
		t:        t,
		mock:     newOfficialMock(t),
		out:      filepath.Join(t.TempDir(), "receipts"),
		now:      fixedNow,
		terminal: true,
		timeout:  2 * time.Second,
	}
}

func (h *harness) deps(stdout io.Writer) deps {
	tokenFile := filepath.Join(h.t.TempDir(), "openapi-token.json")
	return deps{
		now:             func() time.Time { return h.now },
		random:          bytes.NewReader([]byte{0xde, 0xad, 0xbe, 0xef, 0x01, 0x02, 0x03, 0x04}),
		stdout:          stdout,
		stdinIsTerminal: func() bool { return h.terminal },
		openClient: func(string) (headerSource, string, error) {
			h.opened++
			client := official.New(official.Credentials{APIKey: sentinelAPIKey, SecretKey: sentinelSecret},
				tokenFile, official.WithBaseURL(h.mock.server.URL))
			return client, client.BaseURL(), nil
		},
		sendTimeout: h.timeout,
	}
}

// run 은 도구를 한 번 돌리고 (종료 코드, 출력, 오류) 를 돌려줌.
func (h *harness) run(args ...string) (int, string, error) {
	var stdout bytes.Buffer
	code, err := run(context.Background(), append(args, "--out", h.out), h.deps(&stdout))
	return code, stdout.String(), err
}

var tokenLine = regexp.MustCompile(`confirm token: (\S+)`)

// previewToken 은 미리보기를 돌려 출력에서 토큰을 읽음.
func (h *harness) previewToken(symbol, qty string) string {
	h.t.Helper()
	code, out, err := h.run("--symbol", symbol, "--qty", qty)
	if err != nil || code != exitOK {
		h.t.Fatalf("preview: code=%d err=%v out=%s", code, err, out)
	}
	match := tokenLine.FindStringSubmatch(out)
	if match == nil {
		h.t.Fatalf("preview printed no confirm token:\n%s", out)
	}
	return match[1]
}

func (h *harness) receiptFiles() []string {
	h.t.Helper()
	entries, err := os.ReadDir(h.out)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		h.t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, filepath.Join(h.out, entry.Name()))
		}
	}
	return names
}

func (h *harness) onlyReceipt() []byte {
	h.t.Helper()
	files := h.receiptFiles()
	if len(files) != 1 {
		h.t.Fatalf("receipt files = %v, want exactly one", files)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		h.t.Fatal(err)
	}
	return data
}
