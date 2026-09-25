package main

// console_lifecycle_attach.go — a114: 콘솔이 엔진 포지션 정책 control plane(lifecycle)에
// **엔진이 늦게 뜨거나 재시작해도 다시 붙는** 기계.
//
// # 병 (a109 design 선언된 생략 P2-7)
//
// 콘솔은 부팅 때 한 번 `positionpolicyrpc.Dial` 하고 결과를 굳혔다. 엔진이 콘솔보다 늦게
// 뜨면(descriptor 부재) commander 는 nil 이 되어 정책 화면이 「배선되지 않았다」로 영구히
// 남고, 부팅 때 붙은 client 는 엔진이 재시작하면(새 포트·새 토큰) 영구 실패한다. 둘 다
// 콘솔을 재시작해야만 풀렸다.
//
// # 원형 — a109 D4 의 httpapi 재부착(`httpapi_strategy_attach.go`)
//
// 같은 논지다: 재부착은 endpoint 주인이 아니라 **소비자의 것**이다. 같은 네 규칙을 옮긴다.
//
//	① 자리의 모든 상태를 감싼다 — 부재뿐 아니라 부착 후 실패한 live client 도 재부착 대상.
//	② 요청 goroutine 은 dial 하지 않는다 — `Dial` 은 health GET(연결)을 품는다. 시도는
//	   백그라운드 single-flight 이고 최소 간격(rate limit) 아래에서만 돈다.
//	③ 전이 시에만 한 줄 로그 — 30초마다 실패를 찍지 않는다.
//	④ 밀려난 client 는 놓아 준다(io.Closer 면 닫는다).
//
// # 새 파일인 이유
//
// 기존 파일에 덧붙이면 logic-map 게이트가 삽입 지점의 이웃 함수까지 「수정됨」으로 본다.
//
// # ⛔ 이 wrapper 는 명령을 **다시 보내지 않는다**
//
// Preview·Apply·격리 해제는 엔진의 상태를 바꾸는 명령이다. 호출이 실패해도 그 명령을 새
// client 로 재전송하지 않는다 — 재부착은 **다음** 호출을 위한 것이고, 실패한 호출의 결과는
// 운영자에게 그대로 간다(엔진이 이미 적용했을 수 있는 명령을 두 번 보내지 않는다).

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// consolePositionPolicyRedialInterval 은 재부착 시도 사이의 최소 간격이다.
//
// httpapi 의 `strategyRuntimeRedialInterval` 과 같은 30초다. package 변수인 이유도 같다 —
// 테스트가 주입할 수 있어야 한다(30초를 기다리는 테스트는 아무도 안 돌린다).
var consolePositionPolicyRedialInterval = 30 * time.Second

// errPositionPolicyLifecycleDetached 는 붙어 있지 않을 때 화면이 받는 오류다.
//
// # 이 문장은 엔진 부재를 **단정하지 않는다** (a114 spec · a109 D3a-2 준용)
//
// 붙지 못한 이유는 여럿이다: 엔진이 내려갔다, 아직 기동 중이다, 엔진이 이 표면 없이 강등
// 부팅했다(a109 D3), descriptor 를 읽을 수 없다. 운영자가 「엔진이 없다」를 읽고 엔진을
// 재시작하면 강등 원인이 결정적인 경우 같은 상태가 재현된다. 그래서 가능성을 나열하고,
// 무엇을 확인하면 되는지와 회복이 저절로 온다는 사실을 함께 적는다.
var errPositionPolicyLifecycleDetached = errors.New(
	"엔진 포지션 정책 control plane 에 붙어 있지 않다 — 엔진이 내려갔거나 아직 기동 중이거나 " +
		"이 표면 없이 강등 부팅했을 수 있다(엔진 로그의 강등 보고를 확인하라). " +
		"엔진이 endpoint 를 발행하면 콘솔 재시작 없이 다시 붙는다")

