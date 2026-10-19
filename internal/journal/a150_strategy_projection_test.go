//go:build tossos_testseams

package journal

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func TestA150ProjectionBindsExitPolicyAndRetainsV3(t *testing.T) {
	descriptor := strategyflow.Descriptors()[0]
	market := string(descriptor.Market)
	symbol := map[string]string{"KR": "005930", "US": "AAPL"}[market]
	entryMinor, stopMinor, targetMinor, entry, stop, target := pairedProjectionPrices(market)
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, "acct-a150", symbol, "campaign-a150", 20, entryMinor, stopMinor, targetMinor)
	if err != nil {
		t.Fatal(err)
	}
	version, err := QFinalPolicyVersion("guardian-v1", "risk-a150")
	if err != nil {
		t.Fatal(err)
	}
	risk := RiskIntent{AccountRef: "acct-a150", Market: market, Symbol: symbol, Side: "BUY", Quantity: "10", EntryPrice: entry, StopPrice: stop, TargetPrice: target, PolicyVersion: version}
	request := StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk, ActivationManifestDigest: "activation-a150", CreatedAt: testIssued(t)}
	legacy, err := ProjectAcceptedStrategyflowLineage(request)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/a150-baseline-v3.json")
	if err != nil || legacy.DecisionPayload != string(golden) {
		t.Fatalf("pre-a150 v3 bytes changed: %v", err)
	}
	if !strings.Contains(legacy.DecisionPayload, `"schema_version":"journal-strategyflow-risk-binding:v3"`) {
		t.Fatalf("legacy %s", legacy.DecisionPayload)
	}
	snapshot, err := NewStrategyExitPolicySnapshot("")
	if err != nil {
		t.Fatal(err)
	}
	request.ExitPolicy = &snapshot
	current, err := ProjectAcceptedStrategyflowLineage(request)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(current.DecisionPayload, `"schema_version":"journal-strategyflow-risk-binding:v4"`) || current.DecisionIdentity == legacy.DecisionIdentity {
		t.Fatalf("v4 not bound: %+v", current)
	}
	same, err := ProjectAcceptedStrategyflowLineage(request)
	if err != nil || same != current {
		t.Fatalf("replay: %v", err)
	}
	decision, err := (DecisionRequest{ID: "a150-decision", AccountRef: risk.AccountRef,
		SafetyClass: SafetyClassExposureRaising, Kind: KindPlace, Preimage: risk,
		LimitsJSON: `{"max_quantity":"10"}`, Nonce: "a150-nonce", IssuedAt: testIssued(t), ExpiresAt: testIssued(t).Add(time.Minute)}).build()
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyStrategyRiskBinding(decision, current); err != nil {
		t.Fatalf("v4 binding: %v", err)
	}
	changed := snapshot
	changed.Identity.Digest = "sha256:" + strings.Repeat("0", 64)
	request.ExitPolicy = &changed
	if _, err = ProjectAcceptedStrategyflowLineage(request); err == nil {
		t.Fatal("unrecognized policy digest accepted")
	}
}

func a150V4FirstLegFixture(t *testing.T, j *Journal) QFinalCampaignFirstLegRequest {
	t.Helper()
	descriptor := strategyflow.Descriptors()[0]
	request := strategyflowFirstLegFixture(t, j, descriptor, "a150-v4")
	risk := request.Issue.Issue.Decision.Preimage.(RiskIntent)
	entry, stop, target, _, _, _ := pairedProjectionPrices(risk.Market)
	result, err := strategyflow.AcceptedResultForJournalTest(descriptor, risk.AccountRef, risk.Symbol, request.Campaign.CampaignID, 20, entry, stop, target)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := NewStrategyExitPolicySnapshot("")
	if err != nil {
		t.Fatal(err)
	}
	request.Strategy.Lineage, err = ProjectAcceptedStrategyflowLineage(StrategyflowLineageProjectionRequest{Result: result, RiskIntent: risk, ActivationManifestDigest: request.Strategy.ActivationManifestDigest, CreatedAt: request.Strategy.CreatedAt, ExitPolicy: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestA150FirstLegAdmissionWritesMarkerAndReplays(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	request := a150V4FirstLegFixture(t, j)
	first, err := j.RecordQFinalCampaignFirstLeg(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	var marker int
	if err := j.db.QueryRowContext(ctx, `SELECT strategy_binding_version FROM decisions WHERE id=?`, request.Issue.Issue.Decision.ID).Scan(&marker); err != nil || marker != 4 {
		t.Fatalf("marker=%d err=%v", marker, err)
	}
	if _, err := j.db.ExecContext(ctx, `UPDATE decisions SET strategy_binding_version=0 WHERE id=?`, request.Issue.Issue.Decision.ID); err == nil {
		t.Fatal("v4 marker downgrade succeeded")
	}
	replay, err := j.RecordQFinalCampaignFirstLeg(ctx, request)
	if err != nil || !replay.Idempotent || first.DecisionID != replay.DecisionID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
}

func TestA150V4ReservationFailureRollsBackDecisionAndMarker(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	request := a150V4FirstLegFixture(t, j)
	if _, err := j.db.ExecContext(ctx, `CREATE TRIGGER a150_injected_reservation_failure BEFORE INSERT ON risk_reservations BEGIN SELECT RAISE(ABORT, 'injected reservation failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordQFinalCampaignFirstLeg(ctx, request); err == nil || !strings.Contains(err.Error(), "injected reservation failure") {
		t.Fatalf("injected failure: %v", err)
	}
	for _, table := range []string{"decisions", "risk_reservations", "strategy_decision_lineage", "position_campaigns", "risk_bucket_final_decisions", "strategy_first_leg_bindings"} {
		if got := countRiskBucketRows(t, j, table); got != 0 {
			t.Errorf("%s=%d; transaction did not roll back", table, got)
		}
	}
}
