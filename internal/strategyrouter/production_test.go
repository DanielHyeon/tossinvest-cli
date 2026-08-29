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

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
	_ "modernc.org/sqlite"
)

func TestPairedProductionRouteAuthorityLoadsExactFourLanesIndependently(t *testing.T) {
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
		if len(request.Candidates) != 4 || request.Snapshot.Revision != 1 || len(request.Snapshot.Owners) != 0 {
			t.Fatalf("%s route reconstruction=%+v", market, request)
		}
		// 생산 권한은 RouteSet 이다. 평가 전에 한 가족을 고르지 않고
		// 자격 있는 네 가족을 전부 내보내는지 본다.
		routed := RouteSet(request)
		if routed.Code != RefusalNone || !routed.Valid() || len(routed.Decisions) != 4 || routed.ExistingOwner {
			t.Fatalf("%s route set=%+v", market, routed)
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
	if err != nil || RouteSet(us.Request()).Code != RefusalNone {
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
		if request.Key.Symbol != test.symbol || len(request.Candidates) != 4 || RouteSet(request).Code != RefusalNone {
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
			if !ok || RouteSet(authority.Request()).Code != RefusalNone || authority.Request().Key.Symbol != symbol {
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
	// 후보는 필드 이름으로 적는다. 자리 순서로 적으면 매니페스트에 봉인 필드가
	// 하나 늘 때마다 조용히 엉뚱한 자리로 밀린다.
	prefix := map[Market]string{MarketKR: "kr", MarketUS: "us"}[market]
	lane := func(family Family, horizon Horizon, laneID, laneVersion, name string, scorePPM uint32) productionRouteCandidate {
		return productionRouteCandidate{Family: family, Horizon: horizon, LaneID: laneID, LaneVersion: laneVersion,
			ScorePPM: scorePPM, Eligible: true, Desired: StateOn, Effective: StateOn,
			EvidenceDigest: prefix + "-" + name + "-evidence", ConfigDigest: prefix + "-" + name + "-config"}
	}
	lanes := []productionRouteCandidate{}
	if market == MarketKR {
		lanes = []productionRouteCandidate{
			lane(FamilyContinuation, HorizonShort, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, "continuation", 300_000),
			lane(FamilyReversal, HorizonShort, reversallane.KRReversalLaneID, reversallane.LaneVersionV1, "reversal", 200_000),
			lane(FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, "weekly", 100_000),
			lane(FamilyBreakoutRetest, HorizonShort, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, "breakout", 50_000),
		}
	} else {
		lanes = []productionRouteCandidate{
			lane(FamilyContinuation, HorizonShort, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, "continuation", 300_000),
			lane(FamilyReversal, HorizonShort, reversallane.USReversalLaneID, reversallane.LaneVersionV1, "reversal", 200_000),
			lane(FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, "weekly", 100_000),
			lane(FamilyBreakoutRetest, HorizonShort, breakoutlane.USLaneID, breakoutlane.LaneVersionV1, "breakout", 50_000),
		}
	}
	return productionRouteBody{SchemaVersion: productionRouteSchema, Domain: productionRouteDomain, SignatureAlgorithm: productionRouteAlgorithm,
		KeyID: "route-key-v1", Generation: 1, AccountRef: "acct", Market: market,
		MarketRevision: 1, ActivationDigest: "activation-" + string(market), ActivationExpiresAt: fixture.now.Add(time.Hour).Format(time.RFC3339Nano),
		CalendarGeneration: "calendar-generation-" + string(market), CalendarDigest: "calendar-digest-" + string(market), Timezone: timezone,
		SessionScope: "REGULAR", ConfigVersion: "scheduler-config-" + string(market),
		ArbitrationScoreVersion: "arbitration-score-v1", CalibrationDigest: "sha256:calibration-" + string(market),
		Actor: "human-approver", ObservedAt: fixture.now.Add(-time.Minute).Format(time.RFC3339Nano),
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

// 태스크 4.3: 시장마다 정확히 네 가족이다. 세 가족짜리 레거시 매니페스트는
// 더 이상 활성화 권한이 아니다.
func TestProductionRouteDescriptorsCoverFourFamiliesPerMarket(t *testing.T) {
	want := map[Market]map[string]productionLaneDescriptor{
		MarketKR: {
			continuationlane.KRContinuationLaneID: {FamilyContinuation, HorizonShort, "v1"},
			reversallane.KRReversalLaneID:         {FamilyReversal, HorizonShort, "v1"},
			weeklyvaluelane.KRWeeklyLaneID:        {FamilyWeeklyValue, HorizonWeekly, "v1"},
			breakoutlane.KRLaneID:                 {FamilyBreakoutRetest, HorizonShort, "v1"},
		},
		MarketUS: {
			continuationlane.USContinuationLaneID: {FamilyContinuation, HorizonShort, "v1"},
			reversallane.USReversalLaneID:         {FamilyReversal, HorizonShort, "v1"},
			weeklyvaluelane.USWeeklyLaneID:        {FamilyWeeklyValue, HorizonWeekly, "v1"},
			breakoutlane.USLaneID:                 {FamilyBreakoutRetest, HorizonShort, "v1"},
		},
	}
	// 표가 네 가족을 정확히 한 번씩 담는다는 사실을 여기서 지킨다.
	// validProductionRouteCandidates 는 매니페스트를 이 표에 맞추기만 하므로,
	// 표가 틀리면 매니페스트 검증도 같이 틀린다.
	enum := []Family{FamilyContinuation, FamilyReversal, FamilyWeeklyValue, FamilyBreakoutRetest}
	for market, lanes := range want {
		got := productionRouteDescriptors(market)
		if len(got) != 4 {
			t.Fatalf("%s has %d production descriptors, want 4", market, len(got))
		}
		for laneID, expected := range lanes {
			descriptor, ok := got[laneID]
			if !ok {
				t.Fatalf("%s is missing production descriptor %s", market, laneID)
			}
			if descriptor != expected {
				t.Fatalf("%s/%s descriptor drifted: got %+v want %+v", market, laneID, descriptor, expected)
			}
		}
		families := make(map[Family]string, 4)
		for laneID, descriptor := range got {
			if previous, duplicated := families[descriptor.Family]; duplicated {
				t.Fatalf("%s binds family %s to both %s and %s", market, descriptor.Family, previous, laneID)
			}
			families[descriptor.Family] = laneID
		}
		for _, family := range enum {
			if families[family] == "" {
				t.Fatalf("%s has no lane for family %s", market, family)
			}
		}
	}
	if got := productionRouteDescriptors(Market("XX")); got != nil {
		t.Fatalf("an unknown market produced descriptors: %+v", got)
	}
}

func TestProductionRouteCandidatesRejectLegacyThreeFamilyAndPartialSets(t *testing.T) {
	full := func(market Market) []productionRouteCandidate {
		values := make([]productionRouteCandidate, 0, 4)
		for laneID, descriptor := range productionRouteDescriptors(market) {
			values = append(values, productionRouteCandidate{Family: descriptor.Family, Horizon: descriptor.Horizon,
				LaneID: laneID, LaneVersion: descriptor.LaneVersion, ScorePPM: 10, Eligible: true,
				Desired: StateOn, Effective: StateOn, EvidenceDigest: "evidence-" + laneID, ConfigDigest: "config-" + laneID})
		}
		return values
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		values := full(market)
		if !validProductionRouteCandidates(market, values) {
			t.Fatalf("%s four-family candidate set was rejected", market)
		}
		if validProductionRouteCandidates(market, values[:3]) {
			t.Fatalf("%s legacy three-family manifest was accepted as authority", market)
		}
		duplicated := append(append([]productionRouteCandidate(nil), values[:3]...), values[0])
		if validProductionRouteCandidates(market, duplicated) {
			t.Fatalf("%s duplicated candidate was accepted", market)
		}
		unknown := append([]productionRouteCandidate(nil), values...)
		unknown[0].LaneID = "xx_unknown_lane_v1"
		if validProductionRouteCandidates(market, unknown) {
			t.Fatalf("%s unknown lane id was accepted", market)
		}
		mismatched := append([]productionRouteCandidate(nil), values...)
		mismatched[0].LaneVersion = "v2"
		if validProductionRouteCandidates(market, mismatched) {
			t.Fatalf("%s mismatched lane version was accepted", market)
		}
	}
}

// 태스크 4.3(채점 봉인): 생산 경로가 후보에 원시 점수를 실어 주지 않는다.
// 원시 int64 점수는 단위가 다른 신호를 견주므로 가족 간 선택 권한이 될 수 없다.
// 점수가 사라지면 레거시 Route 는 자격 있는 네 후보를 구분할 수 없어 닫아서 거절한다 —
// 즉 "평가 전 사전 선택"이 약속이 아니라 컴파일된 사실이 된다.
func TestProductionRouteCandidatesCarryNoRawArbitrationScore(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		authority, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market])
		if err != nil {
			t.Fatalf("%s load: %v", market, err)
		}
		for _, candidate := range authority.Request().Candidates {
			if candidate.Score != 0 {
				t.Fatalf("%s/%s still carries a raw arbitration score %d", market, candidate.LaneID, candidate.Score)
			}
		}
		if routed := Route(authority.Request()); routed.Code != RefusalAmbiguous {
			t.Fatalf("%s legacy Route still preselected a family: %+v", market, routed)
		}
	}
}

// 태스크 4.3(레거시 거절): 이전 스키마 이름을 단 매니페스트는 활성화 권한이 아니다.
func TestProductionRouteManifestRefusesTheLegacySchemaVersion(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	for _, market := range []Market{MarketKR, MarketUS} {
		body := fixture.body(market)
		body.SchemaVersion = "strategy-lane-authority:v1"
		fixture.write(t, market, body)
		if _, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market]); err == nil {
			t.Fatalf("%s accepted a legacy v1 manifest as activation authority", market)
		}
	}
}

// 태스크 4.3(가족 봉인): 레인 이름이 네 개여도 가족이 네 개가 아니면 권한이 아니다.
func TestProductionRouteCandidatesRejectFamilyDriftAndPartialFamilyCoverage(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		values := make([]productionRouteCandidate, 0, 4)
		for laneID, descriptor := range productionRouteDescriptors(market) {
			values = append(values, productionRouteCandidate{Family: descriptor.Family, Horizon: descriptor.Horizon,
				LaneID: laneID, LaneVersion: descriptor.LaneVersion, ScorePPM: 10, Eligible: true,
				Desired: StateOn, Effective: StateOn, EvidenceDigest: "evidence-" + laneID, ConfigDigest: "config-" + laneID})
		}
		if !validProductionRouteCandidates(market, values) {
			t.Fatalf("%s four-family set was rejected", market)
		}
		missing := append([]productionRouteCandidate(nil), values...)
		missing[0].Family = ""
		if validProductionRouteCandidates(market, missing) {
			t.Fatalf("%s accepted a candidate with no family", market)
		}
		unknown := append([]productionRouteCandidate(nil), values...)
		unknown[0].Family = Family("SCALP")
		if validProductionRouteCandidates(market, unknown) {
			t.Fatalf("%s accepted a family outside the exact enum", market)
		}
		// 네 레인은 그대로 두고 가족만 하나로 몰아 준다. 레인 수만 세는 검증은
		// 이것을 통과시키고, 가족 봉인만이 잡는다.
		collapsed := append([]productionRouteCandidate(nil), values...)
		for index := range collapsed {
			collapsed[index].Family = values[0].Family
		}
		if validProductionRouteCandidates(market, collapsed) {
			t.Fatalf("%s accepted four lanes that all claim one family", market)
		}
		swapped := append([]productionRouteCandidate(nil), values...)
		swapped[0].Family, swapped[1].Family = swapped[1].Family, swapped[0].Family
		if validProductionRouteCandidates(market, swapped) {
			t.Fatalf("%s accepted a family bound to the wrong lane", market)
		}
	}
}

