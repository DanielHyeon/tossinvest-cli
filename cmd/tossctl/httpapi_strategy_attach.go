package main

// httpapi_strategy_attach.go — a109 §2.3 (design D4): 조회 데몬이 엔진이 돌아오면
// 전략 화면에 **다시 붙는** 기계.
//
// # 왜 새 파일인가
//
// 기존 파일에 덧붙이면 tools/logic-map/check_analysis.py 가 삽입 지점의 이웃 함수까지
// 「수정됨」으로 보고, 아무도 편집하지 않은 함수의 증거 묶음을 요구한다(a102·a108 이
// 배운 것). 실제로 이 코드를 httpapi.go 안에 두자 `unavailableStrategyRuntime.Read` 가
// 그 이유만으로 잡혔다.
//
// httpapi.go 에 남는 것은 두 개다: 부팅 1회의 해석(`resolveStrategyRuntimeReader`)과
// 그것을 이 wrapper 로 감싸는 `strategyRuntimeReaderFor`.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// strategyRuntimeRedialInterval 은 재부착 시도 사이의 최소 간격이다 — a109 D4.
//
// 저장소 안의 선례(`positionPolicyRuntimeDescriptorReader`)는 같은 문제를 「rate limit
// 없이 매 read 재해결」로 푼다. 여기서는 그럴 수 없다: 전략 read 는 화면 폴링·SSE 로
// 고빈도이고 `strategyprojectionrpc.Dial` 본문에는 200ms connect probe 가 있다
// (transport_unix.go 의 `projectionSocketAccepts`). 엔진이 내려간 동안 매 read 재해결은
// read 마다 그 비용을 낸다.
//
// package 변수인 이유는 이 값이 **테스트 주입 가능**해야 하기 때문이다(freeze QA ②).
// 30초를 기다리는 테스트는 아무도 돌리지 않고, 안 돌리는 테스트는 없는 테스트다.
var strategyRuntimeRedialInterval = 30 * time.Second

// strategyRuntimeAttachment 는 전략 reader 자리를 감싸서 **엔진이 돌아오면 다시 붙는다**.
//
// # 무엇을 감싸는가 — 전부다 (freeze P0-1)
//
// nil(부재)·sentinel(붙지 못함)만 감싸면 세 번째 상태가 빠진다: 부팅 때 붙은 live
// client 는 엔진이 재시작하면 socket 도 토큰도 새 것이라 **영구 실패**하는데, 부착돼
// 있으니 재부착 대상이 아니다. 그래서 재부착 트리거는 「reader 가 부재이거나 **직전
// Read 가 실패했다**」이고, wrapper 는 live 를 포함한 세 값 전부를 감싼다.
//
// # 요청 goroutine 은 dial 하지 않는다 (freeze P0-2)
//
// `Dial` 본문에 200ms connect probe 가 있으므로 요청 경로에서 동기로 부르면
// 「요청 경로 비차단」이 구현 시점에 거짓이 된다. 요청이 하는 일은 **시도를 깨우는
// 것까지**이고, 응답은 언제나 지금 값으로 즉시 나간다.
//
// # 보고는 상태 전이 시 1회다 (freeze P2-4)
//
// 30초마다 실패를 stderr 에 찍지 않는다 — 엔진이 하루 동안 내려가 있으면 그 로그는
// 2880줄이고, 그 안에서 진짜 사건은 보이지 않는다. 시도 실패는 침묵, 부착·탈착
// **전이**만 한 줄씩 남긴다.
type strategyRuntimeAttachment struct {
	// resolve 는 디스크에서 reader 를 다시 구한다. 두 번째 반환값이 「live 로 붙었다」다.
	resolve func(context.Context) (httpapi.StrategyRuntimeReader, bool)
	// ctx 는 데몬의 수명이다. 시도 goroutine 이 이것을 쓴다 — 요청의 ctx 를 쓰면
	// 요청이 끝나는 순간 시도가 취소된다.
	ctx      context.Context
	now      func() time.Time
	interval time.Duration

	// log 는 **두 goroutine 이 함께 쓴다**: 부착 전이는 시도 goroutine 이, 탈착
	// 전이는 요청 goroutine 이 말한다. 그래서 자기 잠금을 단다 — 상태 잠금(mu)을
	// 쓰지 않는 이유는 느린 writer 하나가 모든 요청의 상태 조회를 멈추게 하면 안
	// 되기 때문이다.
	logMu sync.Mutex
	log   io.Writer

	mu       sync.Mutex
	reader   httpapi.StrategyRuntimeReader // nil 이면 **부재** — 화면은 dormant 다
	attached bool                          // 지금 붙어 있다고 보고한 상태 (전이 1회 로그)
	failed   bool                          // 부재이거나 직전 Read 가 실패했다 = 시도 대상
	lastTry  time.Time
	trying   bool // single-flight
	// seat 는 **reader 자리의 세대**다. 자리에 새 값이 앉을 때마다 하나 오른다.
	//
	// 왜 reader 값을 직접 비교하지 않는가: `Read` 는 잠금 밖에서 reader 를 부르므로
	// 그 사이에 시도가 성공해 다른 값이 앉을 수 있고, 그때 도착한 옛 실패는 **이미
	// 없는 endpoint** 의 소식이다(a109 §2-fix F2). 그것을 가리려면 「내가 읽은 그
	// 자리인가」를 물어야 하는데, 인터페이스 값의 `==` 는 동적 타입이 비교 불가면
	// **패닉한다** — `strategyprojection.Snapshot` 은 map 을 들고 있으므로 스냅샷을
	// 품은 값 reader(테스트의 `a109Projection` 이 그 모양이다)가 정확히 그 경우다.
	// 세대 번호는 같은 질문을 패닉 없이 답한다.
	seat uint64
}

