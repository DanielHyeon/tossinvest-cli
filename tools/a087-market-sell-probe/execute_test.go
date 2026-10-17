package main

// execute_test.go 는 미리보기→실행 전체를 mock 에 대고 돌림 — 단일 전송·영수증·거절 경로.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestPreviewTouchesNothing(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	code, out, err := h.run("--symbol", "005930", "--qty", "1")
	if err != nil || code != exitOK {
		t.Fatalf("preview: code=%d err=%v", code, err)
	}
	want := `{"symbol":"005930","side":"SELL","orderType":"MARKET","quantity":"1","clientOrderId":"a087ms-20261012T101500-deadbeef","confirmHighValueOrder":false}`
	if !strings.Contains(out, want) {
		t.Fatalf("preview did not print the wire body verbatim:\n%s", out)
	}
	if !strings.Contains(out, "nothing was sent") || !strings.Contains(out, "--execute --confirm") {
		t.Fatalf("preview output lacks the no-send notice or the execute line:\n%s", out)
	}
	if h.mock.totalCalls() != 0 || h.opened != 0 {
		t.Fatalf("preview reached the broker (%d calls) or opened credentials (%d)", h.mock.totalCalls(), h.opened)
	}
	if files := h.receiptFiles(); len(files) != 0 {
		t.Fatalf("preview wrote receipts %v", files)
	}
}

func TestPreviewRefusesGuardedShapes(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{
		{"--symbol", "AAPL", "--qty", "1"},
		{"--symbol", "005930", "--qty", "3"},
		{"--symbol", "005930", "--qty", "0"},
		{"--symbol", "005930"},
	} {
		h := newHarness(t)
		code, out, err := h.run(args...)
		if err == nil || code != exitError || strings.Contains(out, "confirm token") {
			t.Errorf("%v: code=%d err=%v out=%q", args, code, err, out)
		}
	}
}

// TestExecuteSendsOnceAndRecordsAcceptance 는 정상 경로 전부를 봄: 서버가 받은 바이트 = 미리보기 바이트 =
// 영수증 바이트, 요청 1회, 계좌·인증 헤더가 실렸음.
func TestExecuteSendsOnceAndRecordsAcceptance(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	code, out, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token)
	if err != nil || code != exitOK {
		t.Fatalf("execute: code=%d err=%v out=%s", code, err, out)
	}
	if n := h.mock.count("POST " + orderPath); n != 1 {
		t.Fatalf("order requests = %d, want exactly 1", n)
	}
	var r receipt
	if err := json.Unmarshal(h.onlyReceipt(), &r); err != nil {
		t.Fatal(err)
	}
	if r.State != stateAnswered || r.Outcome != outcomeAccepted || r.Response.OrderID != "ORDER-MOCK-1" {
		t.Fatalf("receipt state=%s outcome=%s response=%+v", r.State, r.Outcome, r.Response)
	}
	gotBody, hdr := h.mock.orderRequest(0)
	if string(gotBody) != r.Request.WireBody {
		t.Fatalf("the server received\n %s\nbut the receipt records\n %s", gotBody, r.Request.WireBody)
	}
	if hdr.Get("Authorization") != "Bearer "+sentinelToken || hdr.Get("X-Tossinvest-Account") != sentinelSeq ||
		hdr.Get("Content-Type") != "application/json" {
		t.Fatalf("request headers = %v", hdr)
	}
	if r.Response.EchoMatchesSent == nil || !*r.Response.EchoMatchesSent {
		t.Fatalf("the echoed key was not recorded as matching: %+v", r.Response)
	}
	if !strings.Contains(out, "outcome: accepted") || !strings.Contains(out, "order id: ORDER-MOCK-1") {
		t.Fatalf("execute output:\n%s", out)
	}
}

// TestExecuteRefusesMismatchedConfirm 는 토큰 불일치면 자격증명조차 열지 않음을 봄.
func TestExecuteRefusesMismatchedConfirm(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	for _, args := range [][]string{
		{"--symbol", "005930", "--qty", "2", "--execute", "--confirm", token},
		{"--symbol", "000660", "--qty", "1", "--execute", "--confirm", token},
		{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token[:len(token)-2] + "zz"},
		{"--symbol", "005930", "--qty", "1", "--execute"},
	} {
		code, _, err := h.run(args...)
		if err == nil || code != exitError {
			t.Errorf("%v: code=%d err=%v, want a refusal", args, code, err)
		}
	}
	if h.mock.totalCalls() != 0 || h.opened != 0 || len(h.receiptFiles()) != 0 {
		t.Fatalf("a refused confirm reached the broker: calls=%d opened=%d receipts=%v",
			h.mock.totalCalls(), h.opened, h.receiptFiles())
	}
}

