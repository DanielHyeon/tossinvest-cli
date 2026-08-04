//go:build tossos_testseams

package engine

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	_ "modernc.org/sqlite"
)

type riskLoaderFee struct {
	FixedBaseMinor, PerUnitBaseMinor, MinimumBaseMinor string
	Version, Digest                                    string
}

func (value riskLoaderFee) MarshalJSON() ([]byte, error) {
	type wire struct {
		FixedBaseMinor   string `json:"fixed_base_minor"`
		PerUnitBaseMinor string `json:"per_unit_base_minor"`
		MinimumBaseMinor string `json:"minimum_base_minor"`
		Version          string `json:"version"`
		Digest           string `json:"digest"`
	}
	return json.Marshal(wire{value.FixedBaseMinor, value.PerUnitBaseMinor, value.MinimumBaseMinor, value.Version, value.Digest})
}

type riskLoaderStrategy struct {
	LaneID      string             `json:"lane_id"`
	LaneVersion string             `json:"lane_version"`
	Horizon     riskbucket.Horizon `json:"horizon"`
	RiskID      string             `json:"risk_id"`
	RiskVersion string             `json:"risk_version"`
	LimitMinor  string             `json:"limit_minor"`
}
type riskLoaderSymbol struct {
	Symbol           string `json:"symbol"`
	Sector           string `json:"sector"`
	SectorLimitMinor string `json:"sector_limit_minor"`
	SymbolLimitMinor string `json:"symbol_limit_minor"`
}
type riskLoaderBody struct {
	SchemaVersion      string                        `json:"schema_version"`
	Domain             string                        `json:"domain"`
	SignatureAlgorithm string                        `json:"signature_algorithm"`
	KeyID              string                        `json:"key_id"`
	Generation         uint64                        `json:"generation"`
	Market             riskbucket.Market             `json:"market"`
	AccountID          string                        `json:"account_id"`
	AccountCurrency    string                        `json:"account_currency"`
	QuoteCurrency      string                        `json:"quote_currency"`
	PolicyVersion      string                        `json:"policy_version"`
	Approver           string                        `json:"approver"`
	ObservedAt         string                        `json:"observed_at"`
	FreshUntil         string                        `json:"fresh_until"`
	Revoked            bool                          `json:"revoked"`
	Fee                riskLoaderFee                 `json:"fee"`
	HorizonLimits      map[riskbucket.Horizon]string `json:"horizon_limits"`
	MarketLimitMinor   string                        `json:"market_limit_minor"`
	Strategies         []riskLoaderStrategy          `json:"strategies"`
	Symbols            []riskLoaderSymbol            `json:"symbols"`
}
type riskLoaderManifest struct {
	riskLoaderBody
	Signature string `json:"signature"`
}

func TestStrategyRiskAuthorityLoaderPairedKRUSSameWave(t *testing.T) {
	fixture := newStrategyRiskLoaderFixture(t)
	pair := fixture.loader.collect(context.Background(), fixture.results, fixture.fx)
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		got := pair.Snapshot().For(market)
		if !got.Ready || got.Reason != StrategyRiskReady || got.BucketCount != 5 || got.BundleDigest == "" || got.Symbol == "" {
			t.Fatalf("%s risk snapshot=%+v", market, got)
		}
	}
}

func TestStrategyRiskAuthorityLoaderPreservesPeerOnMarketFailure(t *testing.T) {
	fixture := newStrategyRiskLoaderFixture(t)
	if err := os.Chmod(filepath.Join(fixture.dir, riskbucket.ProductionRiskPolicyFileName(riskbucket.MarketKR)), 0o600); err != nil {
		t.Fatal(err)
	}
	pair := fixture.loader.collect(context.Background(), fixture.results, fixture.fx).Snapshot()
	if pair.KR.Ready || pair.KR.Reason != StrategyRiskAuthorityUnavailable {
		t.Fatalf("KR=%+v", pair.KR)
	}
	if !pair.US.Ready || pair.US.Reason != StrategyRiskReady {
		t.Fatalf("US peer=%+v", pair.US)
	}
}

func TestStrategyRiskAuthorityLoaderSkipsFilesUntilLaneAndFXReady(t *testing.T) {
	fixture := newStrategyRiskLoaderFixture(t)
	if err := os.Remove(filepath.Join(fixture.dir, riskbucket.ProductionRiskPolicyFileName(riskbucket.MarketKR))); err != nil {
		t.Fatal(err)
	}
	fixture.results.kr.ready = false
	fixture.fx.us = strategyFXMarketAuthority{market: StrategyMarketUS, snapshot: StrategyFXMarketSnapshot{Market: StrategyMarketUS}}
	pair := fixture.loader.collect(context.Background(), fixture.results, fixture.fx).Snapshot()
	if pair.KR.Reason != StrategyRiskLaneNotReady || pair.US.Reason != StrategyRiskFXNotReady {
		t.Fatalf("pair=%+v", pair)
	}
}

type strategyRiskLoaderFixture struct {
	dir     string
	loader  *strategyRiskAuthorityLoader
	results strategyResultAuthorityPair
	fx      strategyFXAuthorityPair
}

