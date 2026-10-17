package engine

// a135 — 엔진이 재시작할 때 읽을 capability attestation 의 만료 · 범위 상실은 재시작 전에 알려진다.
//
// 판정(judgeAttestation)은 인터록과 같은 술어(attest.Verify + RequiredEndpoints)를 쓰고, 실행자(attestationWatch)는
// 단계마다 한 번만 알린다. 여기서 재는 것:
//   - 경계: D-8 없음 · D-7/D-3 normal · D-2/D-1 critical · 만료 critical
//   - 범위 상실 · 계좌 불일치 · 파일 없음 → 즉시 critical
//   - 중복 방지(같은 단계 두 번 안 보냄) · 갱신(신원 변경) 뒤 다시 셈 · 기록 실패면 다음 판정에 다시 보냄
//   - 알림이 꺼진 엔진은 critical 대신 normal 타입
//   - 판정 패닉 · 오류가 실행자를 멈추지 않음, 게이트 꺼짐은 파일을 읽지 않음, 취소는 취소로 돌려줌
//   - 알림 문구 · 필드에 계좌 없음

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a135Account 는 시험용 계좌임 — 알림 바이트에 원문 · 숫자열 · 마스크 어느 꼴로도 나오면 안 됨.
const a135Account = "987-65-432109"

var a135Now = time.Date(2026, 10, 14, 3, 0, 0, 0, time.UTC)

// a135Attestation 은 now 기준 remaining 뒤 만료되는, 엔진 필수 endpoint 전부를 덮는 attestation 임.
func a135Attestation(remaining time.Duration) attest.Attestation {
	issued := a135Now.Add(remaining).Add(-30 * 24 * time.Hour)
	return attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    a135Account,
		IssuedAt:      issued,
		ExpiresAt:     a135Now.Add(remaining),
		SoakDays:      3,
		Endpoints:     RequiredEndpoints(),
		SupervisedBy: []attest.Proof{
			{Endpoint: "POST /api/v1/orders", At: issued.Add(-time.Hour), AccountRef: attest.Mask(a135Account), Market: "KR"},
			{Endpoint: "POST /api/v1/orders/{id}/cancel", At: issued.Add(-time.Hour), AccountRef: attest.Mask(a135Account), Market: "KR"},
		},
	}
}

// without 은 endpoint 하나를 뺀 사본임 — 갱신이 POST 범위를 잃은 모양.
func without(a attest.Attestation, endpoint string) attest.Attestation {
	var kept []string
	for _, e := range a.Endpoints {
		if e != endpoint {
			kept = append(kept, e)
		}
	}
	a.Endpoints = kept
	return a
}

func TestA135StageBoundaries(t *testing.T) {
	day := 24 * time.Hour
	cases := []struct {
		name      string
		remaining time.Duration
		stage     attestationStage
		days      int
		critical  bool
	}{
		{"D-8 은 아직 알리지 않음", 7*day + time.Second, stageHealthy, 8, false},
		{"30일 남음", 30 * day, stageHealthy, 30, false},
		{"정확히 7일 = D-7", 7 * day, stageNotice, 7, false},
		{"6.5일 = D-7", 6*day + 12*time.Hour, stageNotice, 7, false},
		{"정확히 3일 = D-3", 3 * day, stageNotice, 3, false},
		{"2일 1초 = D-3", 2*day + time.Second, stageNotice, 3, false},
		{"정확히 2일 = D-2", 2 * day, stageImminent, 2, true},
		{"1.5일 = D-2", day + 12*time.Hour, stageImminent, 2, true},
		{"정확히 1일 = D-1", day, stageImminent, 1, true},
		{"1초 = D-1", time.Second, stageImminent, 1, true},
		{"만료 순간", 0, stageExpired, 0, true},
		{"만료 1시간 뒤", -time.Hour, stageExpired, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := judgeAttestation(a135Attestation(tc.remaining), nil, a135Now, a135Account)
			if v.Stage != tc.stage {
				t.Fatalf("stage=%s want %s", v.Stage, tc.stage)
			}
			if tc.stage != stageExpired && v.DaysLeft != tc.days {
				t.Fatalf("days=%d want %d", v.DaysLeft, tc.days)
			}
			if v.critical() != tc.critical {
				t.Fatalf("critical=%v want %v", v.critical(), tc.critical)
			}
		})
	}
}

