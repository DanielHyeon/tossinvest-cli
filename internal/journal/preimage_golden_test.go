package journal

import (
	"context"
	"strings"
	"testing"
	"time"
)

// preimage_golden_test.go is change add-net-rr-measurement task 2.3: the evidence
// that this change did not touch the decision contract.
//
// The change adds a table whose whole purpose is to hold the same three prices
// the decision preimage already holds. That duplication is deliberate (design D1)
// and it has exactly one failure mode worth guarding: somebody deciding, later,
// that the tidy thing to do is add the observation's fields to RiskIntent. Doing
// so changes the canonical bytes, which changes risk_hash, which invalidates every
// decision already on disk — the gateway re-derives that hash and compares.
//
// TestPreimageCanonicalFormIsStable (decision_test.go) already pins the text. This
// file pins the *hash* of that text and the value that reaches the column, so the
// contract is nailed at both ends of the derivation rather than only at the start.

// goldenRiskIntentCanonical and goldenRiskIntentHash are the sealed pair. Both
// literals are contracts: a change to either is a change to what every stored
// decision means, and the only correct way to produce one is a new preimage kind.
const (
	goldenRiskIntentCanonical = `{"kind":"RISK_INTENT","account_ref":"acct-1","market":"us","symbol":"AAPL",` +
		`"side":"BUY","quantity":"10","entry_price":"200.5","stop_price":"190",` +
		`"target_price":"","policy_version":"risk/v1"}`
	goldenRiskIntentHash = "27a283409a50efcb127aac4ee9a94c95d84dfa1bbaac5371754c6bb87f8fe64c"
)

// TestRiskIntentHashIsByteStable pins SHA-256 over the canonical text.
func TestRiskIntentHashIsByteStable(t *testing.T) {
	intent := RiskIntent{
		AccountRef: "acct-1", Market: "US", Symbol: "aapl", Side: "buy",
		Quantity: "10.0", EntryPrice: "200.50", StopPrice: "190", TargetPrice: "",
		PolicyVersion: "risk/v1",
	}
	canonical, err := intent.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if canonical != goldenRiskIntentCanonical {
		t.Fatalf("the canonical text drifted:\n got %s\nwant %s", canonical, goldenRiskIntentCanonical)
	}
	if got := HashPreimage(canonical); got != goldenRiskIntentHash {
		t.Fatalf("risk hash drifted:\n got %s\nwant %s", got, goldenRiskIntentHash)
	}

	// The observation table holds these three prices too. Reading them out of the
	// preimage must give back exactly what went in, because a reconstruction is
	// only deterministic if the parse is.
	parsed, err := ParsePreimage(PreimageKindRiskIntent, canonical)
	if err != nil {
		t.Fatal(err)
	}
	round, err := parsed.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if round != goldenRiskIntentCanonical {
		t.Errorf("the parse/canonicalise round trip drifted: %s", round)
	}
}

// TestStoredRiskHashMatchesTheGolden takes the derivation all the way to disk:
// what the issuance writes into decisions.risk_hash is the hash of the text, in a
// v8 journal, with the observation table present and populated.
//
// The observation row is written on purpose. It is the concrete form of "the two
// tables are joined by nothing the database enforces": an observation carrying the
// same prices exists beside the decision and changes nothing about it.
func TestStoredRiskHashMatchesTheGolden(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	now := j.Now()

	if _, err := j.RecordDecisionAndReserve(ctx,
		issueRequest(j, "decision-1", "acct-1", "100", "0", "1000000",
			mustVersion(t, j, "acct-1"))); err != nil {
		t.Fatal(err)
	}
	obs := observationAt("obs-1", now)
	obs.Outcome = OutcomeAllowedIssued
	obs.StoppedStep, obs.ReasonCode = "", ""
	obs.DecisionID = "decision-1"
	obs.IssuedAt = now
	if err := j.RecordEntryObservation(ctx, obs); err != nil {
		t.Fatal(err)
	}

	dec, err := j.LookupDecision(ctx, "decision-1")
	if err != nil {
		t.Fatal(err)
	}
	// issueRequest's own preimage, spelled out. Written as a literal rather than
	// recomputed from the fixture so that a change to either one is visible here.
	const wantCanonical = `{"kind":"RISK_INTENT","account_ref":"acct-1","market":"kr","symbol":"005930",` +
		`"side":"BUY","quantity":"10","entry_price":"70000","stop_price":"68000",` +
		`"target_price":"","policy_version":"test-1"}`
	const wantHash = "808c6fd547c9613e1fc85793483c4f3d293e3e7c26ccf92a97e32af0d7fe1bc1"
	if dec.RiskPreimage != wantCanonical {
		t.Errorf("stored preimage:\n got %s\nwant %s", dec.RiskPreimage, wantCanonical)
	}
	if dec.RiskHash != wantHash {
		t.Errorf("stored risk_hash:\n got %s\nwant %s", dec.RiskHash, wantHash)
	}
	if dec.RiskHash != HashPreimage(dec.RiskPreimage) {
		t.Error("the stored hash must be the hash of the stored text")
	}

	// And it is still the same after the observation table has been written to,
	// swept, and written to again — the sweep is on the other table only.
	if _, err := j.PruneEntryObservations(ctx, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	after, err := j.LookupDecision(ctx, "decision-1")
	if err != nil {
		t.Fatal(err)
	}
	if after.RiskPreimage != dec.RiskPreimage || after.RiskHash != dec.RiskHash {
		t.Error("the observation sweep must not reach the decision contract")
	}
}

// TestRiskIntentFieldSetIsClosed is the structural half of the same guard. The
// canonical form is a fixed field list in a fixed order, so an added field would
// have to appear here — and ParsePreimage refuses unknown fields, which means a
// journal written by a build that added one is unreadable by a build that did not.
//
// Listing the names is what makes the failure a conversation rather than a
// surprise: a reviewer seeing this test change knows the decision contract is
// being altered.
func TestRiskIntentFieldSetIsClosed(t *testing.T) {
	withEverything := RiskIntent{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "BUY",
		Quantity: "1", EntryPrice: "100", StopPrice: "99", TargetPrice: "102",
		PolicyVersion: "v1",
	}
	canonical, err := withEverything.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"kind":"RISK_INTENT","account_ref":"acct-1","market":"kr","symbol":"005930",` +
		`"side":"BUY","quantity":"1","entry_price":"100","stop_price":"99",` +
		`"target_price":"102","policy_version":"v1"}`
	if canonical != want {
		t.Fatalf("the risk intent field set or its order changed:\n got %s\nwant %s", canonical, want)
	}

	// Nothing from the observation vocabulary leaked in. These are the fields
	// this change deliberately kept out of the preimage (design D1): the ratios,
	// the break-even, the fingerprint and the scope live in the other table.
	for _, forbidden := range []string{
		"net_reward_risk", "gross_reward_risk", "break_even",
		"cost_model_fingerprint", "cost_scope", "outcome", "stopped_step",
	} {
		if strings.Contains(canonical, forbidden) {
			t.Errorf("%q reached the hashed preimage; observations belong in "+
				"entry_decision_observations", forbidden)
		}
	}
}
