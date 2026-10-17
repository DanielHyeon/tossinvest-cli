package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestRunRefusesToStartWithoutTheAPIKey 는 키 부재 시 시작 거부 + 사유를 고정함.
// 키 검사는 플래그·자격증명·네트워크보다 앞서므로 아무 인자 없이도 같은 거부가 나와야 함.
func TestRunRefusesToStartWithoutTheAPIKey(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "   "} {
		getenv := func(name string) string {
			if name == typeSafeAPIKeyEnv {
				return value
			}
			return ""
		}
		// --config-dir 를 빈 임시 디렉터리로: 키 검사가 사라지는 변이가 생겨도 실 자격증명에 닿지 않음.
		err := dispatch(context.Background(), []string{"run", "--market", "KR", "--config-dir", t.TempDir()}, getenv, io.Discard)
		if !errors.Is(err, errMissingAPIKey) {
			t.Fatalf("run without %s: err = %v", typeSafeAPIKeyEnv, err)
		}
		if !strings.Contains(err.Error(), typeSafeAPIKeyEnv) || !strings.Contains(err.Error(), "refuses to start") {
			t.Fatalf("the refusal does not say why: %v", err)
		}
	}
	if _, err := newJudgeClient("", "", ""); !errors.Is(err, errMissingAPIKey) {
		t.Fatalf("newJudgeClient without a key: %v", err)
	}
}

func TestJudgeSendsTheDocumentedRequestAndReadsNoul(t *testing.T) {
	t.Parallel()
	mock := newTypeSafeMock(t)
	client := mock.judgeClient(t)
	result, err := client.judge(context.Background(), PublicState{Market: "KR", Symbol: "005930", News: newsAbsent}, probeQuestions())
	if err != nil {
		t.Fatal(err)
	}
	if result.Model != "jev-1.13.0" || result.P["j1_up_within_horizon"] != 0.93 || result.P["j2_hold_off_red_flag"] != 0.12 {
		t.Fatalf("result = %+v", result)
	}
	if result.InputTokens != 321 || result.Attempts != 1 {
		t.Fatalf("usage/attempts = %+v", result)
	}
	var request map[string]json.RawMessage
	if err := json.Unmarshal(mock.requests()[0], &request); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"state", "model", "questions"} {
		if _, found := request[key]; !found {
			t.Fatalf("request has no %q: %s", key, mock.requests()[0])
		}
	}
	if len(request) != 3 {
		t.Fatalf("request carries undocumented top-level fields: %s", mock.requests()[0])
	}
	var model string
	_ = json.Unmarshal(request["model"], &model)
	if model != defaultJudgeModel {
		t.Fatalf("model = %q, want the pinned %q", model, defaultJudgeModel)
	}
	var questions map[string]noulQuestion
	_ = json.Unmarshal(request["questions"], &questions)
	if questions["j1_up_within_horizon"].Type != "noul" || questions["j2_hold_off_red_flag"].Instructions == "" {
		t.Fatalf("questions = %+v", questions)
	}
}

func TestJudgeRetriesTransientStatusesThenGivesUp(t *testing.T) {
	t.Parallel()
	mock := newTypeSafeMock(t)
	mock.statuses = []int{http.StatusTooManyRequests, 529}
	client := mock.judgeClient(t)
	var delays []time.Duration
	client.sleep = func(_ context.Context, d time.Duration) error { delays = append(delays, d); return nil }
	result, err := client.judge(context.Background(), PublicState{}, probeQuestions())
	if err != nil || result.Attempts != 3 {
		t.Fatalf("two transient refusals then success: result=%+v err=%v", result, err)
	}
	if len(delays) != 2 || delays[0] != 500*time.Millisecond || delays[1] != time.Second {
		t.Fatalf("backoff = %v, want [500ms 1s]", delays)
	}

	mock.statuses = []int{500, 502, 503}
	if _, err := client.judge(context.Background(), PublicState{}, probeQuestions()); err == nil ||
		!strings.Contains(err.Error(), "gave up after 3 attempts") {
		t.Fatalf("three 5xx must give up after 3 attempts: %v", err)
	}
}

func TestJudgeDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()
	for _, status := range []int{http.StatusUnauthorized, http.StatusUnprocessableEntity} {
		mock := newTypeSafeMock(t)
		mock.statuses = []int{status}
		client := mock.judgeClient(t)
		_, err := client.judge(context.Background(), PublicState{}, probeQuestions())
		var httpErr *judgeHTTPError
		if !errors.As(err, &httpErr) || httpErr.Status != status {
			t.Fatalf("status %d: err = %v", status, err)
		}
		if len(mock.requests()) != 1 {
			t.Fatalf("status %d was retried %d times", status, len(mock.requests())-1)
		}
		if strings.Contains(err.Error(), testTypeSafeKey) {
			t.Fatalf("the error carries the API key: %v", err)
		}
	}
}

func TestJudgeHonoursRetryAfterSeconds(t *testing.T) {
	t.Parallel()
	client := &judgeClient{}
	if got := client.retryDelay(1, &judgeHTTPError{Status: 429, RetryAfter: "3"}); got != 3*time.Second {
		t.Fatalf("Retry-After 3 → %v", got)
	}
	if got := client.retryDelay(1, &judgeHTTPError{Status: 429, RetryAfter: "999"}); got != judgeRetryAfterCap {
		t.Fatalf("Retry-After is capped: %v", got)
	}
	if got := client.retryDelay(5, &judgeHTTPError{Status: 503}); got != judgeBackoffMax {
		t.Fatalf("backoff is capped: %v", got)
	}
}

func TestJudgeRefusesContractViolations(t *testing.T) {
	t.Parallel()
	for name, answer := range map[string]string{
		"wrong type":   `{"type":"choice","choice":"a"}`,
		"no value":     `{"type":"noul"}`,
		"out of range": `{"type":"noul","noul":1.5}`,
		"negative":     `{"type":"noul","noul":-0.1}`,
	} {
		mock := newTypeSafeMock(t)
		mock.answer = func(string) string { return answer }
		if _, err := mock.judgeClient(t).judge(context.Background(), PublicState{}, probeQuestions()); err == nil ||
			!strings.Contains(err.Error(), "contract") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if _, err := decodeJudgeResponse([]byte(`{"model":"jev-1.13.0","answers":{}}`), probeQuestions()); err == nil {
		t.Fatal("a missing answer must be refused")
	}
	if _, err := decodeJudgeResponse([]byte(`{"answers":{}}`), nil); err == nil {
		t.Fatal("a response without model must be refused")
	}
}
