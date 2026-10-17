//go:build tossos_testseams

package strategyproposal

// a112 4.5(-a · -b) · BTM L 행(Manager 판정 2026-10-04 (A)(i)): reversal · weekly 의 생산 lane-input 구성(buildLaneInput 의 reversal 갈래
// B8~B15 · weekly 갈래 B16~B21, LoadProductionAuthorityBatch 의 weekly 원장 열기 B7)이 **실 적재기** 를 한 번도 지나지 않았다(BTM 재측정
// `measurements/gate-8.1-8.3-2026-10-04/btm-remeasure-prod.tsv` class L). 여기서 KR · US 짝으로 실제 LoadProductionAuthorityBatch 를 돈다:
// 서명 매니페스트 파일 · 실 증거 저장소(봉인 스냅숏) · 실 경로 권한(RouteSet) · weekly 는 실 원장(journal.Open → ReserveWeeklyMarket, 적재기는
// OpenReadOnly 로 다시 읽음). 기존 continuation 시험(production_test.go)과 합쳐 제안 측 6/6(벽 아래). breakout 은 벽 앞에서 정상 부재 —
// 두 시장 모두(기존 시험은 KR 만).
//
// 매니페스트 서명 · 파일 쓰기는 production_test.go 의 productionFixtureOn 과 같은 모양이다(그 함수는 continuation 증거로 고정 — 손대지 않고
// 여기 새로 둔다).

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

type a112LaneFamily string

const (
	a112Reversal a112LaneFamily = "reversal"
	a112Weekly   a112LaneFamily = "weekly"
)

type a112LaneFixture struct {
	config        ProductionConfig
	target        ProductionTarget
	fx            officialfx.Evidence
	laneID        string
	laneVersion   string
	horizon       strategyrouter.Horizon
	campaignID    string
	reservationID string
	snapshot      strategyevidence.Snapshot
}

