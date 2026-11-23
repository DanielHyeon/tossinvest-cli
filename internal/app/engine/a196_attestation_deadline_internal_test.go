package engine

// a196 — 재시작이 거절되기 시작할 시각(기한)은 만료와 감독 증거의 기한 중 이른 쪽이고, 거절 상태는 그 결과로 하루 한 번 알려짐.
//
// 여기서 재는 것(design §5):
//   - 감독 증거의 기한 = 실린 증거 중 가장 이른 것 + (만료 − 발급) — 만료 30일 뒤여도 감독 증거가 7일 뒤면 D-7
//   - 같은 감독 증거의 재발급은 날수 단계를 새로 세지 않음, 새 감독 증거는 새로 셈
//   - 기한이 지났는데 파일이 통과하면 proof_overdue critical 하루 한 번
//   - 거절 상태는 재발급을 넘겨 하루 한 번, 누락 집합·단계·KST 날짜가 바뀌면 새로 알림 — outbox Key 문자열로도 확인
//   - 거절 알림은 제목과 본문 첫 문장에서 「지금 엔진을 재시작하면 기동이 거부된다」 를 단정
// 픽스처는 전부 합성 값(a135 의 계좌 · 시각 상수 재사용).

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	a196Day      = 24 * time.Hour
	a196Validity = 30 * a196Day
	// a196Refused 는 거절 상태 알림이 제목과 본문 첫 문장에서 단정해야 하는 결과임.
	a196Refused = "지금 엔진을 재시작하면 기동이 거부된다"
	// a196Overdue 는 새 단계의 영문 코드임(구현 전 RED 가 컴파일되게 문자열로 둠).
	a196Overdue = attestationStage("proof_overdue")
)

// a196Attestation 은 issued 에 30일 유효로 발급되고 주문 · 취소 감독 증거를 orderAt · cancelAt 에 가진 attestation 임.
// 목록 순서는 취소 → 주문 — 가장 이른 증거(주문)가 첫 원소가 아니게 해서 binding 을 우회한 기한 계산을 잡음(R2-3 iv).
func a196Attestation(issued, orderAt, cancelAt time.Time) attest.Attestation {
	return attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    a135Account,
		IssuedAt:      issued,
		ExpiresAt:     issued.Add(a196Validity),
		SoakDays:      3,
		Endpoints:     RequiredEndpoints(),
		SupervisedBy: []attest.Proof{
			{Endpoint: "POST /api/v1/orders/{id}/cancel", At: cancelAt, AccountRef: attest.Mask(a135Account), Market: "KR"},
			{Endpoint: "POST /api/v1/orders", At: orderAt, AccountRef: attest.Mask(a135Account), Market: "KR"},
		},
	}
}

// a196ProofDue 는 now 에 발급되어 만료는 30일 뒤이고, 감독 증거의 기한이 now+left 인 attestation 임.
// 취소 증거는 주문 증거보다 3일 새것 — 기한을 정하는 것은 주문 증거.
func a196ProofDue(now time.Time, left time.Duration) attest.Attestation {
	order := now.Add(left).Add(-a196Validity)
	return a196Attestation(now, order, order.Add(3*a196Day))
}

// a196Reissued 는 같은 감독 증거를 싣고 at 에 다시 발급된 사본임 — 갱신 타이머가 쓰는 모양.
func a196Reissued(a attest.Attestation, at time.Time) attest.Attestation {
	a.IssuedAt = at
	a.ExpiresAt = at.Add(a196Validity)
	return a
}

