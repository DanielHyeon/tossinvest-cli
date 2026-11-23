package engine

// a194 — 대사 차단 때문에 편입되지 못한 보유와 운영자만 풀 수 있는 차단 자체의 보고(design §4).
//
// a095 의 생산 배선 하네스(실 원장 · 실 obs.Notifier · 실 배달 실행자)를 그대로 씀. 「한 번」「재알림 창」 주장은 outbox 행 수가
// 아니라 Notify 호출 수(a194Capture)와 실제 전송 수(a098RecordingPublisher)로 잼 — outbox 는 재발송 때 같은 행을 UPDATE 해서
// 행 수 1 로는 변이를 못 가름. 픽스처는 합성 값만(종목 SYN*, 수량 4271, 가격 61234).

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

const (
	a194Qty   = "4271"
	a194Price = 61234
)

// a194Capture 는 Notify 호출을 기록하고 실 알림기로 넘김. fail 은 종류별 남은 기록 실패 주입 횟수, onNotify 는 호출 순간의 훅.
type a194Capture struct {
	inner    ExitAlerter
	events   []obs.Event
	fail     map[obs.EventType]int
	onNotify func(obs.Event)
}

func (c *a194Capture) Notify(ctx context.Context, e obs.Event) error {
	c.events = append(c.events, e)
	if c.onNotify != nil {
		c.onNotify(e)
	}
	if c.fail[e.Type] > 0 {
		c.fail[e.Type]--
		return errors.New("a194 injected record failure")
	}
	if c.inner == nil {
		return nil
	}
	return c.inner.Notify(ctx, e)
}

// of 는 종류(와 사실 칸 — 빈 값이면 아무 사실)로 거른 호출임.
func (c *a194Capture) of(t obs.EventType, fact string) []obs.Event {
	var out []obs.Event
	for _, e := range c.events {
		if e.Type != t {
			continue
		}
		if fact != "" && e.Fields["fact"] != fact {
			continue
		}
		out = append(out, e)
	}
	return out
}

// blockedReports 는 이 change 의 보유 보고(두 새 사실 값, 등급 무관)임.
func (c *a194Capture) blockedReports() []obs.Event {
	var out []obs.Event
	for _, e := range c.events {
		if f := e.Fields["fact"]; f == factBlockedOperatorOnly || f == factBlockedStale {
			out = append(out, e)
		}
	}
	return out
}

type a194Fixture struct {
	*a095Fixture
	cap *a194Capture
}

func newA194(t *testing.T, options ...a095Option) *a194Fixture {
	t.Helper()
	f := newA095(t, options...)
	c := &a194Capture{inner: f.n, fail: map[obs.EventType]int{}}
	f.opts.Alerts = c
	f.rebuild()
	return &a194Fixture{a095Fixture: f, cap: c}
}

func (f *a194Fixture) block(symbol, cause string) {
	f.t.Helper()
	if _, _, err := f.j.EnterReconcile(context.Background(), journal.EnterReconcileRequest{
		AccountRef: a095Account, Symbol: symbol, Cause: cause, Evidence: "a194 synthetic block",
	}); err != nil {
		f.t.Fatalf("EnterReconcile(%q, %s): %v", symbol, cause, err)
	}
}

func (f *a194Fixture) release(symbol, cause string) {
	f.t.Helper()
	if _, _, err := f.j.ReleaseReconcile(context.Background(), journal.ReleaseReconcileRequest{
		AccountRef: a095Account, Symbol: symbol, Cause: journal.ReconcileReleaseOperator,
		Evidence: "a194 synthetic release", ExpectCause: cause,
	}); err != nil {
		f.t.Fatalf("ReleaseReconcile(%q): %v", symbol, err)
	}
}

func (f *a194Fixture) sendsOf(t obs.EventType) int {
	f.pub.mu.Lock()
	defer f.pub.mu.Unlock()
	n := 0
	for _, s := range f.pub.sent {
		if s.Type == t {
			n++
		}
	}
	return n
}

