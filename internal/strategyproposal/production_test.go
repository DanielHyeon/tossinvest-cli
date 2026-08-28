//go:build tossos_testseams

package strategyproposal

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

func TestProductionProposalAuthorityRecognizesExactPairedSixLaneMatrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		market  strategyrouter.Market
		lane    string
		version string
		horizon strategyrouter.Horizon
	}{
		{strategyrouter.MarketKR, continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort},
		{strategyrouter.MarketUS, continuationlane.USContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort},
		{strategyrouter.MarketKR, reversallane.KRReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort},
		{strategyrouter.MarketUS, reversallane.USReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort},
		{strategyrouter.MarketKR, weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly},
		{strategyrouter.MarketUS, weeklyvaluelane.USWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly},
	}
	for _, test := range tests {
		if !laneMatchesMarket(test.market, test.lane, test.version, test.horizon) {
			t.Fatalf("missing production lane market=%s lane=%s horizon=%s", test.market, test.lane, test.horizon)
		}
		peer := strategyrouter.MarketKR
		if test.market == strategyrouter.MarketKR {
			peer = strategyrouter.MarketUS
		}
		if laneMatchesMarket(peer, test.lane, test.version, test.horizon) {
			t.Fatalf("cross-market production lane accepted market=%s lane=%s", peer, test.lane)
		}
	}
}

func TestProductionProposalAuthorityLoadsPairedSignedKRUSSnapshots(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	type result struct {
		market    strategyrouter.Market
		authority ProductionAuthority
		err       error
	}
	results := make(chan result, 2)
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		market := market
		config, target, fx := productionFixture(t, market, now)
		go func() {
			batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
			authority, _ := batch.For(target.Approved.Symbol())
			results <- result{market: market, authority: authority, err: err}
		}()
	}
	seen := map[strategyrouter.Market]bool{}
	for range 2 {
		result := <-results
		if result.err != nil || !result.authority.Proposal().ValidProposal() || result.authority.Proposal().Quantity != 8 {
			t.Fatalf("%s authority=%+v err=%v", result.market, result.authority, result.err)
		}
		seen[result.market] = true
	}
	if !seen[strategyrouter.MarketKR] || !seen[strategyrouter.MarketUS] {
		t.Fatalf("paired markets=%v", seen)
	}
}

func TestProductionProposalAuthorityFailureIsMarketLocal(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	krConfig, krTarget, krFX := productionFixture(t, strategyrouter.MarketKR, now)
	usConfig, usTarget, usFX := productionFixture(t, strategyrouter.MarketUS, now)
	krConfig.ManifestDigest = "sha256:" + string(make([]byte, 64))
	if batch, err := LoadProductionAuthorityBatch(context.Background(), krConfig, []ProductionTarget{krTarget}, krFX); err == nil || batch.Len() != 0 {
		t.Fatalf("invalid KR authority batch=%+v err=%v", batch, err)
	}
	batch, err := LoadProductionAuthorityBatch(context.Background(), usConfig, []ProductionTarget{usTarget}, usFX)
	authority, ok := batch.For(usTarget.Approved.Symbol())
	if err != nil || !ok || !authority.Proposal().ValidProposal() {
		t.Fatalf("US gated by KR failure: batch=%+v err=%v", batch, err)
	}
}

