package official

// reconcile_detail_reads_test.go — a164 tasks 2.1: by-id 대사 읽기의 분류(statement · silent · 거절)와 엄격 해독.
//
// 원 바이트 골든은 0.6 영수증(analysis/receipts/expiry-probe-2026-10-22.json)의 **값**을 OpenAPI 스키마
// (`ConditionalOrderDetailResponse` · `ConditionalOrderCondition` · `ErrorResponse`, docs/migration/openapi.latest.json)의
// **모양**에 넣어 만든다 — 영수증은 CLI 도메인 출력(snake_case)이라 원 응답 골든으로 쓸 수 없다(design 「0.6 영수증 대조」 C1).
// 수량 값은 영수증에서 제거됐으므로 계약 예시 값("1")을 쓴다(판정 입력 아님).

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const detailTargetID = "DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg"

// detailGoldenBody 는 영수증 ① 의 값을 계약 모양에 넣은 상세 응답이다. second·first.triggeredOrderId 는 영수증에 없다(C2 —
// 키 없음·null 구분 불가) → 계약상 nullable 이므로 이 골든은 키를 빼고, 아래 변형이 null 을 넣는다.
const detailGoldenResult = `{"conditionalOrderId":"` + detailTargetID + `","type":"SINGLE","status":"EXPIRED","symbol":"333430",` +
	`"market":"KR","quantity":"1","orderType":"MARKET","expireDate":"2026-08-07",` +
	`"first":{"type":"STOP","status":"EXPIRED","orderSide":"SELL","triggerPrice":"3200","targetProfitRate":null,"orderPrice":null},` +
	`"createdAt":"2026-07-31T18:49:05+09:00"}`

// detailNotFoundBody 는 M40 실측 문자열(code · message · requestId 모양)의 404 본문이다.
const detailNotFoundBody = `{"error":{"requestId":"tTXXXQMKNPt7Zuqw","code":"conditional-order-not-found","message":"존재하지 않는 설정입니다"}}`

type detailServer struct {
	*httptest.Server
	mu      sync.Mutex
	seen    []string
	headers []string
}

// newDetailServer 는 by-id 경로에 status·body 를 그대로 돌려준다(token 은 정상).
func newDetailServer(t *testing.T, status int, body string) *detailServer {
	t.Helper()
	s := &detailServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.seen = append(s.seen, r.Method+" "+r.URL.EscapedPath())
		s.headers = append(s.headers, r.Header.Get("X-Tossinvest-Account"))
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/token" {
			fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(s.Server.Close)
	return s
}

func detailClient(t *testing.T, s *detailServer) *Client {
	t.Helper()
	return New(Credentials{APIKey: "k", SecretKey: "s"}, filepath.Join(t.TempDir(), "t.json"),
		WithBaseURL(s.URL), WithHTTPClient(s.Client()), WithAccountSeq(7))
}

func wrapResult(result string) string { return `{"result":` + result + `}` }