// §4-1 · §4-12 — 운영자 전용 차단 아래 보유: 첫 주기 critical, 편입 0, 래치 없이 매 관측 Notify, 창 안에서는 전송이 늘지 않음.
func TestA194OperatorOnlyBlockedHoldingIsReportedOnTheFirstCycle(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)

	cycle := f.cycle()
	observed := f.clk.Now()
	p := f.position("SYNA01")
	if cycle.Adopted != 0 || p.Adopted() {
		t.Fatalf("adopted = %d (position adopted %v) under an operator-only block", cycle.Adopted, p.Adopted())
	}
	if cycle.Unmanaged != 0 {
		t.Errorf("cycle.Unmanaged = %d; the a194 report is not the unmanaged census", cycle.Unmanaged)
	}
	got := f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly)
	if len(got) != 1 {
		t.Fatalf("blocked_operator_only critical Notify calls = %d, want 1; events = %+v", len(got), f.cap.events)
	}
	e := got[0]
	if want := reconcileUnmanagedKey(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly, p.ID); e.Key != want {
		t.Errorf("key = %q, want %q", e.Key, want)
	}
	for _, want := range []string{kstText(observed) + " 대사에서", "편입하지 못했다", "걸려 있지 않았다",
		journal.ReconcileCauseAttributionFailed, "reconcile-resolve", "이 차단은 스스로 풀리지 않는다", "앱에서 직접 관리하라",
		"거절하며 그 사유를 말한다",
		"늦게 도착했다면"} {
		if !strings.Contains(e.Body, want) {
			t.Errorf("body lacks %q: %q", want, e.Body)
		}
	}
	if strings.Contains(e.Body, "걸려 있지 않다.") {
		t.Errorf("body speaks in the present tense, which is false if it is delivered after the block clears: %q", e.Body)
	}
	if n := f.sendsOf(obs.EventExitPositionAdoptionFailed); n != 1 {
		t.Errorf("critical sends = %d after the first cycle, want 1", n)
	}

	f.cycle()
	if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly)); n != 2 {
		t.Errorf("Notify calls after two cycles = %d, want 2 — a critical report goes through no memory latch", n)
	}
	if n := f.sendsOf(obs.EventExitPositionAdoptionFailed); n != 1 {
		t.Errorf("critical sends after two cycles = %d, want 1 — the outbox key and the remind window dedupe", n)
	}
}

// §4-2 — 스스로 풀리는 차단의 10분 유예: 9분 59초 침묵, 정확히 10분에 보고(>=), 운영자 해제 안내 없음.
func TestA194ASelfReleasingBlockIsReportedAfterTenMinutes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		advance time.Duration // 첫 판정 뒤 다음 판정까지(다음 주기의 안정화 2초 포함 전)
		want    int
	}{
		{"9m59s is still a transition", 9*time.Minute + 57*time.Second, 0},
		{"10m is no longer a transition", 9*time.Minute + 58*time.Second, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA194(t)
			f.holds("SYNA01", a194Qty, a194Price)
			f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
			f.cycle()
			if n := len(f.cap.blockedReports()); n != 0 {
				t.Fatalf("reports on the first cycle = %d; a self-releasing block is silent inside the grace", n)
			}
			f.clk.Advance(tc.advance)
			f.cycle()
			got := f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedStale)
			if len(got) != tc.want {
				t.Fatalf("blocked_stale reports = %d, want %d", len(got), tc.want)
			}
			if tc.want == 1 {
				// 스스로 풀리는 차단에는 「운영자만 풀 수 있다」고 단정하지 않고, 해제가 필요할 수 있다고만 말함(D6, 구현 리뷰 P2-5).
				if strings.Contains(got[0].Body, "이 차단은 스스로 풀리지 않는다") {
					t.Errorf("a self-releasing block must not be called operator-only: %q", got[0].Body)
				}
				for _, want := range []string{"필요할 수 있다", "reconcile-resolve", "앱에서 직접 관리하라"} {
					if !strings.Contains(got[0].Body, want) {
						t.Errorf("stale body lacks %q: %q", want, got[0].Body)
					}
				}
				if !strings.Contains(got[0].Body, "10분") || !strings.Contains(got[0].Body, journal.ReconcileCauseQuantityMismatch) {
					t.Errorf("body must say the block outlived the grace and name its cause: %q", got[0].Body)
				}
			}
		})
	}
}

