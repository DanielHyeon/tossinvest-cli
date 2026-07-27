package soak_test

// soak_test.go covers the survey loop: what one cycle reads, what it records and
// what it refuses to conclude.
//
// The runner is driven through soak.Reads — an interface with no mutating method
// on it — so these tests cannot reach a broker even by accident. The HTTP-level
// evidence (that the wired-up tool issues nothing but GETs) lives in
// cmd/tossctl/soak_test.go, against an httptest server.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

var soakStart = time.Date(2026, 7, 26, 9, 0, 0, 0, time.UTC)

// --- a stub broker ----------------------------------------------------------

// stubReads answers the read surface from a script. Every field is plain data:
// the point of the type is that there is nowhere to put a mutation.
type stubReads struct {
	mu sync.Mutex

	accounts []string
	holdings int
	prices   int

	// pages maps status -> cursor -> the page returned for it. The empty cursor
	// is the first page.
	pages map[string]map[string]soak.OrderPage

	// errs maps a soak endpoint constant to the error that probe returns.
	errs map[string]error

	// requireStatus makes the order list behave like the real broker: `status` is
	// a required query parameter, and a request without one is refused rather
	// than answered with "everything".
	requireStatus bool

	// failStatus refuses the walk of one status group and answers the other, so a
	// test can ask what half a list is worth.
	failStatus string

	ordersRead    []string
	statusesAsked []string
	symbolsAsked  []string
	pageRequests  int
}

// errStatusRequired is what GET /api/v1/orders answers when `status` is absent:
// HTTP 400, code invalid-request, field "status" (46 of 46 live soak cycles,
// requestId rkqNEGbfWwUq0jPx among them). openapi.latest.json marks the parameter
// required with enum {OPEN, CLOSED} and no "everything" member.
var errStatusRequired = errors.New(`official: 400 invalid-request {"field":"status"}`)

func newStubReads() *stubReads {
	return &stubReads{
		accounts: []string{"123-45-678901"},
		holdings: 2,
		prices:   1,
		pages: map[string]map[string]soak.OrderPage{
			"CLOSED": {"": {IDs: []string{"o2"}}},
			"OPEN":   {"": {IDs: []string{"o1"}}},
		},
		errs: map[string]error{},
	}
}

func (s *stubReads) fail(endpoint string, err error) *stubReads {
	s.errs[endpoint] = err
	return s
}

func (s *stubReads) Accounts(context.Context) ([]string, error) {
	if err := s.errs[soak.EndpointAccounts]; err != nil {
		return nil, err
	}
	return append([]string(nil), s.accounts...), nil
}

func (s *stubReads) BuyingPower(_ context.Context, _ string) error {
	return s.errs[soak.EndpointBuyingPower]
}

func (s *stubReads) Holdings(context.Context) (int, error) {
	if err := s.errs[soak.EndpointHoldings]; err != nil {
		return 0, err
	}
	return s.holdings, nil
}

func (s *stubReads) OrdersPage(_ context.Context, status, cursor string) (soak.OrderPage, error) {
	s.mu.Lock()
	s.pageRequests++
	s.statusesAsked = append(s.statusesAsked, status)
	s.mu.Unlock()
	if s.requireStatus && strings.TrimSpace(status) == "" {
		return soak.OrderPage{}, errStatusRequired
	}
	if s.failStatus != "" && s.failStatus == status {
		return soak.OrderPage{}, errors.New("official: the connection was reset")
	}
	if err := s.errs[soak.EndpointOrders]; err != nil {
		return soak.OrderPage{}, err
	}
	byCursor, ok := s.pages[status]
	if !ok {
		return soak.OrderPage{}, nil
	}
	return byCursor[cursor], nil
}

func (s *stubReads) Order(_ context.Context, id string) error {
	s.mu.Lock()
	s.ordersRead = append(s.ordersRead, id)
	s.mu.Unlock()
	return s.errs[soak.EndpointOrderByID]
}

func (s *stubReads) Prices(_ context.Context, symbols []string) (int, error) {
	s.mu.Lock()
	s.symbolsAsked = append(s.symbolsAsked, symbols...)
	s.mu.Unlock()
	if err := s.errs[soak.EndpointPrices]; err != nil {
		return 0, err
	}
	return s.prices, nil
}

