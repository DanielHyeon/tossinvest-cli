package main

// a108_the_aggregate_snapshot_outlives_the_strategy_read_test.go — 6.8 ② (design D4-2).
//
// `httpAPIReader.Snapshot` 은 일곱 개의 조회를 모아 한 JSON 으로 만든다. 마지막이
// 전략 projection 이고, 그 `Read` 가 실패하면 **집계 전체**가 오류로 죽었다
// (`httpAPIReader.Snapshot` 의 마지막 조회). 그러면 엔진이 재시작하는 몇 초 동안 — 또는 잔재 회수로
// projection 이 잠깐 사라진 동안 — 운영자는 포지션도 주문도 못 본다.
//
// 화면 하나가 사라지는 것과 화면 전체가 사라지는 것은 다른 사고다. 콘솔은 이미 그렇게
// 판정한다(`internal/console` 의 `buildMultiMarketStrategyRuntimePage`): reader 가 **없으면**
// dormant(NOT_CONFIGURED), reader 는 있는데 **읽지 못하면** unavailable(RUNTIME_UNAVAILABLE).
// 두 사실을 한 값으로 접지 않는 것도 그 선례를 따른다.
//
// 새 파일인 이유는 겹2·겹3 테스트와 같다: 기존 파일에 덧붙이면 check_analysis 가
// 이웃 함수까지 「수정됨」으로 본다.

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

var a108SnapshotNow = time.Date(2026, 8, 14, 3, 0, 0, 0, time.UTC)

// a108StrategyRuntimeFixture 는 살아 있다가 죽는 projection 이다.
type a108StrategyRuntimeFixture struct {
	snapshot strategyprojection.Snapshot
	err      error
	calls    int
}

func (f *a108StrategyRuntimeFixture) Read(context.Context) (strategyprojection.Snapshot, error) {
	f.calls++
	if f.err != nil {
		return strategyprojection.Snapshot{}, f.err
	}
	return strategyprojection.Clone(f.snapshot), nil
}

// a108SignalsFixture 는 아무 후보도 없는 정상 판독이다.
type a108SignalsFixture struct{}

func (a108SignalsFixture) Signals(context.Context) (console.SignalsReading, error) {
	return console.SignalsReading{At: a108SnapshotNow}, nil
}

// a108PerformanceFixture 는 빈 대시보드다. 귀속 행은 「이 배포에 없음」으로 답한다 —
// `Performance` 가 그 오류만 건너뛰기 때문이다(`httpAPIReader.Performance`).
type a108PerformanceFixture struct{}

func (a108PerformanceFixture) Dashboard(context.Context, performance.Query) (performance.DashboardView, error) {
	return performance.DashboardView{}, nil
}

func (a108PerformanceFixture) AttributionRows(context.Context, string, performance.AttributionQuery, int) (
	[]performance.Attribution, error) {
	return nil, performance.ErrAttributionUnavailable
}

// a108SnapshotReader 는 일곱 조회가 전부 성공하는 production adapter 다.
//
// 여기서 재는 것은 「전략 하나가 죽으면 나머지 여섯이 어떻게 되는가」이므로, 나머지
// 여섯은 전부 성공해야 한다. 하나라도 실패하면 이 테스트는 잘못된 이유로 통과한다.
func a108SnapshotReader(t *testing.T, strategy *a108StrategyRuntimeFixture) *httpAPIReader {
	t.Helper()
	ctx := context.Background()
	registry, err := optimization.CoreRegistry(ctx)
	if err != nil {
		t.Fatal(err)
	}
	store, err := optimization.Open(ctx, optimization.Options{
		Path: filepath.Join(t.TempDir(), optimization.DatabaseFileName), Registry: registry, Actor: "a108-fixture",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	return &httpAPIReader{
		now:          func() time.Time { return a108SnapshotNow },
		engineMarker: filepath.Join(t.TempDir(), "engine-active.json"),
		holdings: &httpAPIHoldingsFixture{rows: []domain.Position{{
			Symbol: "005930", Name: "삼성전자", MarketCode: "KR", Quantity: 3, AveragePrice: 70000, CurrentPrice: 71000,
		}}},
		orders: &httpAPIOrdersFixture{reading: console.OrdersReading{
			AccountRef: "raw-account-1234", Open: []console.OrderRecord{{
				ID: "order-1", Symbol: "005930", Market: "KR", Side: "BUY", Status: "OPEN", Quantity: "1",
			}},
		}},
		signals:           a108SignalsFixture{},
		accountRef:        func() (string, error) { return "raw-account-1234", nil },
		performance:       a108PerformanceFixture{},
		optimization:      store,
		adoptionDesired:   func() (config.Adoption, string, error) { return config.Adoption{}, "", nil },
		managementRuntime: httpAPIManagementRuntimeFixture{},
		strategyRuntime:   strategy,
	}
}

// a108AggregateSnapshot 은 집계 JSON 중 이 테스트가 보는 부분이다.
type a108AggregateSnapshot struct {
	SchemaVersion string `json:"schemaVersion"`
	Engine        struct {
		Status string `json:"status"`
	} `json:"engine"`
	Positions struct {
		Items []struct {
			Symbol string `json:"symbol"`
		} `json:"items"`
	} `json:"positions"`
	Orders struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	} `json:"orders"`
	StrategyRuntime strategyprojection.Snapshot `json:"strategyRuntime"`
}

func a108ReadAggregate(t *testing.T, reader *httpAPIReader) a108AggregateSnapshot {
	t.Helper()
	body, err := reader.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("집계 스냅샷이 실패했다: %v", err)
	}
	var out a108AggregateSnapshot
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("집계 스냅샷 파싱: %v\n%s", err, body)
	}
	return out
}

