package soak

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRenewalStatusRoundTripAndRejectsUnsafeInputs(t *testing.T) {
	if !renewalStatusReaderSupported() {
		t.Skip("secure renewal-status reader is intentionally unavailable on this platform")
	}
	now := time.Now().UTC().Truncate(time.Second)
	path := filepath.Join(t.TempDir(), "capability-attestation.json.renewal-status.json")
	want := RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonStreak}}
	if err := SaveRenewalStatus(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadRenewalStatus(path, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != RenewalRefused || len(got.ReasonCodes) != 1 || got.ReasonCodes[0] != ReasonStreak {
		t.Fatalf("got %+v", got)
	}
	if err := os.WriteFile(path, []byte(`{"format_version":1,"attempted_at":"2026-01-01T00:00:00Z","outcome":"issued","reason_codes":[],"unknown":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRenewalStatus(path, now); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestRenewalStatusIssuedNeedsConsistentExpiry(t *testing.T) {
	now := time.Now().UTC()
	if err := (RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalIssued}).Validate(now); err == nil {
		t.Fatal("issued status without expiry accepted")
	}
}

func TestRenewalStatusRejectsEmptyFailureReasonsAndNonUTCTimestamps(t *testing.T) {
	now := time.Now().UTC()
	if err := (RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalFailed}).Validate(now); err == nil {
		t.Fatal("failed status without a reason was accepted")
	}
	offset := time.FixedZone("KST", 9*60*60)
	if err := (RenewalStatus{FormatVersion: 1, AttemptedAt: now.In(offset), Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonInput}}).Validate(now.Add(time.Hour)); err == nil {
		t.Fatal("non-UTC-offset attempted_at was accepted")
	}
}

func TestRenewalStatusRejectsFutureDuplicateAndOversizeSchemas(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	base := RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonStreak}}
	future := base
	future.AttemptedAt = now.Add(time.Second)
	if err := future.Validate(now); err == nil {
		t.Fatal("future status accepted")
	}
	duplicate := base
	duplicate.ReasonCodes = []RenewalReasonCode{ReasonStreak, ReasonStreak}
	if err := duplicate.Validate(now); err == nil {
		t.Fatal("duplicate codes accepted")
	}
	path := filepath.Join(t.TempDir(), "status.json")
	if err := os.WriteFile(path, make([]byte, renewalStatusMaxBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRenewalStatus(path, now); err == nil {
		t.Fatal("oversize status accepted")
	}
}

func TestLoadRenewalStatusRejectsInsecureFileShapes(t *testing.T) {
	if !renewalStatusReaderSupported() {
		t.Skip("secure renewal-status reader is intentionally unavailable on this platform")
	}
	now := time.Now().UTC().Truncate(time.Second)
	dir := t.TempDir()
	path := filepath.Join(dir, "capability-attestation.json.renewal-status.json")
	if err := SaveRenewalStatus(path, RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonStreak}}); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRenewalStatus(path, now.Add(time.Second)); err == nil {
		t.Fatal("group-readable renewal status accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "missing"), path); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRenewalStatus(path, now.Add(time.Second)); err == nil {
		t.Fatal("symlink renewal status accepted")
	}
}

func TestSaveRenewalStatusDoesNotReplacePriorStatusWhenTemporaryWriteFails(t *testing.T) {
	if !renewalStatusReaderSupported() {
		t.Skip("secure renewal-status reader is intentionally unavailable on this platform")
	}
	now := time.Now().UTC().Truncate(time.Second)
	path := filepath.Join(t.TempDir(), "capability-attestation.json.renewal-status.json")
	prior := RenewalStatus{FormatVersion: 1, AttemptedAt: now.Add(-time.Minute), Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonStreak}}
	if err := SaveRenewalStatus(path, prior); err != nil {
		t.Fatal(err)
	}
	original := createRenewalStatusTemp
	createRenewalStatusTemp = func(dir, pattern string) (*os.File, error) {
		file, err := original(dir, pattern)
		if err == nil {
			_ = file.Close()
		}
		return file, err
	}
	t.Cleanup(func() { createRenewalStatusTemp = original })
	if err := SaveRenewalStatus(path, RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalFailed, ReasonCodes: []RenewalReasonCode{ReasonInput}}); err == nil {
		t.Fatal("closed temporary file was accepted")
	}
	got, err := LoadRenewalStatus(path, now.Add(time.Second))
	if err != nil {
		t.Fatalf("Load prior status: %v", err)
	}
	if got.Outcome != prior.Outcome || len(got.ReasonCodes) != 1 || got.ReasonCodes[0] != ReasonStreak {
		t.Errorf("prior status was replaced after temporary write failure: %+v", got)
	}
}
