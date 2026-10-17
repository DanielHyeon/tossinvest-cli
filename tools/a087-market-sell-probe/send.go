package main

// send.go 는 주문 본문을 정확히 한 번 전송하고 돌아온 것을 해석 없이 담음.
//
// 재시도가 없는 이유: 타임아웃·5xx·연결 끊김 뒤에는 첫 요청이 브로커에 닿았는지 알 수 없음.
// 다시 보내면 이중 주문이 될 수 있으므로 "불명" 으로 기록하고 사람이 주문 조회로 확인함.
// 그래서 internal/official 의 send(401 이면 토큰을 갈아 같은 본문을 다시 보냄)를 쓰지 않고
// 선례 execgw.HTTPReplay 와 같은 모양 — 공유 토큰 관리자의 헤더만 빌린 단일 POST — 으로 보냄.
//
// net/http 가 몰래 다시 보내는 길 둘도 막음:
//   - 리다이렉트: 307/308 은 본문을 다시 실어 따라감 → CheckRedirect 로 따라가지 않음.
//   - 연결 재사용 재시도: Transport 는 재사용 연결이 끊기면 replayable 요청을 다시 보냄 →
//     DisableKeepAlives 로 재사용 연결 자체를 없앰(POST 이고 Idempotency-Key 헤더도 없지만 이중으로 막음).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// maxResponseBytes 는 읽어 올 응답 상한임. 주문 응답은 작은 JSON 이고 그보다 크면 프록시 오류 페이지임.
const maxResponseBytes = 1 << 20

// 결과 분류 — 브로커 응답의 의미를 해석하지 않고 "주문이 생겼을 수 있는가" 만 가름.
const (
	outcomeAccepted = "accepted" // 2xx + orderId
	outcomeRefused  = "refused"  // 4xx(409 제외) — 브로커가 요청을 받아 거절함
	outcomeUnknown  = "unknown"  // 응답 없음·본문 못 읽음·5xx·409·2xx 인데 orderId 없음
)

// headerSource 는 공유 토큰 관리자에서 인증·계좌 헤더를 얻는 표면임(official.Client.AuthHeaders).
type headerSource interface {
	AuthHeaders(ctx context.Context) (map[string]string, error)
}

// sendResult 는 전송 한 번의 원자료임.
type sendResult struct {
	Sent           bool // 요청을 http.Client 에 넘겼는가
	HTTPStatus     int
	Body           []byte
	TransportError string
	SentAt         time.Time
	AnsweredAt     time.Time
}

// newSendClient 는 리다이렉트를 따라가지 않고 연결을 재사용하지 않는 전송 전용 client 임.
func newSendClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   true,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// buildRequest 는 본문 바이트를 그대로 실은 요청을 만듦. 헤더는 받은 것 + Content-Type 뿐임.
func buildRequest(ctx context.Context, baseURL, body string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(baseURL, "/")+orderPath, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// sendOnce 는 요청을 정확히 한 번 넘기고 그 답을 담음. 이 함수 안에 반복문이 없다는 것이
// 단일 전송의 계약이며 static_test.go 가 구조로 고정함.
func sendOnce(client *http.Client, req *http.Request, now func() time.Time) sendResult {
	result := sendResult{Sent: true, SentAt: now()}
	resp, err := client.Do(req)
	result.AnsweredAt = now()
	if err != nil {
		result.TransportError = err.Error()
		return result
	}
	defer resp.Body.Close()
	result.HTTPStatus = resp.StatusCode
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	result.Body = body
	if err != nil {
		result.TransportError = "reading the response body: " + err.Error()
	}
	return result
}

// brokerAnswer 는 응답 본문에서 읽어 둘 칸임 — 성공 `{"result":{orderId,clientOrderId}}`,
// 실패 `{"error":{code,message,data,requestId}}`(openapi OrderResponse·ErrorResponse).
type brokerAnswer struct {
	OrderID          string
	ClientOrderEcho  *string
	ErrorCode        string
	ErrorMessage     string
	EchoMatchesSent  *bool
	BodyIsJSONObject bool
}

func readAnswer(body []byte, sentKey string) brokerAnswer {
	var envelope struct {
		Result *struct {
			OrderID       string  `json:"orderId"`
			ClientOrderID *string `json:"clientOrderId"`
		} `json:"result"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	var answer brokerAnswer
	if err := json.Unmarshal(body, &envelope); err != nil {
		return answer
	}
	answer.BodyIsJSONObject = true
	if envelope.Result != nil {
		answer.OrderID = envelope.Result.OrderID
		answer.ClientOrderEcho = envelope.Result.ClientOrderID
		if answer.ClientOrderEcho != nil {
			match := *answer.ClientOrderEcho == sentKey
			answer.EchoMatchesSent = &match
		}
	}
	if envelope.Error != nil {
		answer.ErrorCode = envelope.Error.Code
		answer.ErrorMessage = envelope.Error.Message
	}
	return answer
}

// classify 는 "주문이 생겼을 수 있는가" 로만 가름. 의심스러우면 unknown 쪽으로 기움.
func classify(result sendResult, answer brokerAnswer) string {
	switch {
	case result.TransportError != "" || result.HTTPStatus == 0:
		return outcomeUnknown
	case result.HTTPStatus >= 200 && result.HTTPStatus < 300:
		// 다른 키를 되돌려준 2xx 는 이 요청에 대한 답이라고 볼 수 없음(orders_write.go PlaceOrder 와 같은 판단).
		if answer.OrderID == "" || (answer.EchoMatchesSent != nil && !*answer.EchoMatchesSent) {
			return outcomeUnknown
		}
		return outcomeAccepted
	case result.HTTPStatus == http.StatusConflict:
		// 409 request-in-progress: 같은 키의 요청이 처리 중 — 주문이 생겼을 수 있음.
		return outcomeUnknown
	case result.HTTPStatus >= 400 && result.HTTPStatus < 500:
		return outcomeRefused
	default:
		// 5xx·1xx·3xx(리다이렉트를 따라가지 않으므로 3xx 도 여기로 옴).
		return outcomeUnknown
	}
}

// headerNames 는 보낸 헤더의 이름만 정렬해 돌려줌 — 값(토큰·계좌 번호)은 영수증에 남기지 않음.
func headerNames(req *http.Request) []string {
	names := make([]string, 0, len(req.Header))
	for name := range req.Header {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var errNoHeaders = errors.New("the shared token manager returned no headers")
