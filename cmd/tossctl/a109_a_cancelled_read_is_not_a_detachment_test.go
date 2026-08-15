//go:build unix

package main

// a109_a_cancelled_read_is_not_a_detachment_test.go — a109 §2-fix F1·F2 (A2 P2-1·P2-2).
//
// 재부착 wrapper 의 `observe` 는 「방금 읽기가 실패했다」를 **엔진 판정**으로 쓴다:
// 탈착을 보고하고, 재-dial 시도를 깨우고, 그 시도가 성공하면 reader 를 갈아끼운다.
// 그 해석이 틀리는 경우가 둘 있고, 둘 다 요청 경로에서 **정상적으로** 일어난다.
//
//	F1 요청자가 취소했다. REST 경로는 `request.Context()` 를 그대로 넘기므로
//	   (`internal/httpapi/router.go` 의 `read`), 브라우저 탭을 닫은 요청 하나가
//	   `context.Canceled` 를 만든다. 엔진은 멀쩡한데 wrapper 는 「죽었다」로 읽고,
//	   건강한 live client 가 재-dial 로 교체된다.
//	F2 늦게 끝난 읽기다. `Read` 는 **잠금 밖에서** reader 를 부르므로, 그 사이에
//	   백그라운드 시도가 성공해 새 reader 가 자리에 앉을 수 있다. 그때 도착하는
//	   옛 실패가 방금 성립한 부착을 탈착으로 뒤엎는다.
//
// 둘 다 「실패했다」는 사실 자체는 참이다. 거짓인 것은 **그 실패가 지금 자리에 앉아
// 있는 endpoint 의 상태를 말한다**는 해석이다.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a109ContextReader 는 **요청 ctx 를 존중하는** reader 다. 실제 client 가 그렇다:
// 취소된 요청의 transport 는 `context.Canceled` 를 감싼 오류를 돌려준다.
type a109ContextReader struct {
	mu    sync.Mutex
	reads int
}

func (r *a109ContextReader) Read(ctx context.Context) (strategyprojection.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
	if err := ctx.Err(); err != nil {
		return strategyprojection.Snapshot{},
			fmt.Errorf("strategy projection runtime: %w", err)
	}
	return strategyprojection.DormantSnapshot(time.Unix(0, 0).UTC()), nil
}

func (r *a109ContextReader) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reads
}

// a109BlockingReader 는 테스트가 놓아 줄 때까지 **잠금 밖에서** 읽는 중인 reader 다.
// 그 창이 F2 가 재는 인터리빙이다.
type a109BlockingReader struct {
	entered chan struct{}
	release chan struct{}
	err     error
	once    sync.Once
}

func (r *a109BlockingReader) Read(context.Context) (strategyprojection.Snapshot, error) {
	r.once.Do(func() { close(r.entered) })
	<-r.release
	return strategyprojection.Snapshot{}, r.err
}

// TestACancelledRequestDoesNotDetachAHealthyClient 는 F1 이다.
//
// 취소는 endpoint 의 상태가 아니라 **요청자의 상태**다. 그것을 탈착으로 읽으면 화면
// 폴링이 잦은 배포에서 건강한 client 가 요청 취소 하나마다 교체 후보가 된다.
func TestACancelledRequestDoesNotDetachAHealthyClient(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	log := &a109SyncWriter{}
	replacement := &a109ContextReader{}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, log,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return replacement, true
		})
	live := &a109ContextReader{}
	attachment.attach(live, true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := attachment.Read(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("취소된 요청의 오류 = %v, want context.Canceled 계열", err)
	}

	// 시도는 백그라운드이므로 즉시 세면 놓친다.
	time.Sleep(100 * time.Millisecond)
	if got := attempts.Load(); got != 0 {
		t.Errorf("요청 취소가 재-dial 시도를 %d번 만들었다, want 0 — 엔진은 멀쩡한데 "+
			"브라우저 탭 하나를 닫은 것이 200ms probe 를 부른다", got)
	}
	if strings.Contains(log.String(), "읽기가 실패했다") {
		t.Errorf("요청 취소를 탈착으로 보고했다 — 운영자는 endpoint 가 죽은 줄 안다:\n%s",
			log.String())
	}

	// 건강한 client 가 그대로 자리에 있는가. 교체됐다면 다음 읽기는 replacement 로 간다.
	if _, err := attachment.Read(context.Background()); err != nil {
		t.Fatalf("취소 뒤의 정상 요청이 실패했다: %v", err)
	}
	if live.count() != 2 {
		t.Errorf("live client 의 읽기 수 = %d, want 2 — 요청 취소가 건강한 client 를 갈아치웠다",
			live.count())
	}
	if got := replacement.count(); got != 0 {
		t.Errorf("교체된 reader 가 %d번 읽혔다, want 0", got)
	}
}

// TestALateReadFailureDoesNotUnseatTheNewAttachment 는 F2 다.
//
// `Read` 가 잠금 밖에서 옛 reader 를 부르는 동안 시도가 성공하면, 뒤늦게 도착한 옛
// 실패는 **이미 없는 endpoint** 의 소식이다. 그것으로 새 부착을 뒤엎으면 회복이
// 성립한 직후에 다시 탈착으로 떨어지고, 다음 시도까지 화면이 비어 있는다.
func TestALateReadFailureDoesNotUnseatTheNewAttachment(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	log := &a109SyncWriter{}
	fresh := &a109FakeReader{}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, log,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return fresh, true
		})
	stale := &a109BlockingReader{
		entered: make(chan struct{}), release: make(chan struct{}),
		err: errors.New("strategy projection runtime: socket has no listener"),
	}
	attachment.attach(stale, true)

	read := make(chan error, 1)
	go func() {
		_, err := attachment.Read(context.Background())
		read <- err
	}()
	<-stale.entered // 옛 reader 가 잠금 밖에서 읽는 중이다

	// 그 창 안에서 시도가 성공해 새 reader 가 자리에 앉는다.
	attachment.wake()
	a109WaitFor(t, "재부착", func() bool { return attempts.Load() == 1 && !attachment.inFlight() })

	close(stale.release) // 옛 읽기가 이제야 실패로 끝난다
	if err := <-read; err == nil {
		t.Fatal("죽은 reader 가 성공을 보고했다 — 대조군이 이미 틀렸다")
	}

	if strings.Contains(log.String(), "읽기가 실패했다") {
		t.Errorf("사라진 endpoint 의 늦은 실패가 **새 부착**을 탈착으로 보고했다:\n%s",
			log.String())
	}
	// 창을 지나 다시 묻는다. 상태가 뒤집혔다면 여기서 두 번째 시도가 뜬다.
	clock.advance(2 * time.Minute)
	attachment.StrategyRuntimeConfigured()
	time.Sleep(100 * time.Millisecond)
	if got := attempts.Load(); got != 1 {
		t.Errorf("재-dial 시도 %d회, want 1 — 늦은 실패가 새 부착을 failed 로 뒤집었고 "+
			"방금 붙은 endpoint 를 다시 dial 한다", got)
	}
}
