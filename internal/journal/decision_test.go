package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Decision persistence tests (extend-execution-contract task 1.4, design D1).
//
// The property under test throughout is the trust boundary: what a decision
// authorises is what the *row* says, and the row is written before anything is
// submitted. Everything else here — canonical form, hashing, class/preimage
// pairing — exists to make that row unambiguous.

func testIssued(t *testing.T) time.Time {
	t.Helper()
	issued, err := time.Parse(time.RFC3339, testNow)
	if err != nil {
		t.Fatal(err)
	}
	return issued
}

func reductionRequest(t *testing.T) DecisionRequest {
	t.Helper()
	issued := testIssued(t)
	return DecisionRequest{
		ID:          "dec-red",
		AccountRef:  "acct-1",
		SafetyClass: SafetyClassRiskReducing,
		Kind:        KindCancel,
		Preimage: ReductionIntent{
			AccountRef: "acct-1", Market: "us", Symbol: "AAPL", Side: "SELL",
			MaxQuantity: "10", Reason: "flatten",
		},
		Nonce:     "nonce-red",
		IssuedAt:  issued,
		ExpiresAt: issued.Add(time.Minute),
	}
}

func riskRequest(t *testing.T) DecisionRequest {
	t.Helper()
	issued := testIssued(t)
	return DecisionRequest{
		ID:          "dec-risk",
		AccountRef:  "acct-1",
		SafetyClass: SafetyClassExposureRaising,
		Kind:        KindPlace,
		Preimage: RiskIntent{
			AccountRef: "acct-1", Market: "us", Symbol: "AAPL", Side: "BUY",
			Quantity: "10", EntryPrice: "200.5", StopPrice: "190", TargetPrice: "230",
			PolicyVersion: "risk/v1",
		},
		LimitsJSON: `{"max_quantity":"100"}`,
		Nonce:      "nonce-risk",
		IssuedAt:   issued,
		ExpiresAt:  issued.Add(time.Minute),
	}
}

// TestRecordAndLookupDecision is the round trip the gateway depends on: what the
// issuer wrote is what the verifier reads, including the derived key and hash.
func TestRecordAndLookupDecision(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	stored, err := j.RecordDecision(ctx, riskRequest(t))
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}
	if stored.ClientOrderID != DeriveClientOrderID("dec-risk", 0) {
		t.Errorf("client_order_id = %q, want the derived key", stored.ClientOrderID)
	}
	if stored.RiskHash != HashPreimage(stored.RiskPreimage) {
		t.Errorf("risk_hash is not the hash of the stored preimage")
	}

	read, err := j.LookupDecision(ctx, "dec-risk")
	if err != nil {
		t.Fatalf("LookupDecision: %v", err)
	}
	if read.RiskPreimage != stored.RiskPreimage || read.RiskHash != stored.RiskHash ||
		read.ClientOrderID != stored.ClientOrderID || read.Nonce != stored.Nonce ||
		read.SafetyClass != stored.SafetyClass || read.PreimageKind != stored.PreimageKind ||
		read.LimitsJSON != stored.LimitsJSON || read.AccountRef != stored.AccountRef ||
		read.Generation != stored.Generation {
		t.Fatalf("round trip mismatch:\nstored %+v\n  read %+v", stored, read)
	}
	if !read.IssuedAt.Equal(stored.IssuedAt.UTC().Truncate(time.Second)) {
		t.Errorf("issued_at = %s, want %s", read.IssuedAt, stored.IssuedAt)
	}
	if !read.ExpiresAt.Equal(stored.ExpiresAt.UTC().Truncate(time.Second)) {
		t.Errorf("expires_at = %s, want %s", read.ExpiresAt, stored.ExpiresAt)
	}
}

// TestReductionDecisionCarriesNoKeyOrLimits: a cancel or a reduce-only exit is
// not limit-checked (§0.3) and has no idempotency key to carry — the broker's
// cancel and modify endpoints take none (openapi).
func TestReductionDecisionCarriesNoKeyOrLimits(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	stored, err := j.RecordDecision(ctx, reductionRequest(t))
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}
	if stored.ClientOrderID != "" {
		t.Errorf("a cancel decision carries no key, got %q", stored.ClientOrderID)
	}
	if stored.LimitsJSON != "" {
		t.Errorf("a RISK_REDUCING decision carries no limits, got %q", stored.LimitsJSON)
	}
	if stored.PreimageKind != PreimageKindReductionIntent {
		t.Errorf("preimage_kind = %q", stored.PreimageKind)
	}
}

