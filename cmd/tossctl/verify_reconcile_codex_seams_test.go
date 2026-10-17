//go:build tossos_testseams

package main

// verify_reconcile_codex_seams_test.go — codex CG-4 의 채널 시험(전 경로 — Q1 seam 이 필요해 tossos_testseams).

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

// TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion — 두 마스크·계좌 수와 y/N 질문이 같은 단말 채널로 간다.
// stdout 을 리다이렉트해도(여기서는 별도 버퍼) 운영자는 질문과 함께 계좌 대조를 본다.
func TestReconcileShowsTheAccountsOnTheSameChannelAsTheQuestion(t *testing.T) {
	configDir := testenv.Isolate(t)
	seamA063Record(t, configDir, time.Now().Add(-2*time.Hour))
	seamPolicies(t)
	terminal := setReconcileTerminal(t, "y\n", true)
	seamProbeFactory(t, configDir)
	var stdout strings.Builder
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetContext(context.Background())
	if err := runVerifyReconcile(cmd, &rootOptions{configDir: configDir}, &verifyReconcileOptions{market: verifylive.MarketKR}); err != nil {
		t.Fatalf("every condition holds: %v", err)
	}
	shown := terminal.String()
	mask := attest.Mask("123-45-678901")
	if strings.Count(shown, mask) < 2 || !strings.Contains(strings.ToLower(shown), "y/n") {
		t.Fatalf("the terminal channel must carry both masks and the question:\n%s", shown)
	}
	if strings.Contains(stdout.String(), mask) {
		t.Fatalf("the account comparison went to stdout instead of the terminal channel:\n%s", stdout.String())
	}
}
