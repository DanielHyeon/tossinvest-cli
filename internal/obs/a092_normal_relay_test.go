package obs_test

// a092 착지 단위 ⑤ — 22.3 C8 · 23.3 K13: exit 관측 goroutine 의 일반 등급 알림은 유계 버퍼에 넣고 반환함. 별도 보조 실행자가 비우며,
// 버퍼가 차면 버리고 **무엇을 버렸는지 기록**함. 종료 배수 때 남은 것도 버림으로 기록함.

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type a092RecordingPublisher struct {
	mu   sync.Mutex
	sent []obs.Notification
}

func (p *a092RecordingPublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, n)
	return nil
}

func (p *a092RecordingPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sent)
}

func a092NormalEvent(key string) obs.Event {
	return obs.Event{Type: obs.EventExitProposalCapped, Key: key, Title: "capped"}
}

// a092LockedBuffer 는 실행자 goroutine 이 쓰는 동안 시험이 읽어도 경합이 없는 로그 버퍼임.
type a092LockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *a092LockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *a092LockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func a092RelayNotifier(pub obs.Publisher) (*obs.Notifier, *a092LockedBuffer) {
	buf := &a092LockedBuffer{}
	return &obs.Notifier{Publisher: pub,
		Log: obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.NewFake(obsNow)})}, buf
}

// 일반 등급은 publisher 를 부르지 않고 곧바로 반환 — 원격 전송이 멈춰 있어도(주문 제출 앞의 capped 알림, 25라운드 A#1).
func TestA092ANormalAlertIsHandedOffNotSent(t *testing.T) {
	if obs.SeverityOf(obs.EventExitProposalCapped) == obs.SeverityCritical {
		t.Fatal("fixture: the capped event must be normal-grade")
	}
	pub := &stuckPublisher{}
	n, _ := a092RelayNotifier(pub)
	relay := obs.NewNormalRelay(n, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := a092Within(t, 2*time.Second, func() error {
		return obs.RecordOnly{N: n, Relay: relay}.Notify(ctx, a092NormalEvent("k1"))
	}); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0 — the exit goroutine must not publish", pub.count())
	}
}

// 보조 실행자가 비움 — 넘겨진 알림이 실제로 나감.
func TestA092TheRelayPublishesWhatWasHandedOff(t *testing.T) {
	pub := &a092RecordingPublisher{}
	n, _ := a092RelayNotifier(pub)
	relay := obs.NewNormalRelay(n, 4)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx) }()
	_ = obs.RecordOnly{N: n, Relay: relay}.Notify(ctx, a092NormalEvent("k1"))
	deadline := time.Now().Add(5 * time.Second)
	for pub.count() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	if err := <-done; err != context.Canceled {
		t.Errorf("Run returned %v, want context.Canceled — a cancellation must be reported as one", err)
	}
	if pub.count() != 1 {
		t.Errorf("published %d, want 1", pub.count())
	}
}

// 버퍼가 차면 버리고 기록 — 넘김은 막히지 않음.
func TestA092AFullRelayDropsAndSaysWhat(t *testing.T) {
	n, buf := a092RelayNotifier(&stuckPublisher{})
	relay := obs.NewNormalRelay(n, 1)
	ctx := context.Background()
	r := obs.RecordOnly{N: n, Relay: relay}
	if err := a092Within(t, 2*time.Second, func() error {
		_ = r.Notify(ctx, a092NormalEvent("first"))
		return r.Notify(ctx, a092NormalEvent("second-dropped"))
	}); err != nil {
		t.Fatal(err)
	}
	log := buf.String()
	if !strings.Contains(log, string(obs.EventNormalAlertDropped)) || !strings.Contains(log, "second-dropped") {
		t.Errorf("the drop was not recorded with its key:\n%s", log)
	}
	if strings.Contains(log, `"first"`) && strings.Count(log, string(obs.EventNormalAlertDropped)) > 1 {
		t.Errorf("more than the overflowing alert was dropped:\n%s", log)
	}
}

