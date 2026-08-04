package strategyrouter

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
	_ "modernc.org/sqlite"
)

func TestPairedProductionRouteAuthorityLoadsExactThreeLanesIndependently(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		authority, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market])
		if err != nil {
			t.Fatalf("%s load: %v", market, err)
		}
		if authority.ManifestDigest() == "" || authority.OwnerDigest() == "" {
			t.Fatalf("%s scalar provenance missing: %+v", market, authority)
		}
		request := authority.Request()
		if len(request.Candidates) != 3 || request.Snapshot.Revision != 1 || len(request.Snapshot.Owners) != 0 {
			t.Fatalf("%s route reconstruction=%+v", market, request)
		}
		routed := Route(request)
		wantLane := map[Market]string{MarketKR: continuationlane.KRContinuationLaneID, MarketUS: continuationlane.USContinuationLaneID}[market]
		if routed.Code != RefusalNone || routed.Decision.LaneID != wantLane || routed.Decision.ExistingOwner {
			t.Fatalf("%s route=%+v", market, routed)
		}
	}
}

func TestProductionRouteAuthorityFailureIsMarketLocal(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	kr := fixture.config[MarketKR]
	kr.ManifestDigest = "sha256:" + string(make([]byte, 64))
	if _, err := LoadProductionRouteAuthority(context.Background(), kr); err == nil {
		t.Fatal("corrupt KR digest accepted")
	}
	us, err := LoadProductionRouteAuthority(context.Background(), fixture.config[MarketUS])
	if err != nil || Route(us.Request()).Code != RefusalNone {
		t.Fatalf("US was gated by KR failure: authority=%+v err=%v", us, err)
	}

	kr = fixture.config[MarketKR]
	if err := os.Chmod(filepath.Join(fixture.dir, ProductionRouteFileName(MarketKR)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionRouteAuthority(context.Background(), kr); err == nil {
		t.Fatal("wrong manifest mode accepted")
	}
	if _, err := LoadProductionRouteAuthority(context.Background(), fixture.config[MarketUS]); err != nil {
		t.Fatalf("US failed after KR mode refusal: %v", err)
	}
}

func TestProductionRouteAuthoritySelectsEverySignedSymbolScope(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, test := range []struct {
		market Market
		symbol string
	}{
		{MarketKR, "035420"},
		{MarketUS, "MSFT"},
	} {
		body := fixture.body(test.market)
		second := body.Scopes[0]
		second.Symbol = test.symbol
		for index := range second.Candidates {
			second.Candidates[index].EvidenceDigest += "-second"
			second.Candidates[index].ConfigDigest += "-second"
		}
		body.Scopes = append(body.Scopes, second)
		fixture.write(t, test.market, body)
		config := fixture.config[test.market]
		config.Symbol = test.symbol
		config.PositionGeneration = 0
		authority, err := LoadProductionRouteAuthority(context.Background(), config)
		if err != nil {
			t.Fatalf("%s %s load: %v", test.market, test.symbol, err)
		}
		request := authority.Request()
		if request.Key.Symbol != test.symbol || len(request.Candidates) != 3 || Route(request).Code != RefusalNone {
			t.Fatalf("%s %s scope=%+v", test.market, test.symbol, request)
		}
	}
}

func TestProductionRouteAuthorityBatchUsesEverySignedScopeInOneMarketSnapshot(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, test := range []struct {
		market Market
		second string
		absent string
	}{
		{MarketKR, "035420", "999999"},
		{MarketUS, "MSFT", "ZZZZ"},
	} {
		body := fixture.body(test.market)
		first := body.Scopes[0].Symbol
		second := body.Scopes[0]
		second.Symbol = test.second
		for index := range second.Candidates {
			second.Candidates[index].EvidenceDigest += "-batch-second"
			second.Candidates[index].ConfigDigest += "-batch-second"
		}
		body.Scopes = append(body.Scopes, second)
		fixture.write(t, test.market, body)
		batch, err := LoadProductionRouteAuthorityBatch(context.Background(), fixture.config[test.market], []ProductionRouteTarget{
			{Symbol: test.second}, {Symbol: first}, {Symbol: test.absent},
		})
		if err != nil {
			t.Fatalf("%s batch load: %v", test.market, err)
		}
		if batch.Len() != 2 || batch.ManifestDigest() != fixture.config[test.market].ManifestDigest {
			t.Fatalf("%s batch=%+v", test.market, batch)
		}
		for _, symbol := range []string{first, test.second} {
			authority, ok := batch.For(symbol)
			if !ok || Route(authority.Request()).Code != RefusalNone || authority.Request().Key.Symbol != symbol {
				t.Fatalf("%s symbol=%s authority=%+v ok=%v", test.market, symbol, authority, ok)
			}
		}
		if _, ok := batch.For(test.absent); ok {
			t.Fatalf("%s absent scope %s was synthesized", test.market, test.absent)
		}
	}
}

func TestProductionRouteAuthorityRestoresExactActiveOwner(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	market := MarketKR
	acquired := fixture.now.Add(-2 * time.Minute).Format(time.RFC3339Nano)
	db, err := sql.Open("sqlite", fixture.journal)
	if err != nil {
		t.Fatal(err)
	}
	prospective := "prospective-owner-kr"
	campaign := "campaign-active-kr"
	if _, err := db.Exec(`INSERT INTO position_campaigns(id,account_ref,market,symbol,lane_id,lane_version,prospective_token,actual_position_generation,state,entry_blocked) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		campaign, "acct", "KR", "005930", reversallane.KRReversalLaneID, reversallane.LaneVersionV1, prospective, 1, "ACTIVE", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,actual_generation,acquired_at,released_at,risk_overage_latched,unknown_actual_latched) VALUES(?,?,?,?,?,?,?,?,NULL,0,0)`,
		"acct", "KR", "005930", prospective, reversallane.KRReversalLaneID, campaign, "1", acquired); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fixture.journal, 0o600); err != nil {
		t.Fatal(err)
	}
	body := fixture.body(market)
	body.Scopes[0].OwnerRevision = 2
	fixture.write(t, market, body)
	authority, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market])
	if err != nil {
		t.Fatal(err)
	}
	routed := Route(authority.Request())
	if routed.Code != RefusalNone || !routed.Decision.ExistingOwner || routed.Decision.CampaignID != campaign ||
		routed.Decision.LaneID != reversallane.KRReversalLaneID {
		t.Fatalf("active owner not restored: %+v", routed)
	}

	// A prospective row without an actual generation cannot be converted into
	// the numeric position generation expected by the router.
	db, err = sql.Open("sqlite", fixture.journal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE risk_bucket_owners SET actual_generation=NULL WHERE campaign_id=?`, campaign); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(fixture.journal, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market]); err == nil {
		t.Fatal("prospective token was treated as actual position generation")
	}
}

type productionRouteFixture struct {
	dir, journal string
	now          time.Time
	public       ed25519.PublicKey
	private      ed25519.PrivateKey
	config       map[Market]ProductionRouteConfig
}

func newProductionRouteFixture(t *testing.T) *productionRouteFixture {
	t.Helper()
	dir := t.TempDir()
	journal := filepath.Join(dir, "journal.db")
	db, err := sql.Open("sqlite", journal)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`PRAGMA user_version=27`,
		`CREATE TABLE risk_bucket_owners(account_ref TEXT,market TEXT,symbol TEXT,prospective_generation TEXT,lane_id TEXT,campaign_id TEXT,actual_generation TEXT,acquired_at TEXT,released_at TEXT,risk_overage_latched INTEGER,unknown_actual_latched INTEGER)`,
		`CREATE TABLE position_campaigns(id TEXT PRIMARY KEY,account_ref TEXT,market TEXT,symbol TEXT,lane_id TEXT,lane_version TEXT,prospective_token TEXT,actual_position_generation INTEGER,state TEXT,entry_blocked INTEGER)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(journal, 0o600); err != nil {
		t.Fatal(err)
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &productionRouteFixture{dir: dir, journal: journal, now: time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC), public: public, private: private,
		config: make(map[Market]ProductionRouteConfig, 2)}
	for _, market := range []Market{MarketKR, MarketUS} {
		body := fixture.body(market)
		scope := body.Scopes[0]
		fixture.config[market] = ProductionRouteConfig{ConfigDir: dir, JournalPath: journal, AccountRef: "acct",
			Market: market, Symbol: scope.Symbol, PositionGeneration: scope.PositionGeneration, TrustedKeyID: "route-key-v1", TrustedKey: public, ObservedAt: fixture.now,
			ActivationDigest: body.ActivationDigest, CalendarGeneration: body.CalendarGeneration, CalendarDigest: body.CalendarDigest,
			SchedulerConfigVersion: body.ConfigVersion, ActivationExpiresAt: fixture.now.Add(time.Hour)}
		fixture.write(t, market, body)
	}
	return fixture
}