// §4-2 — 빠지지 않은 주기가 시계를 지움: 해제된 뒤 다시 차단되면 새 10분.
func TestA194TheGraceClockIsClearedWhenTheHoldingIsNotExcluded(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.cycle()

	// 차단을 풀고 시세를 지워 편입을 연기시킴 — 보유는 후보로 판정되어 빠지지 않음(편입은 안 됨).
	f.release("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.prices.mu.Lock()
	delete(f.prices.last, "SYNA01")
	f.prices.mu.Unlock()
	f.clk.Advance(time.Minute)
	f.cycle()
	if f.position("SYNA01").Adopted() {
		t.Fatal("arrangement: the holding must stay unadopted")
	}

	f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.clk.Advance(10 * time.Minute)
	f.cycle()
	if n := len(f.cap.blockedReports()); n != 0 {
		t.Errorf("reports = %d right after a fresh block; the clock of the earlier block must have been cleared", n)
	}
}

// §4-3 — 판정 전에 끝난 주기(안정화 미수렴)가 끼어도 시계는 이어짐.
func TestA194ASkippedCycleDoesNotResetTheGraceClock(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.cycle()

	f.clk.Advance(5 * time.Minute)
	calls := 0
	f.holdings.onCall = func() {
		calls++
		if calls == 2 {
			f.holdings.items[0].Quantity = "4272" // 둘째 수집만 다름 — 안정화 미수렴
		}
	}
	if c := f.cycle(); c.Stable {
		t.Fatal("arrangement: the cycle must not stabilise")
	}
	f.holdings.onCall = nil
	f.holdings.items[0].Quantity = a194Qty

	f.clk.Advance(4*time.Minute + 56*time.Second)
	f.cycle() // 판정 시각 = 첫 판정 + 10분
	if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedStale)); n != 1 {
		t.Errorf("blocked_stale reports = %d, want 1 — a cycle that never judged is not an observation that the holding left", n)
	}
}

// §4-4 · §4-6 — 편입 대상이 아닌 보유(꺼짐∧미지정 · 제외)는 대상 아님, include 지정이면 대상.
func TestA194OnlyAdoptionCandidatesAreReported(t *testing.T) {
	for _, tc := range []struct {
		name     string
		adoption config.Adoption
		want     int
	}{
		{"adoption off and not designated", config.Adoption{}, 0},
		{"excluded", config.Adoption{Enabled: true, DefaultStopPct: 0.05, ExcludeSymbols: []string{"SYNA01"}}, 0},
		{"designated with adoption off", config.Adoption{DefaultStopPct: 0.05, IncludeSymbols: []string{"SYNA01"}}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA194(t, a095Adoption(tc.adoption))
			f.holds("SYNA01", a194Qty, a194Price)
			f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
			f.cycle()
			if n := len(f.cap.blockedReports()); n != tc.want {
				t.Errorf("blocked reports = %d, want %d", n, tc.want)
			}
		})
	}
}

// §4-4a — SNAPSHOT_STALE 은 operator_only 인데 사유가 ReasonReconcileMismatch: Release 로 가르면 즉시, 사유로 가르면 유예.
func TestA194ASnapshotCauseIsOperatorOnlyByItsRelease(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseSnapshotStale)
	f.cycle()
	if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly)); n != 1 {
		t.Errorf("blocked_operator_only reports = %d, want 1 on the first cycle", n)
	}
}

// §4-5 — 알림이 꺼진 엔진: 일반 등급 한 번, 차단 보고는 로그 한 줄, critical 의 결과(outbox · 래치 · 모드) 0.
func TestA194NotificationsOffNeverRecordsACritical(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options []a095Option
	}{
		{"off with the topic kept", []a095Option{a095NotificationsOff}},
		{"off without a topic", []a095Option{a095NotificationsOff, a095NoTopic}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newA194(t, tc.options...)
			var buf bytes.Buffer
			f.opts.Log = obs.NewLogger(obs.LogOptions{Writer: &buf, JSON: true})
			f.rebuild()
			f.holds("SYNA01", a194Qty, a194Price)
			f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
			f.cycle()
			f.cycle()

			if n := len(f.cap.of(obs.EventExitPositionUnmanaged, factBlockedOperatorOnly)); n != 1 {
				t.Errorf("normal blocked_operator_only reports = %d over two cycles, want 1 (latched)", n)
			}
			if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, "")) + len(f.cap.of(obs.EventReconcilePermanent, "")); n != 0 {
				t.Errorf("critical-typed Notify calls = %d on a notifications-off engine, want 0", n)
			}
			f.assertNoCriticalConsequence(t)
			if n := strings.Count(buf.String(), `"event":"`+string(obs.EventReconcilePermanent)+`"`); n != 1 {
				t.Errorf("block report log lines = %d over two cycles, want 1:\n%s", n, buf.String())
			}
		})
	}
}

