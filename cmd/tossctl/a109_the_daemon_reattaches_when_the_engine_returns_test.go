//go:build unix

package main

// a109_the_daemon_reattaches_when_the_engine_returns_test.go — T2 §2.3 (design D4).
//
// a108 이 남긴 기본값을 닫는다: 조회 데몬은 전략 endpoint 를 **부팅 1회**만 해결하고
// 그 값에 영원히 머문다. 결과는 두 가지다.
//
//	① 냉부팅 순서 — httpapi 가 엔진보다 먼저 뜨면 descriptor 가 없어 dormant 로 굳고,
//	   엔진이 나중에 발행해도 데몬을 **손으로 재시작**할 때까지 전략 화면이 없다.
//	② 가동 중 재시작 — live client 를 잡은 뒤 엔진이 재시작하면 socket 도 토큰도 새
//	   것이라 그 client 는 **영구 실패**한다. 부착돼 있으니 재부착 대상도 아니다
//	   (freeze P0-1 — 원 설계가 nil·sentinel 만 감쌌다면 이쪽이 통째로 빠졌다).
//
// 두 시나리오를 **둘 다** 잰다. 하나만 재면 나머지 하나가 기전 없는 SHALL 로 남는다.
//
// # 왜 실패 주입이 아니라 진짜 endpoint 인가
//
// 여기서 재는 것은 「엔진이 돌아왔다」를 데몬이 알아채는가다. 그 사실은 디스크의
// descriptor·socket 으로만 존재하므로 주입한 seam 으로는 애초에 잴 수 없다.
// a108 의 겹3 이 같은 이유로 seam 을 쓰지 않았다.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
)

// a109Projection 은 엔진이 내보내는 전략 projection 이다. 시장 하나에 뚜렷한 판정을
// 실어 두면 「붙었다」를 화면 값으로 구별할 수 있다.
type a109Projection struct{ snapshot strategyprojection.Snapshot }

func (p a109Projection) Read(context.Context) (strategyprojection.Snapshot, error) {
	return strategyprojection.Clone(p.snapshot), nil
}

func a109LiveProjection() a109Projection {
	at := time.Date(2026, 8, 15, 6, 0, 0, 0, time.UTC)
	return a109Projection{snapshot: strategyprojection.WithMarketFailure(
		strategyprojection.DormantSnapshot(at), strategyprojection.MarketKR,
		strategyprojection.RefusalEvidenceStale, at)}
}

