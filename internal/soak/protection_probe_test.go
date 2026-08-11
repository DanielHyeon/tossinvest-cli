package soak_test

// protection_probe_test.go covers the three conditional-order reads the resident
// protection path makes (a100 tasks 0.10 (a)).
//
// The endpoints themselves are ordinary GETs and prove nothing new about the
// survey. What these tests are actually about is the blast radius of adding
// them: the attestation the engine's automation gate is interlocked on is built
// from this record, and three more reads per cycle is three more ways for a
// cycle to look worse than the account is.
//
// So each test below pins one thing the new reads must NOT be able to reach.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// errConditionalRead is what a refused conditional read looks like to the probe.
var errConditionalRead = errors.New("official: 503 service unavailable")

// The stub's three conditional reads. They live here rather than beside the
// other stubReads methods so that adding them does not land a diff hunk against
// stubReads.Prices — the Function Logic Map gate reads an edit adjacent to a
// function as an edit to it, and a stub method is not evidence about production
// logic.
//
// conditionalIDs is what the list hands back, and it is what decides whether the
// by-id read has an identifier to use at all. It has no default: an account with
// no resting protection is the ordinary state, so a test that wants one says so.
func (s *stubReads) ConditionalOrders(_ context.Context, status string) ([]string, error) {
	s.mu.Lock()
	s.conditionalStatusesAsked = append(s.conditionalStatusesAsked, status)
	s.mu.Unlock()
	if err := s.errs[soak.EndpointConditionalOrders]; err != nil {
		return nil, err
	}
	return append([]string(nil), s.conditionalIDs...), nil
}

func (s *stubReads) ConditionalOrder(_ context.Context, id string) error {
	s.mu.Lock()
	s.conditionalsRead = append(s.conditionalsRead, id)
	s.mu.Unlock()
	return s.errs[soak.EndpointConditionalByID]
}

func (s *stubReads) SellableQuantity(_ context.Context, symbol string) error {
	s.mu.Lock()
	s.sellableAsked = append(s.sellableAsked, symbol)
	s.mu.Unlock()
	return s.errs[soak.EndpointSellableQuantity]
}

// failAllConditionalReads is the stub with every conditional read refused: the
// worst case the new probes can produce.
func failAllConditionalReads() *stubReads {
	stub := newStubReads()
	stub.fail(soak.EndpointConditionalOrders, errConditionalRead)
	stub.fail(soak.EndpointConditionalByID, errConditionalRead)
	stub.fail(soak.EndpointSellableQuantity, errConditionalRead)
	return stub
}

// TestConditionalReadFailureLeavesTheCredentialStreakIntact.
//
// The consecutive-day streak is a claim about credentials and nothing else, and
// Evaluate requires three days of it. If a refused conditional read could mark
// the credential bad, one broker hiccup on an endpoint nothing currently needs
// would cost three days of streak — and the attestation expiring on 2026-08-29
// has to be re-issued whether a100 ships or not.
func TestConditionalReadFailureLeavesTheCredentialStreakIntact(t *testing.T) {
	r, _ := newRunner(t, failAllConditionalReads(), nil)

	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if !c.Credential.OK {
		t.Errorf("Credential.OK = false because a conditional read failed; the credential judgement "+
			"comes from GET /api/v1/accounts alone (class %s)", c.Credential.Class)
	}
	if c.Credential.Class != soak.ClassOK {
		t.Errorf("Credential.Class = %q, want %q — a failed conditional read is not a credential failure",
			c.Credential.Class, soak.ClassOK)
	}
}

// TestConditionalReadFailureLeavesCompletenessIntact.
//
// completenessOf answers "if the engine reads this API, does it see everything
// that exists?" over the order walk and the quotes. A conditional read is not an
// input to that question, and a cycle inside the window that fails the
// completeness check is a reason Evaluate refuses to attest at all.
func TestConditionalReadFailureLeavesCompletenessIntact(t *testing.T) {
	r, _ := newRunner(t, failAllConditionalReads(), nil)

	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if !c.Completeness.Evaluated {
		t.Error("Completeness.Evaluated = false; the order walk succeeded, so it should have been evaluated")
	}
	if !c.Completeness.OK {
		t.Errorf("Completeness.OK = false (%s); a refused conditional read is not a completeness failure",
			c.Completeness.Detail)
	}
}

