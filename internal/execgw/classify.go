package execgw

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// classifyMutation turns one broker call's result into the journal's dispatch
// classification.
//
// The rule the whole design rests on: this broker has no idempotency key, so a
// wrong "it did not happen" duplicates a live order. Anything that does not
// *prove* the outcome becomes AMBIGUOUS, which sends the attempt to IN_DOUBT and
// the resolution procedure — never to a retry.
func classifyMutation(send journal.SendState, result domain.MutationResult, err error) journal.DispatchOutcome {
	if err == nil {
		return journal.DispatchOutcome{
			Class:         journal.DispatchAcked,
			BrokerOrderID: brokerOrderID(result),
			ReasonCode:    string(ReasonBrokerAccepted),
		}
	}

	// Refusals raised by trading.Service happen before any broker contact, so
	// they are provably unsent whatever the tracker says.
	if reason, refused := policyRefusal(err); refused {
		return journal.DispatchOutcome{
			Class:      journal.DispatchNotSent,
			ReasonCode: string(reason),
			Detail:     err.Error(),
			Err:        err,
		}
	}

	// A broker answer that asks for a human — an auth challenge, an FX consent, a
	// deposit — is a well-formed refusal: the request was understood and declined,
	// so it definitively did not execute. It is classified here, before the
	// transport tracker gets a say, because the meaning comes from the answer and
	// not from how far the bytes got.
	if reason, refused := ClassifyBrokerRefusal(err); refused {
		class := journal.DispatchRejected
		var branch *trading.BranchRequiredError
		if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation {
			// The broker asked for confirmation *after* accepting a prepare. What
			// that leaves behind is not something we can assert, so it is doubt.
			class = journal.DispatchAmbiguous
		}
		return journal.DispatchOutcome{
			Class:      class,
			ReasonCode: string(reason),
			Detail:     err.Error(),
			Err:        err,
		}
	}

	// A status the broker actually returned classifies the outcome; the transport
	// tracker only decides the cases where there is no status at all.
	if status, known := statusOf(err); known {
		outcome := journal.ClassifyHTTPMutation(send, status, nil)
		outcome.Err = err
		if outcome.Detail == "" {
			outcome.Detail = err.Error()
		} else {
			outcome.Detail += ": " + err.Error()
		}
		outcome.ReasonCode = reasonForClass(outcome.Class, outcome.ReasonCode)
		return outcome
	}

	outcome := journal.ClassifyHTTPMutation(send, 0, err)
	outcome.ReasonCode = reasonForClass(outcome.Class, outcome.ReasonCode)
	return outcome
}

// reasonForClass keeps the gateway's reason vocabulary on outcomes the journal
// classified, without discarding a more specific code the journal already chose.
func reasonForClass(class journal.DispatchClass, existing string) string {
	if existing != "" {
		return existing
	}
	switch class {
	case journal.DispatchAcked:
		return string(ReasonBrokerAccepted)
	case journal.DispatchRejected:
		return string(ReasonBrokerRejected)
	case journal.DispatchNotSent:
		return journal.ReasonDispatchNotSent
	default:
		return string(ReasonBrokerOutcomeUnknown)
	}
}

// statusOf recovers the HTTP status behind an official-client error.
//
// internal/official maps statuses onto sentinels and drops the code, so this is
// the inverse mapping. It matters because the journal's classifier treats a
// well-formed 4xx refusal as definitive and everything else as ambiguous, and the
// difference decides whether a symbol gets blocked.
func statusOf(err error) (int, bool) {
	var apiErr *official.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code, true
	}
	switch {
	case errors.Is(err, official.ErrAuth), errors.Is(err, official.ErrIPNotAllowed):
		// 401/403: the request was understood and refused. It did not execute.
		return http.StatusUnauthorized, true
	case errors.Is(err, official.ErrRateLimited):
		// 429 is ambiguous by design: a limiter can answer after the order
		// reached the matching engine.
		return http.StatusTooManyRequests, true
	case errors.Is(err, official.ErrServer):
		return http.StatusInternalServerError, true
	default:
		return 0, false
	}
}

// policyRefusal recognises the errors trading.Service raises before it touches a
// broker. They are the config gate, the confirm-token gate and the capability
// check — all local, all provably unsent.
func policyRefusal(err error) (ReasonCode, bool) {
	var disabled *trading.DisabledActionError
	switch {
	case errors.As(err, &disabled):
		return ReasonPolicyDisabled, true
	case errors.Is(err, trading.ErrPlaceUnsupported):
		return ReasonUnsupportedOrderType, true
	case errors.Is(err, trading.ErrLiveActionsDisabled),
		errors.Is(err, trading.ErrExecuteRequired),
		errors.Is(err, trading.ErrConfirmMismatch),
		errors.Is(err, trading.ErrLiveMutationPending):
		return ReasonPolicyDisabled, true
	default:
		return "", false
	}
}

// brokerOrderID picks the order number a mutation result is addressable by. An
// amend reports the *new* order number in CurrentOrderID (the official API issues
// a fresh id), a place reports OrderID.
//
// The chosen candidate is returned exactly as the broker sent it. `orderId` is an
// opaque token — openapi contracts no shape, no prefix and no length for it — so
// the only transformation allowed here is none. Trimming is used to decide
// whether a candidate is a name at all, never to produce the value that gets
// stored and later compared byte-for-byte.
func brokerOrderID(res domain.MutationResult) string {
	for _, candidate := range []string{res.CurrentOrderID, res.OrderID} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}
