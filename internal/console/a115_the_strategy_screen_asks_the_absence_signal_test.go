package console

// a115 — 콘솔 전략 화면은 nil 이 아니라 **부재 신호**를 묻는다 (design D2).
//
// a115 는 콘솔 부팅의 전략 reader 자리에 재부착 wrapper(cmd/tossctl `strategyRuntimeAttachment`)를
// 꽂는다. wrapper 는 정의상 non-nil 이라 화면의 `== nil` 판정을 그대로 두면 「이 배포는 전략 화면을
// 안 쓴다」(dormant)가 「runtime projection을 읽지 못했다」(도달 불가)로 회귀한다 — a109 freeze P1-4 와
// 같은 병의 반대편. 그래서 두 소비자(페이지·설정 요약)가 한 벌 판정
// (`strategyprojection.StrategyRuntimeAbsent`)을 쓰는지 여기서 잰다.
//
// 가짜 reader 는 wrapper 와 같은 모양이다: `Read` 와 `StrategyRuntimeConfigured` 를 둘 다 가진다.
// seam(`MultiMarketStrategyRuntimeReader`)은 여전히 `Read` 하나다(static_test 강제) — presence 는
// 판정 함수 안의 타입 단언으로만 묻는다.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a115PresenceReader 는 부재 신호를 스스로 말하는 reader 다. 몇 번 읽혔는지·몇 번 물었는지 센다.
type a115PresenceReader struct {
	mu         sync.Mutex
	configured bool
	snapshot   strategyprojection.Snapshot
	err        error
	reads      int
	asked      int
}

func (r *a115PresenceReader) Read(context.Context) (strategyprojection.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
	return r.snapshot, r.err
}

func (r *a115PresenceReader) StrategyRuntimeConfigured() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.asked++
	return r.configured
}

func (r *a115PresenceReader) counts() (reads, asked int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reads, r.asked
}

// a115CountingReader 는 부재 신호를 말하지 않는 reader 다(오늘의 스텁·raw client 모양).
type a115CountingReader struct {
	mu       sync.Mutex
	snapshot strategyprojection.Snapshot
	reads    int
}

func (r *a115CountingReader) Read(context.Context) (strategyprojection.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
	return r.snapshot, nil
}

// TestAnUnconfiguredWrapperStillRendersDormant 은 spec 「미구성은 그대로 미구성이다」의 화면 쪽이다.
//
// 부재 fixture 가 **유효한 live 스냅샷**을 드는 이유(a109 원장 M22 의 교훈): 판정을 nil 검사로
// 되돌리면 이 reader 는 non-nil 이라 Read 가 불리고, 스냅샷이 유효하니 화면이 **live 값**을 그린다.
// 빈 스냅샷을 들었으면 Validate 실패가 「읽지 못했다」를 그려 다른 이유로 빨개졌을 것이다 — 여기서는
// 부재 판정 하나만이 dormant 를 만들 수 있다.
func TestAnUnconfiguredWrapperStillRendersDormant(t *testing.T) {
	reader := &a115PresenceReader{configured: false, snapshot: consoleProjectionPair(t)}
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = reader })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"runtime endpoint 미기동", "NOT_CONFIGURED"} {
		if !strings.Contains(page, want) {
			t.Errorf("부재 wrapper 의 화면에 %q 가 없다 — dormant 가 사라졌다", want)
		}
	}
	for _, unwanted := range []string{"runtime projection을 읽지 못했다", "evidence-KR", "RUNTIME_UNAVAILABLE"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("부재 wrapper 의 화면에 %q 가 있다 — 부재가 도달 불가·live 로 오귀속됐다", unwanted)
		}
	}
	reads, asked := reader.counts()
	if reads != 0 {
		t.Errorf("부재 wrapper 를 %d번 Read 했다 — 부재를 먼저 묻지 않았다", reads)
	}
	// 한 번 물어 두 자리(Unwired·Read 여부)에 쓴다 — presence 질문은 재부착 시도를 깨우는 부작용이 있다.
	if asked != 1 {
		t.Errorf("화면 한 번에 부재 신호를 %d번 물었다, want 1", asked)
	}
}

// TestAConfiguredButUnreachableWrapperRendersUnavailable 은 spec 「descriptor 가 남은 다운」의 화면 쪽이다.
// 구성됐으나(신호 true) 못 읽으면 도달 불가다 — NOT_CONFIGURED 가 아니다.
func TestAConfiguredButUnreachableWrapperRendersUnavailable(t *testing.T) {
	reader := &a115PresenceReader{configured: true,
		err: errors.New("strategy runtime projection unavailable: socket has no listener")}
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = reader })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"runtime projection을 읽지 못했다", "RUNTIME_UNAVAILABLE"} {
		if !strings.Contains(page, want) {
			t.Errorf("도달 불가 wrapper 의 화면에 %q 가 없다", want)
		}
	}
	for _, unwanted := range []string{"runtime endpoint 미기동", "NOT_CONFIGURED", "socket has no listener"} {
		if strings.Contains(page, unwanted) {
			t.Errorf("도달 불가 wrapper 의 화면에 %q 가 있다 — 도달 불가가 미구성으로 접혔거나 원문 오류가 샜다", unwanted)
		}
	}
	if reads, asked := reader.counts(); reads != 1 || asked != 1 {
		t.Errorf("reads=%d asked=%d, want 1/1", reads, asked)
	}
}

// TestTheSummaryAsksThePresenceSignalToo 는 설정 화면 요약이 **같은 판정**을 쓰는지다(정정의 단위는 값).
func TestTheSummaryAsksThePresenceSignalToo(t *testing.T) {
	absent := &a115PresenceReader{configured: false, snapshot: consoleProjectionPair(t)}
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = absent })
	got := h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	if got != "KR OFF/UNKNOWN · US OFF/UNKNOWN — dormant 미배선" {
		t.Errorf("부재 wrapper 의 요약 = %q — 페이지와 다른 판정이다", got)
	}
	if reads, asked := absent.counts(); reads != 0 || asked != 1 {
		t.Errorf("부재 요약 reads=%d asked=%d, want 0/1", reads, asked)
	}

	unreachable := &a115PresenceReader{configured: true, err: errors.New("dial")}
	h = newHarness(t, func(options *Options) { options.StrategyRuntime = unreachable })
	got = h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	if got != "읽지 못함 — 전략 lane 판독이 유효하지 않다" {
		t.Errorf("도달 불가 wrapper 의 요약 = %q", got)
	}
}

// TestASignallessReaderIsStillWired 는 오늘의 뜻 그대로다: 신호를 말하지 않는 non-nil reader 는 「있다」.
// 이것이 깨지면 기존 스텁·raw client 가 전부 dormant 로 떨어진다.
func TestASignallessReaderIsStillWired(t *testing.T) {
	reader := &a115CountingReader{snapshot: consoleProjectionPair(t)}
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = reader })
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	if !strings.Contains(page, "evidence-KR") || strings.Contains(page, "runtime endpoint 미기동") {
		t.Error("신호 없는 reader 가 wired 로 그려지지 않았다")
	}
	_ = h.strategyRuntimeSummary(httptest.NewRequest(http.MethodGet, "/settings", nil))
	reader.mu.Lock()
	defer reader.mu.Unlock()
	if reader.reads != 2 {
		t.Errorf("신호 없는 reader 를 %d번 읽었다, want 2(페이지·요약)", reader.reads)
	}
}
