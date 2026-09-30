package obs

// a091 tasks 3.8 (escalate 팔) — 승격의 두 로그 줄(성공 · 실패)이 계좌 원문을 싣지 않음.
//
// a091 의 새 critical 은 기록이 실패하면 recordCritical → escalate 로 흐름. 그래서 그 두 줄은 a091 이 새로 도달시키는 표면이고
// (Manager 판정 2026-10-01 — 불변식 8(a)), 계좌 필드를 뺌. 알림기는 계좌 하나에 묶이므로 줄의 뜻은 잃지 않음.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA091TheEscalationLinesCarryNoAccount(t *testing.T) {
	cases := map[string]func(t *testing.T) *Notifier{
		// 성공: 새 원장 · NORMAL 계정 → 승격이 바뀜(changed) → Warn 줄.
		"escalated": func(t *testing.T) *Notifier {
			return &Notifier{Journal: a092OpenJournal(t), AccountRef: a092Acct, Clock: clock.System()}
		},
		// 실패: 닫힌 원장 → 승격 쓰기 실패 → Error 줄.
		"failed": func(t *testing.T) *Notifier {
			j := a092OpenJournal(t)
			_ = j.Close()
			return &Notifier{Journal: j, AccountRef: a092Acct, Clock: clock.System()}
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			buf := &a092SyncBuffer{}
			n := build(t)
			n.Log = a092JSONLogger(buf)
			included, _ := n.escalate(context.Background(), Event{Type: EventExitStopSoldNothing})
			if !included {
				t.Fatalf("arrangement: the escalation was not attempted")
			}
			found := false
			for _, l := range a092Lines(t, buf.String()) {
				if a092Value(l, "msg") != string(EventOperatingMode) {
					continue
				}
				found = true
				if v := a092Value(l, FieldAccount); v != "" {
					t.Errorf("the escalation line carries the account field %q", v)
				}
			}
			if !found {
				t.Fatalf("no escalation line:\n%s", buf.String())
			}
			if strings.Contains(buf.String(), a092Acct) {
				t.Errorf("the account %q appears in the escalation log:\n%s", a092Acct, buf.String())
			}
		})
	}
}
