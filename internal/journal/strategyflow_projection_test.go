//go:build tossos_testseams

package journal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func TestProjectAcceptedStrategyflowLineagePairsAllEightKRUSLanes(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	if len(descriptors) != 8 {
		t.Fatalf("descriptors=%d, want paired eight", len(descriptors))
	}
	markets := map[string]int{}
	for index, descriptor := range descriptors {
		market := string(descriptor.Market)
		account := "acct-" + market
		symbol := map[string]string{"KR": "005930", "US": "AAPL"}[market]
		campaign := "campaign-projection-" + descriptor.LaneID
		entryMinor, stopMinor, targetMinor, entryMajor, stopMajor, targetMajor := pairedProjectionPrices(market)
		result, err := strategyflow.AcceptedResultForJournalTest(descriptor, account, symbol, campaign, 20, entryMinor, stopMinor, targetMinor)
		if err != nil {
			t.Fatal(err)
		}
		policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", "risk-tx-projection-"+market+string(rune('a'+index)))
		if err != nil {
			t.Fatal(err)
		}
		risk := RiskIntent{AccountRef: account, Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: entryMajor, StopPrice: stopMajor, TargetPrice: targetMajor, PolicyVersion: policy}
		request := StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk,
			ActivationManifestDigest: "sha256:activation-" + descriptor.LaneID, CreatedAt: testIssued(t)}
		first, err := ProjectAcceptedStrategyflowLineage(request)
		if err != nil {
			t.Fatalf("%s projection: %v", descriptor.LaneID, err)
		}
		replay, err := ProjectAcceptedStrategyflowLineage(request)
		if err != nil || first != replay {
			t.Fatalf("%s non-deterministic replay: err=%v", descriptor.LaneID, err)
		}
		if first.Market != market || first.Symbol != symbol || first.LaneID != descriptor.LaneID || first.LaneVersion != descriptor.LaneVersion ||
			first.Quantity != "10" || first.EntryPrice != entryMajor || first.StopPrice != stopMajor || first.TargetPrice != targetMajor || first.PolicyVersion != policy {
			t.Fatalf("%s incomplete projection: %+v", descriptor.LaneID, first)
		}
		outer, err := decodeStrategyflowRiskBinding(first.DecisionPayload)
		if err != nil {
			t.Fatal(err)
		}
		inner, err := strategyflow.VerifyAcceptedProjection(string(outer.Strategyflow))
		if err != nil || inner.ExecutionTerms().Quantity() != 20 || inner.Lineage().PositionGeneration != 1 ||
			inner.ExecutionTerms().Policy().Identity() == "engine.automation_gate/risk-policy-v1" {
			t.Fatalf("%s candidate/q_final policy domains collapsed: inner=%+v err=%v", descriptor.LaneID, inner, err)
		}
		decision, err := (DecisionRequest{ID: "decision-projection-" + descriptor.LaneID, AccountRef: account,
			SafetyClass: SafetyClassExposureRaising, Kind: KindPlace, Preimage: risk, LimitsJSON: `{"max_quantity":"10"}`,
			Nonce: "nonce-projection-" + descriptor.LaneID, IssuedAt: testIssued(t), ExpiresAt: testIssued(t).Add(time.Minute)}).build()
		if err != nil {
			t.Fatal(err)
		}
		if err := verifyStrategyRiskBinding(decision, first); err != nil {
			t.Fatalf("%s binding: %v", descriptor.LaneID, err)
		}
		markets[market]++
	}
	if markets["KR"] != 4 || markets["US"] != 4 {
		t.Fatalf("unpaired projection matrix: %+v", markets)
	}
}