// --- helpers ----------------------------------------------------------------

func newRunner(t *testing.T, reads soak.Reads, mutate func(*soak.Options)) (*soak.Runner, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(soakStart)
	opts := soak.Options{
		Reads:    reads,
		Clock:    fake,
		Symbols:  []string{"005930"},
		Currency: "KRW",
	}
	if mutate != nil {
		mutate(&opts)
	}
	r, err := soak.New(opts)
	if err != nil {
		t.Fatalf("soak.New: %v", err)
	}
	return r, fake
}

func endpointResult(t *testing.T, c soak.Cycle, endpoint string) soak.EndpointResult {
	t.Helper()
	for _, e := range c.Endpoints {
		if e.Endpoint == endpoint {
			return e
		}
	}
	t.Fatalf("cycle recorded no result for %s (recorded: %v)", endpoint, c.Endpoints)
	return soak.EndpointResult{}
}

// --- tests ------------------------------------------------------------------

// TestNewRejectsAnOptionsSetItCannotSurvey. A runner without a read surface
// would record a clean sheet of failures and call that a soak.
func TestNewRejectsAnOptionsSetItCannotSurvey(t *testing.T) {
	if _, err := soak.New(soak.Options{Clock: clock.NewFake(soakStart)}); err == nil {
		t.Fatal("soak.New accepted Options with no Reads")
	}
}

// TestRunCycleSurveysEveryRequiredEndpoint is the shape of one pass: each of the
// endpoints the attestation will claim gets exercised and recorded.
func TestRunCycleSurveysEveryRequiredEndpoint(t *testing.T) {
	r, _ := newRunner(t, newStubReads(), nil)

	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	for _, want := range soak.RequiredEndpoints() {
		got := endpointResult(t, c, want)
		if !got.OK {
			t.Errorf("%s: OK = false (%s / %s)", want, got.Class, got.Error)
		}
	}
	if c.AccountRef != "123-45-678901" {
		t.Errorf("AccountRef = %q, want the broker's account number", c.AccountRef)
	}
	if !c.Credential.OK {
		t.Error("Credential.OK = false although every read succeeded")
	}
	if !c.Completeness.Evaluated || !c.Completeness.OK {
		t.Errorf("Completeness = %+v, want an evaluated pass", c.Completeness)
	}
	if c.Kind != "cycle" || c.FormatVersion != soak.RecordFormatVersion {
		t.Errorf("record header = (%q, %d), want (cycle, %d)", c.Kind, c.FormatVersion, soak.RecordFormatVersion)
	}
}

// TestRunCycleSkipsOrderByIDWhenThereIsNoOrderToRead. Reading order zero is not
// a thing that can be done, and inventing a success for it would put an
// unverified endpoint into an attestation.
func TestRunCycleSkipsOrderByIDWhenThereIsNoOrderToRead(t *testing.T) {
	stub := newStubReads()
	stub.pages = map[string]map[string]soak.OrderPage{"CLOSED": {"": {}}, "OPEN": {"": {}}}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	got := endpointResult(t, c, soak.EndpointOrderByID)
	if !got.Skipped {
		t.Fatalf("%s: Skipped = false, want a skip", soak.EndpointOrderByID)
	}
	if got.OK {
		t.Error("a skipped endpoint must not be recorded as a success")
	}
	if strings.TrimSpace(got.SkipReason) == "" {
		t.Error("a skip must say why")
	}
	if len(stub.ordersRead) != 0 {
		t.Errorf("read orders %v although the list was empty", stub.ordersRead)
	}
}

