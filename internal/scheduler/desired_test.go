package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testManifestClaim struct {
	binding   ActivationBinding
	expiresAt time.Time
	revokedAt time.Time
}

func (m testManifestClaim) verifyActivation(_ context.Context, binding ActivationBinding, now time.Time) error {
	if !m.revokedAt.IsZero() {
		return ErrManifestRevoked
	}
	if m.expiresAt.IsZero() || !now.Before(m.expiresAt) {
		return ErrManifestExpired
	}
	if m.binding != binding {
		return ErrManifestMismatch
	}
	return nil
}

func approvedDesired() DesiredState {
	return DesiredState{
		Version: SchedulerVersion, Enabled: true, AutoStart: true,
		Market: MarketScopeKR, Session: SessionRegular,
		Actor: "local-operator", ApprovedAt: atNoTest("2026-07-31T00:00:00Z"),
		CalendarVersion: "calendar-v1", ConfigVersion: "config-v7",
	}
}

func atNoTest(raw string) time.Time {
	got, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		panic(err)
	}
	return got.UTC()
}

func TestMissingDesiredStateUsesClosedDefaults(t *testing.T) {
	store := NewDesiredStore(filepath.Join(t.TempDir(), "scheduler.json"))
	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultDesiredState()
	if got != want {
		t.Fatalf("default = %+v, want %+v", got, want)
	}
	if got.Enabled || got.AutoStart || got.Market != MarketScopeNone || got.Session != SessionRegular {
		t.Fatalf("unsafe default: %+v", got)
	}
}

