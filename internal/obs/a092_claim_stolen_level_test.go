package obs_test

// a092 26라운드 보이스 C #2: 인수 줄 등급을 INFO 로 내린 logClaimHeld 는 「죽은 발송자의 held 행」 신호를 logClaimStolen(WARN)에
// 넘겼음 — 그 줄의 **등급**과 누가 쥐고 있었는지를 못 박음(줄 존재만 보던 a099 시험은 WARN → INFO 변이를 통과시켰음).

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA092ATakeoverIsAWarningThatNamesTheDeadSender(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	buf := &bytes.Buffer{}
	n := &obs.Notifier{
		Log:         obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
		Publisher:   &a099Publisher{},
		Journal:     j,
		Gate:        execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{}),
		Clock:       clock.System(),
		Attempts:    1,
		RetryDelay:  time.Millisecond,
		RemindAfter: a099Remind,
	}
	id, err := j.EnqueueAlert(ctx, journal.Alert{EventKey: a099Event().Key, Type: string(a099Event().Type), Severity: "critical"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ClaimAlertByID(ctx, id, "the-sender-that-died"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(journal.DefaultAlertLease + time.Second)
	if err := n.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	found := false
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) != nil || m["msg"] != string(obs.EventAlertClaimStolen) {
			continue
		}
		found = true
		if m["level"] != "WARN" {
			t.Errorf("takeover line level = %v, want WARN — it is the signal that a sender died holding a row", m["level"])
		}
		if m["claimed_by"] != "the-sender-that-died" {
			t.Errorf("claimed_by = %v, want the dead sender", m["claimed_by"])
		}
	}
	if !found {
		t.Fatalf("no %s line:\n%s", obs.EventAlertClaimStolen, buf.String())
	}
}
