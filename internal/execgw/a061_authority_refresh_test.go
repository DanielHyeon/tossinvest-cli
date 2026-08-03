package execgw_test

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestEntryGateRefreshesDurableAuthorityBeforeEveryEntryDecision(t *testing.T) {
	gate := execgw.NewEntryGate(clock.NewFake(fixedNow), map[execgw.RequiredQuery]time.Duration{})
	calls := 0
	gate.SetAuthorityRefresh(func() error {
		calls++
		gate.BlockSymbol("us", "AAPL", execgw.ReasonReconcilePermanent, "durable conflict")
		return nil
	})

	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("entry decision = %v, want freshly projected durable block", rejected)
	}
	if calls != 1 {
		t.Fatalf("authority refresh calls = %d, want one per entry decision", calls)
	}
}

func TestEntryGateFailsClosedWhenDurableAuthorityCannotBeRead(t *testing.T) {
	gate := execgw.NewEntryGate(clock.NewFake(fixedNow), map[execgw.RequiredQuery]time.Duration{})
	gate.SetAuthorityRefresh(func() error { return errors.New("journal unavailable") })

	rejected := gate.CheckEntryFor("us", "AAPL")
	if rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("entry decision = %v, want fail-closed reconcile refusal", rejected)
	}
}
