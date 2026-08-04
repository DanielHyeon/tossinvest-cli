//go:build tossos_testseams

package execgw_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// TestQFinalAccountBaseGuardianPairedKRUS is the a072 paired GREEN contract.
// The test-only sealed capability exists only in explicitly tagged binaries;
// normal builds must present opaque officialfx.Evidence.
func TestQFinalAccountBaseGuardianPairedKRUS(t *testing.T) {
	for _, test := range []struct {
		name, market, rate string
		costMarket         costs.Market
	}{
		{name: "KR identity", market: "KR", costMarket: costs.MarketKR, rate: "1"},
		{name: "US official", market: "US", costMarket: costs.MarketUS, rate: "1400"},
	} {
		t.Run(test.name, func(t *testing.T) {
			decisionID := "qfinal-account-base-" + strings.ToLower(test.market)
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			})
			request := qFinalAccountBaseRequest(t, rig, strings.ToLower(test.market), test.market)
			fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, test.costMarket, guardianPolicy(), test.rate, "1")
			if err != nil {
				t.Fatal(err)
			}
			bindPairedAccountBaseFXForTest(&request, fx)
			precheck, err := rig.guardian.PrecheckQFinalEntry(request)
			if err != nil {
				t.Fatalf("%s account-base precheck: %v", test.market, err)
			}
			if precheck.QCandidate() != 20 || precheck.QFinal() != 10 || rig.collections != 0 {
				t.Fatalf("%s candidate=%d final=%d collections=%d, want 20/10/0", test.market, precheck.QCandidate(), precheck.QFinal(), rig.collections)
			}
			issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
			if err != nil {
				t.Fatalf("%s account-base issue: %v", test.market, err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			limits, err := execgw.DecodeLimits(decision.LimitsJSON)
			if err != nil {
				t.Fatal(err)
			}
			if limits.AccountBaseFX == nil || limits.AccountBaseFX.QuoteCurrency != request.Currency ||
				limits.AccountBaseFX.AccountCurrency != "KRW" || limits.AccountBaseFX.EvidenceDigest != fx.Digest() {
				t.Fatalf("%s persisted FX binding = %+v", test.market, limits.AccountBaseFX)
			}
		})
	}
}

func TestQFinalRejectsDifferentGuardianAndBucketFXPairedKRUS(t *testing.T) {
	for _, test := range []struct {
		market, rate string
		costMarket   costs.Market
	}{
		{market: "KR", costMarket: costs.MarketKR, rate: "1"},
		{market: "US", costMarket: costs.MarketUS, rate: "1400"},
	} {
		t.Run(test.market, func(t *testing.T) {
			rig := newGuardian(t, nil)
			request := qFinalAccountBaseRequest(t, rig, "fx-divergence-"+strings.ToLower(test.market), test.market)
			fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, test.costMarket, guardianPolicy(), test.rate, "1")
			if err != nil {
				t.Fatal(err)
			}
			bindPairedAccountBaseFXForTest(&request, fx)
			request.Admission.Admission.Policy.FX.Evidence.Digest = "sha256:" + strings.Repeat("f", 64)
			request.SetFXAuthorityForTest(request.Admission.Admission.Policy.FX)
			if _, err := rig.guardian.PrecheckQFinalEntry(request); err == nil {
				t.Fatal("different Guardian/bucket FX authorities were accepted")
			}
			if rig.collections != 0 {
				t.Fatalf("divergent FX recollected %d times, want zero", rig.collections)
			}
		})
	}
}

func TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllSixLanes(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	if len(descriptors) != 6 {
		t.Fatalf("descriptors=%d, want six paired lanes", len(descriptors))
	}
	markets := map[string]int{}
	for _, descriptor := range descriptors {
		descriptor := descriptor
		t.Run(descriptor.LaneID, func(t *testing.T) {
			market := string(descriptor.Market)
			suffix := "first-leg-" + descriptor.LaneID
			decisionID := "qfinal-" + suffix
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			})
			entry := qFinalAccountBaseRequest(t, rig, suffix, market)
			entry.Admission.Owner.Key.ProspectiveGeneration = ""
			entry.Admission.Owner.LaneID = descriptor.LaneID
			entry.Admission.Owner.CampaignID = "campaign-" + suffix
			costMarket := map[string]costs.Market{"KR": costs.MarketKR, "US": costs.MarketUS}[market]
			rate := map[string]string{"KR": "1", "US": "1400"}[market]
			fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, costMarket, guardianPolicy(), rate, "1")
			if err != nil {
				t.Fatal(err)
			}
			bindPairedAccountBaseFXForTest(&entry, fx)
			entryMinor, stopMinor, targetMinor := pairedStrategyflowMinorPrices(entry)
			result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-7", entry.Symbol,
				entry.Admission.Owner.CampaignID, 20, entryMinor, stopMinor, targetMinor)
			if err != nil {
				t.Fatal(err)
			}
			manifest := "sha256:" + strings.Repeat("5", 64)
			request := execgw.QFinalCampaignFirstLegIssuance{
				Entry: entry, Result: result, ActivationManifestDigest: manifest,
				AttemptID: "attempt-" + suffix, Revision: 1,
				Campaign: journal.FirstLegCampaignRequest{CampaignID: entry.Admission.Owner.CampaignID,
					ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
					CreateCommandKey: "create-" + suffix, FirstLegCommandKey: "command-" + suffix, FirstLegPlanID: "plan-" + suffix},
			}
			request.Weekly = reserveWeeklyFirstLegForTest(t, rig, descriptor, request.Campaign.CampaignID, suffix)
			precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
			if err != nil {
				t.Fatalf("%s first-leg precheck: %v: %v", descriptor.LaneID, err, errors.Unwrap(err))
			}
			if precheck.QCandidate() != 20 || precheck.QFinal() != 10 || rig.collections != 0 {
				t.Fatalf("%s precheck candidate/final/collections=%d/%d/%d", descriptor.LaneID, precheck.QCandidate(), precheck.QFinal(), rig.collections)
			}
			receipt, err := rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck)
			if err != nil {
				t.Fatalf("%s atomic first-leg issue: %v", descriptor.LaneID, err)
			}
			if receipt.DecisionID != decisionID || receipt.QFinal != 10 || receipt.Market != market ||
				receipt.AttemptID != request.AttemptID || receipt.CampaignID != request.Campaign.CampaignID {
				t.Fatalf("%s first-leg receipt = %+v", descriptor.LaneID, receipt)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), decisionID)
			if err != nil {
				t.Fatal(err)
			}
			wantPolicy, err := journal.QFinalPolicyVersion(rig.guardian.PolicyVersion(), entry.Admission.TransactionID)
			if err != nil {
				t.Fatal(err)
			}
			preimage, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
			if err != nil {
				t.Fatal(err)
			}
			intent, ok := preimage.(journal.RiskIntent)
			if !ok || intent.Quantity != "10" || intent.PolicyVersion == result.ExecutionTerms.Policy().Identity() ||
				intent.PolicyVersion != wantPolicy {
				t.Fatalf("%s Guardian policy/q_final collapsed: %#v", descriptor.LaneID, preimage)
			}
			if plans, err := rig.journal.PendingStrategyPlans(context.Background(), "acct-7"); err != nil || len(plans) != 0 {
				t.Fatalf("%s standalone/dispatch plan leaked: plans=%+v err=%v", descriptor.LaneID, plans, err)
			}
			markets[market]++
		})
	}
	if markets["KR"] != 3 || markets["US"] != 3 {
		t.Fatalf("unpaired lane matrix: %+v", markets)
	}
}