// positionPolicyLifecycleAttachment 는 lifecycle client 자리를 감싸 엔진이 돌아오면 다시 붙는다.
//
// 자리의 상태는 셋이다: 비어 있음(한 번도 못 붙음 · 부팅 때 엔진 부재), live, 그리고 live
// 였지만 직전 호출이 **endpoint 실패**로 끝남. 뒤의 둘을 가르는 것이 `failed` 다.
type positionPolicyLifecycleAttachment struct {
	// resolve 는 디스크에서 client 를 다시 구한다. 이 함수만 dial 한다.
	resolve func(context.Context) (positionPolicyLifecycleClient, error)
	// ctx 는 콘솔의 수명이다. 시도 goroutine 이 쓴다 — 요청 ctx 를 쓰면 요청이 끝날 때
	// 시도도 취소된다.
	ctx      context.Context
	now      func() time.Time
	interval time.Duration

	// log 는 시도 goroutine 과 요청 goroutine 이 함께 쓴다 — 자기 잠금을 단다(상태 잠금과
	// 분리: 느린 writer 가 모든 요청을 멈추게 하면 안 된다).
	logMu sync.Mutex
	log   io.Writer

	mu       sync.Mutex
	client   positionPolicyLifecycleClient // nil 이면 비어 있음
	attached bool                          // 지금 붙어 있다고 보고한 상태(전이 1회 로그)
	failed   bool                          // 비어 있거나 직전 호출이 endpoint 실패 = 시도 대상
	lastTry  time.Time
	trying   bool   // single-flight
	seat     uint64 // 자리의 세대 — 옛 자리의 늦은 실패가 새 부착을 뒤엎지 못하게 한다
}

// attach 는 부팅 1회의 해석 결과를 받는다. 실패도 유효한 출발점이다(비어 있음).
func (a *positionPolicyLifecycleAttachment) attach(client positionPolicyLifecycleClient) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.client, a.attached, a.failed = client, client != nil, client == nil
	a.seat++
}

// current 는 지금 자리의 client 와 그 세대, 시도 대상 여부다.
func (a *positionPolicyLifecycleAttachment) current() (positionPolicyLifecycleClient, uint64, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.client, a.seat, a.failed
}

// inFlight 는 시도가 돌고 있는지다 — single-flight 를 테스트가 관측하기 위한 것.
func (a *positionPolicyLifecycleAttachment) inFlight() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.trying
}

// seatFor 는 요청 경로의 공통 입구다: 지금 client 를 돌려주고, 시도 대상이면 시도를 깨운다.
// 비어 있으면 client 는 nil 이다 — 호출자는 dial 하지 않고 즉시 detached 오류로 답한다.
func (a *positionPolicyLifecycleAttachment) seatFor() (positionPolicyLifecycleClient, uint64) {
	client, seat, wanted := a.current()
	if wanted {
		a.wake()
	}
	return client, seat
}

// wake 는 시도를 깨운다. 절대 막히지 않는다 — goroutine 을 띄우거나 아무것도 안 한다.
func (a *positionPolicyLifecycleAttachment) wake() {
	a.mu.Lock()
	now := a.now()
	tooSoon := !a.lastTry.IsZero() && now.Sub(a.lastTry) < a.interval
	if a.trying || tooSoon || a.ctx.Err() != nil {
		a.mu.Unlock()
		return
	}
	a.trying, a.lastTry = true, now
	a.mu.Unlock()
	go a.attempt()
}

// attempt 는 **여기서만** dial 한다. 이 goroutine 은 요청 경로 밖이다.
func (a *positionPolicyLifecycleAttachment) attempt() {
	client, err := a.resolve(a.ctx)
	a.mu.Lock()
	a.trying = false
	if err != nil || client == nil {
		// 실패는 침묵이고 자리를 흔들지 않는다 — 붙어 있던 client 를 실패한 해석으로
		// 비우면 「잠깐 못 읽는 중」이 「비어 있음」으로 바뀐다. 자리는 다음 성공이 바꾼다.
		a.mu.Unlock()
		return
	}
	evicted := a.client
	a.client, a.failed = client, false
	a.seat++
	announce := !a.attached
	a.attached = true
	a.mu.Unlock()
	closeEvictedLifecycleClient(evicted)
	if announce {
		a.reportAttached()
	}
}

// closeEvictedLifecycleClient 는 자리에서 밀려난 client 를 놓아 준다.
//
// io.Closer 면 닫는다. `positionpolicyrpc.Client` 는 오늘 Close 가 없다 — 그 client 의 유휴
// 연결은 엔진 서버의 IdleTimeout(15s, internal/app/engine/position_policy_transport.go)이
// 닫거나, 엔진이 죽었으면 커널이 닫는다. 그래서 밀려난 값이 쥔 것은 길어야 15초다
// (issues.md — Close 추가는 이 change 의 파일 표면 밖).
func closeEvictedLifecycleClient(client positionPolicyLifecycleClient) {
	if closer, ok := client.(io.Closer); ok {
		_ = closer.Close()
	}
}