func TestA135ScopeLossIsCritical(t *testing.T) {
	for _, lost := range []string{"POST /api/v1/orders", "POST /api/v1/orders/{id}/cancel"} {
		// 만료는 30일 뒤 — 날수로는 알릴 일이 없어도 범위 상실은 즉시 critical 이어야 함.
		v := judgeAttestation(without(a135Attestation(30*24*time.Hour), lost), nil, a135Now, a135Account)
		if v.Stage != stageScopeLost || !v.critical() {
			t.Fatalf("%s lost: stage=%s critical=%v, want scope_lost critical", lost, v.Stage, v.critical())
		}
		if len(v.Missing) != 1 || v.Missing[0] != lost {
			t.Fatalf("missing=%v want [%s]", v.Missing, lost)
		}
	}
}

func TestA135UnusableAttestationsAreCritical(t *testing.T) {
	good := a135Attestation(30 * 24 * time.Hour)
	incomplete := good
	incomplete.Endpoints = nil
	cases := []struct {
		name    string
		att     attest.Attestation
		loadErr error
		account string
		stage   attestationStage
	}{
		{"파일 없음", attest.Attestation{}, fmt.Errorf("%w at x", attest.ErrMissing), a135Account, stageUnreadable},
		{"깨진 파일", attest.Attestation{}, fmt.Errorf("%w: parse", attest.ErrMalformed), a135Account, stageUnreadable},
		{"endpoint 없음", incomplete, nil, a135Account, stageIncomplete},
		{"다른 계좌", good, nil, "111-22-333444", stageAccountMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := judgeAttestation(tc.att, tc.loadErr, a135Now, tc.account)
			if v.Stage != tc.stage || !v.critical() {
				t.Fatalf("stage=%s critical=%v want %s critical", v.Stage, v.critical(), tc.stage)
			}
		})
	}
}

// --- 실행자 -------------------------------------------------------------------

type a135Notifier struct {
	mu     sync.Mutex
	events []obs.Event
	fail   int // 남은 실패 횟수
}

func (n *a135Notifier) Notify(_ context.Context, e obs.Event) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.fail > 0 {
		n.fail--
		return errors.New("outbox write failed")
	}
	n.events = append(n.events, e)
	return nil
}

func (n *a135Notifier) got() []obs.Event {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]obs.Event(nil), n.events...)
}

// a135Watch 는 fake 시계 · 기록 알림기 · 바꿔 끼우는 파일 내용으로 감시를 만듦.
func a135Watch(clk *clock.Fake, n *a135Notifier, critical bool, att *attest.Attestation) *attestationWatch {
	return &attestationWatch{
		enabled:    true,
		path:       "/unused/capability-attestation.json",
		accountRef: a135Account,
		critical:   critical,
		notifier:   n,
		clock:      clk,
		interval:   attestationWatchInterval,
		load: func(string) (attest.Attestation, error) {
			if att == nil {
				return attest.Attestation{}, fmt.Errorf("%w at x", attest.ErrMissing)
			}
			return *att, nil
		},
	}
}