// §4-7 — 보유 없는 운영자 전용 차단: 그날 한 번, 다음 KST 날짜에 한 번 더, 풀리면 더 없음.
func TestA194AnOperatorOnlyBlockIsReportedOncePerDay(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	f.cycle()
	got := f.cap.of(obs.EventReconcilePermanent, "")
	if len(got) != 1 {
		t.Fatalf("block report Notify calls = %d over two cycles of one day, want 1", len(got))
	}
	if want := strings.Join([]string{string(obs.EventReconcilePermanent), "symbol", "SYNB01",
		journal.ReconcileCauseAttributionFailed, "2026-09-30"}, "|"); got[0].Key != want {
		t.Errorf("key = %q, want %q", got[0].Key, want)
	}
	if !strings.Contains(got[0].Body, " KST 대사에서") || !strings.Contains(got[0].Body, "reconcile-resolve") {
		t.Errorf("body must carry the observation moment and the release command: %q", got[0].Body)
	}

	f.clk.Advance(24 * time.Hour)
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Errorf("block reports after the next day = %d, want 2", n)
	}
	f.release("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(24 * time.Hour)
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Errorf("block reports after the release = %d, want still 2", n)
	}
}

// §4-7 — 계좌 전체 운영자 전용 차단은 종목마다가 아니라 한 통.
func TestA194AnAccountWideBlockIsOneReport(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.holds("SYNA02", a194Qty, a194Price)
	f.block("", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 1 {
		t.Errorf("block reports = %d, want 1 for one account-wide block", n)
	}
	if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly)); n != 4 {
		t.Errorf("holding reports = %d over two cycles, want 4 (two holdings, every observation)", n)
	}
}

// §4-8 — KST 자정과 UTC 자정을 가름.
func TestA194TheDayIsTheKoreanCalendarDay(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.cycle() // 01:00:02 UTC = 10:00 KST 09-30
	f.clk.Advance(13*time.Hour + 58*time.Minute + 56*time.Second)
	f.cycle() // 14:59:00 UTC = 23:59 KST 09-30
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 1 {
		t.Fatalf("reports by 23:59 KST = %d, want 1", n)
	}
	f.clk.Advance(time.Minute + 58*time.Second)
	f.cycle() // 15:01:00 UTC(같은 UTC 날짜) = 00:01 KST 10-01
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Fatalf("reports after KST midnight = %d, want 2", n)
	}
	f.clk.Advance(9*time.Hour + 28*time.Minute + 58*time.Second)
	f.cycle() // 00:31:00 UTC 10-01(UTC 날짜 바뀜) = 09:31 KST 10-01
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Errorf("reports after UTC midnight = %d, want still 2 — the day is the KST day", n)
	}
}

// §4-9 — 차단 보고의 기록 실패는 보낸 것으로 세지 않음: 다음 주기가 다시 시도, 성공 뒤로 그날 더 없음.
func TestA194AFailedBlockReportIsRetriedNextCycle(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.cap.fail[obs.EventReconcilePermanent] = 1
	f.cycle()
	f.cycle()
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Errorf("block report Notify calls = %d, want 2 (failed, retried, then latched)", n)
	}
	rows := 0
	for _, r := range f.rows() {
		if r.typ == string(obs.EventReconcilePermanent) {
			rows++
		}
	}
	if rows != 1 {
		t.Errorf("outbox rows for the block report = %d, want 1 after the retry", rows)
	}
}

// §4-9a — 보유 보고의 기록 실패 → 같은 사실 재관측 → 기록 성공(critical 앞 메모리 래치 금지).
func TestA194AFailedHoldingReportIsRecordedOnTheNextObservation(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.cap.fail[obs.EventExitPositionAdoptionFailed] = 1
	f.cycle()
	f.cycle()
	n := 0
	for _, r := range f.rows() {
		if r.typ == string(obs.EventExitPositionAdoptionFailed) {
			n++
		}
	}
	if n != 1 {
		t.Errorf("outbox rows for the holding report = %d after a failed then a retried record, want 1", n)
	}
}

// §4-9b — 전달 완료 뒤 재알림 창(하네스 10분)이 지나면 전송이 한 번 더.
func TestA194ADeliveredHoldingReportIsRemindedAfterTheWindow(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	f.cycle()
	if n := f.sendsOf(obs.EventExitPositionAdoptionFailed); n != 1 {
		t.Fatalf("sends inside the window = %d, want 1", n)
	}
	f.clk.Advance(11 * time.Minute)
	f.cycle()
	if n := f.sendsOf(obs.EventExitPositionAdoptionFailed); n != 2 {
		t.Errorf("sends after the window = %d, want 2", n)
	}
}

