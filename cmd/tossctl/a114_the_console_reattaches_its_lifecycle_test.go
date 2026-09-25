//go:build unix

package main

// a114 — 콘솔이 엔진 포지션 정책 control plane(lifecycle)에 다시 붙는다.
//
// 원형은 a109 D4 의 httpapi 재부착 테스트다(`a109_the_request_path_never_dials_test.go`·
// `a109_the_daemon_reattaches_when_the_engine_returns_test.go`). **결과**(엔진이 늦게 뜨거나
// 재시작해도 콘솔 재시작 없이 붙는다)와 **기전**(요청 경로는 dial 하지 않는다 · single-flight ·
// rate limit · 전이 1회 로그 · 취소·늦은 실패·엔진이 답한 거절은 탈착이 아니다 · 명령은 다시 보내지
// 않는다)을 둘 다 잰다. 결과만 재면 틀린 기전도 통과한다.

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// ---- 가짜 client 와 wrapper 세우기 -------------------------------------------------

// a114FakeLifecycle 은 붙은 client 하나다. err 를 채우면 모든 호출이 그 오류로 끝난다.
type a114FakeLifecycle struct {
	mu     sync.Mutex
	err    error
	calls  map[string]int
	closed bool
	states []positionpolicy.State
	noQuar bool // true 면 격리 메서드를 가진 척하지 않는다(→ 아래 a114PreA079 사용)
}

func (f *a114FakeLifecycle) record(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[name]++
	return f.err
}

func (f *a114FakeLifecycle) count(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[name]
}

func (f *a114FakeLifecycle) fail(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

func (f *a114FakeLifecycle) List(context.Context) ([]positionpolicy.State, error) {
	return f.states, f.record("List")
}
func (f *a114FakeLifecycle) Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error) {
	return positionpolicy.Preview{}, f.record("Preview")
}
func (f *a114FakeLifecycle) Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return positionpolicy.State{}, f.record("Apply")
}
func (f *a114FakeLifecycle) Quarantines(context.Context) ([]exitquarantine.Row, error) {
	return nil, f.record("Quarantines")
}
func (f *a114FakeLifecycle) PreviewQuarantineRelease(context.Context, exitquarantine.Request) (exitquarantine.Preview, error) {
	return exitquarantine.Preview{}, f.record("PreviewQuarantineRelease")
}
func (f *a114FakeLifecycle) ReleaseQuarantine(context.Context, exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	return exitquarantine.Result{}, f.record("ReleaseQuarantine")
}
func (f *a114FakeLifecycle) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

// a114PreA079 는 격리 해제를 모르는 엔진의 client 다(lifecycle 세 메서드뿐).
type a114PreA079 struct{ inner *a114FakeLifecycle }

func (p a114PreA079) List(ctx context.Context) ([]positionpolicy.State, error) {
	return p.inner.List(ctx)
}
func (p a114PreA079) Preview(ctx context.Context, r positionpolicy.Request) (positionpolicy.Preview, error) {
	return p.inner.Preview(ctx, r)
}
func (p a114PreA079) Apply(ctx context.Context, r positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return p.inner.Apply(ctx, r)
}

// a114Attachment 는 wrapper 하나를 테스트가 쥔 resolve·시계와 함께 세운다.
func a114Attachment(t *testing.T, clock *a109Clock, log io.Writer,
	resolve func(context.Context) (positionPolicyLifecycleClient, error)) *positionPolicyLifecycleAttachment {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &positionPolicyLifecycleAttachment{
		resolve: resolve, ctx: ctx, now: clock.Now, interval: time.Minute, log: log,
	}
}

func a114Clock() *a109Clock { return &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()} }

// a114EngineDir 는 엔진 control 디렉터리가 들어갈 짧은 비공개 디렉터리다.
func a114EngineDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "a114-engine-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

// a114Commands 는 엔진 쪽 lifecycle 서비스다. marker 로 어느 엔진 인스턴스가 답했는지 구별한다.
type a114Commands struct{ marker string }

func (c a114Commands) List(context.Context) ([]positionpolicy.State, error) {
	return []positionpolicy.State{{PositionID: c.marker}}, nil
}
func (a114Commands) Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error) {
	return positionpolicy.Preview{}, positionpolicy.ErrInvalidRequest
}
func (a114Commands) Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return positionpolicy.State{}, positionpolicy.ErrCapabilityInvalid
}