// TestRunCycleAsksTheOrderListForAStatusItActuallyAccepts is the regression for
// the defect that made 46 consecutive live cycles unattestable.
//
// The survey used to open with an unfiltered walk — OrdersPage(ctx, "", "") — on
// the assumption that "no status" meant "every order". It does not: `status` is
// a required parameter with enum {OPEN, CLOSED}, and the real broker answers a
// request without one with 400 invalid-request. Every cycle therefore recorded
// GET /api/v1/orders as a failure, so the endpoint could never reach the success
// rate an attestation requires, so no attestation could ever be issued.
func TestRunCycleAsksTheOrderListForAStatusItActuallyAccepts(t *testing.T) {
	stub := newStubReads()
	stub.requireStatus = true
	stub.pages = map[string]map[string]soak.OrderPage{
		"CLOSED": {"": {IDs: []string{"o2"}}},
		"OPEN":   {"": {IDs: []string{"o1"}}},
	}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	got := endpointResult(t, c, soak.EndpointOrders)
	if !got.OK {
		t.Fatalf("%s failed against a broker that requires status: %s / %s",
			soak.EndpointOrders, got.Class, got.Error)
	}
	for _, status := range stub.statusesAsked {
		if status != "CLOSED" && status != "OPEN" {
			t.Errorf("the walk asked for status %q; the broker's enum is {OPEN, CLOSED} and has no member for \"everything\"",
				status)
		}
	}
	if !asked(stub.statusesAsked, "CLOSED") || !asked(stub.statusesAsked, "OPEN") {
		t.Errorf("statuses asked = %v, want both groups walked — their union is the whole list", stub.statusesAsked)
	}
	if c.Completeness.OrderIDs != 2 {
		t.Errorf("OrderIDs = %d, want the 2 orders the two groups hold between them", c.Completeness.OrderIDs)
	}
	if !c.Completeness.Evaluated || !c.Completeness.OK {
		t.Errorf("Completeness = %+v, want an evaluated pass", c.Completeness)
	}
}

func asked(statuses []string, want string) bool {
	for _, s := range statuses {
		if s == want {
			return true
		}
	}
	return false
}

// TestRunCycleWalksEveryOrderPage: the completeness claim is about pagination,
// so the walk has to actually follow the cursor.
func TestRunCycleWalksEveryOrderPage(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{
		"":   {IDs: []string{"o1", "o2"}, NextCursor: "c1", HasNext: true},
		"c1": {IDs: []string{"o3"}, NextCursor: "c2", HasNext: true},
		"c2": {IDs: []string{"o4"}},
	}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if c.Completeness.OrderPages != 4 {
		t.Errorf("OrderPages = %d, want 4 (3 closed pages + the single open one)", c.Completeness.OrderPages)
	}
	if c.Completeness.OrderIDs != 5 {
		t.Errorf("OrderIDs = %d, want 5 (4 closed + 1 open)", c.Completeness.OrderIDs)
	}
	if !c.Completeness.OK {
		t.Errorf("a clean three-page walk failed the completeness check: %s", c.Completeness.Detail)
	}
	if got := endpointResult(t, c, soak.EndpointOrders); got.Requests < 4 {
		t.Errorf("Requests = %d, want at least 4 (3 closed pages + the open-order page)", got.Requests)
	}
}

// TestRunCycleWalksTheOpenGroupInOneRequestAndPagesTheClosedOne is the
// pagination contract of GET /api/v1/orders as the spec states it: OPEN ignores
// cursor and limit and returns every working order at once, CLOSED is the only
// group a cursor walks. The survey exists to observe that contract, so it has to
// be able to tell the two behaviours apart in what it records.
func TestRunCycleWalksTheOpenGroupInOneRequestAndPagesTheClosedOne(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{
		"":   {IDs: []string{"c-1", "c-2"}, NextCursor: "p2", HasNext: true},
		"p2": {IDs: []string{"c-3"}},
	}
	// The open group answers the same page whatever cursor it is handed, because
	// it does not read the cursor at all — and it never claims a next page, which
	// is what stops the walk after one request.
	stub.pages["OPEN"] = map[string]soak.OrderPage{"": {IDs: []string{"o-1", "o-2"}}}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if got := countAsked(stub.statusesAsked, "OPEN"); got != 1 {
		t.Errorf("the open group was requested %d time(s), want exactly 1 — it is not paginated", got)
	}
	if got := countAsked(stub.statusesAsked, "CLOSED"); got != 2 {
		t.Errorf("the closed group was requested %d time(s), want 2 — its cursor has to be followed", got)
	}
	if c.Completeness.OrderIDs != 5 {
		t.Errorf("OrderIDs = %d, want 5 (3 closed + 2 open)", c.Completeness.OrderIDs)
	}
	if c.Completeness.OpenOrders != 2 {
		t.Errorf("OpenOrders = %d, want 2", c.Completeness.OpenOrders)
	}
	if c.Completeness.OrdersInBothStatuses != 0 {
		t.Errorf("OrdersInBothStatuses = %d, want 0 — no order was in both groups", c.Completeness.OrdersInBothStatuses)
	}
	if !c.Completeness.OK {
		t.Errorf("a clean two-group walk failed the completeness check: %s", c.Completeness.Detail)
	}
}

