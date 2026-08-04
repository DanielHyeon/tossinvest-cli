package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const (
	strategyRouteKRManifestDigestEnv = "TOSSOS_STRATEGY_LANE_KR_MANIFEST_SHA256"
	strategyRouteUSManifestDigestEnv = "TOSSOS_STRATEGY_LANE_US_MANIFEST_SHA256"
	strategyRouteKeyIDEnv            = "TOSSOS_STRATEGY_LANE_KEY_ID"
	strategyRoutePublicKeyEnv        = "TOSSOS_STRATEGY_LANE_PUBLIC_KEY_BASE64"
)

type StrategyRouteReason string

const (
	StrategyRouteReady             StrategyRouteReason = "READY"
	StrategyRouteScheduleNotReady  StrategyRouteReason = "SCHEDULE_NOT_READY"
	StrategyRouteCandidateNotReady StrategyRouteReason = "CANDIDATE_NOT_READY"
	StrategyRouteNoCandidate       StrategyRouteReason = "NO_APPROVED_CANDIDATE"
	StrategyRouteAuthorityInvalid  StrategyRouteReason = "ROUTE_AUTHORITY_INVALID"
	StrategyRouteInternalFailure   StrategyRouteReason = "INTERNAL_FAILURE"
)

type StrategyRouteMarketSnapshot struct {
	Market                         StrategyMarket
	Ready                          bool
	Reason                         StrategyRouteReason
	ApprovedCount, RoutedCount     int
	RefusedCount                   int
	ManifestDigest, OwnerSetDigest string
}

type PairedStrategyRouteSnapshot struct {
	ObservedAt time.Time
	KR, US     StrategyRouteMarketSnapshot
}

func (snapshot PairedStrategyRouteSnapshot) For(market StrategyMarket) StrategyRouteMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyRouteMarketSnapshot{Market: market, Reason: StrategyRouteInternalFailure}
}

type strategyRouteEntryAuthority struct {
	approved strategy.ApprovedSnapshot
	route    strategyrouter.ProductionRouteAuthority
}

type strategyRouteMarketAuthority struct {
	market   StrategyMarket
	entries  []strategyRouteEntryAuthority
	snapshot StrategyRouteMarketSnapshot
}

type strategyRouteAuthorityPair struct {
	observedAt time.Time
	kr, us     strategyRouteMarketAuthority
}

func (pair strategyRouteAuthorityPair) forMarket(market StrategyMarket) strategyRouteMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyRouteMarketAuthority{market: market}
}

func (pair strategyRouteAuthorityPair) Snapshot() PairedStrategyRouteSnapshot {
	return PairedStrategyRouteSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type strategyRouteAuthorityLoader struct {
	configDir, journalPath, accountRef string
	getenv                             func(string) string
	load                               func(context.Context, strategyrouter.ProductionRouteConfig, []strategyrouter.ProductionRouteTarget) (strategyrouter.ProductionRouteBatchAuthority, error)
}

func newStrategyRouteAuthorityLoader(configDir, journalPath, accountRef string, getenv func(string) string) *strategyRouteAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	return &strategyRouteAuthorityLoader{configDir: strings.TrimSpace(configDir), journalPath: strings.TrimSpace(journalPath),
		accountRef: strings.TrimSpace(accountRef), getenv: getenv, load: strategyrouter.LoadProductionRouteAuthorityBatch}
}

func (loader *strategyRouteAuthorityLoader) collect(ctx context.Context, schedule strategyScheduleAuthorityPair, candidates strategyCandidateAuthorityPair) strategyRouteAuthorityPair {
	if loader == nil || ctx == nil || schedule.observedAt.IsZero() || !schedule.observedAt.Equal(candidates.observedAt) {
		return failedStrategyRoutePair(schedule.observedAt, StrategyRouteInternalFailure)
	}
	type outcome struct {
		market StrategyMarket
		value  strategyRouteMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			value := strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market, Reason: StrategyRouteInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						value = strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market, Reason: StrategyRouteInternalFailure}}
					}
				}()
				value = loader.collectMarket(ctx, schedule.forMarket(market), candidates.forMarket(market), schedule.observedAt)
			}()
			outcomes <- outcome{market: market, value: value}
		}()
	}
	pair := strategyRouteAuthorityPair{observedAt: schedule.observedAt}
	for index := 0; index < 2; index++ {
		value := <-outcomes
		if value.market == StrategyMarketKR {
			pair.kr = value.value
		} else {
			pair.us = value.value
		}
	}
	return pair
}

