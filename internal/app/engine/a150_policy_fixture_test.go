package engine

import "github.com/JungHoonGhae/tossinvest-cli/internal/journal"

func a150RatchetPolicyFixture() *journal.StrategyExitPolicySnapshot {
	value, err := journal.NewStrategyExitPolicySnapshot("")
	if err != nil {
		panic(err)
	}
	return &value
}
