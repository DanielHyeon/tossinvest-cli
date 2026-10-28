package execgw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const cancelAlreadyProcessingMarker = "[a165 "

// CancelRefusalPolicy 는 취소의 이미 처리 중 거절을 경로별로 다루는 정책임.
type CancelRefusalPolicy struct {
	Retries    int
	MaxWait    time.Duration
	Window     time.Duration
	Definitive bool
}

// EngineCancelRefusal 은 엔진 취소 거절을 같은 주기 재전송 없이 종결하는 정책임.
func EngineCancelRefusal() *CancelRefusalPolicy {
	return &CancelRefusalPolicy{Definitive: true}
}

// FlattenCancelRefusal 은 flatten 취소 거절을 제한된 창 안에서 재전송하는 정책임.
func FlattenCancelRefusal() *CancelRefusalPolicy {
	return &CancelRefusalPolicy{Retries: 2, MaxWait: 5 * time.Second, Window: 8 * time.Second}
}

// CancelRefusalPolicy 는 현재 게이트웨이에 주입된 취소 거절 정책을 복사해 돌려줌.
func (g *Gateway) CancelRefusalPolicy() (CancelRefusalPolicy, bool) {
	if g.cancelRefusal == nil {
		return CancelRefusalPolicy{}, false
	}
	return *g.cancelRefusal, true
}

// withCancelRefusal 는 취소 409 already-processing만 원 dispatch 바깥에서 처리함.
func (g *Gateway) withCancelRefusal(plan mutationPlan, dispatch journal.DispatchFunc) journal.DispatchFunc {
	if g.cancelRefusal == nil || plan.kind != journal.KindCancel {
		return dispatch
	}
	policy := *g.cancelRefusal
	return func(dctx context.Context, a *journal.Attempt) journal.DispatchOutcome {
		start := g.clk.Now()
		first := dispatch(dctx, a)
		wait, ok := alreadyProcessingRefusal(first.Err)
		if !ok || first.Class != journal.DispatchAmbiguous {
			return first
		}
		if policy.Definitive {
			out := decorateCancelRefusal(first, 1, "definitive", "definitive", journal.DispatchRejected)
			out.ReasonCode = string(ReasonBrokerRejected)
			return out
		}
		if policy.Retries <= 0 || policy.MaxWait <= 0 || policy.Window <= 0 {
			return decorateCancelRefusal(first, 1, "policy", "today", first.Class)
		}

		// winCtx 시한은 판정이 아니라 매달린 재전송을 끊는 실시간 상한임. 생산에서는 systemClock과 같은 시계라 창 판정과 일치함. Fake 시계 시험에서는 실시간 8초 전에 Clock 판정이 먼저 끝남.
		winCtx, cancel := context.WithTimeout(dctx, policy.Window)
		defer cancel()
		waited := time.Duration(0)
		sends := 1
		for retry := 0; retry < policy.Retries; retry++ {
			d := wait
			if remaining := policy.MaxWait - waited; d > remaining {
				d = remaining
			}
			if d <= 0 {
				return decorateCancelRefusal(first, sends, "max_wait", "today", first.Class)
			}
			if g.clk.Since(start)+d >= policy.Window {
				return decorateCancelRefusal(first, sends, "window", "today", first.Class)
			}
			if err := g.clk.Sleep(winCtx, d); err != nil {
				return decorateCancelRefusal(first, sends, "cancelled", "today", first.Class)
			}
			waited += d
			if g.clk.Since(start) >= policy.Window {
				return decorateCancelRefusal(first, sends, "window", "today", first.Class)
			}

			sends++
			next := dispatch(winCtx, a)
			// 재전송 중 context 종료는 응답 불명으로 보고 첫 409를 보존함.
			if winCtx.Err() != nil && (next.Class == journal.DispatchAmbiguous || next.Class == journal.DispatchNotSent) {
				return decorateCancelRefusal(first, sends, "cancelled", "today", first.Class)
			}
			nextWait, nextIsAlreadyProcessing := alreadyProcessingRefusal(next.Err)
			if !nextIsAlreadyProcessing || next.Class != journal.DispatchAmbiguous {
				if next.Class == journal.DispatchNotSent {
					return decorateCancelRefusal(first, sends, "not_sent", "today", first.Class)
				}
				return decorateCancelRefusal(next, sends, "ack", "retried", next.Class)
			}
			wait = nextWait
		}
		return decorateCancelRefusal(first, sends, "retries", "today", first.Class)
	}
}

// alreadyProcessingRefusal 은 두 code 자리가 모순되지 않는 409만 인식함.
func alreadyProcessingRefusal(err error) (time.Duration, bool) {
	var apiErr *official.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 409 {
		return 0, false
	}
	top, nested, ok := refusalCodes(apiErr.Body)
	if !ok || (top.present && nested.present && !strings.EqualFold(top.value, nested.value)) {
		return 0, false
	}
	code := top
	if !code.present {
		code = nested
	}
	if !code.present || !strings.EqualFold(code.value, "already-processing") {
		return 0, false
	}
	return retryAfter(apiErr.Body), true
}

// retryAfter 는 브로커의 대기 초를 보수적으로 기본 1초로 읽음.
func retryAfter(body string) time.Duration {
	var payload struct {
		Data struct {
			RetryAfterSeconds json.RawMessage `json:"retryAfterSeconds"`
		} `json:"data"`
		Error struct {
			Data struct {
				RetryAfterSeconds json.RawMessage `json:"retryAfterSeconds"`
			} `json:"data"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(body), &payload) != nil {
		return time.Second
	}
	raw := payload.Data.RetryAfterSeconds
	if len(raw) == 0 {
		raw = payload.Error.Data.RetryAfterSeconds
	}
	if len(raw) == 0 {
		return time.Second
	}
	var seconds float64
	if err := json.Unmarshal(raw, &seconds); err != nil || seconds <= 0 {
		return time.Second
	}
	return time.Duration(seconds * float64(time.Second))
}

// decorateCancelRefusal 은 latch가 읽는 공통 꼬리표를 dispatch 결과 끝에 붙임.
func decorateCancelRefusal(out journal.DispatchOutcome, sends int, stop, outcome string, class journal.DispatchClass) journal.DispatchOutcome {
	out.Class = class
	out.Detail = strings.TrimSpace(out.Detail)
	if out.Detail != "" {
		out.Detail += " "
	}
	out.Detail += fmt.Sprintf("%ssends=%d branch=cancel_already_processing stop=%s outcome=%s]", cancelAlreadyProcessingMarker, sends, stop, outcome)
	if out.ReasonCode == "" {
		out.ReasonCode = string(ReasonBrokerOutcomeUnknown)
	}
	return out
}