// TestA196ProofDeadlineGovernsTheDayCount: 만료는 30일 뒤여도 감독 증거의 기한이 7일 뒤면 D-7(§5-1, Scenario 「감독 증거의 기한 7일 전」).
func TestA196ProofDeadlineGovernsTheDayCount(t *testing.T) {
	if v := judgeAttestation(a196ProofDue(a135Now, 7*a196Day+time.Second), nil, a135Now, a135Account); v.Stage != stageHealthy {
		t.Fatalf("7d1s before the proof deadline: stage=%s days=%d, want healthy", v.Stage, v.DaysLeft)
	}
	att := a196ProofDue(a135Now, 7*a196Day)
	v := judgeAttestation(att, nil, a135Now, a135Account)
	if v.Stage != stageNotice || v.DaysLeft != 7 {
		t.Fatalf("proof deadline in 7d, expiry in 30d: stage=%s days=%d, want notice D-7", v.Stage, v.DaysLeft)
	}

	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 || got[0].Type != obs.EventCredentialExpiryNotice {
		t.Fatalf("events=%+v, want one normal notice", got)
	}
	deadline := a135Now.Add(7 * a196Day)
	if got[0].Fields["deadline_kind"] != "supervised_proof" || got[0].Fields["deadline_at"] != deadline.UTC().Format(time.RFC3339) {
		t.Fatalf("fields=%v, want the supervised-proof deadline %s", got[0].Fields, deadline.UTC().Format(time.RFC3339))
	}
	body := got[0].Body
	for _, want := range []string{deadline.In(kst).Format("2006-01-02 15:04 KST"), "다른 유효한 감독 증거가 없으면", "tossctl verify run"} {
		if !strings.Contains(body, want) {
			t.Errorf("notice body lacks %q: %s", want, body)
		}
	}
}

// TestA196SameProofReissueIsNotAnnouncedAgain: 같은 감독 증거를 실은 재발급은 같은 날수 단계 안에서 다시 알리지 않고, 날수가 줄면 알림(§5-2).
func TestA196SameProofReissueIsNotAnnouncedAgain(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a196ProofDue(a135Now, 6*a196Day+12*time.Hour) // D-7
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())

	clk.Advance(6 * time.Hour) // 갱신 타이머 — 발급 · 만료는 6시간 밀리고 증거는 그대로, 여전히 D-7
	att = a196Reissued(att, clk.Now())
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 {
		t.Fatalf("a reissue carrying the same proof was announced again: %+v", got)
	}

	clk.Set(a135Now.Add(6*a196Day + 12*time.Hour).Add(-5*a196Day - 12*time.Hour)) // D-6
	att = a196Reissued(att, clk.Now())
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 || got[0].Key == got[1].Key {
		t.Fatalf("D-6 after D-7: %+v, want a second notice with a new key", got)
	}
}

// TestA196DeadlineBoundaries: 기한 경계 — 엔진 쪽 계산이 발급자 경계(나이 == 유효 기간이면 제외)와 같은 값을 씀(D2 의 engine 쪽).
func TestA196DeadlineBoundaries(t *testing.T) {
	cases := []struct {
		name  string
		left  time.Duration
		stage attestationStage
		days  int
	}{
		{"2일 1초 = D-3", 2*a196Day + time.Second, stageNotice, 3},
		{"정확히 2일 = D-2", 2 * a196Day, stageImminent, 2},
		{"정확히 1일 = D-1", a196Day, stageImminent, 1},
		{"증거 나이 V−1ns = 1ns 남음 = D-1", time.Nanosecond, stageImminent, 1},
		{"증거 나이 정확히 V = 기한 정각 = 지남", 0, a196Overdue, 0},
		{"증거 나이 V+1ns = 지남", -time.Nanosecond, a196Overdue, 0},
		{"하루 반 지남", -a196Day - 12*time.Hour, a196Overdue, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := judgeAttestation(a196ProofDue(a135Now, tc.left), nil, a135Now, a135Account)
			if v.Stage != tc.stage || v.DaysLeft != tc.days {
				t.Fatalf("stage=%s days=%d, want %s %d", v.Stage, v.DaysLeft, tc.stage, tc.days)
			}
			if !v.critical() && tc.stage != stageNotice {
				t.Fatalf("stage %s must be critical", v.Stage)
			}
		})
	}
}

// TestA196OverdueIsCriticalOncePerDay: 기한이 지났는데 파일이 통과하면 critical 하루 한 번 — 같은 날 재판정 · 재발급은 0(§5-3,
// Scenario 「감독 증거의 기한이 지났고 갱신이 실패한다」).
func TestA196OverdueIsCriticalOncePerDay(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a196ProofDue(a135Now, -time.Hour)
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	clk.Advance(time.Hour)
	w.checkOnce(t.Context())
	clk.Advance(5 * time.Hour)
	att = a196Reissued(att, clk.Now())
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 || got[0].Type != obs.EventCredentialExpiring || got[0].Fields[obs.FieldReason] != string(a196Overdue) {
		t.Fatalf("first overdue day: %+v, want one critical proof_overdue", got)
	}
	clk.Set(a135Now.Add(a196Day))
	w.checkOnce(t.Context())
	clk.Set(a135Now.Add(2 * a196Day))
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 3 {
		t.Fatalf("overdue for two more days: %d events, want 3: %+v", len(got), got)
	}
	// 마지막 알림은 재발급된 파일에서 나옴 — 본문의 만료 시각은 그 파일의 것.
	for _, want := range []string{"기한이 지났다", "다른 유효한 감독 증거가 없으면", att.ExpiresAt.In(kst).Format("2006-01-02 15:04 KST"), "tossctl verify run"} {
		if !strings.Contains(got[2].Body, want) {
			t.Errorf("overdue body lacks %q: %s", want, got[2].Body)
		}
	}
}