func a112PairedLaneFixture(t *testing.T, market strategyrouter.Market, family a112LaneFamily, now time.Time, mutate ...func(*productionScope)) a112LaneFixture {
	t.Helper()
	dir := t.TempDir()
	kr := market == strategyrouter.MarketKR
	clockMarket, symbol, currency := marketclock.MarketKR, "005930", "KRW"
	if !kr {
		clockMarket, symbol, currency = marketclock.MarketUS, "AAPL", "USD"
	}
	effectiveDate := map[bool]string{true: "2026-08-04", false: "2026-08-03"}[kr]

	// 증거: 가족별 실 봉투 하나(레인 생산 시험 reversallane · weeklyvaluelane production_proposal_test.go 와 같은 페이로드).
	var header strategyevidence.Header
	var payload string
	switch family {
	case a112Reversal:
		kind, authority, schema := strategyevidence.KindKRNetFlow, strategyevidence.AuthorityKRX, "kr-absorption-v1"
		payload = `{"absorbed_notional_minor":"1","aggressive_sell_notional_minor":"4","absorption_ppm":"250000"}`
		if !kr {
			kind, authority, schema = strategyevidence.KindUSParticipation, strategyevidence.AuthorityTossOpenAPI, "us-dislocation-v1"
			payload = `{"reference_price_minor":"100","dislocation_low_price_minor":"90","dislocation_volume_shares":"150","baseline_volume_shares":"100","drawdown_ppm":"100000","relative_volume_ppm":"1500000"}`
		}
		header = strategyevidence.Header{EvidenceID: "reversal-" + string(market), Market: clockMarket, Symbol: symbol, IssuerIdentity: "issuer-" + symbol,
			IssuerMappingVersion: "mapping-v1", Kind: kind, SchemaVersion: schema, Authority: authority, SourceRecordID: "record-" + string(market),
			RevisionIdentity: "revision-1", MarketEffectiveDate: effectiveDate, SourceEventAt: now.Add(-3 * time.Second),
			SourceAvailableAt: now.Add(-2 * time.Second), ObservedAt: now.Add(-time.Second), IngestedAt: now, Currency: currency, Unit: "minor-v1",
			Availability: strategyevidence.AvailabilityAvailable, Confidence: strategyevidence.ConfidenceVerified}
	case a112Weekly:
		source, authority, schema := weeklyvaluelane.SourceOpenDART, strategyevidence.AuthorityOpenDART, weeklyvaluelane.KRDisclosureSchemaV1
		filing, scale, dilution := "202608040001", "0", "OBSERVED"
		if !kr {
			source, authority, schema = weeklyvaluelane.SourceEDGAR, strategyevidence.AuthoritySEC, weeklyvaluelane.USDisclosureSchemaV1
			filing, scale, dilution = "0000320193-26-000077", "2", "NONE"
		}
		asOf := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
		observed, ingested, cutoff := now.Add(-10*time.Minute), now.Add(-9*time.Minute), now.Add(-8*time.Minute)
		payload = fmt.Sprintf(`{"schema_version":%q,"market":%q,"source":%q,"symbol":%q,"issuer_id":%q,"filing_id":%q,"report_id":"quarterly-2026-q2","revision_id":"rev-1","superseded_revision_id":"NONE","revision_sequence":"1","as_of":%q,"observed_at":%q,"ingested_at":%q,"cutoff_at":%q,"evaluated_at":%q,"fresh_until":%q,"currency":%q,"monetary_unit":"MINOR","monetary_scale":%q,"diluted_shares":"100","shares_unit":"SHARES","dilution_status":%q,"dilution_facts_digest":"dilution-digest","dilution_as_of":%q,"financial_inputs":[{"name":"equity_value","value_minor":"110000","unit":%q}],"model_id":"weekly-value","model_version":"weekly-model-v1","model_config_digest":%q,"threshold_digest":%q,"equity_value_minor":"110000","fair_value_minor":"1100"}`,
			schema, market, source, symbol, "issuer-"+symbol, filing, asOf.Format(time.RFC3339Nano), observed.Format(time.RFC3339Nano),
			ingested.Format(time.RFC3339Nano), cutoff.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), now.Add(7*24*time.Hour).Format(time.RFC3339Nano),
			currency, scale, dilution, asOf.Format(time.RFC3339Nano), currency+"_MINOR", a112WeeklyModelConfig(market), a112WeeklyThreshold(market))
		header = strategyevidence.Header{EvidenceID: "weekly-" + string(market), Market: clockMarket, Symbol: symbol, IssuerIdentity: "issuer-" + symbol,
			IssuerMappingVersion: "mapping-v1", Kind: strategyevidence.KindDisclosure, SchemaVersion: schema, Authority: authority,
			SourceRecordID: filing, RevisionIdentity: "rev-1", MarketEffectiveDate: effectiveDate, SourceEventAt: asOf,
			SourceAvailableAt: observed.Add(-time.Minute), ObservedAt: observed, IngestedAt: ingested, Currency: currency, Unit: "minor-v1",
			Availability: strategyevidence.AvailabilityAvailable, Confidence: strategyevidence.ConfidenceVerified}
	default:
		t.Fatalf("unknown family %q", family)
	}
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
	campaignID := "campaign-" + string(market)
	journalPath := filepath.Join(dir, "journal.db")
	scope := productionScope{Symbol: symbol, PositionGeneration: 1, CandidateID: approved.CandidateLifeID(), CampaignID: campaignID,
		SnapshotID: snapshot.ID, SnapshotDigest: snapshot.Digest, RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14,
		PolicyDigest: "risk-policy", AccountCurrency: "KRW", QuoteCurrency: currency, LegOrdinal: 1, SavedEffectiveStopMinor: "90",
		FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano)}
	fixture := a112LaneFixture{snapshot: snapshot, campaignID: campaignID}
	switch family {
	case a112Reversal:
		fixture.laneID, fixture.laneVersion, fixture.horizon = reversallane.KRReversalLaneID, reversallane.LaneVersionV1, strategyrouter.HorizonShort
		if !kr {
			fixture.laneID = reversallane.USReversalLaneID
		}
		scope.ConfigDigest, scope.ConfigVersion, scope.ThresholdSet = "config-"+string(market), "config-v1", "threshold-"+string(market)
		scope.MinimumAbsorptionPPM, scope.MinimumDrawdownPPM, scope.MinimumRelativeVolumePPM = 250000, 100000, 1500000
		scope.StructuralWindowNS = int64(time.Minute)
		scope.EntryPriceMinor, scope.TargetPriceMinor = "100", "120"
		scope.Stop = productionStop{PriceMinor: "95", Source: "structure", Policy: "stop-v1", Version: "v1", Digest: "stop-digest",
			ObservedAt: now.Add(-time.Second).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano)}
	case a112Weekly:
		fixture.laneID, fixture.laneVersion, fixture.horizon = weeklyvaluelane.KRWeeklyLaneID, weeklyvaluelane.LaneVersionV1, strategyrouter.HorizonWeekly
		provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W32"
		if !kr {
			fixture.laneID = weeklyvaluelane.USWeeklyLaneID
			provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W32"
		}
		// 실 원장에 주간 예약 하나 — 적재기는 이것을 OpenReadOnly 로 다시 읽어 결속한다.
		writable, err := journal.Open(context.Background(), journal.Options{Path: journalPath, Clock: marketclock.NewFake(now),
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
		if err != nil {
			t.Fatalf("journal.Open: %v", err)
		}
		reservation, err := writable.ReserveWeeklyMarket(context.Background(), journal.WeeklyMarketReservationRequest{
			ReservationID: "reservation-" + string(market), CampaignID: campaignID, Market: string(market), StableWeek: stable,
			Provider: provider, TimeZone: zone, SessionDate: "2026-08-03", CalendarGeneration: "generation-A", CalendarDigest: "calendar-A",
			IdempotencyKey: "a112-" + string(market), PlannedOrdinal: 1, ExpectedVersion: 0,
			ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Hour), EvaluatedAt: now})
		if err != nil {
			t.Fatalf("ReserveWeeklyMarket: %v", err)
		}
		if err := writable.Close(); err != nil {
			t.Fatal(err)
		}
		fixture.reservationID = reservation.ReservationID
		scope.ConfigDigest, scope.ModelVersion, scope.ThresholdDigest = a112WeeklyModelConfig(market), "weekly-model-v1", a112WeeklyThreshold(market)
		scope.StableWeek, scope.WeeklyReservationID = stable, reservation.ReservationID
		scope.EntryPriceMinor, scope.StagedTargetMinor, scope.EntryCostsMinor, scope.EstimatedExitCostsMinor, scope.MinimumRRPPM = "100", "1000", "1", "1", 1
		scope.Stop = productionStop{PriceMinor: "90", Source: "structure", Policy: "stop-policy-v1", Version: "stop-v1", Digest: "stop-digest",
			ObservedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), FreshUntil: now.Add(time.Hour).Format(time.RFC3339Nano)}
	}
	scope.Horizon, scope.LaneID, scope.LaneVersion = fixture.horizon, fixture.laneID, fixture.laneVersion
	for _, change := range mutate {
		change(&scope) // 서명 전에 — 서명 뒤에 고치면 매니페스트 검증이 먼저 거절해 안쪽 갈래에 못 닿는다
	}

	route, err := strategyrouter.ProductionRouteAuthorityForTest(key, fixture.horizon, fixture.laneID, fixture.laneVersion, snapshot.Digest, scope.ConfigDigest, now)
	if err != nil {
		t.Fatal(err)
	}
	fixture.target = ProductionTarget{Approved: approved, Router: route.Request()}

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	body := productionBody{SchemaVersion: productionSchema, Domain: productionDomain, SignatureAlgorithm: productionAlgorithm, KeyID: "proposal-key",
		Generation: 1, AccountRef: "acct", Market: market, RouteManifestDigest: "route-manifest", ActivationDigest: "activation",
		CalendarGeneration: "calendar-generation", CalendarDigest: "calendar-digest", SchedulerConfigVersion: "scheduler-v1",
		EvidenceDBIdentity: "evidence-db-" + string(market), Actor: "risk-committee", ObservedAt: now.Add(-time.Second).Format(time.RFC3339Nano),
		FreshUntil: now.Add(time.Minute).Format(time.RFC3339Nano), Scopes: []productionScope{scope}}
	canonical, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(productionManifest{productionBody: body, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, canonical))})
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
	fixture.fx, err = officialfx.EvidenceForAuthorityTest(currency, "KRW", map[bool]string{true: "1", false: "1300"}[kr], "1.1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	fixture.config = ProductionConfig{ConfigDir: dir, EvidencePath: evidencePath, JournalPath: journalPath, AccountRef: "acct", Market: market,
		ManifestDigest: digest(data), TrustedKeyID: "proposal-key", TrustedKey: public, ObservedAt: now, RouteManifestDigest: "route-manifest",
		ActivationDigest: "activation", CalendarGeneration: "calendar-generation", CalendarDigest: "calendar-digest", SchedulerConfigVersion: "scheduler-v1",
		EvidenceDBIdentity: "evidence-db-" + string(market)}
	return fixture
}