// TestReconcileByIDReadsTheReceiptShapedStatement 는 수락 갈래다 — 영수증 ① 값의 statement 를 그대로 옮기고, 계좌 헤더를 단
// GET 하나만 낸다.
func TestReconcileByIDReadsTheReceiptShapedStatement(t *testing.T) {
	s := newDetailServer(t, http.StatusOK, wrapResult(detailGoldenResult))
	got, err := detailClient(t, s).ReconcileConditionalOrder(context.Background(), detailTargetID)
	if err != nil {
		t.Fatalf("a contract-shaped detail must be a statement: %v", err)
	}
	want := ReconcileConditionalDetail{ID: detailTargetID, Type: "SINGLE", Status: "EXPIRED", Symbol: "333430", Market: "KR",
		FirstStatus: "EXPIRED"}
	digest := got.ResultDigest
	got.ResultDigest = ""
	if got != want {
		t.Fatalf("statement %+v, want %+v", got, want)
	}
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+32 {
		t.Fatalf("result digest %q, want sha256 truncated to 16 bytes", digest)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var reads []string
	for i, r := range s.seen {
		if strings.HasPrefix(r, "GET /api/v1/conditional-orders/") {
			reads = append(reads, r)
			if s.headers[i] != "7" {
				t.Fatalf("by-id read sent account header %q, want 7", s.headers[i])
			}
		} else if r != "POST /oauth2/token" {
			t.Fatalf("unexpected request %s", r)
		}
	}
	if len(reads) != 1 || reads[0] != "GET /api/v1/conditional-orders/"+detailTargetID {
		t.Fatalf("by-id requests %v", reads)
	}
}

// TestReconcileByIDTreatsAbsentAndNullOptionalKeysAlike 는 design D3 정정(C2)이다 — triggeredOrderId·second 는 계약상 nullable
// 이라 키 없음과 null 이 같게 「없음」이고, 값이 있으면 형만 본다.
func TestReconcileByIDTreatsAbsentAndNullOptionalKeysAlike(t *testing.T) {
	base, err := DecodeReconcileConditionalDetail([]byte(detailGoldenResult))
	if err != nil {
		t.Fatalf("golden: %v", err)
	}
	withNulls := strings.Replace(detailGoldenResult, `"orderPrice":null}`, `"orderPrice":null,"triggeredOrderId":null},"second":null`, 1)
	got, err := DecodeReconcileConditionalDetail([]byte(withNulls))
	if err != nil {
		t.Fatalf("null optional keys: %v", err)
	}
	if got.TriggeredOrderID != "" || got.HasSecond || got.Status != base.Status || got.FirstStatus != base.FirstStatus {
		t.Fatalf("null optional keys decoded as %+v, want the same ending as absent keys %+v", got, base)
	}
	fired := strings.Replace(detailGoldenResult, `"orderPrice":null}`, `"orderPrice":null,"triggeredOrderId":"ORD-1"},"second":{"type":"STOP","status":"CANCELED","orderSide":"SELL"}`, 1)
	got, err = DecodeReconcileConditionalDetail([]byte(fired))
	if err != nil {
		t.Fatalf("present optional keys: %v", err)
	}
	if got.TriggeredOrderID != "ORD-1" || !got.HasSecond {
		t.Fatalf("present optional keys decoded as %+v — the trigger and the second leg must surface", got)
	}
}

// TestReconcileByIDClassifiesOnlyTheMeasuredNotFoundAsSilent 는 silent 의 정의다 — 404 이면서 본문 error.code 가 정확히
// conditional-order-not-found(M40) 일 때만. requestId 는 함께 옮긴다.
func TestReconcileByIDClassifiesOnlyTheMeasuredNotFoundAsSilent(t *testing.T) {
	s := newDetailServer(t, http.StatusNotFound, detailNotFoundBody)
	got, err := detailClient(t, s).ReconcileConditionalOrder(context.Background(), detailTargetID)
	if err != nil {
		t.Fatalf("the measured not-found must be silent, not an error: %v", err)
	}
	if !got.Silent || got.RequestID != "tTXXXQMKNPt7Zuqw" || got.Status != "" || got.ID != "" {
		t.Fatalf("silent detail %+v", got)
	}
}

// TestReconcileByIDRefusesEveryOtherOutcome 는 그 밖 전부가 거절(오류)임을 잰다 — 다른 code 의 404 · 본문 해독 불능 404 ·
// 400 · 401 · 403 · 429 · 5xx. silent 로 접히면 안 된다.
func TestReconcileByIDRefusesEveryOtherOutcome(t *testing.T) {
	for name, c := range map[string]struct {
		status int
		body   string
		is     error
	}{
		"404-other-code":   {http.StatusNotFound, `{"error":{"requestId":"r","code":"order-not-found","message":"x"}}`, nil},
		"404-no-code":      {http.StatusNotFound, `{"error":{"requestId":"r","message":"x"}}`, nil},
		"404-unreadable":   {http.StatusNotFound, `<html>not found</html>`, nil},
		"404-empty":        {http.StatusNotFound, ``, nil},
		"404-code-in-text": {http.StatusNotFound, `{"error":"conditional-order-not-found"}`, nil},
		"400":              {http.StatusBadRequest, `{"error":{"requestId":"r","code":"conditional-order-not-found","message":"x"}}`, nil},
		"401":              {http.StatusUnauthorized, `{"error":{"code":"unauthorized"}}`, ErrAuth},
		"403":              {http.StatusForbidden, `{"error":{"code":"forbidden"}}`, ErrAuth},
		"403-ip":           {http.StatusForbidden, `{"error":{"code":"ip not allowed"}}`, ErrIPNotAllowed},
		"429":              {http.StatusTooManyRequests, `{"error":{"code":"conditional-order-not-found"}}`, ErrRateLimited},
		"500":              {http.StatusInternalServerError, `{"error":{"code":"conditional-order-not-found"}}`, ErrServer},
		"503":              {http.StatusServiceUnavailable, ``, ErrServer},
	} {
		t.Run(name, func(t *testing.T) {
			s := newDetailServer(t, c.status, c.body)
			got, err := detailClient(t, s).ReconcileConditionalOrder(context.Background(), detailTargetID)
			if err == nil {
				t.Fatalf("HTTP %d %s was accepted as %+v — only the measured not-found is silent", c.status, c.body, got)
			}
			if c.is != nil && !errors.Is(err, c.is) {
				t.Fatalf("HTTP %d: want %v, got %v", c.status, c.is, err)
			}
		})
	}
}

// TestReconcileByIDRefusesATransportError — 서버가 닫혀 응답이 없으면 전송 오류(거절)다.
func TestReconcileByIDRefusesATransportError(t *testing.T) {
	s := newDetailServer(t, http.StatusOK, wrapResult(detailGoldenResult))
	c := detailClient(t, s)
	s.Close()
	if _, err := c.ReconcileConditionalOrder(context.Background(), detailTargetID); err == nil {
		t.Fatal("a closed server produced no error")
	}
}

// TestReconcileByIDRefusesSchemaHoles 는 statement 의 엄격 해독이다 — 계약 required 다섯(conditionalOrderId · status · symbol ·
// market · first.status)의 결측·null·형 불일치, first 결측·비객체, null result, nullable 둘의 형 불일치는 전부
// ErrReconcileSchema(빈 값으로 접지 않음 — F2 와 같은 이유).
func TestReconcileByIDRefusesSchemaHoles(t *testing.T) {
	drop := func(key string) string {
		return strings.Replace(detailGoldenResult, `"`+key+`":`, `"x-`+key+`":`, 1)
	}
	set := func(old, new string) string { return strings.Replace(detailGoldenResult, old, new, 1) }
	cases := map[string]string{
		"result-null":               `null`,
		"result-array":              `[]`,
		"result-string":             `"x"`,
		"id-missing":                drop("conditionalOrderId"),
		"id-null":                   set(`"conditionalOrderId":"`+detailTargetID+`"`, `"conditionalOrderId":null`),
		"id-number":                 set(`"conditionalOrderId":"`+detailTargetID+`"`, `"conditionalOrderId":7`),
		"status-missing":            set(`"status":"EXPIRED","symbol"`, `"symbol"`),
		"status-null":               set(`"status":"EXPIRED","symbol"`, `"status":null,"symbol"`),
		"status-bool":               set(`"status":"EXPIRED","symbol"`, `"status":true,"symbol"`),
		"symbol-missing":            drop("symbol"),
		"symbol-null":               set(`"symbol":"333430"`, `"symbol":null`),
		"market-missing":            drop("market"),
		"market-null":               set(`"market":"KR"`, `"market":null`),
		"first-missing":             drop("first"),
		"first-null":                set(`"first":{"type":"STOP","status":"EXPIRED","orderSide":"SELL","triggerPrice":"3200","targetProfitRate":null,"orderPrice":null}`, `"first":null`),
		"first-array":               set(`"first":{"type":"STOP","status":"EXPIRED","orderSide":"SELL","triggerPrice":"3200","targetProfitRate":null,"orderPrice":null}`, `"first":[]`),
		"first-status-missing":      set(`"type":"STOP","status":"EXPIRED"`, `"type":"STOP"`),
		"first-status-null":         set(`"type":"STOP","status":"EXPIRED"`, `"type":"STOP","status":null`),
		"first-status-number":       set(`"type":"STOP","status":"EXPIRED"`, `"type":"STOP","status":3`),
		"triggered-order-id-number": set(`"orderPrice":null}`, `"orderPrice":null,"triggeredOrderId":5}`),
		"triggered-order-id-object": set(`"orderPrice":null}`, `"orderPrice":null,"triggeredOrderId":{}}`),
		"second-string":             set(`"orderPrice":null}`, `"orderPrice":null},"second":"none"`),
		"second-array":              set(`"orderPrice":null}`, `"orderPrice":null},"second":[]`),
		"type-number":               set(`"type":"SINGLE"`, `"type":1`),
		// 중복 키는 마지막 값이 앞 값을 덮어 반대 증거를 지움(gstack 리뷰 Codex P1) — 엄격 해독이 거절해야 함.
		"dup-status":    set(`"status":"EXPIRED","symbol"`, `"status":"COMPLETED","status":"EXPIRED","symbol"`),
		"dup-first":     set(`"first":{`, `"first":{"type":"STOP","status":"COMPLETED","orderSide":"SELL","triggeredOrderId":"ORD-1"},"first":{`),
		"dup-triggered": set(`"orderPrice":null}`, `"orderPrice":null,"triggeredOrderId":"ORD-1","triggeredOrderId":null}`),
		"dup-second":    set(`"orderPrice":null}`, `"orderPrice":null},"second":{"type":"STOP","status":"CANCELED","orderSide":"SELL"},"second":null`),
	}
	for name, result := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeReconcileConditionalDetail([]byte(result)); !errors.Is(err, ErrReconcileSchema) {
				t.Fatalf("result %s: want ErrReconcileSchema, got %v", result, err)
			}
			s := newDetailServer(t, http.StatusOK, wrapResult(result))
			if _, err := detailClient(t, s).ReconcileConditionalOrder(context.Background(), detailTargetID); err == nil {
				t.Fatalf("client accepted result %s", result)
			}
		})
	}
}
