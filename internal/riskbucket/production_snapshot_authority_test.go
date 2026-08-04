//go:build tossos_testseams

package riskbucket

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

type productionRiskFixture struct {
	config   ProductionRiskSnapshotConfig
	input    ProductionRiskSnapshotInput
	body     productionRiskPolicyBody
	private  ed25519.PrivateKey
	filePath string
}

func TestProductionRiskSnapshotAuthorityPairedKRUSSameWave(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	fixtures := []productionRiskFixture{
		newProductionRiskFixture(t, MarketKR, now),
		newProductionRiskFixture(t, MarketUS, now),
	}
	for _, fixture := range fixtures {
		bundle, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
		if err != nil {
			t.Fatalf("%s load: %v", fixture.config.Market, err)
		}
		scope := bundle.Scope()
		if scope.Market != fixture.config.Market || scope.AccountID != fixture.config.AccountID || scope.AsOf != now ||
			bundle.Policy().AccountCurrency != "KRW" || len(bundle.Entries()) != 5 || bundle.Digest() == "" {
			t.Fatalf("%s bundle mismatch: scope=%+v entries=%d digest=%q", fixture.config.Market, scope, len(bundle.Entries()), bundle.Digest())
		}
		for index, entry := range bundle.Entries() {
			if entry.Bucket.Key.Dimension != RequiredDimensionOrder()[index] || entry.Bucket.FilledMinor != "0" || entry.Bucket.HeldMinor != "0" {
				t.Fatalf("%s entry %d mismatch: %+v", fixture.config.Market, index, entry)
			}
		}
		if err := bundle.Validate(scope); err != nil {
			t.Fatalf("%s validate: %v", fixture.config.Market, err)
		}
	}
}

func TestProductionRiskSnapshotAuthorityFailureIsMarketLocal(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	kr := newProductionRiskFixture(t, MarketKR, now)
	us := newProductionRiskFixture(t, MarketUS, now)
	if err := os.Chmod(kr.filePath, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), kr.config, kr.input); !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("KR wrong mode error=%v", err)
	}
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), us.config, us.input); err != nil {
		t.Fatalf("US peer authority was contaminated: %v", err)
	}

	kr = newProductionRiskFixture(t, MarketKR, now)
	us = newProductionRiskFixture(t, MarketUS, now)
	us.config.ManifestDigest = productionRiskDigest([]byte("wrong"))
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), us.config, us.input); !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("US wrong digest error=%v", err)
	}
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), kr.config, kr.input); err != nil {
		t.Fatalf("KR peer authority was contaminated: %v", err)
	}
}

func TestProductionRiskSnapshotAuthorityReadsExactUsageAndRefusesLatch(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	fixture := newProductionRiskFixture(t, MarketKR, now)
	db := openProductionRiskDB(t, fixture.config.JournalPath)
	values := map[Dimension]string{DimensionHorizon: "SHORT", DimensionMarket: "KR", DimensionStrategy: "continuation",
		DimensionSector: "technology", DimensionSymbol: "005930"}
	for index, dimension := range RequiredDimensionOrder() {
		version := "historical-v1"
		snapshotID := "prior-snapshot-" + string(dimension)
		if _, err := db.Exec(`INSERT INTO risk_bucket_policies(bucket_dimension,bucket_value,policy_version,record_digest) VALUES(?,?,?,?)`,
			string(dimension), values[dimension], version, "policy-record-"+string(dimension)); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO risk_bucket_snapshots(snapshot_id,bucket_dimension,bucket_value,policy_version) VALUES(?,?,?,?)`,
			snapshotID, string(dimension), values[dimension], version); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,account_ref,bucket_dimension,bucket_value,policy_version,snapshot_id,held_minor,filled_minor,state,risk_overage_latched,unknown_actual_latched) VALUES(?,?,?,?,?,?,?,?,?,0,0)`,
			"prior-reservation-"+string(dimension), fixture.config.AccountID, string(dimension), values[dimension], version, snapshotID,
			index+1, (index+1)*10, "HELD"); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	bundle, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	for index, entry := range bundle.Entries() {
		if entry.Bucket.HeldMinor != string(rune('1'+index)) {
			t.Fatalf("held[%d]=%s", index, entry.Bucket.HeldMinor)
		}
	}
	db = openProductionRiskDB(t, fixture.config.JournalPath)
	if _, err := db.Exec(`UPDATE risk_bucket_reservations SET unknown_actual_latched=1 WHERE bucket_dimension='sector'`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), fixture.config, fixture.input); !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("latched load error=%v", err)
	}
}