func a112WeeklyModelConfig(market strategyrouter.Market) string {
	return map[strategyrouter.Market]string{strategyrouter.MarketKR: "model-config-kr", strategyrouter.MarketUS: "model-config-us"}[market]
}

func a112WeeklyThreshold(market strategyrouter.Market) string {
	return map[strategyrouter.Market]string{strategyrouter.MarketKR: "threshold-kr", strategyrouter.MarketUS: "threshold-us"}[market]
}

func TestReversalAndWeeklyProposalsLoadThroughTheRealProductionLoaderInBothMarkets(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, family := range []a112LaneFamily{a112Reversal, a112Weekly} {
		for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
			t.Run(string(family)+"/"+string(market), func(t *testing.T) {
				fixture := a112PairedLaneFixture(t, market, family, now)
				batch, err := LoadProductionAuthorityBatch(context.Background(), fixture.config, []ProductionTarget{fixture.target}, fixture.fx)
				if err != nil {
					t.Fatalf("LoadProductionAuthorityBatch: %v", err)
				}
				if absence, faulted := batch.Fault(); faulted {
					t.Fatalf("the batch recorded a fault: %+v", absence)
				}
				authority, ok := batch.For(fixture.target.Approved.Symbol())
				if !ok || batch.Len() != 1 {
					t.Fatalf("batch len=%d, authority present=%v — the %s scope produced no proposal", batch.Len(), ok, family)
				}
				proposal := authority.Proposal()
				if !proposal.ValidProposal() || proposal.Quantity == 0 {
					t.Fatalf("%s proposal is not a sealed q_candidate: %+v", family, proposal)
				}
				// 정확한 계보: 경로 결정(RouteSet)이 말한 레인 · 버전 · 수평선 · 증거와 매니페스트의 캠페인 · 스냅숏.
				routed := strategyrouter.RouteSet(fixture.target.Router)
				if routed.Code != strategyrouter.RefusalNone || len(routed.Decisions) != 1 {
					t.Fatalf("arrangement: route set = %+v", routed)
				}
				decision := routed.Decisions[0]
				lineage := proposal.Lineage
				if lineage.LaneID != fixture.laneID || lineage.LaneVersion != fixture.laneVersion || lineage.Horizon != fixture.horizon ||
					decision.LaneID != fixture.laneID || lineage.RouterEvidenceDigest != decision.EvidenceDigest ||
					lineage.LaneEvidenceDigest != fixture.snapshot.Digest || lineage.CampaignID != fixture.campaignID ||
					lineage.Market != market || lineage.CandidateLifeID != fixture.target.Approved.CandidateLifeID() {
					t.Fatalf("%s lineage mismatch: %+v (route decision %+v)", family, lineage, decision)
				}
				if authority.SnapshotID() != fixture.snapshot.ID || authority.SnapshotDigest() != fixture.snapshot.Digest {
					t.Fatalf("authority snapshot %s/%s, want %s/%s", authority.SnapshotID(), authority.SnapshotDigest(), fixture.snapshot.ID, fixture.snapshot.Digest)
				}
				binding := authority.WeeklyBinding()
				switch family {
				case a112Weekly:
					if binding == nil || binding.ReservationID != fixture.reservationID || binding.StableWeek == "" || binding.RecordDigest == "" {
						t.Fatalf("weekly proposal carries no durable reservation binding: %+v", binding)
					}
				default:
					if binding != nil {
						t.Fatalf("a %s proposal minted a weekly reservation binding: %+v", family, binding)
					}
				}
			})
		}
	}
}