// §4-10 — 스스로 풀리는 차단만 활성이면 차단 보고 0.
func TestA194ASelfReleasingBlockIsNotABlockReport(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseQuantityMismatch)
	f.cycle()
	f.clk.Advance(time.Hour)
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 0 {
		t.Errorf("block reports = %d for a self-releasing block, want 0", n)
	}
}

// §4-11 — blocked() 는 참인데 Blocks() 에 덮는 블록이 없는 순간: 조용히 건너뛰지 않고 유예 뒤 원인 「미상」으로.
func TestA194ABlockTheTrackerNoLongerListsTakesTheGracePath(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.prices.mu.Lock()
	delete(f.prices.last, "SYNA01") // 편입 연기 — 포지션 인스턴스만 만듦
	f.prices.mu.Unlock()
	f.cycle()
	p := f.position("SYNA01")
	held := []blockedHolding{{position: p, market: "kr", symbol: "SYNA01"}}
	start := f.clk.Now()

	f.d.reportBlockedWith(context.Background(), held, nil, start)
	if n := len(f.cap.blockedReports()); n != 0 {
		t.Fatalf("reports inside the grace = %d, want 0", n)
	}
	f.d.reportBlockedWith(context.Background(), held, nil, start.Add(blockedHoldingGrace))
	got := f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedStale)
	if len(got) != 1 {
		t.Fatalf("blocked_stale reports = %d, want 1 — a disagreement between the two reads must not be a silent skip", len(got))
	}
	if !strings.Contains(got[0].Body, "미상") {
		t.Errorf("body must say the cause is unknown: %q", got[0].Body)
	}
}

// §4-12 — 늦게 배달되는 행: 전송이 막혀 PENDING 인 동안 편입이 성공해도, 배달되는 본문은 관측 시각의 과거형.
func TestA194ALateDeliveredReportIsStillTrue(t *testing.T) {
	f := newA194(t)
	f.pub.fail = fmt.Errorf("a194 transport down")
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	moment := kstText(f.clk.Now())

	f.release("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	if c := f.cycle(); c.Adopted != 1 {
		t.Fatalf("adopted = %d on the cycle after the release", c.Adopted)
	}
	f.pub.fail = nil
	if err := f.del.cycle(context.Background()); err != nil {
		t.Fatalf("deliverer cycle: %v", err)
	}
	f.pub.mu.Lock()
	var body, blockBody string
	for _, s := range f.pub.sent {
		switch s.Type {
		case obs.EventExitPositionAdoptionFailed:
			body = s.Body
		case obs.EventReconcilePermanent:
			blockBody = s.Body
		}
	}
	f.pub.mu.Unlock()
	if !strings.Contains(body, moment) || !strings.Contains(body, "걸려 있지 않았다") {
		t.Errorf("late-delivered body = %q; it must speak of the moment %s in the past tense", body, moment)
	}
	// 차단 보고도 늦게 배달되면 관측 시각의 과거형이어야 함(구현 리뷰 P2-2).
	for _, want := range []string{moment, "활성이었다", "막혀 있었고", "걸려 있지 않았다", "늦게 도착했다면"} {
		if !strings.Contains(blockBody, want) {
			t.Errorf("late-delivered block report lacks %q: %q", want, blockBody)
		}
	}
	for _, present := range []string{"막혀 있고", "걸리지 않는다.", "로만 풀린다"} {
		if strings.Contains(blockBody, present) {
			t.Errorf("late-delivered block report speaks in the present tense (%q): %q", present, blockBody)
		}
	}
}

// §4-13 — 제목·본문·필드에 계좌 · 수량 · 가격 문자열 0.
func TestA194TheReportsCarryNoAccountQuantityOrPrice(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	reports := append(f.cap.blockedReports(), f.cap.of(obs.EventReconcilePermanent, "")...)
	if len(reports) != 2 {
		t.Fatalf("reports = %d, want the holding report and the block report", len(reports))
	}
	for _, e := range reports {
		text := e.Title + "\n" + e.Body + "\n" + fmt.Sprint(e.Fields)
		for _, forbidden := range []string{a095Account, a194Qty, fmt.Sprint(a194Price)} {
			if strings.Contains(text, forbidden) {
				t.Errorf("%s carries %q: %s", e.Type, forbidden, text)
			}
		}
	}
}

