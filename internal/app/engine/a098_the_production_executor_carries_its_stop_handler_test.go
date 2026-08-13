package engine_test

// a098 §6.4 — 독립 리뷰가 찾고 뮤테이션이 실측한 **증거 구멍 셋**을 닫는다.
//
//	M1  프로덕션 실행자가 정지 처리기를 안 달아도 네 패키지가 초록이다
//	M3  배달되는 알림의 제목·본문을 비워도 아무도 안 잡는다
//	M4  한 건 승인을 전체 승인으로 바꿔도 아무도 안 잡는다
//
// 셋 다 「테스트가 있다」와 「테스트가 그것을 잰다」가 다르다는 같은 형태다.
//
// # M1 — 프로덕션 생성자가 돌려준 실행자가 정지 처리기를 달고 있는가
//
// R3·R15·R17 은 「발송자가 죽으면 진입이 잠긴다」를 이미 잰다. 그런데 그 셋은
// 테스트가 손으로 지은 `a098DeliveryAux` 의 `OnStop` 을 쓴다. 그래서 그 셋이 재는
// 것은 *처리기가 하는 일*이고, **프로덕션이 그 처리기를 다는지는 아무도 안 잰다.**
//
// §6.4 독립 리뷰(보이스 B, 발견 2)가 그 틈을 지적했고 뮤테이션 M1 이 실측했다 —
// `Context.AlertDeliverer` 에서 `OnStop` 을 지우면
// engine·execgw·cmd/tossctl·obs **네 패키지가 전부 초록이다.**
// 그 구현에서 발송자는 죽고 진입 게이트는 열려 있다. 이 파일이 그 뮤테이션의
// 사망 원인이다.
//
// ⛔ 「처리기가 nil 이 아니다」로는 부족하다. 빈 함수도 nil 이 아니고, 그 구현에서
// 결과는 M1 과 똑같다. 그래서 아래는 처리기를 **부르고 게이트를 읽는다.**

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestTheProductionDelivererCarriesItsStopHandler(t *testing.T) {
	clk := clock.NewFake(exitNow)
	j := a098OpenJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: 15 * time.Second,
	})

	// 프로덕션 조립부가 부르는 그 생성자. 여기서 `AuxiliaryExecutor` 를 손으로
	// 지으면 M1 이 바꾸는 자리를 테스트가 안 지나가고, 뮤테이션이 안 보인다.
	aux, err := (&engine.Context{Journal: j, Entry: gate}).AlertDeliverer(clk)
	if err != nil {
		t.Fatalf("AlertDeliverer: %v", err)
	}

	if got := gate.Blocks(); len(got) != 0 {
		t.Fatalf("전제가 안 섰다 — 게이트가 시작부터 잠겨 있다: %v", got)
	}
	if aux.OnStop == nil {
		t.Fatal("프로덕션 실행자에 정지 처리기가 없다 — 발송자가 죽어도 진입이 열려 있다")
	}

	// 실패 문자열에 원장 내용을 일부러 섞는다. 걸쇠 detail 이 그것을 옮기는지도
	// 같이 재기 위해서다 (안전 불변식 8).
	aux.OnStop(context.Background(), errors.New("delivery executor died on alert row 7 for 005930"))

	blocks := gate.Blocks()
	detail, latched := blocks[execgw.ReasonAlertSenderDown]
	if !latched {
		t.Fatalf("정지 뒤에도 %s 가 안 걸렸다: %v", execgw.ReasonAlertSenderDown, blocks)
	}
	if _, also := blocks[execgw.ReasonAlertUndelivered]; also {
		t.Error("밀린 알림 걸쇠까지 같이 걸렸다 — 운영자가 backlog 를 승인하면 " +
			"「아무도 안 보낸다」까지 풀린다 (사용자 결정 8-1)")
	}
	if strings.TrimSpace(detail) == "" {
		t.Error("걸쇠에 할 말이 없다 — 운영자가 무엇을 해야 하는지 읽을 곳이 없다")
	}
	for _, leak := range []string{"alert row 7", "005930", "died"} {
		if strings.Contains(detail, leak) {
			t.Errorf("걸쇠 detail 이 실패 내용을 그대로 옮겼다 (%q): %q", leak, detail)
		}
	}
}

// a098RecordedPublisher 는 나간 알림을 **내용째로** 붙잡아 둔다.
//
// 횟수만 세는 발송자로는 M3 이 안 보인다 — 제목과 본문이 빈 알림도 한 번은 나간다.
type a098RecordedPublisher struct {
	mu   sync.Mutex
	sent []obs.Notification
	got  chan struct{}
	once sync.Once
}

