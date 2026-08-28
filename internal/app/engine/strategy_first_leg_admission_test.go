//go:build tossos_testseams

package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func TestStrategyFirstLegAdmissionDormantPairedKRUS(t *testing.T) {
	for _, market := range pairedStrategyFirstLegResults(t) {
		t.Run(string(market.result.Lineage.Market), func(t *testing.T) {
			issuer := &firstLegGuardianSpy{}
			outcome := newStrategyFirstLegAdmissionBridge(issuer, nil).admit(context.Background(), market.result)
			if outcome.Code != StrategyFirstLegAuthorityUnavailable || outcome.Market != string(market.result.Lineage.Market) ||
				issuer.precheckCalls != 0 || issuer.issueCalls != 0 {
				t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
			}
		})
	}
}

func TestStrategyFirstLegAdmissionUsesGuardianOnlyAllEightLanes(t *testing.T) {
	markets := map[string]int{}
	for _, lane := range allPairedStrategyFirstLegResults(t) {
		lane := lane
		t.Run(lane.result.Lineage.LaneID, func(t *testing.T) {
			issuance := firstLegBridgeIssuanceFixture(lane.result)
			loader := &firstLegAuthorityLoaderStub{issuance: issuance}
			issuer := &firstLegGuardianSpy{receipt: journal.QFinalCampaignFirstLegReceipt{
				DecisionID: "decision-" + lane.result.Lineage.LaneID, CampaignID: lane.result.Lineage.CampaignID,
				Market: string(lane.result.Lineage.Market), Symbol: lane.result.Lineage.Symbol, QFinal: 10,
			}}
			outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), lane.result)
			if outcome.Code != StrategyFirstLegAdmitted || loader.calls != 1 || issuer.precheckCalls != 1 || issuer.issueCalls != 1 {
				t.Fatalf("outcome=%+v loader=%d precheck=%d issue=%d", outcome, loader.calls, issuer.precheckCalls, issuer.issueCalls)
			}
			if issuer.events != "precheck,issue" || issuer.requests[0].Result.Lineage.Identity != lane.result.Lineage.Identity ||
				outcome.Receipt.QFinal != 10 || outcome.Receipt.CampaignID != lane.result.Lineage.CampaignID {
				t.Fatalf("events=%q request=%+v receipt=%+v", issuer.events, issuer.requests, outcome.Receipt)
			}
			markets[string(lane.result.Lineage.Market)]++
		})
	}
	// 태스크 4.3: 시장마다 정확히 네 가족이다.
	if markets["KR"] != 4 || markets["US"] != 4 {
		t.Fatalf("unpaired lane matrix: %+v", markets)
	}
}

func TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority(t *testing.T) {
	for _, market := range pairedStrategyFirstLegResults(t) {
		invalid := market.result
		invalid.Lineage.LegOrdinal = 2
		loader := &firstLegAuthorityLoaderStub{}
		issuer := &firstLegGuardianSpy{}
		outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), invalid)
		if outcome.Code != StrategyFirstLegResultInvalid || loader.calls != 0 || issuer.precheckCalls != 0 || issuer.issueCalls != 0 {
			t.Fatalf("outcome=%+v loader=%d issuer=%+v", outcome, loader.calls, issuer)
		}
	}
}

func TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS(t *testing.T) {
	paired := pairedStrategyFirstLegResults(t)
	for index, market := range paired {
		peer := paired[1-index]
		for _, test := range []struct {
			name   string
			mutate func(*execgw.QFinalCampaignFirstLegIssuance)
		}{
			{name: "cross-market", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { *v = firstLegBridgeIssuanceFixture(peer.result) }},
			{name: "candidate", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Entry.QCandidate-- }},
			{name: "generation", mutate: func(v *execgw.QFinalCampaignFirstLegIssuance) { v.Campaign.ExpectedPositionGeneration = 1 }},
		} {
			t.Run(string(market.result.Lineage.Market)+"-"+test.name, func(t *testing.T) {
				issuance := firstLegBridgeIssuanceFixture(market.result)
				test.mutate(&issuance)
				loader := &firstLegAuthorityLoaderStub{issuance: issuance}
				issuer := &firstLegGuardianSpy{}
				outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
				if outcome.Code != StrategyFirstLegAuthorityMismatch || issuer.precheckCalls != 0 || issuer.issueCalls != 0 {
					t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
				}
			})
		}
	}
}

func TestStrategyFirstLegAdmissionClassifiesPairedFailures(t *testing.T) {
	for _, market := range pairedStrategyFirstLegResults(t) {
		t.Run(string(market.result.Lineage.Market)+"-collection", func(t *testing.T) {
			loader := &firstLegAuthorityLoaderStub{err: context.Canceled}
			issuer := &firstLegGuardianSpy{}
			outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
			if outcome.Code != StrategyFirstLegAuthorityCollectionFailed || issuer.precheckCalls != 0 || issuer.issueCalls != 0 {
				t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
			}
		})
		t.Run(string(market.result.Lineage.Market)+"-precheck", func(t *testing.T) {
			loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
			issuer := &firstLegGuardianSpy{precheckErr: errors.New("guardian refused")}
			outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
			if outcome.Code != StrategyFirstLegAuthorityMismatch || issuer.precheckCalls != 1 || issuer.issueCalls != 0 {
				t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
			}
		})
		t.Run(string(market.result.Lineage.Market)+"-atomic", func(t *testing.T) {
			loader := &firstLegAuthorityLoaderStub{issuance: firstLegBridgeIssuanceFixture(market.result)}
			issuer := &firstLegGuardianSpy{issueErr: journal.ErrGenerationConflict}
			outcome := newStrategyFirstLegAdmissionBridge(issuer, loader).admit(context.Background(), market.result)
			if outcome.Code != StrategyFirstLegAtomicAdmissionFailed || issuer.precheckCalls != 1 || issuer.issueCalls != 1 {
				t.Fatalf("outcome=%+v issuer=%+v", outcome, issuer)
			}
		})
	}
}