// TestA196NewSupervisedProofRestartsTheCount: 새 감독 증거로 기한이 30일 뒤가 되면 알림 0, 그 뒤 새 기한으로 다시 셈(§5-4).
func TestA196NewSupervisedProofRestartsTheCount(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := a196ProofDue(a135Now, 2*a196Day)
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())

	clk.Advance(time.Hour)
	fresh := clk.Now().Add(-time.Hour)
	att = a196Attestation(clk.Now(), fresh, fresh)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 {
		t.Fatalf("a fresh supervised proof was announced: %+v", got)
	}
	clk.Set(fresh.Add(a196Validity).Add(-7 * a196Day))
	att = a196Reissued(att, clk.Now())
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 || got[1].Type != obs.EventCredentialExpiryNotice || got[1].Fields["days_left"] != 7 {
		t.Fatalf("new deadline D-7: %+v", got)
	}
}

// TestA196WithoutSupervisedProofCountsTheExpiry: 감독 증거가 실리지 않은 파일은 만료만 셈(§5-5).
func TestA196WithoutSupervisedProofCountsTheExpiry(t *testing.T) {
	att := a196Attestation(a135Now.Add(7*a196Day-a196Validity), time.Time{}, time.Time{})
	att.SupervisedBy = nil
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, true, &att)
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 1 || got[0].Fields["days_left"] != 7 || got[0].Fields["deadline_kind"] != "expiry" {
		t.Fatalf("expiry-only D-7: %+v", got)
	}
	// 어느 기한을 세는지를 문구로도 말함(정본) — 만료가 기한인 통지가 감독 증거 문구로 새지 않게(gstack review 시험 보강).
	if !strings.Contains(got[0].Title, "검증서(attestation) 만료 D-7") || strings.Contains(got[0].Title, "감독하 실주문 검증") {
		t.Errorf("expiry notice title=%q", got[0].Title)
	}
	if !strings.Contains(got[0].Body, att.ExpiresAt.In(kst).Format("2006-01-02 15:04 KST")) || strings.Contains(got[0].Body, "tossctl verify run") {
		t.Errorf("expiry notice body does not tell the expiry deadline: %s", got[0].Body)
	}
}

// TestA196ScopeLossOncePerDayAcrossReissues: 범위 상실은 재발급을 넘겨 하루 한 번, 누락 집합이 바뀌면 같은 날에도 알림(§5-6).
func TestA196ScopeLossOncePerDayAcrossReissues(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	base := a196ProofDue(a135Now, 20*a196Day)
	att := without(base, "POST /api/v1/orders")
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	for i := 0; i < 2; i++ { // 같은 결함을 가진 파일이 갱신으로 두 번 더 쓰임
		clk.Advance(3 * time.Hour)
		att = a196Reissued(att, clk.Now())
		w.checkOnce(t.Context())
	}
	if got := n.got(); len(got) != 1 {
		t.Fatalf("scope loss re-announced on reissue within the day: %+v", got)
	}
	clk.Set(a135Now.Add(a196Day))
	att = a196Reissued(att, clk.Now())
	w.checkOnce(t.Context())
	clk.Advance(time.Hour)
	att = without(att, "POST /api/v1/orders/{id}/cancel") // 같은 날 하나 더 빠짐
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 3 {
		t.Fatalf("next day + a larger missing set: %d events, want 3: %+v", len(got), got)
	}
	if got[1].Key == got[2].Key {
		t.Fatalf("different missing sets share outbox key %q", got[1].Key)
	}
}

