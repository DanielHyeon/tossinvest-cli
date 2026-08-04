//go:build tossos_testseams

package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycandidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestStrategyRouteAuthorityLoadsKRUSConcurrently(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	loader := testStrategyRouteLoader(t, now)
	arrived := make(chan StrategyMarket, 2)
	release := make(chan struct{})
	loader.load = func(_ context.Context, config strategyrouter.ProductionRouteConfig, targets []strategyrouter.ProductionRouteTarget) (strategyrouter.ProductionRouteBatchAuthority, error) {
		arrived <- routeStrategyMarket(config.Market)
		<-release
		return testRouteBatchAuthority(config, targets, nil)
	}
	done := make(chan strategyRouteAuthorityPair, 1)
	go func() {
		done <- loader.collect(context.Background(), routeReadySchedulePair(now), routeCandidatePair(now, []string{"005930"}, []string{"AAPL"}))
	}()

	seen := map[StrategyMarket]bool{}
	for len(seen) < 2 {
		select {
		case market := <-arrived:
			seen[market] = true
		case <-time.After(time.Second):
			t.Fatal("KR and US route loading did not overlap")
		}
	}
	close(release)
	pair := <-done
	if !pair.kr.snapshot.Ready || !pair.us.snapshot.Ready || pair.kr.snapshot.RoutedCount != 1 || pair.us.snapshot.RoutedCount != 1 {
		t.Fatalf("paired route snapshots = KR:%+v US:%+v", pair.kr.snapshot, pair.us.snapshot)
	}
}

func TestStrategyRouteAuthorityKeepsMarketFailureLocal(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	loader := testStrategyRouteLoader(t, now)
	loader.load = func(_ context.Context, config strategyrouter.ProductionRouteConfig, targets []strategyrouter.ProductionRouteTarget) (strategyrouter.ProductionRouteBatchAuthority, error) {
		if config.Market == strategyrouter.MarketKR {
			return strategyrouter.ProductionRouteBatchAuthority{}, errors.New("synthetic KR authority failure")
		}
		return testRouteBatchAuthority(config, targets, nil)
	}
	pair := loader.collect(context.Background(), routeReadySchedulePair(now), routeCandidatePair(now, []string{"005930"}, []string{"AAPL"}))
	if pair.kr.snapshot.Ready || pair.kr.snapshot.Reason != StrategyRouteAuthorityInvalid {
		t.Fatalf("KR route snapshot = %+v", pair.kr.snapshot)
	}
	if !pair.us.snapshot.Ready || pair.us.snapshot.RoutedCount != 1 {
		t.Fatalf("US route snapshot = %+v", pair.us.snapshot)
	}
}

func TestStrategyRouteAuthorityRoutesAllSymbolsAndCountsLocalRefusal(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	loader := testStrategyRouteLoader(t, now)
	loader.load = func(_ context.Context, config strategyrouter.ProductionRouteConfig, targets []strategyrouter.ProductionRouteTarget) (strategyrouter.ProductionRouteBatchAuthority, error) {
		var omit map[string]bool
		if config.Market == strategyrouter.MarketKR {
			omit = map[string]bool{"000660": true}
		}
		return testRouteBatchAuthority(config, targets, omit)
	}
	pair := loader.collect(context.Background(), routeReadySchedulePair(now), routeCandidatePair(now,
		[]string{"005930", "000660"}, []string{"AAPL", "MSFT"}))
	if !pair.kr.snapshot.Ready || pair.kr.snapshot.ApprovedCount != 2 || pair.kr.snapshot.RoutedCount != 1 || pair.kr.snapshot.RefusedCount != 1 {
		t.Fatalf("KR route snapshot = %+v", pair.kr.snapshot)
	}
	if !pair.us.snapshot.Ready || pair.us.snapshot.ApprovedCount != 2 || pair.us.snapshot.RoutedCount != 2 || pair.us.snapshot.RefusedCount != 0 {
		t.Fatalf("US route snapshot = %+v", pair.us.snapshot)
	}
}