// 종료 배수 — 남은 알림을 버림으로 기록하고 context 오류를 돌려줌(K13).
func TestA092TheRelayRecordsWhatShutdownLeftBehind(t *testing.T) {
	n, buf := a092RelayNotifier(&stuckPublisher{})
	relay := obs.NewNormalRelay(n, 4)
	r := obs.RecordOnly{N: n, Relay: relay}
	_ = r.Notify(context.Background(), a092NormalEvent("left-behind"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := relay.Run(ctx); err != context.Canceled {
		t.Errorf("Run = %v, want context.Canceled", err)
	}
	if !strings.Contains(buf.String(), "left-behind") || !strings.Contains(buf.String(), string(obs.EventNormalAlertDropped)) {
		t.Errorf("the shutdown did not record the undelivered alert:\n%s", buf.String())
	}
}

// 이관 수단이 없으면(Relay nil) 동기 발행으로 떨어지지 않고 버림을 기록.
func TestA092WithoutARelayANormalAlertIsDroppedNotSent(t *testing.T) {
	pub := &stuckPublisher{}
	n, buf := a092RelayNotifier(pub)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := a092Within(t, 2*time.Second, func() error {
		return obs.RecordOnly{N: n}.Notify(ctx, a092NormalEvent("no-relay"))
	}); err != nil {
		t.Fatal(err)
	}
	if pub.count() != 0 || !strings.Contains(buf.String(), "no-relay") {
		t.Errorf("publish=%d log=%s", pub.count(), buf.String())
	}
}

// 새 이벤트 타입들은 critical 이 아님 — 버림 기록이 outbox 에 들어가 진입을 막으면 안 됨.
func TestA092TheRelayEventsAreNotCritical(t *testing.T) {
	for _, e := range []obs.EventType{obs.EventNormalAlertDropped, obs.EventNormalAlertRelayStopped} {
		if obs.SeverityOf(e) == obs.SeverityCritical {
			t.Errorf("%s is critical", e)
		}
	}
}

// 26라운드 codex #2: 실행자가 멈춘 뒤 넘긴 알림도 버림으로 기록 — 아무도 안 보내는 큐에 조용히 쌓이지 않음.
func TestA092AHandOffAfterTheRelayStoppedIsRecorded(t *testing.T) {
	n, buf := a092RelayNotifier(&a092RecordingPublisher{})
	relay := obs.NewNormalRelay(n, 4)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = relay.Run(ctx)
	relay.Offer(a092NormalEvent("after-stop"))
	if !strings.Contains(buf.String(), "after-stop") {
		t.Errorf("a hand-off after the relay stopped was not recorded:\n%s", buf.String())
	}
}

// 26라운드 codex #2: 발행 중 패닉 — 그 알림과 남은 큐를 버림으로 기록하고 패닉은 위로(런타임이 정지로 기록).
type a092PanickingPublisher struct{}

func (a092PanickingPublisher) Publish(context.Context, obs.Notification) error {
	panic("transport exploded")
}

func TestA092APanicInTheRelayRecordsWhatItHeld(t *testing.T) {
	n, buf := a092RelayNotifier(a092PanickingPublisher{})
	relay := obs.NewNormalRelay(n, 4)
	relay.Offer(a092NormalEvent("in-flight"))
	relay.Offer(a092NormalEvent("queued"))
	func() {
		defer func() { _ = recover() }()
		_ = relay.Run(context.Background())
	}()
	log := buf.String()
	if !strings.Contains(log, "in-flight") || !strings.Contains(log, "queued") {
		t.Errorf("the panic lost alerts without a record:\n%s", log)
	}
	relay.Offer(a092NormalEvent("after-panic"))
	if !strings.Contains(buf.String(), "after-panic") {
		t.Errorf("a hand-off after the panic was not recorded:\n%s", buf.String())
	}
}

// 26라운드 codex #3: 전송기가 없거나 발행이 실패하면 유형 + 키로 기록.
func TestA092APublishFailureOrNoPublisherIsRecordedWithItsKey(t *testing.T) {
	for name, pub := range map[string]obs.Publisher{"no publisher": nil, "publish fails": &failingPublisher{fail: true}} {
		t.Run(name, func(t *testing.T) {
			n, buf := a092RelayNotifier(pub)
			if pub == nil {
				n.Publisher = nil
			}
			relay := obs.NewNormalRelay(n, 4)
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan error, 1)
			go func() { done <- relay.Run(ctx) }()
			relay.Offer(a092NormalEvent("lost-key"))
			deadline := time.Now().Add(5 * time.Second)
			for !strings.Contains(buf.String(), "lost-key") && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			cancel()
			<-done
			if !strings.Contains(buf.String(), "lost-key") {
				t.Errorf("the lost alert's key is not in the log:\n%s", buf.String())
			}
		})
	}
}