func TestVerifyStrategyflowRiskBindingV2Compatibility(t *testing.T) {
	descriptor := strategyflow.Descriptors()[1]
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-v2-us", "AAPL", "campaign-v2-us", 20, "5000", "4505", "6001")
	if err != nil {
		t.Fatal(err)
	}
	projection, err := strategyflow.ProjectAccepted(result)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", "risk-v2-us")
	if err != nil {
		t.Fatal(err)
	}
	risk := RiskIntent{AccountRef: "acct-v2-us", Market: "US", Symbol: "AAPL", Side: "BUY", Quantity: "10",
		EntryPrice: "5000", StopPrice: "4505", TargetPrice: "6001", PolicyVersion: policy}
	risk, err = exactCanonicalRiskIntent(risk)
	if err != nil {
		t.Fatal(err)
	}
	created := testIssued(t)
	record := strategyflowRiskBindingPayload{SchemaVersion: strategyflowRiskBindingSchemaVersionV2,
		Strategyflow: json.RawMessage(projection.Payload()), StrategyflowPayloadDigest: projection.PayloadDigest(), RiskIntent: risk,
		ActivationManifestDigest: "sha256:historical-v2", CreatedAt: created.Format(time.RFC3339Nano)}
	identity, err := strategyflowDecisionIdentity(record)
	if err != nil {
		t.Fatal(err)
	}
	record.Identity = identity
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	lineage := strategyflowLineageFromProjection(projection, risk, record.ActivationManifestDigest, created, identity, string(payload), record.SchemaVersion)
	decision, err := (DecisionRequest{ID: "decision-v2-us", AccountRef: risk.AccountRef, SafetyClass: SafetyClassExposureRaising,
		Kind: KindPlace, Preimage: risk, LimitsJSON: `{"max_quantity":"10"}`, Nonce: "nonce-v2-us",
		IssuedAt: created, ExpiresAt: created.Add(time.Minute)}).build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(lineage.DecisionIdentity, "strategy-decision:v2:sha256:") || lineage.EntryPrice != "5000" {
		t.Fatalf("historical v2 semantics changed: %+v", lineage)
	}
	if err := verifyStrategyRiskBinding(decision, lineage); err != nil {
		t.Fatalf("historical v2 binding rejected: %v", err)
	}
}

func pairedProjectionPrices(market string) (entryMinor, stopMinor, targetMinor, entryMajor, stopMajor, targetMajor string) {
	if market == "US" {
		return "5000", "4505", "6001", "50", "45.05", "60.01"
	}
	return "50000", "45050", "60010", "50000", "45050", "60010"
}

func strategyflowFirstLegFixture(t *testing.T, j *Journal, descriptor strategyflow.Descriptor, suffix string) QFinalCampaignFirstLegRequest {
	t.Helper()
	market := string(descriptor.Market)
	account := "acct-" + strings.ToLower(market) + "-" + suffix
	symbol := map[string]string{"KR": "005930", "US": "AAPL"}[market]
	request := firstLegAtomicFixture(t, j, suffix, account, market, symbol)
	entryMinor, stopMinor, targetMinor, entryMajor, stopMajor, targetMajor := pairedProjectionPrices(market)
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, account, symbol, request.Campaign.CampaignID, 20, entryMinor, stopMinor, targetMinor)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", request.Issue.Admission.TransactionID)
	if err != nil {
		t.Fatal(err)
	}
	risk := RiskIntent{AccountRef: account, Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: entryMajor, StopPrice: stopMajor, TargetPrice: targetMajor, PolicyVersion: policy}
	request.Issue.Issue.Decision.Preimage = risk
	request.Issue.Admission.Owner.LaneID = descriptor.LaneID
	lineage, err := ProjectAcceptedStrategyflowLineage(StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk,
		ActivationManifestDigest: request.Strategy.ActivationManifestDigest, CreatedAt: request.Strategy.CreatedAt})
	if err != nil {
		t.Fatal(err)
	}
	request.Strategy.Lineage = lineage
	if descriptor.LaneID == "kr_weekly_disclosure_value_v1" || descriptor.LaneID == "us_weekly_disclosure_value_v1" {
		provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W14"
		if market == "US" {
			provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W14"
		}
		reservation, reserveErr := j.ReserveWeeklyMarket(context.Background(), WeeklyMarketReservationRequest{
			ReservationID: "weekly-" + suffix, CampaignID: request.Campaign.CampaignID, Market: market, StableWeek: stable,
			Provider: provider, TimeZone: zone, SessionDate: "2026-03-30", CalendarGeneration: "generation-A", CalendarDigest: "calendar-A",
			IdempotencyKey: "weekly-key-" + suffix, PlannedOrdinal: 1, ObservedAt: request.Strategy.CreatedAt.Add(-time.Minute),
			FreshUntil: request.Strategy.CreatedAt.Add(time.Hour), EvaluatedAt: request.Strategy.CreatedAt,
		})
		if reserveErr != nil {
			t.Fatal(reserveErr)
		}
		request.Weekly = &WeeklyFirstLegReservationBinding{ReservationID: reservation.ReservationID, StableWeek: reservation.StableWeek,
			PlannedOrdinal: reservation.PlannedOrdinal, ScopeVersion: reservation.ScopeVersion, RequestDigest: reservation.RequestDigest,
			RecordDigest: reservation.RecordDigest, CalendarGeneration: reservation.CalendarGeneration, CalendarDigest: reservation.CalendarDigest}
	}
	return request
}

