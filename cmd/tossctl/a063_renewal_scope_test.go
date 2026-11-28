package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// a063 D6: 감독 검증 증거 없이 발급한 갱신은 진단에 범위 상실을 남김. 종료 코드·파일 내용은 그대로임.
func TestA063RenewalWithoutSupervisedEvidenceRecordsTheUncoveredScope(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedQualifyingRecord(t, filepath.Join(configDir, soak.FileName))
	if _, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest", "--record-renewal-status"); err != nil {
		t.Fatalf("soak attest: %v (exit status must stay success when a file was issued)", err)
	}
	attestationPath := filepath.Join(configDir, attest.FileName)
	a, err := attest.Load(attestationPath)
	if err != nil {
		t.Fatalf("Load attestation: %v", err)
	}
	if len(a.MissingEndpoints(soak.LiveOnlyEndpoints())) == 0 {
		t.Fatal("fixture precondition: the issued file should lack the supervised endpoints")
	}
	status, err := soak.LoadRenewalStatus(soak.RenewalStatusPath(attestationPath), time.Now().UTC())
	if err != nil {
		t.Fatalf("LoadRenewalStatus: %v", err)
	}
	if status.Outcome != soak.RenewalIssued || !status.ExpiresAt.Equal(a.ExpiresAt) {
		t.Errorf("status = %+v, want issued with the attestation expiry", status)
	}
	if len(status.ReasonCodes) != 1 || status.ReasonCodes[0] != soak.ReasonLiveUncovered {
		t.Errorf("reason codes = %v, want [%s]: a renewal that lost the supervised endpoints must not read as a plain success",
			status.ReasonCodes, soak.ReasonLiveUncovered)
	}
}
