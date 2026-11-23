package execgw

// a198 D7 분류 단위 시험 — 엔진 취소 정책의 `withCancelRefusal` 이 「이미 끝난 주문」 409 를 확정 거절로 닫는 경계.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func a198Outcome(status int, body string) journal.DispatchOutcome {
	return journal.DispatchOutcome{
		Class:      journal.DispatchAmbiguous,
		ReasonCode: journal.ReasonDispatchAmbiguous,
		Err:        &official.APIError{Code: status, Body: body},
		Detail:     "broker refused cancel",
	}
}

func a198Run(policy *CancelRefusalPolicy, kind journal.MutationKind, first journal.DispatchOutcome) (journal.DispatchOutcome, int) {
	g := &Gateway{clk: clock.NewFake(time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)), cancelRefusal: policy}
	calls := 0
	out := g.withCancelRefusal(mutationPlan{kind: kind}, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		calls++
		return first
	})(context.Background(), nil)
	return out, calls
}

func TestA198EngineSettledOrderRefusalsAreDefinitive(t *testing.T) {
	for _, body := range []string{
		`{"error":{"requestId":"r","code":"already-filled","message":"m"}}`,
		`{"error":{"requestId":"r","code":"already-canceled","message":"m"}}`,
		`{"error":{"requestId":"r","code":"already-rejected","message":"m"}}`,
		`{"code":"already-filled"}`,
		`{"error":{"code":"ALREADY-CANCELED"}}`,
		`{"code":"already-rejected","error":{"code":"Already-Rejected"}}`,
	} {
		t.Run(body, func(t *testing.T) {
			out, calls := a198Run(EngineCancelRefusal(), journal.KindCancel, a198Outcome(409, body))
			if calls != 1 || out.Class != journal.DispatchRejected || out.ReasonCode != journal.ReasonDispatchRejected {
				t.Fatalf("calls=%d outcome=%+v, want one call settled as a definitive %s", calls, out, journal.ReasonDispatchRejected)
			}
		})
	}
}

func TestA198OtherAnswersKeepTheirOutcome(t *testing.T) {
	cases := []struct {
		name   string
		policy *CancelRefusalPolicy
		kind   journal.MutationKind
		status int
		body   string
	}{
		{"already-modified stays in doubt", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":"already-modified"}}`},
		{"contradictory code places", EngineCancelRefusal(), journal.KindCancel, 409, `{"code":"already-filled","error":{"code":"other"}}`},
		{"unreadable code value", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":7}}`},
		{"suffixed code", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":"already-filled-v2"}}`},
		{"message only", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":"other","message":"already-filled"}}`},
		{"no code", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"message":"x"}}`},
		{"not json", EngineCancelRefusal(), journal.KindCancel, 409, `already-filled`},
		{"request-in-progress", EngineCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":"request-in-progress"}}`},
		{"place is not classified", EngineCancelRefusal(), journal.KindPlace, 409, `{"error":{"code":"already-filled"}}`},
		{"amend is not classified", EngineCancelRefusal(), journal.KindAmend, 409, `{"error":{"code":"already-filled"}}`},
		{"flatten policy is untouched", FlattenCancelRefusal(), journal.KindCancel, 409, `{"error":{"code":"already-canceled"}}`},
		{"nil policy is untouched", nil, journal.KindCancel, 409, `{"error":{"code":"already-canceled"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := a198Outcome(tc.status, tc.body)
			out, _ := a198Run(tc.policy, tc.kind, first)
			if out.Class != journal.DispatchAmbiguous || out.ReasonCode != first.ReasonCode {
				t.Fatalf("outcome=%+v, want the original ambiguous outcome", out)
			}
		})
	}
}

func TestA198AlreadyProcessingKeepsTheA165Path(t *testing.T) {
	out, calls := a198Run(EngineCancelRefusal(), journal.KindCancel,
		a198Outcome(409, `{"error":{"code":"already-processing","data":{"retryAfterSeconds":1}}}`))
	if calls != 1 || out.Class != journal.DispatchRejected || out.ReasonCode != string(ReasonBrokerRejected) {
		t.Fatalf("calls=%d outcome=%+v, want a165's engine definitive refusal (%s)", calls, out, ReasonBrokerRejected)
	}
}

func TestA198ANonAmbiguousOutcomeIsNotReclassified(t *testing.T) {
	first := a198Outcome(409, `{"error":{"code":"already-filled"}}`)
	first.Class = journal.DispatchAcked
	out, _ := a198Run(EngineCancelRefusal(), journal.KindCancel, first)
	if out.Class != journal.DispatchAcked {
		t.Fatalf("outcome=%+v, want the acknowledged outcome untouched", out)
	}
}
