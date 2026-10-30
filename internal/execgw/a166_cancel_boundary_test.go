//go:build linux

package execgw_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

type a166CancelBroker struct{ a165BoundaryBroker }

func (b *a166CancelBroker) CancelPendingOrder(ctx context.Context, in orderintent.CancelIntent) (domain.MutationResult, error) {
	if b.calls == 0 {
		return b.a165BoundaryBroker.CancelPendingOrder(ctx, in)
	}
	b.calls++
	return domain.MutationResult{}, accountbudget.ErrNotSent
}
func TestA166CancelRetryAdmissionRefusalPreservesFirstAmbiguous(t *testing.T) {
	clk := a165AdvancingClock{Fake: clock.NewFake(fixedNow)}
	j := openJournal(t, clk)
	broker := &a166CancelBroker{a165BoundaryBroker{retryAfter: "0.001"}}
	gw, err := execgw.New(execgw.Options{Journal: j, Trading: trading.NewService(config.Trading{Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true}, broker), Clock: clk, AccountRef: "acct-7", Source: "a166-boundary", CancelRefusal: &execgw.CancelRefusalPolicy{Retries: 1, MaxWait: time.Second, Window: time.Second}})
	if err != nil {
		t.Fatal(err)
	}
	out, _ := a165BoundaryCancel(t, gw, j, clk)
	if broker.calls != 2 || out.State != journal.StateInDoubt {
		t.Fatalf("calls=%d state=%s", broker.calls, out.State)
	}
}
