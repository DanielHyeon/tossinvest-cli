package flatten_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// a165FlattenBroker 는 bounded retry가 실제 flatten 정착 경로를 지나는지 확인함.
type a165FlattenBroker struct{ calls int }

func (b *a165FlattenBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	return domain.MutationResult{}, errors.New("not used")
}

func (b *a165FlattenBroker) CancelPendingOrder(context.Context, orderintent.CancelIntent) (domain.MutationResult, error) {
	b.calls++
	return domain.MutationResult{}, &official.APIError{Code: 409, Body: `{"code":"already-processing","data":{"retryAfterSeconds":1}}`}
}

func (b *a165FlattenBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	return domain.MutationResult{}, errors.New("not used")
}

func (b *a165FlattenBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func TestA165FlattenWiringPolicyKeepsExhaustedCancelInDoubt(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{
		order("O-a165", "005930", "BUY", "10", "70000", "KRW"),
	}})
	broker := &a165FlattenBroker{}
	clk := autoClock{h.clock}
	gw, err := execgw.New(execgw.Options{
		Journal: h.journal,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock: clk, AccountRef: "acct-7", Source: "a165-flatten-test", Entry: h.gate,
		CancelRefusal: execgw.FlattenCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	h.gateway = gw
	s := h.saga(false)
	s.Clock = clk
	report, err := s.CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if broker.calls != 3 {
		t.Fatalf("cancel calls=%d, want initial plus two retries", broker.calls)
	}
	if report.InDoubt != 1 || report.Failed != 0 || report.Cancelled != 0 {
		t.Fatalf("flatten report=%+v, want one IN_DOUBT", report)
	}
	steps, err := h.journal.FlattenSteps(context.Background(), report.SagaID)
	if err != nil || len(steps) != 1 || steps[0].AttemptID == "" {
		t.Fatalf("flatten steps=%+v err=%v, want one journalled attempt", steps, err)
	}
	record, err := h.journal.LookupAttempt(context.Background(), steps[0].AttemptID)
	if err != nil || record.State != journal.StateInDoubt {
		t.Fatalf("cancel record=%+v err=%v, want IN_DOUBT", record, err)
	}
}
