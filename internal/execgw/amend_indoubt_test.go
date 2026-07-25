package execgw_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// CANCEL/AMEND IN_DOUBT resolution tests (harden-execution-base task 2.8).
//
// A cancel or an amend in doubt is harder than a place in doubt: the official API
// answers a modify with a *new* order number, so "did it happen" is answered by
// looking at the original order's state and, for an amend, by finding the
// successor that carries the remainder.

// fakeOrderReader answers OrderByID-style single-order reads.
type fakeOrderReader struct {
	mu     sync.Mutex
	orders map[string]string
	err    error
}

func newOrderReader() *fakeOrderReader {
	return &fakeOrderReader{orders: map[string]string{}}
}

func (r *fakeOrderReader) set(id, body string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[id] = body
}

func (r *fakeOrderReader) OrderRaw(_ context.Context, orderID string) (json.RawMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	body, ok := r.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order %s not found", orderID)
	}
	return json.RawMessage(body), nil
}

// cancellableOrderJSON builds an order payload with the fields the derivation
// reads: canceledAt and the cumulative filled quantity.
func cancellableOrderJSON(id, symbol, side, status, qty, price, orderedAt, canceledAt, filled string) string {
	canceled := "null"
	if canceledAt != "" {
		canceled = fmt.Sprintf("%q", canceledAt)
	}
	return fmt.Sprintf(`{"orderId":%q,"symbol":%q,"side":%q,"status":%q,"quantity":%q,`+
		`"price":%q,"currency":"KRW","orderedAt":%q,"canceledAt":%s,`+
		`"execution":{"filledQuantity":%q}}`,
		id, symbol, side, status, qty, price, orderedAt, canceled, filled)
}

// cancelInDoubt drives a cancel of O-1 into IN_DOUBT and returns the attempt id.
func (f *doubtFixture) cancelInDoubt(t *testing.T) string {
	t.Helper()
	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	out, err := f.gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   intent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: goodDecision(t, execgw.CancelHash(intent), f.clk),
	})
	if err == nil {
		t.Fatal("the fixture broker must fail so the cancel lands IN_DOUBT")
	}
	if out.State != journal.StateInDoubt {
		t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
	}
	return out.AttemptID
}

// amendInDoubt drives an amend of O-1 (price → 70500) into IN_DOUBT.
func (f *doubtFixture) amendInDoubt(t *testing.T) string {
	t.Helper()
	newPrice := 70500.0
	intent := orderintent.AmendIntent{OrderID: "O-1", Price: &newPrice}
	out, err := f.gw.Amend(context.Background(), execgw.AmendRequest{
		Intent:   intent,
		Symbol:   "005930",
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: goodDecision(t, execgw.AmendHash(intent), f.clk),
	})
	if err == nil {
		t.Fatal("the fixture broker must fail so the amend lands IN_DOUBT")
	}
	if out.State != journal.StateInDoubt {
		t.Fatalf("state: got %s, want IN_DOUBT (%s)", out.State, out.Detail)
	}
	return out.AttemptID
}

func (f *doubtFixture) resolverWith(reader execgw.OrderReader) *execgw.Resolver {
	r := f.resolver()
	r.Order = reader
	return r
}

// --- cancel -----------------------------------------------------------------