func testStrategyRouteLoader(t *testing.T, now time.Time) *strategyRouteAuthorityLoader {
	t.Helper()
	key := make([]byte, ed25519.PublicKeySize)
	for index := range key {
		key[index] = byte(index + 1)
	}
	env := map[string]string{
		strategyRouteKRManifestDigestEnv: "sha256:test-route-KR",
		strategyRouteUSManifestDigestEnv: "sha256:test-route-US",
		strategyRouteKeyIDEnv:            "route-key-1",
		strategyRoutePublicKeyEnv:        base64.StdEncoding.EncodeToString(key),
	}
	dir := t.TempDir()
	return newStrategyRouteAuthorityLoader(dir, filepath.Join(dir, "journal.db"), "acct", func(name string) string { return env[name] })
}

func routeReadySchedulePair(now time.Time) strategyScheduleAuthorityPair {
	market := func(value StrategyMarket) strategyScheduleMarketAuthority {
		scope := strategySchedulerMarket(value)
		desired := scheduler.DesiredState{Revision: 7, Version: scheduler.SchedulerVersion, Enabled: true, AutoStart: true,
			Market: scope, Session: scheduler.SessionRegular, Actor: "test-human", ApprovedAt: now.Add(-time.Minute),
			CalendarVersion: "calendar-" + string(value), ConfigVersion: "config-v1"}
		activation := scheduler.ActivationForTest(desired.ActivationBinding("build-test"))
		return strategyScheduleMarketAuthority{market: value, desired: desired,
			calendar: scheduler.CalendarSnapshot{Version: desired.CalendarVersion},
			restore:  scheduler.RestoreResult{Restored: true, Reason: scheduler.ResumeExactManifest, Activation: activation},
			snapshot: StrategyScheduleMarketSnapshot{Market: value, Ready: true, Reason: scheduler.ResumeExactManifest,
				CalendarVersion: desired.CalendarVersion, ActivationManifestDigest: "sha256:test-activation-" + string(value)}}
	}
	return strategyScheduleAuthorityPair{observedAt: now, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func routeCandidatePair(now time.Time, krSymbols, usSymbols []string) strategyCandidateAuthorityPair {
	market := func(value StrategyMarket, symbols []string) strategyCandidateMarketAuthority {
		approved := make([]strategy.ApprovedSnapshot, 0, len(symbols))
		for _, symbol := range symbols {
			approved = append(approved, strategy.ApprovedSnapshotForTest(string(value), symbol, now))
		}
		return strategyCandidateMarketAuthority{market: value, approved: strategycandidate.ApprovedBatchForTest(approved...),
			snapshot: StrategyCandidateMarketSnapshot{Market: value, Ready: true, Reason: StrategyCandidateReady,
				CandidateCount: len(symbols), ApprovedCount: len(symbols)}}
	}
	return strategyCandidateAuthorityPair{observedAt: now, kr: market(StrategyMarketKR, krSymbols), us: market(StrategyMarketUS, usSymbols)}
}

func testRouteBatchAuthority(config strategyrouter.ProductionRouteConfig, targets []strategyrouter.ProductionRouteTarget,
	omit map[string]bool,
) (strategyrouter.ProductionRouteBatchAuthority, error) {
	authorities := make([]strategyrouter.ProductionRouteAuthority, 0, len(targets))
	for _, target := range targets {
		if omit[target.Symbol] {
			continue
		}
		key, err := strategyrouter.NewOwnerKey(config.AccountRef, config.Market, target.Symbol, 1)
		if err != nil {
			return strategyrouter.ProductionRouteBatchAuthority{}, err
		}
		laneID := continuationlane.KRContinuationLaneID
		if config.Market == strategyrouter.MarketUS {
			laneID = continuationlane.USContinuationLaneID
		}
		authority, err := strategyrouter.ProductionRouteAuthorityForTest(key, strategyrouter.HorizonShort, laneID,
			continuationlane.LaneVersionV1, "sha256:test-route-evidence", "sha256:test-route-config", config.ObservedAt)
		if err != nil {
			return strategyrouter.ProductionRouteBatchAuthority{}, err
		}
		authorities = append(authorities, authority)
	}
	return strategyrouter.ProductionRouteBatchAuthorityForTest(config.ManifestDigest, authorities...), nil
}

func routeStrategyMarket(market strategyrouter.Market) StrategyMarket {
	if market == strategyrouter.MarketKR {
		return StrategyMarketKR
	}
	return StrategyMarketUS
}