// TestConditionalReadsAreRecordedAfterTheQuoteRead.
//
// Position in the cycle is a rate-budget decision, not a cosmetic one. On
// 2026-07-27 the order walk opened a 429 penalty window that the read following
// it fell into (measurements.md M8). Three more requests placed before the reads
// the attestation already depends on could push those into the same window;
// placed last, a penalty the new reads provoke lands on the new reads.
func TestConditionalReadsAreRecordedAfterTheQuoteRead(t *testing.T) {
	r, _ := newRunner(t, newStubReads(), nil)

	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	at := func(endpoint string) int {
		t.Helper()
		for i, e := range c.Endpoints {
			if e.Endpoint == endpoint {
				return i
			}
		}
		t.Fatalf("cycle recorded no result for %s", endpoint)
		return -1
	}

	prices := at(soak.EndpointPrices)
	for _, later := range []string{
		soak.EndpointConditionalOrders,
		soak.EndpointConditionalByID,
		soak.EndpointSellableQuantity,
	} {
		if at(later) < prices {
			t.Errorf("%s is recorded before %s; the conditional reads go last so a 429 they provoke "+
				"cannot catch an endpoint the attestation already depends on", later, soak.EndpointPrices)
		}
	}
}

// TestConditionalOrderByIDIsSkippedWhenTheAccountHasNone.
//
// Same contract as the order-by-id probe: with nothing to read, the result is a
// recorded skip and not a success. An account with no conditional order has not
// proven the endpoint, and saying otherwise would put an unexercised call into
// an attestation the engine treats as permission to trade unattended.
func TestConditionalOrderByIDIsSkippedWhenTheAccountHasNone(t *testing.T) {
	stub := newStubReads()
	stub.conditionalIDs = nil

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	got := endpointResult(t, c, soak.EndpointConditionalByID)
	if !got.Skipped {
		t.Errorf("%s: Skipped = false (OK=%v); with no conditional order there is no id to read",
			soak.EndpointConditionalByID, got.OK)
	}
	if got.OK {
		t.Errorf("%s: OK = true although nothing was read", soak.EndpointConditionalByID)
	}
	if strings.TrimSpace(got.SkipReason) == "" {
		t.Errorf("%s: skipped with no reason; the operator has to be told why", soak.EndpointConditionalByID)
	}
}

// TestConditionalReadsAreProvenWhenTheAccountHasOne is the other half: given a
// conditional order, the by-id read is exercised against its identifier.
func TestConditionalReadsAreProvenWhenTheAccountHasOne(t *testing.T) {
	stub := newStubReads()
	stub.conditionalIDs = []string{"c1"}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	for _, want := range []string{
		soak.EndpointConditionalOrders,
		soak.EndpointConditionalByID,
		soak.EndpointSellableQuantity,
	} {
		if got := endpointResult(t, c, want); !got.OK {
			t.Errorf("%s: OK = false (%s / %s)", want, got.Class, got.Error)
		}
	}
	if len(stub.conditionalsRead) != 1 || stub.conditionalsRead[0] != "c1" {
		t.Errorf("conditional read by id = %v, want the identifier the list returned", stub.conditionalsRead)
	}
	if len(stub.sellableAsked) != 1 || stub.sellableAsked[0] != "005930" {
		t.Errorf("sellable-quantity asked for %v, want the surveyed symbol", stub.sellableAsked)
	}
}

// TestTheConditionalListRecordsBothGroupRequests.
//
// Requests is what sizes the account's rate budget, and this account has already
// lost endpoints to a 429 opened by a burst nobody had counted (measurements.md
// M8). Two calls reported as one probe would understate the survey's own burst
// by half, and the understatement would be invisible: the cycle would look
// cheaper than it is right up until an endpoint starts failing.
func TestTheConditionalListRecordsBothGroupRequests(t *testing.T) {
	stub := newStubReads()
	r, _ := newRunner(t, stub, nil)

	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	got := endpointResult(t, c, soak.EndpointConditionalOrders)
	if got.Requests != 2 {
		t.Errorf("%s: Requests = %d, want 2 — the walk reads OPEN and CLOSED",
			soak.EndpointConditionalOrders, got.Requests)
	}
	if len(stub.conditionalStatusesAsked) != 2 {
		t.Errorf("the list was read %d time(s) (%v), want once per status group",
			len(stub.conditionalStatusesAsked), stub.conditionalStatusesAsked)
	}
}

// TestAFailedConditionalGroupIsNotAPartialList. Half a list is not a read of the
// list — the same rule probeOrders states — and an identifier salvaged from the
// group that answered would let the by-id read claim an endpoint the failed walk
// never established.
func TestAFailedConditionalGroupIsNotAPartialList(t *testing.T) {
	stub := newStubReads()
	stub.conditionalIDs = []string{"c1"}
	stub.fail(soak.EndpointConditionalOrders, errConditionalRead)

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if got := endpointResult(t, c, soak.EndpointConditionalOrders); got.OK {
		t.Error("the conditional list reported OK although a group read failed")
	}
	if got := endpointResult(t, c, soak.EndpointConditionalByID); !got.Skipped {
		t.Errorf("%s: Skipped = false after the list failed; there is no identifier to trust",
			soak.EndpointConditionalByID)
	}
	if len(stub.conditionalsRead) != 0 {
		t.Errorf("the by-id read ran with %v after the list failed", stub.conditionalsRead)
	}
	// And it must not say the account is empty when nobody could look.
	if reason := endpointResult(t, c, soak.EndpointConditionalByID).SkipReason; !strings.Contains(reason, "could not be read") {
		t.Errorf("skip reason = %q; a list that failed to read is not an account with no "+
			"conditional order, and an operator reading this has to be able to tell them apart", reason)
	}
}