func a114StartEngine(t *testing.T, dir, marker string) *engine.PositionPolicyCommandServer {
	t.Helper()
	server, err := engine.StartPositionPolicyCommandServer(dir, a114Commands{marker: marker})
	if err != nil {
		t.Fatalf("엔진 lifecycle endpoint 기동: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return server
}

// a114ListedBy 는 지금 화면이 읽는 lifecycle 목록이 어느 엔진의 것인지다("" = 못 읽음).
func a114ListedBy(commander *consolePositionPolicyCommander) string {
	states, err := commander.List(context.Background())
	if err != nil || len(states) != 1 {
		return ""
	}
	return states[0].PositionID
}

// ---- 결과: 늦은 기동과 재시작 ---------------------------------------------------------

// TestTheConsoleAttachesWhenTheEngineStartsLater 는 spec 시나리오 「엔진이 콘솔보다 늦게 뜬다」다.
//
// 진짜 endpoint 로 잰다 — 「엔진이 돌아왔다」는 디스크의 descriptor 와 loopback 포트로만 존재한다.
func TestTheConsoleAttachesWhenTheEngineStartsLater(t *testing.T) {
	dir := a114EngineDir(t)
	log := &a109SyncWriter{}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	restore := consolePositionPolicyRedialInterval
	consolePositionPolicyRedialInterval = 0
	t.Cleanup(func() { consolePositionPolicyRedialInterval = restore })

	commander := consolePositionPolicyCommanderFor(ctx, dir, log)
	if commander == nil {
		t.Fatal("engineDir 가 있는데 commander 가 nil 이다 — 화면이 「미배선」으로 굳는다")
	}
	if _, err := commander.List(context.Background()); !errors.Is(err, errPositionPolicyLifecycleDetached) {
		t.Fatalf("엔진이 없을 때 List 오류 = %v, want detached", err)
	}
	if strings.TrimSpace(log.String()) != "" {
		t.Fatalf("descriptor 부재는 부팅 경고가 아니다(오늘처럼 조용히): %q", log.String())
	}

	a114StartEngine(t, dir, "first")
	a109WaitFor(t, "늦게 뜬 엔진에 콘솔 재시작 없이 붙기", func() bool {
		return a114ListedBy(commander) == "first"
	})
	if !strings.Contains(log.String(), "붙었다") {
		t.Fatalf("부착 전이를 한 줄로 알리지 않았다: %q", log.String())
	}
}

// TestTheConsoleReattachesAfterTheEngineRestarts 는 spec 시나리오 「가동 중 엔진 재시작」이다.
// 새 엔진은 새 ephemeral 포트와 새 토큰을 갖는다 — 부팅 때 잡은 client 는 영구 실패한다.
func TestTheConsoleReattachesAfterTheEngineRestarts(t *testing.T) {
	dir := a114EngineDir(t)
	log := &a109SyncWriter{}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	restore := consolePositionPolicyRedialInterval
	consolePositionPolicyRedialInterval = 0
	t.Cleanup(func() { consolePositionPolicyRedialInterval = restore })

	first := a114StartEngine(t, dir, "first")
	commander := consolePositionPolicyCommanderFor(ctx, dir, log)
	if got := a114ListedBy(commander); got != "first" {
		t.Fatalf("부팅 때 떠 있는 엔진에 즉시 붙지 않았다: %q", got)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	// 그 사이 화면은 부재를 **상태로** 받는다 — 오류이고, 지어낸 목록이 아니다.
	if _, err := commander.List(context.Background()); err == nil {
		t.Fatal("엔진이 내려갔는데 목록을 받았다")
	}
	a114StartEngine(t, dir, "second")
	a109WaitFor(t, "재시작한 엔진에 콘솔 재시작 없이 다시 붙기", func() bool {
		return a114ListedBy(commander) == "second"
	})
	text := log.String()
	if !strings.Contains(text, "실패했다") || !strings.Contains(text, "붙었다") {
		t.Fatalf("탈착·부착 전이를 한 줄씩 남기지 않았다: %q", text)
	}
}

// ---- 기전: 요청 경로 · single-flight · rate limit -----------------------------------

// TestTheLifecycleRequestPathNeverWaitsForADial 은 SHALL NOT(요청 경로 dial)의 직접 측정이다.
// `positionpolicyrpc.Dial` 은 health GET 을 품는다 — 동기로 부르면 화면이 그 연결을 기다린다.
func TestTheLifecycleRequestPathNeverWaitsForADial(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			once.Do(func() { close(entered) })
			<-release
			return nil, errors.New("아직 없다")
		})
	defer close(release)
	attachment.attach(nil)

	answered := make(chan error, 1)
	go func() {
		_, err := attachment.List(context.Background())
		answered <- err
	}()
	select {
	case err := <-answered:
		if !errors.Is(err, errPositionPolicyLifecycleDetached) {
			t.Fatalf("부착 전 List 오류 = %v, want detached", err)
		}
	case <-time.After(time.Second):
		t.Fatal("요청 경로가 dial 을 기다렸다")
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("시도가 아예 일어나지 않았다 — 비차단으로 만들면서 재부착을 지웠다")
	}
}

// TestTheLifecycleAttemptIsSingleFlight 는 rate limit 을 끄고(간격 0) 겹침만 잰다(a109 M10 교훈).
func TestTheLifecycleAttemptIsSingleFlight(t *testing.T) {
	entered, release := make(chan struct{}, 64), make(chan struct{})
	var calls atomic.Int64
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			entered <- struct{}{}
			<-release
			return nil, errors.New("아직 없다")
		})
	attachment.interval = 0
	attachment.attach(nil)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = attachment.List(context.Background())
		}()
	}
	wg.Wait()
	<-entered
	time.Sleep(100 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("동시 요청 20건이 dial 을 %d번 불렀다, want 1", got)
	}
	close(release)
	a109WaitFor(t, "첫 시도 종료", func() bool { return !attachment.inFlight() })
}