func TestExecuteRefusesAnExpiredToken(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	h.now = h.now.Add(confirmWindow + time.Second)
	code, _, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token)
	if err == nil || code != exitError || !strings.Contains(err.Error(), "the window is") {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if h.mock.totalCalls() != 0 {
		t.Fatalf("an expired token reached the broker (%d calls)", h.mock.totalCalls())
	}
}

func TestExecuteNeedsATerminal(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	h.terminal = false
	code, _, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token)
	if err == nil || code != exitError || !strings.Contains(err.Error(), "not a terminal") {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if h.mock.totalCalls() != 0 || h.opened != 0 {
		t.Fatal("a non-terminal execute reached the broker")
	}
}

// sendAndRead 는 주어진 응답으로 실행을 한 번 돌리고 영수증을 읽음.
func sendAndRead(t *testing.T, h *harness) (int, receipt, error) {
	t.Helper()
	token := h.previewToken("005930", "1")
	code, _, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token)
	var r receipt
	if decodeErr := json.Unmarshal(h.onlyReceipt(), &r); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	return code, r, err
}

// TestNoRetryOnAnyFailure 는 5xx·401·429·409·타임아웃·리다이렉트 어디서도 두 번째 요청이 없음을 봄.
// 401 은 internal/official 의 send 라면 토큰을 갈아 같은 본문을 다시 보내는 자리임.
func TestNoRetryOnAnyFailure(t *testing.T) {
	t.Parallel()
	errorBody := func(status int, code string) orderResponder {
		return func(w http.ResponseWriter, _ *http.Request, _ []byte) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = io.WriteString(w, `{"error":{"code":"`+code+`","message":"mock","requestId":"REQ-1"}}`)
		}
	}
	cases := []struct {
		name      string
		respond   orderResponder
		timeout   time.Duration
		wantCode  int
		wantOut   string
		wantState string
	}{
		{"500", errorBody(500, "internal-error"), 0, exitUnknown, outcomeUnknown, stateAnswered},
		{"503", errorBody(503, "maintenance"), 0, exitUnknown, outcomeUnknown, stateAnswered},
		{"401", errorBody(401, "expired-token"), 0, exitRefused, outcomeRefused, stateAnswered},
		{"429", errorBody(429, "rate-limit-exceeded"), 0, exitRefused, outcomeRefused, stateAnswered},
		{"409", errorBody(409, "request-in-progress"), 0, exitUnknown, outcomeUnknown, stateAnswered},
		{"redirect", func(w http.ResponseWriter, r *http.Request, _ []byte) {
			http.Redirect(w, r, "/api/v1/orders-elsewhere", http.StatusTemporaryRedirect)
		}, 0, exitUnknown, outcomeUnknown, stateAnswered},
		{"timeout", func(_ http.ResponseWriter, r *http.Request, _ []byte) {
			select {
			case <-r.Context().Done():
			case <-time.After(3 * time.Second):
			}
		}, 150 * time.Millisecond, exitUnknown, outcomeUnknown, stateNoAnswer},
		{"200 without orderId", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
			_, _ = io.WriteString(w, `{"result":{}}`)
		}, 0, exitUnknown, outcomeUnknown, stateAnswered},
		{"200 echoing another key", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
			_, _ = io.WriteString(w, `{"result":{"orderId":"X","clientOrderId":"someone-else"}}`)
		}, 0, exitUnknown, outcomeUnknown, stateAnswered},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			if tc.timeout != 0 {
				h.timeout = tc.timeout
			}
			h.mock.setRespond(tc.respond)
			code, r, err := sendAndRead(t, h)
			if err != nil || code != tc.wantCode {
				t.Fatalf("code=%d err=%v, want code %d", code, err, tc.wantCode)
			}
			if n := h.mock.count("POST " + orderPath); n != 1 {
				t.Fatalf("order requests = %d, want exactly 1 (no retry)", n)
			}
			if n := h.mock.count("POST /api/v1/orders-elsewhere") + h.mock.count("GET /api/v1/orders-elsewhere"); n != 0 {
				t.Fatalf("a redirect was followed (%d)", n)
			}
			if r.Outcome != tc.wantOut || r.State != tc.wantState {
				t.Fatalf("receipt outcome=%s state=%s, want %s/%s", r.Outcome, r.State, tc.wantOut, tc.wantState)
			}
			if tc.wantOut == outcomeUnknown && !strings.Contains(r.Guidance, "Do NOT resend") {
				t.Fatalf("unknown guidance = %q", r.Guidance)
			}
			if tc.wantState == stateNoAnswer && (r.Response != nil || r.TransportError == "") {
				t.Fatalf("a timeout must record no response and a transport error: %+v", r)
			}
		})
	}
}

