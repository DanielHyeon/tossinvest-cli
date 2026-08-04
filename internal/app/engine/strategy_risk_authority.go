package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

const (
	strategyRiskKRManifestDigestEnv = "TOSSOS_RISK_BUCKET_KR_MANIFEST_SHA256"
	strategyRiskUSManifestDigestEnv = "TOSSOS_RISK_BUCKET_US_MANIFEST_SHA256"
	strategyRiskKeyIDEnv            = "TOSSOS_RISK_BUCKET_POLICY_KEY_ID"
	strategyRiskPublicKeyEnv        = "TOSSOS_RISK_BUCKET_POLICY_PUBLIC_KEY_BASE64"
)

type StrategyRiskReason string

const (
	StrategyRiskReady                StrategyRiskReason = "READY"
	StrategyRiskLaneNotReady         StrategyRiskReason = "LANE_NOT_READY"
	StrategyRiskFXNotReady           StrategyRiskReason = "FX_NOT_READY"
	StrategyRiskAuthorityUnavailable StrategyRiskReason = "RISK_AUTHORITY_UNAVAILABLE"
	StrategyRiskInternalFailure      StrategyRiskReason = "INTERNAL_FAILURE"
)

type StrategyRiskMarketSnapshot struct {
	Market         StrategyMarket
	Ready          bool
	Reason         StrategyRiskReason
	Horizon        string
	StrategyRiskID string
	Sector         string
	Symbol         string
	BundleDigest   string
	BucketCount    int
}

type PairedStrategyRiskSnapshot struct {
	ObservedAt time.Time
	KR         StrategyRiskMarketSnapshot
	US         StrategyRiskMarketSnapshot
}

func (snapshot PairedStrategyRiskSnapshot) For(market StrategyMarket) StrategyRiskMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyRiskMarketSnapshot{Market: market, Reason: StrategyRiskInternalFailure}
}

type strategyResultMarketAuthority struct {
	market StrategyMarket
	ready  bool
	result strategyflow.Result
}

type strategyResultAuthorityPair struct {
	observedAt time.Time
	kr, us     strategyResultMarketAuthority
}

func (pair strategyResultAuthorityPair) forMarket(market StrategyMarket) strategyResultMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyResultMarketAuthority{market: market}
}

type strategyRiskMarketAuthority struct {
	market   StrategyMarket
	bundle   riskbucket.RiskSnapshotAuthorityBundle
	snapshot StrategyRiskMarketSnapshot
}

type strategyRiskAuthorityPair struct {
	observedAt time.Time
	kr, us     strategyRiskMarketAuthority
}

func (pair strategyRiskAuthorityPair) forMarket(market StrategyMarket) strategyRiskMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyRiskMarketAuthority{market: market}
}

func (pair strategyRiskAuthorityPair) Snapshot() PairedStrategyRiskSnapshot {
	return PairedStrategyRiskSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type strategyRiskAuthorityLoader struct {
	configDir, journalPath, accountID, accountCurrency string
	observedAt                                         time.Time
	digests                                            map[StrategyMarket]string
	keyID                                              string
	key                                                ed25519.PublicKey
}

func newStrategyRiskAuthorityLoader(configDir, journalPath, accountID, accountCurrency string, observedAt time.Time,
	getenv func(string) string,
) *strategyRiskAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	encoded := strings.TrimSpace(getenv(strategyRiskPublicKeyEnv))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		key = nil
	}
	return &strategyRiskAuthorityLoader{configDir: filepath.Clean(strings.TrimSpace(configDir)),
		journalPath: filepath.Clean(strings.TrimSpace(journalPath)), accountID: strings.TrimSpace(accountID),
		accountCurrency: strings.ToUpper(strings.TrimSpace(accountCurrency)), observedAt: observedAt.UTC(),
		digests: map[StrategyMarket]string{StrategyMarketKR: strings.TrimSpace(getenv(strategyRiskKRManifestDigestEnv)),
			StrategyMarketUS: strings.TrimSpace(getenv(strategyRiskUSManifestDigestEnv))},
		keyID: strings.TrimSpace(getenv(strategyRiskKeyIDEnv)), key: ed25519.PublicKey(key)}
}