func TestProductionRiskSnapshotAuthorityRejectsCrossMarketAndSymlink(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	kr := newProductionRiskFixture(t, MarketKR, now)
	us := newProductionRiskFixture(t, MarketUS, now)
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), kr.config, us.input); !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("cross-market result/FX error=%v", err)
	}
	realPath := kr.filePath + ".real"
	if err := os.Rename(kr.filePath, realPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realPath, kr.filePath); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionRiskSnapshotAuthority(context.Background(), kr.config, kr.input); !errors.Is(err, ErrProductionRiskSnapshotUnavailable) {
		t.Fatalf("symlink policy error=%v", err)
	}
}

func newProductionRiskFixture(t *testing.T, market Market, now time.Time) productionRiskFixture {
	t.Helper()
	dir := t.TempDir()
	journalPath := filepath.Join(dir, "journal.db")
	createProductionRiskDB(t, journalPath)
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := productionRiskDescriptor(t, market)
	symbol := "005930"
	entry, stop, target := "100", "95", "120"
	if market == MarketUS {
		symbol, entry, stop, target = "AAPL", "10000", "9500", "12000"
	}
	result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, "acct-production-risk", symbol, "campaign-production-risk", 8,
		entry, stop, target, now.Add(-time.Second), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	quote := productionRiskQuoteCurrency(market)
	fxRate, haircut := "1", "1"
	if market == MarketUS {
		fxRate, haircut = "1400", "1.05"
	}
	fx, err := officialfx.EvidenceForAuthorityTest(quote, "KRW", fxRate, haircut, now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	body := productionRiskPolicyBody{SchemaVersion: productionRiskPolicySchema, Domain: productionRiskPolicyDomain,
		SignatureAlgorithm: productionRiskPolicyAlgorithm, KeyID: "risk-policy-key-1", Generation: 1, Market: market,
		AccountID: "acct-production-risk", AccountCurrency: "KRW", QuoteCurrency: quote, PolicyVersion: "risk-policy-v1",
		Approver: "risk-committee", ObservedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano),
		Fee:           productionRiskFeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "1", MinimumBaseMinor: "1", Version: "fee-v1", Digest: productionRiskDigest([]byte("fee-" + string(market)))},
		HorizonLimits: map[Horizon]string{HorizonShort: "10000000", HorizonMedium: "20000000"}, MarketLimitMinor: "10000000",
		Strategies: []productionRiskStrategyPolicy{{LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, Horizon: Horizon(descriptor.Horizon),
			RiskID: "continuation", RiskVersion: "continuation-risk-v1", LimitMinor: "5000000"}},
		Symbols: []productionRiskSymbolPolicy{{Symbol: symbol, Sector: "technology", SectorLimitMinor: "3000000", SymbolLimitMinor: "2000000"}}}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionRiskPolicyManifest{productionRiskPolicyBody: body,
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, bodyJSON))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dir, ProductionRiskPolicyFileName(market))
	if err := os.WriteFile(filePath, data, 0o400); err != nil {
		t.Fatal(err)
	}
	return productionRiskFixture{config: ProductionRiskSnapshotConfig{ConfigDir: dir, JournalPath: journalPath, Market: market,
		AccountID: body.AccountID, AccountCurrency: body.AccountCurrency, ManifestDigest: productionRiskDigest(data), TrustedKeyID: body.KeyID,
		TrustedKey: public, ObservedAt: now}, input: ProductionRiskSnapshotInput{Result: result, FX: fx}, body: body, private: private, filePath: filePath}
}

func productionRiskDescriptor(t *testing.T, market Market) strategyflow.Descriptor {
	t.Helper()
	for _, descriptor := range strategyflow.Descriptors() {
		if Market(descriptor.Market) == market && descriptor.Horizon == strategyrouter.HorizonShort {
			return descriptor
		}
	}
	t.Fatalf("no %s short descriptor", market)
	return strategyflow.Descriptor{}
}

func createProductionRiskDB(t *testing.T, path string) {
	t.Helper()
	db := openProductionRiskDB(t, path)
	statements := []string{
		`PRAGMA user_version=27`,
		`CREATE TABLE risk_bucket_policies(bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT,record_digest TEXT,PRIMARY KEY(bucket_dimension,bucket_value,policy_version))`,
		`CREATE TABLE risk_bucket_snapshots(snapshot_id TEXT PRIMARY KEY,bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT)`,
		`CREATE TABLE risk_bucket_reservations(reservation_id TEXT PRIMARY KEY,account_ref TEXT,bucket_dimension TEXT,bucket_value TEXT,policy_version TEXT,snapshot_id TEXT,held_minor TEXT,filled_minor TEXT,state TEXT,risk_overage_latched INTEGER,unknown_actual_latched INTEGER)`,
		`CREATE TABLE risk_bucket_scope_latches(account_ref TEXT,market TEXT,symbol TEXT)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
}

func openProductionRiskDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
