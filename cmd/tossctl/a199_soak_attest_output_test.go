package main

// a199_soak_attest_output_test.go — `soak attest` 가 운영자에게 게이트의 귀결을 사실대로 말하는지(gstack review testing 지적).
//
//   - 주문 실행(POST) endpoint 가 빠진 attestation: 게이트가 켜진 엔진은 기동 거절이 아니라 청산 전용으로 뜸 — 그렇게 말함.
//   - 다 덮인 attestation: 검증된 기동이지만 새 진입은 여전히 게이트웨이에서 막힘(보호 미배선) — 그렇게 말함.
// 기존 시험 함수 본문은 건드리지 않음(새 파일 — 기존 함수 편집은 FLM 번들을 요구함).

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

func TestA199SoakAttestSaysAPartialAttestationStartsExitOnly(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedQualifyingRecord(t, filepath.Join(configDir, soak.FileName))

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest")
	if err != nil {
		t.Fatalf("soak attest: %v", err)
	}
	if !strings.Contains(out, "will start exit-only") || strings.Contains(out, "refuse to start") {
		t.Errorf("soak attest output misstates the gate outcome for a partial attestation:\n%s", out)
	}
}

func TestA199SoakAttestSaysAFullAttestationStillRefusesEntriesAtTheGateway(t *testing.T) {
	configDir := testenv.Isolate(t)
	seedQualifyingRecord(t, filepath.Join(configDir, soak.FileName))
	seedSupervisedRecord(t, filepath.Join(configDir, verifylive.RecordFileName(verifylive.MarketKR)),
		attest.Mask("123-45-678901"), time.Now().UTC().Add(-2*time.Hour))

	out, _, err := runCLI(t, "--config-dir", configDir, "soak", "attest")
	if err != nil {
		t.Fatalf("soak attest: %v", err)
	}
	if !strings.Contains(out, "starts verified") || !strings.Contains(out, "refused at the gateway") ||
		strings.Contains(out, "exit-only") || strings.Contains(out, "refuse to start") {
		t.Errorf("soak attest output misstates the gate outcome for a full attestation:\n%s", out)
	}
}