func TestCancelInDoubtResolvesFromTheOriginalOrder(t *testing.T) {
	cases := []struct {
		name   string
		order  string
		want   journal.AttemptState
		reason execgw.ReasonCode
	}{
		{
			name: "cancelled with nothing filled",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "0"),
			want: journal.StateConfirmed,
		},
		{
			name: "cancelled after a partial fill",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "1"),
			want: journal.StateConfirmed,
		},
		{
			name: "already fully filled — the cancel lost the race",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "2"),
			want: journal.StateFailedConfirmed,
		},
		{
			name: "still open after stabilisation — the cancel never landed",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "OPEN", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "0"),
			want: journal.StateFailedConfirmed,
		},
		{
			name: "contradictory payload — closed with an unexplained remainder",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "1"),
			want:   journal.StateUnresolvedInDoubt,
			reason: execgw.ReasonUnresolvedInDoubt,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newDoubtFixture(t)
			attemptID := f.cancelInDoubt(t)
			reader := newOrderReader()
			reader.set("O-1", tc.order)

			res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if res.State != tc.want {
				t.Fatalf("state: got %s, want %s (%s)", res.State, tc.want, res.Detail)
			}
			if tc.reason != "" && res.Reason != tc.reason {
				t.Errorf("reason: got %q, want %q", res.Reason, tc.reason)
			}
		})
	}
}

// TestCancelInDoubtUnreadableOrderParks: if the original order cannot be read at
// all, nothing is proven and the attempt must not be closed either way.
func TestCancelInDoubtUnreadableOrderParks(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.cancelInDoubt(t)
	reader := newOrderReader() // O-1 is absent → every read errors

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// --- amend ------------------------------------------------------------------

// TestAmendInDoubtFindsSuccessorAndRecordsLineage is the spec's "부분체결 중 정정"
// scenario: the original closes carrying its fill, a successor carries the
// remainder, and both facts are committed together.
func TestAmendInDoubtFindsSuccessorAndRecordsLineage(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)

	reader := newOrderReader()
	// The original was replaced: closed, cancelled, one of two filled.
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
		"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "1"))
	// The successor carries the amended price, on the second page so the scan has
	// to walk the list to find it.
	f.orders.set("OPEN",
		[]string{orderJSON("O-noise", "005930", "SELL", "OPEN", "5", "71000", "2026-03-30T10:30:02+09:00")},
		[]string{orderJSON("O-2", "005930", "BUY", "OPEN", "2", "70500", "2026-03-30T10:30:02+09:00")},
	)

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", res.State, res.Detail)
	}
	if res.BrokerOrderID != "O-2" {
		t.Errorf("broker order id: got %q, want the successor O-2", res.BrokerOrderID)
	}

	ctx := context.Background()
	edges, err := f.journal.LineageChildren(ctx, "O-1")
	if err != nil {
		t.Fatalf("LineageChildren: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("lineage edges from O-1: got %d, want 1", len(edges))
	}
	edge := edges[0]
	if edge.ChildOrderID != "O-2" || edge.Relation != journal.RelationReplaces {
		t.Errorf("edge: %+v", edge)
	}
	if edge.ParentFilledQuantity != "1" {
		t.Errorf("parent filled quantity: got %q, want 1", edge.ParentFilledQuantity)
	}
	if edge.RequestedQuantity != "2" {
		t.Errorf("requested quantity: got %q, want 2", edge.RequestedQuantity)
	}
	if edge.AttemptID != attemptID {
		t.Errorf("edge attempt id: got %q, want %q", edge.AttemptID, attemptID)
	}

	// The settle and the edge are one transaction: the attempt is CONFIRMED and
	// the edge exists, or neither does.
	rec, err := f.journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateConfirmed || rec.BrokerOrderID != "O-2" {
		t.Errorf("journal attempt: state=%s brokerOrderID=%q", rec.State, rec.BrokerOrderID)
	}
}

// TestAmendInDoubtStillOpenIsFailed: the original is untouched at its original
// price after stabilisation, so the amend did not land.
func TestAmendInDoubtStillOpenIsFailed(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)
	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "OPEN", "2", "70000",
		"2026-03-30T10:20:00+09:00", "", "0"))

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateFailedConfirmed {
		t.Errorf("state: got %s, want FAILED_CONFIRMED (%s)", res.State, res.Detail)
	}
	if edges, _ := f.journal.LineageChildren(context.Background(), "O-1"); len(edges) != 0 {
		t.Errorf("a failed amend must record no lineage, got %d edge(s)", len(edges))
	}
}

