package verifylive

import (
	"fmt"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func m0PendingFixture() M0Checkpoint {
	return M0Checkpoint{Kind: "pending-create", ClientOrderID: "TRIGGER-recovery", Symbol: "005930", Market: "KR",
		Type: "SINGLE", Quantity: "1", Side: "SELL", OrderType: "MARKET", ConditionType: "STOP", Trigger: "69900", ExpireDate: "2026-08-02"}
}

func appendM0Pending(t *testing.T, h *harness, checkpoint M0Checkpoint) {
	t.Helper()
	recorder, err := OpenRecorder(h.record)
	if err != nil {
		t.Fatal(err)
	}
	if err := recorder.Append(Entry{Kind: KindM0Checkpoint, M0Checkpoint: &checkpoint}); err != nil {
		_ = recorder.Close()
		t.Fatal(err)
	}
	if err := recorder.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestM0PendingRecoveryZeroOneMultipleHoldWithoutMutation(t *testing.T) {
	pending := m0PendingFixture()
	match := officialRawForPending(pending, "co-recovered")
	for _, tc := range []struct {
		name    string
		orders  []official.RawConditionalOrder
		wantOne bool
		dup     bool
	}{
		{name: "zero"},
		{name: "one", orders: []official.RawConditionalOrder{match}, wantOne: true},
		{name: "one-deduped-across-open-and-closed", orders: []official.RawConditionalOrder{match}, wantOne: true, dup: true},
		{name: "multiple", orders: []official.RawConditionalOrder{match, officialRawForPending(pending, "co-second")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) { f.recoveryOrders, f.recoveryDuplicateAcrossStatus = tc.orders, tc.dup })
			appendM0Pending(t, h, pending)
			summary, err := h.run(triggerOptions(t, DefaultTriggerWindow))
			if err == nil || !summary.Halted || !strings.Contains(summary.Halt, "M0 causal receipt HOLD") {
				t.Fatalf("recovery = summary=%+v err=%v, want terminal HOLD", summary, err)
			}
			for _, request := range h.broker.requests {
				if strings.HasPrefix(request, "POST /conditional-orders") || strings.HasPrefix(request, "DELETE ") {
					t.Fatalf("recovery made a mutation: %s", request)
				}
			}
			entries := h.entries()
			gotOne := false
			for _, entry := range entries {
				if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "parent-created" {
					gotOne = true
				}
			}
			if gotOne != tc.wantOne {
				t.Fatalf("parent-created checkpoint = %v, want %v", gotOne, tc.wantOne)
			}
		})
	}
}

func TestM0MultiplePendingOwnersHoldBeforeRecoveryReadOrMutation(t *testing.T) {
	h := triggerHarness(t, nil)
	first := m0PendingFixture()
	second := first
	second.ClientOrderID = "TRIGGER-second"
	appendM0Pending(t, h, first)
	appendM0Pending(t, h, second)
	if _, err := h.run(triggerOptions(t, DefaultTriggerWindow)); err == nil || !strings.Contains(err.Error(), "multiple unresolved pending") {
		t.Fatalf("multiple pending error = %v, want pre-factory ambiguity HOLD", err)
	}
	for _, request := range h.broker.requests {
		if strings.Contains(request, "/conditional-orders") {
			t.Fatalf("ambiguous pending made broker request: %s", request)
		}
	}
}

func TestM0RestartWithDurableParentOrChildOwnerNeverReentersTriggerMutation(t *testing.T) {
	for _, checkpoint := range []M0Checkpoint{
		func() M0Checkpoint {
			c := m0PendingFixture()
			c.Kind, c.ParentConditionalID = "parent-created", "parent-crash"
			return c
		}(),
		func() M0Checkpoint {
			c := m0PendingFixture()
			c.Kind, c.ParentConditionalID, c.ChildOrderID = "child-observed", "parent-child", "child-crash"
			return c
		}(),
	} {
		t.Run(checkpoint.Kind, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 1, 1) })
			seedM0TriggerPrerequisites(t, h)
			appendM0Pending(t, h, checkpoint)
			if _, err := h.run(triggerOptions(t, DefaultTriggerWindow)); err == nil || !strings.Contains(err.Error(), "unsettled "+checkpoint.Kind+" owner") {
				t.Fatalf("unsettled %s did not fail before broker work: %v", checkpoint.Kind, err)
			}
			if got := h.broker.countRequests("POST /conditional-orders"); got != 0 {
				t.Fatalf("unsettled %s made %d conditional POST(s)", checkpoint.Kind, got)
			}
			if got := h.broker.countRequests("GET /orders/child-crash"); got != 0 {
				t.Fatalf("unsettled %s reread child %d time(s)", checkpoint.Kind, got)
			}
			entries := h.entries()
			if got := PendingCleanup(entries); len(got) != 0 {
				t.Fatalf("PendingCleanup=%+v, want manual-only owner", got)
			}
			if got := AbortTargets(entries); len(got) != 0 {
				t.Fatalf("AbortTargets=%+v, want manual-only owner", got)
			}
		})
	}
}

func TestM0RecoveryRejectsRawFieldMismatchAndBrokenPagination(t *testing.T) {
	pending := m0PendingFixture()
	match := officialRawForPending(pending, "co-recovered")
	for _, tc := range []struct {
		name   string
		orders []official.RawConditionalOrder
		script func(string, string) (official.RawConditionalOrderList, error)
	}{
		{name: "raw-order-type-mismatch", orders: []official.RawConditionalOrder{func() official.RawConditionalOrder { bad := match; bad.OrderType = "LIMIT"; return bad }()}},
		{name: "empty-next-cursor", script: func(status, cursor string) (official.RawConditionalOrderList, error) {
			if status == "OPEN" {
				return official.RawConditionalOrderList{HasNext: true}, nil
			}
			return official.RawConditionalOrderList{}, nil
		}},
		{name: "repeated-cursor", script: func(status, cursor string) (official.RawConditionalOrderList, error) {
			if status != "OPEN" {
				return official.RawConditionalOrderList{}, nil
			}
			return official.RawConditionalOrderList{HasNext: true, NextCursor: "repeat"}, nil
		}},
		{name: "page-cap", script: func(status, cursor string) (official.RawConditionalOrderList, error) {
			if status != "OPEN" {
				return official.RawConditionalOrderList{}, nil
			}
			return official.RawConditionalOrderList{HasNext: true, NextCursor: fmt.Sprintf("page-%s", cursor)}, nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) { f.recoveryOrders, f.recoveryPageScript = tc.orders, tc.script })
			appendM0Pending(t, h, pending)
			_, err := h.run(triggerOptions(t, DefaultTriggerWindow))
			if err == nil || !strings.Contains(err.Error(), "M0 causal receipt HOLD") {
				t.Fatalf("recovery err=%v, want HOLD", err)
			}
			if n := h.broker.countRequests("POST /conditional-orders"); n != 0 {
				t.Fatalf("recovery mismatch/pagination made %d POST(s)", n)
			}
		})
	}
}

func officialRawForPending(p M0Checkpoint, id string) official.RawConditionalOrder {
	return official.RawConditionalOrder{ID: id, ClientOrderID: p.ClientOrderID, Symbol: p.Symbol, Market: p.Market,
		Type: "SINGLE", Status: "WATCHING", FirstStatus: "WATCHING", OrderType: "MARKET", OrderSide: "SELL",
		Quantity: "1", TriggerPrice: p.Trigger, ConditionType: "STOP", ExpireDate: p.ExpireDate}
}