func TestQFinalCampaignFirstLegWeeklyBindingPairedKRUS(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	for _, descriptor := range descriptors {
		if descriptor.LaneID != weeklyvaluelane.KRWeeklyLaneID && descriptor.LaneID != weeklyvaluelane.USWeeklyLaneID {
			continue
		}
		descriptor := descriptor
		t.Run(string(descriptor.Market), func(t *testing.T) {
			decisionID := "weekly-opaque-" + strings.ToLower(string(descriptor.Market))
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			})
			request := qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, decisionID)
			precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
			if err != nil {
				t.Fatal(err)
			}
			request.Weekly.RecordDigest = "caller-mutated-after-precheck"
			receipt, err := rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck)
			if err != nil || receipt.DecisionID != decisionID {
				t.Fatalf("opaque weekly issue receipt=%+v err=%v", receipt, err)
			}

			badID := decisionID + "-bad"
			badRig := newGuardian(t, func(options *execgw.RiskGuardianOptions) { options.NewID = fixedIDs(badID, badID+"-nonce") })
			bad := qFinalCampaignFirstLegIssuanceFixture(t, badRig, descriptor, badID)
			bad.Weekly.StableWeek = map[strategyrouter.Market]string{
				strategyrouter.MarketKR: "US-XNYS-2026-W14", strategyrouter.MarketUS: "KR-XKRX-2026-W14",
			}[descriptor.Market]
			badPrecheck, err := badRig.guardian.PrecheckQFinalCampaignFirstLeg(bad)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := badRig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), badPrecheck); err == nil {
				t.Fatal("cross-market weekly reservation was issued")
			}
			if decisionOnDisk(t, badRig.journal, badID) {
				t.Fatal("cross-market weekly refusal persisted a decision")
			}
		})
	}

	for _, descriptor := range descriptors[:2] {
		rig := newGuardian(t, nil)
		request := qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, "non-weekly-smuggle-"+string(descriptor.Market))
		request.Weekly = &journal.WeeklyFirstLegReservationBinding{ReservationID: "smuggled", StableWeek: "smuggled", PlannedOrdinal: 1,
			ScopeVersion: 1, RequestDigest: "smuggled", RecordDigest: "smuggled", CalendarGeneration: "smuggled", CalendarDigest: "smuggled"}
		if _, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request); err == nil {
			t.Fatalf("%s non-weekly lane accepted weekly authority", descriptor.Market)
		}
	}
}

func TestQFinalCampaignFirstLegRejectsTamperWithoutWritesPairedKRUS(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {
		market := string(descriptor.Market)
		for _, test := range []struct {
			name   string
			mutate func(*execgw.QFinalCampaignFirstLegIssuance)
		}{
			{name: "candidate", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Entry.QCandidate = 19 }},
			{name: "campaign", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Campaign.CampaignID += "-peer" }},
			{name: "generation", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Campaign.ExpectedPositionGeneration = 1 }},
			{name: "activation", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.ActivationManifestDigest = "" }},
			{name: "projection-purity", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Result.CommonSafetyIndependent = false }},
		} {
			t.Run(market+"-"+test.name, func(t *testing.T) {
				decisionID := "tamper-" + strings.ToLower(market) + "-" + test.name
				rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) { options.NewID = fixedIDs(decisionID) })
				entry := qFinalAccountBaseRequest(t, rig, decisionID, market)
				entry.Admission.Owner.Key.ProspectiveGeneration = ""
				entry.Admission.Owner.LaneID = descriptor.LaneID
				entry.Admission.Owner.CampaignID = "campaign-" + decisionID
				fx, err := risk.TestOnlySealAccountBaseFX(fixedNow,
					map[string]costs.Market{"KR": costs.MarketKR, "US": costs.MarketUS}[market], guardianPolicy(),
					map[string]string{"KR": "1", "US": "1400"}[market], "1")
				if err != nil {
					t.Fatal(err)
				}
				bindPairedAccountBaseFXForTest(&entry, fx)
				entryMinor, stopMinor, targetMinor := pairedStrategyflowMinorPrices(entry)
				result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-7", entry.Symbol,
					entry.Admission.Owner.CampaignID, 20, entryMinor, stopMinor, targetMinor)
				if err != nil {
					t.Fatal(err)
				}
				request := execgw.QFinalCampaignFirstLegIssuance{Entry: entry, Result: result,
					ActivationManifestDigest: "sha256:" + strings.Repeat("a", 64), AttemptID: "attempt-" + decisionID, Revision: 1,
					Campaign: journal.FirstLegCampaignRequest{CampaignID: entry.Admission.Owner.CampaignID,
						ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0, CreateCommandKey: "create-" + decisionID,
						FirstLegCommandKey: "leg-" + decisionID, FirstLegPlanID: "plan-" + decisionID}}
				test.mutate(&request)
				if _, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request); err == nil {
					t.Fatal("tampered issuance was accepted")
				}
				if rig.collections != 0 || decisionOnDisk(t, rig.journal, decisionID) {
					t.Fatalf("tamper caused collection/write: collections=%d", rig.collections)
				}
			})
		}
	}
}