func countAsked(statuses []string, want string) int {
	n := 0
	for _, s := range statuses {
		if s == want {
			n++
		}
	}
	return n
}

// TestRunCycleDetectsACursorLoop. A cursor that points at itself is an infinite
// walk; a soak that hung on it would report nothing at all.
func TestRunCycleDetectsACursorLoop(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{
		"":   {IDs: []string{"o1"}, NextCursor: "c1", HasNext: true},
		"c1": {IDs: []string{"o2"}, NextCursor: "c1", HasNext: true},
	}

	r, _ := newRunner(t, stub, nil)
	c, err := r.RunCycle(context.Background())
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if !c.Completeness.CursorLoop {
		t.Error("CursorLoop = false on a self-referential cursor")
	}
	if c.Completeness.OK {
		t.Error("a cursor loop passed the completeness check")
	}
}

// TestRunCycleDetectsTruncatedPagination: hasNext with no cursor means the rest
// of the list is unreachable.
func TestRunCycleDetectsTruncatedPagination(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{
		"": {IDs: []string{"o1"}, HasNext: true, NextCursor: ""},
	}

	r, _ := newRunner(t, stub, nil)
	c, _ := r.RunCycle(context.Background())

	if !c.Completeness.TruncatedPagination {
		t.Error("TruncatedPagination = false although hasNext was true with no cursor")
	}
	if c.Completeness.OK {
		t.Error("truncated pagination passed the completeness check")
	}
}

// TestRunCycleDetectsDuplicateOrderIDs. The same order on two pages of one
// status group means the walk cannot be trusted to have seen every order exactly
// once.
func TestRunCycleDetectsDuplicateOrderIDs(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{
		"":   {IDs: []string{"o1", "o2"}, NextCursor: "c1", HasNext: true},
		"c1": {IDs: []string{"o2", "o3"}},
	}

	r, _ := newRunner(t, stub, nil)
	c, _ := r.RunCycle(context.Background())

	if c.Completeness.DuplicateOrderIDs != 1 {
		t.Errorf("DuplicateOrderIDs = %d, want 1", c.Completeness.DuplicateOrderIDs)
	}
	if !strings.Contains(c.Completeness.Detail, "o2") {
		t.Errorf("Detail = %q, want it to name the repeated order", c.Completeness.Detail)
	}
	if c.Completeness.OK {
		t.Error("a duplicated order id passed the completeness check")
	}
}

// TestRunCycleDetectsADuplicateInsideTheOpenGroupToo. The open group is the one
// the engine reconciles its exposure against; a repeat inside it is at least as
// serious as a repeat in the history, and the old single-list check never looked.
func TestRunCycleDetectsADuplicateInsideTheOpenGroup(t *testing.T) {
	stub := newStubReads()
	stub.pages["OPEN"] = map[string]soak.OrderPage{
		"":   {IDs: []string{"o1"}, NextCursor: "c1", HasNext: true},
		"c1": {IDs: []string{"o1"}},
	}

	r, _ := newRunner(t, stub, nil)
	c, _ := r.RunCycle(context.Background())

	if c.Completeness.DuplicateOrderIDs != 1 {
		t.Errorf("DuplicateOrderIDs = %d, want 1", c.Completeness.DuplicateOrderIDs)
	}
	if c.Completeness.OK {
		t.Error("an order repeated inside the open group passed the completeness check")
	}
}

