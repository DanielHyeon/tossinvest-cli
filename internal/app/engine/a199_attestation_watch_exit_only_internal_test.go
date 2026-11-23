package engine

// a199_attestation_watch_exit_only_internal_test.go — 감시가 사실대로 말하기(a199 D7 · Manager Q-A · Q-C).
//
//   - 낡음 단계(expired · scope_lost)의 귀결은 「지금 엔진을 재시작하면 새 진입이 막힌 청산 전용으로 뜬다」.
//   - 날수 · 지남 통지(notice · imminent · proof_overdue)의 기한 뒤 귀결도 「청산 전용」이지 「기동 거부」가 아님.
//   - 엔진이 청산 전용으로 뜬 프로세스면(기동 표지) 파일 단계 알림의 단계 · 등급 · 키는 그대로, 본문에 한 문장이 덧붙음.
//   - 파일이 healthy 인데 청산 전용이면 일반 등급 회복 안내 하루 한 번 — 같은 날 재발급으로 기한이 바뀌어도 다시 안 나감(R2-3).
//   - 기동 표지가 정상이면 덧붙는 문장 · 회복 안내 0.

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	a199OnRestartExitOnly = "지금 엔진을 재시작하면 새 진입이 막힌 청산 전용으로 뜬다"
	a199RunningExitOnly   = "지금 이 엔진은 청산 전용으로 돌고 있다"
	a199Recovered         = "증명은 회복됐다 — 엔진을 재시작하면 attestation 사유의 진입 차단이 풀린다"
)

func TestA199StaleStagesTellTheExitOnlyConsequence(t *testing.T) {
	cases := map[string]attest.Attestation{
		"expired":    a135Attestation(-time.Hour),
		"scope_lost": without(a135Attestation(20*24*time.Hour), "POST /api/v1/orders"),
	}
	for stage, att := range cases {
		t.Run(stage, func(t *testing.T) {
			n := &a135Notifier{}
			att := att
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 || got[0].Fields[obs.FieldReason] != stage {
				t.Fatalf("events = %+v", got)
			}
			e := got[0]
			if !strings.Contains(e.Title, a199OnRestartExitOnly) || !strings.HasPrefix(e.Body, a199OnRestartExitOnly) {
				t.Errorf("title/body do not state the exit-only consequence:\n%q\n%q", e.Title, e.Body)
			}
			if strings.Contains(e.Title+e.Body, "기동이 거부된다") || strings.Contains(e.Body, "자동 손절이 멈춘다") {
				t.Errorf("a stale stage still says the start is refused or the stop halts: %q", e.Body)
			}
			if e.Type != obs.EventCredentialExpiring {
				t.Errorf("type = %s, want the critical type (grading unchanged)", e.Type)
			}
			if strings.Contains(e.Body, a199RunningExitOnly) {
				t.Errorf("a normally started engine must not say it is running exit-only: %q", e.Body)
			}
		})
	}
}

func TestA199DayCountAndOverdueNoticesTellTheExitOnlyConsequence(t *testing.T) {
	cases := map[string]attest.Attestation{
		"notice (expiry)":   a135Attestation(5 * 24 * time.Hour),
		"imminent (expiry)": a135Attestation(36 * time.Hour),
		// gstack review(testing): 감독 증거 기한의 통지 · 지남 단계도 같은 귀결을 말해야 함(attestation_watch.go 의 proof 갈래).
		"notice (proof)": a196ProofDue(a135Now, 5*a196Day),
		"proof_overdue":  a196ProofDue(a135Now, -time.Hour),
	}
	for name, att := range cases {
		t.Run(name, func(t *testing.T) {
			n := &a135Notifier{}
			att := att
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 {
				t.Fatalf("events = %+v", got)
			}
			text := got[0].Title + got[0].Body
			if strings.Contains(text, "재시작이 거부") || strings.Contains(text, "기동이 거부") {
				t.Errorf("a day-count notice still predicts a refused restart: %q", text)
			}
			if !strings.Contains(text, "청산 전용") {
				t.Errorf("a day-count notice does not say the restart comes up exit-only: %q", text)
			}
		})
	}
}

