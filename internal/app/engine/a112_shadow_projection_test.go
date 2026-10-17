//go:build tossos_testseams

package engine

// a112 7.3.1 SHADOW 투영의 소거 규칙(브리프 v3.3 §5.1 · §13 R1).
//
// 관측은 판정 함수 하나(shadowObservationUsable — 파도 등식 ∧ 미만료 ∧ 나이 상한)로만 투영에 쓰인다:
//   (a)~(d) 다음 파도의 철회 · 만료 · 닫힘 · 단계 실패 → UNOBSERVED.
//   (e) 새 파도 없음 + 매니페스트 만료 시각 → UNOBSERVED(경계 −1ns 는 SHADOW).
//   (f) 새 파도 없음 + 실패 없음(supervisor 정지) + 관측 나이 MaxAge → UNOBSERVED(경계 −1ns 는 SHADOW).
//   의도된 간극: 건강한 주기에서 record 가 파도를 올린 뒤 다음 게시 전은 UNOBSERVED, 게시 뒤 SHADOW.
//   R1 두 시계: supervisor 가 그 시장의 평가를 abandon 으로 기록했으면 SHADOW 를 보이지 않는다.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyshadow"
)

func a112ShadowAt(t *testing.T) *a112ShadowWorld {
	t.Helper()
	world := newA112ShadowWorld(t, true, strategyrouter.FamilyActivation{})
	world.c.strategyRefresh.dispatch = nil
	if err := world.cycle(t, StrategyMarketKR); err != nil || a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
		t.Fatalf("arrangement: no SHADOW at the first wave (err=%v)", err)
	}
	return world
}

func (world *a112ShadowWorld) withConfig(config strategyshadow.Config) {
	batch := world.c.strategyRefresh.shadow.kr
	batch.config = config
	world.c.strategyRefresh.shadow = strategyShadowPair{kr: batch}
}

func TestTheNextWaveClearsAnUnusableShadow(t *testing.T) {
	t.Run("(a) revoked", func(t *testing.T) {
		world := a112ShadowAt(t)
		// 같은 결속의 폐기 매니페스트를 새 자리에 쓰고 핀을 그 바이트로 옮긴다(재발급 절차의 「폐기」 모양).
		config := a112ShadowManifestConfig(t, world.clk.Now(), false, time.Hour)
		data, err := strategyshadow.EncodeProductionFamilyShadow(strategyshadow.Document{Market: strategyrouter.MarketKR, Generation: 2,
			RouteManifestDigest: config.RouteManifestDigest, CalibrationDigest: config.CalibrationDigest, CalendarVersion: config.CalendarVersion,
			RiskPolicyDigest: config.RiskPolicyDigest, BuildDigest: config.BuildDigest, Actor: "operator",
			ApprovedAt: world.clk.Now().Add(-2 * time.Hour), IssuedAt: world.clk.Now().Add(-time.Hour), ExpiresAt: world.clk.Now().Add(time.Hour),
			Revoked: true, Shadow: []strategyrouter.Family{strategyrouter.FamilyContinuation}})
		if err != nil {
			t.Fatal(err)
		}
		config.ManifestDigest = a112WriteShadowManifest(t, config.ConfigDir, data)
		world.withConfig(config)
		if err := world.cycle(t, StrategyMarketKR); err != nil {
			t.Fatal(err)
		}
		if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
			t.Fatalf("%d SHADOW lanes survive a revoked manifest on the next wave", n)
		}
	})
	t.Run("(b) expired", func(t *testing.T) {
		world := a112ShadowAt(t)
		world.withConfig(a112ShadowManifestConfig(t, world.clk.Now(), true, 0, strategyrouter.FamilyContinuation))
		if err := world.cycle(t, StrategyMarketKR); err != nil {
			t.Fatal(err)
		}
		if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
			t.Fatalf("%d SHADOW lanes survive an expired manifest on the next wave", n)
		}
	})
	t.Run("(c) market closed before coordination", func(t *testing.T) {
		world := a112ShadowAt(t)
		world.c.strategyRefresh.shadow = strategyShadowPair{} // 조정 앞 닫힘 = 부재 값
		if err := world.cycle(t, StrategyMarketKR); err != nil {
			t.Fatal(err)
		}
		if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
			t.Fatalf("%d SHADOW lanes survive a closed market", n)
		}
	})
	t.Run("(d) shadow step failure", func(t *testing.T) {
		world := a112ShadowAt(t)
		t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, func(context.Context, strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
			panic("a112 shadow step failure")
		}))
		if err := world.cycle(t, StrategyMarketKR); err != nil {
			t.Fatal(err)
		}
		if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
			t.Fatalf("%d SHADOW lanes survive a failed step", n)
		}
	})
}

