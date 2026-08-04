package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

const (
	strategyFXManifestDigestEnv = "TOSSOS_FX_RISK_POLICY_MANIFEST_SHA256"
	strategyFXKeyIDEnv          = "TOSSOS_FX_RISK_POLICY_KEY_ID"
	strategyFXPublicKeyEnv      = "TOSSOS_FX_RISK_POLICY_PUBLIC_KEY_BASE64"
)

type StrategyFXReason string

const (
	StrategyFXReady                StrategyFXReason = "READY"
	StrategyFXCandidateNotReady    StrategyFXReason = "CANDIDATE_NOT_READY"
	StrategyFXAuthorityUnavailable StrategyFXReason = "FX_AUTHORITY_UNAVAILABLE"
	StrategyFXInternalFailure      StrategyFXReason = "INTERNAL_FAILURE"
)

type StrategyFXMarketSnapshot struct {
	Market          StrategyMarket
	Ready           bool
	Reason          StrategyFXReason
	QuoteCurrency   string
	AccountCurrency string
	Digest          string
}

type PairedStrategyFXSnapshot struct {
	ObservedAt time.Time
	KR         StrategyFXMarketSnapshot
	US         StrategyFXMarketSnapshot
}

func (snapshot PairedStrategyFXSnapshot) For(market StrategyMarket) StrategyFXMarketSnapshot {
	if market == StrategyMarketKR {
		return snapshot.KR
	}
	if market == StrategyMarketUS {
		return snapshot.US
	}
	return StrategyFXMarketSnapshot{Market: market, Reason: StrategyFXInternalFailure}
}

type strategyFXRead struct {
	valid                          bool
	quoteCurrency, accountCurrency string
	digest                         string
	evidence                       officialfx.Evidence
}

type strategyFXCollector interface {
	collectKR(context.Context) (strategyFXRead, error)
	collectUS(context.Context) (strategyFXRead, error)
}

type productionStrategyFXCollector struct {
	service         *officialfx.ProductionAuthorityService
	at              time.Time
	accountCurrency string
}

func (collector productionStrategyFXCollector) collectKR(ctx context.Context) (strategyFXRead, error) {
	return collector.collect(ctx, StrategyMarketKR)
}

func (collector productionStrategyFXCollector) collectUS(ctx context.Context) (strategyFXRead, error) {
	return collector.collect(ctx, StrategyMarketUS)
}

func (collector productionStrategyFXCollector) collect(ctx context.Context, market StrategyMarket) (strategyFXRead, error) {
	var evidence officialfx.Evidence
	var err error
	if market == StrategyMarketKR {
		evidence, err = collector.service.CollectKR(ctx)
	} else {
		evidence, err = collector.service.CollectUS(ctx)
	}
	if err != nil {
		return strategyFXRead{}, err
	}
	quote := "KRW"
	if market == StrategyMarketUS {
		quote = "USD"
	}
	if _, err := evidence.EvidenceAt(collector.at, quote, collector.accountCurrency); err != nil {
		return strategyFXRead{}, err
	}
	return strategyFXRead{valid: true, quoteCurrency: quote, accountCurrency: collector.accountCurrency,
		digest: evidence.Digest(), evidence: evidence}, nil
}

type strategyFXMarketAuthority struct {
	market   StrategyMarket
	read     strategyFXRead
	snapshot StrategyFXMarketSnapshot
}

type strategyFXAuthorityPair struct {
	observedAt time.Time
	kr         strategyFXMarketAuthority
	us         strategyFXMarketAuthority
}

func (pair strategyFXAuthorityPair) forMarket(market StrategyMarket) strategyFXMarketAuthority {
	if market == StrategyMarketKR {
		return pair.kr
	}
	if market == StrategyMarketUS {
		return pair.us
	}
	return strategyFXMarketAuthority{market: market}
}

func (pair strategyFXAuthorityPair) Snapshot() PairedStrategyFXSnapshot {
	return PairedStrategyFXSnapshot{ObservedAt: pair.observedAt, KR: pair.kr.snapshot, US: pair.us.snapshot}
}

type strategyFXAuthorityLoader struct {
	config    officialfx.ProductionAuthorityConfig
	collector strategyFXCollector
}

