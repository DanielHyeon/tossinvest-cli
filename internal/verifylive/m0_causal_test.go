package verifylive

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestM0CausalReceiptUsesOnlyRunSaltedIDTags(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	receipt, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer receipt.Close()
	lease, err := receipt.AcquireRunLease()
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err := lease.RecordCausal("parent-child-id", m0CausalFieldsV1{
		ParentRequestTag: receipt.tag("parent-request-opaque"), ParentResponseTag: receipt.tag("parent-response-opaque"),
		PendingClientTag: receipt.tag("pending-client-opaque"), ParentClientTag: receipt.tag("parent-client-opaque"),
		ParentChildTag: receipt.tag("parent-child-opaque"), ChildCheckpointTag: receipt.tag("child-checkpoint-opaque"),
		Symbol: "005930", RequestedMarket: "KR", Market: "KR", Type: "SINGLE", OrderType: "MARKET", Quantity: "1", Side: "SELL",
		RootStatus: "TRIGGERED", FirstStatus: "TRIGGERED", Condition: "STOP", Trigger: "69900", Expiry: "2026-08-02",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := lease.RecordCausal("child-first-observed-fill", m0CausalFieldsV1{
		ChildRequestTag: receipt.tag("child-request-opaque"), ChildResponseTag: receipt.tag("child-response-opaque"),
		Symbol: "005930", RequestedMarket: "KR", Market: "KR", OrderType: "MARKET", Quantity: "1", Side: "SELL", ChildStatus: "FILLED", Currency: "KRW",
		FilledQuantity: "1", AverageFilledPrice: "70000", FilledAmount: "70000", FilledAt: "",
	}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, receipt.RunID()+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rawID := range []string{"parent-request-opaque", "parent-response-opaque", "pending-client-opaque", "parent-client-opaque", "parent-child-opaque", "child-checkpoint-opaque", "child-request-opaque", "child-response-opaque"} {
		if strings.Contains(string(body), rawID) {
			t.Fatalf("causal receipt leaked raw opaque id %q: %s", rawID, body)
		}
	}
	for _, required := range []string{`"extracted_v1"`, `"schema":1`, `"parent_request_tag"`, `"parent_response_tag"`, `"pending_client_tag"`, `"parent_client_tag"`, `"parent_child_tag"`, `"child_checkpoint_tag"`, `"child_request_tag"`, `"child_response_tag"`, `"root_status":"TRIGGERED"`, `"first_status":"TRIGGERED"`, `"condition_type":"STOP"`, `"requested_market":"KR"`, `"child_status":"FILLED"`, `"currency":"KRW"`, `"filled_quantity":"1"`, `"average_filled_price":"70000"`, `"filled_amount":"70000"`} {
		if !strings.Contains(string(body), required) {
			t.Fatalf("causal receipt missing %s: %s", required, body)
		}
	}
}

func TestM0TriggerCannotPassWithoutTracedOfficialReadBoundary(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(1, 1, 1)
		f.suppressAttemptTrace = true
	})
	seedM0TriggerPrerequisites(t, h)
	if _, err := h.run(triggerOptions(t, DefaultTriggerWindow)); err == nil || !strings.Contains(err.Error(), "official-client raw read source") {
		t.Fatalf("arbitrary Broker error = %v, want pre-run official-client rejection", err)
	}
}

func TestM0ForgedBrokerCannotMintOfficialTransportEvidence(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(1, 1, 1)
		// The fake has no exported delivery function to call. Suppressing its
		// httptest official.Client leaves it as the arbitrary direct Broker a
		// production caller could supply to verifylive.New.
		f.suppressAttemptTrace = true
	})
	seedM0TriggerPrerequisites(t, h)
	if _, err := h.run(triggerOptions(t, DefaultTriggerWindow)); err == nil || !strings.Contains(err.Error(), "official-client raw read source") {
		t.Fatalf("forged broker error = %v, want pre-run official-client rejection", err)
	}
}

