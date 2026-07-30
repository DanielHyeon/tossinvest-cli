//go:build tossos_testseams

package engine

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// ConfigureCLIRegressionForTest enables only the seams the cross-package
// command assembly regression needs. This symbol does not exist without the
// tossos_testseams build tag.
func ConfigureCLIRegressionForTest(
	opts *Options,
	prober journal.FSProber,
	constructed func(),
) {
	opts.journalFSProber = prober
	opts.productionGuardianFactory = func(gate config.AutomationGate, j *journal.Journal, clk clock.Clock,
		accountRef string, announcer journal.ModeAnnouncer,
	) (execgw.Guardian, error) {
		if constructed != nil {
			constructed()
		}
		return defaultProductionGuardianFactory(gate, j, clk, accountRef, announcer)
	}
}