// TestTheLifecycleAttemptIsRateLimited 는 창 밖에서만 다시 시도함을 손 시계로 잰다.
func TestTheLifecycleAttemptIsRateLimited(t *testing.T) {
	var calls atomic.Int64
	clock := a114Clock()
	attachment := a114Attachment(t, clock, io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return nil, errors.New("아직 없다")
		})
	attachment.attach(nil)
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "첫 시도", func() bool { return calls.Load() == 1 && !attachment.inFlight() })
	for range 10 {
		_, _ = attachment.List(context.Background())
	}
	time.Sleep(50 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("간격 안에서 시도 %d번, want 1", got)
	}
	clock.advance(time.Minute)
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "간격 뒤 두 번째 시도", func() bool { return calls.Load() == 2 })
}

// TestTheProductionLifecycleRedialIntervalIsThirtySeconds 는 운영 값을 못 박는다(a109 M9 교훈).
func TestTheProductionLifecycleRedialIntervalIsThirtySeconds(t *testing.T) {
	if consolePositionPolicyRedialInterval != 30*time.Second {
		t.Fatalf("재부착 간격 = %s, want 30s", consolePositionPolicyRedialInterval)
	}
}

// ---- 기전: 판정 --------------------------------------------------------------------

// TestAnEngineAnswerIsNotADetachment — 엔진이 판정해 돌려준 거절은 endpoint 가 살아 있다는 증거다.
// 그것을 실패로 읽으면 운영자의 거절된 Apply 한 번이 탈착 로그와 재-dial 을 만든다.
func TestAnEngineAnswerIsNotADetachment(t *testing.T) {
	var calls atomic.Int64
	log := &a109SyncWriter{}
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	for _, answer := range positionPolicyEngineAnswers {
		live.fail(fmt.Errorf("%w: 엔진이 답했다", answer))
		_, err := attachment.Apply(context.Background(), positionpolicy.ApplyRequest{})
		if !errors.Is(err, answer) {
			t.Fatalf("엔진의 답 %v 가 운영자에게 그대로 가지 않았다: %v", answer, err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	if got := calls.Load(); got != 0 {
		t.Fatalf("엔진이 답한 거절 뒤 재-dial %d번 — 거절은 탈착이 아니다", got)
	}
	if strings.TrimSpace(log.String()) != "" {
		t.Fatalf("엔진이 답한 거절에 탈착 로그를 남겼다: %q", log.String())
	}
	// 대조군: 연결 오류(transport)는 재부착 대상이다 — 전부를 「답」으로 읽는 구현을 거른다.
	live.fail(&url.Error{Op: "Get", URL: "http://127.0.0.1:1/v1/positions", Err: syscall.ECONNREFUSED})
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "모르는 오류 뒤 재부착 시도", func() bool { return calls.Load() >= 1 })
}

// TestEveryEngineAnswerIsOnTheList 는 답 목록의 완전성이다 — 두 패키지의 exported `Err*` 전부.
func TestEveryEngineAnswerIsOnTheList(t *testing.T) {
	declared := map[string]bool{}
	for pkg, file := range map[string]string{
		"positionpolicy": "../../internal/positionpolicy/model.go",
		"exitquarantine": "../../internal/exitquarantine/model.go",
	} {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				for _, name := range spec.(*ast.ValueSpec).Names {
					if strings.HasPrefix(name.Name, "Err") {
						declared[pkg+"."+name.Name] = true
					}
				}
			}
		}
	}
	listed := map[string]bool{}
	source, err := parser.ParseFile(token.NewFileSet(), "console_lifecycle_attach.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(source, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || spec.Names[0].Name != "positionPolicyEngineAnswers" {
			return true
		}
		ast.Inspect(spec, func(inner ast.Node) bool {
			if sel, ok := inner.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok {
					listed[pkg.Name+"."+sel.Sel.Name] = true
				}
			}
			return true
		})
		return false
	})
	if len(declared) == 0 {
		t.Fatal("선언을 하나도 못 셌다 — 경로가 틀렸다")
	}
	for name := range declared {
		if !listed[name] {
			t.Errorf("%s 가 답 목록에 없다 — 그 거절이 탈착 로그와 재-dial 을 만든다", name)
		}
	}
	for name := range listed {
		if !declared[name] {
			t.Errorf("답 목록의 %s 는 선언된 오류가 아니다", name)
		}
	}
	if len(listed) != len(positionPolicyEngineAnswers) {
		t.Errorf("목록 항목 %d, 구문으로 센 것 %d", len(positionPolicyEngineAnswers), len(listed))
	}
}

