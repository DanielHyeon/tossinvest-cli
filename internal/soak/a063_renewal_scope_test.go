package soak

import (
	"path/filepath"
	"testing"
	"time"
)

// a063 D6: 발급은 됐지만 감독 검증 endpoint 를 잃은 갱신은 issued + live_endpoint_uncovered 로 남음.
// issued 에 붙을 수 있는 사유는 그 코드 하나뿐임.
func TestA063IssuedStatusMayCarryOnlyTheUncoveredScopeReason(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	issued := func(codes ...RenewalReasonCode) RenewalStatus {
		return RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalIssued,
			ReasonCodes: codes, ExpiresAt: now.Add(30 * 24 * time.Hour)}
	}
	if err := issued(ReasonLiveUncovered).Validate(now); err != nil {
		t.Fatalf("issued status naming the uncovered scope was rejected: %v", err)
	}
	if err := issued().Validate(now); err != nil {
		t.Fatalf("fully covered issued status was rejected: %v", err)
	}
	for _, codes := range [][]RenewalReasonCode{{ReasonStreak}, {ReasonLiveUncovered, ReasonStreak}, {ReasonInput}} {
		if err := issued(codes...).Validate(now); err == nil {
			t.Errorf("issued status with reasons %v was accepted; only %q may accompany issued", codes, ReasonLiveUncovered)
		}
	}
	refused := RenewalStatus{FormatVersion: 1, AttemptedAt: now, Outcome: RenewalRefused, ReasonCodes: []RenewalReasonCode{ReasonLiveUncovered}}
	if err := refused.Validate(now); err != nil {
		t.Errorf("the new code is not a closed, valid reason: %v", err)
	}
	if !renewalStatusReaderSupported() {
		return
	}
	path := filepath.Join(t.TempDir(), "capability-attestation.json"+RenewalStatusSuffix)
	if err := SaveRenewalStatus(path, issued(ReasonLiveUncovered)); err != nil {
		t.Fatalf("SaveRenewalStatus: %v", err)
	}
	got, err := LoadRenewalStatus(path, now.Add(time.Second))
	if err != nil {
		t.Fatalf("LoadRenewalStatus: %v", err)
	}
	if got.Outcome != RenewalIssued || len(got.ReasonCodes) != 1 || got.ReasonCodes[0] != ReasonLiveUncovered {
		t.Fatalf("round trip = %+v", got)
	}
}