// 태스크 4.3(채점 봉인): ppm 은 0..1,000,000 이다. 그 위는 권한이 아니다.
func TestProductionRouteCandidatesRejectAScorePPMAboveTheApprovedRange(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		values := make([]productionRouteCandidate, 0, 4)
		for laneID, descriptor := range productionRouteDescriptors(market) {
			values = append(values, productionRouteCandidate{Family: descriptor.Family, Horizon: descriptor.Horizon,
				LaneID: laneID, LaneVersion: descriptor.LaneVersion, ScorePPM: productionRouteScorePPMMax, Eligible: true,
				Desired: StateOn, Effective: StateOn, EvidenceDigest: "evidence-" + laneID, ConfigDigest: "config-" + laneID})
		}
		if !validProductionRouteCandidates(market, values) {
			t.Fatalf("%s rejected the maximum approved score ppm", market)
		}
		over := append([]productionRouteCandidate(nil), values...)
		over[0].ScorePPM = productionRouteScorePPMMax + 1
		if validProductionRouteCandidates(market, over) {
			t.Fatalf("%s accepted a score ppm above %d", market, productionRouteScorePPMMax)
		}
	}
}

// 태스크 4.3(보정 봉인): 승인된 점수 버전이나 보정 다이제스트가 없는 매니페스트는
// 활성화 권한이 아니다. 하나만 빠져도 시장 전체가 닫힌다.
func TestProductionRouteManifestRefusesAMissingCalibrationSeal(t *testing.T) {
	for _, test := range []struct {
		name   string
		break_ func(*productionRouteBody)
	}{
		{"score version", func(body *productionRouteBody) { body.ArbitrationScoreVersion = "" }},
		{"calibration digest", func(body *productionRouteBody) { body.CalibrationDigest = "" }},
		{"both", func(body *productionRouteBody) {
			body.ArbitrationScoreVersion, body.CalibrationDigest = "", ""
		}},
	} {
		for _, market := range []Market{MarketKR, MarketUS} {
			fixture := newProductionRouteFixture(t)
			body := fixture.body(market)
			test.break_(&body)
			fixture.write(t, market, body)
			if _, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market]); err == nil {
				t.Fatalf("%s accepted a manifest with no %s", market, test.name)
			}
		}
	}
}