func TestA199AnExitOnlyEngineAddsOneSentenceToTheStageAlert(t *testing.T) {
	att := a135Attestation(-time.Hour)
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.exitOnly = true
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 || got[0].Fields[obs.FieldReason] != "expired" || got[0].Type != obs.EventCredentialExpiring {
		t.Fatalf("events = %+v, want the expired stage alert, critical", got)
	}
	if !strings.Contains(got[0].Body, a199RunningExitOnly) {
		t.Errorf("body does not say the engine is running exit-only now: %q", got[0].Body)
	}

	// 거부 유지 단계도 귀결 문장은 그대로, 한 문장만 덧붙음.
	n2 := &a135Notifier{}
	w2 := a135Watch(clock.NewFake(a135Now), n2, true, nil) // 파일 없음
	w2.exitOnly = true
	w2.checkOnce(t.Context())
	got2 := n2.got()
	if len(got2) != 1 || !strings.HasPrefix(got2[0].Body, "지금 엔진을 재시작하면 기동이 거부된다") ||
		!strings.Contains(got2[0].Body, a199RunningExitOnly) {
		t.Fatalf("unreadable on an exit-only engine = %+v", got2)
	}
}

func TestA199ANoticeOnAnExitOnlyEngineCarriesTheRecoveryAndIsNotDoubled(t *testing.T) {
	att := a135Attestation(5 * 24 * time.Hour) // notice
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.exitOnly = true
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 {
		t.Fatalf("events = %d, want one — the day-count notice carries the recovery sentence, nothing separate", len(got))
	}
	if got[0].Fields[obs.FieldReason] != "notice" || got[0].Type != obs.EventCredentialExpiryNotice {
		t.Fatalf("event = %+v, want the notice stage at normal grade", got[0])
	}
	if !strings.Contains(got[0].Body, a199Recovered) {
		t.Errorf("notice body does not carry the recovery sentence: %q", got[0].Body)
	}
}

func TestA199AHealthyFileOnAnExitOnlyEngineIsAnnouncedOncePerDay(t *testing.T) {
	clk := clock.NewFake(a135Now)
	att := a135Attestation(20 * 24 * time.Hour) // healthy
	n := &a135Notifier{}
	w := a135Watch(clk, n, true, &att)
	w.exitOnly = true
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 {
		t.Fatalf("events = %d, want the recovery notice", len(got))
	}
	if got[0].Type != obs.EventCredentialExpiryNotice {
		t.Errorf("type = %s, want the normal grade — a recovered proof is not an emergency", got[0].Type)
	}
	if !strings.Contains(got[0].Title, a199Recovered) {
		t.Errorf("title = %q, want the recovery sentence", got[0].Title)
	}

	// 같은 날 재발급으로 기한이 바뀜 — 회복 안내는 다시 안 나감(R2-3: sent 가 기한 변경으로 비워져도).
	clk.Advance(time.Hour)
	att = a135Attestation(25 * 24 * time.Hour)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 {
		t.Fatalf("events after a same-day reissue = %d, want still 1", len(got))
	}

	// 다음 날 다시 한 번.
	clk.Advance(24 * time.Hour)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 {
		t.Fatalf("events on the next day = %d, want 2", len(got))
	}
}

func TestA199AHealthyFileOnANormalEngineSaysNothing(t *testing.T) {
	att := a135Attestation(20 * 24 * time.Hour)
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 0 {
		t.Fatalf("a healthy file on a normally started engine produced %d event(s)", len(got))
	}
}

func TestA199TheWatchReadsTheBootMarkerAndWritesNothing(t *testing.T) {
	c := &Context{Automation: AutomationStatus{Enabled: true, ExitOnly: true, AttestationFile: "/x"}}
	before := c.Automation
	w := newAttestationWatch(c, clock.NewFake(a135Now))
	if !w.exitOnly {
		t.Fatal("the watch did not take the boot-time exit-only marker")
	}
	if c.Automation != before {
		t.Fatal("building the watch changed the engine's automation status")
	}
}
