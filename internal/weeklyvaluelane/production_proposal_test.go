//go:build tossos_testseams

package weeklyvaluelane

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

func TestProductionProposalAuthoritiesDeliverPairedKRUSFromDurableWeeks(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	for _, market := range []Market{MarketKR, MarketUS} {
		input := productionProposalFixture(t, market, now)
		if market == MarketKR {
			authority, err := BuildProductionKRProposalAuthority(input)
			if err != nil {
				t.Fatal(err)
			}
			assertProductionProposal(t, market, ProposeKR(authority.Request()), input, authority.SnapshotID(), authority.SnapshotDigest())
			continue
		}
		authority, err := BuildProductionUSProposalAuthority(input)
		if err != nil {
			t.Fatal(err)
		}
		assertProductionProposal(t, market, ProposeUS(authority.Request()), input, authority.SnapshotID(), authority.SnapshotDigest())
	}
}

func TestProductionProposalAuthorityRejectsSnapshotAndReservationTamper(t *testing.T) {
	now := time.Date(2026, 8, 4, 1, 0, 3, 0, time.UTC)
	input := productionProposalFixture(t, MarketUS, now)
	input.Snapshot.Digest = "tampered"
	if _, err := BuildProductionUSProposalAuthority(input); err == nil {
		t.Fatal("tampered immutable evidence snapshot accepted")
	}
	input = productionProposalFixture(t, MarketUS, now)
	input.Reservation.StableWeek = "US-XNYS-2026-W33"
	if _, err := BuildProductionUSProposalAuthority(input); err == nil {
		t.Fatal("reservation projection detached from durable week accepted")
	}
}

func assertProductionProposal(t *testing.T, market Market, outcome Outcome, input ProductionProposalInput, snapshotID, snapshotDigest string) {
	t.Helper()
	if outcome.Kind != OutcomeDecision || outcome.Code != "" || outcome.Quantity != 2 || outcome.Lineage.EvidenceDigest != input.Snapshot.Digest ||
		outcome.Lineage.ReservationID != input.Reservation.ReservationID || snapshotID != input.Snapshot.ID || snapshotDigest != input.Snapshot.Digest {
		t.Fatalf("%s proposal=%+v", market, outcome)
	}
}