// TestA196RestartedWatchReusesTheRejectionKey: 재시작한 감시가 같은 날 같은 거절 상태를 다시 판정하면 outbox Key 가 같음 —
// 가짜 알림기는 중복을 거르지 않으므로 건수가 아니라 Key 문자열로 잼(R2-3 i). 실제 outbox 는 이 Key 로 같은 행을 찾음.
func TestA196RestartedWatchReusesTheRejectionKey(t *testing.T) {
	keyOf := func(att attest.Attestation, at time.Time) string {
		n := &a135Notifier{}
		w := a135Watch(clock.NewFake(at), n, true, &att)
		w.checkOnce(t.Context())
		got := n.got()
		if len(got) != 1 {
			t.Fatalf("events=%+v", got)
		}
		return got[0].Key
	}
	lost := without(a196ProofDue(a135Now, 20*a196Day), "POST /api/v1/orders")
	k1 := keyOf(lost, a135Now)
	k2 := keyOf(a196Reissued(lost, a135Now.Add(6*time.Hour)), a135Now.Add(6*time.Hour))
	if k1 != k2 {
		t.Fatalf("a reissue of the same scope loss got a new outbox key: %q vs %q", k1, k2)
	}
	both := without(lost, "POST /api/v1/orders/{id}/cancel")
	if k3 := keyOf(both, a135Now); k3 == k1 {
		t.Fatalf("a different missing set reused outbox key %q", k1)
	}
	expA := a135Attestation(-time.Hour)
	expB := a135Attestation(-2 * time.Hour)
	if ka, kb := keyOf(expA, a135Now), keyOf(expB, a135Now); ka != kb {
		t.Fatalf("two expired files on the same day got different outbox keys: %q vs %q", ka, kb)
	}
}

// TestA196RejectionDayIsTheKSTDate: UTC 날짜는 같고 KST 날짜만 바뀌는 두 시각(23:59 → 00:01 KST)에 각각 알림(R2-3 ii).
func TestA196RejectionDayIsTheKSTDate(t *testing.T) {
	before := time.Date(2026, 10, 14, 14, 59, 0, 0, time.UTC) // KST 23:59
	after := time.Date(2026, 10, 14, 15, 1, 0, 0, time.UTC)   // KST 다음 날 00:01, 같은 UTC 날짜
	clk := clock.NewFake(before)
	n := &a135Notifier{}
	att := a135Attestation(-time.Hour)
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	clk.Set(after)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 || got[0].Key == got[1].Key {
		t.Fatalf("KST midnight did not start a new day: %+v", got)
	}
}

// TestA196ScopeLossThenExpiryAreTwoAlerts: 같은 날 범위 상실 → 만료로 단계가 바뀌면 두 알림이 각각, Key 도 다름(R2-3 iii).
func TestA196ScopeLossThenExpiryAreTwoAlerts(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	att := without(a135Attestation(time.Hour), "POST /api/v1/orders")
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	clk.Advance(2 * time.Hour) // 파일 그대로 만료 시각을 넘김
	w.checkOnce(t.Context())
	got := n.got()
	if len(got) != 2 || got[0].Fields[obs.FieldReason] != string(stageScopeLost) || got[1].Fields[obs.FieldReason] != string(stageExpired) {
		t.Fatalf("scope_lost → expired: %+v", got)
	}
	if got[0].Key == got[1].Key {
		t.Fatalf("scope_lost and expired share outbox key %q", got[0].Key)
	}
}

// TestA196RefusalTellsTheConsequenceFirst: 재시작이 거절될 상태의 알림은 제목과 본문 첫 문장에서 결과를 단정하고, 본문에 손절 · 할 일(§5-7, Q3).
func TestA196RefusalTellsTheConsequenceFirst(t *testing.T) {
	good := a196ProofDue(a135Now, 20*a196Day)
	incomplete := good
	incomplete.Endpoints = nil
	cases := map[string]struct {
		att     *attest.Attestation
		account string
	}{
		"scope_lost":       {a196Ptr(without(good, "POST /api/v1/orders")), a135Account},
		"expired":          {a196Ptr(a135Attestation(-time.Hour)), a135Account},
		"incomplete":       {&incomplete, a135Account},
		"account_mismatch": {&good, "111-22-333444"},
		"unreadable":       {nil, a135Account},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			n := &a135Notifier{}
			w := a135Watch(clock.NewFake(a135Now), n, true, tc.att)
			w.accountRef = tc.account
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 || got[0].Fields[obs.FieldReason] != name {
				t.Fatalf("events=%+v", got)
			}
			e := got[0]
			if !strings.Contains(e.Title, a196Refused) || strings.Contains(e.Title, "수 있다") {
				t.Errorf("title does not state the refusal as a fact: %q", e.Title)
			}
			if !strings.HasPrefix(e.Body, a196Refused) {
				t.Errorf("body's first sentence is not the refusal: %q", e.Body)
			}
			if !strings.Contains(e.Body, "손절") || !strings.Contains(e.Body, "할 일") {
				t.Errorf("body lacks the stop-loss consequence or the remedy: %q", e.Body)
			}
			if name == "scope_lost" {
				for _, want := range []string{"POST /api/v1/orders", "tossctl verify run"} {
					if !strings.Contains(e.Body, want) {
						t.Errorf("scope-loss body lacks %q: %q", want, e.Body)
					}
				}
			}
		})
	}
}