// weekly 의 원장 결속이 장식이 아님: 원장 파일이 없으면 배치가 서지 않고(LoadProductionAuthorityBatch 의 원장 열기 갈래), 매니페스트의 예약
// ID 가 원장의 것과 다르면 그 스코프는 제안을 내지 않는다(buildLaneInput 의 예약 대조 갈래) — 두 시장 모두.
func TestAWeeklyScopeWithoutItsDurableReservationProposesNothing(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		t.Run(string(market)+"/journal missing", func(t *testing.T) {
			fixture := a112PairedLaneFixture(t, market, a112Weekly, now)
			if err := os.Remove(fixture.config.JournalPath); err != nil {
				t.Fatal(err)
			}
			batch, err := LoadProductionAuthorityBatch(context.Background(), fixture.config, []ProductionTarget{fixture.target}, fixture.fx)
			if err == nil || batch.Len() != 0 {
				t.Fatalf("a weekly batch stood without its journal: len=%d err=%v", batch.Len(), err)
			}
		})
		t.Run(string(market)+"/reservation id differs", func(t *testing.T) {
			fixture := a112PairedLaneFixture(t, market, a112Weekly, now, func(scope *productionScope) {
				scope.WeeklyReservationID = "reservation-not-in-the-journal"
			})
			batch, err := LoadProductionAuthorityBatch(context.Background(), fixture.config, []ProductionTarget{fixture.target}, fixture.fx)
			if err != nil {
				t.Fatalf("LoadProductionAuthorityBatch: %v", err)
			}
			if batch.Len() != 0 {
				t.Fatalf("a weekly scope proposed with a reservation the journal does not hold: len=%d", batch.Len())
			}
		})
	}
}

// 결정 49 의 벽: breakout 스코프는 두 시장 모두 정상 부재(고장 아님 — 시장을 닫지 않음). 기존 시험(lost_proposal_test.go)은 KR 만 돌았다.
func TestABreakoutScopeIsAbsenceNotFaultInBothMarkets(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []strategyrouter.Market{strategyrouter.MarketKR, strategyrouter.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			laneID := breakoutlane.KRLaneID
			if market == strategyrouter.MarketUS {
				laneID = breakoutlane.USLaneID
			}
			config, target, fx := productionFixtureOn(t, market, now,
				&productionLaneOverride{laneID: laneID, version: breakoutlane.LaneVersionV1, horizon: strategyrouter.HorizonShort}, nil)
			batch, err := LoadProductionAuthorityBatch(context.Background(), config, []ProductionTarget{target}, fx)
			if err != nil {
				t.Fatalf("LoadProductionAuthorityBatch: %v", err)
			}
			if batch.Len() != 0 {
				t.Fatalf("the breakout lane produced %d proposals behind the decision-49 wall", batch.Len())
			}
			if absence, faulted := batch.Fault(); faulted {
				t.Fatalf("breakout evidence absence was recorded as a fault (the market would close every cycle): %+v", absence)
			}
		})
	}
}