func (loader *strategyRiskAuthorityLoader) collect(ctx context.Context, results strategyResultAuthorityPair, fx strategyFXAuthorityPair) strategyRiskAuthorityPair {
	if loader == nil || ctx == nil || loader.observedAt.IsZero() || !results.observedAt.Equal(loader.observedAt) || !fx.observedAt.Equal(loader.observedAt) {
		return failedStrategyRiskPair(loaderTime(loader), StrategyRiskInternalFailure)
	}
	type outcome struct {
		market    StrategyMarket
		authority strategyRiskMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			result := strategyRiskMarketAuthority{market: market, snapshot: StrategyRiskMarketSnapshot{Market: market, Reason: StrategyRiskInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						result = strategyRiskMarketAuthority{market: market, snapshot: StrategyRiskMarketSnapshot{Market: market, Reason: StrategyRiskInternalFailure}}
					}
				}()
				result = loader.collectMarket(ctx, market, results.forMarket(market), fx.forMarket(market))
			}()
			outcomes <- outcome{market: market, authority: result}
		}()
	}
	pair := strategyRiskAuthorityPair{observedAt: loader.observedAt}
	for index := 0; index < 2; index++ {
		value := <-outcomes
		if value.market == StrategyMarketKR {
			pair.kr = value.authority
		} else {
			pair.us = value.authority
		}
	}
	return pair
}

func (loader *strategyRiskAuthorityLoader) collectMarket(ctx context.Context, market StrategyMarket, result strategyResultMarketAuthority,
	fx strategyFXMarketAuthority,
) strategyRiskMarketAuthority {
	fail := func(reason StrategyRiskReason) strategyRiskMarketAuthority {
		return strategyRiskMarketAuthority{market: market, snapshot: StrategyRiskMarketSnapshot{Market: market, Reason: reason}}
	}
	if !result.ready {
		return fail(StrategyRiskLaneNotReady)
	}
	if !fx.snapshot.Ready || !fx.read.valid {
		return fail(StrategyRiskFXNotReady)
	}
	bucketMarket := riskbucket.MarketKR
	if market == StrategyMarketUS {
		bucketMarket = riskbucket.MarketUS
	}
	bundle, err := riskbucket.LoadProductionRiskSnapshotAuthority(ctx, riskbucket.ProductionRiskSnapshotConfig{
		ConfigDir: loader.configDir, JournalPath: loader.journalPath, Market: bucketMarket, AccountID: loader.accountID,
		AccountCurrency: loader.accountCurrency, ManifestDigest: loader.digests[market], TrustedKeyID: loader.keyID,
		TrustedKey: loader.key, ObservedAt: loader.observedAt,
	}, riskbucket.ProductionRiskSnapshotInput{Result: result.result, FX: fx.read.evidence})
	if err != nil {
		return fail(StrategyRiskAuthorityUnavailable)
	}
	scope := bundle.Scope()
	if string(scope.Market) != string(market) || scope.AccountID != loader.accountID || !scope.AsOf.Equal(loader.observedAt) || len(bundle.Entries()) != 5 {
		return fail(StrategyRiskAuthorityUnavailable)
	}
	return strategyRiskMarketAuthority{market: market, bundle: bundle, snapshot: StrategyRiskMarketSnapshot{Market: market, Ready: true,
		Reason: StrategyRiskReady, Horizon: string(scope.Horizon), StrategyRiskID: scope.StrategyRiskID, Sector: scope.Sector,
		Symbol: scope.Symbol, BundleDigest: bundle.Digest(), BucketCount: len(bundle.Entries())}}
}

func failedStrategyRiskPair(observedAt time.Time, reason StrategyRiskReason) strategyRiskAuthorityPair {
	market := func(value StrategyMarket) strategyRiskMarketAuthority {
		return strategyRiskMarketAuthority{market: value, snapshot: StrategyRiskMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyRiskAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}

func loaderTime(loader *strategyRiskAuthorityLoader) time.Time {
	if loader == nil {
		return time.Time{}
	}
	return loader.observedAt
}