// §4-13a — 한 주기의 보유 보고 Notify 는 최대 3, 넘는 보유는 다음 주기(처음 빠진 시각 → 포지션 id 순).
func TestA194AtMostThreeHoldingReportsPerCycle(t *testing.T) {
	f := newA194(t)
	symbols := []string{"SYNC01", "SYNC02", "SYNC03", "SYNC04", "SYNC05"}
	for _, s := range symbols {
		f.holds(s, a194Qty, a194Price)
		f.block(s, journal.ReconcileCauseAttributionFailed)
	}
	f.cycle()
	var ids []string
	for _, s := range symbols {
		ids = append(ids, f.position(s).ID)
	}
	sort.Strings(ids)
	positions := func(events []obs.Event) []string {
		var out []string
		for _, e := range events {
			out = append(out, fmt.Sprint(e.Fields["position_id"]))
		}
		return out
	}
	first := positions(f.cap.blockedReports())
	if strings.Join(first, ",") != strings.Join(ids[:3], ",") {
		t.Fatalf("first cycle reported %v, want %v (at most three, by position id)", first, ids[:3])
	}
	f.cycle()
	second := positions(f.cap.blockedReports())[3:]
	if strings.Join(second, ",") != strings.Join(ids[3:], ",") {
		t.Errorf("second cycle reported %v, want the remaining %v", second, ids[3:])
	}
}

// §4-13b — 새 보고는 adopt 뒤: 보고가 전송되는 순간 같은 주기의 다른 후보는 이미 편입돼 있음.
func TestA194TheReportComesAfterTheCyclesAdoptions(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.holds("SYND01", a194Qty, a194Price)
	adoptedAtReport, sawReport := false, false
	f.cap.onNotify = func(e obs.Event) {
		if e.Fields["fact"] == factBlockedOperatorOnly && !sawReport {
			sawReport = true
			p, err := f.j.CurrentPosition(context.Background(), a095Account, "kr", "SYND01")
			adoptedAtReport = err == nil && p.Adopted()
		}
	}
	if c := f.cycle(); c.Adopted != 1 {
		t.Fatalf("adopted = %d, want the unblocked candidate", c.Adopted)
	}
	if !sawReport || !adoptedAtReport {
		t.Errorf("report seen %v, the other candidate adopted at that moment %v — the report must come after adopt", sawReport, adoptedAtReport)
	}
}

// §4-13a — 순서는 처음 빠진 시각이 먼저, 같을 때만 포지션 id: 먼저 막힌 보유가 id 가 더 커도 앞에 섬.
func TestA194EarlierExcludedHoldingsAreReportedFirst(t *testing.T) {
	f := newA194(t)
	f.holds("SYNE01", a194Qty, a194Price)
	f.holds("SYNE02", a194Qty, a194Price)
	f.prices.mu.Lock()
	delete(f.prices.last, "SYNE01")
	delete(f.prices.last, "SYNE02")
	f.prices.mu.Unlock()
	f.cycle() // 두 포지션 인스턴스만 만듦(편입 연기)
	a, b := f.position("SYNE01"), f.position("SYNE02")
	early, late := a, b // early 는 id 가 더 큰 쪽 — 시각 순서와 id 순서가 갈리게
	if a.ID < b.ID {
		early, late = b, a
	}
	f.block(early.Symbol, journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	f.cycle()
	f.block(late.Symbol, journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	f.cycle()
	var order []string
	for _, e := range f.cap.blockedReports() {
		order = append(order, fmt.Sprint(e.Fields["position_id"]))
	}
	want := []string{early.ID, early.ID, late.ID} // 둘째 주기 early, 셋째 주기 early → late
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("report order = %v, want %v (earlier exclusion first)", order, want)
	}
}

// 구현 리뷰 P2-4 — 같은 날 풀렸다 다시 걸린 차단은 다른 차단이라 다시 알림. 풀린 차단의 래치 key 는 지워짐(생존 변이 M23).
func TestA194ABlockRaisedAgainTheSameDayIsReportedAgain(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	f.release("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 1 {
		t.Fatalf("block reports after the release = %d, want 1", n)
	}
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	f.cycle()
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 2 {
		t.Errorf("block reports after the same-day re-block = %d, want 2 — a block raised again is another block", n)
	}
}

// 구현 리뷰 M3 — 덮는 운영자 전용 블록이 둘이면 가장 이른 Since 의 원인·시각이 본문에 실림(목록 순서와 무관).
func TestA194TheEarliestCoveringOperatorBlockIsNamed(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed) // 먼저 — Blocks() 정렬에서는 계좌 행 뒤에 섬
	earliest := f.clk.Now()
	f.clk.Advance(time.Minute)
	f.block("", journal.ReconcileCauseIdentifierConflict)
	f.cycle()
	got := f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedOperatorOnly)
	if len(got) != 1 {
		t.Fatalf("holding reports = %d, want 1", len(got))
	}
	want := "원인 " + journal.ReconcileCauseAttributionFailed + ", " + kstText(earliest) + "부터"
	if !strings.Contains(got[0].Body, want) {
		t.Errorf("body = %q, want the earliest covering block %q", got[0].Body, want)
	}
}