func TestStrategyFirstLegAdmissionSourceHasOnlyGuardianCapability(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("strategy_first_leg_admission.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{
		"RecordQFinalCampaignFirstLegWithRecollection", "RecordQFinalDecisionAndReserve", "*journal.Journal",
		"ClaimStrategyDispatchLease", "StrategyDispatchLease", "internal/broker", "execgw.Gateway", "NewGateway", ".Submit(",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("bridge contains forbidden capability %q", forbidden)
		}
	}
	for _, required := range []string{"internal/execgw", "PrecheckQFinalCampaignFirstLeg", "IssuePrecheckedQFinalCampaignFirstLeg"} {
		if !strings.Contains(text, required) {
			t.Fatalf("bridge omitted Guardian boundary %q", required)
		}
	}
}

type firstLegAuthorityLoaderStub struct {
	issuance execgw.QFinalCampaignFirstLegIssuance
	err      error
	calls    int
}

func (l *firstLegAuthorityLoaderStub) collectStrategyFirstLegAuthority(context.Context, strategyFirstLegAccepted) (execgw.QFinalCampaignFirstLegIssuance, error) {
	l.calls++
	if l.err != nil {
		return execgw.QFinalCampaignFirstLegIssuance{}, l.err
	}
	return l.issuance, nil
}

type firstLegGuardianSpy struct {
	precheckCalls int
	issueCalls    int
	requests      []execgw.QFinalCampaignFirstLegIssuance
	events        string
	precheckErr   error
	issueErr      error
	receipt       journal.QFinalCampaignFirstLegReceipt
}

func (s *firstLegGuardianSpy) PrecheckQFinalCampaignFirstLeg(request execgw.QFinalCampaignFirstLegIssuance) (execgw.QFinalCampaignFirstLegPrecheck, error) {
	s.precheckCalls++
	s.requests = append(s.requests, request)
	s.events = strings.TrimPrefix(s.events+",precheck", ",")
	return execgw.QFinalCampaignFirstLegPrecheck{}, s.precheckErr
}

func (s *firstLegGuardianSpy) IssuePrecheckedQFinalCampaignFirstLeg(context.Context, execgw.QFinalCampaignFirstLegPrecheck) (journal.QFinalCampaignFirstLegReceipt, error) {
	s.issueCalls++
	s.events = strings.TrimPrefix(s.events+",issue", ",")
	return s.receipt, s.issueErr
}

type pairedStrategyFirstLegCase struct{ result strategyflow.Result }

func pairedStrategyFirstLegResults(t *testing.T) []pairedStrategyFirstLegCase {
	t.Helper()
	all := allPairedStrategyFirstLegResults(t)
	return []pairedStrategyFirstLegCase{all[0], all[1]}
}

func allPairedStrategyFirstLegResults(t *testing.T) []pairedStrategyFirstLegCase {
	t.Helper()
	descriptors := strategyflow.Descriptors()
	out := make([]pairedStrategyFirstLegCase, 0, len(descriptors))
	for _, descriptor := range descriptors {
		market := string(descriptor.Market)
		entryMinor, stopMinor, targetMinor := "50000", "45050", "60010"
		if market == "US" {
			entryMinor, stopMinor, targetMinor = "5000", "4505", "6001"
		}
		result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct", map[string]string{"KR": "005930", "US": "AAPL"}[market],
			"campaign-"+descriptor.LaneID, 20, entryMinor, stopMinor, targetMinor)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, pairedStrategyFirstLegCase{result: result})
	}
	return out
}

func firstLegBridgeIssuanceFixture(result strategyflow.Result) execgw.QFinalCampaignFirstLegIssuance {
	lineage, terms := result.Lineage, result.ExecutionTerms
	market := string(lineage.Market)
	entryMajor, _ := terms.Entry().MajorDecimal()
	stopMajor, _ := terms.EffectiveStop().MajorDecimal()
	targetMajor, _ := terms.Target().MajorDecimal()
	return execgw.QFinalCampaignFirstLegIssuance{
		Entry: execgw.QFinalEntryIssuance{Market: market, Currency: map[string]string{"KR": "KRW", "US": "USD"}[market],
			Symbol: lineage.Symbol, QCandidate: result.Quantity, EntryPrice: entryMajor,
			StopPrice: stopMajor, TargetPrice: targetMajor,
			Admission: journal.RiskBucketAdmissionPlan{TransactionID: "risk-" + lineage.LaneID,
				Admission: riskbucket.AdmissionRequest{QCandidate: result.Quantity},
				Owner: riskbucket.OwnerClaim{Key: riskbucket.OwnerKey{AccountID: lineage.AccountRef, Market: riskbucket.Market(lineage.Market), Symbol: lineage.Symbol},
					LaneID: lineage.LaneID, CampaignID: lineage.CampaignID}}},
		Result: result, ActivationManifestDigest: "sha256:activation-" + lineage.LaneID,
		AttemptID: "attempt-" + lineage.LaneID, Revision: 1,
		Campaign: journal.FirstLegCampaignRequest{CampaignID: lineage.CampaignID, ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
			CreateCommandKey: "create-" + lineage.LaneID, FirstLegCommandKey: "leg-" + lineage.LaneID, FirstLegPlanID: "plan-" + lineage.LaneID},
	}
}

var _ strategyFirstLegGuardian = (*firstLegGuardianSpy)(nil)
var _ strategyFirstLegAuthorityLoader = (*firstLegAuthorityLoaderStub)(nil)