// pump 는 **화면 요청이 없어도** 시도를 깨운다 — a114 freeze P1-2.
//
// 콘솔에는 httpapi 의 publisher 같은 상시 구동원이 없다. 화면이 안 열려 있으면 엔진이 떠도 아무도
// wake 를 부르지 않고, 콘솔 자신이 autostart 한 엔진도 그렇다. 그래서 간격마다 한 번, 자리가 시도
// 대상일 때만 깨운다. 과빈도는 wake 의 rate limit·single-flight 가 막는다. 간격이 0 이하면(테스트)
// 띄우지 않는다 — 0 주기 ticker 는 패닉이다.
func (a *positionPolicyLifecycleAttachment) pump() {
	if a.interval <= 0 {
		return
	}
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if _, _, wanted := a.current(); wanted {
				a.wake()
			}
		}
	}
}

func (a *positionPolicyLifecycleAttachment) reportAttached() {
	a.report("note: 엔진 포지션 정책 control plane 에 붙었다 — 정책 화면이 엔진 상태를 반영한다.\n")
}

func (a *positionPolicyLifecycleAttachment) report(format string, args ...any) {
	a.logMu.Lock()
	defer a.logMu.Unlock()
	fmt.Fprintf(a.log, format, args...)
}

// positionPolicyEngineAnswers 는 엔진이 **답한** 타입 있는 거절들이다 — 목록은 이 한 곳에만 둔다.
//
// 이 오류들은 endpoint 가 살아서 요청을 판정했다는 증거다(검증 실패·version 충돌·capability 만료 등).
// 그것을 endpoint 실패로 읽으면 운영자의 거절된 Apply 한 번이 탈착 로그와 재-dial 을 만든다.
// 완전성은 `TestEveryEngineAnswerIsOnTheList` 가 두 패키지의 exported Err 선언을 세어 고정한다.
var positionPolicyEngineAnswers = []error{
	positionpolicy.ErrInvalidRequest, positionpolicy.ErrPositionNotFound,
	positionpolicy.ErrVersionMismatch, positionpolicy.ErrExitConflict,
	positionpolicy.ErrUnknownState, positionpolicy.ErrIneligible,
	positionpolicy.ErrCapabilityInvalid, positionpolicy.ErrCapabilityTooEarly,
	positionpolicy.ErrCapabilityExpired, positionpolicy.ErrCapabilityStale,
	positionpolicy.ErrConfirmationRequired,
	exitquarantine.ErrInvalidRequest, exitquarantine.ErrNotQuarantined,
	exitquarantine.ErrVersionMismatch, exitquarantine.ErrCapabilityInvalid,
	exitquarantine.ErrCapabilityTooEarly, exitquarantine.ErrCapabilityExpired,
	exitquarantine.ErrConfirmationRequired, exitquarantine.ErrUnwired,
}

// 다른 파일의 문구 셋이다 — 원본에서 찾는 테스트(`TestTheAnswerPhrasesAreTheSourcesOwn`)가 변경을 잡는다.
const (
	// 두 decoder 가 **모르는 코드**(엔진의 `internal` 등)에 쓰는 문구
	// (internal/positionpolicyrpc/client.go `decodeRemoteError`, exit_quarantine_client.go).
	positionPolicyRemoteFailurePhrase = "position policy control: remote failure"
	exitQuarantineRemoteFailurePhrase = "exit quarantine control: remote failure"
	// 엔진 auth 가 토큰을 거절할 때의 고정 문구(internal/app/engine/position_policy_transport.go `auth`).
	engineBearerRejectedPhrase = "local bearer token rejected"
)

// endpointAnswered 는 「이 오류는 엔진이 판정해서 돌려준 답인가」다 — a114 freeze P1-1 로 넓혔다.
//
// 여섯 메서드가 자리 하나를 공유한다. 한 메서드가 엔진의 `internal`(500) 을 반복해서 받는 동안 다른
// 메서드가 성공하면, 그것을 탈착으로 읽는 판정은 매 렌더 「붙었다/실패했다」를 번갈아 찍고 멀쩡한
// client 를 갈아끼운다. 그래서 **코드가 붙은 모든 rpcError** 를 답으로 본다: 타입 있는 거절과, 두
// decoder 가 모르는 코드에 쓰는 remote failure 문구.
//
// 예외 하나 — **토큰 거절은 탈착이다.** 재시작한 엔진이 같은 포트에 다른 토큰으로 뜨면 옛 client 가
// 받는 답이 이것이고, 답으로 읽으면 콘솔이 옛 토큰에 영구히 묶인다.
//
// 그 밖(연결 거부·timeout 의 `*url.Error`, 코드 없는 `HTTP %d`, 응답 해독 실패 — 그 포트에 우리 엔진이
// 아닌 것이 있다)은 전부 탈착이다 — **모르는 오류의 기본값은 재부착**이다.
func endpointAnswered(err error) bool {
	text := err.Error()
	if strings.Contains(text, engineBearerRejectedPhrase) {
		return false
	}
	for _, answer := range positionPolicyEngineAnswers {
		if errors.Is(err, answer) {
			return true
		}
	}
	return strings.Contains(text, positionPolicyRemoteFailurePhrase) ||
		strings.Contains(text, exitQuarantineRemoteFailurePhrase)
}

