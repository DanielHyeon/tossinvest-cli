package main

// judge.go 는 TypeSafe System One HTTP API(Noul) 호출 하나를 담당함.
//
// 계약은 analysis/typesafe-api-contract.md 에 원문 인용으로 동결됨 — 여기의 경로·헤더·필드명은
// 전부 그 문서에서 옮긴 것이고, 문서에 없는 필드는 읽지도 보내지도 않음.
// SDK 를 쓰지 않는 이유: Go SDK 부재 + 새 외부 의존 금지(design.md 「TypeSafe 연동」).

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 계약 상수 — typesafe-api-contract.md 1·4·5 절.
const (
	typeSafeAPIKeyEnv   = "TYPESAFE_API_KEY"
	defaultTypeSafeBase = "https://api.typesafe.ai"
	systemOnePath       = "/v1/systemone"
	// 버전 핀: alias(jev-latest)는 측정 도중 움직일 수 있어 p 의 모집단을 바꿈(models.md 인용).
	defaultJudgeModel = "jev-1.13.0"

	judgeAttemptTimeout = 10 * time.Second
	judgeMaxRetries     = 2
	judgeBackoffInitial = 500 * time.Millisecond
	judgeBackoffMax     = 5 * time.Second
	judgeRetryAfterCap  = 30 * time.Second
	judgeErrorBodyLimit = 200
)

// errMissingAPIKey 는 키 없이 시작하려 할 때의 거부 사유임.
var errMissingAPIKey = errors.New(typeSafeAPIKeyEnv + " is not set; the probe refuses to start without a TypeSafe API key " +
	"(supply it through the environment only — never store it in the repository, memory or logs)")

// noulQuestion 은 Noul 질문 하나의 전송 모양(api.md 「Noul」)임.
type noulQuestion struct {
	Type         string        `json:"type"`
	Instructions string        `json:"instructions"`
	Criteria     *noulCriteria `json:"criteria,omitempty"`
}

type noulCriteria struct {
	True  string `json:"true"`
	False string `json:"false"`
}

// question 은 원장에 남길 질문 식별(id·rev·원문 digest)과 전송 본문을 함께 묶음.
// 질문이 바뀌면 p 의 모집단이 달라지므로 rev 와 digest 를 행마다 박음(design.md 「판단 설계」).
type question struct {
	ID   string
	Rev  string
	Body noulQuestion
}