func TestA135WatchAnnouncesEachStageOnce(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a135Attestation(8 * 24 * time.Hour) // D-8
	w := a135Watch(clk, n, true, &att)
	ctx := t.Context()

	w.checkOnce(ctx)
	if got := n.got(); len(got) != 0 {
		t.Fatalf("D-8 announced %d event(s): %+v", len(got), got)
	}

	clk.Advance(24 * time.Hour) // D-7
	w.checkOnce(ctx)
	clk.Advance(time.Hour) // 같은 D-7 — 다시 보내면 안 됨
	w.checkOnce(ctx)
	got := n.got()
	if len(got) != 1 || got[0].Type != obs.EventCredentialExpiryNotice {
		t.Fatalf("D-7: events=%+v, want one %s", got, obs.EventCredentialExpiryNotice)
	}
	if obs.SeverityOf(got[0].Type) != obs.SeverityNormal {
		t.Fatalf("D-7 severity=%s want normal", obs.SeverityOf(got[0].Type))
	}

	clk.Set(att.ExpiresAt.Add(-6 * 24 * time.Hour)) // 다음 날 D-6 — 하루 한 번
	w.checkOnce(ctx)
	w.checkOnce(ctx)
	clk.Set(att.ExpiresAt.Add(-3 * 24 * time.Hour)) // D-3
	w.checkOnce(ctx)
	clk.Set(att.ExpiresAt.Add(-2 * 24 * time.Hour)) // D-2
	w.checkOnce(ctx)
	clk.Set(att.ExpiresAt.Add(-24 * time.Hour)) // D-1
	w.checkOnce(ctx)
	w.checkOnce(ctx)
	clk.Set(att.ExpiresAt) // 만료
	w.checkOnce(ctx)
	w.checkOnce(ctx)

	got = n.got()
	want := []obs.EventType{
		obs.EventCredentialExpiryNotice, // D-7
		obs.EventCredentialExpiryNotice, // D-6
		obs.EventCredentialExpiryNotice, // D-3
		obs.EventCredentialExpiring,     // D-2
		obs.EventCredentialExpiring,     // D-1
		obs.EventCredentialExpiring,     // 만료
	}
	if len(got) != len(want) {
		t.Fatalf("events=%d want %d: %+v", len(got), len(want), got)
	}
	keys := map[string]bool{}
	for i, e := range got {
		if e.Type != want[i] {
			t.Errorf("event %d type=%s want %s", i, e.Type, want[i])
		}
		if keys[e.Key] {
			t.Errorf("event %d reuses key %q — outbox 가 다른 단계를 같은 조건으로 읽음", i, e.Key)
		}
		keys[e.Key] = true
	}
	if obs.SeverityOf(obs.EventCredentialExpiring) != obs.SeverityCritical {
		t.Fatal("EventCredentialExpiring 은 critical 이어야 함(engine-safety 등급화된 알림)")
	}
}

func TestA135RenewalResetsAndScopeLossIsImmediate(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a135Attestation(2 * 24 * time.Hour) // D-2
	w := a135Watch(clk, n, true, &att)
	ctx := t.Context()

	w.checkOnce(ctx)
	if got := n.got(); len(got) != 1 || got[0].Type != obs.EventCredentialExpiring {
		t.Fatalf("D-2: %+v", got)
	}

	// 갱신: 새 신원 · 30일 — 알릴 것 없음.
	clk.Advance(time.Hour)
	att = a135Attestation(30 * 24 * time.Hour)
	att.IssuedAt = clk.Now()
	att.ExpiresAt = clk.Now().Add(30 * 24 * time.Hour)
	w.checkOnce(ctx)
	if got := n.got(); len(got) != 1 {
		t.Fatalf("healthy renewal announced: %+v", got[1:])
	}

	// 다음 갱신이 POST cancel 을 잃음 — 날수와 무관하게 즉시 critical, 한 번만.
	clk.Advance(time.Hour)
	att = without(a135Attestation(30*24*time.Hour), "POST /api/v1/orders/{id}/cancel")
	att.IssuedAt = clk.Now()
	att.ExpiresAt = clk.Now().Add(30 * 24 * time.Hour)
	w.checkOnce(ctx)
	w.checkOnce(ctx)
	got := n.got()
	if len(got) != 2 || got[1].Type != obs.EventCredentialExpiring {
		t.Fatalf("scope loss: %+v", got)
	}
	if !strings.Contains(got[1].Body, "POST /api/v1/orders/{id}/cancel") {
		t.Errorf("scope-loss body does not name the missing endpoint: %q", got[1].Body)
	}
}