func a196Ptr(a attest.Attestation) *attest.Attestation { return &a }

// TestA196NewAlertsCarryNoAccount: 새 문구(감독 증거 날수 · 임박 · 지남)에도 계좌 원문 · 숫자열 · 마스크 꼴이 없음(§5-8).
func TestA196NewAlertsCarryNoAccount(t *testing.T) {
	forms := []string{a135Account, "98765432109", attest.Mask(a135Account), attest.Mask("98765432109")}
	for name, left := range map[string]time.Duration{"notice": 5 * a196Day, "imminent": a196Day, "overdue": -time.Hour} {
		t.Run(name, func(t *testing.T) {
			att := a196ProofDue(a135Now, left)
			n := &a135Notifier{}
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 {
				t.Fatalf("events=%+v", got)
			}
			text := got[0].Title + "\n" + got[0].Body + "\n" + got[0].Key + "\n" + fmt.Sprint(got[0].Fields)
			for _, form := range forms {
				if strings.Contains(text, form) {
					t.Fatalf("alert carries account form %q:\n%s", form, text)
				}
			}
		})
	}
}

// TestA196NotificationsOffDowngradesOverdue: 알림이 꺼진 엔진은 proof_overdue 도 일반 등급으로 기록만(§5-9).
func TestA196NotificationsOffDowngradesOverdue(t *testing.T) {
	att := a196ProofDue(a135Now, -time.Hour)
	n := &a135Notifier{}
	w := a135Watch(clock.NewFake(a135Now), n, false, &att)
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 1 || got[0].Type != obs.EventCredentialExpiryNotice {
		t.Fatalf("notifications off: %+v, want one normal event", got)
	}
}

// TestA196ImminentProofNoticeTellsTheProofDeadline: 감독 증거가 기한인 D-2 · D-1 critical 도 감독 증거의 기한 시각과 할 일
// `tossctl verify run` 을 말함 — 만료 문구로 새지 않게(구현 리뷰 P1-1).
func TestA196ImminentProofNoticeTellsTheProofDeadline(t *testing.T) {
	for _, left := range []time.Duration{2 * a196Day, a196Day} {
		att := a196ProofDue(a135Now, left)
		n := &a135Notifier{}
		w := a135Watch(clock.NewFake(a135Now), n, true, &att)
		w.checkOnce(t.Context())
		got := n.got()
		if len(got) != 1 || got[0].Type != obs.EventCredentialExpiring {
			t.Fatalf("left=%s: %+v, want one critical", left, got)
		}
		deadline := a135Now.Add(left).In(kst).Format("2006-01-02 15:04 KST")
		for _, want := range []string{"감독하 실주문 검증", deadline, "tossctl verify run", "다른 유효한 감독 증거가 없으면"} {
			if !strings.Contains(got[0].Title+"\n"+got[0].Body, want) {
				t.Errorf("left=%s: imminent proof alert lacks %q:\n%s\n%s", left, want, got[0].Title, got[0].Body)
			}
		}
	}
}

