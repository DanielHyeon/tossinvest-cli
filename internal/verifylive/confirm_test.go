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