func TestA135NotificationsOffNeverRaisesCritical(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a135Attestation(-time.Hour) // 만료
	w := a135Watch(clk, n, false, &att)
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 || got[0].Type != obs.EventCredentialExpiryNotice {
		t.Fatalf("notifications off: %+v, want one normal %s", got, obs.EventCredentialExpiryNotice)
	}
}

func TestA135FailedRecordIsRetried(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{fail: 1}
	w := a135Watch(clk, n, true, nil) // 파일 없음 → critical
	w.checkOnce(t.Context())
	if len(n.got()) != 0 {
		t.Fatal("a failed record was counted as sent")
	}
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 || got[0].Type != obs.EventCredentialExpiring {
		t.Fatalf("retry: %+v", got)
	}
}

func TestA135AlertCarriesNoAccount(t *testing.T) {
	forms := []string{a135Account, "98765432109", attest.Mask(a135Account), attest.Mask("98765432109")}
	cases := map[string]attest.Attestation{
		"notice":     a135Attestation(5 * 24 * time.Hour),
		"imminent":   a135Attestation(24 * time.Hour),
		"expired":    a135Attestation(-time.Hour),
		"scope_lost": without(a135Attestation(30*24*time.Hour), "POST /api/v1/orders"),
		"incomplete": func() attest.Attestation { a := a135Attestation(30 * 24 * time.Hour); a.Endpoints = nil; return a }(),
	}
	for name, att := range cases {
		t.Run(name, func(t *testing.T) {
			n := &a135Notifier{}
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 {
				t.Fatalf("events=%+v", got)
			}
			e := got[0]
			text := e.Title + "\n" + e.Body + "\n" + e.Key
			for k, v := range e.Fields {
				text += "\n" + k + "=" + fmt.Sprint(v)
			}
			for _, form := range forms {
				if strings.Contains(text, form) {
					t.Fatalf("alert carries account form %q:\n%s", form, text)
				}
			}
			for k := range e.Fields {
				if strings.Contains(strings.ToLower(k), "account") {
					t.Fatalf("alert field %q is an account key", k)
				}
			}
		})
	}
	// 계좌 불일치도 — 두 계좌 어느 쪽도 나오면 안 됨.
	n := &a135Notifier{}
	att := a135Attestation(30 * 24 * time.Hour)
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.accountRef = "111-22-333444"
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 {
		t.Fatalf("mismatch events=%+v", got)
	}
	text := got[0].Title + got[0].Body + got[0].Key + fmt.Sprint(got[0].Fields)
	for _, form := range append(forms, "111-22-333444", "11122333444", attest.Mask("111-22-333444")) {
		if strings.Contains(text, form) {
			t.Fatalf("mismatch alert carries %q: %s", form, text)
		}
	}
}