// TestACancelledLifecycleCallIsNotADetachment — 브라우저 탭을 닫은 것은 endpoint 판정이 아니다(a109 M29).
func TestACancelledLifecycleCallIsNotADetachment(t *testing.T) {
	var calls atomic.Int64
	log := &a109SyncWriter{}
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	live.fail(context.Canceled)
	_, _ = attachment.List(ctx)
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 0 || strings.TrimSpace(log.String()) != "" {
		t.Fatalf("취소된 요청이 재-dial %d번·로그 %q 를 만들었다", calls.Load(), log.String())
	}
	if client, _, _ := attachment.current(); client != positionPolicyLifecycleClient(live) {
		t.Fatal("취소된 요청이 멀쩡한 client 를 교체 후보로 만들었다")
	}
}

// TestALateLifecycleFailureDoesNotUnseatTheNewAttachment — 옛 자리의 늦은 실패는 새 부착을 뒤엎지 않는다(a109 M30).
func TestALateLifecycleFailureDoesNotUnseatTheNewAttachment(t *testing.T) {
	log := &a109SyncWriter{}
	old := &a114FakeLifecycle{}
	fresh := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) { return fresh, nil })
	attachment.attach(old)
	_, seat, _ := attachment.current()
	// 시도가 새 client 를 앉힌다(옛 호출이 아직 잠금 밖에서 도는 사이).
	attachment.attempt()
	attachment.observe(context.Background(), seat, errors.New("옛 엔진: connection refused"))
	client, _, failed := attachment.current()
	if client != positionPolicyLifecycleClient(fresh) || failed {
		t.Fatalf("옛 자리의 늦은 실패가 새 부착을 뒤엎었다 (failed=%v)", failed)
	}
	if strings.Contains(log.String(), "실패했다") {
		t.Fatalf("옛 자리의 소식으로 탈착을 보고했다: %q", log.String())
	}
	if !old.closed {
		t.Fatal("밀려난 client 를 놓아 주지 않았다(io.Closer)")
	}
}

