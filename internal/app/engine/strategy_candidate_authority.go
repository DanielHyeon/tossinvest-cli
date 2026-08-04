package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	candidatepkg "github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycandidate"
)

const strategyCandidateActivationEnvPrefix = "TOSSOS_CANDIDATE_THRESHOLD_"

type StrategyCandidateReason string

const (
	StrategyCandidateReady            StrategyCandidateReason = "READY"
	StrategyCandidateScheduleNotReady StrategyCandidateReason = "SCHEDULE_NOT_READY"
	StrategyCandidateThresholdInvalid StrategyCandidateReason = "THRESHOLD_AUTHORITY_INVALID"
	StrategyCandidateStoreUnavailable StrategyCandidateReason = "DISCOVERY_STORE_UNAVAILABLE"
	StrategyCandidateAssessmentFailed StrategyCandidateReason = "ASSESSMENT_FAILED"
	StrategyCandidateInternalFailure  StrategyCandidateReason = "INTERNAL_FAILURE"
)

// StrategyCandidateMarketSnapshot contains observations only. It cannot
// reconstruct a ThresholdSet, ApprovedCandidate or discovery-store handle.
type StrategyCandidateMarketSnapshot struct {
	Market             StrategyMarket
	Ready              bool
	Reason             StrategyCandidateReason
	ThresholdVersion   string
	ThresholdSetDigest string
	EvidenceDigest     string
	ActivationDigest   string
	CandidateCount     int
	ApprovedCount      int
	RefusedCount       int
}

type PairedStrategyCandidateSnapshot struct {
	ObservedAt time.Time
	KR         StrategyCandidateMarketSnapshot
	US         StrategyCandidateMarketSnapshot
}

func (snapshot PairedStrategyCandidateSnapshot) For(market StrategyMarket) StrategyCandidateMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyCandidateMarketSnapshot{Market: market, Reason: StrategyCandidateInternalFailure}
}

type strategyCandidateMarketAuthority struct {
	market   StrategyMarket
	approved strategycandidate.ApprovedBatch
	snapshot StrategyCandidateMarketSnapshot
}

type strategyCandidateAuthorityPair struct {
	observedAt time.Time
	kr         strategyCandidateMarketAuthority
	us         strategyCandidateMarketAuthority
}

func (pair strategyCandidateAuthorityPair) forMarket(market StrategyMarket) strategyCandidateMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyCandidateMarketAuthority{market: market}
}

func (pair strategyCandidateAuthorityPair) Snapshot() PairedStrategyCandidateSnapshot {
	return PairedStrategyCandidateSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type strategyCandidateAuthorityLoader struct {
	configDir     string
	storePath     string
	getenv        func(string) string
	loadAuthority func(context.Context, candidatepkg.ProductionThresholdAuthorityConfig, time.Time, time.Duration) (candidatepkg.ProductionThresholdAuthority, error)
	openStore     func(context.Context, candidatepkg.Options) (*candidatepkg.Store, error)
}

func newStrategyCandidateAuthorityLoader(configDir, storePath string, getenv func(string) string) *strategyCandidateAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	return &strategyCandidateAuthorityLoader{
		configDir:     filepath.Clean(strings.TrimSpace(configDir)),
		storePath:     filepath.Clean(strings.TrimSpace(storePath)),
		getenv:        getenv,
		loadAuthority: candidatepkg.LoadProductionThresholdAuthority,
		openStore:     candidatepkg.OpenReadOnly,
	}
}

