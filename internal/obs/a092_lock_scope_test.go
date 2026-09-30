package obs_test

// a092 착지 단위 ③ — 잠금 범위(21.4 GREEN)와 모든 발송자의 원칙 E(22.3 C2 · C27, 23.3 K2 · K4, 24.3 M5).
//
// 정본 델타: 「exit 관측 goroutine이 기다리는 잠금은 원격 전송을 덮어서는 안 된다」 — 그 잠금을 쥐는 모든 보유자(범위 밖 동기 발송 포함)는
// 전송 동안 잠금을 놓아야 함. 그러면 운영자 승인(셈~해제)이 전송 중에 끼어들 수 있으므로 래치 세 자리는 근거 확정 순간의 해제 세대로
// 판정해야 함(「모든 발송자」 문단).

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// parkingPublisher 는 첫 Publish 를 release 가 닫힐 때까지 세움(원격 전송이 멈춘 발송자). 뒤의 호출은 곧바로 성공.
type parkingPublisher struct {
	mu      sync.Mutex
	calls   int
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func newParkingPublisher() *parkingPublisher {
	return &parkingPublisher{entered: make(chan struct{}), release: make(chan struct{})}
}

func (p *parkingPublisher) Publish(ctx context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	first := p.calls == 1
	p.mu.Unlock()
	if first {
		p.once.Do(func() { close(p.entered) })
		select {
		case <-p.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// a092Entered 는 첫 발송이 전송에 들어갈 때까지 기다림 — 들어가지 않으면(발송이 없으면) 이름 있는 실패로 멈춤.
func a092Entered(t *testing.T, pub *parkingPublisher) {
	t.Helper()
	select {
	case <-pub.entered:
	case <-time.After(10 * time.Second):
		t.Fatal("no publish began within 10s — there was no send in flight to measure against")
	}
}

func a092OtherEvent() obs.Event {
	e := a096Event()
	e.Key = string(obs.EventExitProposalRefused) + "|pos-2|LADDER_PARTIAL|1"
	return e
}

// 21.3 (f) · 21.4: 범위 밖 호출자의 동기 발송이 원격 전송 중일 때 exit 기록(기록 전용 입구)은 그 전송을 기다리지 않음.
func TestA092ARecordDoesNotWaitForAnotherSendersTransport(t *testing.T) {
	pub := newParkingPublisher()
	n, j, _, _ := a096Notifier(t, pub)
	done := make(chan error, 1)
	go func() { done <- n.Notify(context.Background(), a096Event()) }()
	defer func() { close(pub.release); <-done }()
	a092Entered(t, pub)

	if err := a092Within(t, 3*time.Second, func() error {
		return obs.RecordOnly{N: n}.Notify(context.Background(), a092OtherEvent())
	}); err != nil {
		t.Fatalf("record: %v", err)
	}
	rows, _ := j.PendingAlerts(context.Background(), 0)
	if len(rows) != 2 {
		t.Errorf("pending rows = %d, want 2 (the send in flight and the record)", len(rows))
	}
}

// 운영자 승인도 남의 전송을 기다리지 않고, 전송 중인 행을 승인으로 선점하면 그 발송자는 덮지 않고 선점을 기록하며 래치하지 않음.
func TestA092AnAcknowledgementPreemptsASendInFlight(t *testing.T) {
	pub := newParkingPublisher()
	n, j, buf, _ := a099LoggedSender(t, pub, 3)
	n.Publisher = pub
	gate := n.Gate
	done := make(chan error, 1)
	go func() { done <- n.Notify(context.Background(), a099Event()) }()
	released := false
	defer func() {
		if !released {
			close(pub.release)
		}
	}()
	a092Entered(t, pub)

	if err := a092Within(t, 3*time.Second, func() error {
		return n.Acknowledge(context.Background(), "operator-1")
	}); err != nil {
		t.Fatalf("Acknowledge: %v", err)
	}
	released = true
	close(pub.release)
	if err := <-done; err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if rej := gate.CheckEntry(); rej != nil {
		t.Errorf("entry blocked after an acknowledged send in flight: %v", rej)
	}
	if count, _ := j.UndeliveredCount(context.Background()); count != 0 {
		t.Errorf("undelivered = %d, want 0 — the acknowledgement must not be undone", count)
	}
	if lineFor(logLines(t, buf), obs.EventAlertClaimLost) == nil {
		t.Errorf("the preemption was not recorded; log:\n%s", buf.String())
	}
}

// --- 원칙 E: 세 래치 자리 -------------------------------------------------------------------------------

// deletesAndSucceeds 는 발행 순간 행을 지우고 성공을 돌려줌 → MarkAlertDelivered 가 행 없음 → :484(발행됐으나 정산 불가).
type deletesAndSucceeds struct {
	t    *testing.T
	path string
}

func (p *deletesAndSucceeds) Publish(context.Context, obs.Notification) error {
	db, err := sql.Open("sqlite", p.path)
	if err != nil {
		p.t.Fatalf("arranging the disappearance: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`DELETE FROM alert_outbox`); err != nil {
		p.t.Fatalf("arranging the disappearance: %v", err)
	}
	return nil
}

// clearsDuringTransport 는 전송 중(근거 확정 **앞**)에 해제하고 실패함.
type clearsDuringTransport struct{ gate *execgw.EntryGate }

func (p *clearsDuringTransport) Publish(context.Context, obs.Notification) error {
	p.gate.Clear(execgw.ReasonAlertUndelivered)
	return errors.New("transport is down")
}

type a092Site struct {
	name      string
	escalates bool
	attempts  int
	publisher func(t *testing.T, j *journal.Journal) obs.Publisher
	// onStage 는 훅 단계마다 먼저 불림 — release-missing 자리는 반납 직전(`release`)에 행을 지움.
	onStage func(t *testing.T, stage string, j *journal.Journal)
}

// a092DeleteRows 는 outbox 행을 원장 밖 연결로 지움(원장이 쥐고 있던 행을 잃은 상태).
func a092DeleteRows(t *testing.T, j *journal.Journal) {
	t.Helper()
	db, err := sql.Open("sqlite", j.Path())
	if err != nil {
		t.Fatalf("arranging the disappearance: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`DELETE FROM alert_outbox`); err != nil {
		t.Fatalf("arranging the disappearance: %v", err)
	}
}

var a092Sites = []a092Site{
	{name: "unrecorded", escalates: true, attempts: 1,
		publisher: func(t *testing.T, j *journal.Journal) obs.Publisher { return &deletesAndSucceeds{t: t, path: j.Path()} }},
	{name: "vanished", escalates: false, attempts: 2,
		publisher: func(t *testing.T, j *journal.Journal) obs.Publisher { return &deletesTheRow{t: t, path: j.Path()} }},
	{name: "exhausted", escalates: true, attempts: 1,
		publisher: func(*testing.T, *journal.Journal) obs.Publisher { return &alwaysFails{} }},
	// 25라운드 codex P0: 반납이 행 없음으로 돌아오면 선점이 아니라 잠금(승격 없음).
	{name: "release-missing", escalates: false, attempts: 1,
		publisher: func(*testing.T, *journal.Journal) obs.Publisher { return &alwaysFails{} },
		onStage: func(t *testing.T, stage string, j *journal.Journal) {
			if stage == "release" {
				a092DeleteRows(t, j)
			}
		}},
}

// a092SiteRun 은 한 자리에서 발송을 실패시키고, stage 에서 act 를 부름. 반환: 게이트 · 원장 · 모드.
func a092SiteRun(t *testing.T, site a092Site, account string, stage string,
	act func(ctx context.CancelFunc, gate *execgw.EntryGate)) (*execgw.EntryGate, string, []string) {
	t.Helper()
	n, j, _, _ := a099LoggedSender(t, nil, site.attempts)
	n.Publisher = site.publisher(t, j)
	n.AccountRef = account
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var seen []string
	obs.SetDeliveryHookForTest(n, func(s string) {
		seen = append(seen, s)
		if site.onStage != nil {
			site.onStage(t, s, j)
		}
		if s == stage && act != nil {
			act(cancel, n.Gate)
		}
	})
	_ = n.Notify(ctx, a099Event())
	mode := journal.ModeNormal
	if account != "" {
		cur, err := j.CurrentOperatingMode(context.Background(), account)
		if err != nil {
			t.Fatalf("CurrentOperatingMode: %v", err)
		}
		mode = cur.Mode
	}
	return n.Gate, mode, seen
}

func a092Latched(gate *execgw.EntryGate) bool {
	rej := gate.CheckEntry()
	return rej != nil && rej.Reason == execgw.ReasonAlertUndelivered
}

// 기준: 해제가 없으면 세 자리 모두 잠그고, 승격 포함 자리는 승격함. 훅이 두 단계를 모두 지나감.
func TestA092EachLatchSiteLatchesWithoutARelease(t *testing.T) {
	for _, site := range a092Sites {
		t.Run(site.name, func(t *testing.T) {
			gate, mode, seen := a092SiteRun(t, site, "acct-a092", "", nil)
			if !a092Latched(gate) {
				t.Error("no latch")
			}
			if want := map[bool]string{true: journal.ModeEntryBlocked, false: journal.ModeNormal}[site.escalates]; mode != want {
				t.Errorf("mode = %s, want %s", mode, want)
			}
			joined := strings.Join(seen, ",")
			if !strings.Contains(joined, "evidence:"+site.name) || !strings.Contains(joined, "epoch:"+site.name) {
				t.Errorf("stages seen = %v, want evidence and epoch at %s", seen, site.name)
			}
		})
	}
}

// C2 · C27: 세대 읽기 **뒤**의 해제 → 차단을 다시 세우지 않음. 승격 포함 자리는 승격은 섬.
func TestA092AReleaseAfterTheEpochReadIsHonoured(t *testing.T) {
	for _, site := range a092Sites {
		t.Run(site.name, func(t *testing.T) {
			gate, mode, _ := a092SiteRun(t, site, "acct-a092", "epoch:"+site.name,
				func(_ context.CancelFunc, g *execgw.EntryGate) { g.Clear(execgw.ReasonAlertUndelivered) })
			if a092Latched(gate) {
				t.Error("the latch was re-applied over an operator release that came after the evidence")
			}
			if want := map[bool]string{true: journal.ModeEntryBlocked, false: journal.ModeNormal}[site.escalates]; mode != want {
				t.Errorf("mode = %s, want %s — a timely escalation is not undone by the release", mode, want)
			}
		})
	}
}

// K4: 근거 확정과 세대 읽기 **사이**의 해제는 「앞」으로 봄 → 다시 잠금(보수 방향).
func TestA092AReleaseBeforeTheEpochReadRelatches(t *testing.T) {
	for _, site := range a092Sites {
		t.Run(site.name, func(t *testing.T) {
			gate, _, _ := a092SiteRun(t, site, "acct-a092", "evidence:"+site.name,
				func(_ context.CancelFunc, g *execgw.EntryGate) { g.Clear(execgw.ReasonAlertUndelivered) })
			if !a092Latched(gate) {
				t.Error("a release before the epoch read was treated as after — the latch was lost")
			}
		})
	}
}

// C2: 근거 확정 **앞**(전송 중)의 해제는 판정을 바꾸지 않음 — 그 뒤의 실패는 잠금.
func TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict(t *testing.T) {
	n, _, _, _ := a099LoggedSender(t, nil, 1)
	n.Publisher = &clearsDuringTransport{gate: n.Gate}
	_ = n.Notify(context.Background(), a099Event())
	if !a092Latched(n.Gate) {
		t.Error("a release during the transport suppressed the failure that followed it")
	}
}

// K2 · M5: 승격 포함 판정에서 해제로 조건부 차단이 생략되고 승격 쓰기가 실패하면 무조건 잠금. 승격 미포함(계정 없음)은 추가 차단 없음.
func TestA092AFailedEscalationLatchesUnconditionally(t *testing.T) {
	for _, site := range a092Sites {
		if !site.escalates {
			continue
		}
		t.Run(site.name, func(t *testing.T) {
			gate, mode, _ := a092SiteRun(t, site, "acct-a092", "epoch:"+site.name,
				func(cancel context.CancelFunc, g *execgw.EntryGate) {
					g.Clear(execgw.ReasonAlertUndelivered)
					cancel() // 승격 쓰기를 실패시킴
				})
			if mode != journal.ModeNormal {
				t.Fatalf("mode = %s — the escalation was supposed to fail", mode)
			}
			if !a092Latched(gate) {
				t.Error("both the conditional latch and the escalation were lost")
			}
		})
		t.Run(site.name+"/no-account", func(t *testing.T) {
			gate, _, _ := a092SiteRun(t, site, "", "epoch:"+site.name,
				func(cancel context.CancelFunc, g *execgw.EntryGate) {
					g.Clear(execgw.ReasonAlertUndelivered)
					cancel()
				})
			if a092Latched(gate) {
				t.Error("a judgement without an escalation latched unconditionally")
			}
		})
	}
}

// gstack 리뷰(게이트 준비, 불변식 8 (a)): 승격 실패의 무조건 래치 설명에 원문 오류(원장 문구 「… transition for <계좌>」)가
// 실리지 않음 — 게이트 설명은 상태 출력이 어디서나 읽는 칸.
func TestA092AFailedEscalationLatchKeepsTheAccountOut(t *testing.T) {
	for _, site := range a092Sites {
		if !site.escalates {
			continue
		}
		t.Run(site.name, func(t *testing.T) {
			gate, _, _ := a092SiteRun(t, site, "acct-a092", "epoch:"+site.name,
				func(cancel context.CancelFunc, g *execgw.EntryGate) {
					g.Clear(execgw.ReasonAlertUndelivered)
					cancel() // 승격 쓰기를 실패시킴
				})
			detail, ok := gate.Blocks()[execgw.ReasonAlertUndelivered]
			if !ok {
				t.Fatal("arrangement: no unconditional latch")
			}
			if strings.Contains(detail, "acct-a092") {
				t.Errorf("the gate description carries the account: %q", detail)
			}
		})
	}
}