// 태스크 4.3: 세 봉인은 서로 다른 것을 증명하고, 재료가 바뀌면 각자 따로 달라진다.
func TestProductionRouteAuthorityCarriesThreeIndependentSeals(t *testing.T) {
	fixture := newProductionRouteFixture(t)
	seals := make(map[Market]ProductionRouteSeals, 2)
	for _, market := range []Market{MarketKR, MarketUS} {
		authority, err := LoadProductionRouteAuthority(context.Background(), fixture.config[market])
		if err != nil {
			t.Fatalf("%s load: %v", market, err)
		}
		if !authority.SealsValid() {
			t.Fatalf("%s seals did not verify: %+v", market, authority.Seals())
		}
		value := authority.Seals()
		if value.Family == "" || value.Scoring == "" || value.Calibration == "" {
			t.Fatalf("%s left a seal empty: %+v", market, value)
		}
		if value.Family == value.Scoring || value.Family == value.Calibration || value.Scoring == value.Calibration {
			t.Fatalf("%s produced two identical seals: %+v", market, value)
		}
		calibration := authority.Calibration()
		if calibration.ScoreVersion != "arbitration-score-v1" || calibration.CalibrationDigest != "sha256:calibration-"+string(market) {
			t.Fatalf("%s calibration=%+v", market, calibration)
		}
		scores := authority.FamilyScores()
		if len(scores) != 4 {
			t.Fatalf("%s carried %d family scores, want 4", market, len(scores))
		}
		families := make(map[Family]bool, 4)
		for index, score := range scores {
			if index > 0 && scores[index-1].Family >= score.Family {
				t.Fatalf("%s family scores are not in family order: %+v", market, scores)
			}
			families[score.Family] = true
		}
		for _, family := range []Family{FamilyContinuation, FamilyReversal, FamilyWeeklyValue, FamilyBreakoutRetest} {
			if !families[family] {
				t.Fatalf("%s is missing family %s in the sealed scores", market, family)
			}
		}
		// 복사본을 고쳐도 봉인된 원본은 그대로다.
		scores[0].ScorePPM = 999
		if again := authority.FamilyScores(); again[0].ScorePPM == 999 {
			t.Fatalf("%s FamilyScores handed out the sealed slice itself", market)
		}
		if !authority.SealsValid() {
			t.Fatalf("%s seals broke after a caller edited its own copy", market)
		}
		seals[market] = value
	}
	if seals[MarketKR].Family == seals[MarketUS].Family {
		t.Fatal("KR and US share a family seal despite different lane ids")
	}
	if seals[MarketKR].Calibration == seals[MarketUS].Calibration {
		t.Fatal("KR and US share a calibration seal despite different calibration digests")
	}

	// 점수만 하나 바꾸면 채점 봉인만 달라져야 한다. 가족 구성은 그대로이므로
	// 가족 봉인은 같은 값이어야 한다.
	body := fixture.body(MarketKR)
	body.Scopes[0].Candidates[0].ScorePPM = 123_456
	fixture.write(t, MarketKR, body)
	changed, err := LoadProductionRouteAuthority(context.Background(), fixture.config[MarketKR])
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if changed.Seals().Family != seals[MarketKR].Family {
		t.Fatal("a score change moved the family seal")
	}
	if changed.Seals().Calibration != seals[MarketKR].Calibration {
		t.Fatal("a score change moved the calibration seal")
	}
	if changed.Seals().Scoring == seals[MarketKR].Scoring {
		t.Fatal("a score change left the scoring seal unchanged")
	}
	if !changed.SealsValid() {
		t.Fatal("reloaded seals did not verify")
	}
	if (ProductionRouteAuthority{}).SealsValid() {
		t.Fatal("an empty authority reported valid seals")
	}
}
