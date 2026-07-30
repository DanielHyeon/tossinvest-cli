package verifylive

// endpoints_test.go pins what a verification record is allowed to prove.
//
// The record is the only evidence in this system that a mutating broker call
// works: the soak that proves everything else contains no mutation transport at
// all. So the rule that decides what counts as proof has to be stated on its own,
// where a change to it fails a test rather than quietly widening what an
// automation gate will start on.

import (
	"testing"
	"time"
)

func at(t *testing.T, h, m int) time.Time {
	t.Helper()
	return time.Date(2026, 7, 27, h, m, 0, 0, time.UTC)
}

func evidenceRecord(t *testing.T) []Entry {
	t.Helper()
	return []Entry{{
		Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictPass,
		AccountRef: "*******5921",
		Calls: []Call{
			{Endpoint: "POST /api/v1/orders", At: at(t, 12, 0)},
			{Endpoint: "POST /api/v1/orders/{id}/cancel", At: at(t, 12, 1)},
			{Endpoint: "GET /api/v1/orders", At: at(t, 12, 2)},
		},
	}}
}

func TestSucceededEndpointsTakesTheCallsThatWorked(t *testing.T) {
	ev := SucceededEndpoints(evidenceRecord(t), at(t, 13, 0), time.Hour*24)

	for _, want := range []string{
		"POST /api/v1/orders", "POST /api/v1/orders/{id}/cancel", "GET /api/v1/orders",
	} {
		if _, ok := ev.Endpoints[want]; !ok {
			t.Errorf("%s succeeded in this record but is not in the evidence: %+v", want, ev.Endpoints)
		}
	}
	if len(ev.AccountRefs) != 1 || ev.AccountRefs[0] != "*******5921" {
		t.Errorf("AccountRefs = %v, want exactly the one the record names", ev.AccountRefs)
	}
}

// TestSucceededEndpointsIgnoresACallThatFailed. A 422 that refused an order is
// not evidence that placing orders works — it is the opposite.
func TestSucceededEndpointsIgnoresACallThatFailed(t *testing.T) {
	entries := []Entry{{
		Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictFail, AccountRef: "*******5921",
		Calls: []Call{{
			Endpoint: "POST /api/v1/orders", At: at(t, 12, 0),
			Error: "official: API error 422: order-hours-closed", ErrorCode: "order-hours-closed",
		}},
	}}

	if _, ok := SucceededEndpoints(entries, at(t, 13, 0), 24*time.Hour).Endpoints["POST /api/v1/orders"]; ok {
		t.Fatal("a refused call was taken as proof that the endpoint works")
	}
}

// TestSucceededEndpointsKeepsTheNewestSuccess: the age check downstream is about
// "when was this last known to work", so the newest success is the one that
// answers it.
func TestSucceededEndpointsKeepsTheNewestSuccess(t *testing.T) {
	entries := []Entry{{
		Kind: KindStep, StepID: StepOrderCancel, Verdict: VerdictPass, AccountRef: "*******5921",
		Calls: []Call{
			{Endpoint: "POST /api/v1/orders", At: at(t, 9, 0)},
			{Endpoint: "POST /api/v1/orders", At: at(t, 12, 0)},
			{Endpoint: "POST /api/v1/orders", At: at(t, 10, 0)},
		},
	}}

	got := SucceededEndpoints(entries, at(t, 13, 0), 24*time.Hour).Endpoints["POST /api/v1/orders"]
	if !got.Equal(at(t, 12, 0)) {
		t.Fatalf("newest success = %s, want %s", got, at(t, 12, 0))
	}
}

// TestSucceededEndpointsDropsEvidenceOlderThanTheWindow.
func TestSucceededEndpointsDropsEvidenceOlderThanTheWindow(t *testing.T) {
	entries := evidenceRecord(t)
	now := at(t, 12, 0).Add(48 * time.Hour)

	if got := SucceededEndpoints(entries, now, 24*time.Hour); len(got.Endpoints) != 0 {
		t.Fatalf("evidence two days old was kept inside a one-day window: %+v", got.Endpoints)
	}
}

// TestSucceededEndpointsWindowBoundary states the boundary rather than leaving it
// to whoever reads the comparison next: exactly maxAge old is already too old.
func TestSucceededEndpointsWindowBoundary(t *testing.T) {
	entries := evidenceRecord(t)
	success := at(t, 12, 0)

	inside := SucceededEndpoints(entries, success.Add(24*time.Hour-time.Nanosecond), 24*time.Hour)
	if _, ok := inside.Endpoints["POST /api/v1/orders"]; !ok {
		t.Error("a success one nanosecond inside the window was dropped")
	}
	on := SucceededEndpoints(entries, success.Add(24*time.Hour), 24*time.Hour)
	if _, ok := on.Endpoints["POST /api/v1/orders"]; ok {
		t.Error("a success exactly maxAge old was kept; the boundary is exclusive")
	}
}

// TestSucceededEndpointsRefusesEvidenceFromTheFuture: a call stamped after now is
// a clock that cannot be trusted, and the safe reading of an untrustworthy clock
// is "this proves nothing".
func TestSucceededEndpointsRefusesEvidenceFromTheFuture(t *testing.T) {
	entries := evidenceRecord(t)

	if got := SucceededEndpoints(entries, at(t, 11, 0), 24*time.Hour); len(got.Endpoints) != 0 {
		t.Fatalf("evidence stamped in the future was accepted: %+v", got.Endpoints)
	}
}

// TestSucceededEndpointsReportsEveryAccountItSaw. One record naming two accounts
// is a misconfiguration the caller has to refuse, so it must be able to see it.
func TestSucceededEndpointsReportsEveryAccountItSaw(t *testing.T) {
	entries := append(evidenceRecord(t), Entry{
		Kind: KindStep, StepID: StepOrderAmend, Verdict: VerdictPass, AccountRef: "*******0000",
		Calls: []Call{{Endpoint: "POST /api/v1/orders/{id}/modify", At: at(t, 12, 5)}},
	})

	if got := SucceededEndpoints(entries, at(t, 13, 0), 24*time.Hour).AccountRefs; len(got) != 2 {
		t.Fatalf("AccountRefs = %v, want both accounts so the caller can refuse", got)
	}
}

func TestSucceededEndpointsOnAnEmptyRecord(t *testing.T) {
	got := SucceededEndpoints(nil, at(t, 13, 0), 24*time.Hour)
	if len(got.Endpoints) != 0 || len(got.AccountRefs) != 0 {
		t.Fatalf("an empty record proved something: %+v", got)
	}
}
