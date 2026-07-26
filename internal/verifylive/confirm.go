package verifylive

// confirm.go is the human gate in front of every live mutation this tool makes.
//
// It is internal/flatten/confirm.go's design, applied per mutation instead of per
// run, and the reasoning transfers unchanged: a confirmation string with no
// expiry can be copied into a runbook and typed months later against an account
// that has changed, so the nonce is derived from *this* mutation and stops being
// valid shortly after it is shown. There is no --yes, no environment variable
// that stands in for one, and Confirm requires an interactive terminal. An agent
// cannot answer this prompt, which is the point (WORKFLOW §0.7).
//
// # Why every mutation and not just the first
//
// flatten-all confirms once because its two phases are one decision: cancel
// everything, then sell everything. This tool's mutations are unrelated to each
// other — a limit buy for the idempotency check, a conditional modify twenty
// minutes later — and a single up-front "yes" would be consent to a list the
// operator has not read yet. So each one is shown, in full, immediately before it
// is sent, and each one is refusable on its own.
//
// The cancels are confirmed too, which is the one place this differs from
// flatten. There, §0.3 forbids putting a prompt in front of a cancel because
// cancelling only reduces exposure. Here the cancel is part of a measurement — it
// is the step's own claim under test — and skipping the prompt would mean the
// record could not say a person watched it happen. The prompt says plainly that
// the cancel reduces exposure, so an operator is never misled about which way the
// risk points.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

// ConfirmationTTL is how long a confirmation nonce stays valid. Two minutes, as
// flatten's is: long enough to read the summary and think, short enough that the
// account cannot meaningfully change between the plan and the answer.
const ConfirmationTTL = 2 * time.Minute

// Errors the confirmation can produce.
var (
	// ErrNotATerminal means stdin is not a terminal. There is no override, and
	// there is no step that proceeds without one.
	ErrNotATerminal = errors.New(
		"verify: every live mutation must be confirmed at a terminal — this command has no automation flag by design")
	// ErrConfirmationExpired means the nonce timed out. It is a refusal, not a
	// failure: nothing was sent.
	ErrConfirmationExpired = errors.New("verify: the confirmation expired; nothing was sent")
	// ErrRefused means the typed text did not match, or the operator sent EOF.
	ErrRefused = errors.New("verify: the confirmation did not match; nothing was sent")
)

// Mutation is what the operator is shown and has to type back.
type Mutation struct {
	// Step is the step asking.
	Step StepID
	// Action is the one-line description: "place a limit buy", "cancel order".
	Action string
	// Account is the masked account identifier.
	Account string
	// Detail is the mutation's own summary — symbol, side, quantity, price,
	// notional. Every number in it is one the operator can check against their
	// own screen before typing.
	Detail []string
	// Reversal says how the exposure ends. A mutation with no answer here is a
	// mutation this tool should not be making.
	Reversal string
	// Nonce is what the operator types. It is meaningless outside this mutation.
	Nonce string
	// ExpiresAt is when the nonce stops being accepted.
	ExpiresAt time.Time
}

// NewMutation builds the prompt for one mutation.
//
// accountRef is masked here rather than by the caller, so there is no path that
// prints a full account number by forgetting to.
func NewMutation(step StepID, action, accountRef string, detail []string, reversal string, now time.Time) Mutation {
	return Mutation{
		Step:      step,
		Action:    action,
		Account:   attest.Mask(accountRef),
		Detail:    detail,
		Reversal:  reversal,
		Nonce:     newNonce(),
		ExpiresAt: now.Add(ConfirmationTTL),
	}
}

// Prompt renders what the operator sees.
func (m Mutation) Prompt() string {
	var b strings.Builder
	fmt.Fprintf(&b, "\nLIVE MUTATION — step %s\n\n", m.Step)
	fmt.Fprintf(&b, "  action           %s\n", m.Action)
	fmt.Fprintf(&b, "  account          %s\n", m.Account)
	for _, d := range m.Detail {
		fmt.Fprintf(&b, "  %s\n", d)
	}
	if strings.TrimSpace(m.Reversal) != "" {
		fmt.Fprintf(&b, "  reversal         %s\n", m.Reversal)
	}
	fmt.Fprintf(&b, "  confirmation     %s (expires %s)\n", m.Nonce, m.ExpiresAt.UTC().Format("15:04:05Z"))
	b.WriteString("\nThis sends a real request to the real account.\n")
	b.WriteString("Type the confirmation string to proceed, anything else to refuse this step: ")
	return b.String()
}

// Verify checks a typed answer.
func (m Mutation) Verify(input string, now time.Time) error {
	if !now.Before(m.ExpiresAt) {
		return ErrConfirmationExpired
	}
	if strings.TrimSpace(input) != m.Nonce {
		return ErrRefused
	}
	return nil
}

// Confirm shows the prompt and reads the answer.
//
// interactive must be the caller's real TTY check (tui.IsInteractive). It is a
// parameter rather than a call inside this function so the tests can exercise the
// non-terminal refusal — and so that a caller cannot accidentally satisfy it by
// redirecting stdin, which is the case the check exists to catch.
func Confirm(in io.Reader, out io.Writer, m Mutation, interactive bool, now time.Time) error {
	if !interactive {
		return ErrNotATerminal
	}
	if _, err := io.WriteString(out, m.Prompt()); err != nil {
		return err
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("verify: reading the confirmation: %w", err)
	}
	return m.Verify(line, now)
}

// Confirmer is how the runner asks. The command supplies one bound to the real
// terminal; the tests supply one that answers from a script.
type Confirmer func(m Mutation) error

// newNonce returns a short, unambiguous, typeable token.
//
// Base32 without padding, as flatten's is: short enough to retype without a
// copy-paste, because copy-paste is how a confirmation stops being read.
func newNonce() string { return newToken("VERIFY") }
