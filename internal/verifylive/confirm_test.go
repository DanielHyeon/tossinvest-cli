package verifylive

// confirm_test.go covers the gate in front of every live mutation.
//
// It is internal/flatten/confirm_test's set of properties, because it is the same
// gate: no terminal is a refusal, an expired nonce is a refusal, a near-miss is a
// refusal, and there is no shape of input that stands in for the typed string.

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func testMutation(now time.Time) Mutation {
	return NewMutation(StepIdempotency, "place a live LIMIT order", "123-45-678901",
		[]string{"symbol           005930 (KR)", "quantity         1 share(s)"},
		"cancelled at the end of this step", now)
}

func TestConfirmRefusesWithoutATerminal(t *testing.T) {
	now := time.Now()
	m := testMutation(now)
	var out bytes.Buffer
	err := Confirm(strings.NewReader(m.Nonce+"\n"), &out, m, false, now)
	if !errors.Is(err, ErrNotATerminal) {
		t.Fatalf("err = %v, want ErrNotATerminal — piping the right answer in must not work", err)
	}
	if out.Len() != 0 {
		t.Error("the prompt was printed to a non-terminal; nothing should have been offered")
	}
}

func TestConfirmAcceptsTheTypedNonce(t *testing.T) {
	now := time.Now()
	m := testMutation(now)
	var out bytes.Buffer
	if err := Confirm(strings.NewReader(m.Nonce+"\n"), &out, m, true, now); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	prompt := out.String()
	for _, want := range []string{"LIVE MUTATION", "place a live LIMIT order", "005930", m.Nonce, "expires"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the prompt does not contain %q:\n%s", want, prompt)
		}
	}
}

func TestConfirmMasksTheAccount(t *testing.T) {
	m := testMutation(time.Now())
	if strings.Contains(m.Prompt(), "123-45-678901") {
		t.Error("the prompt printed the full account number")
	}
	if !strings.Contains(m.Prompt(), "8901") {
		t.Errorf("the prompt shows no account reference at all:\n%s", m.Prompt())
	}
}

func TestConfirmRejectsAnythingButTheNonce(t *testing.T) {
	now := time.Now()
	m := testMutation(now)
	for _, answer := range []string{"y", "yes", "Y\n", "", "\n", strings.ToLower(m.Nonce), m.Nonce + "x"} {
		var out bytes.Buffer
		err := Confirm(strings.NewReader(answer), &out, m, true, now)
		if !errors.Is(err, ErrRefused) {
			t.Errorf("answer %q produced %v, want ErrRefused", answer, err)
		}
	}
}

func TestConfirmExpires(t *testing.T) {
	now := time.Now()
	m := testMutation(now)
	var out bytes.Buffer
	err := Confirm(strings.NewReader(m.Nonce+"\n"), &out, m, true, now.Add(ConfirmationTTL+time.Second))
	if !errors.Is(err, ErrConfirmationExpired) {
		t.Fatalf("err = %v, want ErrConfirmationExpired", err)
	}
}

// TestNoncesDoNotRepeat. A reused nonce is a nonce somebody could have typed from
// a runbook.
func TestNoncesDoNotRepeat(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		m := testMutation(time.Now())
		if seen[m.Nonce] {
			t.Fatalf("nonce %q repeated", m.Nonce)
		}
		seen[m.Nonce] = true
		if !strings.HasPrefix(m.Nonce, "VERIFY-") {
			t.Fatalf("nonce %q does not identify what it confirms", m.Nonce)
		}
	}
}

// TestPromptStatesTheReversal. An operator has to be told how the exposure ends
// before they type anything, not afterwards.
func TestPromptStatesTheReversal(t *testing.T) {
	m := testMutation(time.Now())
	if !strings.Contains(m.Prompt(), "reversal") {
		t.Errorf("the prompt does not say how the exposure ends:\n%s", m.Prompt())
	}
}

// --- the batch approval ----------------------------------------------------------
//
// The same four properties, for the gate that is now the default.