func TestRecordQFinalCampaignFirstLegAcceptsAllStrategyflowLanesPaired(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	markets := map[string]int{}
	for index, descriptor := range descriptors {
		descriptor := descriptor
		t.Run(descriptor.LaneID, func(t *testing.T) {
			j := openTestJournal(t)
			request := strategyflowFirstLegFixture(t, j, descriptor, fmt.Sprintf("sf-first-%d", index))
			outer, err := decodeStrategyflowRiskBinding(request.Strategy.Lineage.DecisionPayload)
			if err != nil {
				t.Fatal(err)
			}
			inner, err := strategyflow.VerifyAcceptedProjection(string(outer.Strategyflow))
			if err != nil || inner.ExecutionTerms().Quantity() != 20 || request.Strategy.Lineage.Quantity != "10" ||
				inner.Lineage().PositionGeneration != 1 || request.Campaign.ExpectedPositionGeneration != 0 || request.Campaign.ExpectedPositionVersion != 0 ||
				strings.HasPrefix(request.Strategy.Lineage.PolicyVersion, inner.ExecutionTerms().Policy().Identity()) {
				t.Fatalf("candidate/q_final or generation domains collapsed: inner=%+v request=%+v err=%v", inner, request, err)
			}
			receipt, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Market != string(descriptor.Market) || receipt.QFinal != 10 || receipt.AttemptID != request.Strategy.AttemptID || receipt.Idempotent {
				t.Fatalf("receipt=%+v", receipt)
			}
			if got := countRiskBucketRows(t, j, "strategy_dispatch_leases"); got != 0 {
				t.Fatalf("projection/first-leg created %d dispatch leases", got)
			}
			markets[string(descriptor.Market)]++
		})
	}
	if markets["KR"] != 4 || markets["US"] != 4 {
		t.Fatalf("unpaired first-leg matrix: %+v", markets)
	}
}

func TestRecordQFinalCampaignFirstLegRejectsStrategyflowTamperWithoutRows(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {
		market := string(descriptor.Market)
		for _, test := range []struct {
			name   string
			mutate func(*QFinalCampaignFirstLegRequest)
		}{
			{"cross-market-column", func(v *QFinalCampaignFirstLegRequest) {
				v.Strategy.Lineage.Market = map[string]string{"KR": "US", "US": "KR"}[market]
			}},
			{"q-final-column", func(v *QFinalCampaignFirstLegRequest) { v.Strategy.Lineage.Quantity = "9" }},
			{"policy-column", func(v *QFinalCampaignFirstLegRequest) { v.Strategy.Lineage.PolicyVersion = "forged" }},
			{"outer-schema", func(v *QFinalCampaignFirstLegRequest) {
				v.Strategy.Lineage.DecisionPayload = strings.Replace(v.Strategy.Lineage.DecisionPayload, strategyflowRiskBindingSchemaVersion, "journal-strategyflow-risk-binding:v9", 1)
			}},
			{"inner-lineage-seal", func(v *QFinalCampaignFirstLegRequest) {
				v.Strategy.Lineage.DecisionPayload = strings.Replace(v.Strategy.Lineage.DecisionPayload, `strategy-lineage:v1:sha256:`, `strategy-lineage:v1:sha256:0`, 1)
			}},
		} {
			t.Run(market+"/"+test.name, func(t *testing.T) {
				j := openTestJournal(t)
				request := strategyflowFirstLegFixture(t, j, descriptor, "sf-tamper-"+strings.ToLower(market)+"-"+test.name)
				test.mutate(&request)
				if _, err := j.RecordQFinalCampaignFirstLeg(context.Background(), request); err == nil {
					t.Fatal("tampered strategyflow first-leg accepted")
				}
				for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "strategy_attempt_lineage", "position_campaigns", "position_campaign_claims", "campaign_legs", "risk_bucket_final_decisions", "risk_bucket_owners", "risk_bucket_reservations", "strategy_first_leg_bindings", "strategy_dispatch_leases"} {
					if got := countRiskBucketRows(t, j, table); got != 0 {
						t.Fatalf("%s retained %d rows after refusal", table, got)
					}
				}
			})
		}
	}
}