// TestAFailedLifecycleAttemptKeepsTheCurrentClient — 실패한 해석은 붙어 있던 자리를 비우지 않는다(a109 M11).
func TestAFailedLifecycleAttemptKeepsTheCurrentClient(t *testing.T) {
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) { return nil, errors.New("못 붙음") })
	attachment.attach(live)
	attachment.attempt()
	if client, _, _ := attachment.current(); client != positionPolicyLifecycleClient(live) {
		t.Fatal("실패한 시도가 현재 client 를 갈아끼웠다")
	}
}

// TestTheLifecycleAttachmentReportsOnlyTransitions — 30초마다 실패를 찍지 않는다(a109 M15).
func TestTheLifecycleAttachmentReportsOnlyTransitions(t *testing.T) {
	log := &a109SyncWriter{}
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) { return nil, errors.New("못 붙음") })
	attachment.attach(live)
	live.fail(errors.New("connection refused"))
	for range 5 {
		_, _ = attachment.List(context.Background())
	}
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	if got := strings.Count(log.String(), "실패했다"); got != 1 {
		t.Fatalf("탈착 전이 로그 %d줄, want 1: %q", got, log.String())
	}
	// 같은 client 로 회복하면 부착 전이 1줄(a109 G3) — 그 뒤의 진짜 사망도 다시 한 줄로 말해야 한다.
	live.fail(nil)
	_, _ = attachment.List(context.Background())
	live.fail(errors.New("connection refused"))
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	if got := strings.Count(log.String(), "실패했다"); got != 2 {
		t.Fatalf("회복 뒤 다시 죽었는데 탈착 로그 %d줄, want 2: %q", got, log.String())
	}
}

// ---- 명령 · 격리 해제 · 문구 · 부팅 ---------------------------------------------------

// TestTheLifecycleWrapperNeverResendsACommand — 실패한 Apply·격리 해제는 새 client 로 다시 가지 않는다.
// 엔진이 이미 적용했을 수 있는 명령을 두 번 보내면 안 된다.
func TestTheLifecycleWrapperNeverResendsACommand(t *testing.T) {
	old := &a114FakeLifecycle{}
	fresh := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) { return fresh, nil })
	attachment.interval = 0
	attachment.attach(old)
	old.fail(errors.New("connection reset by peer"))
	if _, err := attachment.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: "c"}); err == nil {
		t.Fatal("실패한 Apply 가 성공으로 보고됐다")
	}
	if _, err := attachment.ReleaseQuarantine(context.Background(), exitquarantine.ApplyRequest{Capability: "c"}); err == nil {
		t.Fatal("실패한 격리 해제가 성공으로 보고됐다")
	}
	a109WaitFor(t, "재부착", func() bool {
		client, _, _ := attachment.current()
		return client == positionPolicyLifecycleClient(fresh)
	})
	time.Sleep(50 * time.Millisecond)
	if old.count("Apply") != 1 || old.count("ReleaseQuarantine") != 1 ||
		fresh.count("Apply") != 0 || fresh.count("ReleaseQuarantine") != 0 {
		t.Fatalf("명령 전송 횟수 old=%v fresh=%v — 명령은 정확히 한 번만 간다", old.calls, fresh.calls)
	}
}

