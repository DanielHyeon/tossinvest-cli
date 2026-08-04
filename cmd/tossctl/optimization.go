package main

import (
	"context"
	"path/filepath"
	"time"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/optimizationevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

const consolePerformanceDatabaseFileName = "performance.db"

// consolePerformanceCapabilities deliberately retain only the two read
// interfaces required by the transports. The performance read-only adapter has
// no Collect/Prune authority, and its concrete handle never leaves this closure.
type consolePerformanceCapabilities struct {
	Performance performanceDashboardReader
	Evidence    strategyopt.EvidenceProvider
	close       func() error
}

type performanceDashboardReader interface {
	Dashboard(context.Context, performance.Query) (performance.DashboardView, error)
	AttributionRows(context.Context, string, performance.AttributionQuery, int) ([]performance.Attribution, error)
}

func openConsolePerformanceCapabilities(dataDir string, now func() time.Time) (consolePerformanceCapabilities, error) {
	store, err := performance.OpenReadOnly(filepath.Join(dataDir, consolePerformanceDatabaseFileName))
	if err != nil {
		return consolePerformanceCapabilities{}, err
	}
	return consolePerformanceCapabilities{
		Performance: store,
		Evidence:    optimizationevidence.New(store, now),
		close:       store.Close,
	}, nil
}

func (c consolePerformanceCapabilities) Close() error {
	if c.close == nil {
		return nil
	}
	return c.close()
}

// newConsoleOptimizationCommander constructs the canonical optimization
// control-plane service beside, but never inside, the trading journal. The
// returned type exposes no broker/order/LIVE/lane/gate mutation method.
func newConsoleOptimizationCommander(
	ctx context.Context,
	journalPath string,
	evidence strategyopt.EvidenceProvider,
) (*strategyopt.Store, error) {
	registry, err := strategyopt.CoreRegistry(ctx)
	if err != nil {
		return nil, err
	}
	return strategyopt.Open(ctx, strategyopt.Options{
		Path:     filepath.Join(filepath.Dir(journalPath), strategyopt.DatabaseFileName),
		Registry: registry,
		Actor:    "console-operator",
		// The provider exposes only ReadEvidence over a049 Dashboard. A nil
		// provider is the explicit startup fallback: evidence-backed candidates
		// remain blocked while finite owner-provided server presets stay visible.
		Evidence: evidence,
	})
}