func TestQFinalCampaignFirstLegRejectsZeroPrecheckWithoutWritesPairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			decisionID := "zero-precheck-" + strings.ToLower(market)
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) { options.NewID = fixedIDs(decisionID) })
			if _, err := rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), execgw.QFinalCampaignFirstLegPrecheck{}); err == nil {
				t.Fatal("zero precheck was issued")
			}
			if rig.collections != 0 || decisionOnDisk(t, rig.journal, decisionID) {
				t.Fatalf("zero precheck caused collection/write: collections=%d", rig.collections)
			}
		})
	}
}

func TestQFinalCampaignFirstLegIssueRefusalsWriteZeroPairedKRUS(t *testing.T) {
	descriptors := strategyflow.Descriptors()
	for _, descriptor := range []strategyflow.Descriptor{descriptors[0], descriptors[1]} {
		market := string(descriptor.Market)
		for _, test := range []struct {
			name        string
			mutateEntry func(*execgw.QFinalEntryIssuance, *guardianRig)
			afterCheck  func(*guardianRig)
		}{
			{name: "expired", afterCheck: func(rig *guardianRig) { rig.clock.Advance(2 * time.Minute) }},
			{name: "collection", mutateEntry: func(entry *execgw.QFinalEntryIssuance, _ *guardianRig) {
				entry.Collect = func(context.Context, int) (execgw.ExposureSnapshot, error) {
					return execgw.ExposureSnapshot{}, errors.New("snapshot unavailable")
				}
			}},
			{name: "usage-currency", mutateEntry: func(entry *execgw.QFinalEntryIssuance, rig *guardianRig) {
				entry.Collect = func(ctx context.Context, attempt int) (execgw.ExposureSnapshot, error) {
					snapshot, err := rig.collect(ctx, attempt)
					snapshot.OpenExposure.Currency = "USD"
					return snapshot, err
				}
			}},
		} {
			t.Run(market+"-"+test.name, func(t *testing.T) {
				decisionID := "issue-refusal-" + strings.ToLower(market) + "-" + test.name
				rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) { options.NewID = fixedIDs(decisionID, decisionID+"-nonce") })
				request := qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, decisionID)
				if test.mutateEntry != nil {
					test.mutateEntry(&request.Entry, rig)
				}
				precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
				if err != nil {
					t.Fatal(err)
				}
				if test.afterCheck != nil {
					test.afterCheck(rig)
				}
				if _, err := rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck); err == nil {
					t.Fatal("issue refusal case was admitted")
				}
				if decisionOnDisk(t, rig.journal, decisionID) {
					t.Fatal("issue refusal persisted a decision")
				}
			})
		}
	}
}