// observe 는 방금 호출의 결과로 자리 상태를 갱신하고 **전이**만 보고한다.
//
// 요청자가 취소한 호출(`requestCancelled` — httpapi 재부착과 같은 판정)은 아무것도 바꾸지
// 않는다. 옛 자리의 늦은 결과(seat 불일치)도 바꾸지 않는다. 엔진이 답한 거절은 성공과 같다 —
// 자리는 살아 있다.
func (a *positionPolicyLifecycleAttachment) observe(ctx context.Context, seat uint64, err error) {
	if err != nil && requestCancelled(ctx, err) {
		return
	}
	a.mu.Lock()
	if seat != a.seat {
		a.mu.Unlock()
		return
	}
	if err == nil || endpointAnswered(err) {
		a.failed = false
		announce := !a.attached
		a.attached = true
		a.mu.Unlock()
		if announce {
			a.reportAttached()
		}
		return
	}
	a.failed = true
	announce := a.attached
	a.attached = false
	interval := a.interval
	a.mu.Unlock()
	if announce {
		a.report("note: 엔진 포지션 정책 control plane 호출이 실패했다 (%v)\n"+
			"      콘솔은 그대로 돈다 — 엔진이 돌아오면 최소 %s 간격의 재시도가 다시 붙인다.\n",
			err, interval)
	}
	// 직전 호출이 실패했다 = 지금부터 시도 대상이다. 다음 요청을 기다리지 않는다.
	a.wake()
}

// ---- lifecycle 세 메서드 -------------------------------------------------------------
//
// 셋 다 같은 모양이다: 지금 자리의 client 로 **한 번** 부르고, 그 결과로 자리를 판정한다.
// 자리가 비어 있으면 dial 하지 않고 즉시 detached 오류로 답한다(시도는 깨운다).

func (a *positionPolicyLifecycleAttachment) List(ctx context.Context) ([]positionpolicy.State, error) {
	client, seat := a.seatFor()
	if client == nil {
		return nil, errPositionPolicyLifecycleDetached
	}
	states, err := client.List(ctx)
	a.observe(ctx, seat, err)
	return states, err
}

func (a *positionPolicyLifecycleAttachment) Preview(ctx context.Context,
	req positionpolicy.Request) (positionpolicy.Preview, error) {
	client, seat := a.seatFor()
	if client == nil {
		return positionpolicy.Preview{}, errPositionPolicyLifecycleDetached
	}
	preview, err := client.Preview(ctx, req)
	a.observe(ctx, seat, err)
	return preview, err
}

// Apply 는 엔진 상태를 바꾸는 명령이다. 실패해도 **다시 보내지 않는다** — 재부착은 다음 호출을 위한 것이다.
func (a *positionPolicyLifecycleAttachment) Apply(ctx context.Context,
	req positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	client, seat := a.seatFor()
	if client == nil {
		return positionpolicy.State{}, errPositionPolicyLifecycleDetached
	}
	state, err := client.Apply(ctx, req)
	a.observe(ctx, seat, err)
	return state, err
}

// ---- 격리 해제 세 메서드 (a079) ------------------------------------------------------
//
// `consolePositionPolicyCommander.quarantineClient` 는 lifecycle 자리에 `exitQuarantineClient` 를
// 타입 단언해 격리 해제를 발견한다. 이 wrapper 가 세 메서드를 갖지 않으면 **격리 해제 버튼이
// 사라진다** — 격리된 포지션의 손절 포함 미판정 상태를 푸는 유일한 장중 경로다.
//
// 붙은 client 가 그 메서드를 모르면(a079 이전 엔진) 오늘처럼 `exitquarantine.ErrUnwired` 다.
// 비어 있으면 detached 오류다 — ErrUnwired 로 감싸지 않는다: 그 화면 문구는 「이 빌드에 배선되지
// 않았다」를 포함하므로 부착 전 상태에는 거짓이다.