func a108MarketRefusal(t *testing.T, snapshot strategyprojection.Snapshot,
	market strategyprojection.Market) strategyprojection.RefusalCode {
	t.Helper()
	projection, ok := snapshot.Markets[market]
	if !ok {
		t.Fatalf("전략 스냅샷에 %s 시장이 없다: %+v", market, snapshot.Markets)
	}
	if projection.Error == nil {
		return strategyprojection.RefusalNone
	}
	return projection.Error.Code
}

// TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding 은 6.8 ② 그 자체다.
//
// 살아 있던 projection 을 먼저 읽어 「대조군」을 만든다. 대조군이 없으면
// 「항상 unavailable 을 준다」는 구현이 이 테스트를 통과한다.
func TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding(t *testing.T) {
	live := strategyprojection.WithMarketFailure(
		strategyprojection.DormantSnapshot(a108SnapshotNow),
		strategyprojection.MarketKR, strategyprojection.RefusalEvidenceStale, a108SnapshotNow)
	strategy := &a108StrategyRuntimeFixture{snapshot: live}
	reader := a108SnapshotReader(t, strategy)

	alive := a108ReadAggregate(t, reader)
	if got := a108MarketRefusal(t, alive.StrategyRuntime, strategyprojection.MarketKR); got !=
		strategyprojection.RefusalEvidenceStale {
		t.Fatalf("살아 있는 projection 의 KR 판정 = %q, want %q — 집계가 실제 읽기를 통과시키지 않는다",
			got, strategyprojection.RefusalEvidenceStale)
	}
	if len(alive.Positions.Items) != 1 || len(alive.Orders.Items) != 1 {
		t.Fatalf("대조군 집계가 이미 비어 있다: positions=%d orders=%d",
			len(alive.Positions.Items), len(alive.Orders.Items))
	}

	// 엔진이 재시작한다. projection socket 은 사라지고 읽기는 실패한다.
	strategy.err = errors.New("strategy projection runtime: socket is invalid")

	dead := a108ReadAggregate(t, reader)
	if len(dead.Positions.Items) != 1 || dead.Positions.Items[0].Symbol != "005930" {
		t.Errorf("전략 읽기 실패가 포지션 화면을 같이 죽였다: %+v", dead.Positions.Items)
	}
	if len(dead.Orders.Items) != 1 || dead.Orders.Items[0].ID != "order-1" {
		t.Errorf("전략 읽기 실패가 주문 화면을 같이 죽였다: %+v", dead.Orders.Items)
	}
	if dead.Engine.Status == "" || dead.SchemaVersion == "" {
		t.Errorf("집계 봉투가 무너졌다: %+v", dead)
	}
	for _, market := range []strategyprojection.Market{strategyprojection.MarketKR, strategyprojection.MarketUS} {
		if got := a108MarketRefusal(t, dead.StrategyRuntime, market); got !=
			strategyprojection.RefusalRuntimeUnavailable {
			t.Errorf("죽은 projection 의 %s 판정 = %q, want %q — 실패를 「설정 안 됨」으로 접으면 "+
				"운영자는 엔진이 죽은 것과 기능을 안 켠 것을 구별할 수 없다",
				market, got, strategyprojection.RefusalRuntimeUnavailable)
		}
	}
	if strategy.calls != 2 {
		t.Errorf("전략 읽기 호출 %d회, want 2 — 흡수가 읽기 자체를 건너뛰었다", strategy.calls)
	}
}

// TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable 은 위 판정의 반대편이다.
//
// reader 가 아예 없는 배포(전략 화면 미구성)는 「읽지 못했다」가 아니라 「설정되지
// 않았다」다. 이 구분이 없으면 두 사실이 한 화면으로 접힌다.
func TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable(t *testing.T) {
	reader := a108SnapshotReader(t, nil)
	reader.strategyRuntime = nil

	snapshot := a108ReadAggregate(t, reader)
	for _, market := range []strategyprojection.Market{strategyprojection.MarketKR, strategyprojection.MarketUS} {
		if got := a108MarketRefusal(t, snapshot.StrategyRuntime, market); got !=
			strategyprojection.RefusalNotConfigured {
			t.Errorf("reader 부재의 %s 판정 = %q, want %q", market, got, strategyprojection.RefusalNotConfigured)
		}
	}
}