func newStrategyFXAuthorityLoader(configDir, accountID, accountCurrency string, observedAt time.Time,
	client *official.Client, getenv func(string) string,
) *strategyFXAuthorityLoader {
	if getenv == nil {
		getenv = os.Getenv
	}
	encoded := strings.TrimSpace(getenv(strategyFXPublicKeyEnv))
	key, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || base64.StdEncoding.EncodeToString(key) != encoded || len(key) != ed25519.PublicKeySize {
		key = nil
	}
	config := officialfx.ProductionAuthorityConfig{
		ConfigDir: filepath.Clean(strings.TrimSpace(configDir)), AccountID: strings.TrimSpace(accountID),
		AccountCurrency: strings.ToUpper(strings.TrimSpace(accountCurrency)),
		ManifestDigest:  strings.TrimSpace(getenv(strategyFXManifestDigestEnv)),
		TrustedKeyID:    strings.TrimSpace(getenv(strategyFXKeyIDEnv)), TrustedKey: ed25519.PublicKey(key),
		Now: func() time.Time { return observedAt.UTC() },
	}
	service := officialfx.NewProductionAuthorityService(config, client)
	return &strategyFXAuthorityLoader{config: config, collector: productionStrategyFXCollector{
		service: service, at: observedAt.UTC(), accountCurrency: config.AccountCurrency,
	}}
}

func (loader *strategyFXAuthorityLoader) collect(ctx context.Context, candidates strategyCandidateAuthorityPair) strategyFXAuthorityPair {
	if loader == nil || ctx == nil || loader.collector == nil || candidates.observedAt.IsZero() ||
		loader.config.Now == nil || !loader.config.Now().Equal(candidates.observedAt) {
		return failedStrategyFXPair(candidates.observedAt, StrategyFXInternalFailure)
	}
	type outcome struct {
		market    StrategyMarket
		authority strategyFXMarketAuthority
	}
	outcomes := make(chan outcome, 2)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		market := market
		go func() {
			result := strategyFXMarketAuthority{market: market,
				snapshot: StrategyFXMarketSnapshot{Market: market, Reason: StrategyFXInternalFailure}}
			func() {
				defer func() {
					if recover() != nil {
						result = strategyFXMarketAuthority{market: market,
							snapshot: StrategyFXMarketSnapshot{Market: market, Reason: StrategyFXInternalFailure}}
					}
				}()
				result = loader.collectMarket(ctx, market, candidates.forMarket(market))
			}()
			outcomes <- outcome{market: market, authority: result}
		}()
	}
	pair := strategyFXAuthorityPair{observedAt: candidates.observedAt}
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

func (loader *strategyFXAuthorityLoader) collectMarket(ctx context.Context, market StrategyMarket, candidate strategyCandidateMarketAuthority) strategyFXMarketAuthority {
	fail := func(reason StrategyFXReason) strategyFXMarketAuthority {
		return strategyFXMarketAuthority{market: market,
			snapshot: StrategyFXMarketSnapshot{Market: market, Reason: reason}}
	}
	if !candidate.snapshot.Ready {
		return fail(StrategyFXCandidateNotReady)
	}
	var read strategyFXRead
	var err error
	if market == StrategyMarketKR {
		read, err = loader.collector.collectKR(ctx)
	} else {
		read, err = loader.collector.collectUS(ctx)
	}
	quote := "KRW"
	if market == StrategyMarketUS {
		quote = "USD"
	}
	if err != nil || !read.valid || read.quoteCurrency != quote ||
		read.accountCurrency != loader.config.AccountCurrency || !validStrategyFXDigest(read.digest) {
		return fail(StrategyFXAuthorityUnavailable)
	}
	return strategyFXMarketAuthority{market: market, read: read,
		snapshot: StrategyFXMarketSnapshot{Market: market, Ready: true, Reason: StrategyFXReady,
			QuoteCurrency: quote, AccountCurrency: read.accountCurrency, Digest: read.digest}}
}

func validStrategyFXDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func failedStrategyFXPair(observedAt time.Time, reason StrategyFXReason) strategyFXAuthorityPair {
	market := func(value StrategyMarket) strategyFXMarketAuthority {
		return strategyFXMarketAuthority{market: value, snapshot: StrategyFXMarketSnapshot{Market: value, Reason: reason}}
	}
	return strategyFXAuthorityPair{observedAt: observedAt, kr: market(StrategyMarketKR), us: market(StrategyMarketUS)}
}
