package main

// console_strategy_attach.go — a115: 콘솔 전략 화면이 엔진의 전략 projection 에 **엔진이 늦게 뜨거나
// 재시작해도 다시 붙는** 배선.
//
// # 병 (a109 review §2 A2 P1-1 — 선언된 생략)
//
// 콘솔은 부팅 때 한 번 `strategyprojectionrpc.Dial` 하고, 실패를 **nil 로 접었다**. 화면은 nil 을
// 「이 배포는 전략 화면을 안 쓴다」(dormant)로 그리므로, descriptor 가 남은 채 엔진이 연결을 받지
// 않는 상태가 미구성으로 오귀속됐다. 그리고 부팅 1회라 엔진이 늦게 뜨거나 재시작하면 콘솔을
// 재시작할 때까지 회복하지 않았다.
//
// # 기계는 옮겨 적지 않고 재사용한다 (design D1)
//
// httpapi 의 재부착 wrapper(`strategyRuntimeAttachment`, httpapi_strategy_attach.go — a109 D4)가
// 세 자리 상태(부재·sentinel·live)와 요청 경로 비차단·single-flight·rate limit·전이 1회 로그·취소·
// 늦은 실패·밀려난 client Close 를 이미 갖췄다. 복사한 기계는 어긋나기 시작한 기계다. 여기에 더하는
// 것은 둘뿐이다: 콘솔용 해석(경고 문구만 다르다)과 **펌프**(콘솔엔 publisher 가 없다).
//
// # 새 파일인 이유
//
// 기존 파일에 덧붙이면 logic-map 게이트가 삽입 지점의 이웃 함수까지 「수정됨」으로 본다.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
)

// consoleStrategyRuntimeRedialInterval 은 콘솔 전략 재부착 시도 사이의 최소 간격이다.
//
// httpapi 의 `strategyRuntimeRedialInterval` 과 같은 30초지만 **공유하지 않는다**(freeze 리뷰 P2-5):
// a109 시험이 그 변수를 30s 로 고정하고 주입하므로, 공유하면 한쪽 시험의 주입이 다른 쪽 배선을 흔든다.
// package 변수인 이유는 시험 주입 — 30초를 기다리는 시험은 아무도 안 돌린다.
var consoleStrategyRuntimeRedialInterval = 30 * time.Second

// resolveConsoleStrategyRuntime 은 디스크의 전략 endpoint 상태를 reader 하나로 접는다 — 콘솔판.
//
// 판정은 httpapi 의 `resolveStrategyRuntimeReader` 와 **같다**(동치는 `TestTheConsoleResolutionMatchesTheDaemons`
// 가 같은 디스크 상태에서 고정한다): descriptor 부재 → nil(부재, 조용히) · 비부재 stat 오류 → 경고 + nil
// (선언된 접힘 — httpapi 와 동일, design 「구성」의 뜻) · dial 실패 → **sentinel**(도달 불가) · 성공 → client.
// 다른 것은 경고 문구뿐이다: httpapi 판은 「데몬」을 말하고 root 에서 디렉터리를 다시 푼다. 콘솔은 이미 푼
// engineDir 를 쓴다.
//
// ⛔ 문구는 「전략 화면은 dormant로 뜬다」를 말하지 않는다 — a115 뒤 죽은 descriptor 는 도달 불가로 뜨고,
// 어느 경우든 콘솔 재시작 없이 다시 붙는다(freeze 리뷰 P2-3).
func resolveConsoleStrategyRuntime(ctx context.Context, engineDir string,
	errOut io.Writer) (httpapi.StrategyRuntimeReader, bool) {
	descriptor := strategyprojectionrpc.DescriptorPath(engineDir)
	if _, statErr := os.Stat(descriptor); statErr != nil {
		// 부재는 「엔진이 아직 발행하지 않았다」라는 정상 상태이므로 경고하지 않는다(오늘의 동작).
		if !errors.Is(statErr, os.ErrNotExist) {
			fmt.Fprintf(errOut, "엔진 전략 runtime endpoint를 확인할 수 없다 (%v). "+
				"전략 화면은 확인할 수 있을 때까지 미기동 안내를 표시하고, 경로·볼륨 배치가 고쳐지면 "+
				"콘솔 재시작 없이 다시 붙는다.\n", statErr)
		}
		return nil, false
	}
	client, dialErr := strategyprojectionrpc.Dial(ctx, descriptor)
	if dialErr != nil {
		fmt.Fprintf(errOut, "엔진 전략 runtime projection에 지금 연결할 수 없다 (%v). "+
			"전략 화면은 붙기 전까지 도달 불가를 표시하고, 엔진이 돌아오면 콘솔 재시작 없이 다시 붙는다.\n", dialErr)
		// nil 이 아니라 sentinel 이다. descriptor 가 **있는데** 못 붙은 것은 부재가 아니므로 부재의 화면
		// 값(dormant)을 쓰면 안 된다 — 이것이 a115 가 지우는 접힘이다.
		return unavailableStrategyRuntime{cause: dialErr}, false
	}
	return client, true
}

