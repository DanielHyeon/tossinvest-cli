package main

// a199_soak_attest_output_test.go — `soak attest` 가 운영자에게 게이트의 귀결을 사실대로 말하는지(gstack review testing 지적).
//
//   - 주문 실행(POST) endpoint 가 빠진 attestation: 귀결은 원장에 달림(G2) — 이 계좌로 엔진 주문이 확정된 적이 없으면 기동 거절,
//     있으면 청산 전용. 둘 다 말함.
//   - 다 덮인 attestation: 파일이 덮는 범위만 말하고(계좌와 나머지 조항은 기동 때 판정 — G6) 새 진입은 여전히 게이트웨이에서
//     막힘(보호 미배선)을 말함.
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
	// a199 G2: 귀결은 원장에 달림 — 엔진 주문이 확정된 적 없는 계좌는 거절, 있으면 청산 전용. 둘 다 말해야 함.
	if !strings.Contains(out, "refuses to start") || !strings.Contains(out, "starts exit-only if the engine already had an order confirmed") {
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
	// gstack review G6: 파일이 덮는 범위만 말함 — 계좌와 나머지 조항은 기동 때 판정(「starts verified」 로 단정하지 않음).
	if !strings.Contains(out, "judged at engine start") || !strings.Contains(out, "refused at the gateway") ||
		strings.Contains(out, "starts verified") ||
		strings.Contains(out, "exit-only") || strings.Contains(out, "refuse to start") {
		t.Errorf("soak attest output misstates the gate outcome for a full attestation:\n%s", out)
	}
}
