//go:build unix

package main

// a109_the_reattachment_does_not_strand_itself_test.go — a109 §2b.3 G1·G2·G3·G5.
//
// 재부착 자리(`strategyRuntimeAttachment`)는 세 상태를 감싼다. gstack 8패스가 찾은 것은
// 그 자리가 **스스로를 가두는** 네 가지 방법이다.
//
//	G1  빈 자리에서 시도가 sentinel 을 버린다 → descriptor 잔재 상태가 영구
//	    NOT_CONFIGURED 로 남는다(a108 D4-2 가 지운 접힘의 재발).
//	G2  깨우기가 `failed` 게이트 뒤에 있다 → 「live 로 보이지만 실제로 죽은」 자리는
//	    집계가 고장 난 배포에서 **영원히** 안 깨어난다(§2-fix F3 의 반쪽).
//	G3  성공한 읽기가 `attached` 를 복원하지 않는다 → 깜빡임 1회 뒤의 **진짜 사망**이
//	    조용해진다(탈착 보고가 영구 침묵).
//	G5  자리에서 밀려난 client 를 아무도 닫지 않는다 → 재부착마다 유휴 연결이 쌓인다.
//
// 넷 다 「화면이 언젠가 돌아온다」는 결과 테스트로는 보이지 않는다. 그래서 여기서는
// 자리를 직접 세우고 그 자리의 **상태 전이**를 잰다.

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
)

// a109ClosingReader 는 자리에서 밀려날 때 닫히는지 세는 reader 다.
//
// 실제 client(`strategyprojectionrpc.Client`)를 쓰지 않는 이유는 여기서 재는 것이
// **자리의 규율**이기 때문이다: 자리는 이전 값이 io.Closer 면 닫는다. client 쪽 계약은
// 자기 패키지의 테스트가 진다.
type a109ClosingReader struct {
	a109FakeReader
	closes atomic.Int64
}

func (r *a109ClosingReader) Close() error {
	r.closes.Add(1)
	return nil
}

// TestAnEmptySeatTakesTheUnavailableSentinel 은 G1 이다.
//
// 데몬이 엔진보다 먼저 뜨면 자리는 **비어 있다**(nil = 부재 = dormant 화면). 그 뒤
// 엔진이 descriptor 를 남기고 죽거나 반쪽 잔재가 생기면 해석은 「붙지 못했다」를
// 뜻하는 sentinel 을 돌려준다 — 그런데 시도는 `!live` 한 줄로 그 값을 **버렸다**.
// 자리는 계속 nil 이고, 화면은 「이 배포는 전략 화면을 안 쓴다」로 영구히 남는다.
func TestAnEmptySeatTakesTheUnavailableSentinel(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	sentinel := unavailableStrategyRuntime{
		cause: errors.New("strategy projection runtime: socket has no listener"),
	}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return sentinel, false // descriptor 는 있는데 dial 이 실패했다
		})
	attachment.attach(nil, false)
	if !httpapi.StrategyRuntimeAbsent(attachment) {
		t.Fatal("대조군: 출발 상태가 부재가 아니다 — 이 테스트가 재려는 자리가 아니다")
	}

	attachment.wake()
	a109WaitFor(t, "잔재 위의 시도", func() bool {
		return attempts.Load() >= 1 && !attachment.inFlight()
	})

	if httpapi.StrategyRuntimeAbsent(attachment) {
		t.Fatal("descriptor 잔재 위에서 자리가 그대로 **부재**다 — 화면은 " +
			"「엔진이 죽었다」 대신 「이 배포는 전략 화면을 안 쓴다」를 영구히 그린다")
	}
	if _, err := attachment.Read(context.Background()); err == nil {
		t.Fatal("붙지 못한 자리가 읽기를 성공시켰다")
	}
}

