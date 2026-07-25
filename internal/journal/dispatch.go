package journal

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptrace"
	"sync/atomic"
)

// DispatchClass is what we know about a mutation request's fate.
//
// The classification exists because the broker has no idempotency key: an
// automatic retry can double an order, so the *only* safe answer to "did it
// happen?" is one of these four, and three of them forbid retrying.
type DispatchClass string

const (
	// DispatchNotSent means we can prove the request never reached the broker —
	// for example the connection was refused before a single byte was written.
	// This is the only class that closes the attempt without any resolution work.
	DispatchNotSent DispatchClass = "NOT_SENT"
	// DispatchAcked means the broker accepted the mutation.
	DispatchAcked DispatchClass = "ACKED"
	// DispatchRejected means the broker refused it in a well-formed way, so it
	// definitively did not execute.
	DispatchRejected DispatchClass = "REJECTED"
	// DispatchAmbiguous means we cannot tell. The attempt becomes IN_DOUBT and
	// the resolution procedure — never a retry — decides.
	DispatchAmbiguous DispatchClass = "AMBIGUOUS"
)

// DispatchOutcome is what a DispatchFunc reports back.
type DispatchOutcome struct {
	Class DispatchClass
	// BrokerOrderID is the order number the broker named. Required for an ACK to
	// be confirmable: an ack we cannot address is not trackable.
	BrokerOrderID string
	// ReasonCode is a stable code recorded on the attempt.
	ReasonCode string
	// Detail is human-readable context for operators.
	Detail string
	// Err is the underlying transport or API error, if any.
	Err error
}

// DispatchFunc performs the actual broker call. It must not retry: the journal
// records exactly one dispatch per attempt.
type DispatchFunc func(ctx context.Context, a *Attempt) DispatchOutcome

// Result is the outcome of a journalled mutation.
type Result struct {
	AttemptID     string
	IntentID      string
	Final         AttemptState
	Class         DispatchClass
	BrokerOrderID string
	ReasonCode    string
	Detail        string
	// Err carries the dispatch error when there was one. It is informational:
	// Final is the authoritative outcome.
	Err error
}

// Blocking reports whether this outcome leaves the symbol blocked for new
// mutations (anything not settled, i.e. IN_DOUBT or a stuck ACK).
func (r Result) Blocking() bool { return !r.Final.IsTerminal() }

// Run performs one mutation with the ordering the order-execution spec requires:
//
//	Prepare (RECORDED, committed + fsynced)
//	  → MarkDispatchStarted (committed)
//	    → dispatch exactly once
//	      → settle from the reported class
//
// Every step that fails stops the sequence before the broker is touched. Run never
// retries a mutation.
func (j *Journal) Run(ctx context.Context, req PrepareRequest, dispatch DispatchFunc) (Result, error) {
	if dispatch == nil {
		return Result{}, fmt.Errorf("%w: dispatch function is nil", ErrInvalidRequest)
	}
	a, err := j.Prepare(ctx, req)
	if err != nil {
		return Result{}, err
	}
	return a.Dispatch(ctx, dispatch)
}

// Dispatch runs the second half of Run against an already-recorded attempt. It is
// separate so a restart or an operator flow can resume an attempt that was
// prepared by an earlier process.
func (a *Attempt) Dispatch(ctx context.Context, dispatch DispatchFunc) (Result, error) {
	if dispatch == nil {
		return Result{}, fmt.Errorf("%w: dispatch function is nil", ErrInvalidRequest)
	}
	res := Result{AttemptID: a.id, IntentID: a.intentID}

	// The broker is only called after this commit lands: a mutation the journal
	// does not know it started is exactly what crash recovery cannot reason about.
	if err := a.MarkDispatchStarted(ctx); err != nil {
		return res, err
	}

	outcome := dispatch(ctx, a)
	res.Class = outcome.Class
	res.BrokerOrderID = outcome.BrokerOrderID
	res.Detail = outcome.Detail
	res.Err = outcome.Err

	switch outcome.Class {
	case DispatchNotSent:
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchNotSent)
		if err := a.Settle(ctx, StateNotDispatched, res.ReasonCode, outcome.Detail); err != nil {
			return res, err
		}
		res.Final = StateNotDispatched

	case DispatchRejected:
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchRejected)
		if err := a.Settle(ctx, StateFailedConfirmed, res.ReasonCode, outcome.Detail); err != nil {
			return res, err
		}
		res.Final = StateFailedConfirmed

	case DispatchAcked:
		if outcome.BrokerOrderID == "" {
			// An ack with no order number cannot be tracked or cancelled, so it
			// is treated as doubt rather than success.
			res.ReasonCode = ReasonAckWithoutOrderID
			if err := a.MarkInDoubt(ctx, res.ReasonCode,
				"broker acknowledged the mutation without an order id"); err != nil {
				return res, err
			}
			res.Final = StateInDoubt
			break
		}
		if err := a.MarkAcked(ctx, outcome.BrokerOrderID); err != nil {
			return res, err
		}
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonBrokerAcknowledged)
		if err := a.Settle(ctx, StateConfirmed, res.ReasonCode, outcome.Detail); err != nil {
			return res, err
		}
		res.Final = StateConfirmed

	case DispatchAmbiguous:
		res.ReasonCode = firstNonEmpty(outcome.ReasonCode, ReasonDispatchAmbiguous)
		if err := a.MarkInDoubt(ctx, res.ReasonCode, outcome.Detail); err != nil {
			return res, err
		}
		res.Final = StateInDoubt

	default:
		// A class this build does not know is treated as doubt, never as success.
		res.ReasonCode = ReasonUnknownDispatchClass
		detail := fmt.Sprintf("dispatch reported class %q", outcome.Class)
		if err := a.MarkInDoubt(ctx, res.ReasonCode, detail); err != nil {
			return res, err
		}
		res.Final = StateInDoubt
		res.Detail = detail
	}

	return res, nil
}