// TestTheQuarantineSurfaceRidesTheAttachment — wrapper 가 격리 해제를 가리면 버튼이 사라진다.
func TestTheQuarantineSurfaceRidesTheAttachment(t *testing.T) {
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) { return nil, errors.New("못 붙음") })
	commander := &consolePositionPolicyCommander{lifecycle: attachment}
	// 부착 전: 발견은 되지만(버튼 경로가 산다) 호출은 detached 로 즉시 끝난다 — 「배선 안 됨」이 아니다.
	attachment.attach(nil)
	if _, err := commander.Quarantines(context.Background()); !errors.Is(err, errPositionPolicyLifecycleDetached) ||
		errors.Is(err, exitquarantine.ErrUnwired) {
		t.Fatalf("부착 전 격리 목록 오류 = %v, want detached(ErrUnwired 아님 — 그 문구는 빌드 탓을 포함한다)", err)
	}
	// 부착 후: 세 메서드가 붙은 client 로 간다.
	attachment.attach(live)
	_, _ = commander.Quarantines(context.Background())
	_, _ = commander.PreviewQuarantineRelease(context.Background(), exitquarantine.Request{})
	_, _ = commander.ReleaseQuarantine(context.Background(), exitquarantine.ApplyRequest{})
	for _, name := range []string{"Quarantines", "PreviewQuarantineRelease", "ReleaseQuarantine"} {
		if live.count(name) != 1 {
			t.Fatalf("%s 가 붙은 client 로 가지 않았다", name)
		}
	}
	// a079 이전 엔진: 격리 메서드가 없는 client 는 오늘처럼 ErrUnwired(버튼 없음)다.
	attachment.attach(a114PreA079{inner: live})
	if _, err := commander.Quarantines(context.Background()); !errors.Is(err, exitquarantine.ErrUnwired) {
		t.Fatalf("격리 해제를 모르는 엔진 = %v, want ErrUnwired", err)
	}
}

// TestTheDetachedMessageDoesNotAssertTheEngineIsAbsent — spec: 부재를 상태로 표시하되 단정하지 않는다.
func TestTheDetachedMessageDoesNotAssertTheEngineIsAbsent(t *testing.T) {
	text := errPositionPolicyLifecycleDetached.Error()
	for _, must := range []string{"내려갔거나", "강등", "다시 붙는다"} {
		if !strings.Contains(text, must) {
			t.Errorf("detached 문구에 %q 가 없다: %q", must, text)
		}
	}
	for _, never := range []string{"엔진이 없다", "엔진이 실행 중이 아니다", "not running", "미배선"} {
		if strings.Contains(text, never) {
			t.Errorf("detached 문구가 원인을 단정한다(%q): %q", never, text)
		}
	}
}

// TestRunConsoleNeverDialsTheLifecycleItself 는 부팅 1회 dial 재도입의 구조 핀이다.
// runConsole 이 `positionpolicyrpc.Dial` 을 직접 부르면 그 client 는 굳는다 — dial 은 wrapper 의
// resolve(백그라운드)에만 있어야 한다.
func TestRunConsoleNeverDialsTheLifecycleItself(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "console.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var run *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "runConsole" {
			run = fn
		}
	}
	if run == nil {
		t.Fatal("runConsole 을 찾지 못했다")
	}
	dials, wires := 0, 0
	ast.Inspect(run.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "positionpolicyrpc" && fun.Sel.Name == "Dial" {
				dials++
			}
		case *ast.Ident:
			if fun.Name == "consolePositionPolicyCommanderFor" {
				wires++
			}
		}
		return true
	})
	if dials != 0 {
		t.Fatalf("runConsole 이 positionpolicyrpc.Dial 을 %d번 직접 부른다 — 부팅 1회 client 는 엔진 재시작 뒤 영구 실패한다", dials)
	}
	if wires != 1 {
		t.Fatalf("runConsole 이 consolePositionPolicyCommanderFor 를 %d번 부른다, want 1", wires)
	}
}

// ---- freeze 리뷰 P1-1 · P1-2 · P2 ---------------------------------------------------

// a114LiveSeat 은 List 는 성공하고 격리 목록만 실패하는 붙은 자리다(freeze P1-1 의 깜빡임 모양).
type a114LiveSeat struct {
	*a114FakeLifecycle
	quarantineErr error
}

func (s a114LiveSeat) List(ctx context.Context) ([]positionpolicy.State, error) {
	_ = s.a114FakeLifecycle.record("List")
	return nil, nil
}

func (s a114LiveSeat) Quarantines(context.Context) ([]exitquarantine.Row, error) {
	_ = s.a114FakeLifecycle.record("Quarantines")
	return nil, s.quarantineErr
}

