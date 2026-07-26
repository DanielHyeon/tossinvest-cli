package execgw_test

// amend_indoubt_status10_test.go drives the cancel/amend resolution procedures
// with the OrderStatus values the API actually returns (extend-execution-contract
// task 5.1). amend_indoubt_test.go covers the same procedures with the legacy
// OPEN/CLOSED group labels; both must hold, because the derivation accepts both.

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// TestCancelInDoubtWithDocumentedStatuses is the cancel table over the ten-value
// enum. The row that matters most is PENDING_CANCEL: the broker is still
// processing a cancel, so the procedure must not conclude "the cancel never
// landed" — that would settle a live mutation as FAILED_CONFIRMED.
func TestCancelInDoubtWithDocumentedStatuses(t *testing.T) {
	cases := []struct {
		name string
		// order is the payload GET /api/v1/orders/{orderId} answers with.
		order string
		want  journal.AttemptState
		// wantDetail, when set, pins which branch produced the outcome — two
		// different readings can both end UNRESOLVED for opposite reasons.
		wantDetail string
	}{
		{
			name: "CANCELED with nothing filled",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CANCELED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "0"),
			want: journal.StateConfirmed,
		},
		{
			name: "CANCELED after a partial fill",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CANCELED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "1"),
			want: journal.StateConfirmed,
		},
		{
			name: "FILLED — the cancel lost the race",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "FILLED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "2"),
			want: journal.StateFailedConfirmed,
		},
		{
			name: "REJECTED — the broker refused the order itself",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "REJECTED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "0"),
			want: journal.StateFailedConfirmed,
		},
		{
			name: "PENDING throughout — the cancel never landed",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "PENDING", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "0"),
			want: journal.StateFailedConfirmed,
		},
		{
			name: "PENDING_CANCEL throughout — in flight is not an answer",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "PENDING_CANCEL", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "0"),
			want:       journal.StateUnresolvedInDoubt,
			wantDetail: "CANCEL_PENDING",
		},
		{
			name: "CANCEL_REJECTED record — shape unmeasured, so it blocks",
			order: cancellableOrderJSON("O-1", "005930", "BUY", "CANCEL_REJECTED", "2", "70000",
				"2026-03-30T10:20:00+09:00", "", "0"),
			want:       journal.StateUnresolvedInDoubt,
			wantDetail: "cancel_rejected_record",
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
			if tc.wantDetail != "" && !strings.Contains(res.Detail, tc.wantDetail) {
				t.Errorf("detail %q does not mention %q, so a different branch produced it",
					res.Detail, tc.wantDetail)
			}
		})
	}
}

// TestAmendInDoubtReplacedOriginalFindsSuccessor is the shape a real amend
// resolution meets: the original reports REPLACED and the journal has no lineage
// edge yet, because writing that edge is what this procedure does when it
// succeeds. The derivation fails closed on that row (nothing names the order
// carrying the remainder), and the procedure has to read that as "go find the
// successor" rather than "unresolvable" — otherwise every amend resolution parks.
func TestAmendInDoubtReplacedOriginalFindsSuccessor(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)

	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "REPLACED", "2", "70000",
		"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "1"))
	f.orders.set("OPEN", []string{
		orderJSON("O-2", "005930", "BUY", "PENDING", "2", "70500", "2026-03-30T10:30:02+09:00"),
	})

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

	edges, err := f.journal.LineageChildren(context.Background(), "O-1")
	if err != nil {
		t.Fatalf("LineageChildren: %v", err)
	}
	if len(edges) != 1 || edges[0].ChildOrderID != "O-2" {
		t.Fatalf("lineage edges from O-1: %+v", edges)
	}
	if edges[0].ParentFilledQuantity != "1" {
		t.Errorf("parent filled quantity: got %q, want 1", edges[0].ParentFilledQuantity)
	}
}

// TestAmendInDoubtReplacedOriginalWithoutSuccessorParks: the same REPLACED row
// with nothing to attribute the remainder to stays fail-closed.
func TestAmendInDoubtReplacedOriginalWithoutSuccessorParks(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)

	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "REPLACED", "2", "70000",
		"2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00", "1"))

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestAmendInDoubtPendingReplaceIsNotEvidence: PENDING_REPLACE is very likely
// this amend being processed. Counting it as "still open and unreplaced" would
// settle the attempt as FAILED_CONFIRMED while the broker is mid-replace.
func TestAmendInDoubtPendingReplaceIsNotEvidence(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)

	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "PENDING_REPLACE", "2", "70000",
		"2026-03-30T10:20:00+09:00", "", "0"))

	res, err := resolveAsync(t, f.resolverWith(reader), f.clk, attemptID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if res.State == journal.StateFailedConfirmed {
		t.Fatalf("a replace in flight must not be settled as FAILED_CONFIRMED (%s)", res.Detail)
	}
	if res.State != journal.StateUnresolvedInDoubt {
		t.Errorf("state: got %s, want UNRESOLVED_IN_DOUBT (%s)", res.State, res.Detail)
	}
}

// TestAmendInDoubtRejectedOriginalIsFailed: the broker refused the original, so
// there was never anything to amend.
func TestAmendInDoubtRejectedOriginalIsFailed(t *testing.T) {
	f := newDoubtFixture(t)
	attemptID := f.amendInDoubt(t)

	reader := newOrderReader()
	reader.set("O-1", cancellableOrderJSON("O-1", "005930", "BUY", "REJECTED", "2", "70000",
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
