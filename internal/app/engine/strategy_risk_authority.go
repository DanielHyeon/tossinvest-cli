package engine

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
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
	// scoped 는 서명 활성화된 시장의 소유자 범위별 결과다(a112 5.2.2.2, 조정자 순서). 활성화 없는 시장은 비어 있고 위의 result 하나다.
	scoped []strategyflow.Result
}

// results 는 이 시장의 범위별 결과다 — 활성화 없는 시장은 준비된 result 하나(오늘 그대로).
func (authority strategyResultMarketAuthority) results() []strategyflow.Result {
	if len(authority.scoped) != 0 {
		return authority.scoped
	}
	if authority.ready {
		return []strategyflow.Result{authority.result}
	}
	return nil
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
	// scopes 는 소유자 범위별 위험 권한이다(a112 5.2.2.2). 활성화 없는 시장은 원소 하나.
	scopes []strategyRiskScopeAuthority
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
	// a112 5.2.2.2: 결과 권한의 범위마다 위험 번들 하나(활성화 없는 시장은 하나 — 오늘 그대로). 한 범위의 적재 실패는 그 범위만 준비 안 됨
	// (Manager 판정 J3); 준비된 범위가 하나도 없으면 시장 전체가 준비 안 됨(오늘의 사유 그대로).
	scopes := make([]strategyRiskScopeAuthority, 0, len(result.results()))
	for _, scoped := range result.results() {
		key, keyed := strategyOwnerKeyOf(scoped.Lineage)
		entry := strategyRiskScopeAuthority{key: key, reason: StrategyRiskAuthorityUnavailable,
			cause: errors.New("production risk owner scope key invalid")}
		if keyed {
			bundle, err := riskbucket.LoadProductionRiskSnapshotAuthority(ctx, riskbucket.ProductionRiskSnapshotConfig{
				ConfigDir: loader.configDir, JournalPath: loader.journalPath, Market: bucketMarket, AccountID: loader.accountID,
				AccountCurrency: loader.accountCurrency, ManifestDigest: loader.digests[market], TrustedKeyID: loader.keyID,
				TrustedKey: loader.key, ObservedAt: loader.observedAt,
				// a127 D2: 이 빌드가 이해하는 원장 스키마(상수) — 적재기는 원장 user_version 이 이 값과 같을 때만 읽음.
				JournalSchemaVersion: journal.SchemaVersion,
			}, riskbucket.ProductionRiskSnapshotInput{Result: scoped, FX: fx.read.evidence})
			scope := bundle.Scope()
			switch {
			case err != nil:
				// 원인을 접지 않고 운반(5.2.2.2 리뷰 수리 — 범위 국소 거절인지 결함인지는 1차 레그가 이 원인의 신원으로 가름).
				entry.cause = err
			case string(scope.Market) == string(market) && scope.AccountID == loader.accountID &&
				scope.AsOf.Equal(loader.observedAt) && len(bundle.Entries()) == 5:
				entry.bundle, entry.ready, entry.reason, entry.cause = bundle, true, StrategyRiskReady, nil
			default:
				entry.cause = errors.New("production risk bundle does not match the loader's market, account or observation")
			}
		}
		scopes = append(scopes, entry)
	}
	return strategyRiskMarketFromScopes(market, scopes)
}

// strategyRiskMarketFromScopes 는 범위별 위험 권한에서 시장 권한을 만든다. 시장 칸(bundle · snapshot)은 첫 준비된 범위의 것 — 범위가
// 하나면 오늘과 같은 값이다. 범위가 둘 이상이면 BundleDigest 는 준비된 범위들의 digest 를 조정자 순서로 묶은 값(worker 증거 digest 가 범위
// 집합을 담게).
func strategyRiskMarketFromScopes(market StrategyMarket, scopes []strategyRiskScopeAuthority) strategyRiskMarketAuthority {
	var first *strategyRiskScopeAuthority
	ready := make([]string, 0, len(scopes))
	for index := range scopes {
		if scopes[index].ready {
			if first == nil {
				first = &scopes[index]
			}
			ready = append(ready, scopes[index].bundle.Digest())
		}
	}
	if first == nil {
		return strategyRiskMarketAuthority{market: market, scopes: scopes,
			snapshot: StrategyRiskMarketSnapshot{Market: market, Reason: StrategyRiskAuthorityUnavailable}}
	}
	scope := first.bundle.Scope()
	digest := first.bundle.Digest()
	if len(scopes) > 1 {
		digest = strategyWorkerEvidenceDigest(ready...)
	}
	return strategyRiskMarketAuthority{market: market, bundle: first.bundle, scopes: scopes, snapshot: StrategyRiskMarketSnapshot{Market: market,
		Ready: true, Reason: StrategyRiskReady, Horizon: string(scope.Horizon), StrategyRiskID: scope.StrategyRiskID, Sector: scope.Sector,
		Symbol: scope.Symbol, BundleDigest: digest, BucketCount: len(first.bundle.Entries())}}
}

// forScope 는 그 소유자 범위의 준비된 위험 번들이다. 없거나 준비 안 됐으면 false — 봉투 값으로 폴백하지 않는다.
func (authority strategyRiskMarketAuthority) forScope(key strategyrouter.OwnerKey) (riskbucket.RiskSnapshotAuthorityBundle, bool) {
	for _, scope := range authority.scopes {
		if scope.key == key && scope.ready {
			return scope.bundle, true
		}
	}
	return riskbucket.RiskSnapshotAuthorityBundle{}, false
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
