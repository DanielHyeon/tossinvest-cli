package engine

// a069 task 4: the engine-owned command side of a quarantine release.
//
// The claims worth pinning are the ones an operator's safety rests on: the
// grant is one-time, the danger delay is real, an unconfirmed apply changes
// nothing and costs nothing, the version is compared against the ledger and not
// against a browser's memory, and the release touches exactly one row.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// fakeQuarantineRepository is a policy repository that also holds quarantines,
// which is exactly the shape the production journal has.
type fakeQuarantineRepository struct {
	fakePolicyRepository
	rows       []exitquarantine.Row
	released   []releasedQuarantine
	listErr    error
	releaseErr error
	// otherWrites counts anything a release must never do. The production
	// repository interface cannot express those calls at all, so this exists to
	// document the claim rather than to catch a regression the compiler misses.
	otherWrites int
}

type releasedQuarantine struct {
	positionID string
	generation int64
	version    int64
	kind       string
	evidence   string
}

func (f *fakeQuarantineRepository) ActiveExitSnapshotQuarantines(context.Context) ([]exitquarantine.Row, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]exitquarantine.Row(nil), f.rows...), nil
}

func (f *fakeQuarantineRepository) ReleaseExitSnapshotQuarantine(_ context.Context, positionID string,
	generation, expectedVersion int64, kind, evidence string) error {
	if f.releaseErr != nil {
		return f.releaseErr
	}
	for i, row := range f.rows {
		if row.PositionID == positionID && row.Generation == generation && row.Version == expectedVersion {
			f.rows = append(f.rows[:i], f.rows[i+1:]...)
			f.released = append(f.released, releasedQuarantine{positionID, generation, expectedVersion, kind, evidence})
			return nil
		}
	}
	return journal.ErrExitSnapshotReleaseStale
}

func quarantineFixture(t *testing.T, now time.Time) (*PositionPolicyCommandService, *fakeQuarantineRepository, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(now)
	repo := &fakeQuarantineRepository{
		rows: []exitquarantine.Row{{
			PositionID: "pos-1", Market: "kr", Symbol: "466100", Generation: 1, Version: 1,
			Reason: "ambiguous_recovery", Evidence: "exitpolicy: recovery candidate identity mismatch",
			QuarantinedAt: "2026-08-03T09:03:40Z", Protection: "24929",
		}},
	}
	return &PositionPolicyCommandService{j: repo, clk: clk}, repo, clk
}

func releaseRequest() exitquarantine.Request {
	return exitquarantine.Request{
		PositionID: "pos-1", Generation: 1, Version: 1, Actor: exitquarantine.ActorLocalOperator,
	}
}

func TestThePreviewShowsTheProtectionARleaseWouldKeep(t *testing.T) {
	service, _, _ := quarantineFixture(t, time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))

	preview, err := service.PreviewQuarantineRelease(context.Background(), releaseRequest())
	if err != nil {
		t.Fatalf("PreviewQuarantineRelease: %v", err)
	}
	if preview.Capability == "" {
		t.Fatal("a preview without a grant cannot be approved")
	}
	if preview.WaitSeconds != 3 {
		t.Fatalf("wait = %ds, want the 3s danger delay", preview.WaitSeconds)
	}
	if preview.Row.Protection != "24929" || preview.Row.Reason != "ambiguous_recovery" {
		t.Fatalf("preview row = %+v, want the stored protection and reason", preview.Row)
	}
}

func TestAPositionWithNoActiveQuarantineCannotBeReleased(t *testing.T) {
	service, repo, _ := quarantineFixture(t, time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC))
	repo.rows = nil

	_, err := service.PreviewQuarantineRelease(context.Background(), releaseRequest())

	if !errors.Is(err, exitquarantine.ErrNotQuarantined) {
		t.Fatalf("err = %v, want ErrNotQuarantined", err)
	}
}

func TestAReleaseWithoutTheDangerAcknowledgementChangesNothingAndKeepsTheGrant(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	service, repo, clk := quarantineFixture(t, now)
	ctx := context.Background()

	preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(4 * time.Second)

	if _, err := service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: false,
	}); !errors.Is(err, exitquarantine.ErrConfirmationRequired) {
		t.Fatalf("err = %v, want ErrConfirmationRequired", err)
	}
	if len(repo.released) != 0 {
		t.Fatalf("an unconfirmed apply released something: %+v", repo.released)
	}
	// Recoverable, so the same preview must still be approvable.
	if _, err := service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	}); err != nil {
		t.Fatalf("the grant did not survive a recoverable refusal: %v", err)
	}
}

func TestAReleaseBeforeTheDangerDelayIsRefused(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	service, repo, clk := quarantineFixture(t, now)
	ctx := context.Background()

	preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(2 * time.Second)

	_, err = service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	})
	if !errors.Is(err, exitquarantine.ErrCapabilityTooEarly) {
		t.Fatalf("err = %v, want ErrCapabilityTooEarly", err)
	}
	if len(repo.released) != 0 {
		t.Fatalf("an early apply released something: %+v", repo.released)
	}
}