// TestAmendInDoubtClosedWithoutSuccessorParks: the original is gone but nothing
// carries its remainder. That could be a replace whose successor we cannot see,
// or a cancellation — the difference decides whether the engine is exposed, so it
// is not something to assume.
func TestAmendInDoubtClosedWithoutSuccessorParks(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)
	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
		"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "0"))

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestAmendInDoubtAmbiguousSuccessorParks: two candidate successors means we
// cannot say which order carries the remainder.
func TestAmendInDoubtAmbiguousSuccessorParks(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)
	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "CLOSED", "2", "70000",
		"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "0"))
	f.orders.set("OPEN", []string{
		orderJSON("O-2", "005930", "BUY", "OPEN", "2", "70500", "2026-03-30T10:30:02+09:00"),
		orderJSON("O-3", "005930", "BUY", "OPEN", "2", "70500", "2026-03-30T10:30:03+09:00"),
	})

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestLineageChainResolvesToTheCurrentOrder is the spec's "다단계 정정 체인": after
// two amendments, the id the engine first knew must still lead to the live order.
func TestLineageChainResolvesToTheCurrentOrder(t *testing.T) {
	f := newDoubtFixture(t)
	ctx := context.Background()

	// Two replace edges, recorded the way an amend resolution records them.
	f.seedLineage(t, "O-1", "O-2", "1", "2")
	f.seedLineage(t, "O-2", "O-3", "0", "1")

	current, err := f.journal.ResolveCurrentOrderID(ctx, "O-1")
	if err != nil {
		t.Fatalf("ResolveCurrentOrderID: %v", err)
	}
	if current != "O-3" {
		t.Errorf("current order for O-1: got %q, want O-3", current)
	}
	if current, err := f.journal.ResolveCurrentOrderID(ctx, "O-3"); err != nil || current != "O-3" {
		t.Errorf("a leaf must resolve to itself: got %q (%v)", current, err)
	}
	if current, err := f.journal.ResolveCurrentOrderID(ctx, "O-unknown"); err != nil || current != "O-unknown" {
		t.Errorf("an unknown order must resolve to itself: got %q (%v)", current, err)
	}
}

// seedLineage records one replace edge through a resolved amend attempt, which is
// the only way an edge is ever written.
func (f *doubtFixture) seedLineage(t *testing.T, parent, child, parentFilled, requested string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := f.journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-" + parent + "-" + child, Market: "kr", TradingDay: "2026-03-30",
			AccountRef: "acct-7", Symbol: "005930", Side: "BUY", OrderType: "LIMIT",
			Quantity: requested, Price: "70500", Currency: "KRW", Source: "test",
			Fingerprint: "fp-" + parent + child,
		},
		Kind:          journal.KindAmend,
		AttemptID:     "attempt-" + parent + "-" + child,
		TargetOrderID: parent,
	})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkInDoubt(ctx, "test", "seeded"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	if err := attempt.ResolveConfirmedWithLineage(ctx, journal.LineageEdge{
		ParentOrderID:        parent,
		ChildOrderID:         child,
		Relation:             journal.RelationReplaces,
		ParentFilledQuantity: parentFilled,
		RequestedQuantity:    requested,
	}, journal.ReasonResolvedFound, "seeded by test"); err != nil {
		t.Fatalf("ResolveConfirmedWithLineage: %v", err)
	}
}

// TestLineageCycleIsRefused: a malformed chain must not hang the resolver.
func TestLineageCycleIsRefused(t *testing.T) {
	f := newDoubtFixture(t)
	f.seedLineage(t, "A", "B", "0", "1")
	f.seedLineage(t, "B", "A", "0", "1")

	if _, err := f.journal.ResolveCurrentOrderID(context.Background(), "A"); err == nil ||
		!strings.Contains(err.Error(), "cycle") {
		t.Errorf("a lineage cycle must be refused, got %v", err)
	}
}