// TestA135CheckPanicDoesNotStopTheWatch: 판정 한 번의 패닉은 그 판정에서 끝나고 다음 판정이 돎.
func TestA135CheckPanicDoesNotStopTheWatch(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	w := a135Watch(clk, n, true, nil)
	calls := 0
	w.load = func(string) (attest.Attestation, error) {
		calls++
		if calls == 1 {
			panic("boom")
		}
		return attest.Attestation{}, fmt.Errorf("%w at x", attest.ErrMissing)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	if !clk.WaitForSleepers(1, 5*time.Second) {
		t.Fatal("the watch did not reach its sleep after a panicking check")
	}
	clk.Advance(attestationWatchInterval)
	deadline := time.Now().Add(5 * time.Second)
	for len(n.got()) == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := n.got(); len(got) != 1 {
		t.Fatalf("second check did not run after the panic: %+v", got)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run returned %v, want context.Canceled (graceful stop)", err)
	}
}

// TestA135GateOffReadsNothing: 게이트가 꺼진 엔진은 attestation 에 의존하지 않으므로 읽지도 알리지도 않음.
func TestA135GateOffReadsNothing(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	w := a135Watch(clk, n, true, nil)
	w.enabled = false
	w.load = func(string) (attest.Attestation, error) {
		t.Error("gate-off watch read the attestation")
		return attest.Attestation{}, nil
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := w.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run=%v want context.Canceled", err)
	}
	if len(n.got()) != 0 {
		t.Fatal("gate-off watch announced")
	}
}

// TestA135ProductionExecutorShape: 생산 실행자는 감독 밖 · 자기 정지 이벤트 · 정지 처리 없음(래치 없음).
func TestA135ProductionExecutorShape(t *testing.T) {
	c := &Context{AccountRef: a135Account}
	c.Automation = AutomationStatus{Enabled: true, Verified: true, AttestationFile: "/x/capability-attestation.json"}
	aux := c.AttestationWatch(clock.NewFake(a135Now))
	if aux.Name != "attestation-watch" || aux.Run == nil {
		t.Fatalf("aux=%+v", aux)
	}
	if aux.StopEvent != obs.EventAttestationWatchFailed {
		t.Fatalf("StopEvent=%q want %q — 배달 실행자의 이벤트를 빌리면 그 필터가 이 정지를 잡음", aux.StopEvent, obs.EventAttestationWatchFailed)
	}
	if aux.OnStop != nil {
		t.Fatal("the attestation watch has a stop handler; its stop must not latch anything")
	}
	if obs.SeverityOf(obs.EventAttestationWatchFailed) != obs.SeverityNormal ||
		obs.SeverityOf(obs.EventCredentialExpiryNotice) != obs.SeverityNormal {
		t.Fatal("the watch's own events must be normal-grade")
	}
	// nil 알림기 포인터가 nil 이 아닌 인터페이스로 들어가면 Notify 가 패닉함 — 생성자가 막아야 함.
	w := newAttestationWatch(c, clock.NewFake(a135Now))
	if w.notifier != nil {
		t.Fatal("a nil *obs.Notifier became a non-nil notifier interface")
	}
	if w.path != c.Automation.AttestationFile || w.accountRef != a135Account || !w.enabled {
		t.Fatalf("watch=%+v does not read the interlock's file for the engine's account", w)
	}
}

// TestA135UnreadableAlertCarriesNoAccountOrPath: 파일 없음 단계의 알림도 계좌 · 경로를 싣지 않음.
func TestA135UnreadableAlertCarriesNoAccountOrPath(t *testing.T) {
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, nil)
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 || got[0].Fields["reason"] != string(stageUnreadable) {
		t.Fatalf("events=%+v", got)
	}
	text := got[0].Title + got[0].Body + got[0].Key + fmt.Sprint(got[0].Fields)
	for _, form := range []string{a135Account, "98765432109", attest.Mask(a135Account), w.path, "capability-attestation.json"} {
		if strings.Contains(text, form) {
			t.Fatalf("unreadable alert carries %q: %s", form, text)
		}
	}
}

// TestA135SameStageOnANewAttestationIsAnnouncedAgain: 신원이 바뀌면 같은 단계라도 다시 알림(a135 리뷰 P2 — 신원 재설정 고정).
func TestA135SameStageOnANewAttestationIsAnnouncedAgain(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a135Attestation(-time.Hour) // 만료된 A
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())

	b := a135Attestation(-2 * time.Hour) // 만료된 B — 같은 날 · 같은 단계, 다른 신원
	att = b
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 2 {
		t.Fatalf("a different expired attestation at the same stage was swallowed: %+v", got)
	}
	if got[0].Key == got[1].Key {
		t.Fatalf("two attestations share outbox key %q — 알림기가 둘째를 「이미 보냄」으로 삼킴", got[0].Key)
	}

	// 같은 단계 · 같은 날수의 갱신도 다시 셈: D-5 의 A, 같은 D-5 의 C(발급만 다름).
	clk2 := clock.NewFake(a135Now)
	n2 := &a135Notifier{}
	a := a135Attestation(5 * 24 * time.Hour)
	w2 := a135Watch(clk2, n2, true, &a)
	w2.checkOnce(t.Context())
	a.IssuedAt = a.IssuedAt.Add(time.Minute)
	w2.checkOnce(t.Context())
	if got := n2.got(); len(got) != 2 || got[0].Key == got[1].Key {
		t.Fatalf("issued-at change was not a new identity: %+v", got)
	}
}

