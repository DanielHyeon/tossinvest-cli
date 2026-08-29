package engine

// a112 L5 — 결정 54의 「읽을 수 없는 숫자는 적을 수 없다」.
//
// 스케줄러를 깨우려면 사람이 활성화 매니페스트에 config digest 와 build digest 를
// 적어야 한다. 엔진은 두 값을 계산해서 스냅샷 내부에 채워 왔지만 어떤 운영 표면으로도
// 내보내지 않았다. 그래서 운영자가 적을 수 없었고, 적지 못하면 스케줄러는 안 깨어난다.
//
// 이 파일이 재는 것은 **읽는 경로**다: 엔진의 projection reader 가 돌려주는 스냅샷에
// 이 프로세스가 실제로 쓰는 값이 그대로 들어 있는가.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest 는 결정 54가 요구한
// 증명 의무 그대로다: **투영된 값 == strategyRuntimeConfigDigest()**.
//
// 그리고 일부러 dormant(두 시장 다 UNKNOWN) 스냅샷으로 잰다. 운영자가 이 숫자를
// 가장 필요로 하는 순간이 바로 아무것도 활성화되지 않은 그 상태이기 때문이다 —
// 시장 레코드 안에 뒀다면 정확히 그때 사라졌을 것이다.
func TestStrategyRuntimeReadExposesThisProcessConfigAndBuildDigest(t *testing.T) {
	observed := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(observed))
	if err != nil {
		t.Fatal(err)
	}
	ctx := &Context{strategyProjection: store}

	snapshot, err := ctx.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := strategyprojection.Validate(snapshot); err != nil {
		t.Fatalf("노출된 스냅샷이 스스로의 계약을 깬다: %v", err)
	}

	if snapshot.Runtime.ConfigDigest == nil || snapshot.Runtime.BuildDigest == nil {
		t.Fatalf("운영자가 적어야 하는 숫자가 여전히 밖으로 나오지 않는다: %+v", snapshot.Runtime)
	}

	// ① 형식 — **생산자의 변환을 되풀이하지 않는 단언**이다.
	//
	// 기대값을 production 이 쓰는 변환으로 만들면 그 변환이 틀려도 둘이 같이 움직여
	// 테스트가 통과한다. 그래서 여기서는 소비자가 요구하는 모양을 직접 적는다:
	// `scheduler.validateProductionActivationManifest` 는 매니페스트의
	// `ConfigVersion`/`BuildDigest` 를 binding 값과 **정확히 문자열 비교**하고,
	// binding 쪽은 `sha256:` 접두사를 달고 있다
	// (`strategy_schedule_authority.go` 의 prepare → `scheduler.CurrentBinding`).
	for name, got := range map[string]string{"config": *snapshot.Runtime.ConfigDigest, "build": *snapshot.Runtime.BuildDigest} {
		if !strings.HasPrefix(got, "sha256:") {
			t.Fatalf("%s digest = %q — 접두사가 없다. 이대로 매니페스트에 적으면 엔진이 거절한다", name, got)
		}
		body := strings.TrimPrefix(got, "sha256:")
		if len(body) != 64 || strings.ToLower(body) != body || strings.Trim(body, "0123456789abcdef") != "" {
			t.Fatalf("%s digest = %q — 소문자 64자리 hex 가 아니다", name, got)
		}
	}

	// ② 신원 — 이 프로세스의 값인가.
	if *snapshot.Runtime.ConfigDigest != strategyRuntimeConfigDigest() {
		t.Fatalf("config digest = %q, want %q", *snapshot.Runtime.ConfigDigest, strategyRuntimeConfigDigest())
	}
	if *snapshot.Runtime.BuildDigest != strategyRuntimeBuildDigest() {
		t.Fatalf("build digest = %q, want %q", *snapshot.Runtime.BuildDigest, strategyRuntimeBuildDigest())
	}

	// 두 시장이 여전히 정직한 UNKNOWN 이어야 한다. 이 change 는 관측을 하나 더
	// 내보낼 뿐, 없는 준비 상태를 만들어 내지 않는다.
	for _, market := range []strategyprojection.Market{strategyprojection.MarketKR, strategyprojection.MarketUS} {
		if item := snapshot.Markets[market]; item.Status != strategyprojection.StatusUnknown ||
			item.Lane.Effective != strategyprojection.StateOff {
			t.Fatalf("%s 가 관측 노출만으로 상태를 얻었다: %+v", market, item)
		}
	}
}

// TestStrategyRuntimeReadWithoutAStoreStaysAnError 는 노출이 **부재를 가리지 않음**을
// 잡는다. store 가 없으면 지어낸 digest 가 아니라 지금까지처럼 오류다.
func TestStrategyRuntimeReadWithoutAStoreStaysAnError(t *testing.T) {
	ctx := &Context{}
	if _, err := ctx.Read(context.Background()); err == nil {
		t.Fatal("store 없는 Read 가 성공했다 — 없는 런타임의 digest 를 지어낼 수 있다")
	}
	var absent *Context
	if _, err := absent.Read(context.Background()); err == nil {
		t.Fatal("nil Context 의 Read 가 성공했다")
	}
}

// TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket 는 이 함수의 **덮어쓰기
// 구간**을 잰다.
//
// Read 는 identity 를 붙인 뒤 latch 된 시장을 RUNTIME_UNAVAILABLE 로 덮는다. 그 덮기가
// envelope 를 새로 만들면서 identity 를 떨어뜨리면, 무언가 잘못돼 운영자가 화면을 여는
// 바로 그 순간에 매니페스트용 숫자가 사라진다.
//
// 이 함수의 latch 구간에는 이 lot 전까지 어떤 테스트도 없었다.
func TestStrategyRuntimeReadKeepsTheIdentityWhileLatchingAMarket(t *testing.T) {
	observed := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	snapshot := strategyprojection.DormantSnapshot(observed)
	snapshot.Markets[strategyprojection.MarketKR] = currentMarketProjectionForTest(strategyprojection.MarketKR, observed)
	store, err := strategyprojection.NewStore(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	// KR 만 latch 된 supervisor. 이 테스트는 Snapshot 읽기만 쓴다.
	supervisor := &StrategyEntrySupervisor{workers: map[StrategyMarket]*strategyMarketRuntime{
		StrategyMarketKR: {effective: true, latched: true},
		StrategyMarketUS: {effective: false, latched: false},
	}}
	ctx := &Context{strategyProjection: store, strategySupervisor: supervisor}

	got, err := ctx.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := strategyprojection.Validate(got); err != nil {
		t.Fatal(err)
	}
	kr := got.Markets[strategyprojection.MarketKR]
	if kr.Status != strategyprojection.StatusUnknown || kr.Error == nil ||
		kr.Error.Code != strategyprojection.RefusalRuntimeUnavailable {
		t.Fatalf("latch 된 KR 이 덮이지 않았다: %+v", kr)
	}
	if got.Runtime.ConfigDigest == nil || *got.Runtime.ConfigDigest != strategyRuntimeConfigDigest() ||
		got.Runtime.BuildDigest == nil || *got.Runtime.BuildDigest != strategyRuntimeBuildDigest() {
		t.Fatalf("시장 하나가 latch 되자 프로세스 전체의 identity 가 사라졌다: %+v", got.Runtime)
	}
}

// TestStrategyRuntimeReadOnAFailedStoreInventsNothing 는 오류 경로가 identity 를 붙이기
// **전에** 끝남을 잡는다. 읽지 못한 것에 이 프로세스의 숫자를 얹으면 안 된다.
func TestStrategyRuntimeReadOnAFailedStoreInventsNothing(t *testing.T) {
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := (&Context{strategyProjection: store}).Read(cancelled)
	if err == nil {
		t.Fatal("취소된 컨텍스트의 Read 가 성공했다")
	}
	if got.Runtime.ConfigDigest != nil || got.Runtime.BuildDigest != nil {
		t.Fatalf("읽지 못한 스냅샷에 digest 가 붙었다: %+v", got.Runtime)
	}
}

// currentMarketProjectionForTest 는 CURRENT 시장 하나를 만든다. latch 덮어쓰기는
// CURRENT 인 시장에만 일어나므로, 그 가지를 재려면 CURRENT 가 필요하다.
func currentMarketProjectionForTest(market strategyprojection.Market, observed time.Time) strategyprojection.MarketProjection {
	laneID, laneVersion := "lane-"+string(market), "v1"
	evidenceID, evidenceDigest := "evidence-"+string(market), strings.Repeat("e", 64)
	campaignID, legID := "campaign-"+string(market), "1"
	bucket, policy := "SHORT", "risk-v1"
	calendarSource, calendarVersion := "official-open-api-"+string(market), "2026.08"
	activationDigest := strings.Repeat("a", 64)
	at := observed
	return strategyprojection.MarketProjection{Market: market, Status: strategyprojection.StatusCurrent,
		Lane: strategyprojection.LaneProjection{ID: &laneID, Version: &laneVersion,
			Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn},
		Evidence: strategyprojection.EvidenceProjection{ID: &evidenceID, Digest: &evidenceDigest,
			Freshness: strategyprojection.FreshnessCurrent},
		Campaign:    strategyprojection.CampaignProjection{ID: &campaignID, LegID: &legID},
		HorizonRisk: strategyprojection.HorizonRiskProjection{Bucket: &bucket, PolicyVersion: &policy, Status: strategyprojection.ComponentCurrent},
		Scheduler: strategyprojection.SchedulerProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
			CalendarSource: &calendarSource, CalendarVersion: &calendarVersion, CalendarFreshness: strategyprojection.FreshnessCurrent},
		Activation: strategyprojection.ActivationProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
			Digest: &activationDigest, Status: strategyprojection.ActivationConfigured},
		Protection:     strategyprojection.ProtectionProjection{Readiness: strategyprojection.ProtectionWired, Refusal: strategyprojection.RefusalNone},
		Reconciliation: strategyprojection.ReconciliationProjection{Status: strategyprojection.ReconciliationHealthy, Refusal: strategyprojection.RefusalNone},
		FirstRefusal:   strategyprojection.RefusalNone, ObservedAt: &at}
}
