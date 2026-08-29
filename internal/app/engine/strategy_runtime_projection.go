package engine

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// StrategyRuntimeProjection returns the engine-owned observation capability.
// The returned interface has no activation, order, journal, or toggle method.
func (c *Context) StrategyRuntimeProjection() strategyprojection.Reader {
	if c == nil {
		return nil
	}
	return c
}

// Read implements strategyprojection.Reader and overlays the live supervisor's
// latch state on the most recent paired authority observation.
func (c *Context) Read(ctx context.Context) (strategyprojection.Snapshot, error) {
	if c == nil || ctx == nil {
		return strategyprojection.Snapshot{}, errors.New("engine: strategy runtime projection unavailable")
	}
	c.strategyProjectionMu.RLock()
	store, supervisor := c.strategyProjection, c.strategySupervisor
	c.strategyProjectionMu.RUnlock()
	if store == nil {
		return strategyprojection.Snapshot{}, errors.New("engine: strategy runtime projection unavailable")
	}
	snapshot, err := store.Read(ctx)
	if err != nil {
		return snapshot, err
	}
	// 이 프로세스가 실제로 쓰는 두 digest 를 여기서 붙인다 — a112 결정 54.
	//
	// 왜 store 가 아니라 여기인가: store 의 내용은 전략 assembly refresh 가 채우는데,
	// 그 refresh 는 아직 한 번도 안 돌았거나 실패했을 수 있다. 운영자가 이 숫자를 가장
	// 필요로 하는 순간이 바로 그때다. Read 는 REST·SSE·콘솔·Unix transport 가 모두
	// 지나는 단 하나의 출구라서, 여기 붙이면 어느 표면에서도 같은 값이 보인다.
	//
	// **접두사를 벗기지 않는다.** 활성화 매니페스트 검증은 이 문자열을 그대로 비교한다
	// (`scheduler.validateProductionActivationManifest` 의 `body.ConfigVersion !=
	// binding.ConfigVersion`). 벗겨서 보여 주면 운영자가 옮겨 적은 매니페스트를 엔진이
	// 거절한다 — 이 기능이 없애려던 바로 그 실패다.
	snapshot = strategyprojection.WithRuntimeIdentity(snapshot,
		strategyRuntimeConfigDigest(), strategyRuntimeBuildDigest())
	if supervisor == nil {
		return snapshot, nil
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		worker, ok := supervisor.Snapshot(market)
		if !ok || !worker.Latched {
			continue
		}
		projectionMarket := strategyprojection.Market(market)
		if current := snapshot.Markets[projectionMarket]; current.Status == strategyprojection.StatusCurrent {
			snapshot = strategyprojection.WithMarketFailure(snapshot, projectionMarket,
				strategyprojection.RefusalRuntimeUnavailable, snapshot.GeneratedAt)
		}
	}
	return snapshot, nil
}

func (c *Context) publishStrategyRuntime(assembly StrategyEntryProductionAssembly) error {
	if c == nil || assembly.Supervisor == nil {
		return errors.New("engine: strategy runtime projection unavailable")
	}
	next := strategyProjectionFromAssembly(assembly)
	c.strategyProjectionMu.Lock()
	if c.strategyProjection == nil {
		store, err := strategyprojection.NewStore(next)
		if err != nil {
			c.strategyProjectionMu.Unlock()
			return err
		}
		c.strategyProjection = store
		c.strategySupervisor = assembly.Supervisor
		c.strategyProjectionMu.Unlock()
		return nil
	}
	store := c.strategyProjection
	c.strategyProjectionMu.Unlock()
	if err := store.Replace(next); err != nil {
		return err
	}
	c.strategyProjectionMu.Lock()
	if c.strategySupervisor == nil {
		c.strategySupervisor = assembly.Supervisor
	}
	c.strategyProjectionMu.Unlock()
	return nil
}