// consoleStrategyRuntimeReaderFor 는 콘솔의 전략 reader 를 만든다 — a115.
//
// 부팅 1회 해석은 **오늘처럼 동기로** 한 번 한다: 엔진이 이미 떠 있으면 첫 화면부터 붙어 있다. 달라진
// 것은 그 결과가 굳지 않는다는 것이다 — 세 값(부재·sentinel·client) 어느 것으로 출발하든 wrapper 가
// 이후를 맡는다. 부팅 해석은 `lastTry` 를 찍지 않으므로 첫 wake 가 간격만큼 막히지 않는다(a114 C4).
//
// engineDir 가 있을 때만 부른다. 반환은 **언제나 non-nil** 이다(freeze 리뷰 P2-4): nil
// `*strategyRuntimeAttachment` 가 interface 에 담기면 presence 질문이 nil mutex 에서 패닉한다. 화면의
// dormant 는 engineDir 를 풀지 못한 진짜 미배선(nil interface)과 부재 신호에만 남는다.
func consoleStrategyRuntimeReaderFor(ctx context.Context, engineDir string,
	errOut io.Writer) *strategyRuntimeAttachment {
	attachment := &strategyRuntimeAttachment{
		resolve: func(attemptCtx context.Context) (httpapi.StrategyRuntimeReader, bool) {
			// 재시도의 경고는 버린다 — 보고는 상태 전이 시 1회다(wrapper 의 전이 로그). errOut 을 그대로
			// 넘기면 엔진이 하루 내려간 배포에서 같은 경고가 2880번 찍힌다(httpapi.go 선례, 리뷰 P2-3).
			return resolveConsoleStrategyRuntime(attemptCtx, engineDir, io.Discard)
		},
		ctx: ctx, now: time.Now, interval: consoleStrategyRuntimeRedialInterval, log: errOut,
	}
	attachment.attach(resolveConsoleStrategyRuntime(ctx, engineDir, errOut))
	go pumpConsoleStrategyRuntime(attachment)
	return attachment
}

// pumpConsoleStrategyRuntime 은 **화면 요청이 없어도** 재부착 시도를 깨운다 — a115 design D1.
//
// 콘솔에는 httpapi 의 publisher 같은 상시 구동원이 없다. 화면이 안 열려 있으면 엔진이 떠도 아무도 wake 를
// 부르지 않고, 콘솔 자신이 autostart 한 엔진도 그렇다.
//
// # 깨우기는 **무조건**이다 (freeze 리뷰 P1-2)
//
// 「시도 대상(`failed`)일 때만 깨우기」는 a109 G2 가 기각한 게이트다(httpapi_strategy_attach.go
// `StrategyRuntimeConfigured` 주석): 엔진이 재시작하면 부팅 때 잡은 client 는 **live 로 보이지만 죽어**
// 있고, 그 사실은 누군가 Read 해야 드러난다. 렌더가 없는 동안 그 자리는 죽은 채 남고 첫 렌더가 「읽지
// 못했다」를 그린다. 과빈도는 게이트가 아니라 wake 안의 rate limit·single-flight 가 막으므로 비용은
// 「간격당 시도 1회」로 고정된다.
//
// 틱은 간격의 **절반**이다(a114 post-review P2-2): 틱이 간격과 같으면 스케줄 지터로 `now-lastTry < interval`
// 인 틱이 건너뛰어져 재시도 간격이 두 배로 벌어진다. 간격이 0 이하면 띄우지 않는다 — 0 주기 ticker 는
// 패닉이다. 콘솔 ctx 가 끝나면 돌아온다.
func pumpConsoleStrategyRuntime(attachment *strategyRuntimeAttachment) {
	if attachment.interval <= 0 {
		return
	}
	ticker := time.NewTicker(max(attachment.interval/2, time.Millisecond))
	defer ticker.Stop()
	for {
		select {
		case <-attachment.ctx.Done():
			return
		case <-ticker.C:
			attachment.wake()
		}
	}
}

// 컴파일 결속(design D2, freeze 리뷰 P1-1): 구조적 인터페이스만으로는 메서드 이름이 바뀌어도 컴파일이
// 통과하고, 콘솔 화면의 부재가 조용히 「읽지 못했다」로 회귀한다. 콘솔이 쓰는 판정의 인터페이스에 직접 묶는다.
var _ strategyprojection.StrategyRuntimePresence = (*strategyRuntimeAttachment)(nil)