// (e) · (f): 새 파도가 없을 때의 두 상한 — 만료와 나이. 경계는 등호 쪽이 거부, −1ns 가 통과.
func TestWithoutANewWaveExpiryAndAgeStillEndTheShadow(t *testing.T) {
	for _, tc := range []struct {
		name string
		at   func(world *a112ShadowWorld, observedAt, expiresAt time.Time) time.Time
	}{
		{"(e) manifest expiry", func(_ *a112ShadowWorld, _, expiresAt time.Time) time.Time { return expiresAt }},
		{"(f) observation age", func(_ *a112ShadowWorld, observedAt, _ time.Time) time.Time {
			return observedAt.Add(strategyShadowObservationMaxAge)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			world := newA112ShadowWorld(t, false, strategyrouter.FamilyActivation{})
			world.c.strategyRefresh.dispatch = nil
			expiresIn := 10 * time.Minute
			if tc.name == "(e) manifest expiry" {
				expiresIn = 30 * time.Second // 나이 상한(74s)보다 먼저 만료
			}
			world.withConfig(a112ShadowManifestConfig(t, world.clk.Now(), true, expiresIn, strategyrouter.FamilyContinuation))
			if err := world.cycle(t, StrategyMarketKR); err != nil || a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
				t.Fatalf("arrangement: no SHADOW (err=%v)", err)
			}
			var observedAt, expiresAt time.Time
			world.lanes.mu.RLock()
			for _, observation := range world.lanes.shadowObserved[StrategyMarketKR] {
				observedAt, expiresAt = observation.observedAt, observation.expiresAt
			}
			world.lanes.mu.RUnlock()
			edge := tc.at(world, observedAt, expiresAt)
			world.clk.Set(edge.Add(-time.Nanosecond))
			if a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
				t.Fatalf("1ns before the edge %s the SHADOW must still show", edge)
			}
			world.clk.Set(edge)
			if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
				t.Fatalf("at the edge %s %d SHADOW lanes still show", edge, n)
			}
		})
	}
}

// 나이 상한은 세 상수에서 유도한 값이고, 건강한 관측 간격(폴 + 주기 한도 + 단계 마감)을 나이만으로 거부하지 않는다.
func TestTheShadowAgeLimitIsDerivedAndRejectsNoHealthyGap(t *testing.T) {
	if want := 2 * (DefaultStrategyCycleLimit + MaximumStrategyCycleLimit + strategyShadowStepDeadline); strategyShadowObservationMaxAge != want {
		t.Fatalf("max age=%s, want 2×(poll+cycle limit+step deadline)=%s", strategyShadowObservationMaxAge, want)
	}
	healthyGap := DefaultStrategyCycleLimit + MaximumStrategyCycleLimit + strategyShadowStepDeadline
	observedAt := time.Date(2026, 10, 5, 1, 0, 0, 0, time.UTC)
	observation := strategyShadowObservation{wave: 3, observedAt: observedAt, expiresAt: observedAt.Add(time.Hour)}
	if !shadowObservationUsable(observedAt.Add(healthyGap), 3, observation) {
		t.Fatal("a same-wave observation one healthy gap old must not be rejected by age alone")
	}
	if shadowObservationUsable(observedAt, 4, observation) {
		t.Fatal("an observation from an older wave must not be used")
	}
}

// 의도된 간극: 건강한 주기에서 evaluate 가 파도를 올린 뒤 새 게시 전에는 UNOBSERVED, 게시 뒤 SHADOW(이전 파도 수용안 기각 — 철회 소거 약화).
func TestTheGapBetweenRecordAndPublishIsUnobservedByDesign(t *testing.T) {
	world := a112ShadowAt(t)
	release := make(chan struct{})
	t.Cleanup(setStrategyShadowLoadHookForTest(world.lanes, func(ctx context.Context, config strategyshadow.Config) (strategyshadow.FamilyShadow, error) {
		<-release
		return strategyshadow.LoadProductionFamilyShadow(context.Background(), config)
	}))
	if err := world.c.productionStrategyCycle(world.clk, StrategyMarketKR)(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("between record and publish %d lanes show SHADOW — the previous wave must not be used", n)
	}
	close(release)
	a112WaitShadowIdle(t, world.lanes, StrategyMarketKR)
	if a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
		t.Fatal("after the publish the new wave must show SHADOW")
	}
}

// R1 두 시계: 클로저가 성공으로 본 주기를 supervisor 가 abandon 으로 기록하면(경계 동시 준비) 그 시장의 SHADOW 는 보이지 않는다.
func TestAMarketWhoseEvaluationWasAbandonedShowsNoShadow(t *testing.T) {
	world := a112ShadowAt(t)
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{Clock: world.clk, CycleLimit: MaximumStrategyCycleLimit,
		Workers: []StrategyMarketWorker{
			{Market: StrategyMarketKR, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: func(context.Context) error { return nil }},
			{Market: StrategyMarketUS, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true, Cycle: func(context.Context) error { return nil }},
		}})
	if err != nil {
		t.Fatal(err)
	}
	world.c.strategyProjectionMu.Lock()
	world.c.strategySupervisor = supervisor
	world.c.strategyProjectionMu.Unlock()
	if a112ShadowCount(a112ShadowLanes(t, world.c)) == 0 {
		t.Fatal("control: before any abandon the SHADOW must show")
	}
	supervisor.markAbandoned(supervisor.workers[StrategyMarketKR])
	if n := a112ShadowCount(a112ShadowLanes(t, world.c)); n != 0 {
		t.Fatalf("after the supervisor recorded an abandoned KR evaluation %d lanes still show SHADOW", n)
	}
}

// a112WriteShadowManifest 는 바이트를 dir 에 0400 으로 쓰고 그 핀을 돌려준다.
func a112WriteShadowManifest(t *testing.T, dir string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, strategyshadow.ProductionFamilyShadowFileName(strategyrouter.MarketKR))
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
