package main

import (
	"context"
	"path/filepath"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
)

// newConsoleOptimizationCommander constructs the canonical optimization
// control-plane service beside, but never inside, the trading journal. The
// returned type exposes no broker/order/LIVE/lane/gate mutation method.
func newConsoleOptimizationCommander(ctx context.Context, journalPath string) (*strategyopt.Store, error) {
	registry, err := strategyopt.CoreRegistry(ctx)
	if err != nil {
		return nil, err
	}
	return strategyopt.Open(ctx, strategyopt.Options{
		Path:     filepath.Join(filepath.Dir(journalPath), strategyopt.DatabaseFileName),
		Registry: registry,
		Actor:    "console-operator",
		// a049 is not yet integrated at this HEAD. Nil is intentional: the
		// lifecycle reports `unavailable`, blocks evidence-backed candidates and
		// still permits a human to preview a finite owner-provided server preset.
		Evidence: nil,
	})
}