func testBatch(now time.Time) Batch {
	return NewBatch(Plan{
		RunID:   "run-TEST",
		Account: "********8901",
		Mutations: []PlannedMutation{{
			Ordinal: 1, Step: StepOrderCancel, Kind: MutatePlaceOrder,
			Symbol: "005930", Side: "buy", Quantity: "1 share", MaxQuantity: 1,
			Pricing: PriceFarBuy.Describe(DefaultOffset, MarketKR),
			Ends:    "cancelled inside this step",
		}},
	}, false, now)
}

func TestConfirmBatchRefusesWithoutATerminal(t *testing.T) {
	now := time.Now()
	b := testBatch(now)
	var out bytes.Buffer
	err := ConfirmBatch(strings.NewReader(b.Nonce+"\n"), &out, b, false, now)
	if !errors.Is(err, ErrNotATerminal) {
		t.Fatalf("err = %v, want ErrNotATerminal — piping the right answer in must not work", err)
	}
	if out.Len() != 0 {
		t.Error("the plan was printed to a non-terminal; nothing should have been offered")
	}
}

func TestConfirmBatchAcceptsTheTypedNonce(t *testing.T) {
	now := time.Now()
	b := testBatch(now)
	var out bytes.Buffer
	if err := ConfirmBatch(strings.NewReader(b.Nonce+"\n"), &out, b, true, now); err != nil {
		t.Fatalf("ConfirmBatch: %v", err)
	}
	prompt := out.String()
	for _, want := range []string{
		"LIVE MUTATION BATCH", "1 request(s)", "order-cancel", "005930", "BUY", "cancelled inside this step",
		b.Nonce, "expires", "abort",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the batch prompt does not contain %q:\n%s", want, prompt)
		}
	}
}

func TestConfirmBatchRejectsAnythingButTheNonce(t *testing.T) {
	now := time.Now()
	b := testBatch(now)
	for _, answer := range []string{"y", "approve", "ok", "", "\n", strings.ToLower(b.Nonce), b.Nonce + "x"} {
		var out bytes.Buffer
		err := ConfirmBatch(strings.NewReader(answer), &out, b, true, now)
		if !errors.Is(err, ErrRefused) {
			t.Errorf("answer %q produced %v, want ErrRefused", answer, err)
		}
	}
}

func TestConfirmBatchExpires(t *testing.T) {
	now := time.Now()
	b := testBatch(now)
	var out bytes.Buffer
	err := ConfirmBatch(strings.NewReader(b.Nonce+"\n"), &out, b, true, now.Add(BatchApprovalTTL+time.Second))
	if !errors.Is(err, ErrConfirmationExpired) {
		t.Fatalf("err = %v, want ErrConfirmationExpired", err)
	}
}

// TestBatchNoncesDoNotRepeatAndDoNotLookLikeMutationNonces. Two different gates
// mean two answers an operator could confuse; the prefixes keep them apart.
func TestBatchNoncesDoNotRepeatAndDoNotLookLikeMutationNonces(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		b := testBatch(time.Now())
		if seen[b.Nonce] {
			t.Fatalf("batch nonce %q repeated", b.Nonce)
		}
		seen[b.Nonce] = true
		if !strings.HasPrefix(b.Nonce, "APPROVE-") {
			t.Fatalf("batch nonce %q does not identify what it approves", b.Nonce)
		}
		if strings.HasPrefix(b.Nonce, "VERIFY-") {
			t.Fatalf("batch nonce %q is indistinguishable from a per-mutation one", b.Nonce)
		}
	}
}

// TestTheBatchApprovalOutlastsAReadOfTheList, but not by much: it has to survive
// somebody reading a dozen orders and it has to expire while the quotes it was
// derived from are still recognisable.
func TestTheBatchApprovalOutlastsAReadOfTheList(t *testing.T) {
	if BatchApprovalTTL <= ConfirmationTTL {
		t.Errorf("the batch window (%s) is no longer than a single mutation's (%s), although there is a whole "+
			"list to read", BatchApprovalTTL, ConfirmationTTL)
	}
	if BatchApprovalTTL > 15*time.Minute {
		t.Errorf("the batch window is %s; an approval typed that long after the list was priced is an "+
			"approval of different numbers", BatchApprovalTTL)
	}
}
