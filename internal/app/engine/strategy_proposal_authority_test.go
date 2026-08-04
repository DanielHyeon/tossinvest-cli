//go:build tossos_testseams

package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestStrategyProposalAuthorityLoadsKRUSConcurrently(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	loader := testStrategyProposalLoader(t)
	started := make(chan StrategyMarket, 2)
	release := make(chan struct{})
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		started <- StrategyMarket(config.Market)
		<-release
		return testProposalBatch(t, config, targets, now), nil
	}
	done := make(chan strategyProposalAuthorityPair, 1)
	go func() {
		done <- loader.collect(context.Background(), routeReadySchedulePair(now), proposalRoutePair(t, now), proposalFXPair(now))
	}()
	seen := map[StrategyMarket]bool{<-started: true, <-started: true}
	if !seen[StrategyMarketKR] || !seen[StrategyMarketUS] {
		t.Fatalf("same-wave starts=%v", seen)
	}
	close(release)
	pair := <-done
	if !pair.kr.snapshot.Ready || !pair.us.snapshot.Ready || pair.kr.snapshot.ProposedCount != 1 || pair.us.snapshot.ProposedCount != 1 ||
		!pair.ResultAuthority().kr.ready || !pair.ResultAuthority().us.ready {
		t.Fatalf("paired proposals KR=%+v US=%+v", pair.kr.snapshot, pair.us.snapshot)
	}
}

func TestStrategyProposalAuthorityKeepsMarketFailureLocal(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 2, 3, 0, time.UTC)
	loader := testStrategyProposalLoader(t)
	loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
		if config.Market == strategyrouter.MarketKR {
			return strategyproposal.ProductionBatchAuthority{}, errors.New("synthetic KR proposal failure")
		}
		return testProposalBatch(t, config, targets, now), nil
	}
	pair := loader.collect(context.Background(), routeReadySchedulePair(now), proposalRoutePair(t, now), proposalFXPair(now))
	if pair.kr.snapshot.Ready || pair.kr.snapshot.Reason != StrategyProposalAuthorityInvalid {
		t.Fatalf("KR=%+v", pair.kr.snapshot)
	}
	if !pair.us.snapshot.Ready || pair.us.snapshot.ProposedCount != 1 {
		t.Fatalf("US=%+v", pair.us.snapshot)
	}
}

// Alias keeps fake loader signatures readable while remaining exactly the
// production officialfx.Evidence type.
type interfaceOfficialFX = officialfx.Evidence

func testStrategyProposalLoader(t *testing.T) *strategyProposalAuthorityLoader {
	t.Helper()
	key := make(ed25519.PublicKey, ed25519.PublicKeySize)
	for index := range key {
		key[index] = byte(index + 1)
	}
	env := map[string]string{strategyProposalKRManifestDigestEnv: "sha256:" + strings.Repeat("a", 64), strategyProposalUSManifestDigestEnv: "sha256:" + strings.Repeat("b", 64),
		strategyProposalKeyIDEnv: "proposal-key", strategyProposalPublicKeyEnv: base64.StdEncoding.EncodeToString(key),
		strategyProposalKREvidenceIDEnv: "evidence-KR", strategyProposalUSEvidenceIDEnv: "evidence-US"}
	dir := t.TempDir()
	return newStrategyProposalAuthorityLoader(dir, filepath.Join(dir, "evidence.db"), filepath.Join(dir, "journal.db"), "acct", func(key string) string { return env[key] })
}

func proposalRoutePair(t *testing.T, now time.Time) strategyRouteAuthorityPair {
	t.Helper()
	market := func(value StrategyMarket, symbol string) strategyRouteMarketAuthority {
		approved := strategy.ApprovedSnapshotForTest(string(value), symbol, now)
		routerMarket, lane := strategyrouter.MarketKR, continuationlane.KRContinuationLaneID
		if value == StrategyMarketUS {
			routerMarket, lane = strategyrouter.MarketUS, continuationlane.USContinuationLaneID
		}
		key, err := strategyrouter.NewOwnerKey("acct", routerMarket, symbol, 1)
		if err != nil {
			t.Fatal(err)
		}
		route, err := strategyrouter.ProductionRouteAuthorityForTest(key, strategyrouter.HorizonShort, lane, continuationlane.LaneVersionV1, "lane-evidence", "lane-config", now)
		if err != nil {
			t.Fatal(err)
		}
		entry := strategyRouteEntryAuthority{approved: approved, route: route}
		return strategyRouteMarketAuthority{market: value, entries: []strategyRouteEntryAuthority{entry}, snapshot: StrategyRouteMarketSnapshot{Market: value, Ready: true, Reason: StrategyRouteReady, RoutedCount: 1, ManifestDigest: "route-" + string(value)}}
	}
	return strategyRouteAuthorityPair{observedAt: now, kr: market(StrategyMarketKR, "005930"), us: market(StrategyMarketUS, "AAPL")}
}

func proposalFXPair(now time.Time) strategyFXAuthorityPair {
	market := func(value StrategyMarket) strategyFXMarketAuthority {
		return strategyFXMarketAuthority{market: value, read: strategyFXRead{valid: true}, snapshot: StrategyFXMarketSnapshot{Market: value, Ready: true, Reason: StrategyFXReady}}
	}
	return strategyFXAuthorityPair{observedAt: now, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func testProposalBatch(t *testing.T, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, now time.Time) strategyproposal.ProductionBatchAuthority {
	t.Helper()
	values := make(map[string]strategyflow.Result, len(targets))
	for _, target := range targets {
		descriptor := riskLoaderDescriptor(t, StrategyMarket(config.Market))
		result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, config.AccountRef, target.Approved.Symbol(), "campaign-"+string(config.Market), 8, "100", "90", "120", now.Add(-time.Second), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		values[target.Approved.Symbol()] = result
	}
	return strategyproposal.ProductionBatchAuthorityForTest(config.ManifestDigest, values)
}