// 구현 리뷰 M12 — 알림기가 없으면 보낸 것으로 세지 않음: 다음 주기에 알림기가 생기면 그날 보고가 나감.
func TestA194NoAlerterIsNotARecordedBlockReport(t *testing.T) {
	f := newA194(t)
	f.block("SYNB01", journal.ReconcileCauseAttributionFailed)
	f.d.opts.Alerts = nil
	f.cycle()
	f.d.opts.Alerts = f.cap
	f.cycle()
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 1 {
		t.Errorf("block reports once an alerter exists = %d, want 1", n)
	}
}

// 구현 리뷰 M19 — 제목의 종목 표시는 기존 label(이름(코드)).
func TestA194TheTitlesUseTheInstrumentLabel(t *testing.T) {
	f := newA194(t)
	f.opts.Names.Learn("SYNA01", "합성종목가")
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.cycle()
	label := f.opts.Names.Label("SYNA01")
	if !strings.Contains(label, "합성종목가") {
		t.Fatalf("arrangement: label = %q", label)
	}
	for _, e := range append(f.cap.blockedReports(), f.cap.of(obs.EventReconcilePermanent, "")...) {
		if !strings.HasPrefix(e.Title, label+" ") {
			t.Errorf("%s title = %q, want it to start with the label %q", e.Type, e.Title, label)
		}
	}
}

// exit-policy Scenario 「스스로 풀리는 차단이 유예 안에 풀린다」 — 보호 누락 알림 없음, 편입 이벤트 통지.
func TestA194ABlockClearedInsideTheGraceAdoptsSilently(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.cycle()
	f.release("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.clk.Advance(5 * time.Minute)
	if c := f.cycle(); c.Adopted != 1 {
		t.Fatalf("adopted = %d after the block cleared", c.Adopted)
	}
	if n := len(f.cap.blockedReports()); n != 0 {
		t.Errorf("protection-missing reports = %d, want 0 inside the grace", n)
	}
	if n := len(f.cap.of(obs.EventExitPositionAdopted, "")); n != 1 {
		t.Errorf("adoption events = %d, want 1", n)
	}
}

// a194FlakyHoldings 는 fail 이 참인 동안 보유 읽기를 실패시킴 — 대사 주기 자체의 실패.
type a194FlakyHoldings struct {
	inner *a095Holdings
	fail  bool
}

func (h *a194FlakyHoldings) Positions(ctx context.Context) ([]domain.Position, error) {
	if h.fail {
		return nil, errors.New("a194 synthetic holdings outage")
	}
	return h.inner.Positions(ctx)
}

func (h *a194FlakyHoldings) PositionsRaw(ctx context.Context) ([]reconcile.RawHolding, error) {
	if h.fail {
		return nil, errors.New("a194 synthetic holdings outage")
	}
	return h.inner.PositionsRaw(ctx)
}

// reconciliation 의 exit-policy Scenario 「대사 주기 실패는 연속을 끊지 않는다」 — 빠진 지 6분 뒤 두 번 실패, 10분 되는 주기에 기록.
func TestA194TwoFailedCyclesAtSixMinutesDoNotBreakTheRun(t *testing.T) {
	f := newA194(t)
	flaky := &a194FlakyHoldings{inner: f.holdings}
	f.opts.Collector.Positions = flaky
	f.rebuild()
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseQuantityMismatch)
	f.cycle()
	first := f.clk.Now()

	f.clk.Advance(6 * time.Minute)
	flaky.fail = true
	for i := 0; i < 2; i++ {
		if c := f.d.RunOnce(context.Background()); c.Err == nil {
			t.Fatalf("arrangement: cycle %d must fail", i+1)
		}
		f.clk.Advance(time.Minute)
	}
	flaky.fail = false
	f.clk.Advance(first.Add(blockedHoldingGrace).Sub(f.clk.Now()) - reconcile.DefaultStabilisationInterval)
	f.cycle() // 판정 시각 = 처음 빠진 시각 + 10분
	if !f.clk.Now().Equal(first.Add(blockedHoldingGrace)) {
		t.Fatalf("arrangement: judged at %s, want %s", f.clk.Now(), first.Add(blockedHoldingGrace))
	}
	if n := len(f.cap.of(obs.EventExitPositionAdoptionFailed, factBlockedStale)); n != 1 {
		t.Errorf("blocked_stale reports = %d, want 1 on the cycle ten minutes after the first exclusion", n)
	}
}