func qFinalCampaignFirstLegIssuanceFixture(t *testing.T, rig *guardianRig, descriptor strategyflow.Descriptor, suffix string) execgw.QFinalCampaignFirstLegIssuance {
	t.Helper()
	market := string(descriptor.Market)
	entry := qFinalAccountBaseRequest(t, rig, suffix, market)
	entry.Admission.Owner.Key.ProspectiveGeneration = ""
	entry.Admission.Owner.LaneID = descriptor.LaneID
	entry.Admission.Owner.CampaignID = "campaign-" + suffix
	fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, map[string]costs.Market{"KR": costs.MarketKR, "US": costs.MarketUS}[market],
		guardianPolicy(), map[string]string{"KR": "1", "US": "1400"}[market], "1")
	if err != nil {
		t.Fatal(err)
	}
	bindPairedAccountBaseFXForTest(&entry, fx)
	entryMinor, stopMinor, targetMinor := pairedStrategyflowMinorPrices(entry)
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-7", entry.Symbol, entry.Admission.Owner.CampaignID,
		20, entryMinor, stopMinor, targetMinor)
	if err != nil {
		t.Fatal(err)
	}
	request := execgw.QFinalCampaignFirstLegIssuance{Entry: entry, Result: result,
		ActivationManifestDigest: "sha256:" + strings.Repeat("a", 64), AttemptID: "attempt-" + suffix, Revision: 1,
		Campaign: journal.FirstLegCampaignRequest{CampaignID: entry.Admission.Owner.CampaignID,
			ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0, CreateCommandKey: "create-" + suffix,
			FirstLegCommandKey: "leg-" + suffix, FirstLegPlanID: "plan-" + suffix}}
	request.Weekly = reserveWeeklyFirstLegForTest(t, rig, descriptor, request.Campaign.CampaignID, suffix)
	return request
}

func pairedStrategyflowMinorPrices(entry execgw.QFinalEntryIssuance) (string, string, string) {
	if entry.Market == "US" {
		return "5000", "4500", "6000"
	}
	return entry.EntryPrice, entry.StopPrice, entry.TargetPrice
}

func reserveWeeklyFirstLegForTest(t *testing.T, rig *guardianRig, descriptor strategyflow.Descriptor, campaignID, suffix string) *journal.WeeklyFirstLegReservationBinding {
	t.Helper()
	if descriptor.LaneID != weeklyvaluelane.KRWeeklyLaneID && descriptor.LaneID != weeklyvaluelane.USWeeklyLaneID {
		return nil
	}
	market := string(descriptor.Market)
	provider, zone, stable := "XKRX_OFFICIAL", "Asia/Seoul", "KR-XKRX-2026-W14"
	if market == "US" {
		provider, zone, stable = "XNYS_OFFICIAL", "America/New_York", "US-XNYS-2026-W14"
	}
	snapshot, err := rig.journal.ReserveWeeklyMarket(context.Background(), journal.WeeklyMarketReservationRequest{
		ReservationID: "weekly-" + suffix, CampaignID: campaignID, Market: market, StableWeek: stable,
		Provider: provider, TimeZone: zone, SessionDate: "2026-03-30", CalendarGeneration: "calendar-generation-" + market,
		CalendarDigest: "calendar-digest-" + market, IdempotencyKey: "weekly-idempotency-" + suffix,
		PlannedOrdinal: 1, ExpectedVersion: 0, ObservedAt: fixedNow.Add(-time.Minute), FreshUntil: fixedNow.Add(time.Hour), EvaluatedAt: fixedNow,
	})
	if err != nil {
		t.Fatalf("reserve weekly %s: %v", descriptor.LaneID, err)
	}
	return &journal.WeeklyFirstLegReservationBinding{ReservationID: snapshot.ReservationID, StableWeek: snapshot.StableWeek,
		PlannedOrdinal: snapshot.PlannedOrdinal, ScopeVersion: snapshot.ScopeVersion, RequestDigest: snapshot.RequestDigest,
		RecordDigest: snapshot.RecordDigest, CalendarGeneration: snapshot.CalendarGeneration, CalendarDigest: snapshot.CalendarDigest}
}

func bindPairedAccountBaseFXForTest(request *execgw.QFinalEntryIssuance, fx risk.AccountBaseFX) {
	request.Admission.Admission.Policy.FX.Evidence.Source = fx.Source()
	request.Admission.Admission.Policy.FX.Evidence.Version = fx.Version()
	request.Admission.Admission.Policy.FX.Evidence.Digest = fx.Digest()
	request.SetFXAuthorityForTest(request.Admission.Admission.Policy.FX)
	request.SetAccountBaseFXForTest(fx)
}
