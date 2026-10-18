package engine

// a157 A2 — 조립 감독자(`(*Context).NewPairedStrategyEntryProductionAssembly` 의 감독자)도 생산 감독자(`NewRefreshingPairedStrategyEntrySupervisor`)
// 와 같은 일일 손실 알림 · 건너뜀 알림 함수와 같은 진입 게이트를 받음(design D4 · D5, spec MODIFIED).
//
// 이 감독자의 worker 주기는 생산 권한 갱신 전체라 Run 으로는 기록자까지 닿지 않음(freeze 리뷰 Opus P2-2). 그래서 감독자 칸을 직접 보고,
// 패키지 내부 기록자를 그 감독자에 직접 불러 `obs` 이벤트가 `c.Log` 까지 가는지 잼. 생산 감독자를 같은 방법으로 대조군으로 둠.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// KST 10:00 — 계좌 일 키 2026-10-18(알림 dedup 이 일 키를 셀 수 있는 시각).
var a157AssemblyNow = time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)

// a157AssemblySupervisor 는 최소 Context 로 조립 감독자를 만듦(빈 Context 의 파도는 권한을 하나도 못 읽고 성공함 — a112 파도 시험과 같은 모양).
func a157AssemblySupervisor(t *testing.T, c *Context) *StrategyEntrySupervisor {
	t.Helper()
	assembly, err := c.NewPairedStrategyEntryProductionAssembly(context.Background(), clock.NewFake(a157AssemblyNow))
	if err != nil || assembly.Supervisor == nil {
		t.Fatalf("assembly: supervisor=%v err=%v", assembly.Supervisor, err)
	}
	return assembly.Supervisor
}

// a157ProductionSupervisor 는 대조군 — 생산 감독자(진입 게이트가 있어야 만들어짐).
func a157ProductionSupervisor(t *testing.T, c *Context) *StrategyEntrySupervisor {
	t.Helper()
	supervisor, err := c.NewRefreshingPairedStrategyEntrySupervisor(clock.NewFake(a157AssemblyNow))
	if err != nil {
		t.Fatal(err)
	}
	return supervisor
}

func a157LoggedContext(buf *strings.Builder, gate *execgw.EntryGate) *Context {
	return &Context{Log: obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true}), Entry: gate}
}

// 알림 칸 둘이 채워짐(로그가 있는 엔진). 로그가 없으면 두 칸 모두 비고(알림 함수가 nil 을 돌려줌) 조립은 그대로 성공.
func TestA157TheAssemblySupervisorCarriesBothNotices(t *testing.T) {
	var buf strings.Builder
	gate := execgw.NewEntryGate(clock.NewFake(a157AssemblyNow), nil)
	for name, supervisor := range map[string]*StrategyEntrySupervisor{
		"assembly":            a157AssemblySupervisor(t, a157LoggedContext(&buf, gate)),
		"production(control)": a157ProductionSupervisor(t, a157LoggedContext(&buf, gate)),
	} {
		if supervisor.dailyLossNotice == nil || supervisor.entrySkipNotice == nil {
			t.Errorf("%s: dailyLossNotice set=%v entrySkipNotice set=%v — want both", name,
				supervisor.dailyLossNotice != nil, supervisor.entrySkipNotice != nil)
		}
	}
	quiet := a157AssemblySupervisor(t, &Context{})
	if quiet.dailyLossNotice != nil || quiet.entrySkipNotice != nil {
		t.Fatal("an engine without a logger must leave both notice seams empty")
	}
}

// 기록자를 감독자에 직접 부르면 경고 줄이 c.Log 까지 감 — 측정 불가 한 줄 · 건너뜀 한 줄.
func TestA157TheAssemblyNoticesReachTheLog(t *testing.T) {
	for _, kind := range []string{"assembly", "production(control)"} {
		var buf strings.Builder
		c := a157LoggedContext(&buf, execgw.NewEntryGate(clock.NewFake(a157AssemblyNow), nil))
		supervisor := a157AssemblySupervisor(t, c)
		if kind != "assembly" {
			supervisor = a157ProductionSupervisor(t, c)
		}
		worker := supervisor.workers[StrategyMarketKR]
		if worker == nil {
			t.Fatalf("%s: no KR worker runtime", kind)
		}
		supervisor.recordDailyLossUnmeasurable(worker, errors.New("dailyloss: today's realised loss cannot be measured: KR 005930 pos-1a2b3c4d"))
		supervisor.recordGuardianCapZeroSkip(worker, &strategyEntrySkip{market: strategyrouter.MarketKR, symbol: "005930",
			entry: "70000", stop: "69000", fx: "unavailable", cause: errors.New("cap zero")})
		lines := buf.String()
		for _, want := range []string{string(obs.EventDailyLossUnmeasurable), string(obs.EventStrategyEntrySkipped), `"account_day":"2026-10-18"`} {
			if !strings.Contains(lines, want) {
				t.Errorf("%s: log %q lacks %q", kind, lines, want)
			}
		}
	}
}

// 진입 게이트: 엔진에 게이트가 없으면 감독자의 게이트 칸은 **nil 인터페이스**(타입 있는 nil 이 아님 — design F10)라 중앙 무결성 고장은
// 오늘처럼 프로세스 전체 fail-closed 로 올라감. 게이트가 있으면 같은 게이트를 받아 그 고장이 신규 진입을 닫음(Block).
func TestA157TheAssemblySupervisorEntryGate(t *testing.T) {
	without := a157AssemblySupervisor(t, &Context{})
	if without.entry != nil {
		t.Fatalf("entry = %#v — want a nil interface when the engine has no entry gate (a typed nil would panic in Block)", without.entry)
	}
	if without.blockEntryOnCentralIntegrity(without.workers[StrategyMarketKR]) {
		t.Fatal("without an entry gate the central fault must not be absorbed as an entry block")
	}

	for _, kind := range []string{"assembly", "production(control)"} {
		gate := execgw.NewEntryGate(clock.NewFake(a157AssemblyNow), nil)
		c := &Context{Entry: gate}
		var supervisor *StrategyEntrySupervisor
		if kind == "assembly" {
			supervisor = a157AssemblySupervisor(t, c)
		} else {
			supervisor = a157ProductionSupervisor(t, c)
		}
		if blocker, ok := supervisor.entry.(*execgw.EntryGate); !ok || blocker != gate {
			t.Fatalf("%s: entry = %#v — want the engine's own entry gate", kind, supervisor.entry)
		}
		if !supervisor.blockEntryOnCentralIntegrity(supervisor.workers[StrategyMarketKR]) {
			t.Fatalf("%s: the central fault was not turned into an entry block", kind)
		}
		if _, blocked := gate.Blocks()[execgw.ReasonStrategyCentralIntegrity]; !blocked {
			t.Fatalf("%s: blocks=%v — want the strategy central integrity latch", kind, gate.Blocks())
		}
	}
}