// quarantineSeat 은 격리 호출의 공통 입구다. ok 가 false 면 err 로 즉시 답한다.
func (a *positionPolicyLifecycleAttachment) quarantineSeat() (exitQuarantineClient, uint64, error) {
	client, seat := a.seatFor()
	if client == nil {
		return nil, seat, errPositionPolicyLifecycleDetached
	}
	quarantine, ok := client.(exitQuarantineClient)
	if !ok {
		return nil, seat, exitquarantine.ErrUnwired
	}
	return quarantine, seat, nil
}

func (a *positionPolicyLifecycleAttachment) Quarantines(ctx context.Context) ([]exitquarantine.Row, error) {
	client, seat, err := a.quarantineSeat()
	if err != nil {
		return nil, err
	}
	rows, err := client.Quarantines(ctx)
	a.observe(ctx, seat, err)
	return rows, err
}

func (a *positionPolicyLifecycleAttachment) PreviewQuarantineRelease(ctx context.Context,
	req exitquarantine.Request) (exitquarantine.Preview, error) {
	client, seat, err := a.quarantineSeat()
	if err != nil {
		return exitquarantine.Preview{}, err
	}
	preview, err := client.PreviewQuarantineRelease(ctx, req)
	a.observe(ctx, seat, err)
	return preview, err
}

// ReleaseQuarantine 은 명령이다 — Apply 와 같이 다시 보내지 않는다.
func (a *positionPolicyLifecycleAttachment) ReleaseQuarantine(ctx context.Context,
	req exitquarantine.ApplyRequest) (exitquarantine.Result, error) {
	client, seat, err := a.quarantineSeat()
	if err != nil {
		return exitquarantine.Result{}, err
	}
	result, err := client.ReleaseQuarantine(ctx, req)
	a.observe(ctx, seat, err)
	return result, err
}

var (
	_ positionPolicyLifecycleClient = (*positionPolicyLifecycleAttachment)(nil)
	_ exitQuarantineClient          = (*positionPolicyLifecycleAttachment)(nil)
)

// ---- 부팅 -------------------------------------------------------------------------

// consolePositionPolicyCommanderFor 는 콘솔의 포지션 정책 commander 를 만든다 — a114.
//
// 부팅 1회 해석은 **오늘처럼 동기로** 한 번 한다: 엔진이 이미 떠 있으면 첫 화면부터 붙어 있다.
// 달라진 것은 그 결과가 굳지 않는다는 것이다 — 실패하면 비어 있는 자리로 출발하고, 붙으면 live
// 로 출발하며, 어느 쪽이든 wrapper 가 이후를 맡는다.
//
// engineDir 가 있을 때만 부른다. 반환은 언제나 non-nil 이다 — 화면의 「배선되지 않았다」는 engineDir
// 를 해석하지 못한 진짜 미배선에만 남는다.
func consolePositionPolicyCommanderFor(ctx context.Context, engineDir string,
	errOut io.Writer) *consolePositionPolicyCommander {
	descriptorPath := positionpolicyrpc.DescriptorPath(engineDir)
	resolve := func(ctx context.Context) (positionPolicyLifecycleClient, error) {
		// descriptor 를 먼저 stat 하는 이유: 부재는 「엔진이 아직 발행하지 않았다」라는 정상 상태이고,
		// 부팅 경고를 남길 사유가 아니다(오늘의 동작). Dial 의 오류는 그것을 감싸 구분이 흐려진다.
		if _, err := os.Stat(descriptorPath); err != nil {
			return nil, err
		}
		client, err := positionpolicyrpc.Dial(ctx, descriptorPath)
		if err != nil {
			// 타입 있는 nil 포인터를 인터페이스에 담아 돌려주지 않는다 — 「nil 이 아닌 부재」가 된다.
			return nil, err
		}
		return client, nil
	}
	attachment := &positionPolicyLifecycleAttachment{
		resolve: resolve, ctx: ctx, now: time.Now,
		interval: consolePositionPolicyRedialInterval, log: errOut,
	}
	client, err := resolve(ctx)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(errOut, "엔진 포지션 정책 control plane에 지금 연결할 수 없다 (%v). "+
			"정책 화면은 붙기 전까지 읽기 실패를 상태로 표시하고, 엔진이 돌아오면 콘솔 재시작 없이 다시 붙는다.\n", err)
	}
	// 부팅 해석은 `lastTry` 를 찍지 않는다(freeze P1-2): 부팅 직후의 첫 wake 가 간격만큼 막히지 않는다.
	attachment.attach(client)
	go attachment.pump()
	return &consolePositionPolicyCommander{
		lifecycle: attachment,
		runtime: positionPolicyRuntimeDescriptorReader{
			descriptorPath: positionpolicyrpc.RuntimeDescriptorPath(engineDir),
		},
	}
}