// TestThePublisherWakesASeatThatOnlyLooksAlive 는 G2 다.
//
// §2-fix F3 은 깨우기를 집계 **밖으로** 꺼냈지만, 깨우기 자체가 「자리가 시도 대상인가」
// (`failed`)에 매여 있었다. 그래서 아직 반쪽이 남는다: 부팅 때 live 로 붙은 자리는
// 엔진이 재시작해도 **아무도 Read 하지 않으면** `failed` 가 켜지지 않는다. 그리고
// 집계가 고장 난 배포에서는 전략 블록에 닿는 Read 자체가 없다(B1–B7 이 앞에서 끝난다).
// 자리는 「live」인 채로 죽어 있고, 상시 구동원은 그것을 못 깨운다.
func TestThePublisherWakesASeatThatOnlyLooksAlive(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	fresh := &a109FakeReader{}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return fresh, true // 엔진은 이미 돌아와 있다 — 깨우기만 하면 붙는다
		})
	// 부팅 때 붙었다. 그 뒤 엔진이 재시작했지만 자리는 여전히 「live」로 보인다.
	dead := &a109FakeReader{}
	dead.fail(errors.New("strategy projection runtime: socket has no listener"))
	attachment.attach(dead, true)

	reader := a108SnapshotReader(t, nil)
	reader.strategyRuntime = attachment
	// 전략과 **무관한** 조회 하나가 고장 난다 — 전략 블록은 그래서 실행되지 않는다.
	reader.adoptionDesired = func() (config.Adoption, string, error) {
		return config.Adoption{}, "", errors.New("optimization: 채택 설정을 읽을 수 없다")
	}
	snapshots, err := newHTTPAPISnapshotCache(reader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := snapshots.Refresh(context.Background()); err == nil {
		t.Fatal("집계가 성공했다 — 이 테스트가 재려는 상태가 아니다")
	}
	stream, err := httpapi.NewStream(httpapi.StreamOptions{}, snapshots.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stream.Close)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		publishHTTPAPISnapshots(ctx, stream, snapshots, attachment, 10*time.Millisecond)
	}()

	a109WaitFor(t, "live 로 보이는 죽은 자리의 재부착", func() bool { return attempts.Load() >= 1 })
	cancel()
	<-done
}

// TestARecoveredReadIsAnAttachmentAgain 은 G3 다.
//
// 보고는 **상태 전이 1회**다. 그런데 성공한 읽기가 `attached` 를 복원하지 않으면 그
// 상태는 한 번 false 로 떨어진 뒤 시도 goroutine 이 새 client 를 앉힐 때까지 돌아오지
// 않는다. 깜빡임(일시적 read 실패 → 다음 read 성공)은 새 client 를 만들지 않으므로,
// 그 뒤의 **진짜 사망**은 `announce := a.attached` 가 false 라 아무 말도 하지 않는다.
// 운영자에게는 endpoint 가 조용히 사라진다.
func TestARecoveredReadIsAnAttachmentAgain(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	log := &a109SyncWriter{}
	attachment := a109Attachment(t, clock, log,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			return nil, false // 시도는 실패한다 — 회복은 **읽기**가 만든다
		})
	live := &a109FakeReader{}
	attachment.attach(live, true)

	// ① 깜빡임: 읽기 하나가 실패한다.
	live.fail(errors.New("strategy projection runtime: read timeout"))
	if _, err := attachment.Read(context.Background()); err == nil {
		t.Fatal("죽은 읽기가 성공을 보고했다")
	}
	if got := strings.Count(log.String(), "읽기가 실패했다"); got != 1 {
		t.Fatalf("대조군: 첫 탈착 보고 %d줄, want 1\n%s", got, log.String())
	}

	// ② 회복: 다음 읽기가 성공한다. 같은 client·같은 자리이므로 시도는 관여하지 않는다.
	live.fail(nil)
	if _, err := attachment.Read(context.Background()); err != nil {
		t.Fatalf("회복된 읽기가 실패했다: %v", err)
	}
	if got := strings.Count(log.String(), "다시 붙었다"); got != 1 {
		t.Fatalf("읽기 회복 보고 %d줄, want 1 — 탈착만 말하고 회복을 말하지 않으면 "+
			"로그의 마지막 줄은 언제나 「죽었다」다\n%s", got, log.String())
	}

	// ③ 진짜 사망: 다시 실패한다. 이것은 **새 전이**이므로 또 보고돼야 한다.
	live.fail(errors.New("strategy projection runtime: socket has no listener"))
	if _, err := attachment.Read(context.Background()); err == nil {
		t.Fatal("죽은 읽기가 성공을 보고했다")
	}
	if got := strings.Count(log.String(), "읽기가 실패했다"); got != 2 {
		t.Fatalf("두 번째 탈착 보고 %d줄, want 2 — 깜빡임 한 번이 그 뒤의 모든 사망을 "+
			"침묵시킨다\n%s", got, log.String())
	}
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
}

// TestTheReplacedReaderIsClosed 는 G5 다(security CRITICAL).
//
// 재부착은 자리의 값을 갈아끼운다. 밀려난 값이 HTTP client 면 그것이 쥔 유휴 연결과
// 그 연결을 지키는 goroutine 은 아무도 놓아 주지 않는다 — 엔진이 오르내릴 때마다
// 하나씩 영구히 쌓인다.
func TestTheReplacedReaderIsClosed(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	fresh := &a109FakeReader{}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return fresh, true
		})
	evicted := &a109ClosingReader{}
	attachment.attach(evicted, true)

	attachment.wake()
	a109WaitFor(t, "재부착", func() bool {
		return attempts.Load() == 1 && !attachment.inFlight()
	})

	if got := evicted.closes.Load(); got != 1 {
		t.Fatalf("자리에서 밀려난 reader 를 %d번 닫았다, want 1 — 엔진이 하루에 열 번 "+
			"오르내리면 유휴 연결 열 벌이 데몬 수명 내내 남는다", got)
	}
}
