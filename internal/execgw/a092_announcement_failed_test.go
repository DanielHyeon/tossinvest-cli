package execgw_test

// a092 25라운드 보이스 B #7: 승격이 커밋되고 통지 기록만 실패하면(ErrModeAnnouncementFailed) 「승격이 원장에 닿지 않았다」가 아님.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

type a092FailingAnnouncer struct{}

func (a092FailingAnnouncer) AnnounceOperatingMode(context.Context, string, journal.OperatingModeRecord) error {
	return errors.New("outbox disk full")
}

func TestA092AnUnannouncedCredentialTighteningIsReportedAsTightened(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	r := &execgw.Retrier{Clock: clk, Gate: gate, Escalate: j, AccountRef: "acct-7", Announcer: a092FailingAnnouncer{}}
	err := r.Query(context.Background(), execgw.QueryHoldings, func(context.Context) error { return authFatal() })
	if err == nil {
		t.Fatal("a rejected credential must be returned to the caller")
	}
	if cur, _ := j.CurrentOperatingMode(context.Background(), "acct-7"); cur.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %s, want ENTRY_BLOCKED — the transition committed", cur.Mode)
	}
	if strings.Contains(err.Error(), "did not reach the operating mode") {
		t.Errorf("err = %q — the tightening did reach the operating mode; only its notice was not recorded", err)
	}
	if !errors.Is(err, journal.ErrModeAnnouncementFailed) {
		t.Errorf("err = %v, want the announcement failure kept for errors.Is", err)
	}
}