// TestA135InvalidStatesAreRemindedDaily: 재시작이 거절될 상태는 고칠 때까지 하루 한 번 다시 알림(한 번 알리고 침묵하지 않게).
func TestA135InvalidStatesAreRemindedDaily(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a135Attestation(-time.Hour)
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	clk.Advance(time.Hour)
	w.checkOnce(t.Context())
	if len(n.got()) != 1 {
		t.Fatalf("expired re-announced within the day: %+v", n.got())
	}
	clk.Advance(24 * time.Hour)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 || got[1].Type != obs.EventCredentialExpiring {
		t.Fatalf("expired was not reminded the next day: %+v", got)
	}
}

// TestA135SupervisedProofAtIsTheBindingProof: 같은 endpoint 의 오래된 증거가 새 증거를 가리지 않음 — endpoint 별 최신의 최소.
func TestA135SupervisedProofAtIsTheBindingProof(t *testing.T) {
	old := a135Now.Add(-40 * 24 * time.Hour)
	fresh := a135Now.Add(-2 * 24 * time.Hour)
	cancelAt := a135Now.Add(-5 * 24 * time.Hour)
	got := binding([]attest.Proof{
		{Endpoint: "POST /api/v1/orders", At: old},
		{Endpoint: "POST /api/v1/orders", At: fresh},
		{Endpoint: "post  /api/v1/orders/{id}/cancel", At: cancelAt},
	})
	if !got.Equal(cancelAt) {
		t.Fatalf("binding=%s want %s (cancel's newest; orders' newest is later)", got, cancelAt)
	}
	if !binding(nil).IsZero() {
		t.Fatal("no proofs must give zero")
	}
}

// a135PanickingNotifier 는 알림기 잠금 안의 패닉을 흉내 냄.
type a135PanickingNotifier struct{}

func (a135PanickingNotifier) Notify(context.Context, obs.Event) error {
	panic("panic under the notifier lock")
}

// TestA135NotifyPanicIsNotRecovered: 알림기 안의 패닉은 감시의 회수에도 런타임의 회수(runAuxiliaryBody)에도 잡히지 않고
// 프로세스를 내림(a135 리뷰 P1). 잡히면 알림기 잠금이 영구히 잡힌 채 엔진이 살아 있어 청산 관측의 알림이 조용히 멈춤.
// 프로세스 종료를 재려고 시험 바이너리를 자식으로 다시 실행함.
func TestA135NotifyPanicIsNotRecovered(t *testing.T) {
	if os.Getenv("A135_NOTIFY_PANIC_CHILD") == "1" {
		w := a135Watch(clock.NewFake(a135Now), nil, true, nil)
		w.notifier = a135PanickingNotifier{}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = runAuxiliaryBody(ctx, AuxiliaryExecutor{Name: attestationWatchName, Run: func(ctx context.Context) error {
			w.checkOnce(ctx)
			return ctx.Err()
		}})
		fmt.Println("A135-SWALLOWED")
		os.Exit(0)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestA135NotifyPanicIsNotRecovered$", "-test.count=1")
	cmd.Env = append(os.Environ(), "A135_NOTIFY_PANIC_CHILD=1")
	out, err := cmd.CombinedOutput()
	if strings.Contains(string(out), "A135-SWALLOWED") {
		t.Fatalf("a panic inside Notify was recovered; the notifier lock would stay held:\n%s", out)
	}
	if err == nil || !strings.Contains(string(out), "panic under the notifier lock") {
		t.Fatalf("child did not die of the notifier panic (err=%v):\n%s", err, out)
	}
}