// StrategyRuntimeConfigured 는 「이 배포에 전략 화면이 있는가」다
// (`httpapi.StrategyRuntimePresence`).
//
// 시도를 여기서도 깨우는 이유는 이것이 **부재 상태에서 요청 경로가 endpoint 를 만지는
// 유일한 지점**이기 때문이다. 소비자는 부재면 `Read` 를 아예 부르지 않으므로
// (`httpAPIReader.Snapshot`·router 의 REST 경로), 깨우기를 `Read` 에만 두면 냉부팅
// 순서 — 엔진이 나중에 뜨는 경우 — 가 영원히 회복하지 못한다.
func (a *strategyRuntimeAttachment) StrategyRuntimeConfigured() bool {
	reader, _, wanted := a.state()
	if wanted {
		a.wake()
	}
	return reader != nil
}

func (a *strategyRuntimeAttachment) Read(ctx context.Context) (strategyprojection.Snapshot, error) {
	reader, seat, wanted := a.state()
	if wanted {
		a.wake()
	}
	if reader == nil {
		// 소비자는 부재를 먼저 묻고 dormant 를 그리므로 여기 오지 않는다. 그래도
		// 값을 지어내지 않는다 — 부재를 스냅샷으로 답하면 그것이 새로운 접힘이다.
		return strategyprojection.Snapshot{},
			errors.New("strategy runtime projection is not attached")
	}
	snapshot, err := reader.Read(ctx)
	if a.observe(ctx, seat, err) {
		// 직전 읽기가 실패했다 = 지금부터 재부착 대상이다. 다음 요청을 기다리지
		// 않는다 — 폴링이 뜸한 시간대에 그 기다림이 곧 화면 공백이 된다.
		a.wake()
	}
	return snapshot, err
}

// state 는 지금 자리에 있는 reader 와 그 **자리의 세대**, 그리고 시도 대상 여부다.
func (a *strategyRuntimeAttachment) state() (httpapi.StrategyRuntimeReader, uint64, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.reader, a.seat, a.failed
}

// inFlight 는 지금 시도가 돌고 있는지다.
//
// single-flight 는 **관측 가능해야** 하는 불변식이다: 이 상태를 밖에서 볼 수 없으면
// 「동시 요청 20건이 dial 을 한 번만 불렀다」를 잰 뒤 그 시도가 끝났는지 모르는 채로
// 다음 단정을 쓰게 되고, 그 테스트는 느린 호스트에서 흔들린다.
func (a *strategyRuntimeAttachment) inFlight() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.trying
}