// a109StartEngineProjection 은 엔진 쪽 endpoint 를 실제로 세운다.
func a109StartEngineProjection(t *testing.T, dir string) *strategyprojectionrpc.Server {
	t.Helper()
	server, err := strategyprojectionrpc.Start(dir, a109LiveProjection())
	if err != nil {
		t.Fatalf("엔진 projection endpoint 기동: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return server
}

// a109Screen 은 데몬이 지금 그리는 KR 화면의 판정이다.
//
// 집계 스냅샷을 통과시키는 이유는 그것이 **운영자가 실제로 보는 값**이기 때문이다 —
// reader 를 직접 Read 하면 「부재」와 「못 읽음」이 같은 오류로 접혀 구분이 사라진다.
// reader 를 한 번만 만들어 재사용하는 이유는 폴링 때문이다: 매번 새로 만들면 폴링
// 하나가 optimization DB 를 수십 번 여는 일이 된다.
func a109Screen(t *testing.T, reader *httpAPIReader) strategyprojection.RefusalCode {
	t.Helper()
	return a108MarketRefusal(t, a108ReadAggregate(t, reader).StrategyRuntime,
		strategyprojection.MarketKR)
}

// a109ScreenReader 는 그 화면을 그리는 집계 reader 하나다.
func a109ScreenReader(t *testing.T, runtime httpapi.StrategyRuntimeReader) *httpAPIReader {
	t.Helper()
	reader := a108SnapshotReader(t, nil)
	reader.strategyRuntime = runtime
	return reader
}

// a109WaitForScreen 은 재부착이 rate limit 창 안에서 일어나기를 기다린다.
//
// 폴링인 이유는 시도가 **백그라운드**이기 때문이다 — 요청 경로는 dial 하지 않으므로
// 첫 요청이 곧바로 회복된 값을 주는 것은 오히려 계약 위반이다.
func a109WaitForScreen(t *testing.T, reader *httpAPIReader, want strategyprojection.RefusalCode,
	within time.Duration) strategyprojection.RefusalCode {
	t.Helper()
	deadline := time.Now().Add(within)
	got := a109Screen(t, reader)
	for got != want && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		got = a109Screen(t, reader)
	}
	return got
}

// a109ReattachWindow 는 이 테스트가 회복을 기다리는 상한이다.
const a109ReattachWindow = 5 * time.Second

// a109FastRedial 은 rate limit 을 이 테스트 동안만 좁힌다.
//
// 30초를 기다리는 테스트는 아무도 돌리지 않고, 안 돌리는 테스트는 없는 테스트다.
// **운영 기본값 자체**는 아래 TestTheProductionRedialIntervalIsThirtySeconds 가 핀한다 —
// 주입 가능성을 「기본값을 안 재도 되는 이유」로 쓰면 그것이 구멍이다.
func a109FastRedial(t *testing.T) {
	t.Helper()
	previous := strategyRuntimeRedialInterval
	strategyRuntimeRedialInterval = 20 * time.Millisecond
	t.Cleanup(func() { strategyRuntimeRedialInterval = previous })
}

// TestTheProductionRedialIntervalIsThirtySeconds 는 배포가 실제로 쓰는 값이다.
func TestTheProductionRedialIntervalIsThirtySeconds(t *testing.T) {
	if strategyRuntimeRedialInterval != 30*time.Second {
		t.Errorf("운영 재부착 간격 = %s, want 30s — Dial 본문의 200ms probe 때문에 "+
			"간격 없는 재시도는 read 마다 그 비용을 낸다", strategyRuntimeRedialInterval)
	}
}

// TestTheDaemonAttachesWhenTheEngineComesUpLater 는 시나리오 ①(냉부팅 순서)이다.
func TestTheDaemonAttachesWhenTheEngineComesUpLater(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a109FastRedial(t)
	errOut := &a109SyncWriter{}
	runtime := strategyRuntimeReaderFor(context.Background(), &rootOptions{configDir: dir}, errOut)
	reader := a109ScreenReader(t, runtime)

	// 엔진은 아직 없다 — 화면은 「이 배포는 전략 화면을 안 쓴다」로 뜬다.
	if got := a109Screen(t, reader); got != strategyprojection.RefusalNotConfigured {
		t.Fatalf("엔진 없는 기동의 화면 = %q, want %q", got, strategyprojection.RefusalNotConfigured)
	}

	// 엔진이 뜬다.
	a109StartEngineProjection(t, dir)

	if got := a109WaitForScreen(t, reader, strategyprojection.RefusalEvidenceStale,
		a109ReattachWindow); got != strategyprojection.RefusalEvidenceStale {
		t.Fatalf("엔진이 돌아온 뒤 %s 안에 화면이 회복되지 않았다 (판정 %q) — "+
			"데몬을 손으로 재시작해야 전략 화면이 돌아온다\n%s",
			a109ReattachWindow, got, errOut.String())
	}
}

// TestTheDaemonReattachesAfterTheEngineRestarts 는 시나리오 ②(가동 중 재시작)이고
// freeze P0-1 이 연 구멍이다.
//
// 여기서 재는 것은 **live 부착 뒤의 실패**다. 엔진을 닫고 다시 세우면 socket 도 토큰도
// 새 것이라 잡고 있던 client 는 영구 실패한다 — 부재도 아니고 sentinel 도 아니므로,
// nil·sentinel 만 감싸는 wrapper 로는 이 경로가 통째로 빠진다.
func TestTheDaemonReattachesAfterTheEngineRestarts(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a109FastRedial(t)
	first := a109StartEngineProjection(t, dir)

	errOut := &a109SyncWriter{}
	runtime := strategyRuntimeReaderFor(context.Background(), &rootOptions{configDir: dir}, errOut)
	reader := a109ScreenReader(t, runtime)
	if got := a109Screen(t, reader); got != strategyprojection.RefusalEvidenceStale {
		t.Fatalf("live 부착의 화면 = %q, want %q — 대조군이 이미 틀렸다",
			got, strategyprojection.RefusalEvidenceStale)
	}

	// 엔진이 재시작한다: 옛 endpoint 가 사라지고 새 것이 선다.
	if err := first.Close(); err != nil {
		t.Fatalf("첫 엔진 endpoint Close: %v", err)
	}
	if got := a109Screen(t, reader); got != strategyprojection.RefusalRuntimeUnavailable {
		t.Fatalf("엔진이 사라진 직후의 화면 = %q, want %q — 「없다」와 「못 읽는다」가 접혔다",
			got, strategyprojection.RefusalRuntimeUnavailable)
	}
	a109StartEngineProjection(t, dir)

	if got := a109WaitForScreen(t, reader, strategyprojection.RefusalEvidenceStale,
		a109ReattachWindow); got != strategyprojection.RefusalEvidenceStale {
		t.Fatalf("엔진 재시작 뒤 %s 안에 화면이 회복되지 않았다 (판정 %q) — "+
			"부팅에서 잡은 client 는 새 socket·새 토큰을 모른다\n%s",
			a109ReattachWindow, got, errOut.String())
	}
}
