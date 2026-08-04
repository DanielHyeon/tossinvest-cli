//go:build tossos_testseams

package reversallane

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

func TestProductionProposalAuthoritiesDeliverPairedKRUSFirstLeg(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []Market{MarketKR, MarketUS} {
		input := productionProposalFixture(t, market, now, 1)
		if market == MarketKR {
			authority, err := BuildProductionKRProposalAuthority(input)
			if err != nil {
				t.Fatal(err)
			}
			outcome := ProposeKR(authority.Request())
			assertProductionProposal(t, market, outcome, input, authority.SnapshotID(), authority.SnapshotDigest(), 2)
			continue
		}
		authority, err := BuildProductionUSProposalAuthority(input)
		if err != nil {
			t.Fatal(err)
		}
		outcome := ProposeUS(authority.Request())
		assertProductionProposal(t, market, outcome, input, authority.SnapshotID(), authority.SnapshotDigest(), 2)
	}
}

func TestProductionProposalAuthoritiesRequireFinalLegStructureForKRUS(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []Market{MarketKR, MarketUS} {
		missing := productionProposalFixture(t, market, now, 3)
		var refused EvaluationResult
		if market == MarketKR {
			authority, err := BuildProductionKRProposalAuthority(missing)
			if err != nil {
				t.Fatal(err)
			}
			refused = ProposeKR(authority.Request())
		} else {
			authority, err := BuildProductionUSProposalAuthority(missing)
			if err != nil {
				t.Fatal(err)
			}
			refused = ProposeUS(authority.Request())
		}
		if refused.Code != RefusalStructuralMissing || refused.Quantity != 0 {
			t.Fatalf("%s final leg without structure=%+v", market, refused)
		}

		confirmed := productionProposalFixture(t, market, now, 3)
		confirmed.Structure = productionStructure(confirmed)
		if market == MarketKR {
			authority, err := BuildProductionKRProposalAuthority(confirmed)
			if err != nil {
				t.Fatal(err)
			}
			assertProductionProposal(t, market, ProposeKR(authority.Request()), confirmed, authority.SnapshotID(), authority.SnapshotDigest(), 8)
		} else {
			authority, err := BuildProductionUSProposalAuthority(confirmed)
			if err != nil {
				t.Fatal(err)
			}
			assertProductionProposal(t, market, ProposeUS(authority.Request()), confirmed, authority.SnapshotID(), authority.SnapshotDigest(), 8)
		}
	}
}

func TestProductionProposalAuthorityRejectsSnapshotTamper(t *testing.T) {
	input := productionProposalFixture(t, MarketUS, time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC), 1)
	input.Snapshot.Digest = "tampered"
	if _, err := BuildProductionUSProposalAuthority(input); err == nil {
		t.Fatal("tampered immutable snapshot accepted")
	}
}

func assertProductionProposal(t *testing.T, market Market, outcome EvaluationResult, input ProductionProposalInput, snapshotID, snapshotDigest string, quantity uint64) {
	t.Helper()
	if outcome.Code != "" || outcome.Quantity != quantity || outcome.Lineage.MetricEvidenceDigest != input.Snapshot.Digest ||
		snapshotID != input.Snapshot.ID || snapshotDigest != input.Snapshot.Digest {
		t.Fatalf("%s proposal=%+v", market, outcome)
	}
}

func productionProposalFixture(t *testing.T, market Market, now time.Time, ordinal int) ProductionProposalInput {
	t.Helper()
	clockMarket := map[Market]marketclock.Market{MarketKR: marketclock.MarketKR, MarketUS: marketclock.MarketUS}[market]
	symbol := map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market]
	kind := map[Market]strategyevidence.EvidenceKind{MarketKR: strategyevidence.KindKRNetFlow, MarketUS: strategyevidence.KindUSParticipation}[market]
	authority := map[Market]strategyevidence.SourceAuthority{MarketKR: strategyevidence.AuthorityKRX, MarketUS: strategyevidence.AuthorityTossOpenAPI}[market]
	schema := map[Market]string{MarketKR: "kr-absorption-v1", MarketUS: "us-dislocation-v1"}[market]
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market]
	payload := `{"absorbed_notional_minor":"1","aggressive_sell_notional_minor":"4","absorption_ppm":"250000"}`
	if market == MarketUS {
		payload = `{"reference_price_minor":"100","dislocation_low_price_minor":"90","dislocation_volume_shares":"150","baseline_volume_shares":"100","drawdown_ppm":"100000","relative_volume_ppm":"1500000"}`
	}
	header := strategyevidence.Header{EvidenceID: "reversal-" + string(market), Market: clockMarket, Symbol: symbol,
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
		ConfigDigest: "config-" + string(market), AccountCurrency: "KRW", QuoteCurrency: currency, ConfigVersion: "config-v1",
		ThresholdSet: "threshold-" + string(market), MinimumAbsorptionPPM: 250000, MinimumDrawdownPPM: 100000,
		MinimumRelativeVolumePPM: 1500000, StructuralWindow: time.Minute, Leg: LegProgress{Ordinal: ordinal}, SavedEffectiveStopMinor: "90",
		Stop: ProductionStopInput{PriceMinor: "95", Source: "structure", Policy: "stop-v1", Version: "v1", Digest: "stop-digest",
			ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)},
		EntryPriceMinor: "100", TargetPriceMinor: "120", FreshUntil: now.Add(time.Minute)}
}

func productionStructure(input ProductionProposalInput) StructuralConfirmation {
	makeEvent := func(kind StructuralEventKind, at time.Time) StructuralEvent {
		return StructuralEvent{Kind: kind, AccountRef: input.AccountRef, Market: input.Market, Symbol: input.Symbol,
			PositionGeneration: input.PositionGeneration, EvidenceVersion: map[Market]string{MarketKR: "kr-absorption-v1", MarketUS: "us-dislocation-v1"}[input.Market],
			RecordID: string(kind) + "-record", Digest: string(kind) + "-digest", At: at, FreshUntil: input.FreshUntil}
	}
	return StructuralConfirmation{Sweep: makeEvent(EventSweep, input.Snapshot.EvaluationAt.Add(-30*time.Second)),
		Break: makeEvent(EventBreak, input.Snapshot.EvaluationAt.Add(-20*time.Second)), Reclaim: makeEvent(EventReclaim, input.Snapshot.EvaluationAt.Add(-10*time.Second))}
}