func (loader *strategyCandidateAuthorityLoader) collect(ctx context.Context, schedule strategyScheduleAuthorityPair) strategyCandidateAuthorityPair {
	if loader == nil || ctx == nil || schedule.observedAt.IsZero() {
		return failedStrategyCandidatePair(schedule.observedAt, StrategyCandidateInternalFailure)
	}
	type outcome struct {
		market    StrategyMarket
		authority strategyCandidateMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			result := strategyCandidateMarketAuthority{market: market,
				snapshot: StrategyCandidateMarketSnapshot{Market: market, Reason: StrategyCandidateInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						result = strategyCandidateMarketAuthority{market: market,
							snapshot: StrategyCandidateMarketSnapshot{Market: market, Reason: StrategyCandidateInternalFailure}}
					}
				}()
				result = loader.collectMarket(ctx, schedule.forMarket(market), schedule.observedAt)
			}()
			outcomes <- outcome{market: market, authority: result}
		}()
	}
	pair := strategyCandidateAuthorityPair{observedAt: schedule.observedAt}
	for index := 0; index < 2; index++ {
		result := <-outcomes
		if result.market == StrategyMarketKR {
			pair.kr = result.authority
		} else {
			pair.us = result.authority
		}
	}
	return pair
}

func (loader *strategyCandidateAuthorityLoader) collectMarket(ctx context.Context, schedule strategyScheduleMarketAuthority, observedAt time.Time) strategyCandidateMarketAuthority {
	market := schedule.market
	fail := func(reason StrategyCandidateReason) strategyCandidateMarketAuthority {
		return strategyCandidateMarketAuthority{market: market,
			snapshot: StrategyCandidateMarketSnapshot{Market: market, Reason: reason}}
	}
	if !schedule.snapshot.Ready {
		return fail(StrategyCandidateScheduleNotReady)
	}
	if !filepath.IsAbs(loader.configDir) || !filepath.IsAbs(loader.storePath) || loader.getenv == nil ||
		loader.loadAuthority == nil || loader.openStore == nil {
		return fail(StrategyCandidateInternalFailure)
	}
	authority, err := loader.loadAuthority(ctx, candidatepkg.ProductionThresholdAuthorityConfig{
		ConfigDir: loader.configDir, Market: string(market),
		ActivationDigest: strings.TrimSpace(loader.getenv(strategyCandidateActivationDigestEnv(market))),
	}, observedAt, 0)
	if err != nil || !authority.Valid() || authority.Market() != string(market) {
		return fail(StrategyCandidateThresholdInvalid)
	}
	store, err := loader.openStore(ctx, candidatepkg.Options{Path: loader.storePath, Clock: clock.NewFake(observedAt)})
	if err != nil {
		return fail(StrategyCandidateStoreUnavailable)
	}
	defer store.Close()
	verdicts, err := candidatepkg.Assess(ctx, store, candidatepkg.AssessOptions{
		Market: string(market), At: observedAt, Thresholds: authority.ThresholdSet().VetoThresholds(),
	})
	if err != nil {
		return fail(StrategyCandidateAssessmentFailed)
	}
	approved := strategycandidate.ApprovedBatch{}
	refused := 0
	for _, verdict := range verdicts {
		updated, err := strategycandidate.Append(approved, verdict, authority, observedAt)
		if err != nil {
			refused++
			continue
		}
		approved = updated
	}
	return strategyCandidateMarketAuthority{market: market, approved: approved,
		snapshot: StrategyCandidateMarketSnapshot{Market: market, Ready: true, Reason: StrategyCandidateReady,
			ThresholdVersion: authority.Version(), ThresholdSetDigest: authority.SetDigest(),
			EvidenceDigest: authority.EvidenceDigest(), ActivationDigest: authority.ActivationDigest(),
			CandidateCount: len(verdicts), ApprovedCount: approved.Len(), RefusedCount: refused}}
}

func strategyCandidateActivationDigestEnv(market StrategyMarket) string {
	return strategyCandidateActivationEnvPrefix + string(market) + "_ACTIVATION_SHA256"
}

func failedStrategyCandidatePair(observedAt time.Time, reason StrategyCandidateReason) strategyCandidateAuthorityPair {
	market := func(value StrategyMarket) strategyCandidateMarketAuthority {
		return strategyCandidateMarketAuthority{market: value,
			snapshot: StrategyCandidateMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyCandidateAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
