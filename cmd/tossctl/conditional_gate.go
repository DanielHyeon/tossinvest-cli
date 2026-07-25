package main

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// conditionalGate is a shim over trading.ConditionalGate, which is where the
// conditional-order write policy now lives (internal/trading/conditional.go) so
// every surface — cobra, MCP, the trading engine — enforces one copy of it.
//
// Kept as a function with this signature because the CLI has a config.File in
// hand, not just its trading section, and because its test pins the four
// outcomes at this seam.
func conditionalGate(cfg config.File, canonical string, execute bool, confirm string) error {
	return trading.ConditionalGate(cfg.Trading, canonical, execute, confirm)
}

// errConditionalPreviewOnly is the same sentinel value the trading package
// returns, so `errors.Is` comparisons on either side agree.
var errConditionalPreviewOnly = trading.ErrConditionalPreviewOnly