func strategyflowBindingFixture(t *testing.T, descriptor strategyflow.Descriptor) (Decision, StrategyDecisionLineage) {
	t.Helper()
	market := string(descriptor.Market)
	account := "acct-binding-" + strings.ToLower(market)
	symbol := map[string]string{"KR": "005930", "US": "AAPL"}[market]
	entryMinor, stopMinor, targetMinor, entryMajor, stopMajor, targetMajor := pairedProjectionPrices(market)
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, account, symbol, "campaign-binding-"+market, 20, entryMinor, stopMinor, targetMinor)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", "risk-tx-binding-"+market)
	if err != nil {
		t.Fatal(err)
	}
	risk := RiskIntent{AccountRef: account, Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: entryMajor, StopPrice: stopMajor, TargetPrice: targetMajor, PolicyVersion: policy}
	lineage, err := ProjectAcceptedStrategyflowLineage(StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk,
		ActivationManifestDigest: "sha256:activation-binding-" + market, CreatedAt: testIssued(t)})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := (DecisionRequest{ID: "decision-binding-" + market, AccountRef: account, SafetyClass: SafetyClassExposureRaising,
		Kind: KindPlace, Preimage: risk, LimitsJSON: `{"max_quantity":"10"}`, Nonce: "nonce-binding-" + market,
		IssuedAt: testIssued(t), ExpiresAt: testIssued(t).Add(time.Minute)}).build()
	if err != nil {
		t.Fatal(err)
	}
	return decision, lineage
}

func TestVerifyAcceptedStrategyflowProjectionRejectsLineageColumnDrift(t *testing.T) {
	for _, descriptor := range []strategyflow.Descriptor{strategyflow.Descriptors()[0], strategyflow.Descriptors()[1]} {
		decision, base := strategyflowBindingFixture(t, descriptor)
		tests := []struct {
			name   string
			mutate func(*StrategyDecisionLineage)
		}{
			{"identity", func(v *StrategyDecisionLineage) { v.DecisionIdentity += "x" }},
			{"candidate", func(v *StrategyDecisionLineage) { v.CandidateLifeID += "x" }},
			{"market", func(v *StrategyDecisionLineage) { v.Market = map[string]string{"KR": "US", "US": "KR"}[v.Market] }},
			{"symbol", func(v *StrategyDecisionLineage) { v.Symbol += "X" }},
			{"threshold-version", func(v *StrategyDecisionLineage) { v.ThresholdVersion += "x" }},
			{"threshold-digest", func(v *StrategyDecisionLineage) { v.ThresholdSetDigest += "x" }},
			{"evidence", func(v *StrategyDecisionLineage) { v.EvidenceDigest += "x" }},
			{"consumed-id", func(v *StrategyDecisionLineage) { v.ConsumedEvidenceSnapshotID = "forged" }},
			{"consumed-digest", func(v *StrategyDecisionLineage) { v.ConsumedEvidenceSnapshotDigest = "forged" }},
			{"lane", func(v *StrategyDecisionLineage) { v.LaneID += "x" }},
			{"lane-version", func(v *StrategyDecisionLineage) { v.LaneVersion += "x" }},
			{"lane-source", func(v *StrategyDecisionLineage) { v.LaneSourceDigest += "x" }},
			{"lane-constants", func(v *StrategyDecisionLineage) { v.LaneConstantsDigest += "x" }},
			{"entry", func(v *StrategyDecisionLineage) { v.EntryPrice = "6" }},
			{"stop", func(v *StrategyDecisionLineage) { v.StopPrice = "3" }},
			{"target", func(v *StrategyDecisionLineage) { v.TargetPrice = "8" }},
			{"quantity", func(v *StrategyDecisionLineage) { v.Quantity = "9" }},
			{"policy", func(v *StrategyDecisionLineage) { v.PolicyVersion += "x" }},
			{"settings", func(v *StrategyDecisionLineage) { v.SettingsDigest += "x" }},
			{"payload", func(v *StrategyDecisionLineage) { v.DecisionPayload += " " }},
			{"payload-digest", func(v *StrategyDecisionLineage) { v.DecisionPayloadDigest += "x" }},
			{"activation", func(v *StrategyDecisionLineage) { v.ActivationManifestDigest += "x" }},
			{"created", func(v *StrategyDecisionLineage) { v.CreatedAt = v.CreatedAt.Add(time.Nanosecond) }},
		}
		for _, test := range tests {
			t.Run(string(descriptor.Market)+"/"+test.name, func(t *testing.T) {
				lineage := base
				test.mutate(&lineage)
				if err := verifyStrategyRiskBinding(decision, lineage); err == nil {
					t.Fatal("drifted persisted lineage accepted")
				}
			})
		}
	}
}