// TestRunCycleTreatsAnOrderInBothGroupsAsOneOrderNotADuplicate.
//
// The two groups are labels over the per-order status and they overlap: the spec
// puts PARTIAL_FILLED in both OPEN and CLOSED. If the union were merged before
// duplicates were counted, every partially filled order on the account would read
// as a corrupt list and would block the attestation for as long as it stayed
// partially filled — the same shape of permanent block this walk was fixed for.
func TestRunCycleTreatsAnOrderInBothGroupsAsOneOrderNotADuplicate(t *testing.T) {
	stub := newStubReads()
	stub.requireStatus = true
	stub.pages["CLOSED"] = map[string]soak.OrderPage{"": {IDs: []string{"filled-1", "partial-1"}}}
	stub.pages["OPEN"] = map[string]soak.OrderPage{"": {IDs: []string{"partial-1", "pending-1"}}}

	r, _ := newRunner(t, stub, nil)
	c, _ := r.RunCycle(context.Background())

	if c.Completeness.DuplicateOrderIDs != 0 {
		t.Errorf("DuplicateOrderIDs = %d, want 0 — the OPEN/CLOSED overlap is the broker's contract, not a fault (%s)",
			c.Completeness.DuplicateOrderIDs, c.Completeness.Detail)
	}
	if c.Completeness.OrdersInBothStatuses != 1 {
		t.Errorf("OrdersInBothStatuses = %d, want 1", c.Completeness.OrdersInBothStatuses)
	}
	if c.Completeness.OrderIDs != 4 {
		t.Errorf("OrderIDs = %d, want 4 identifiers over the two walks (3 distinct orders)", c.Completeness.OrderIDs)
	}
	if !c.Completeness.OK {
		t.Errorf("an account with a partial fill failed the completeness check: %s", c.Completeness.Detail)
	}
}

// TestRunCycleFlagsAnOrderReturnedWithNoIdentifier is what became of the
// open-order coverage check.
//
// Coverage asked whether every open order was in the unfiltered list; with the
// list defined as CLOSED ∪ OPEN that question cannot fail any more. What it could
// really catch is an entry nothing can name — cmd/tossctl's adapter surfaces one
// as an empty id rather than dropping the page — and an order the engine cannot
// name is an order it can never reconcile.
func TestRunCycleFlagsAnOrderReturnedWithNoIdentifier(t *testing.T) {
	stub := newStubReads()
	stub.pages["CLOSED"] = map[string]soak.OrderPage{"": {IDs: []string{"c1", ""}}}
	stub.pages["OPEN"] = map[string]soak.OrderPage{"": {IDs: []string{"o1", ""}}}

	r, _ := newRunner(t, stub, nil)
	c, _ := r.RunCycle(context.Background())

	if c.Completeness.BlankOrderIDs != 2 {
		t.Errorf("BlankOrderIDs = %d, want 2 (one in each group)", c.Completeness.BlankOrderIDs)
	}
	if c.Completeness.OK {
		t.Error("an order with no identifier on it passed the completeness check")
	}
	if !strings.Contains(c.Completeness.Detail, "no identifier") {
		t.Errorf("Detail = %q, want it to say what was wrong", c.Completeness.Detail)
	}
	// Two nameless orders are two problems of one kind, not one order that turned
	// up twice and not an order the broker put in both groups.
	if c.Completeness.DuplicateOrderIDs != 0 {
		t.Errorf("DuplicateOrderIDs = %d, want 0 — a blank id is not a repeated order",
			c.Completeness.DuplicateOrderIDs)
	}
	if c.Completeness.OrdersInBothStatuses != 0 {
		t.Errorf("OrdersInBothStatuses = %d, want 0 — a blank id is not an order in both groups",
			c.Completeness.OrdersInBothStatuses)
	}
}