// attach 는 부팅 1회의 해석 결과를 그대로 받는다. 세 값 전부가 유효한 출발점이다.
func (a *strategyRuntimeAttachment) attach(reader httpapi.StrategyRuntimeReader, live bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reader, a.attached, a.failed = reader, live, !live
	a.seat++
}

// wake 는 시도를 깨운다. 이 함수는 **절대 막히지 않는다**.
func (a *strategyRuntimeAttachment) wake() {
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
func (a *strategyRuntimeAttachment) attempt() {
	reader, live := a.resolve(a.ctx)
	a.mu.Lock()
	a.trying = false
	if !live {
		// 실패는 침묵이고, 지금 화면 값을 흔들지 않는다. 실패한 해석으로 현재 reader 를
		// 갈아끼우면 「붙어 있다가 잠깐 못 읽는 중」이 「이 배포는 전략 화면을 안 쓴다」로
		// 바뀔 수 있다 — 같은 접힘이 뒷문으로 들어온다.
		a.mu.Unlock()
		return
	}
	a.reader, a.failed = reader, false
	a.seat++
	announce := !a.attached
	a.attached = true
	a.mu.Unlock()
	if announce {
		a.report("note: strategy runtime projection에 다시 붙었다 — 전략 화면이 회복된다.\n")
	}
}

// report 는 전이 한 줄을 남긴다. 잠금은 이 한 곳뿐이다.
func (a *strategyRuntimeAttachment) report(format string, args ...any) {
	a.logMu.Lock()
	defer a.logMu.Unlock()
	fmt.Fprintf(a.log, format, args...)
}

// requestCancelled 는 「이 실패가 endpoint 가 아니라 **요청자** 때문인가」다 —
// a109 §2-fix F1 (A2 P2-1).
//
// REST 경로는 `request.Context()` 를 그대로 넘긴다(`internal/httpapi/router.go` 의
// `read`). 그래서 브라우저 탭을 닫거나 요청 timeout 이 지난 것 하나가
// `context.Canceled`·`DeadlineExceeded` 로 도착한다. 그것을 endpoint 판정으로 쓰면
// 멀쩡한 엔진에 붙어 있는 client 가 요청 취소마다 교체 후보가 되고, 화면 폴링이
// 잦을수록 재-dial(200ms probe)이 잦아진다.
//
// **호출자 ctx 가 실제로 끝났을 때만** 그렇게 읽는다. ctx 는 멀쩡한데 endpoint 쪽이
// 취소 계열 오류를 돌려준 경우는 진짜 실패이므로 그대로 판정에 쓴다.
func requestCancelled(ctx context.Context, err error) bool {
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return ctx != nil && ctx.Err() != nil
}

// observe 는 방금 읽기의 결과를 기록하고 **탈착 전이만** 보고한다.
// 반환값은 「지금 재부착 대상인가」다.
//
// seat 는 그 읽기가 **어느 자리의 reader** 로 이뤄졌는지다. `Read` 는 잠금 밖에서
// 읽으므로 그 사이에 시도가 성공해 새 reader 가 앉을 수 있고, 그때 도착하는 옛
// 실패로 방금 성립한 부착을 뒤엎으면 회복 직후에 다시 탈착으로 떨어진다
// (a109 §2-fix F2 — A2 P2-2).
func (a *strategyRuntimeAttachment) observe(ctx context.Context, seat uint64, err error) bool {
	if err != nil && requestCancelled(ctx, err) {
		// 상태를 **하나도** 바꾸지 않는다. 성공도 실패도 아니라 판정 자체가 없다.
		return false
	}
	a.mu.Lock()
	if seat != a.seat {
		// 내가 읽은 자리는 이미 없다. 그 소식으로 지금 자리를 판정하지 않는다.
		a.mu.Unlock()
		return false
	}
	if err == nil {
		a.failed = false
		a.mu.Unlock()
		return false
	}
	a.failed = true
	announce := a.attached
	a.attached = false
	interval := a.interval
	a.mu.Unlock()
	if announce {
		a.report("note: strategy runtime projection 읽기가 실패했다 (%v)\n"+
			"      데몬은 그대로 돈다 — 엔진이 돌아오면 최소 %s 간격의 재시도가 다시 붙인다.\n",
			err, interval)
	}
	return true
}

var _ httpapi.StrategyRuntimePresence = (*strategyRuntimeAttachment)(nil)