// TestReducingPlaceDecisionCarriesAKey: a reduce-only *sell* is still an order
// creation, so it does get a key — the class does not decide that, the endpoint
// does.
func TestReducingPlaceDecisionCarriesAKey(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	req := reductionRequest(t)
	req.Kind = KindPlace
	stored, err := j.RecordDecision(ctx, req)
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}
	if stored.ClientOrderID != DeriveClientOrderID(req.ID, 0) {
		t.Fatalf("client_order_id = %q, want the derived key", stored.ClientOrderID)
	}
}

// TestRecordDecisionRefusals is the table of shapes that are not decisions.
func TestRecordDecisionRefusals(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	cases := []struct {
		name   string
		mutate func(*DecisionRequest)
		want   string
	}{
		{"no id", func(r *DecisionRequest) { r.ID = "" }, "decision id"},
		{"no account", func(r *DecisionRequest) { r.AccountRef = "" }, "account ref"},
		{"no nonce", func(r *DecisionRequest) { r.Nonce = "" }, "nonce"},
		{"advanced generation", func(r *DecisionRequest) { r.Generation = 1 }, "generation 0 only"},
		{"unknown class", func(r *DecisionRequest) { r.SafetyClass = "MOSTLY_SAFE" }, "safety class"},
		{"reserved class", func(r *DecisionRequest) {
			r.SafetyClass = SafetyClassProtectionWeakening
		}, "reserved"},
		{"unknown kind", func(r *DecisionRequest) { r.Kind = MutationKind("FROB") }, "mutation kind"},
		{"no preimage", func(r *DecisionRequest) { r.Preimage = nil }, "risk preimage"},
		{"class and preimage disagree", func(r *DecisionRequest) {
			r.Preimage = ReductionIntent{
				AccountRef: "acct-1", Market: "us", Symbol: "AAPL", Side: "SELL",
				MaxQuantity: "1", Reason: "x",
			}
		}, "preimage"},
		{"preimage for another account", func(r *DecisionRequest) {
			p := r.Preimage.(RiskIntent)
			p.AccountRef = "acct-2"
			r.Preimage = p
		}, "account"},
		{"entry without a stop", func(r *DecisionRequest) {
			p := r.Preimage.(RiskIntent)
			p.StopPrice = ""
			r.Preimage = p
		}, "stop price"},
		{"entry without a policy version", func(r *DecisionRequest) {
			p := r.Preimage.(RiskIntent)
			p.PolicyVersion = ""
			r.Preimage = p
		}, "policy version"},
		{"non-decimal quantity", func(r *DecisionRequest) {
			p := r.Preimage.(RiskIntent)
			p.Quantity = "ten"
			r.Preimage = p
		}, "decimal"},
		{"entry without limits", func(r *DecisionRequest) { r.LimitsJSON = "" }, "limit snapshot"},
		{"no expiry", func(r *DecisionRequest) { r.ExpiresAt = time.Time{} }, "expires_at"},
		{"expiry before issue", func(r *DecisionRequest) {
			r.ExpiresAt = r.IssuedAt.Add(-time.Second)
		}, "expires at or before"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := riskRequest(t)
			tc.mutate(&req)
			_, err := j.RecordDecision(ctx, req)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("want ErrInvalidRequest, got %v", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("the refusal must say why (%q missing): %v", tc.want, err)
			}
		})
	}

	var count int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM decisions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("refused decisions wrote %d rows, want 0", count)
	}
}

// TestReductionDecisionRefusesLimits keeps the exit path clear of anything that
// could refuse it later (§0.3).
func TestReductionDecisionRefusesLimits(t *testing.T) {
	j := openTestJournal(t)
	req := reductionRequest(t)
	req.LimitsJSON = `{"max_quantity":"1"}`
	if _, err := j.RecordDecision(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("want ErrInvalidRequest, got %v", err)
	}
}