func TestAReleaseGrantIsOneTimeAndExpires(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	ctx := context.Background()

	t.Run("재사용 거부", func(t *testing.T) {
		service, _, clk := quarantineFixture(t, now)
		preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(4 * time.Second)
		apply := exitquarantine.ApplyRequest{Capability: preview.Capability, Confirmed: true}
		if _, err := service.ReleaseQuarantine(ctx, apply); err != nil {
			t.Fatalf("first release: %v", err)
		}
		if _, err := service.ReleaseQuarantine(ctx, apply); !errors.Is(err, exitquarantine.ErrCapabilityInvalid) {
			t.Fatalf("replay err = %v, want ErrCapabilityInvalid", err)
		}
	})

	t.Run("만료 거부", func(t *testing.T) {
		service, _, clk := quarantineFixture(t, now)
		preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(6 * time.Minute)
		if _, err := service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); !errors.Is(err, exitquarantine.ErrCapabilityExpired) {
			t.Fatalf("expiry err = %v, want ErrCapabilityExpired", err)
		}
	})

	t.Run("위조 거부", func(t *testing.T) {
		service, _, _ := quarantineFixture(t, now)
		if _, err := service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
			Capability: "not-a-real-grant", Confirmed: true,
		}); !errors.Is(err, exitquarantine.ErrCapabilityInvalid) {
			t.Fatalf("forged err = %v, want ErrCapabilityInvalid", err)
		}
	})
}

func TestAQuarantineRewrittenAfterThePreviewIsNotReleasedBlind(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	service, repo, clk := quarantineFixture(t, now)
	ctx := context.Background()

	preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	// The engine re-quarantined the same generation while the operator was
	// reading: same position, new version.
	repo.rows[0].Version = 2
	clk.Advance(4 * time.Second)

	_, err = service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	})
	if !errors.Is(err, exitquarantine.ErrVersionMismatch) {
		t.Fatalf("err = %v, want ErrVersionMismatch", err)
	}
	if len(repo.released) != 0 {
		t.Fatalf("a stale version released something: %+v", repo.released)
	}
}

func TestAReleaseWritesHumanRepairWithServerComposedEvidence(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	service, repo, clk := quarantineFixture(t, now)
	ctx := context.Background()

	preview, err := service.PreviewQuarantineRelease(ctx, releaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	clk.Advance(4 * time.Second)
	result, err := service.ReleaseQuarantine(ctx, exitquarantine.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	})
	if err != nil {
		t.Fatalf("ReleaseQuarantine: %v", err)
	}

	if len(repo.released) != 1 {
		t.Fatalf("released %d rows, want exactly 1", len(repo.released))
	}
	got := repo.released[0]
	if got.kind != journal.QuarantineReleaseHumanRepair {
		t.Fatalf("release kind = %q, want HUMAN_REPAIR", got.kind)
	}
	if got.positionID != "pos-1" || got.generation != 1 || got.version != 1 {
		t.Fatalf("released %+v, want the previewed row", got)
	}
	if got.evidence == "" {
		t.Fatal("the ledger refuses a blank evidence string")
	}
	if got.evidence != result.Evidence {
		t.Fatalf("returned evidence %q differs from what was written %q", result.Evidence, got.evidence)
	}
	// The operator never typed this: it is composed from what the server knows.
	for _, want := range []string{exitquarantine.ActorLocalOperator, "v1", "ambiguous_recovery"} {
		if !contains(got.evidence, want) {
			t.Fatalf("evidence %q is missing %q", got.evidence, want)
		}
	}
	if repo.otherWrites != 0 {
		t.Fatalf("a release wrote something other than the quarantine row: %d", repo.otherWrites)
	}
}

func TestAControlPlaneWithoutTheQuarantineCapabilitySaysSo(t *testing.T) {
	// A policy repository that predates a069: it has no quarantine methods at
	// all, which is what an older engine's handle looks like through this seam.
	service := &PositionPolicyCommandService{
		j:   &fakePolicyRepository{state: positionpolicy.State{PositionID: "pos-1"}},
		clk: clock.NewFake(time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)),
	}

	if _, err := service.Quarantines(context.Background()); !errors.Is(err, exitquarantine.ErrUnwired) {
		t.Fatalf("list err = %v, want ErrUnwired", err)
	}
	if _, err := service.PreviewQuarantineRelease(context.Background(), releaseRequest()); !errors.Is(err, exitquarantine.ErrUnwired) {
		t.Fatalf("preview err = %v, want ErrUnwired", err)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) &&
		func() bool {
			for i := 0; i+len(needle) <= len(haystack); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		}())
}