func (p *a098RecordedPublisher) Publish(_ context.Context, n obs.Notification) error {
	p.mu.Lock()
	p.sent = append(p.sent, n)
	p.mu.Unlock()
	p.once.Do(func() { close(p.got) })
	return nil
}

func (p *a098RecordedPublisher) first() obs.Notification {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.sent) == 0 {
		return obs.Notification{}
	}
	return p.sent[0]
}

// TestTheDeliveredNotificationCarriesTheRecordedAlert 는 M3 을 죽인다.
//
// 「배달됐다」는 원장의 상태이고, 운영자가 받는 것은 **알림의 내용**이다. 둘을
// 가르지 않는 테스트는 제목도 본문도 빈 알림을 배달로 세고, 그 구현에서 운영자는
// 아무것도 안 적힌 정지 실패 통보를 받는다.
func TestTheDeliveredNotificationCarriesTheRecordedAlert(t *testing.T) {
	clk := clock.NewFake(exitNow)
	j := a098OpenJournal(t, clk)
	gate := execgw.NewEntryGate(clk, nil)

	const (
		title = "STOP_UNSERVED: 005930"
		body  = "정지 주문이 아무것도 못 팔았다"
	)
	if _, err := j.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: "a098-m3", Type: string(obs.EventExitProposalRefused),
		Severity: string(obs.SeverityCritical), Title: title, Body: body,
	}); err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}

	pub := &a098RecordedPublisher{got: make(chan struct{})}
	aux, err := (&engine.Context{
		Journal: j, Entry: gate, Notifier: a098Notifier(j, gate, pub, clk),
	}).AlertDeliverer(clk)
	if err != nil {
		t.Fatalf("AlertDeliverer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = aux.Run(ctx) }()

	select {
	case <-pub.got:
	case <-time.After(10 * time.Second):
		t.Fatal("실행자가 아무것도 안 보냈다")
	}

	sent := pub.first()
	if sent.Title != title {
		t.Errorf("나간 알림의 제목 = %q, 원장의 제목 = %q", sent.Title, title)
	}
	if sent.Body != body {
		t.Errorf("나간 알림의 본문 = %q, 원장의 본문 = %q", sent.Body, body)
	}
	if string(sent.Type) != string(obs.EventExitProposalRefused) {
		t.Errorf("나간 알림의 종류 = %q, 원장의 종류 = %q",
			sent.Type, obs.EventExitProposalRefused)
	}
	if sent.Severity != obs.SeverityCritical {
		t.Errorf("나간 알림의 심각도 = %q, want %q", sent.Severity, obs.SeverityCritical)
	}
}

// TestAcknowledgingOneAlertLeavesTheOthers 는 M4 를 죽인다.
//
// a098 의 승인 테스트는 **전부** id 를 안 넘긴다(= 전체 승인). 그래서 `ids...` 를
// 통째로 버려도 아무도 안 잡고, 그 구현에서 한 건을 승인한 운영자가 backlog 전체를
// 승인한 것이 된다 — 걸쇠가 풀리고 원장에는 그 사람 이름이 남는다.
func TestAcknowledgingOneAlertLeavesTheOthers(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(exitNow)
	j := a098OpenJournal(t, clk)
	gate := execgw.NewEntryGate(clk, nil)

	for _, key := range []string{"a098-m4-a", "a098-m4-b"} {
		if _, err := j.EnqueueAlert(ctx, journal.Alert{
			EventKey: key, Type: string(obs.EventExitProposalRefused),
			Severity: string(obs.SeverityCritical), Title: key, Body: "밀린 행",
		}); err != nil {
			t.Fatalf("EnqueueAlert(%s): %v", key, err)
		}
	}
	before, err := j.PendingAlerts(ctx, 0)
	if err != nil || len(before) != 2 {
		t.Fatalf("PendingAlerts = %d행, err=%v — 전제가 안 섰다", len(before), err)
	}

	ops, err := (&engine.Context{
		Journal: j, Entry: gate, Notifier: a098Notifier(j, gate, nil, clk),
	}).AlertOperations()
	if err != nil {
		t.Fatalf("AlertOperations: %v", err)
	}

	result, err := ops.Acknowledge(ctx, "risk-officer", before[0].ID)
	if err != nil {
		t.Fatalf("Acknowledge(%d): %v", before[0].ID, err)
	}
	if result.RemainingUndelivered != 1 {
		t.Errorf("한 건 승인 뒤 남은 미배달 = %d, want 1", result.RemainingUndelivered)
	}

	after, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(after) != 1 || after[0].ID != before[1].ID {
		t.Fatalf("승인하지 않은 행이 안 남았다: %v (want id=%d 하나)", after, before[1].ID)
	}
}

// a098OpenJournal 은 이 파일의 세 테스트가 쓰는 원장 하나를 연다.
func a098OpenJournal(t *testing.T, clk clock.Clock) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}
