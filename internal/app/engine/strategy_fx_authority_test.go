package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeStrategyFXCollector struct {
	mu      sync.Mutex
	calls   map[StrategyMarket]int
	results map[StrategyMarket]strategyFXRead
	errors  map[StrategyMarket]error
	started chan StrategyMarket
	release <-chan struct{}
}

func (collector *fakeStrategyFXCollector) collectKR(ctx context.Context) (strategyFXRead, error) {
	return collector.collect(ctx, StrategyMarketKR)
}

func (collector *fakeStrategyFXCollector) collectUS(ctx context.Context) (strategyFXRead, error) {
	return collector.collect(ctx, StrategyMarketUS)
}

func (collector *fakeStrategyFXCollector) collect(ctx context.Context, market StrategyMarket) (strategyFXRead, error) {
	collector.mu.Lock()
	if collector.calls == nil {
		collector.calls = map[StrategyMarket]int{}
	}
	collector.calls[market]++
	collector.mu.Unlock()
	if collector.started != nil {
		collector.started <- market
	}
	if collector.release != nil {
		select {
		case <-collector.release:
		case <-ctx.Done():
			return strategyFXRead{}, ctx.Err()
		}
	}
	return collector.results[market], collector.errors[market]
}

func TestStrategyFXAuthorityCollectsKRAndUSInTheSameWave(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	started := make(chan StrategyMarket, 2)
	release := make(chan struct{})
	fake := &fakeStrategyFXCollector{started: started, release: release,
		results: map[StrategyMarket]strategyFXRead{
			StrategyMarketKR: {valid: true, quoteCurrency: "KRW", accountCurrency: "KRW", digest: "sha256:" + strings.Repeat("a", 64)},
			StrategyMarketUS: {valid: true, quoteCurrency: "USD", accountCurrency: "KRW", digest: "sha256:" + strings.Repeat("b", 64)},
		}, errors: map[StrategyMarket]error{}}
	loader := newStrategyFXAuthorityLoader(t.TempDir(), "acct-fx", "KRW", now, nil, func(string) string { return "" })
	loader.collector = fake
	done := make(chan strategyFXAuthorityPair, 1)
	go func() { done <- loader.collect(context.Background(), readyCandidatePair(now)) }()
	seen := map[StrategyMarket]bool{<-started: true, <-started: true}
	if !seen[StrategyMarketKR] || !seen[StrategyMarketUS] {
		t.Fatalf("same-wave starts = %v", seen)
	}
	close(release)
	pair := <-done
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		got := pair.Snapshot().For(market)
		if !got.Ready || got.Reason != StrategyFXReady || got.Digest == "" || got.AccountCurrency != "KRW" {
			t.Fatalf("FX snapshot(%s) = %+v", market, got)
		}
	}
}

func TestStrategyFXAuthorityFailureIsMarketLocal(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	for _, failed := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		fake := &fakeStrategyFXCollector{results: map[StrategyMarket]strategyFXRead{
			StrategyMarketKR: {valid: true, quoteCurrency: "KRW", accountCurrency: "KRW", digest: "sha256:" + strings.Repeat("a", 64)},
			StrategyMarketUS: {valid: true, quoteCurrency: "USD", accountCurrency: "KRW", digest: "sha256:" + strings.Repeat("b", 64)},
		}, errors: map[StrategyMarket]error{failed: errors.New("synthetic FX refusal")}}
		loader := newStrategyFXAuthorityLoader(t.TempDir(), "acct-fx", "KRW", now, nil, func(string) string { return "" })
		loader.collector = fake
		pair := loader.collect(context.Background(), readyCandidatePair(now)).Snapshot()
		peer := StrategyMarketKR
		if failed == StrategyMarketKR {
			peer = StrategyMarketUS
		}
		if pair.For(failed).Ready || pair.For(failed).Reason != StrategyFXAuthorityUnavailable {
			t.Fatalf("failed %s snapshot = %+v", failed, pair.For(failed))
		}
		if !pair.For(peer).Ready || pair.For(peer).Reason != StrategyFXReady {
			t.Fatalf("peer %s snapshot = %+v", peer, pair.For(peer))
		}
	}
}

func TestStrategyFXAuthorityCandidateOffPerformsNoMarketRead(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	fake := &fakeStrategyFXCollector{results: map[StrategyMarket]strategyFXRead{}, errors: map[StrategyMarket]error{}}
	loader := newStrategyFXAuthorityLoader(t.TempDir(), "acct-fx", "KRW", now, nil, func(string) string { return "" })
	loader.collector = fake
	pair := loader.collect(context.Background(), failedStrategyCandidatePair(now, StrategyCandidateScheduleNotReady)).Snapshot()
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.calls[StrategyMarketKR] != 0 || fake.calls[StrategyMarketUS] != 0 {
		t.Fatalf("OFF collector calls = %+v", fake.calls)
	}
	if pair.KR.Reason != StrategyFXCandidateNotReady || pair.US.Reason != StrategyFXCandidateNotReady {
		t.Fatalf("OFF snapshots = KR:%+v US:%+v", pair.KR, pair.US)
	}
}

func TestStrategyFXAuthorityLoaderPinsExactProductionTrustAndFrozenTime(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	key := make(ed25519.PublicKey, ed25519.PublicKeySize)
	for index := range key {
		key[index] = byte(index + 1)
	}
	env := map[string]string{
		strategyFXManifestDigestEnv: "sha256:" + strings.Repeat("c", 64),
		strategyFXKeyIDEnv:          "fx-key-1",
		strategyFXPublicKeyEnv:      base64.StdEncoding.EncodeToString(key),
	}
	loader := newStrategyFXAuthorityLoader(t.TempDir(), "acct-fx", "KRW", now, nil, func(key string) string { return env[key] })
	if loader.config.AccountID != "acct-fx" || loader.config.AccountCurrency != "KRW" ||
		loader.config.ManifestDigest != env[strategyFXManifestDigestEnv] || loader.config.TrustedKeyID != "fx-key-1" ||
		!loader.config.Now().Equal(now) || !stringEqualBytes(loader.config.TrustedKey, key) {
		t.Fatalf("production FX config = %+v", loader.config)
	}
}

func readyCandidatePair(now time.Time) strategyCandidateAuthorityPair {
	market := func(value StrategyMarket) strategyCandidateMarketAuthority {
		return strategyCandidateMarketAuthority{market: value,
			snapshot: StrategyCandidateMarketSnapshot{Market: value, Ready: true, Reason: StrategyCandidateReady}}
	}
	return strategyCandidateAuthorityPair{observedAt: now, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func stringEqualBytes(left, right []byte) bool { return string(left) == string(right) }