// TestNonceIsUniqueAcrossDecisions: the one-shot property starts at issue time.
// Two decisions sharing a nonce would let the second one ride the first's
// consumption record.
func TestNonceIsUniqueAcrossDecisions(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.RecordDecision(ctx, riskRequest(t)); err != nil {
		t.Fatal(err)
	}
	second := riskRequest(t)
	second.ID = "dec-risk-2"
	if _, err := j.RecordDecision(ctx, second); err == nil {
		t.Fatal("a reused nonce must be refused by the store, not just by convention")
	}
}

// TestLookupMissingDecision documents what the gateway sees when a submission
// names a decision nobody wrote.
func TestLookupMissingDecision(t *testing.T) {
	j := openTestJournal(t)
	if _, err := j.LookupDecision(context.Background(), "nope"); !errors.Is(err, ErrDecisionNotFound) {
		t.Fatalf("want ErrDecisionNotFound, got %v", err)
	}
}

// TestPreimageCanonicalFormIsStable pins the exact bytes the hash is taken over.
// Changing them invalidates every decision already on disk, so this literal is a
// contract.
func TestPreimageCanonicalFormIsStable(t *testing.T) {
	risk := RiskIntent{
		AccountRef: "acct-1", Market: "US", Symbol: "aapl", Side: "buy",
		Quantity: "10.0", EntryPrice: "200.50", StopPrice: "190", TargetPrice: "",
		PolicyVersion: "risk/v1",
	}
	got, err := risk.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"kind":"RISK_INTENT","account_ref":"acct-1","market":"us","symbol":"AAPL",` +
		`"side":"BUY","quantity":"10","entry_price":"200.5","stop_price":"190",` +
		`"target_price":"","policy_version":"risk/v1"}`
	if got != want {
		t.Fatalf("canonical risk intent drifted:\n got %s\nwant %s", got, want)
	}

	reduction := ReductionIntent{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Side: "sell",
		MaxQuantity: "3.500", Reason: "flatten-all",
	}
	got, err = reduction.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want = `{"kind":"REDUCTION_INTENT","account_ref":"acct-1","market":"kr","symbol":"005930",` +
		`"side":"SELL","max_quantity":"3.5","reason":"flatten-all"}`
	if got != want {
		t.Fatalf("canonical reduction intent drifted:\n got %s\nwant %s", got, want)
	}
}

// TestParsePreimageRoundTrip: what was stored parses back to a value that
// re-canonicalises to the same bytes, which is what lets the gateway compare
// typed fields against the order while the hash covers the text.
func TestParsePreimageRoundTrip(t *testing.T) {
	original := RiskIntent{
		AccountRef: "acct-1", Market: "us", Symbol: "AAPL", Side: "BUY",
		Quantity: "10", EntryPrice: "200.5", StopPrice: "190", TargetPrice: "230",
		PolicyVersion: "risk/v1",
	}
	canonical, err := original.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParsePreimage(PreimageKindRiskIntent, canonical)
	if err != nil {
		t.Fatalf("ParsePreimage: %v", err)
	}
	if parsed != Preimage(original) {
		t.Fatalf("parsed = %+v, want %+v", parsed, original)
	}
}

// TestParsePreimageRefusesNonCanonicalText is the anti-smuggling guard: a blob
// that decodes but is not the canonical spelling is refused rather than
// re-canonicalised, because the hash was taken over the text.
func TestParsePreimageRefusesNonCanonicalText(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"extra field", `{"kind":"REDUCTION_INTENT","account_ref":"a","market":"us","symbol":"X",` +
			`"side":"SELL","max_quantity":"1","reason":"r","extra":"1"}`},
		{"reordered", `{"account_ref":"a","kind":"REDUCTION_INTENT","market":"us","symbol":"X",` +
			`"side":"SELL","max_quantity":"1","reason":"r"}`},
		{"whitespace", `{"kind":"REDUCTION_INTENT", "account_ref":"a","market":"us","symbol":"X",` +
			`"side":"SELL","max_quantity":"1","reason":"r"}`},
		{"unnormalised decimal", `{"kind":"REDUCTION_INTENT","account_ref":"a","market":"us","symbol":"X",` +
			`"side":"SELL","max_quantity":"1.00","reason":"r"}`},
		{"lowercase symbol", `{"kind":"REDUCTION_INTENT","account_ref":"a","market":"us","symbol":"x",` +
			`"side":"SELL","max_quantity":"1","reason":"r"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParsePreimage(PreimageKindReductionIntent, tc.text); err == nil {
				t.Fatal("non-canonical preimage text must be refused")
			}
		})
	}
}

