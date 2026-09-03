package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// 이 파일은 태스크 5.3.3 의 나머지 절반이다: **레인 잠금을 원장에 남기고,
// 원장에서 되살리고, 증거가 있을 때만 연다.**
//
// 왜 여기인가. `internal/strategyworker` 는 자기 폐포에 쓸 수 있는 저널이 없다는
// 것을 `-deps` 로 증명한 패키지다. durable 기록에는 writer 가 필요하므로 writer 를
// 거기 두면 그 증명이 사라진다. `design.md:223` 이 "lane health/latch projection"
// 을 `internal/app/engine` 에 배정한 것과 같은 이유다.
//
// **복구 증거가 무엇인지는 골라 쓴 것이 아니라 읽은 것이다.**
// `scheduler.Activation` 은 "an opaque capability issued only after an exact
// manifest verification. Callers cannot forge it with a bool or public struct
// literal"(`internal/scheduler/desired.go:236`)이고, `Generation()` 이 돌려주는
// 값은 ed25519 서명 매니페스트 파일의 `manifest.Generation`
// (`internal/scheduler/production_activation.go:159`)이다. 즉 사람이 서명된 파일을
// 바꾸지 않으면 그 수는 오르지 않는다. 그것이 "복구 조건은 성공한 사이클이 아니라
// 증거"(`design.md:199`)를 값으로 만든 것이다.
//
// **이 로트가 없애는 운영 수단 하나를 적어 둔다.** 오늘 잠긴 레인을 여는 방법은
// 엔진 재시작이다. 이 뒤로는 재시작이 그것을 열지 않는다. 그것이 요구된 동작이지만,
// 재시작으로 고치던 사람에게는 수단이 하나 사라지는 것이므로 거절 문구가 무엇이
// 필요한지 말한다(`journal.ErrStrategyLaneLatchRecoveryEvidence`).

// strategyLaneLedger 는 이 런타임이 원장에게 요구하는 전부다.
//
// 인터페이스로 좁히는 이유는 대역을 만들기 위해서가 아니다(시험도 실제 저널을
// 쓴다). 이 타입이 `*journal.Journal` 을 통째로 들면 레인 관측자가 주문 경로의
// 모든 메서드를 손에 쥐게 되고, "관측자는 기록만 한다"가 주석이 된다.
type strategyLaneLedger interface {
	OpenStrategyLaneLatches(ctx context.Context, accountRef string) ([]journal.StrategyLaneLatch, error)
	RecordStrategyLaneLatch(ctx context.Context, latch journal.StrategyLaneLatch) (journal.StrategyLaneLatch, error)
	RecoverStrategyLaneLatch(ctx context.Context, seq int64, activationGeneration uint64) error
}

// strategyLaneLatchRestores 는 원장 기록을 레인이 받는 값으로 옮긴다.
func strategyLaneLatchRestores(latches []journal.StrategyLaneLatch) []strategyworker.LatchRestore {
	restores := make([]strategyworker.LatchRestore, 0, len(latches))
	for _, latch := range latches {
		restores = append(restores, strategyworker.LatchRestore{
			Key: strategyworker.Key{Market: strategyrouter.Market(latch.Market),
				Family: strategyrouter.Family(latch.Family), LaneID: latch.LaneID, LaneVersion: latch.LaneVersion},
			LatchID: latch.LatchID, LatchRevision: latch.LatchRevision, Reason: latch.Reason,
			Abnormal: latch.Abnormal, ObservedAt: latch.ObservedAt,
		})
	}
	return restores
}