func newStrategyRiskLoaderFixture(t *testing.T) strategyRiskLoaderFixture {
	t.Helper()
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.db")
	createStrategyRiskLoaderJournal(t, journalPath)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]string{strategyRiskKeyIDEnv: "risk-policy-key-1", strategyRiskPublicKeyEnv: base64.StdEncoding.EncodeToString(public)}
	results := strategyResultAuthorityPair{observedAt: now}
	fxPair := strategyFXAuthorityPair{observedAt: now}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		descriptor := riskLoaderDescriptor(t, market)
		symbol, entry, stop, target, quote, rate, haircut := "005930", "100", "95", "120", "KRW", "1", "1"
		bucketMarket := riskbucket.MarketKR
		if market == StrategyMarketUS {
			symbol, entry, stop, target, quote, rate, haircut, bucketMarket = "AAPL", "10000", "9500", "12000", "USD", "1400", "1.05", riskbucket.MarketUS
		}
		result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, "acct-risk-loader", symbol,
			"campaign-risk-loader-"+strings.ToLower(string(market)), 8,
			entry, stop, target, now.Add(-time.Second), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		fx, err := officialfx.EvidenceForAuthorityTest(quote, "KRW", rate, haircut, now.Add(-time.Minute), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		body := riskLoaderBody{SchemaVersion: "strategy-risk-bucket-policy:v1", Domain: "TossOS/strategy-risk-bucket-policy/ed25519/v1",
			SignatureAlgorithm: "Ed25519", KeyID: env[strategyRiskKeyIDEnv], Generation: 1, Market: bucketMarket,
			AccountID: "acct-risk-loader", AccountCurrency: "KRW", QuoteCurrency: quote, PolicyVersion: "risk-policy-v1", Approver: "risk-committee",
			ObservedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano),
			Fee:           riskLoaderFee{FixedBaseMinor: "0", PerUnitBaseMinor: "1", MinimumBaseMinor: "1", Version: "fee-v1", Digest: riskLoaderDigest("fee-" + string(market))},
			HorizonLimits: map[riskbucket.Horizon]string{riskbucket.HorizonShort: "10000000", riskbucket.HorizonMedium: "20000000"}, MarketLimitMinor: "10000000",
			Strategies: []riskLoaderStrategy{{LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, Horizon: riskbucket.HorizonShort,
				RiskID: "continuation", RiskVersion: "continuation-risk-v1", LimitMinor: "5000000"}},
			Symbols: []riskLoaderSymbol{{Symbol: symbol, Sector: "technology", SectorLimitMinor: "3000000", SymbolLimitMinor: "2000000"}}}
		bodyJSON, _ := json.Marshal(body)
		manifest := riskLoaderManifest{riskLoaderBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, bodyJSON))}
		data, _ := json.Marshal(manifest)
		if err := os.WriteFile(filepath.Join(dir, riskbucket.ProductionRiskPolicyFileName(bucketMarket)), data, 0o400); err != nil {
			t.Fatal(err)
		}
		digestEnv := strategyRiskKRManifestDigestEnv
		if market == StrategyMarketUS {
			digestEnv = strategyRiskUSManifestDigestEnv
			results.us = strategyResultMarketAuthority{market: market, ready: true, result: result}
			fxPair.us = strategyFXMarketAuthority{market: market, read: strategyFXRead{valid: true, quoteCurrency: quote, accountCurrency: "KRW", digest: fx.Digest(), evidence: fx},
				snapshot: StrategyFXMarketSnapshot{Market: market, Ready: true, Reason: StrategyFXReady}}
		} else {
			results.kr = strategyResultMarketAuthority{market: market, ready: true, result: result}
			fxPair.kr = strategyFXMarketAuthority{market: market, read: strategyFXRead{valid: true, quoteCurrency: quote, accountCurrency: "KRW", digest: fx.Digest(), evidence: fx},
				snapshot: StrategyFXMarketSnapshot{Market: market, Ready: true, Reason: StrategyFXReady}}
		}
		env[digestEnv] = riskLoaderDigestBytes(data)
	}
	loader := newStrategyRiskAuthorityLoader(dir, journalPath, "acct-risk-loader", "KRW", now, func(key string) string { return env[key] })
	return strategyRiskLoaderFixture{dir: dir, loader: loader, results: results, fx: fxPair}
}

func riskLoaderDescriptor(t *testing.T, market StrategyMarket) strategyflow.Descriptor {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if StrategyMarket(descriptor.Market) == market && descriptor.Horizon == strategyrouter.HorizonShort {
			return descriptor
		}
	}
	t.Fatal("missing descriptor")
	return strategyflow.Descriptor{}
}

func createStrategyRiskLoaderJournal(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{`PRAGMA user_version=27`,
		`CREATE TABLE risk_bucket_policies(bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT,record_digest TEXT,PRIMARY KEY(bucket_dimension,bucket_value,policy_version))`,
		`CREATE TABLE risk_bucket_snapshots(snapshot_id TEXT PRIMARY KEY,bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT)`,
		`CREATE TABLE risk_bucket_reservations(reservation_id TEXT PRIMARY KEY,account_ref TEXT,bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT,snapshot_id TEXT,held_minor TEXT,filled_minor TEXT,state TEXT,risk_overage_latched INTEGER,unknown_actual_latched INTEGER)`,
		`CREATE TABLE risk_bucket_scope_latches(account_ref TEXT,market TEXT,symbol TEXT)`} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
}

func riskLoaderDigest(value string) string { return riskLoaderDigestBytes([]byte(value)) }
func riskLoaderDigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
