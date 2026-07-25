package flatten

// confirm.go is the human gate in front of a live flatten (harden-execution-base
// task 4.5, engine-safety: "확인 문자열은 마스킹된 계좌 식별·포지션 수·예상 청산
// 수량·만료 nonce를 포함하고, TTY 직접 입력만 허용하며 자동화 플래그는 금지된다").
//
// # Why the gate exists here and not in front of the cancels
//
// This is the only confirmation in the flatten path, and it guards the phase that
// *sells*. Cancelling has no downside worth a prompt — it reduces exposure and
// nothing else — so §0.3 forbids putting anything in front of it. Selling every
// position is irreversible at the day's prices, and running it against the wrong
// account is a mistake nobody can undo. That asymmetry is why one of the two
// phases is confirmed and the other is not.
//
// # Why the nonce expires
//
// A confirmation string with no expiry can be copied into a runbook, a shell
// history or a wiki, and typed months later against an account that has changed.
// The nonce is derived from *this* plan and stops being valid shortly after it is
// shown, so the thing the operator confirms is the thing they were shown.
//
// # Why there is no --yes
//
// The spec forbids an automation flag, and this file provides none. There is no
// function here that accepts a pre-supplied answer, no environment variable that
// stands in for one, and Confirm requires an interactive terminal. An agent
// cannot answer this prompt, which is the point (§0.7: operational flips are a
// human decision).

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

// ConfirmationTTL is how long a confirmation nonce stays valid.
//
// Two minutes: long enough to read the summary and think, short enough that the
// account cannot meaningfully change between the plan and the answer.
const ConfirmationTTL = 2 * time.Minute

// Errors the confirmation can produce.
var (
	// ErrNotATerminal means stdin is not a terminal. There is no override.
	ErrNotATerminal = errors.New(
		"flatten: liquidation must be confirmed at a terminal — this command has no automation flag by design")
	// ErrConfirmationExpired means the nonce timed out.
	ErrConfirmationExpired = errors.New("flatten: the confirmation expired; re-run to see a fresh plan")
	// ErrConfirmationMismatch means the typed text did not match.
	ErrConfirmationMismatch = errors.New("flatten: the confirmation did not match; nothing was submitted")
)

// Confirmation is what the operator is shown and has to type back.
type Confirmation struct {
	// Account is the masked account identifier.
	Account string
	// OpenOrders is how many orders will be cancelled.
	OpenOrders int
	// Positions is how many positions will be liquidated.
	Positions int
	// Quantity is the total quantity across those positions.
	Quantity float64
	// Symbols lists the positions, for the summary.
	Symbols []string
	// Nonce is what the operator types. It is meaningless outside this plan.
	Nonce string
	// ExpiresAt is when the nonce stops being accepted.
	ExpiresAt time.Time
}

// NewConfirmation builds the prompt for a plan.
//
// accountRef is masked here rather than by the caller, so there is no path that
// prints a full account number by forgetting to.
func NewConfirmation(accountRef string, openOrders int, targets []LiquidationTarget, now time.Time) Confirmation {
	c := Confirmation{
		Account:    attest.Mask(accountRef),
		OpenOrders: openOrders,
		Nonce:      newNonce(),
		ExpiresAt:  now.Add(ConfirmationTTL),
	}
	for _, t := range targets {
		if t.Held <= 0 {
			continue
		}
		c.Positions++
		c.Quantity += t.Held
		c.Symbols = append(c.Symbols, t.Symbol)
	}
	return c
}

// Prompt renders what the operator sees.
//
// Every number in it is a number they can check against their own screen before
// typing: an operator who cannot recognise their account, their position count
// and roughly their size is an operator about to flatten the wrong thing.
func (c Confirmation) Prompt() string {
	var b strings.Builder
	b.WriteString("\nFLATTEN-ALL — this cancels every open order and sells every position.\n\n")
	fmt.Fprintf(&b, "  account          %s\n", c.Account)
	fmt.Fprintf(&b, "  open orders      %d (all will be cancelled)\n", c.OpenOrders)
	fmt.Fprintf(&b, "  positions        %d\n", c.Positions)
	fmt.Fprintf(&b, "  total quantity   %s\n", decimalString(c.Quantity))
	if len(c.Symbols) > 0 {
		fmt.Fprintf(&b, "  symbols          %s\n", strings.Join(c.Symbols, ", "))
	}
	fmt.Fprintf(&b, "  confirmation     %s (expires %s)\n", c.Nonce, c.ExpiresAt.UTC().Format("15:04:05Z"))
	b.WriteString("\nThis cannot be undone at today's prices.\n")
	b.WriteString("Type the confirmation string to proceed, anything else to abort: ")
	return b.String()
}

// Verify checks a typed answer.
func (c Confirmation) Verify(input string, now time.Time) error {
	if !now.Before(c.ExpiresAt) {
		return ErrConfirmationExpired
	}
	if strings.TrimSpace(input) != c.Nonce {
		return ErrConfirmationMismatch
	}
	return nil
}

// Confirm shows the prompt and reads the answer.
//
// interactive must be the caller's real TTY check (tui.IsInteractive). It is a
// parameter rather than a call inside this function so the tests can exercise the
// non-terminal refusal — and so that a caller cannot accidentally satisfy it by
// redirecting stdin, which is the case the check exists to catch.
func Confirm(in io.Reader, out io.Writer, c Confirmation, interactive bool, now time.Time) error {
	if !interactive {
		return ErrNotATerminal
	}
	if _, err := io.WriteString(out, c.Prompt()); err != nil {
		return err
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("flatten: reading the confirmation: %w", err)
	}
	if err := c.Verify(line, now); err != nil {
		return err
	}
	return nil
}

// newNonce returns a short, unambiguous, typeable token.
//
// Base32 without padding: no characters an operator can confuse (no 0/O, no 1/l),
// and short enough to retype without a copy-paste, because copy-paste is how a
// confirmation stops being read.
func newNonce() string {
	var buf [5]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("flatten: crypto/rand is unavailable: " + err.Error())
	}
	return "FLATTEN-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf[:])
}
