package verifylive

// confirm.go is the human gate in front of the live mutations this tool makes.
//
// It is internal/flatten/confirm.go's design, and the reasoning transfers
// unchanged: a confirmation string with no expiry can be copied into a runbook and
// typed months later against an account that has changed, so the nonce is derived
// from *this* run and stops being valid shortly after it is shown. There is no
// --yes, no environment variable that stands in for one, and neither confirmation
// below proceeds without an interactive terminal. An agent cannot answer these
// prompts, which is the point (WORKFLOW §0.7).
//
// # Two gates, one of them the default
//
//	Batch     one expiring string for the whole run. Before anything is sent, every
//	          live request the run can make is listed — action, symbol, side,
//	          quantity, how the price is derived, how the exposure ends — and one
//	          string approves exactly that list. This is the default, and it is what
//	          tasks.md 1.5 settled on (사용자 결정 2026-07-26): the procedure is a
//	          one-off measurement of fifteen or so requests, and a person answering
//	          fifteen separate prompts stops reading them by the fourth.
//	Mutation  one expiring string immediately before each request, opted into with
//	          --confirm-each. It is the finer gate and it is kept, because an
//	          operator who wants to stop halfway through a boundary probe should be
//	          able to.
//
// The batch model is only as strong as the list. A summary that omitted a mutation
// would be consent obtained for something the operator never saw, so the list is
// derived structurally from the step catalogue (plan.go) and Plan.Authorises is
// what mutate.go acts on: a request the approved list does not carry a line for is
// not sent, and the run stops rather than adapting.
//
// The cancels are on the list too, which is the one place this differs from
// flatten. There, §0.3 forbids putting a prompt in front of a cancel because
// cancelling only reduces exposure. Here the cancel is part of a measurement — it
// is the step's own claim under test — and leaving it off the list would mean the
// record could not say a person saw it coming. Every line says plainly which way
// the exposure moves, so an operator is never misled about the risk.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

// ConfirmationTTL is how long a per-mutation nonce stays valid. Two minutes, as
// flatten's is: long enough to read the summary and think, short enough that the
// account cannot meaningfully change between the plan and the answer.
const ConfirmationTTL = 2 * time.Minute

// BatchApprovalTTL is how long the run-wide approval stays valid.
//
// Longer than ConfirmationTTL because there is genuinely more to read — a dozen
// requests with their prices and their reversals — and a window that expires while
// somebody is still reading trains them to type first and read afterwards, which is
// the exact behaviour the prompt exists to prevent. Still minutes, not hours: the
// last trade every price is derived from moves, and an approval typed long after
// the list was printed would be an approval of different numbers.
const BatchApprovalTTL = 5 * time.Minute

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

// --- the batch approval ---------------------------------------------------------

// Batch is the run-wide approval: a complete plan and one nonce that covers it.
type Batch struct {
	// Plan is every live request this run can make. Typing the nonce approves
	// exactly this list and nothing else.
	Plan Plan
	// Resumed reports that the run is continuing a verification already on the
	// record, so the list below is the *remaining* batch. A resumed run is
	// approved again from scratch: the earlier approval covered earlier requests.
	Resumed bool
	// Nonce is what the operator types.
	Nonce string
	// ExpiresAt is when it stops being accepted.
	ExpiresAt time.Time
}

// NewBatch builds the run's approval prompt.
func NewBatch(plan Plan, resumed bool, now time.Time) Batch {
	return Batch{
		Plan:      plan,
		Resumed:   resumed,
		Nonce:     newToken("APPROVE"),
		ExpiresAt: now.Add(BatchApprovalTTL),
	}
}

// Prompt renders the whole list and the one string that approves it.
func (b Batch) Prompt() string {
	var s strings.Builder
	scope := "this run"
	if b.Resumed {
		scope = "the REMAINING part of this verification"
	}
	fmt.Fprintf(&s, "\nLIVE MUTATION BATCH — %d request(s) planned for %s\n\n", len(b.Plan.Mutations), scope)
	fmt.Fprintf(&s, "  account          %s\n", b.Plan.Account)
	fmt.Fprintf(&s, "  run              %s\n\n", b.Plan.RunID)
	s.WriteString("  Everything this run can send is listed below. Each order is a LIMIT order for the " +
		"quantity\n  shown, priced so it cannot fill, and each line says how its exposure ends.\n\n")

	b.Plan.WriteLines(&s)

	fmt.Fprintf(&s, "\n  confirmation     %s (expires %s)\n\n",
		b.Nonce, b.ExpiresAt.UTC().Format("15:04:05Z"))
	s.WriteString("These are real requests against the real account. Typing the confirmation approves the " +
		"list\nabove and nothing else: a step that would have to send something not on it stops the run and\n" +
		"asks again rather than adapting. Prices are re-quoted by the stated rule when each step runs.\n")
	s.WriteString("Type the confirmation string to approve the batch, anything else to abort: ")
	return s.String()
}

// Verify checks a typed answer against the batch.
func (b Batch) Verify(input string, now time.Time) error {
	if !now.Before(b.ExpiresAt) {
		return ErrConfirmationExpired
	}
	if strings.TrimSpace(input) != b.Nonce {
		return ErrRefused
	}
	return nil
}

// ConfirmBatch shows the whole plan and reads the one answer.
//
// interactive is the caller's real TTY check, for the same reason it is on Confirm:
// redirecting stdin must not satisfy the terminal requirement, since redirection is
// exactly the case the requirement exists to catch.
func ConfirmBatch(in io.Reader, out io.Writer, b Batch, interactive bool, now time.Time) error {
	if !interactive {
		return ErrNotATerminal
	}
	if _, err := io.WriteString(out, b.Prompt()); err != nil {
		return err
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("verify: reading the batch approval: %w", err)
	}
	return b.Verify(line, now)
}

// BatchConfirmer is how the runner asks for the run-wide approval.
type BatchConfirmer func(b Batch) error

// newNonce returns a short, unambiguous, typeable token.
//
// Base32 without padding, as flatten's is: short enough to retype without a
// copy-paste, because copy-paste is how a confirmation stops being read.
func newNonce() string { return newToken("VERIFY") }
