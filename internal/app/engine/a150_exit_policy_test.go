package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"testing"
)

func TestA150AdmittedProtectionMatchesExitSnapshot(t *testing.T) {
	p, err := journal.NewStrategyExitPolicySnapshot("")
	if err != nil {
		t.Fatal(err)
	}
	binding := journal.StrategyProtection{Risk: journal.RiskIntent{EntryPrice: "100", StopPrice: "90"}, Policy: &p}
	state := journal.ExitState{PolicyKind: p.Kind, PolicyIdentity: p.Identity, PositionGeneration: 1, EntryPrice: "100", InitialStop: "90"}
	pos := journal.Position{InstanceSeq: 1}
	if !strategyProtectionMatchesExit(binding, pos, state) {
		t.Fatal("admitted snapshot rejected")
	}
	state.InitialStop = "80"
	if strategyProtectionMatchesExit(binding, pos, state) {
		t.Fatal("different stop accepted")
	}
}
