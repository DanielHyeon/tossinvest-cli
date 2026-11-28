package console

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// a063 D6: issued + live_endpoint_uncovered 는 고정 문장으로 보임 — 일반 오류 문구로 떨어지지 않음.
func TestA063DashboardTellsAnIssuedRenewalThatLostTheSupervisedScope(t *testing.T) {
	h := newHarness(t)
	now := time.Now().UTC().Truncate(time.Second)
	expires := now.Add(30 * 24 * time.Hour)
	if err := attest.Save(h.opts.Attestation, attest.Attestation{
		FormatVersion: attest.FormatVersion,
		AccountRef:    "123-45-678901",
		IssuedAt:      now.Add(-time.Hour),
		ExpiresAt:     expires,
		SoakDays:      3,
		Endpoints:     soak.RequiredEndpoints(),
		VerifiedBy:    "fixture",
	}); err != nil {
		t.Fatalf("Save attestation: %v", err)
	}
	if err := soak.SaveRenewalStatus(soak.RenewalStatusPath(h.opts.Attestation), soak.RenewalStatus{
		FormatVersion: 1,
		AttemptedAt:   now.Add(-time.Hour),
		Outcome:       soak.RenewalIssued,
		ReasonCodes:   []soak.RenewalReasonCode{soak.ReasonLiveUncovered},
		ExpiresAt:     expires,
	}); err != nil {
		t.Fatalf("SaveRenewalStatus: %v", err)
	}
	h.authenticate(t)
	page := body(t, h.get(t, pathVerifyConsole))
	for _, want := range []string{"renewal 진단: issued", "감독 검증 증거 없이 발급되었다"} {
		if !strings.Contains(page, want) {
			t.Errorf("dashboard missing %q:\n%s", want, truncateForLog(page))
		}
	}
	if strings.Contains(page, "renewal 입력 또는 증거를 확인할 수 없다") {
		t.Error("the uncovered-scope reason fell through to the generic message")
	}
}