// restoreLatches 는 프로세스가 처음 도는 주기에 한 번, 원장에서 레인을 다시
// 세운다.
//
// 한 번뿐인 이유: 두 시장의 주기가 같은 런타임을 나눠 쓰므로, 두 번째 시장이
// 다시 세우면 첫 시장이 이번 프로세스에서 만든 잠금이 지워진다.
func (runtime *strategyLaneRuntime) restoreLatches(ctx context.Context) error {
	if runtime == nil {
		return nil
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.restored {
		return nil
	}
	// 원장이 없으면 기록도 없다. 그 상태에서 진입은 어차피 불가능하다 —
	// dispatch 는 저널 없이 리스를 낼 수 없다. 그러므로 여기서 막을 것이 없다.
	if runtime.ledger == nil || runtime.accountRef == "" {
		runtime.restored = true
		return nil
	}
	open, err := runtime.ledger.OpenStrategyLaneLatches(ctx, runtime.accountRef)
	if err != nil {
		return fmt.Errorf("engine: reading durable strategy lane latches: %w", err)
	}
	lanes, unmatched := strategyworker.ProductionLanesFrom(runtime.clk, strategyLaneLatchRestores(open))
	runtime.lanes = lanes
	runtime.latches = make(map[strategyworker.Key]journal.StrategyLaneLatch, len(open))
	runtime.unmatched = nil
	stale := map[strategyworker.Key]struct{}{}
	for _, restore := range unmatched {
		stale[restore.Key] = struct{}{}
	}
	for _, latch := range open {
		key := strategyworker.Key{Market: strategyrouter.Market(latch.Market),
			Family: strategyrouter.Family(latch.Family), LaneID: latch.LaneID, LaneVersion: latch.LaneVersion}
		if _, isStale := stale[key]; isStale {
			// 붙지 않은 기록도 **들고 있는다.** 버리면 그 기록을 복구할 방법이
			// 사라지고, 아래 오류는 빠져나갈 문이 없는 fail-closed 가 된다.
			runtime.unmatched = append(runtime.unmatched, latch)
			continue
		}
		runtime.latches[key] = latch
	}
	runtime.restored = true
	return nil
}

// staleLatchError 는 이 시장에 **붙지 않은** 기록이 남아 있다는 답이다.
//
// 조용히 넘기지 않는 이유: lane_id 나 버전이 바뀌면 사람이 잠가 둔 것이 아무 일도
// 하지 않으면서 기록으로만 남는다. 운영자는 잠겨 있다고 믿고 런타임은 열려 있다.
//
// 그리고 이 오류에는 **나가는 문이 있다.** 위 `recoverMarketLanes` 가 붙지 않은
// 기록에도 복구를 시도하므로, 더 큰 서명 활성화 세대가 오면 기록이 닫히고 이
// 오류도 사라진다. 문 없는 fail-closed 는 이 change 가 이미 한 번 만들었다가
// 고친 모양이다.
func (runtime *strategyLaneRuntime) staleLatchError(market StrategyMarket) error {
	routerMarket := strategyRouterMarket(market)
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	stale := []string{}
	for _, latch := range runtime.unmatched {
		if strategyrouter.Market(latch.Market) == routerMarket {
			stale = append(stale, latch.LatchID)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	return fmt.Errorf("engine: durable strategy lane latch record(s) name no lane in this build: %v"+
		" — raise the signed activation generation to close them", stale)
}

// recoverMarketLanes 는 증거가 도착한 레인을 다시 세운다.
//
// 잠금을 **푸는** 함수가 아니다. 그런 함수는 `strategyworker` 에도 여기에도
// 없다. 여기서 하는 일은 원장이 복구를 받아들였을 때(= 더 큰 서명 활성화 세대를
// 보였을 때) 그 레인을 **기록 없이 다시 태어나게** 하는 것이다.
//
// **세대를 여기서 비교하지 않는다.** 첫 판본은 `activationGeneration >
// latch.ActivationGeneration` 인 것만 골라 원장에 물었고, 그러면 판정이 둘이
// 된다 — Go 의 부등호와 SQLite 트리거. 반증이 그 대가를 값으로 보여 줬다:
// 부등호를 `>=` 로 바꾼 변이는 **살아남고**(트리거가 막는다), 트리거를 지운
// 변이는 엔진 시험에서 **살아남는다**(부등호가 막는다). 각자가 상대의 시험을
// 통과시키므로, 둘 중 하나가 조용히 사라져도 아무 색도 변하지 않는다.
//
// 그래서 판정은 하나다: 기록이 사는 곳이 한다. 여기서는 이 시장의 열린 잠금
// 전부를 물어보고 답을 따른다. 거절은 흔한 답이고 오류가 아니다.
func (runtime *strategyLaneRuntime) recoverMarketLanes(ctx context.Context, market StrategyMarket,
	activationGeneration uint64,
) error {
	if runtime == nil || runtime.ledger == nil || activationGeneration == 0 {
		return nil
	}
	routerMarket := strategyRouterMarket(market)
	runtime.mu.Lock()
	pending := []journal.StrategyLaneLatch{}
	for key, latch := range runtime.latches {
		if key.Market == routerMarket {
			pending = append(pending, latch)
		}
	}
	// 붙지 않은 기록도 함께 묻는다. 그것이 위 staleLatchError 의 나가는 문이다.
	for _, latch := range runtime.unmatched {
		if strategyrouter.Market(latch.Market) == routerMarket {
			pending = append(pending, latch)
		}
	}
	runtime.mu.Unlock()

	for _, latch := range pending {
		if err := runtime.ledger.RecoverStrategyLaneLatch(ctx, latch.Seq, activationGeneration); err != nil {
			// 증거가 모자란 것은 여기서 결함이 아니다: 세대가 그사이 되돌아갔다는
			// 뜻이고, 그러면 레인은 잠긴 채로 남는 것이 맞다.
			if errors.Is(err, journal.ErrStrategyLaneLatchRecoveryEvidence) {
				continue
			}
			return fmt.Errorf("engine: recovering strategy lane latch: %w", err)
		}
		runtime.rebornLane(strategyworker.Key{Market: strategyrouter.Market(latch.Market),
			Family: strategyrouter.Family(latch.Family), LaneID: latch.LaneID, LaneVersion: latch.LaneVersion})
		runtime.forgetStaleLatch(latch.Seq)
	}
	return nil
}

// rebornLane 은 레인 하나를 기록 없이 다시 세운 것으로 갈아 끼운다.
//
// 살아 있는 레인의 잠금을 푸는 대신 새 레인을 세우는 이유: 푸는 길이 존재하면
// 그 길은 증거 없이도 불릴 수 있다. 여기서는 원장이 복구를 이미 받아들인 뒤에만
// 새 레인이 생기고, `strategyworker` 에는 잠금을 거짓으로 만드는 자리가 하나도
// 없다는 것을 그 패키지의 셈 시험이 지킨다.
func (runtime *strategyLaneRuntime) rebornLane(key strategyworker.Key) {
	fresh, unmatched := strategyworker.ProductionLanesFrom(runtime.clk, nil)
	if len(unmatched) != 0 {
		return
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	for index, lane := range runtime.lanes {
		if lane.Key() != key {
			continue
		}
		for _, candidate := range fresh {
			if candidate.Key() == key {
				runtime.lanes[index] = candidate
			}
		}
	}
	delete(runtime.latches, key)
}

// forgetStaleLatch 는 복구된 "붙지 않은 기록" 을 목록에서 뺀다.
func (runtime *strategyLaneRuntime) forgetStaleLatch(seq int64) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	kept := runtime.unmatched[:0]
	for _, latch := range runtime.unmatched {
		if latch.Seq != seq {
			kept = append(kept, latch)
		}
	}
	runtime.unmatched = kept
}

// persistMarketLatches 는 이 시장에서 **지금 잠겨 있는** 레인을 원장에 남긴다.
//
// "새로 잠긴 것" 이 아니라 "잠겨 있는 것" 을 매 주기 쓴다. 전이 순간에만 쓰면
// 그 한 번의 쓰기가 실패했을 때 다시 쓸 기회가 영영 없다 — 레인은 이미 잠겨
// 있어서 두 번 잠기지 않기 때문이다. 원장 쪽은 열린 잠금이 있으면 그것을
// 그대로 돌려주므로 이 되풀이는 첫 원인을 덮지 않는다.
func (runtime *strategyLaneRuntime) persistMarketLatches(ctx context.Context, market StrategyMarket,
	activationGeneration uint64, now time.Time,
) error {
	if runtime == nil || runtime.ledger == nil || runtime.accountRef == "" {
		return nil
	}
	for _, lane := range runtime.lanesFor(market) {
		if !lane.Latched() {
			continue
		}
		key := lane.Key()
		runtime.mu.RLock()
		_, known := runtime.latches[key]
		runtime.mu.RUnlock()
		if known {
			continue
		}
		stored, err := runtime.ledger.RecordStrategyLaneLatch(ctx, journal.StrategyLaneLatch{
			AccountRef: runtime.accountRef, Market: string(key.Market), Family: string(key.Family),
			LaneID: key.LaneID, LaneVersion: key.LaneVersion,
			LatchID: strategyLaneLatchID(key, lane.LatchRevision()), LatchRevision: lane.LatchRevision(),
			Reason: lane.FirstFailure(), Abnormal: lane.FirstAbnormal(),
			ActivationGeneration: activationGeneration, ObservedAt: now.UTC(),
		})
		if err != nil {
			return fmt.Errorf("engine: recording durable strategy lane latch: %w", err)
		}
		runtime.mu.Lock()
		runtime.latches[key] = stored
		runtime.mu.Unlock()
	}
	return nil
}

// strategyLaneLatchID 는 레인이 자기 고장 기록에 붙이는 것과 같은 철자다.
// 원장이 신원으로 쓰는 것은 이 문자열이 아니라 자기 순번이고, 이 값은 운영자가
// 두 기록을 눈으로 잇기 위한 것이다.
func strategyLaneLatchID(key strategyworker.Key, revision uint64) string {
	return fmt.Sprintf("lane-latch:%s:%s:%s:%s:%d", key.Market, key.Family, key.LaneID, key.LaneVersion, revision)
}