func productionProposalFixture(t *testing.T, market Market, now time.Time) ProductionProposalInput {
	t.Helper()
	clockMarket := map[Market]marketclock.Market{MarketKR: marketclock.MarketKR, MarketUS: marketclock.MarketUS}[market]
	symbol := map[Market]string{MarketKR: "005930", MarketUS: "AAPL"}[market]
	currency := map[Market]string{MarketKR: "KRW", MarketUS: "USD"}[market]
	source := map[Market]DisclosureSource{MarketKR: SourceOpenDART, MarketUS: SourceEDGAR}[market]
	authority := map[Market]strategyevidence.SourceAuthority{MarketKR: strategyevidence.AuthorityOpenDART, MarketUS: strategyevidence.AuthoritySEC}[market]
	schema := map[Market]string{MarketKR: KRDisclosureSchemaV1, MarketUS: USDisclosureSchemaV1}[market]
	modelConfig := map[Market]string{MarketKR: "model-config-kr", MarketUS: "model-config-us"}[market]
	threshold := map[Market]string{MarketKR: "threshold-kr", MarketUS: "threshold-us"}[market]
	filing := map[Market]string{MarketKR: "202608040001", MarketUS: "0000320193-26-000077"}[market]
	revision, superseded, sequence, scale := "rev-1", "NONE", "1", "0"
	if market == MarketUS {
		revision, superseded, sequence, scale = "rev-1", "NONE", "1", "2"
	}
	asOf := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	observed, ingested, cutoff := now.Add(-10*time.Minute), now.Add(-9*time.Minute), now.Add(-8*time.Minute)
	payload := fmt.Sprintf(`{"schema_version":%q,"market":%q,"source":%q,"symbol":%q,"issuer_id":%q,"filing_id":%q,"report_id":"quarterly-2026-q2","revision_id":%q,"superseded_revision_id":%q,"revision_sequence":%q,"as_of":%q,"observed_at":%q,"ingested_at":%q,"cutoff_at":%q,"evaluated_at":%q,"fresh_until":%q,"currency":%q,"monetary_unit":"MINOR","monetary_scale":%q,"diluted_shares":"100","shares_unit":"SHARES","dilution_status":%q,"dilution_facts_digest":"dilution-digest","dilution_as_of":%q,"financial_inputs":[{"name":"equity_value","value_minor":"110000","unit":%q}],"model_id":"weekly-value","model_version":"weekly-model-v1","model_config_digest":%q,"threshold_digest":%q,"equity_value_minor":"110000","fair_value_minor":"1100"}`,
		schema, market, source, symbol, "issuer-"+symbol, filing, revision, superseded, sequence, asOf.Format(time.RFC3339Nano),
		observed.Format(time.RFC3339Nano), ingested.Format(time.RFC3339Nano), cutoff.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano),
		now.Add(7*24*time.Hour).Format(time.RFC3339Nano), currency, scale, map[Market]string{MarketKR: "OBSERVED", MarketUS: "NONE"}[market],
		asOf.Format(time.RFC3339Nano), currency+"_MINOR", modelConfig, threshold)
	header := strategyevidence.Header{EvidenceID: "weekly-" + string(market), Market: clockMarket, Symbol: symbol, IssuerIdentity: "issuer-" + symbol,
		IssuerMappingVersion: "mapping-v1", Kind: strategyevidence.KindDisclosure, SchemaVersion: schema, Authority: authority,
		SourceRecordID: filing, RevisionIdentity: revision,
		MarketEffectiveDate: map[Market]string{MarketKR: "2026-08-04", MarketUS: "2026-08-03"}[market], SourceEventAt: asOf,
		SourceAvailableAt: observed.Add(-time.Minute), ObservedAt: observed, IngestedAt: ingested, Currency: currency, Unit: "minor-v1",
		Availability: strategyevidence.AvailabilityAvailable, Confidence: strategyevidence.ConfidenceVerified}
	envelope, err := strategyevidence.NewEnvelope(header, []byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	evidenceStore, err := strategyevidence.Open(context.Background(), strategyevidence.Options{Path: filepath.Join(t.TempDir(), "evidence.db"), Clock: marketclock.NewFake(now)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = evidenceStore.Close() })
	if _, err := evidenceStore.Append(context.Background(), envelope); err != nil {
		t.Fatal(err)
	}
	snapshot, err := evidenceStore.SealSnapshot(context.Background(), strategyevidence.SnapshotQuery{Market: clockMarket, Symbol: symbol,
		IssuerIdentity: header.IssuerIdentity, IssuerMappingVersion: header.IssuerMappingVersion, EvaluationAt: now, IngestionCutoff: now})
	if err != nil {
		t.Fatal(err)
	}

	provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W32"
	if market == MarketUS {
		provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W32"
	}
	reservation := ProductionReservationInput{ReservationID: "reservation-" + string(market), CampaignID: "campaign-" + string(market),
		Market: string(market), StableWeek: stable, Provider: provider, TimeZone: zone, SessionDate: "2026-08-03",
		CalendarGeneration: "generation-A", CalendarDigest: "calendar-A", Status: "ACTIVE", RequestDigest: "request-digest",
		RecordDigest: "record-digest", PlannedOrdinal: 1, ScopeVersion: 1, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Hour), EvaluatedAt: now}
	fx, err := officialfx.EvidenceForAuthorityTest(currency, "KRW", map[Market]string{MarketKR: "1", MarketUS: "1300"}[market], "1.1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return ProductionProposalInput{Snapshot: snapshot, FX: fx, Market: market, AccountRef: "acct", Symbol: symbol,
		CandidateID: "candidate-" + string(market), CampaignID: reservation.CampaignID, PositionGeneration: 1, RiskBudgetMinor: "1000",
		PerShareRiskMinor: "10", PlannedQuantity: 14, PolicyDigest: "risk-policy", AccountCurrency: "KRW", QuoteCurrency: currency,
		ModelVersion: "weekly-model-v1", ModelConfigDigest: modelConfig, ThresholdDigest: threshold,
		Reservation: reservation,
		Leg:         LegProgress{Ordinal: 1}, Stop: ProductionStopInput{PriceMinor: "90", Version: "stop-v1", Source: "structure", Policy: "stop-policy-v1",
			Digest: "stop-digest", ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Hour)}, SavedEffectiveStopMinor: "90",
		EntryPriceMinor: "100", StagedTargetMinor: "1000", EntryCostsMinor: "1", EstimatedExitCostsMinor: "1", MinimumRRPPM: 1}
}