// TestA196ScopeLossWordingFollowsTheMissingKind: 범위 상실의 제목 · 원인 · 할 일은 빠진 endpoint 의 종류로 갈림(구현 리뷰 P1-2).
func TestA196ScopeLossWordingFollowsTheMissingKind(t *testing.T) {
	const (
		order  = "POST /api/v1/orders"
		cancel = "POST /api/v1/orders/{id}/cancel"
		read   = "GET /api/v1/holdings"
		cause  = "감독하 실주문 검증 증거 없이 쓰였다"
		redo   = "빠진 endpoint 를 덮는 능력 검증을 다시 실행해 attestation 을 재발급하라"
	)
	cases := []struct {
		name                string
		drop                []string
		title               string
		cause, verify, redo bool
	}{
		{"POST 둘", []string{order, cancel}, "주문 실행 검증이 attestation 에 없다", true, true, false},
		{"POST 하나", []string{cancel}, "주문 실행 검증이 attestation 에 없다", true, true, false},
		{"GET 하나", []string{read}, "엔진이 쓰는 endpoint 가 attestation 에 없다", false, false, true},
		{"POST+GET", []string{order, read}, "엔진이 쓰는 endpoint 가 attestation 에 없다", false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			att := a196ProofDue(a135Now, 20*a196Day)
			for _, e := range tc.drop {
				att = without(att, e)
			}
			n := &a135Notifier{}
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 {
				t.Fatalf("events=%+v", got)
			}
			e := got[0]
			if !strings.Contains(e.Title, a196Refused) || !strings.Contains(e.Title, tc.title) {
				t.Errorf("title=%q, want %q", e.Title, tc.title)
			}
			if !strings.HasPrefix(e.Body, a196Refused+".") {
				t.Errorf("body's first sentence is not the refusal: %q", e.Body)
			}
			for _, d := range tc.drop {
				if !strings.Contains(e.Body, d) {
					t.Errorf("body does not name %s: %q", d, e.Body)
				}
			}
			for _, want := range []string{"자동 손절이 멈춘다", "지금 실행 중인 엔진은 계속 돈다"} {
				if !strings.Contains(e.Body, want) {
					t.Errorf("body lacks %q: %q", want, e.Body)
				}
			}
			if strings.Contains(e.Body, cause) != tc.cause || strings.Contains(e.Body, "tossctl verify run") != tc.verify ||
				strings.Contains(e.Body, redo) != tc.redo {
				t.Errorf("cause=%v verify=%v redo=%v, want %v %v %v: %q", strings.Contains(e.Body, cause),
					strings.Contains(e.Body, "tossctl verify run"), strings.Contains(e.Body, redo), tc.cause, tc.verify, tc.redo, e.Body)
			}
		})
	}
}

// TestA196RejectionAfterAHealthyFileIsAnnouncedAgain: 거절 → 고쳐짐(healthy) → 같은 날 같은 거절 재발은 다시 알림 —
// prepare 의 「보냄」 재설정이 기한 변화로 일어남을 고정(구현 리뷰 P2-1).
func TestA196RejectionAfterAHealthyFileIsAnnouncedAgain(t *testing.T) {
	clk := clock.NewFake(a135Now)
	n := &a135Notifier{}
	good := a196ProofDue(a135Now, 20*a196Day)
	att := without(good, "POST /api/v1/orders")
	w := a135Watch(clk, n, true, &att)
	w.checkOnce(t.Context())
	clk.Advance(time.Hour)
	att = good
	w.checkOnce(t.Context())
	clk.Advance(time.Hour)
	att = without(good, "POST /api/v1/orders")
	w.checkOnce(t.Context())
	if got := n.got(); len(got) != 2 {
		t.Fatalf("scope loss → healthy → same scope loss on the same day: %d events, want 2: %+v", len(got), got)
	}
}

// TestA196NoProofDeadlineWithoutAValidityWindow: 만료 ≤ 발급인 파일은 감독 증거의 기한을 세지 않음 — 유효 기간이 0 이하면
// 「증거 + 0」 이 과거라 거짓 proof_overdue 가 나감(구현 리뷰 P2-2, design D1).
func TestA196NoProofDeadlineWithoutAValidityWindow(t *testing.T) {
	for name, shift := range map[string]time.Duration{"만료 = 발급": 0, "만료 < 발급": time.Hour} {
		t.Run(name, func(t *testing.T) {
			att := a196ProofDue(a135Now, 5*a196Day)
			att.ExpiresAt = a135Now.Add(20 * a196Day)
			att.IssuedAt = att.ExpiresAt.Add(shift)
			v := judgeAttestation(att, nil, a135Now, a135Account)
			if v.Stage != stageHealthy || v.DeadlineKind != deadlineExpiry || v.DaysLeft != 20 {
				t.Fatalf("stage=%s kind=%s days=%d, want healthy expiry D-20", v.Stage, v.DeadlineKind, v.DaysLeft)
			}
		})
	}
}
