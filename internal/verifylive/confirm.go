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
// # Two renderings of one list
//
// The functions above are the terminal's. The console asks the same question with
// a click on a session- and CSRF-gated screen (operator-console: 검증 배치 승인의
// 형식), so it renders Summary — the list without the typed tail — and judges the
// window with Expired. Both readings come from the same Plan.WriteLines, because a
// second rendering would be a second list to keep in step with the first.
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
		"verify: 모든 라이브 mutation은 터미널에서 확인받아야 한다 — 이 명령에는 설계상 자동화 플래그가 없다")
	// ErrConfirmationExpired means the nonce timed out. It is a refusal, not a
	// failure: nothing was sent.
	ErrConfirmationExpired = errors.New("verify: 확인 문자열이 만료되었다. 아무것도 전송되지 않았다")
	// ErrRefused means the typed text did not match, or the operator sent EOF.
	ErrRefused = errors.New("verify: 확인 문자열이 일치하지 않는다. 아무것도 전송되지 않았다")
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
	fmt.Fprintf(&b, "\n라이브 MUTATION — 단계 %s (%s)\n\n", m.Step, StepLabel(m.Step))
	fmt.Fprintf(&b, "  동작             %s\n", m.Action)
	fmt.Fprintf(&b, "  계좌             %s\n", m.Account)
	for _, d := range m.Detail {
		fmt.Fprintf(&b, "  %s\n", d)
	}
	if strings.TrimSpace(m.Reversal) != "" {
		fmt.Fprintf(&b, "  되돌리기         %s\n", m.Reversal)
	}
	fmt.Fprintf(&b, "  확인 문자열      %s (만료 %s)\n", m.Nonce, m.ExpiresAt.UTC().Format("15:04:05Z"))
	b.WriteString("\n이것은 실제 계좌에 실제 요청을 보낸다.\n")
	b.WriteString("진행하려면 확인 문자열을 입력하라. 그 외의 입력은 이 단계를 거부한다: ")
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
		return fmt.Errorf("verify: 확인 입력을 읽는 중: %w", err)
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

// Summary renders the list itself — everything an approver has to read, and
// nothing about how they answer.
//
// It exists because the console approves with a click (operator-console: 검증 배치
// 승인의 형식) and a screen that printed "확인 문자열을 입력하라" would be describing an
// approval method that is not in force there. Splitting the string rather than
// re-rendering the plan in HTML keeps one source for the list: the terminal and
// the browser show the same bytes, which is what makes "the list is complete" a
// claim about one thing instead of two.
func (b Batch) Summary() string {
	var s strings.Builder
	scope := "이 실행"
	if b.Resumed {
		scope = "이 검증의 남은 부분"
	}
	fmt.Fprintf(&s, "\n라이브 MUTATION 배치 — %s에 계획된 요청 %d건\n\n", scope, len(b.Plan.Mutations))
	fmt.Fprintf(&s, "  계좌             %s\n", b.Plan.Account)
	fmt.Fprintf(&s, "  run              %s\n\n", b.Plan.RunID)
	s.WriteString("  이 실행이 보낼 수 있는 전부가 아래에 있다. 각 주문은 표시된 수량의 LIMIT 주문이고,\n" +
		"  체결될 수 없는 가격이며, 각 줄은 그 노출이 어떻게 끝나는지 밝힌다.\n\n")

	b.Plan.WriteLines(&s)

	s.WriteString("\n이것들은 실제 계좌에 나가는 실제 요청이다. 승인되는 것은 위 목록뿐이고 " +
		"그 외에는\n아무것도 승인되지 않는다 — 목록에 없는 것을 보내야 하는 단계는 적응하는 대신 실행을 " +
		"멈추고 다시 묻는다.\n가격은 각 단계가 실행될 때 명시된 규칙으로 재호가된다.\n")
	return s.String()
}

// Prompt renders the whole list and the one string that approves it at a
// terminal. It is Summary plus the typed tail, so the two cannot drift.
func (b Batch) Prompt() string {
	var s strings.Builder
	s.WriteString(b.Summary())
	fmt.Fprintf(&s, "\n  확인 문자열      %s (만료 %s)\n\n",
		b.Nonce, b.ExpiresAt.UTC().Format("15:04:05Z"))
	s.WriteString("배치를 승인하려면 확인 문자열을 입력하라. 그 외의 입력은 중단이다: ")
	return s.String()
}

// Expired reports that the approval window has closed.
//
// It is the same window Verify enforces, spelled once: the console judges an
// expiry without a typed string to compare, and a second clock rule here would be
// a second definition of "too late".
func (b Batch) Expired(now time.Time) bool { return !now.Before(b.ExpiresAt) }

// Verify checks a typed answer against the batch.
func (b Batch) Verify(input string, now time.Time) error {
	if b.Expired(now) {
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
		return fmt.Errorf("verify: 배치 승인 입력을 읽는 중: %w", err)
	}
	return b.Verify(line, now)
}

// BatchConfirmer is how the runner asks for the run-wide approval.
type BatchConfirmer func(b Batch) error

// The channels an approval can arrive through, as the evidence record names them.
//
// The record is what proves later that a person was shown a list and answered, so
// it has to say which act that was. ApprovalChannelTyped is the zero value's
// meaning — every caller that does not say otherwise is at a terminal.
const (
	ApprovalChannelTyped        = "one typed expiring string for the whole run"
	ApprovalChannelConsoleClick = "one click on the console's approval screen (loopback session + CSRF), " +
		"inside the same expiring window"
)

// newNonce returns a short, unambiguous, typeable token.
//
// Base32 without padding, as flatten's is: short enough to retype without a
// copy-paste, because copy-paste is how a confirmation stops being read.
func newNonce() string { return newToken("VERIFY") }