func (fixture *productionRouteFixture) body(market Market) productionRouteBody {
	symbol := map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market]
	timezone := map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[market]
	lanes := []productionRouteCandidate{}
	if market == MarketKR {
		lanes = []productionRouteCandidate{
			{HorizonShort, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, 30, true, StateOn, StateOn, "kr-continuation-evidence", "kr-continuation-config"},
			{HorizonShort, reversallane.KRReversalLaneID, reversallane.LaneVersionV1, 20, true, StateOn, StateOn, "kr-reversal-evidence", "kr-reversal-config"},
			{HorizonWeekly, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, 10, true, StateOn, StateOn, "kr-weekly-evidence", "kr-weekly-config"},
		}
	} else {
		lanes = []productionRouteCandidate{
			{HorizonShort, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, 30, true, StateOn, StateOn, "us-continuation-evidence", "us-continuation-config"},
			{HorizonShort, reversallane.USReversalLaneID, reversallane.LaneVersionV1, 20, true, StateOn, StateOn, "us-reversal-evidence", "us-reversal-config"},
			{HorizonWeekly, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, 10, true, StateOn, StateOn, "us-weekly-evidence", "us-weekly-config"},
		}
	}
	return productionRouteBody{SchemaVersion: productionRouteSchema, Domain: productionRouteDomain, SignatureAlgorithm: productionRouteAlgorithm,
		KeyID: "route-key-v1", Generation: 1, AccountRef: "acct", Market: market,
		MarketRevision: 1, ActivationDigest: "activation-" + string(market), ActivationExpiresAt: fixture.now.Add(time.Hour).Format(time.RFC3339Nano),
		CalendarGeneration: "calendar-generation-" + string(market), CalendarDigest: "calendar-digest-" + string(market), Timezone: timezone,
		SessionScope: "REGULAR", ConfigVersion: "scheduler-config-" + string(market), Actor: "human-approver", ObservedAt: fixture.now.Add(-time.Minute).Format(time.RFC3339Nano),
		FreshUntil: fixture.now.Add(30 * time.Minute).Format(time.RFC3339Nano), Scopes: []productionRouteScope{{Symbol: symbol, PositionGeneration: 1, OwnerRevision: 1, Candidates: lanes}}}
}

func (fixture *productionRouteFixture) write(t *testing.T, market Market, body productionRouteBody) {
	t.Helper()
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionRouteManifest{productionRouteBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(fixture.private, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fixture.dir, ProductionRouteFileName(market))
	if _, err := os.Stat(path); err == nil {
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}
	config := fixture.config[market]
	config.ManifestDigest = productionRouteDigest(data)
	fixture.config[market] = config
}