// digest 는 질문 전송 본문의 sha256 — rev 를 손으로 안 올린 질문 변경도 원장에서 드러나게 함.
func (q question) digest() string {
	raw, _ := json.Marshal(q.Body)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// systemOneRequest 의 State 가 PublicState 타입으로 고정된 것이 계좌 데이터 차단의 첫 겹임:
// 이 요청에는 공개 시장 데이터 구조체 말고 다른 값을 담을 자리가 없음.
type systemOneRequest struct {
	State     PublicState             `json:"state"`
	Model     string                  `json:"model"`
	Questions map[string]noulQuestion `json:"questions"`
}

type noulAnswer struct {
	Type string   `json:"type"`
	Noul *float64 `json:"noul"`
}

type systemOneResponse struct {
	Model   string                `json:"model"`
	Answers map[string]noulAnswer `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// judgeResult 는 한 요청의 결과 — 질문 id 별 p 와 답한 모델 버전.
type judgeResult struct {
	Model        string
	P            map[string]float64
	InputTokens  int
	OutputTokens int
	Attempts     int
}

// judgeClient 는 TypeSafe 호출기임. sleep 은 시험에서 대기 시간을 없애려고 주입함.
type judgeClient struct {
	base  string
	key   string
	model string
	hc    *http.Client
	sleep func(context.Context, time.Duration) error
}

// newJudgeClient 는 키가 비면 거부함 — 키 없는 프로브는 원장에 아무 의미 없는 결손만 쌓음.
func newJudgeClient(key, base, model string) (*judgeClient, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errMissingAPIKey
	}
	if base == "" {
		base = defaultTypeSafeBase
	}
	if model == "" {
		model = defaultJudgeModel
	}
	return &judgeClient{
		base:  strings.TrimRight(base, "/"),
		key:   key,
		model: model,
		hc:    &http.Client{},
		sleep: sleepContext,
	}, nil
}

// judge 는 같은 state 에 질문 여럿을 한 요청으로 묻고(noul.md 「ask more than one question per call」),
// 응답이 계약과 어긋나면 고쳐 읽지 않고 오류를 냄.
func (c *judgeClient) judge(ctx context.Context, state PublicState, questions []question) (judgeResult, error) {
	request := systemOneRequest{State: state, Model: c.model, Questions: map[string]noulQuestion{}}
	for _, q := range questions {
		request.Questions[q.ID] = q.Body
	}
	body, err := json.Marshal(request)
	if err != nil {
		return judgeResult{}, fmt.Errorf("encoding the request: %w", err)
	}
	// 둘째 겹: 직렬화된 바이트 자체를 열쇠 이름으로 다시 검사함(타입이 바뀌어도 여기서 멈춤).
	if err := assertPublicJSON(body); err != nil {
		return judgeResult{}, err
	}

	var lastErr error
	for attempt := 0; attempt <= judgeMaxRetries; attempt++ {
		if attempt > 0 {
			if err := c.sleep(ctx, c.retryDelay(attempt, lastErr)); err != nil {
				return judgeResult{}, err
			}
		}
		raw, status, header, err := c.post(ctx, body)
		if err != nil {
			// 전송 오류는 재시도 대상(SDK 기본 api_connection_error/api_timeout_error = True).
			lastErr = &judgeHTTPError{Transport: err}
			continue
		}
		if status >= 200 && status < 300 {
			result, err := decodeJudgeResponse(raw, questions)
			if err != nil {
				return judgeResult{}, err
			}
			result.Attempts = attempt + 1
			return result, nil
		}
		lastErr = &judgeHTTPError{Status: status, Body: truncate(string(raw), judgeErrorBodyLimit), RetryAfter: header.Get("Retry-After")}
		if !retryableStatus(status) {
			return judgeResult{}, lastErr
		}
	}
	return judgeResult{}, fmt.Errorf("gave up after %d attempts: %w", judgeMaxRetries+1, lastErr)
}

// post 는 시도 한 번 — 시도당 10s 상한(SDK DEFAULT_TIMEOUT = 10.0).
func (c *judgeClient) post(ctx context.Context, body []byte) ([]byte, int, http.Header, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, judgeAttemptTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, c.base+systemOnePath, bytes.NewReader(body))
	if err != nil {
		return nil, 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, 0, nil, err
	}
	return raw, resp.StatusCode, resp.Header, nil
}

// retryDelay 는 Retry-After(초 정수)가 있으면 그것을, 없으면 0.5s 에서 2배씩 최대 5s.
func (c *judgeClient) retryDelay(attempt int, lastErr error) time.Duration {
	var httpErr *judgeHTTPError
	if errors.As(lastErr, &httpErr) && httpErr.RetryAfter != "" {
		if seconds, err := strconv.Atoi(strings.TrimSpace(httpErr.RetryAfter)); err == nil && seconds >= 0 {
			return min(time.Duration(seconds)*time.Second, judgeRetryAfterCap)
		}
	}
	delay := judgeBackoffInitial << (attempt - 1)
	return min(delay, judgeBackoffMax)
}

// retryableStatus — 408·429·5xx(529 포함). 401·422 등은 재시도해도 같은 답임.
func retryableStatus(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500
}

// decodeJudgeResponse 는 p = answers.<id>.noul 경로만 읽음(contract 3 절).
func decodeJudgeResponse(raw []byte, questions []question) (judgeResult, error) {
	var response systemOneResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return judgeResult{}, fmt.Errorf("contract: response is not the documented JSON: %w", err)
	}
	if response.Model == "" {
		return judgeResult{}, errors.New("contract: response carries no model")
	}
	result := judgeResult{
		Model:        response.Model,
		P:            map[string]float64{},
		InputTokens:  response.Usage.InputTokens,
		OutputTokens: response.Usage.OutputTokens,
	}
	for _, q := range questions {
		answer, found := response.Answers[q.ID]
		if !found {
			return judgeResult{}, fmt.Errorf("contract: no answer for question %q", q.ID)
		}
		if answer.Type != "noul" {
			return judgeResult{}, fmt.Errorf("contract: answer %q has type %q, not noul", q.ID, answer.Type)
		}
		if answer.Noul == nil {
			return judgeResult{}, fmt.Errorf("contract: answer %q carries no noul value", q.ID)
		}
		p := *answer.Noul
		if math.IsNaN(p) || p < 0 || p > 1 {
			return judgeResult{}, fmt.Errorf("contract: answer %q noul %v is outside 0..1", q.ID, p)
		}
		result.P[q.ID] = p
	}
	return result, nil
}

// judgeHTTPError 는 상태 코드·본문 앞부분만 싣음. 요청 헤더(키)는 절대 담지 않음.
type judgeHTTPError struct {
	Status     int
	Body       string
	RetryAfter string
	Transport  error
}

func (e *judgeHTTPError) Error() string {
	if e.Transport != nil {
		return "typesafe transport: " + e.Transport.Error()
	}
	return fmt.Sprintf("typesafe HTTP %d: %s", e.Status, e.Body)
}

func truncate(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "…"
}

// sleepContext 는 ctx 취소를 존중하는 대기.
func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