// SendState is how far a request got before something went wrong.
type SendState int

const (
	// SendNotStarted: no request bytes were written (connect/DNS/TLS failure).
	SendNotStarted SendState = iota
	// SendPartial: writing began but did not finish, so bytes may have arrived.
	SendPartial
	// SendComplete: the request was fully written.
	SendComplete
)

func (s SendState) String() string {
	switch s {
	case SendNotStarted:
		return "not_started"
	case SendPartial:
		return "partial"
	case SendComplete:
		return "complete"
	default:
		return "unknown"
	}
}

// SendTracker reports how far an HTTP request got. It is the difference between
// "provably not sent" (safe to close) and "may have executed" (needs resolution),
// so the classification is not left to guessing from the error text.
type SendTracker struct {
	state atomic.Int32
}

// State returns the furthest point the request reached.
func (t *SendTracker) State() SendState { return SendState(t.state.Load()) }

func (t *SendTracker) set(s SendState) {
	for {
		current := t.state.Load()
		if SendState(current) >= s {
			return
		}
		if t.state.CompareAndSwap(current, int32(s)) {
			return
		}
	}
}

// WithSendTracker returns a context that observes how far the outgoing HTTP request
// got, plus the tracker to read afterwards.
//
// This lives in the journal package because the dispatch classification is a
// journal concern: the lifecycle state an attempt lands in depends on it, and the
// same tracker must be used by every caller that records a mutation.
func WithSendTracker(ctx context.Context) (context.Context, *SendTracker) {
	tracker := &SendTracker{}
	trace := &httptrace.ClientTrace{
		WroteHeaderField: func(string, []string) { tracker.set(SendPartial) },
		WroteHeaders:     func() { tracker.set(SendPartial) },
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err != nil {
				tracker.set(SendPartial)
				return
			}
			tracker.set(SendComplete)
		},
	}
	return httptrace.WithClientTrace(ctx, trace), tracker
}

// ClassifyHTTPMutation maps a mutation request's transport result to a class.
//
// The mapping is deliberately pessimistic. 429 and 5xx are ambiguous rather than
// "not executed": a rate limiter or a proxy can answer after the order reached the
// matching engine, and being wrong in that direction duplicates a live order. Only
// well-formed refusals (4xx that describe the request itself) are treated as
// definitive rejections.
func ClassifyHTTPMutation(send SendState, statusCode int, err error) DispatchOutcome {
	if err != nil {
		if send == SendNotStarted {
			return DispatchOutcome{
				Class:      DispatchNotSent,
				ReasonCode: ReasonDispatchNotSent,
				Detail:     "request never left the process: " + err.Error(),
				Err:        err,
			}
		}
		return DispatchOutcome{
			Class:      DispatchAmbiguous,
			ReasonCode: ReasonDispatchAmbiguous,
			Detail: fmt.Sprintf("transport failed with the request %s: %s",
				send.String(), err.Error()),
			Err: err,
		}
	}

	switch {
	case statusCode >= 200 && statusCode < 300:
		return DispatchOutcome{Class: DispatchAcked}
	case isDefinitiveRejection(statusCode):
		return DispatchOutcome{
			Class:      DispatchRejected,
			ReasonCode: ReasonDispatchRejected,
			Detail:     fmt.Sprintf("broker refused the mutation with HTTP %d", statusCode),
			Err:        fmt.Errorf("broker returned HTTP %d", statusCode),
		}
	case statusCode == 0:
		return DispatchOutcome{
			Class:      DispatchAmbiguous,
			ReasonCode: ReasonDispatchAmbiguous,
			Detail:     "no HTTP status and no error were reported",
			Err:        errors.New("dispatch reported neither a status nor an error"),
		}
	default:
		return DispatchOutcome{
			Class:      DispatchAmbiguous,
			ReasonCode: ReasonDispatchAmbiguous,
			Detail:     fmt.Sprintf("HTTP %d does not prove whether the mutation executed", statusCode),
			Err:        fmt.Errorf("broker returned HTTP %d", statusCode),
		}
	}
}

// isDefinitiveRejection lists the statuses that describe the request rather than
// the broker's ability to process it.
func isDefinitiveRejection(statusCode int) bool {
	switch statusCode {
	case 400, 401, 403, 404, 405, 415, 422:
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
