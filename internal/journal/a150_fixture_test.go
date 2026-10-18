package journal

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

// a150TestProjection rebinds a real test-seam-produced projection, entirely in
// the test binary. Verification still runs the production canonical reader.
func a150TestProjection(t *testing.T, market, account, symbol, campaign string) strategyflow.AcceptedProjection {
	return a150NamedTestProjection(t, "a150-projection-"+market, account, symbol, campaign)
}
func a150NamedTestProjection(t *testing.T, name, account, symbol, campaign string) strategyflow.AcceptedProjection {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	original, err := strategyflow.VerifyAcceptedProjection(text)
	if err != nil {
		t.Fatalf("fixture corrupt: %v", err)
	}
	replace := func(old, new string) {
		a, _ := json.Marshal(old)
		b, _ := json.Marshal(new)
		text = strings.ReplaceAll(text, string(a), string(b))
	}
	l := original.Lineage()
	terms := original.ExecutionTerms()
	replace(l.AccountRef, account)
	replace(l.Symbol, symbol)
	replace(l.CampaignID, campaign)
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var record map[string]any
	if err := decoder.Decode(&record); err != nil {
		t.Fatal(err)
	}
	lineage := record["lineage"].(map[string]any)
	h := sha256.New()
	str := func(v any) { s := v.(string); fixtureWriteString(h, s) }
	num := func(v any) {
		n, err := strconv.ParseUint(v.(json.Number).String(), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		fixtureWriteUint(h, n)
	}
	for _, field := range []string{"account_ref", "market", "symbol"} {
		str(lineage[field])
	}
	num(lineage["position_generation"])
	for _, field := range []string{"candidate_state", "candidate_life_id"} {
		str(lineage[field])
	}
	for _, field := range []string{"candidate_first_seen_ns", "candidate_last_seen_ns", "candidate_valid_until_ns", "candidate_approved_at_ns"} {
		num(lineage[field])
	}
	for _, field := range []string{"threshold_version", "threshold_set_digest", "candidate_evidence_digest", "router_evidence_digest", "lane_evidence_digest", "router_id", "router_release", "horizon", "lane_id", "lane_version", "lane_release", "config_digest", "campaign_id"} {
		str(lineage[field])
	}
	num(lineage["leg_ordinal"])
	num(lineage["planned_ceiling"])
	str(lineage["risk_budget_digest"])
	str(lineage["execution_policy_digest"])
	fixtureWriteString(h, "1")
	lineageID := "strategy-lineage:v1:sha256:" + hex.EncodeToString(h.Sum(nil))
	replace(l.Identity, lineageID)
	decoder = json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	record = map[string]any{}
	if err := decoder.Decode(&record); err != nil {
		t.Fatal(err)
	}
	bound := record["execution_terms"].(map[string]any)
	h = sha256.New()
	for _, field := range []string{"account_ref", "market", "symbol", "campaign_id"} {
		str(bound[field])
	}
	num(bound["leg_ordinal"])
	num(bound["quantity"])
	for _, part := range []string{"entry", "effective_stop", "target"} {
		price := bound[part].(map[string]any)
		for _, field := range []string{"price_minor", "source", "version", "digest", "as_of", "currency"} {
			str(price[field])
		}
		num(price["minor_scale"])
		str(price["unit_version"])
	}
	str(bound["policy"].(map[string]any)["identity"])
	str(bound["lineage_identity"])
	termID := "strategy-execution-terms:v1:sha256:" + hex.EncodeToString(h.Sum(nil))
	replace(terms.Identity(), termID)
	projection, err := strategyflow.VerifyAcceptedProjection(text)
	if err != nil {
		t.Fatalf("rebound fixture: %v", err)
	}
	return projection
}

func fixtureWriteUint(h hash.Hash, v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	_, _ = h.Write(b[:])
}
func fixtureWriteString(h hash.Hash, s string) {
	fixtureWriteUint(h, uint64(len(s)))
	_, _ = h.Write([]byte(s))
}

func a150FirstLegAtomicFixture(t *testing.T, j *Journal, suffix, account, market, symbol string) QFinalCampaignFirstLegRequest {
	t.Helper()
	request := legacyFirstLegAtomicFixture(t, j, suffix, account, market, symbol)
	projection := a150TestProjection(t, market, account, symbol, request.Campaign.CampaignID)
	return a150BindFixture(t, request, projection)
}
func a150BindFixture(t *testing.T, request QFinalCampaignFirstLegRequest, projection strategyflow.AcceptedProjection) QFinalCampaignFirstLegRequest {
	t.Helper()
	request.Issue.Admission.Owner.LaneID = projection.Lineage().LaneID
	risk, err := exactCanonicalRiskIntent(request.Issue.Issue.Decision.Preimage.(RiskIntent))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := NewStrategyExitPolicySnapshot("")
	if err != nil {
		t.Fatal(err)
	}
	created := request.Strategy.CreatedAt.UTC()
	record := strategyflowRiskBindingPayload{SchemaVersion: strategyflowRiskBindingSchemaVersionV4, Strategyflow: json.RawMessage(projection.Payload()), StrategyflowPayloadDigest: projection.PayloadDigest(), RiskIntent: risk, ActivationManifestDigest: request.Strategy.ActivationManifestDigest, CreatedAt: created.Format("2006-01-02T15:04:05.999999999Z07:00"), ExitPolicy: &policy}
	record.Identity, err = strategyflowDecisionIdentity(record)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	request.Strategy.Lineage = strategyflowLineageFromProjection(projection, risk, record.ActivationManifestDigest, created, record.Identity, string(payload), record.SchemaVersion)
	if err := verifyStrategyRiskBindingMustBuild(t, request); err != nil {
		t.Fatal(err)
	}
	return request
}
func verifyStrategyRiskBindingMustBuild(t *testing.T, request QFinalCampaignFirstLegRequest) error {
	t.Helper()
	decision, err := request.Issue.Issue.Decision.build()
	if err != nil {
		return err
	}
	if err := verifyStrategyRiskBinding(decision, request.Strategy.Lineage); err != nil {
		return fmt.Errorf("test fixture: %w", err)
	}
	return nil
}