// exit-policy Scenario 「이미 exit 관리 중인 보유」 — 대상이 아님(차단 보고는 따로 나감).
func TestA194AManagedHoldingIsNotAProtectionMissingReport(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	if c := f.cycle(); c.Adopted != 1 {
		t.Fatalf("arrangement: adopted = %d", c.Adopted)
	}
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.clk.Advance(time.Minute)
	f.cycle()
	if n := len(f.cap.blockedReports()); n != 0 {
		t.Errorf("protection-missing reports = %d for a holding already under exit management, want 0", n)
	}
	if n := len(f.cap.of(obs.EventReconcilePermanent, "")); n != 1 {
		t.Errorf("block reports = %d, want 1 (the block itself is still reported)", n)
	}
}

// gstack review — 차례 도중 남은 보유가 빠지면(편입) 다음 주기가 빈손이 되지 않고 새 차례를 시작함.
func TestA194ARoundWhoseRestLeftStartsAgainNextCycle(t *testing.T) {
	f := newA194(t)
	symbols := []string{"SYNF01", "SYNF02", "SYNF03", "SYNF04"}
	for _, s := range symbols {
		f.holds(s, a194Qty, a194Price)
		f.block(s, journal.ReconcileCauseAttributionFailed)
	}
	f.cycle()
	if n := len(f.cap.blockedReports()); n != 3 {
		t.Fatalf("first cycle reports = %d, want 3", n)
	}
	reported := map[string]bool{}
	for _, e := range f.cap.blockedReports() {
		reported[fmt.Sprint(e.Fields["position_id"])] = true
	}
	for _, s := range symbols { // 아직 보고되지 않은 넷째를 풀어 편입시킴
		if !reported[f.position(s).ID] {
			f.release(s, journal.ReconcileCauseAttributionFailed)
		}
	}
	f.clk.Advance(time.Minute)
	if c := f.cycle(); c.Adopted != 1 {
		t.Fatalf("arrangement: adopted = %d, want the released fourth holding", c.Adopted)
	}
	if n := len(f.cap.blockedReports()); n != 6 {
		t.Errorf("reports after the second cycle = %d, want 6 — the three still blocked must be attempted again", n)
	}
}

// Manager 결정 2026-11-23 — 세 본문 모두 해제 명령이 엔진 정지를 요구하고 그동안 모든 포지션의 자동 손절이 멈춘다고 말함.
func TestA194EveryReleaseGuideSaysTheStopsPauseWhileTheEngineIsDown(t *testing.T) {
	f := newA194(t)
	f.holds("SYNA01", a194Qty, a194Price)
	f.block("SYNA01", journal.ReconcileCauseAttributionFailed)
	f.holds("SYNA02", a194Qty, a194Price)
	f.block("SYNA02", journal.ReconcileCauseQuantityMismatch)
	f.cycle()
	f.clk.Advance(blockedHoldingGrace)
	f.cycle()
	bodies := map[string]string{}
	for _, e := range f.cap.blockedReports() {
		bodies[fmt.Sprint(e.Fields["fact"])] = e.Body
	}
	if blocks := f.cap.of(obs.EventReconcilePermanent, ""); len(blocks) > 0 {
		bodies["block report"] = blocks[0].Body
	}
	for _, name := range []string{factBlockedOperatorOnly, factBlockedStale, "block report"} {
		body, ok := bodies[name]
		if !ok {
			t.Fatalf("no %s body; reports = %+v", name, f.cap.events)
		}
		for _, want := range []string{"엔진을 멈춰야", "자동 손절이 돌지 않"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s body lacks %q: %q", name, want, body)
			}
		}
	}
	if !strings.Contains(bodies["block report"], "거절하며 그 사유를 말한다") ||
		strings.Contains(bodies["block report"], "로만 풀 수 있는 원인이었다") {
		t.Errorf("block report guide = %q; it must defer to the command's own refusal and not assert the cause", bodies["block report"])
	}
}
