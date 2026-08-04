//go:build tossos_testseams

package continuationlane

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

func TestProductionProposalAuthoritiesDeliverKRUSFromImmutableSnapshots(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []Market{MarketKR, MarketUS} {
		input := productionProposalFixture(t, market, now)
		if !input.Snapshot.Valid() {
			t.Fatalf("%s fixture snapshot is invalid: %+v", market, input.Snapshot)
		}
		kind := map[Market]strategyevidence.EvidenceKind{MarketKR: strategyevidence.KindKRNetFlow, MarketUS: strategyevidence.KindUSParticipation}[market]
		schema := map[Market]string{MarketKR: KRFlowSchemaV1, MarketUS: USParticipationSchemaV1}[market]
		header, _, evidenceErr := productionEvidence(input, kind, schema)
		if evidenceErr != nil {
			t.Fatalf("%s production evidence: %v header=%+v snapshot=%+v", market, evidenceErr, header, input.Snapshot)
		}
		if _, contextErr := productionContext(input, productionEnvelope(input, header, schema)); contextErr != nil {
			t.Fatalf("%s production context: %v", market, contextErr)
		}
		if market == MarketKR {
			authority, err := BuildProductionKRProposalAuthority(input)
			if err != nil {
				t.Fatal(err)
			}
			outcome := ProposeKR(authority.Request())
			if outcome.Code != RefusalNone || outcome.Quantity != 8 || outcome.Lineage.EvidenceDigest != input.Snapshot.Digest ||
				authority.SnapshotID() != input.Snapshot.ID || authority.SnapshotDigest() != input.Snapshot.Digest {
				t.Fatalf("KR proposal=%+v", outcome)
			}
		} else {
			authority, err := BuildProductionUSProposalAuthority(input)
			if err != nil {
				t.Fatal(err)
			}
			outcome := ProposeUS(authority.Request())
			if outcome.Code != RefusalNone || outcome.Quantity != 8 || outcome.Lineage.EvidenceDigest != input.Snapshot.Digest ||
				authority.SnapshotID() != input.Snapshot.ID || authority.SnapshotDigest() != input.Snapshot.Digest {
				t.Fatalf("US proposal=%+v", outcome)
			}
		}
	}
}

func TestProductionProposalAuthorityRejectsSnapshotTamper(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	input := productionProposalFixture(t, MarketUS, now)
	input.Snapshot.Digest = "tampered"
	if _, err := BuildProductionUSProposalAuthority(input); err == nil {
		t.Fatal("tampered immutable snapshot accepted")
	}
}

func productionProposalFixture(t *testing.T, market Market, now time.Time) ProductionProposalInput {
	t.Helper()
	clockMarket := map[Market]marketclock.Market{MarketKR: marketclock.MarketKR, MarketUS: marketclock.MarketUS}[market]
	symbol := map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market]
	kind := map[Market]strategyevidence.EvidenceKind{MarketKR: strategyevidence.KindKRNetFlow, MarketUS: strategyevidence.KindUSParticipation}[market]
	authority := map[Market]strategyevidence.SourceAuthority{MarketKR: strategyevidence.AuthorityKRX, MarketUS: strategyevidence.AuthorityTossOpenAPI}[market]
	schema := map[Market]string{MarketKR: KRFlowSchemaV1, MarketUS: USParticipationSchemaV1}[market]
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market]
	payload := `{"net_flow_notional_minor":"1","turnover_notional_minor":"10","flow_pressure_ppm":"100000"}`
	if market == MarketUS {
		payload = `{"participating_volume_shares":"2","baseline_volume_shares":"10","reference_price_minor":"100","last_price_minor":"101","participation_ppm":"200000","price_change_ppm":"10000"}`
	}
	header := strategyevidence.Header{EvidenceID: "continuation-" + string(market), Market: clockMarket, Symbol: symbol,
		IssuerIdentity: "issuer-" + symbol, IssuerMappingVersion: "mapping-v1", Kind: kind, SchemaVersion: schema, Authority: authority,
		SourceRecordID: "record-" + string(market), RevisionIdentity: "revision-1",
		MarketEffectiveDate: map[Market]string{MarketKR: "2026-08-04", MarketUS: "2026-08-03"}[market],
		SourceEventAt:       now.Add(-3 * time.Second), SourceAvailableAt: now.Add(-2 * time.Second), ObservedAt: now.Add(-time.Second),
		IngestedAt: now, Currency: currency, Unit: "minor-v1", Availability: strategyevidence.AvailabilityAvailable,
		Confidence: strategyevidence.ConfidenceVerified}
	envelope, err := strategyevidence.NewEnvelope(header, []byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	store, err := strategyevidence.Open(context.Background(), strategyevidence.Options{Path: filepath.Join(t.TempDir(), "evidence.db"), Clock: marketclock.NewFake(now)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if _, err := store.Append(context.Background(), envelope); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.SealSnapshot(context.Background(), strategyevidence.SnapshotQuery{Market: clockMarket, Symbol: symbol,
		IssuerIdentity: header.IssuerIdentity, IssuerMappingVersion: header.IssuerMappingVersion, EvaluationAt: now, IngestionCutoff: now})
	if err != nil {
		t.Fatal(err)
	}
	fx, err := officialfx.EvidenceForAuthorityTest(currency, "KRW", map[Market]string{MarketKR: "1", MarketUS: "1300"}[market], "1.1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return ProductionProposalInput{Snapshot: snapshot, FX: fx, Market: market, AccountRef: "acct", Symbol: symbol,
		CandidateID: "candidate-" + string(market), CampaignID: "campaign-" + string(market), PositionGeneration: 1,
		RiskBudgetMinor: "1000", PerShareRiskMinor: "10", PlannedQuantity: 14, PolicyDigest: "risk-policy",
		ConfigDigest: "config-" + string(market), AccountCurrency: "KRW", QuoteCurrency: currency, ThresholdSetID: "threshold-" + string(market),
		MinimumFlowPressurePPM: 100000, MinimumParticipationPPM: 200000, MinimumPriceChangePPM: 10000,
		Leg: LegProgress{Ordinal: 1}, SavedEffectiveStopMinor: "90",
		Stop: ProductionStopInput{PriceMinor: "95", Source: "structure", Policy: "stop-v1", Version: "v1", Digest: "stop-digest",
			ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)},
		EntryPriceMinor: "110", TargetPriceMinor: "130", FreshUntil: now.Add(time.Minute)}
}