func TestM0TriggerHoldsOnEitherExtractedIdentityGroupMismatch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		script func(*fakeBroker)
		want   string
	}{
		{name: "parent", script: func(f *fakeBroker) { f.rawParentOrderType = "LIMIT" }, want: "parent identity group"},
		{name: "child", script: func(f *fakeBroker) { f.childIdentitySymbol = "MISMATCH" }, want: "child identity group"},
		{name: "parent-raw-decimal", script: func(f *fakeBroker) { f.rawParentTrigger = "69900.0" }, want: "parent identity group"},
		{name: "child-raw-decimal", script: func(f *fakeBroker) { f.childIdentityQty = "1.0" }, want: "child identity group"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) {
				f.firesOnRead(1, 1, 1)
				tc.script(f)
			})
			seedM0TriggerPrerequisites(t, h)
			_, err := h.run(triggerOptions(t, DefaultTriggerWindow))
			if err == nil {
				t.Fatal("identity mismatch completed M0 trigger")
			}
			var reason string
			for _, entry := range h.entries() {
				if entry.StepID == StepConditionalTrigger {
					reason = entry.Reason
				}
			}
			if !strings.Contains(reason, tc.want) {
				t.Fatalf("failure = %q, want %q", reason, tc.want)
			}
		})
	}
}

func TestM0CriticalTransportGapIsIrreversibleAfterRetrySuccess(t *testing.T) {
	for _, status := range []int{401, 429} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) {
				f.firesOnRead(1, 1, 1)
				f.m0ChildStatusOnce = status
			})
			seedM0TriggerPrerequisites(t, h)
			_, err := h.run(triggerOptions(t, DefaultTriggerWindow))
			if err == nil {
				t.Fatalf("status %d followed by success produced an M0 pass", status)
			}
			var reason string
			for _, entry := range h.entries() {
				if entry.StepID == StepConditionalTrigger {
					reason = entry.Reason
				}
			}
			if !strings.Contains(reason, "irreversible read gap") {
				t.Fatalf("status %d failure = %q, want irreversible gap", status, reason)
			}
		})
	}
}

func TestM0InvalidSuccessEnvelopeIsAnIrreversibleHold(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(1, 1, 1)
		f.invalidEnvelopeOnce = true
	})
	seedM0TriggerPrerequisites(t, h)
	_, err := h.run(triggerOptions(t, DefaultTriggerWindow))
	if err == nil {
		t.Fatal("invalid 2xx envelope completed M0 trigger")
	}
	var reason string
	for _, entry := range h.entries() {
		if entry.StepID == StepConditionalTrigger {
			reason = entry.Reason
		}
	}
	if !strings.Contains(reason, "invalid 2xx envelope") {
		t.Fatalf("failure = %q, want invalid-envelope HOLD", reason)
	}
}

func TestM0TriggeredWithoutChildIDIsManualReconcileOnly(t *testing.T) {
	h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 0, 0) })
	seedM0TriggerPrerequisites(t, h)
	_, err := h.run(triggerOptions(t, 2*time.Minute))
	if err == nil {
		t.Fatal("triggered-without-child run reported a clean result")
	}
	entries := h.entries()
	if targets := PendingCleanup(entries); len(targets) != 0 {
		t.Fatalf("PendingCleanup = %+v, want manual parent reconciliation", targets)
	}
	if targets := AbortTargets(entries); len(targets) != 0 {
		t.Fatalf("AbortTargets = %+v, want no automatic parent cancellation", targets)
	}
	runner := h.runner(t, Options{})
	result, abortErr := runner.Abort(context.Background(), "")
	if abortErr != nil || len(result.Targets) != 0 {
		t.Fatalf("Abort = %+v err=%v, want zero M0 mutation", result, abortErr)
	}
	if n := h.broker.countRequests("DELETE /conditional-orders/") + h.broker.countRequests("POST /orders/"); n != 0 {
		t.Fatalf("automatic M0 reconcile sent %d cancel request(s)", n)
	}
}

func TestM0ImportsStayMeasurementOnlyAndMutationMethodsStayExactSix(t *testing.T) {
	for _, path := range []string{"receipt.go", "m0_recovery.go", "m0_manual.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"internal/app", "internal/engine", "internal/protection", "internal/attestation", "internal/journal", "internal/trading"} {
			if strings.Contains(string(source), forbidden) {
				t.Fatalf("M0 source %s imports/wires forbidden %s", path, forbidden)
			}
		}
	}
	want := []string{"PlaceOrder", "CancelOrder", "ModifyOrder", "CreateConditionalOrder", "ModifyConditionalOrderRef", "CancelConditionalOrder"}
	got := MutationMethods()
	if len(got) != len(want) {
		t.Fatalf("MutationMethods=%v, want exact six", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("MutationMethods[%d]=%q, want %q", i, got[i], want[i])
		}
	}
}