// TestAnInternalEngineErrorDoesNotFlapTheSeat — freeze P1-1. 엔진의 `internal`(500)은 코드가 붙은
// **답**이다. 그것을 탈착으로 읽으면 List 성공과 번갈아 매 렌더 「붙었다/실패했다」를 찍고 client 를 갈아끼운다.
func TestAnInternalEngineErrorDoesNotFlapTheSeat(t *testing.T) {
	var calls atomic.Int64
	log := &a109SyncWriter{}
	seat := a114LiveSeat{a114FakeLifecycle: &a114FakeLifecycle{},
		quarantineErr: fmt.Errorf("%s: journal busy", exitQuarantineRemoteFailurePhrase)}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(seat)
	for range 10 { // 렌더 10번: List 성공 → 격리 목록 500
		_, _ = attachment.List(context.Background())
		_, _ = attachment.Quarantines(context.Background())
	}
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 0 || strings.TrimSpace(log.String()) != "" {
		t.Fatalf("엔진의 내부 오류 답이 재-dial %d번·로그 %q 를 만들었다", calls.Load(), log.String())
	}
}

// TestATokenRejectionIsADetachment — 재시작한 엔진이 같은 포트에 다른 토큰으로 뜨면 옛 client 가 받는
// 답이다. 코드가 붙어 있어도 이것만은 탈착이다 — 아니면 콘솔이 옛 토큰에 영구히 묶인다.
func TestATokenRejectionIsADetachment(t *testing.T) {
	var calls atomic.Int64
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	live.fail(fmt.Errorf("%s: %s", positionPolicyRemoteFailurePhrase, engineBearerRejectedPhrase))
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "토큰 거절 뒤 재부착 시도", func() bool { return calls.Load() >= 1 })
}

// TestAHungEngineTimeoutIsADetachment — client 의 5s timeout 도 `context.DeadlineExceeded` 다. 요청자의 ctx
// 가 멀쩡한데 그 오류가 오면 멈춘 엔진이지 취소가 아니다(`requestCancelled` 의 ctx 확인이 가르는 것).
func TestAHungEngineTimeoutIsADetachment(t *testing.T) {
	var calls atomic.Int64
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	live.fail(&url.Error{Op: "Get", URL: "http://127.0.0.1:1/v1/positions", Err: context.DeadlineExceeded})
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "멈춘 엔진 뒤 재부착 시도", func() bool { return calls.Load() >= 1 })
}