func TestVerifyAcceptedStrategyflowProjectionRejectsOuterCanonicalTamper(t *testing.T) {
	decision, base := strategyflowBindingFixture(t, strategyflow.Descriptors()[0])
	wrongHash := decision
	wrongHash.RiskHash = "forged"
	if err := verifyStrategyRiskBinding(wrongHash, base); err == nil {
		t.Fatal("RiskIntent hash drift accepted")
	}
	badPreimage := decision
	badPreimage.RiskPreimage = `{"kind":"risk-intent"}`
	badPreimage.RiskHash = HashPreimage(badPreimage.RiskPreimage)
	if err := verifyStrategyRiskBinding(badPreimage, base); err == nil {
		t.Fatal("malformed RiskIntent accepted")
	}
	reduction := ReductionIntent{AccountRef: decision.AccountRef, Market: "KR", Symbol: "005930", Side: "SELL", MaxQuantity: "1", Reason: "test"}
	reductionPayload, err := reduction.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	wrongKind := decision
	wrongKind.PreimageKind, wrongKind.RiskPreimage, wrongKind.RiskHash = PreimageKindReductionIntent, reductionPayload, HashPreimage(reductionPayload)
	if err := verifyStrategyRiskBinding(wrongKind, base); err == nil {
		t.Fatal("non-RiskIntent preimage accepted")
	}
	wrongAccount := decision
	wrongAccount.AccountRef += "-other"
	if err := verifyStrategyRiskBinding(wrongAccount, base); err == nil {
		t.Fatal("decision account drift accepted")
	}
	payload := base.DecisionPayload
	for name, tampered := range map[string]string{
		"unknown":     strings.TrimSuffix(payload, "}") + `,"unknown":true}`,
		"trailing":    payload + `{}`,
		"whitespace":  " " + payload,
		"schema":      strings.Replace(payload, strategyflowRiskBindingSchemaVersion, "journal-strategyflow-risk-binding:v9", 1),
		"schema-type": strings.Replace(payload, `"schema_version":"`+strategyflowRiskBindingSchemaVersion+`"`, `"schema_version":123`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			lineage := base
			lineage.DecisionPayload = tampered
			if err := verifyStrategyRiskBinding(decision, lineage); err == nil {
				t.Fatal("outer canonical tamper accepted")
			}
		})
	}
}

func TestProjectAcceptedStrategyflowLineageRejectsExactRiskDrift(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {
		market := string(descriptor.Market)
		account := "acct-" + market
		symbol := map[string]string{"KR": "005930", "US": "AAPL"}[market]
		entryMinor, stopMinor, targetMinor, entryMajor, stopMajor, targetMajor := pairedProjectionPrices(market)
		result, err := strategyflow.AcceptedResultForJournalTest(descriptor, account, symbol, "campaign-"+market, 20, entryMinor, stopMinor, targetMinor)
		if err != nil {
			t.Fatal(err)
		}
		policy, _ := QFinalPolicyVersion("engine.automation_gate/risk-policy-v1", "risk-tx-"+market)
		base := RiskIntent{AccountRef: account, Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: entryMajor, StopPrice: stopMajor, TargetPrice: targetMajor, PolicyVersion: policy}
		tests := []struct {
			name   string
			mutate func(*RiskIntent)
		}{
			{"account", func(v *RiskIntent) { v.AccountRef += "-other" }},
			{"cross-market", func(v *RiskIntent) {
				v.Market = map[string]string{"KR": "US", "US": "KR"}[market]
				v.Symbol = map[string]string{"KR": "AAPL", "US": "005930"}[market]
			}},
			{"symbol", func(v *RiskIntent) { v.Symbol += "X" }},
			{"q-final-oversize", func(v *RiskIntent) { v.Quantity = "21" }},
			{"q-final-zero", func(v *RiskIntent) { v.Quantity = "0" }},
			{"q-final-fractional", func(v *RiskIntent) { v.Quantity = "9.5" }},
			{"q-final-alias", func(v *RiskIntent) { v.Quantity = "10 " }},
			{"entry", func(v *RiskIntent) { v.EntryPrice = "6" }},
			{"stop", func(v *RiskIntent) { v.StopPrice = "3" }},
			{"target", func(v *RiskIntent) { v.TargetPrice = "8" }},
		}
		for _, test := range tests {
			t.Run(market+"/"+test.name, func(t *testing.T) {
				risk := base
				test.mutate(&risk)
				if _, err := ProjectAcceptedStrategyflowLineage(StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk,
					ActivationManifestDigest: "activation", CreatedAt: testIssued(t)}); err == nil {
					t.Fatal("drifted RiskIntent projected")
				}
			})
		}
	}
}
