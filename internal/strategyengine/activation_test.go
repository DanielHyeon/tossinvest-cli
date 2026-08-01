package strategyengine

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func bindingFixture(now time.Time) ManifestBinding {
	return ManifestBinding{
		AccountRef: "acct", Profile: "prod", BuildDigest: "build", CommitDigest: "commit", LaneID: LaneID, LaneVersion: LaneVersion,
		LaneSourceDigest: FrozenSourceSetDigest, LaneConstantsDigest: constantsDigest(), ThresholdVersion: "threshold-v1", ThresholdSetDigest: "set", EvidenceDigest: "evidence", SettingsDigest: "settings",
		AttestationDigest: "attest", AttestationExpiresAt: now.Add(time.Hour), GuardianVersion: "guardian-v1", GuardianLimitsDigest: "limits", ReconciliationWatermark: "watermark",
		ProtectionProfile: "broker-stop-v1", ProtectionState: "WIRED", OperatingPolicy: "LIVE", SchedulerScope: "KR/regular", CalendarVersion: "calendar",
		LaneApproved: true, SchedulerApproved: true, AutoStartApproved: true, GateApproved: true, LiveApproved: true, Actor: "operator", AuditID: "audit", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour), Generation: 7,
	}
}
func installedRepository(binding ManifestBinding) *ManifestRepository {
	digest, err := manifestDigest(binding)
	if err != nil {
		panic(err)
	}
	return &ManifestRepository{current: activationManifest{binding: binding, digest: digest}}
}

func TestActivationManifestBindsEveryFieldAndFailsClosed(t *testing.T) {
	now := time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC)
	base := bindingFixture(now)
	wantDigest, _ := manifestDigest(base)
	if got, err := installedRepository(base).Verify(base, now); err != nil || got.Digest() != wantDigest {
		t.Fatalf("verify=(%+v,%v)", got, err)
	}
	tests := []struct {
		name   string
		mutate func(*ManifestBinding)
	}{
		{"account", func(v *ManifestBinding) { v.AccountRef = "other" }}, {"profile", func(v *ManifestBinding) { v.Profile = "other" }},
		{"build", func(v *ManifestBinding) { v.BuildDigest = "other" }}, {"commit", func(v *ManifestBinding) { v.CommitDigest = "other" }},
		{"lane", func(v *ManifestBinding) { v.LaneID = "other" }}, {"lane version", func(v *ManifestBinding) { v.LaneVersion = "other" }},
		{"source", func(v *ManifestBinding) { v.LaneSourceDigest = "other" }}, {"constants", func(v *ManifestBinding) { v.LaneConstantsDigest = "other" }},
		{"threshold version", func(v *ManifestBinding) { v.ThresholdVersion = "other" }}, {"threshold set", func(v *ManifestBinding) { v.ThresholdSetDigest = "other" }},
		{"evidence", func(v *ManifestBinding) { v.EvidenceDigest = "other" }}, {"settings", func(v *ManifestBinding) { v.SettingsDigest = "other" }},
		{"attestation", func(v *ManifestBinding) { v.AttestationDigest = "other" }}, {"attestation expiry", func(v *ManifestBinding) { v.AttestationExpiresAt = v.AttestationExpiresAt.Add(time.Second) }},
		{"guardian", func(v *ManifestBinding) { v.GuardianVersion = "other" }}, {"limits", func(v *ManifestBinding) { v.GuardianLimitsDigest = "other" }},
		{"reconcile", func(v *ManifestBinding) { v.ReconciliationWatermark = "other" }}, {"protection profile", func(v *ManifestBinding) { v.ProtectionProfile = "other" }},
		{"protection state", func(v *ManifestBinding) { v.ProtectionState = "UNWIRED" }}, {"operating", func(v *ManifestBinding) { v.OperatingPolicy = "other" }},
		{"scheduler scope", func(v *ManifestBinding) { v.SchedulerScope = "other" }}, {"calendar", func(v *ManifestBinding) { v.CalendarVersion = "other" }},
		{"lane approval", func(v *ManifestBinding) { v.LaneApproved = false }}, {"scheduler approval", func(v *ManifestBinding) { v.SchedulerApproved = false }},
		{"autostart approval", func(v *ManifestBinding) { v.AutoStartApproved = false }}, {"gate approval", func(v *ManifestBinding) { v.GateApproved = false }},
		{"live approval", func(v *ManifestBinding) { v.LiveApproved = false }}, {"actor", func(v *ManifestBinding) { v.Actor = "other" }},
		{"audit", func(v *ManifestBinding) { v.AuditID = "other" }}, {"issued", func(v *ManifestBinding) { v.IssuedAt = v.IssuedAt.Add(time.Second) }},
		{"expires", func(v *ManifestBinding) { v.ExpiresAt = v.ExpiresAt.Add(time.Second) }}, {"generation", func(v *ManifestBinding) { v.Generation++ }},
	}
	if got, want := len(tests), reflect.TypeOf(base).NumField(); got != want {
		t.Fatalf("field mismatch tests=%d manifest fields=%d", got, want)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changed := base
			tc.mutate(&changed)
			if _, err := installedRepository(base).Verify(changed, now); err == nil {
				t.Fatal("mismatch accepted")
			}
		})
	}
}