func strategyProjectionFromAssembly(assembly StrategyEntryProductionAssembly) strategyprojection.Snapshot {
	observed := assembly.Schedule.ObservedAt.UTC()
	snapshot := strategyprojection.DormantSnapshot(observed)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		projectionMarket := strategyprojection.Market(market)
		schedule := assembly.Schedule.For(market)
		candidate := assembly.Candidate.For(market)
		proposal := assembly.Proposal.For(market)
		risk := assembly.Risk.For(market)
		worker, ok := assembly.Supervisor.Snapshot(market)
		if !ok || !worker.Effective {
			code := strategyprojection.RefusalNotConfigured
			switch {
			case schedule.DesiredEnabled && !schedule.Ready:
				code = strategyprojection.RefusalActivationAbsent
			case schedule.Ready && (!candidate.Ready || !proposal.Ready || !risk.Ready):
				code = strategyprojection.RefusalEvidenceStale
			case schedule.Ready:
				code = strategyprojection.RefusalProtectionUnwired
			}
			snapshot = strategyprojection.WithMarketFailure(snapshot, projectionMarket, code, observed)
			continue
		}
		authority := assembly.proposals.forMarket(market)
		if len(authority.entries) != 1 || !authority.entries[0].authority.Proposal().ValidProposal() {
			snapshot = strategyprojection.WithMarketFailure(snapshot, projectionMarket,
				strategyprojection.RefusalEvidenceStale, observed)
			continue
		}
		result := authority.entries[0].authority.Proposal()
		lineage := result.Lineage
		laneID, laneVersion := lineage.LaneID, lineage.LaneVersion
		evidenceID, evidenceDigest := lineage.CandidateLifeID, projectionDigest(lineage.LaneEvidenceDigest)
		if evidenceDigest == "" {
			evidenceDigest = projectionDigest(lineage.CandidateEvidenceDigest)
		}
		campaignID, legID := lineage.CampaignID, strconv.Itoa(lineage.LegOrdinal)
		bucket, policyVersion := string(lineage.Horizon), risk.BundleDigest
		calendarSource, calendarVersion := "official-open-api-"+string(market), schedule.CalendarVersion
		activationDigest := projectionDigest(schedule.ActivationManifestDigest)
		item := strategyprojection.MarketProjection{Market: projectionMarket, Status: strategyprojection.StatusCurrent,
			Lane: strategyprojection.LaneProjection{ID: &laneID, Version: &laneVersion,
				Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn},
			Evidence: strategyprojection.EvidenceProjection{ID: &evidenceID, Digest: &evidenceDigest,
				Freshness: strategyprojection.FreshnessCurrent},
			Campaign:    strategyprojection.CampaignProjection{ID: &campaignID, LegID: &legID},
			HorizonRisk: strategyprojection.HorizonRiskProjection{Bucket: &bucket, PolicyVersion: &policyVersion, Status: strategyprojection.ComponentCurrent},
			Scheduler: strategyprojection.SchedulerProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
				CalendarSource: &calendarSource, CalendarVersion: &calendarVersion, CalendarFreshness: strategyprojection.FreshnessCurrent},
			Activation: strategyprojection.ActivationProjection{Desired: strategyprojection.StateOn, Effective: strategyprojection.StateOn,
				Digest: &activationDigest, Status: strategyprojection.ActivationConfigured},
			Protection:     strategyprojection.ProtectionProjection{Readiness: strategyprojection.ProtectionWired, Refusal: strategyprojection.RefusalNone},
			Reconciliation: strategyprojection.ReconciliationProjection{Status: strategyprojection.ReconciliationHealthy, Refusal: strategyprojection.RefusalNone},
			FirstRefusal:   strategyprojection.RefusalNone, ObservedAt: &observed}
		snapshot.Markets[projectionMarket] = item
	}
	return snapshot
}

func projectionDigest(value string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "sha256:")
}

var _ strategyprojection.Reader = (*Context)(nil)