func TestDesiredStateRoundTripsActorApprovalMarketAndVersions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.json")
	store := NewDesiredStore(path)
	want := approvedDesired()
	if err := store.Save(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load(context.Background())
	want.Revision = 1
	if err != nil || got != want {
		t.Fatalf("round trip = %+v err=%v, want %+v", got, err, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestDesiredStateRevisionCASPreservesCommittedOperatorOff(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.json")
	first := NewDesiredStore(path)
	second := NewDesiredStore(path)
	if err := first.Save(context.Background(), approvedDesired()); err != nil {
		t.Fatal(err)
	}
	stale, err := first.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	off := stale
	off.Enabled = false
	off.AutoStart = false
	if err := second.Save(context.Background(), off); err != nil {
		t.Fatalf("saving operator OFF: %v", err)
	}
	if err := first.Save(context.Background(), stale); !errors.Is(err, ErrDesiredRevisionConflict) {
		t.Fatalf("stale ON save error = %v, want revision conflict", err)
	}
	got, err := first.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled || got.AutoStart || got.Revision != 2 {
		t.Fatalf("committed operator OFF was overwritten: %+v", got)
	}
}

func TestDesiredStoreHonorsCanceledContext(t *testing.T) {
	store := NewDesiredStore(filepath.Join(t.TempDir(), "scheduler.json"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Load(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Load error = %v, want canceled", err)
	}
	if err := store.Save(ctx, approvedDesired()); !errors.Is(err, context.Canceled) {
		t.Fatalf("Save error = %v, want canceled", err)
	}
}

func TestDesiredRevisionCannotWrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.json")
	desired := approvedDesired()
	desired.Revision = math.MaxUint64
	raw, err := json.Marshal(desired)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := NewDesiredStore(path).Save(context.Background(), desired); err == nil {
		t.Fatal("maximum desired revision wrapped")
	}
}

func TestEnabledDesiredStateCannotInventApproval(t *testing.T) {
	store := NewDesiredStore(filepath.Join(t.TempDir(), "scheduler.json"))
	for _, mutate := range []func(*DesiredState){
		func(d *DesiredState) { d.Actor = "" },
		func(d *DesiredState) { d.ApprovedAt = time.Time{} },
		func(d *DesiredState) { d.Market = MarketScopeNone },
		func(d *DesiredState) { d.ConfigVersion = "" },
		func(d *DesiredState) { d.CalendarVersion = "" },
	} {
		d := approvedDesired()
		mutate(&d)
		if err := store.Save(context.Background(), d); err == nil {
			t.Fatalf("invalid approval was saved: %+v", d)
		}
	}
}

func TestDesiredStateRejectsCombinedMarketWithoutPerMarketBindings(t *testing.T) {
	d := approvedDesired()
	d.Market = MarketScope("KR+US")
	if err := NewDesiredStore(filepath.Join(t.TempDir(), "scheduler.json")).Save(context.Background(), d); err == nil {
		t.Fatal("combined market was accepted without exact per-market calendar bindings")
	}
}

func TestAutoResumeRequiresExactUnexpiredUnrevokedManifest(t *testing.T) {
	now := atNoTest("2026-08-01T01:00:00Z")
	desired := approvedDesired()
	current := CurrentBinding{
		SchedulerVersion: SchedulerVersion, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion,
		BuildDigest: "build-v1",
	}
	binding := desired.ActivationBinding(current.BuildDigest)
	claim := testManifestClaim{binding: binding, expiresAt: now.Add(time.Hour)}
	got := Restore(context.Background(), desired, current, claim, now)
	if !got.Restored || got.Reason != ResumeExactManifest || got.Activation == nil {
		t.Fatalf("exact restore = %+v", got)
	}

	cases := []struct {
		name string
		edit func(*testManifestClaim, *CurrentBinding)
		want ResumeReason
	}{
		{"scheduler mismatch", func(m *testManifestClaim, _ *CurrentBinding) { m.binding.SchedulerVersion = "other" }, ResumeManifestMismatch},
		{"calendar mismatch", func(m *testManifestClaim, _ *CurrentBinding) { m.binding.CalendarVersion = "other" }, ResumeManifestMismatch},
		{"config mismatch", func(_ *testManifestClaim, c *CurrentBinding) { c.ConfigVersion = "other" }, ResumeDesiredMismatch},
		{"expired", func(m *testManifestClaim, _ *CurrentBinding) { m.expiresAt = now }, ResumeManifestExpired},
		{"revoked", func(m *testManifestClaim, _ *CurrentBinding) { m.revokedAt = now.Add(-time.Minute) }, ResumeManifestRevoked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, c := claim, current
			tc.edit(&m, &c)
			got := Restore(context.Background(), desired, c, m, now)
			if got.Restored || got.Reason != tc.want {
				t.Fatalf("restore = %+v, want refusal %q", got, tc.want)
			}
		})
	}
}

func TestDesiredStateLoadRejectsUnknownDuplicateAndTrailingJSON(t *testing.T) {
	for _, raw := range []string{
		`{"version":"scheduler-v1","market":"none","session":"regular","unknown":true}`,
		`{"version":"scheduler-v1","market":"none","market":"KR","session":"regular"}`,
		`{"version":"scheduler-v1","market":"none","session":"regular"}{}`,
	} {
		path := filepath.Join(t.TempDir(), "scheduler.json")
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := NewDesiredStore(path).Load(context.Background()); err == nil {
			t.Fatalf("unsafe JSON accepted: %s", raw)
		}
	}
}

func TestRestoreRejectsFutureApproval(t *testing.T) {
	now := atNoTest("2026-08-01T01:00:00Z")
	desired := approvedDesired()
	desired.ApprovedAt = now.Add(time.Minute)
	current := CurrentBinding{
		SchedulerVersion: desired.Version, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion, BuildDigest: "build-v1",
	}
	claim := testManifestClaim{binding: desired.ActivationBinding(current.BuildDigest), expiresAt: now.Add(time.Hour)}
	got := Restore(context.Background(), desired, current, claim, now)
	if got.Restored || got.Reason != ResumeDesiredMismatch || got.Activation != nil {
		t.Fatalf("future approval restored: %+v", got)
	}
}

func TestDesiredStoreRejectsFutureApprovalOnSaveAndLoad(t *testing.T) {
	desired := approvedDesired()
	desired.ApprovedAt = time.Now().UTC().Add(time.Hour)
	path := filepath.Join(t.TempDir(), "scheduler.json")
	store := NewDesiredStore(path)
	if err := store.Save(context.Background(), desired); err == nil {
		t.Fatal("future approval saved")
	}
	raw, err := json.Marshal(desired)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("future approval loaded")
	}
}

func TestRestoreHonorsCanceledContext(t *testing.T) {
	desired := approvedDesired()
	current := CurrentBinding{SchedulerVersion: desired.Version, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion, BuildDigest: "build-v1"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := Restore(ctx, desired, current, testManifestClaim{
		binding: desired.ActivationBinding(current.BuildDigest), expiresAt: time.Now().Add(time.Hour),
	}, time.Now())
	if got.Restored || got.Activation != nil || got.Reason != ResumeVerificationFailed || !errors.Is(got.Err, context.Canceled) {
		t.Fatalf("canceled restore = %+v", got)
	}
}

func TestAutoResumeWithoutA047VerifierFailsClosed(t *testing.T) {
	desired := approvedDesired()
	current := CurrentBinding{
		SchedulerVersion: SchedulerVersion, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion,
		BuildDigest: "build-v1",
	}
	got := Restore(context.Background(), desired, current, nil, atNoTest("2026-08-01T01:00:00Z"))
	if got.Restored || got.Reason != ResumeManifestUnavailable || !errors.Is(got.Err, ErrManifestUnavailable) {
		t.Fatalf("nil verifier restore = %+v", got)
	}
}