// TestTheAnswerPhrasesAreTheSourcesOwn 은 판정이 기대는 문구 셋이 **원본 파일에 그대로** 있는지 본다.
// 다른 패키지의 문장을 옮겨 적은 판정은 그 문장이 바뀌는 순간 조용히 틀린다.
func TestTheAnswerPhrasesAreTheSourcesOwn(t *testing.T) {
	for file, phrase := range map[string]string{
		"../../internal/positionpolicyrpc/client.go":                 positionPolicyRemoteFailurePhrase,
		"../../internal/positionpolicyrpc/exit_quarantine_client.go": exitQuarantineRemoteFailurePhrase,
		"../../internal/app/engine/position_policy_transport.go":     engineBearerRejectedPhrase,
	} {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		ast.Inspect(parsed, func(node ast.Node) bool {
			if lit, ok := node.(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == fmt.Sprintf("%q", phrase) {
				found = true
			}
			return !found
		})
		if !found {
			t.Errorf("%s 에 문자열 %q 가 없다 — 답 판정이 옛 문구에 묶여 있다", file, phrase)
		}
	}
}

// TestTheConsoleAttachesWithoutAnyRender — freeze P1-2. 화면이 하나도 안 열려 있어도(요청 0) 엔진이
// 뜨면 붙는다. 콘솔에는 httpapi 의 publisher 같은 상시 구동원이 없으므로 펌프가 그 일을 한다.
func TestTheConsoleAttachesWithoutAnyRender(t *testing.T) {
	dir := a114EngineDir(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	restore := consolePositionPolicyRedialInterval
	consolePositionPolicyRedialInterval = 20 * time.Millisecond
	t.Cleanup(func() { consolePositionPolicyRedialInterval = restore })
	commander := consolePositionPolicyCommanderFor(ctx, dir, io.Discard)
	attachment := commander.lifecycle.(*positionPolicyLifecycleAttachment)
	a114StartEngine(t, dir, "first")
	a109WaitFor(t, "렌더 없이 붙기", func() bool {
		client, _, _ := attachment.current()
		return client != nil
	})
}

// TestTheBootResolutionLeavesTheFirstWakeFree — freeze P1-2. 부팅 해석이 `lastTry` 를 찍으면 첫 wake 가
// 간격(운영 30s)만큼 막힌다. 간격을 1시간으로 두고, 엔진이 뜬 뒤 첫 요청이 즉시 붙이는지 본다.
func TestTheBootResolutionLeavesTheFirstWakeFree(t *testing.T) {
	dir := a114EngineDir(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	restore := consolePositionPolicyRedialInterval
	consolePositionPolicyRedialInterval = time.Hour
	t.Cleanup(func() { consolePositionPolicyRedialInterval = restore })
	commander := consolePositionPolicyCommanderFor(ctx, dir, io.Discard)
	a114StartEngine(t, dir, "first")
	_, _ = commander.List(context.Background()) // 첫 요청 — detached 로 답하고 시도를 깨운다
	a109WaitFor(t, "첫 wake 가 막히지 않고 붙기", func() bool { return a114ListedBy(commander) == "first" })
}

// ---- 구현 후 리뷰 (post-review) --------------------------------------------------------

// TestAnInternalListErrorDoesNotFlapTheSeat — post-review N7: position policy decoder 의 문구도 답이다.
func TestAnInternalListErrorDoesNotFlapTheSeat(t *testing.T) {
	var calls atomic.Int64
	log := &a109SyncWriter{}
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), log,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	live.fail(fmt.Errorf("%s: database is locked", positionPolicyRemoteFailurePhrase))
	for range 5 {
		_, _ = attachment.List(context.Background())
	}
	time.Sleep(50 * time.Millisecond)
	if calls.Load() != 0 || strings.TrimSpace(log.String()) != "" {
		t.Fatalf("엔진의 내부 오류 답(List)이 재-dial %d번·로그 %q 를 만들었다", calls.Load(), log.String())
	}
}

// TestAReasonlessRemoteFailureIsNotOurEngine — post-review P2-1. 사유 없는 JSON 거절은 우리 엔진의 답이
// 아니다(옛 포트에 다른 서버). 답으로 읽으면 자리가 죽은 client 에 영구히 묶인다.
func TestAReasonlessRemoteFailureIsNotOurEngine(t *testing.T) {
	var calls atomic.Int64
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return &a114FakeLifecycle{}, nil
		})
	attachment.interval = 0
	attachment.attach(live)
	live.fail(fmt.Errorf("%s: ", positionPolicyRemoteFailurePhrase))
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "사유 없는 거절 뒤 재부착 시도", func() bool { return calls.Load() >= 1 })
}

// TestARecoveredSeatStopsAsking — post-review N2. 한 번 실패한 뒤 **같은 client** 로 회복하면 시도 대상에서
// 빠져야 한다. 아니면 멀쩡한 자리가 간격마다 재-dial·교체된다(로그 없이).
func TestARecoveredSeatStopsAsking(t *testing.T) {
	var calls atomic.Int64
	live := &a114FakeLifecycle{}
	attachment := a114Attachment(t, a114Clock(), io.Discard,
		func(context.Context) (positionPolicyLifecycleClient, error) {
			calls.Add(1)
			return nil, errors.New("잠깐 못 붙음")
		})
	attachment.interval = 0
	attachment.attach(live)
	live.fail(&url.Error{Op: "Get", URL: "http://127.0.0.1:1/v1/positions", Err: syscall.ECONNRESET})
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "실패 뒤 시도 1회", func() bool { return calls.Load() == 1 && !attachment.inFlight() })
	live.fail(nil)
	// 회복 뒤 첫 호출은 성공을 알기 **전에** 입구에서 한 번 더 깨울 수 있다(자리가 아직 시도 대상이다).
	// 그 호출의 성공이 자리를 시도 대상에서 빼야 하므로, 그 뒤로는 시도 수가 늘지 않아야 한다.
	_, _ = attachment.List(context.Background())
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	settled := calls.Load()
	for range 5 {
		_, _ = attachment.List(context.Background())
		a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	}
	if got := calls.Load(); got != settled {
		t.Fatalf("회복한 자리가 계속 시도 대상이다: 회복 뒤 시도 %d번 더, want 0", got-settled)
	}
}

// a114UnusedGuard 는 filepath 를 쓰는 테스트가 빠져도 import 가 깨지지 않게 한다.
var _ = filepath.Join