// TestNormalizeDecimal covers the exact, float-free normalisation the canonical
// form depends on.
func TestNormalizeDecimal(t *testing.T) {
	ok := map[string]string{
		"10":                      "10",
		"10.0":                    "10",
		"010.500":                 "10.5",
		"0.10000000000000000555":  "0.10000000000000000555",
		".5":                      "0.5",
		"0":                       "0",
		"0.000":                   "0",
		"+7":                      "7",
		" 7 ":                     "7",
		"-0.50":                   "-0.5",
		"123456789012345678901.5": "123456789012345678901.5",
	}
	for in, want := range ok {
		got, valid := NormalizeDecimal(in)
		if !valid {
			t.Errorf("NormalizeDecimal(%q) reported invalid", in)
			continue
		}
		if got != want {
			t.Errorf("NormalizeDecimal(%q) = %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"", " ", "ten", "1e5", "1.2.3", "1,000", "0x10", "."} {
		if got, valid := NormalizeDecimal(in); valid {
			t.Errorf("NormalizeDecimal(%q) = %q, want invalid", in, got)
		}
	}
}

// TestAttemptForClientOrderID is how a second submission of the same decision is
// recognised across processes: the key is claimed durably at RECORDED.
func TestAttemptForClientOrderID(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	dec, err := j.RecordDecision(ctx, riskRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.AttemptForClientOrderID(ctx, "acct-1", dec.ClientOrderID); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("before the attempt exists: want ErrAttemptNotFound, got %v", err)
	}

	req := testRequest()
	req.DecisionID = dec.ID
	req.SafetyClass = dec.SafetyClass
	req.ClientOrderID = dec.ClientOrderID
	if _, err := j.Prepare(ctx, req); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	rec, err := j.AttemptForClientOrderID(ctx, "acct-1", dec.ClientOrderID)
	if err != nil {
		t.Fatalf("AttemptForClientOrderID: %v", err)
	}
	if rec.ID != "attempt-1" {
		t.Fatalf("attempt = %q, want attempt-1", rec.ID)
	}
	// Scoped by account: the same key on another account is a different claim.
	if _, err := j.AttemptForClientOrderID(ctx, "acct-2", dec.ClientOrderID); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("cross-account lookup: want ErrAttemptNotFound, got %v", err)
	}
}

// TestPrepareRefusesABindingTheDecisionDoesNotSay: the attempt row cannot claim
// a class, generation or key its decision never carried.
func TestPrepareRefusesABindingTheDecisionDoesNotSay(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	dec, err := j.RecordDecision(ctx, riskRequest(t))
	if err != nil {
		t.Fatal(err)
	}

	base := func() PrepareRequest {
		req := testRequest()
		req.DecisionID = dec.ID
		req.SafetyClass = dec.SafetyClass
		req.ClientOrderID = dec.ClientOrderID
		return req
	}
	cases := []struct {
		name   string
		mutate func(*PrepareRequest)
	}{
		{"class the decision does not carry", func(r *PrepareRequest) {
			r.SafetyClass = SafetyClassRiskReducing
		}},
		{"generation the decision does not carry", func(r *PrepareRequest) { r.Generation = 2 }},
		{"key the decision does not authorise", func(r *PrepareRequest) {
			r.ClientOrderID = DeriveClientOrderID("some-other-decision", 0)
		}},
		{"no key although the decision has one", func(r *PrepareRequest) { r.ClientOrderID = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := base()
			tc.mutate(&req)
			if _, err := j.Prepare(ctx, req); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("want ErrInvalidRequest, got %v", err)
			}
		})
	}
}