// TestRefusalRecordsTheBrokerCode 는 5.2 가 읽을 422 오류 코드·원문이 영수증에 남는지 봄.
func TestRefusalRecordsTheBrokerCode(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.now = time.Date(2026, 10, 12, 16, 5, 0, 0, kst)
	body := `{"error":{"code":"order-hours-closed","data":{"retryAfterAt":"2026-10-13T09:00:00+09:00"},"message":"현재 해당 주문을 접수할 수 없는 시간입니다.","requestId":"REQ-422"}}`
	h.mock.setRespond(func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		_, _ = io.WriteString(w, body)
	})
	code, r, err := sendAndRead(t, h)
	if err != nil || code != exitRefused {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if r.Response.HTTPStatus != 422 || r.Response.ErrorCode != "order-hours-closed" || r.Response.Body != body {
		t.Fatalf("response record = %+v", r.Response)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, r.Response.BodyJSON); err != nil || compact.String() != body ||
		r.Response.BodySHA256 != sha256Hex([]byte(body)) {
		t.Fatalf("raw body not preserved: %s", r.Response.BodyJSON)
	}
	if r.Session.ClockPhase != phaseOutside || r.Session.KSTWeekday != "Monday" {
		t.Fatalf("session context = %+v", r.Session)
	}
}

// TestSecondExecuteWithTheSameTokenIsRefused 는 토큰이 1회용임을 봄 — 같은 명령을 다시 쳐도 안 나감.
func TestSecondExecuteWithTheSameTokenIsRefused(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.mock.setRespond(func(w http.ResponseWriter, _ *http.Request, _ []byte) { w.WriteHeader(500) })
	token := h.previewToken("005930", "1")
	args := []string{"--symbol", "005930", "--qty", "1", "--execute", "--confirm", token}
	if code, _, err := h.run(args...); err != nil || code != exitUnknown {
		t.Fatalf("first: code=%d err=%v", code, err)
	}
	before := h.mock.totalCalls()
	code, _, err := h.run(args...)
	if !errors.Is(err, errAlreadySent) || code != exitError {
		t.Fatalf("second: code=%d err=%v, want errAlreadySent", code, err)
	}
	if h.mock.totalCalls() != before || h.mock.count("POST "+orderPath) != 1 {
		t.Fatalf("the second run reached the broker")
	}
}

// TestCreatePendingReceiptIsExclusive 는 사전 검사(Lstat)와 별개로 파일 생성 자체가 배타임을 봄 —
// 둘이 서로를 가려 한쪽이 지워져도 시험이 통과하는 일을 막으려고 층을 내려 따로 잼.
func TestCreatePendingReceiptIsExclusive(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "k.json")
	if err := createPendingReceipt(path, receipt{State: stateSending}); err != nil {
		t.Fatal(err)
	}
	if err := createPendingReceipt(path, receipt{State: stateSending}); !errors.Is(err, errAlreadySent) {
		t.Fatalf("second create = %v, want errAlreadySent", err)
	}
}

// TestPendingReceiptExistsBeforeTheRequestLeaves 는 요청이 서버에 닿은 순간 이미 "sending" 영수증이
// 디스크에 있음을 봄 — 전송 중 프로세스가 죽어도 흔적이 남는 근거.
func TestPendingReceiptExistsBeforeTheRequestLeaves(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	seenCh := make(chan string, 1)
	h.mock.setHook(func() {
		state := "missing"
		if entries, err := os.ReadDir(h.out); err == nil {
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".json") && !strings.HasPrefix(entry.Name(), ".") {
					data, _ := os.ReadFile(filepath.Join(h.out, entry.Name()))
					var r receipt
					_ = json.Unmarshal(data, &r)
					state = r.State
				}
			}
		}
		seenCh <- state
	})
	if code, _, _ := sendAndRead(t, h); code != exitOK {
		t.Fatalf("code=%d", code)
	}
	if seen := <-seenCh; seen != stateSending {
		t.Fatalf("receipt state at the moment of the request = %q, want %q", seen, stateSending)
	}
}