func (loader *strategyRouteAuthorityLoader) collectMarket(ctx context.Context, schedule strategyScheduleMarketAuthority,
	candidates strategyCandidateMarketAuthority, observedAt time.Time,
) strategyRouteMarketAuthority {
	market := schedule.market
	fail := func(reason StrategyRouteReason) strategyRouteMarketAuthority {
		return strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market, Reason: reason}}
	}
	if !schedule.snapshot.Ready || schedule.restore.Activation == nil {
		return fail(StrategyRouteScheduleNotReady)
	}
	if !candidates.snapshot.Ready {
		return fail(StrategyRouteCandidateNotReady)
	}
	if candidates.approved.Len() == 0 {
		return fail(StrategyRouteNoCandidate)
	}
	if loader.getenv == nil || loader.load == nil || loader.configDir == "" || loader.journalPath == "" || loader.accountRef == "" {
		return fail(StrategyRouteInternalFailure)
	}
	encoded := strings.TrimSpace(loader.getenv(strategyRoutePublicKeyEnv))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		return fail(StrategyRouteAuthorityInvalid)
	}
	digest := strings.TrimSpace(loader.getenv(map[StrategyMarket]string{StrategyMarketKR: strategyRouteKRManifestDigestEnv, StrategyMarketUS: strategyRouteUSManifestDigestEnv}[market]))
	keyID := strings.TrimSpace(loader.getenv(strategyRouteKeyIDEnv))
	entries := make([]strategyRouteEntryAuthority, 0, candidates.approved.Len())
	seen := make(map[string]bool, candidates.approved.Len())
	targets := make([]strategyrouter.ProductionRouteTarget, 0, candidates.approved.Len())
	approvedValues := make([]strategy.ApprovedSnapshot, 0, candidates.approved.Len())
	for index := 0; index < candidates.approved.Len(); index++ {
		approved, ok := candidates.approved.At(index)
		if !ok || !approved.Valid() || approved.Market() != string(market) || seen[approved.Symbol()] {
			return fail(StrategyRouteInternalFailure)
		}
		seen[approved.Symbol()] = true
		targets = append(targets, strategyrouter.ProductionRouteTarget{Symbol: approved.Symbol()})
		approvedValues = append(approvedValues, approved)
	}
	batch, err := loader.load(ctx, strategyrouter.ProductionRouteConfig{ConfigDir: loader.configDir, JournalPath: loader.journalPath,
		AccountRef: loader.accountRef, Market: strategyRouterMarket(market), ManifestDigest: digest,
		TrustedKeyID: keyID, TrustedKey: ed25519.PublicKey(key), ObservedAt: observedAt,
		ActivationDigest: schedule.snapshot.ActivationManifestDigest, CalendarGeneration: schedule.desired.CalendarVersion,
		CalendarDigest: schedule.calendar.Version, SchedulerConfigVersion: schedule.desired.ConfigVersion}, targets)
	if err != nil || batch.ManifestDigest() != digest {
		return strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market,
			Reason: StrategyRouteAuthorityInvalid, ApprovedCount: candidates.approved.Len(), RefusedCount: candidates.approved.Len(), ManifestDigest: digest}}
	}
	refused := 0
	for _, approved := range approvedValues {
		authority, ok := batch.For(approved.Symbol())
		if !ok {
			refused++
			continue
		}
		request := authority.Request()
		routed := strategyrouter.Route(request)
		if routed.Code != strategyrouter.RefusalNone || request.Key.Market != strategyRouterMarket(market) || request.Key.Symbol != approved.Symbol() {
			refused++
			continue
		}
		entries = append(entries, strategyRouteEntryAuthority{approved: approved, route: authority})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].approved.Symbol() < entries[j].approved.Symbol() })
	if len(entries) == 0 {
		return strategyRouteMarketAuthority{market: market, snapshot: StrategyRouteMarketSnapshot{Market: market,
			Reason: StrategyRouteAuthorityInvalid, ApprovedCount: candidates.approved.Len(), RefusedCount: refused, ManifestDigest: digest}}
	}
	h := sha256.New()
	for _, entry := range entries {
		_, _ = h.Write([]byte(entry.approved.Symbol() + "\x00" + entry.route.OwnerDigest() + "\x00"))
	}
	return strategyRouteMarketAuthority{market: market, entries: entries,
		snapshot: StrategyRouteMarketSnapshot{Market: market, Ready: true, Reason: StrategyRouteReady, ApprovedCount: candidates.approved.Len(),
			RoutedCount: len(entries), RefusedCount: refused, ManifestDigest: digest, OwnerSetDigest: "sha256:" + hex.EncodeToString(h.Sum(nil))}}
}

func strategyRouterMarket(market StrategyMarket) strategyrouter.Market {
	if market == StrategyMarketKR {
		return strategyrouter.MarketKR
	}
	if market == StrategyMarketUS {
		return strategyrouter.MarketUS
	}
	return ""
}

func failedStrategyRoutePair(observedAt time.Time, reason StrategyRouteReason) strategyRouteAuthorityPair {
	market := func(value StrategyMarket) strategyRouteMarketAuthority {
		return strategyRouteMarketAuthority{market: value, snapshot: StrategyRouteMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyRouteAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