// TestRunCycleTakesTheOrderByIDProbeFromEitherGroup. The by-id read has to be
// exercised on any account that has an order at all, whichever group holds it —
// a skipped probe is an endpoint Evaluate will not attest.
func TestRunCycleTakesTheOrderByIDProbeFromEitherGroup(t *testing.T) {
	for _, tc := range []struct {
		name    string
		closed  []string
		open    []string
		wantID  string
		wantRun bool
	}{
		{name: "closed only", closed: []string{"c-1"}, wantID: "c-1", wantRun: true},
		{name: "open only", open: []string{"o-1"}, wantID: "o-1", wantRun: true},
		{name: "both", closed: []string{"c-1"}, open: []string{"o-1"}, wantID: "c-1", wantRun: true},
		{name: "neither"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := newStubReads()
			stub.requireStatus = true
			stub.pages = map[string]map[string]soak.OrderPage{
				"CLOSED": {"": {IDs: tc.closed}},
				"OPEN":   {"": {IDs: tc.open}},
			}

			r, _ := newRunner(t, stub, nil)
			c, _ := r.RunCycle(context.Background())

			got := endpointResult(t, c, soak.EndpointOrderByID)
			if !tc.wantRun {
				if !got.Skipped {
					t.Fatalf("%s ran against an account with no orders", soak.EndpointOrderByID)
				}
				return
			}
			if got.Skipped {
				t.Fatalf("%s was skipped although the union held %v / %v: %s",
					soak.EndpointOrderByID, tc.closed, tc.open, got.SkipReason)
			}
			if len(stub.ordersRead) != 1 || stub.ordersRead[0] != tc.wantID {
				t.Errorf("read orders %v, want [%s]", stub.ordersRead, tc.wantID)
			}
		})
	}
}

// TestRunCycleLeavesCompletenessUnevaluatedWhenAWalkFails, for either group. A
// half-read list cannot support a claim about the whole list, and inventing a
// pass from the group that did answer is exactly the failure mode that would let
// a broken read endpoint into an attestation.
func TestRunCycleLeavesCompletenessUnevaluatedWhenAWalkFails(t *testing.T) {
	for _, status := range []string{"CLOSED", "OPEN"} {
		t.Run(status, func(t *testing.T) {
			stub := newStubReads()
			stub.failStatus = status
			stub.pages["CLOSED"] = map[string]soak.OrderPage{"": {IDs: []string{"c-1"}}}
			stub.pages["OPEN"] = map[string]soak.OrderPage{"": {IDs: []string{"o-1"}}}

			r, _ := newRunner(t, stub, nil)
			c, _ := r.RunCycle(context.Background())

			if got := endpointResult(t, c, soak.EndpointOrders); got.OK {
				t.Errorf("%s was recorded OK although the %s walk failed", soak.EndpointOrders, status)
			}
			if c.Completeness.Evaluated {
				t.Error("Completeness.Evaluated = true although a walk did not finish")
			}
			if c.Completeness.OK {
				t.Error("Completeness.OK = true although a walk did not finish")
			}
			if !strings.Contains(c.Completeness.Detail, "not evaluated") {
				t.Errorf("Detail = %q, want it to say the list could not be read", c.Completeness.Detail)
			}
		})
	}
}

// TestRunCycleFlagsAQuoteTheBrokerDidNotReturn.
func TestRunCycleFlagsAQuoteTheBrokerDidNotReturn(t *testing.T) {
	stub := newStubReads()
	stub.prices = 1

	r, _ := newRunner(t, stub, func(o *soak.Options) {
		o.Symbols = []string{"005930", "000660"}
	})
	c, _ := r.RunCycle(context.Background())

	if c.Completeness.QuotesRequested != 2 || c.Completeness.QuotesReturned != 1 {
		t.Errorf("quotes = %d/%d, want 1/2", c.Completeness.QuotesReturned, c.Completeness.QuotesRequested)
	}
	if c.Completeness.OK {
		t.Error("a short quote response passed the completeness check")
	}
}

// TestRunCycleClassifiesAFailure. The class is what separates "the broker
// throttled us" from "our credentials stopped working"; the streak rule depends
// on telling them apart.
func TestRunCycleClassifiesAFailure(t *testing.T) {
	throttled := errors.New("official: rate limited")
	stub := newStubReads().fail(soak.EndpointPrices, throttled)

	r, _ := newRunner(t, stub, func(o *soak.Options) {
		o.Classify = func(err error) soak.Class {
			if errors.Is(err, throttled) {
				return soak.ClassRateLimited
			}
			return soak.ClassOther
		}
	})
	c, _ := r.RunCycle(context.Background())

	got := endpointResult(t, c, soak.EndpointPrices)
	if got.OK {
		t.Fatal("a failed read was recorded as a success")
	}
	if got.Class != soak.ClassRateLimited {
		t.Errorf("Class = %q, want %q", got.Class, soak.ClassRateLimited)
	}
	if !c.Credential.OK {
		t.Error("a throttled quote read must not be reported as a credential failure")
	}
}

