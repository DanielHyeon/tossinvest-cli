package soak

// protection_probe.go surveys the three reads the resident-protection path
// makes (a100 tasks 0.10 (a)).
//
// # Why they are here rather than in the supervised live check
//
// The supervised check can lend the attestation a mutation and nothing else:
// acceptSupervised refuses a read on the grounds that one supervised success is
// not what days of unattended operation prove. The convergence worker calls
// these three on every cycle it runs, so they are exactly the kind of claim the
// soak exists to make and exactly the kind the live check may not make for it.
//
// # What they do not do
//
// Nothing here enters RequiredEndpoints or LiveOnlyEndpoints. Those two lists
// are refusals — one makes `soak attest` refuse, the other makes
// BuildAttestation refuse a supervised proof — and an endpoint reaches the
// attestation without either of them, because BuildAttestation carries every
// endpoint that succeeded inside the window. Probing first and requiring later
// is the only order in which the evidence exists before the refusal does. The
// change that teaches the engine to need these calls is the change that adds
// them to the lists, in the deploy that carries that engine.

import (
	"context"
	"strings"
)

// The conditional-order groups the list accepts, spelled as
// verifylive.ConditionalStatusOpen / ConditionalStatusClosed spell them.
//
// Measured 2026-07-27 (verify-execution-capability measurements.md M17): a value
// outside these two is answered 400 invalid-request with
// allowedValues=["OPEN","CLOSED"]. A conditional's own status (WATCHING) belongs
// to a different vocabulary and is not a filter value.
const (
	conditionalStatusOpen   = "OPEN"
	conditionalStatusClosed = "CLOSED"
)

// probeConditionalOrders reads the conditional-order list, both groups.
//
// Both are read for the same reason probeOrders reads both order groups: the
// gateway's List and its post-cancel confirmation each walk OPEN and then
// CLOSED (protectionofficial/gateway.go:113, :189), so a survey that proved only
// one group would attest an endpoint the engine uses in a way nobody measured.
// One failed group fails the endpoint.
//
// Two things it deliberately does not do. It does not filter by symbol, where
// the gateway does — the unfiltered call is the broader exercise of the same
// endpoint and it is also what finds an identifier for the by-id read below. And
// it reads the first page of each group only: an account's resting protection is
// not a history the way its closed orders are, so paging it would spend the rate
// budget the 429 measurements (M8) say is the scarce thing here. The engine's
// own pagination handling is bounded and fails closed on its own
// (gateway.go:211); what this probe answers is whether the endpoint answers.
//
// Requests counts both group reads rather than the one probe, the same way
// probeOrders sums its two walks: a call the broker answered still spent the
// account's rate budget, and this record is what that budget is sized from
// (measurements.md M8). One probe reporting one request for two calls would
// understate the burst by half.
func (r *Runner) probeConditionalOrders(ctx context.Context) ([]string, EndpointResult) {
	started := r.opts.Clock.Now()

	var ids []string
	var err error
	requests := 0
	for _, status := range []string{conditionalStatusOpen, conditionalStatusClosed} {
		var group []string
		requests++
		if group, err = r.opts.Reads.ConditionalOrders(ctx, status); err != nil {
			break
		}
		ids = append(ids, group...)
	}

	result := r.result(EndpointConditionalOrders, requests, started, err)
	if !result.OK {
		// A partial list is not a list. Whatever the first group returned before
		// the second one failed must not become the identifier the by-id read
		// claims to have proven the endpoint with.
		ids = nil
	}
	return ids, result
}

// probeConditionalByID reads the first conditional order the list returned.
//
// With none, the result is a recorded skip rather than a success — the same
// contract probeOrderByID has, for the same reason: an account with no resting
// protection has not exercised the endpoint, and an attestation that claimed
// otherwise would be claiming a call nobody has seen work.
//
// An empty conditional list is the ordinary state of this account today, so the
// skip is the expected outcome and not a fault. What makes the read provable is
// a conditional order existing at the moment a cycle runs.
//
// listOK separates the two ways there can be no identifier, because they are not
// the same news: an account with no resting protection is ordinary, and a list
// that could not be read is a failure already recorded against another endpoint.
// Reporting the second as the first would tell an operator their account is
// empty when what actually happened is that nobody could look.
func (r *Runner) probeConditionalByID(ctx context.Context, ids []string, listOK bool) EndpointResult {
	id := ""
	for _, candidate := range ids {
		if strings.TrimSpace(candidate) != "" {
			id = strings.TrimSpace(candidate)
			break
		}
	}
	if id == "" {
		reason := "the account has no conditional order, so there is no id to read"
		if !listOK {
			reason = "the conditional-order list could not be read, so no id was available"
		}
		return EndpointResult{
			Endpoint:   EndpointConditionalByID,
			Skipped:    true,
			SkipReason: reason,
		}
	}
	return r.probe(ctx, EndpointConditionalByID, func(ctx context.Context) error {
		return r.opts.Reads.ConditionalOrder(ctx, id)
	})
}

// probeSellableQuantity reads how much of a symbol may be sold.
//
// The quantity is read and discarded: this package records that an endpoint
// answered, never what it answered. The symbol is the first the survey was
// configured to quote, which keeps the probe's input in the operator's config
// rather than derived from the account's holdings — soak.Reads returns a count
// of positions and not their symbols, and it should stay that way.
func (r *Runner) probeSellableQuantity(ctx context.Context) EndpointResult {
	symbol := r.opts.Symbols[0]
	return r.probe(ctx, EndpointSellableQuantity, func(ctx context.Context) error {
		return r.opts.Reads.SellableQuantity(ctx, symbol)
	})
}