// TestUnwritableReceiptDirSendsNothing 는 영수증을 못 쓰면 보내지 않음을 봄 — 경로가 파일에 막힌 경우와
// 디렉터리에 쓰기 권한이 없는 경우(배타 생성 자체가 실패) 둘 다.
func TestUnwritableReceiptDirSendsNothing(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	blocked := func(t *testing.T) string {
		blocker := filepath.Join(t.TempDir(), "a-file")
		if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		return filepath.Join(blocker, "receipts")
	}
	readOnly := func(t *testing.T) string {
		dir := filepath.Join(t.TempDir(), "ro")
		if err := os.Mkdir(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		return dir
	}
	for name, makeDir := range map[string]func(*testing.T) string{"blocked by a file": blocked, "read-only": readOnly} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			h.out = makeDir(t)
			token := h.previewToken("005930", "1")
			code, _, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token)
			if err == nil || code != exitError || !strings.Contains(err.Error(), "nothing was sent") {
				t.Fatalf("code=%d err=%v", code, err)
			}
			if n := h.mock.count("POST " + orderPath); n != 0 {
				t.Fatalf("order requests = %d with no receipt, want 0", n)
			}
		})
	}
}

// TestReceiptSchemaAndNoSecrets 는 영수증의 칸 집합을 고정하고, 시크릿·토큰·계좌 표식이 영수증과 출력
// 어디에도 없음을 봄(mock 은 계좌 목록에 계좌번호를, 헤더에 계좌 seq 를 실제로 실어 보냄).
func TestReceiptSchemaAndNoSecrets(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	token := h.previewToken("005930", "1")
	code, out, err := h.run("--symbol", "005930", "--qty", "1", "--execute", "--confirm", token, "--note", "5.1 test")
	if err != nil || code != exitOK {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if h.mock.count("GET /api/v1/accounts") != 1 || h.mock.count("POST /oauth2/token") != 1 {
		t.Fatal("the mock did not exercise the shared token manager (token exchange + account discovery)")
	}
	data := h.onlyReceipt()
	for _, s := range sentinels {
		if strings.Contains(string(data), s) || strings.Contains(out, s) {
			t.Errorf("sentinel %q leaked into the receipt or the output", s)
		}
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatal(err)
	}
	want := []string{"answered_at_kst", "client_order_id", "confirm_token", "guidance", "issued_at_kst", "note",
		"outcome", "request", "response", "schema", "sent_at_kst", "session_context", "state", "tool"}
	if got := sortedKeys(top); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("receipt keys\n got %v\nwant %v", got, want)
	}
	var r receipt
	_ = json.Unmarshal(data, &r)
	if r.Schema != receiptSchema || r.Request.Method != "POST" || r.Request.Path != orderPath ||
		r.Request.WireBodySHA256 != sha256Hex([]byte(r.Request.WireBody)) {
		t.Fatalf("request record = %+v", r.Request)
	}
	if strings.Join(r.Request.HeaderNames, ",") != "Authorization,Content-Type,X-Tossinvest-Account" {
		t.Fatalf("header names = %v", r.Request.HeaderNames)
	}
	if r.SentAtKST == "" || !strings.HasSuffix(r.SentAtKST, "+09:00") || r.Session.ClockPhase != phaseRegular {
		t.Fatalf("time context sent=%q session=%+v", r.SentAtKST, r.Session)
	}
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestSessionPhases(t *testing.T) {
	t.Parallel()
	cases := map[time.Time]string{
		time.Date(2026, 10, 12, 8, 59, 0, 0, kst):  phaseOutside,
		time.Date(2026, 10, 12, 9, 0, 0, 0, kst):   phaseRegular,
		time.Date(2026, 10, 12, 15, 29, 0, 0, kst): phaseRegular,
		time.Date(2026, 10, 12, 15, 30, 0, 0, kst): phaseOutside,
		time.Date(2026, 10, 10, 11, 0, 0, 0, kst):  phaseWeekend,
	}
	for at, want := range cases {
		if got := sessionAt(at).ClockPhase; got != want {
			t.Errorf("%s: %s, want %s", at, got, want)
		}
	}
}