func TestActivationRejectsMissingExpiryRevocationAndGenerationRollback(t *testing.T) {
	now := time.Now().UTC()
	base := bindingFixture(now)
	if _, err := NewDormantManifestRepository().Verify(base, now); !errors.Is(err, ErrActivationNotConfigured) {
		t.Fatalf("dormant err=%v", err)
	}
	revoked := installedRepository(base)
	revoked.current.revoked = true
	if _, err := revoked.Verify(base, now); err == nil {
		t.Fatal("revoked accepted")
	}
	expired := base
	expired.ExpiresAt = now
	if _, err := installedRepository(expired).Verify(expired, now); err == nil {
		t.Fatal("expired accepted")
	}
	attestExpired := base
	attestExpired.AttestationExpiresAt = now
	if _, err := installedRepository(attestExpired).Verify(attestExpired, now); err == nil {
		t.Fatal("expired attestation accepted")
	}
	rollback := base
	rollback.Generation = 6
	if _, err := installedRepository(base).Verify(rollback, now); err == nil {
		t.Fatal("generation rollback accepted")
	}
}

type guardianSpy struct {
	calls  int
	mutate func()
	err    error
}

func (s *guardianSpy) Authorize(context.Context, EntryDecision) (string, error) {
	s.calls++
	if s.mutate != nil {
		s.mutate()
	}
	return "guardian", s.err
}

type attemptSpy struct {
	calls int
	seen  map[string]bool
	last  PlannedAttempt
}

func (s *attemptSpy) Plan(_ context.Context, a PlannedAttempt) error {
	s.calls++
	if s.seen == nil {
		s.seen = map[string]bool{}
	}
	if s.seen[a.ClientOrderID] {
		return errors.New("duplicate")
	}
	s.seen[a.ClientOrderID] = true
	s.last = a
	return nil
}

type gatewaySpy struct {
	calls int
	last  PlannedAttempt
}

func (s *gatewaySpy) Place(_ context.Context, a PlannedAttempt) (string, error) {
	s.calls++
	s.last = a
	return "broker", nil
}

func acceptedDecision() EntryDecision {
	return EntryDecision{Accepted: true, Identity: "decision", CandidateLifeID: "life", LaneID: LaneID, LaneVersion: LaneVersion}
}
func TestDispatchBlockersCallNoAuthority(t *testing.T) {
	now := time.Now().UTC()
	binding := bindingFixture(now)
	for _, gates := range []GateState{{}, {LaneDesired: true}, {LaneDesired: true, LaneEffective: true, KillSwitch: true}, {LaneDesired: true, LaneEffective: true}, {LaneDesired: true, LaneEffective: true, ProtectionWired: true}} {
		g := &guardianSpy{}
		a := &attemptSpy{}
		w := &gatewaySpy{}
		_ = Dispatch(context.Background(), acceptedDecision(), binding, gates, Dependencies{Manifest: NewDormantManifestRepository(), Guardian: g, Attempts: a, Gateway: w, Now: func() time.Time { return now }})
		if g.calls+a.calls+w.calls != 0 {
			t.Fatalf("blocked calls=%d/%d/%d", g.calls, a.calls, w.calls)
		}
	}
}
func TestDispatchRevalidatesManifestAtSubmitBoundaryAndPreservesOrder(t *testing.T) {
	now := time.Now().UTC()
	binding := bindingFixture(now)
	repo := installedRepository(binding)
	g := &guardianSpy{mutate: func() { repo.mu.Lock(); repo.current.revoked = true; repo.mu.Unlock() }}
	a := &attemptSpy{}
	w := &gatewaySpy{}
	gates := GateState{LaneDesired: true, LaneEffective: true, ProtectionWired: true, GateOpen: true}
	if err := Dispatch(context.Background(), acceptedDecision(), binding, gates, Dependencies{Manifest: repo, Guardian: g, Attempts: a, Gateway: w, Now: func() time.Time { return now }}); err == nil {
		t.Fatal("TOCTOU accepted")
	}
	if g.calls != 1 || a.calls != 0 || w.calls != 0 {
		t.Fatalf("calls=%d/%d/%d", g.calls, a.calls, w.calls)
	}
}
func TestDispatchPlansBeforeOfficialGatewayAndRejectsDuplicateIdentity(t *testing.T) {
	now := time.Now().UTC()
	binding := bindingFixture(now)
	repo := installedRepository(binding)
	g := &guardianSpy{}
	a := &attemptSpy{}
	w := &gatewaySpy{}
	deps := Dependencies{Manifest: repo, Guardian: g, Attempts: a, Gateway: w, Now: func() time.Time { return now }}
	gates := GateState{LaneDesired: true, LaneEffective: true, ProtectionWired: true, GateOpen: true}
	if err := Dispatch(context.Background(), acceptedDecision(), binding, gates, deps); err != nil {
		t.Fatal(err)
	}
	wantDigest, _ := manifestDigest(binding)
	if a.calls != 1 || w.calls != 1 || w.last.ManifestDigest != wantDigest {
		t.Fatalf("calls/attempt=%d/%d %+v", a.calls, w.calls, w.last)
	}
	if err := Dispatch(context.Background(), acceptedDecision(), binding, gates, deps); err == nil {
		t.Fatal("duplicate accepted")
	}
	if w.calls != 1 {
		t.Fatalf("gateway called on duplicate: %d", w.calls)
	}
}
