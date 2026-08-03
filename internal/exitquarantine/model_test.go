package exitquarantine

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAReleaseRequestMustNameOneRealQuarantineVersion(t *testing.T) {
	base := Request{PositionID: "pos-1", Generation: 1, Version: 1, Actor: ActorLocalOperator}
	if err := base.Validate(); err != nil {
		t.Fatalf("a complete request must validate: %v", err)
	}

	for _, tc := range []struct {
		name    string
		mutate  func(*Request)
		wantSub string
	}{
		{"빈 position", func(r *Request) { r.PositionID = "  " }, "position id is required"},
		{"음수 generation", func(r *Request) { r.Generation = -1 }, "generation must not be negative"},
		{"version 0", func(r *Request) { r.Version = 0 }, "quarantine version must be positive"},
		{"음수 version", func(r *Request) { r.Version = -3 }, "quarantine version must be positive"},
		{"임의 actor", func(r *Request) { r.Actor = "SOMEBODY" }, "actor is not the local operator"},
		{"빈 actor", func(r *Request) { r.Actor = "" }, "actor is not the local operator"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			tc.mutate(&req)
			err := req.Validate()
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("want ErrInvalidRequest, got %v", err)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("want %q in %v", tc.wantSub, err)
			}
		})
	}
}

func TestTheServerComposesTheReleaseEvidence(t *testing.T) {
	at := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	row := Row{PositionID: "pos-1", Version: 2, Reason: "ambiguous_recovery"}

	got := ComposeEvidence(ActorLocalOperator, row, at)

	for _, want := range []string{
		"LOCAL_OPERATOR", "quarantine v2", "reason=ambiguous_recovery", "2026-08-04T12:00:00Z",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("evidence %q is missing %q", got, want)
		}
	}
}

func TestComposedEvidenceIsNeverEmptyEvenWithThinInput(t *testing.T) {
	// ReleaseExitSnapshotQuarantine refuses a blank evidence string, so the
	// composer must not be able to produce one. A row with no recorded reason is
	// exactly the case that would otherwise turn into a refusal at the ledger.
	got := ComposeEvidence("", Row{Version: 1}, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))

	if strings.TrimSpace(got) == "" {
		t.Fatal("composed evidence must never be blank")
	}
	if !strings.Contains(got, ActorLocalOperator) || !strings.Contains(got, "reason=unknown") {
		t.Fatalf("thin input should still name the actor and an explicit unknown reason: %q", got)
	}
}