func productionFixture(t *testing.T, market strategyrouter.Market, now time.Time) (ProductionConfig, ProductionTarget, officialfx.Evidence) {
	t.Helper()
	dir := t.TempDir()
	clockMarket := marketclock.MarketKR
	symbol, currency, lane := "005930", "KRW", continuationlane.KRContinuationLaneID
	payload := `{"net_flow_notional_minor":"1","turnover_notional_minor":"10","flow_pressure_ppm":"100000"}`
	kind, authority, schema, effective := strategyevidence.KindKRNetFlow, strategyevidence.AuthorityKRX, continuationlane.KRFlowSchemaV1, "2026-08-04"
	if market == strategyrouter.MarketUS {
		clockMarket, symbol, currency, lane = marketclock.MarketUS, "AAPL", "USD", continuationlane.USContinuationLaneID
		payload = `{"participating_volume_shares":"2","baseline_volume_shares":"10","reference_price_minor":"100","last_price_minor":"101","participation_ppm":"200000","price_change_ppm":"10000"}`
		kind, authority, schema, effective = strategyevidence.KindUSParticipation, strategyevidence.AuthorityTossOpenAPI, continuationlane.USParticipationSchemaV1, "2026-08-03"
	}
	header := strategyevidence.Header{EvidenceID: "proposal-" + string(market), Market: clockMarket, Symbol: symbol, IssuerIdentity: "issuer-" + symbol,
		IssuerMappingVersion: "mapping-v1", Kind: kind, SchemaVersion: schema, Authority: authority, SourceRecordID: "record-" + string(market),
		RevisionIdentity: "revision-1", MarketEffectiveDate: effective, SourceEventAt: now.Add(-3 * time.Second), SourceAvailableAt: now.Add(-2 * time.Second),
		ObservedAt: now.Add(-time.Second), IngestedAt: now, Currency: currency, Unit: "minor-v1", Availability: strategyevidence.AvailabilityAvailable,
		Confidence: strategyevidence.ConfidenceVerified}
	envelope, err := strategyevidence.NewEnvelope(header, []byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(dir, "evidence.db")
	store, err := strategyevidence.Open(context.Background(), strategyevidence.Options{Path: evidencePath, Clock: marketclock.NewFake(now)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Append(context.Background(), envelope); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.SealSnapshot(context.Background(), strategyevidence.SnapshotQuery{Market: clockMarket, Symbol: symbol,
		IssuerIdentity: header.IssuerIdentity, IssuerMappingVersion: header.IssuerMappingVersion, EvaluationAt: now, IngestionCutoff: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(evidencePath, 0o600); err != nil {
		t.Fatal(err)
	}

	approved := strategy.ApprovedSnapshotForTest(string(market), symbol, now)
	key, err := strategyrouter.NewOwnerKey("acct", market, symbol, 1)
	if err != nil {
		t.Fatal(err)
	}
	route, err := strategyrouter.ProductionRouteAuthorityForTest(key, strategyrouter.HorizonShort, lane, continuationlane.LaneVersionV1, snapshot.Digest, "config-"+string(market), now)
	if err != nil {
		t.Fatal(err)
	}
	target := ProductionTarget{Approved: approved, Router: route.Request()}

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	body := productionBody{SchemaVersion: productionSchema, Domain: productionDomain, SignatureAlgorithm: productionAlgorithm, KeyID: "proposal-key",
		Generation: 1, AccountRef: "acct", Market: market, RouteManifestDigest: "route-manifest", ActivationDigest: "activation",
		CalendarGeneration: "calendar-generation", CalendarDigest: "calendar-digest", SchedulerConfigVersion: "scheduler-v1",
		EvidenceDBIdentity: "evidence-db-" + string(market), Actor: "risk-committee", ObservedAt: now.Add(-time.Second).Format(time.RFC3339Nano),
		FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano), Scopes: []productionScope{{Symbol: symbol, PositionGeneration: 1,
			CandidateID: approved.CandidateLifeID(), CampaignID: "campaign-" + string(market), Horizon: strategyrouter.HorizonShort, LaneID: lane,
			LaneVersion: continuationlane.LaneVersionV1, SnapshotID: snapshot.ID, SnapshotDigest: snapshot.Digest, RiskBudgetMinor: "1000",
			PerShareRiskMinor: "10", PlannedQuantity: 14, PolicyDigest: "risk-policy", AccountCurrency: "KRW", QuoteCurrency: currency,
			LegOrdinal: 1, SavedEffectiveStopMinor: "90", Stop: productionStop{PriceMinor: "95", Source: "structure", Policy: "stop-v1",
				Version: "v1", Digest: "stop-digest", ObservedAt: now.Add(-time.Second).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano)},
			EntryPriceMinor: "110", TargetPriceMinor: "130", FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano), ConfigDigest: "config-" + string(market),
			ThresholdSet: "threshold-" + string(market), MinimumFlowPressurePPM: 100000, MinimumParticipationPPM: 200000, MinimumPriceChangePPM: 10000}}}
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	manifest := productionManifest{productionBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, ProductionFileName(market))
	if err := os.WriteFile(manifestPath, data, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(manifestPath, 0o400); err != nil {
		t.Fatal(err)
	}
	fx, err := officialfx.EvidenceForAuthorityTest(currency, "KRW", map[strategyrouter.Market]string{strategyrouter.MarketKR: "1", strategyrouter.MarketUS: "1300"}[market], "1.1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	config := ProductionConfig{ConfigDir: dir, EvidencePath: evidencePath, JournalPath: filepath.Join(dir, "journal.db"), AccountRef: "acct", Market: market,
		ManifestDigest: digest(data), TrustedKeyID: "proposal-key", TrustedKey: public, ObservedAt: now, RouteManifestDigest: "route-manifest",
		ActivationDigest: "activation", CalendarGeneration: "calendar-generation", CalendarDigest: "calendar-digest", SchedulerConfigVersion: "scheduler-v1",
		EvidenceDBIdentity: "evidence-db-" + string(market)}
	return config, target, fx
}

// 태스크 4.3.1 / 2.6.1: 한 종목에 여러 가족 스코프가 오면 전부 살아남아야 한다.
// 예전처럼 종목 하나에 제안 하나로 접으면 뒤 스코프가 조용히 앞을 덮어쓴다.
func TestValidScopesAcceptsSeveralFamiliesForOneSymbolAndStillRejectsDuplicateLanes(t *testing.T) {
	base := productionScope{Symbol: "005930", PositionGeneration: 1, CandidateID: "c", CampaignID: "camp",
		SnapshotID: "s", SnapshotDigest: "d", PlannedQuantity: 1, LegOrdinal: 1}
	continuation := base
	continuation.LaneID, continuation.LaneVersion, continuation.Horizon = continuationlane.KRContinuationLaneID, continuationlane.LaneVersionV1, strategyrouter.HorizonShort
	reversal := base
	reversal.LaneID, reversal.LaneVersion, reversal.Horizon = reversallane.KRReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort
	breakout := base
	breakout.LaneID, breakout.LaneVersion, breakout.Horizon = breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort

	// 레인 이름 사전순: absorption_reversal < breakout_retest < flow_continuation.
	if !validScopes(strategyrouter.MarketKR, []productionScope{reversal, breakout, continuation}) {
		t.Fatal("several families for one symbol were rejected, so only one family per symbol can ever propose")
	}
	if validScopes(strategyrouter.MarketKR, []productionScope{continuation, reversal}) {
		t.Fatal("unordered scopes were accepted, so the manifest order is not deterministic")
	}
	if validScopes(strategyrouter.MarketKR, []productionScope{continuation, continuation}) {
		t.Fatal("the same lane twice was accepted for one symbol")
	}
}

// 태스크 4.3: breakout 레인은 매니페스트에서 시장별로 인식돼야 한다.
func TestLaneMatchesMarketCoversTheBreakoutFamily(t *testing.T) {
	if !laneMatchesMarket(strategyrouter.MarketKR, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort) {
		t.Fatal("KR breakout lane is not recognised by the production manifest")
	}
	if !laneMatchesMarket(strategyrouter.MarketUS, breakoutlane.USLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort) {
		t.Fatal("US breakout lane is not recognised by the production manifest")
	}
	if laneMatchesMarket(strategyrouter.MarketKR, breakoutlane.USLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonShort) {
		t.Fatal("the US breakout lane was accepted in KR")
	}
	if laneMatchesMarket(strategyrouter.MarketKR, breakoutlane.KRLaneID, breakoutlane.LaneVersionV1, strategyrouter.HorizonWeekly) {
		t.Fatal("the breakout lane was accepted on the weekly horizon")
	}
}

// 결정 49: 파생 지표 증거가 없으므로 breakout 레인 입력은 오늘 닫힌 채로 거절돼야 한다.
// 조용히 빈 입력을 만들어 통과시키면 근거 없는 제안이 생긴다.
func TestBreakoutLaneInputFailsClosedWhileTheDerivedMetricEvidenceIsMissing(t *testing.T) {
	scope := productionScope{Symbol: "005930", PositionGeneration: 1, CandidateID: "c", CampaignID: "camp",
		SnapshotID: "s", SnapshotDigest: "d", PlannedQuantity: 1, LegOrdinal: 1,
		LaneID: breakoutlane.KRLaneID, LaneVersion: breakoutlane.LaneVersionV1, Horizon: strategyrouter.HorizonShort,
		Stop:       productionStop{ObservedAt: "2026-08-28T00:00:00Z", FreshUntil: "2026-08-28T01:00:00Z"},
		FreshUntil: "2026-08-28T01:00:00Z"}
	snapshot := strategyevidence.Snapshot{ID: "s", Digest: "d"}
	lane, weekly, err := buildLaneInput(context.Background(), ProductionConfig{Market: strategyrouter.MarketKR, AccountRef: "acct"},
		scope, snapshot, officialfx.Evidence{}, nil)
	if err == nil {
		t.Fatal("a breakout scope produced a lane input even though no ATR/RVOL evidence exists")
	}
	if !errors.Is(err, ErrBreakoutEvidenceUnavailable) {
		t.Fatalf("breakout refusal must be typed, got %v", err)
	}
	if weekly != nil {
		t.Fatal("a breakout scope minted a weekly reservation binding")
	}
	// LaneInput 은 비교할 수 없는 구조라 제로값 자체를 비교하지 못한다.
	// 대신 어떤 descriptor 에도 붙지 않는다는 사실로 "빈 입력"임을 증명한다.
	for _, descriptor := range strategyflow.Descriptors() {
		if strategyflow.LaneInputMatchesForTest(lane, descriptor) {
			t.Fatalf("a refused breakout scope still returned a lane input bound to %s", descriptor.LaneID)
		}
	}
}
