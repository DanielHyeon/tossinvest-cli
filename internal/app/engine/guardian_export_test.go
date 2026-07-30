package engine

import (
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// DisableProductionGuardianForTest leaves a gate-ON profile with no Guardian so
// direct tests can still prove the interlock's Guardian-required refusal.
// TESTS ONLY. The production binary always constructs the real Guardian.
func (o *Options) DisableProductionGuardianForTest() {
	o.disableProductionGuardian = true
}

// FailProductionGuardianForTest captures the journal handed to the production
// constructor and makes construction fail. TESTS ONLY.
func (o *Options) FailProductionGuardianForTest(capture func(*journal.Journal)) {
	o.productionGuardianFactory = func(_ config.AutomationGate, j *journal.Journal, _ clock.Clock,
		_ string, _ journal.ModeAnnouncer,
	) (execgw.Guardian, error) {
		if capture != nil {
			capture(j)
		}
		return nil, errors.New("injected production Guardian construction failure")
	}
}