// TestRunCycleReportsAnAuthFailureAgainstTheCredentials. This is the signal the
// three-day streak is built on: if the unattended refresh stops working, the day
// must not count.
func TestRunCycleReportsAnAuthFailureAgainstTheCredentials(t *testing.T) {
	denied := errors.New("official: authentication failed")
	stub := newStubReads().fail(soak.EndpointAccounts, denied)

	r, _ := newRunner(t, stub, func(o *soak.Options) {
		o.Classify = func(error) soak.Class { return soak.ClassAuth }
	})
	c, _ := r.RunCycle(context.Background())

	if c.Credential.OK {
		t.Error("Credential.OK = true although the authenticated read was refused")
	}
	if c.Credential.Class != soak.ClassAuth {
		t.Errorf("Credential.Class = %q, want %q", c.Credential.Class, soak.ClassAuth)
	}
}

// TestRunObservesAnUnattendedTokenRefresh. An access token whose expiry moved
// forward between cycles is the only direct evidence that the credentials
// renewed themselves with nobody watching.
func TestRunObservesAnUnattendedTokenRefresh(t *testing.T) {
	expiry := soakStart.Add(time.Hour)
	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.TokenExpiry = func() (time.Time, bool) { return expiry, true }
	})

	first, _ := r.RunCycle(context.Background())
	if first.Credential.Refreshed {
		t.Error("the first observation cannot be a refresh; there is nothing to compare it to")
	}
	if !first.Credential.Observed {
		t.Error("Credential.Observed = false although the expiry was readable")
	}

	expiry = expiry.Add(2 * time.Hour)
	second, _ := r.RunCycle(context.Background())
	if !second.Credential.Refreshed {
		t.Error("an expiry that moved forward was not recorded as a refresh")
	}

	third, _ := r.RunCycle(context.Background())
	if third.Credential.Refreshed {
		t.Error("an unchanged expiry was recorded as a refresh")
	}
}

// TestRunRecordsEveryCycleAndStopsAtTheLimit.
func TestRunRecordsEveryCycleAndStopsAtTheLimit(t *testing.T) {
	rec, err := soak.OpenRecorder(t.TempDir() + "/soak.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	r, _ := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 3
		o.Interval = 0
	})
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	cycles, err := soak.LoadCycles(rec.Path())
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 3 {
		t.Fatalf("recorded %d cycles, want 3", len(cycles))
	}
}

// TestRunSleepsTheIntervalBetweenCycles, driven by the fake clock so the test
// does not spend the interval it is asserting about.
func TestRunSleepsTheIntervalBetweenCycles(t *testing.T) {
	rec, err := soak.OpenRecorder(t.TempDir() + "/soak.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	r, fake := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 2
		o.Interval = 6 * time.Hour
	})

	done := make(chan error, 1)
	go func() { done <- r.Run(context.Background()) }()

	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the runner never parked in Sleep; it is not honouring --interval")
	}
	fake.Advance(6 * time.Hour)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not finish after the interval elapsed")
	}

	cycles, err := soak.LoadCycles(rec.Path())
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 2 {
		t.Fatalf("recorded %d cycles, want 2", len(cycles))
	}
	if !cycles[1].StartedAt.Equal(soakStart.Add(6 * time.Hour)) {
		t.Errorf("second cycle at %s, want the interval to have elapsed first", cycles[1].StartedAt)
	}
}

// TestRunStopsWhenTheContextIsCancelled: a soak is a long-running process and
// Ctrl-C has to end it without losing what it already wrote.
func TestRunStopsWhenTheContextIsCancelled(t *testing.T) {
	rec, err := soak.OpenRecorder(t.TempDir() + "/soak.jsonl")
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	r, fake := newRunner(t, newStubReads(), func(o *soak.Options) {
		o.Recorder = rec
		o.Cycles = 0 // run until stopped
		o.Interval = time.Hour
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()

	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the runner never reached its sleep")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run ignored the cancellation")
	}

	cycles, err := soak.LoadCycles(rec.Path())
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("recorded %d cycles, want the one that completed before the cancellation", len(cycles))
	}
}