// TestConditionalEndpointsAreReads keeps BuildAttestation's non-GET refusal off
// the new endpoints: it rejects the whole attestation over one non-GET in the
// record, on the grounds that a read-only tool cannot have measured a write.
func TestConditionalEndpointsAreReads(t *testing.T) {
	for _, e := range []string{
		soak.EndpointConditionalOrders,
		soak.EndpointConditionalByID,
		soak.EndpointSellableQuantity,
	} {
		if !strings.HasPrefix(e, "GET ") {
			t.Errorf("%q is not a read; the soak issues no mutating request", e)
		}
	}
}

// TestAProbedEndpointReachesTheAttestationWithoutBeingRequired is the property
// the whole staged plan rests on.
//
// BuildAttestation carries Window.SuccessfulEndpoints() — every endpoint that
// succeeded inside the streak window — and RequiredEndpoints only decides what
// Evaluate refuses over. That is what lets the evidence be produced first and
// the requirement arrive later, in the deploy that carries the engine needing
// it. If SuccessfulEndpoints were ever narrowed to the required set, this change
// would go on passing its own tests while quietly minting attestations without
// the three reads, and the engine would refuse to start after a deploy that
// looked clean.
func TestAProbedEndpointReachesTheAttestationWithoutBeingRequired(t *testing.T) {
	added := []string{
		soak.EndpointConditionalOrders,
		soak.EndpointConditionalByID,
		soak.EndpointSellableQuantity,
	}
	cycles := threeCleanDays()
	for i := range cycles {
		for _, endpoint := range added {
			cycles[i].Endpoints = append(cycles[i].Endpoints, soak.EndpointResult{
				Endpoint: endpoint, OK: true, Class: soak.ClassOK, Requests: 1,
			})
		}
	}

	s := soak.Summarize(cycles)
	now := s.LastAt.Add(time.Hour)
	a, err := soak.BuildAttestation(s, soak.DefaultCriteria(), now, "tester", "", nil)
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}

	carried := map[string]bool{}
	for _, e := range a.Endpoints {
		carried[e] = true
	}
	required := map[string]bool{}
	for _, e := range soak.RequiredEndpoints() {
		required[e] = true
	}
	for _, endpoint := range added {
		if !carried[endpoint] {
			t.Errorf("%s succeeded in every cycle of the window and the attestation does not carry it; "+
				"probing an endpoint has to be enough to attest it", endpoint)
		}
		if required[endpoint] {
			t.Errorf("%s is in RequiredEndpoints, so this test no longer proves anything about "+
				"endpoints that are probed but not required", endpoint)
		}
	}
}

// TestConditionalEndpointsStayOutOfTheRefusalCatalogs is the latch on a100
// tasks 0.10 (a)'s whole point.
//
// Three lists refuse things: soak.RequiredEndpoints makes `soak attest` refuse,
// soak.LiveOnlyEndpoints makes BuildAttestation refuse a supervised proof, and
// engine.RequiredEndpoints makes the engine refuse to start. Probing an endpoint
// puts it into an attestation on its own — BuildAttestation carries every
// endpoint that succeeded — so widening any of the three here would make the
// refusal arrive before the evidence does.
//
// a100 widens them, in the deploy that also carries the engine that needs them.
// This step only produces the evidence.
func TestConditionalEndpointsStayOutOfTheRefusalCatalogs(t *testing.T) {
	added := map[string]bool{
		soak.EndpointConditionalOrders: true,
		soak.EndpointConditionalByID:   true,
		soak.EndpointSellableQuantity:  true,
	}
	for _, e := range soak.RequiredEndpoints() {
		if added[e] {
			t.Errorf("%s entered soak.RequiredEndpoints; a probe that has not succeeded yet would now "+
				"make `soak attest` refuse, including the re-issuance a100 does not depend on", e)
		}
	}
	for _, e := range soak.LiveOnlyEndpoints() {
		if added[e] {
			t.Errorf("%s entered soak.LiveOnlyEndpoints, which is the borrowable-mutation list; these "+
				"three are reads and reads are never borrowed from supervised evidence", e)
		}
	}
}
