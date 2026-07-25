package engine_test

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// TestContextExposesNoRawMutator is the structural half of the ExecutionGateway
// seal (task 2.5, engine-safety "raw mutator 미노출").
//
// The engine's order mutators used to be exported fields on Context, which meant
// any caller holding the wiring could place, cancel or amend with no
// GuardianDecision, no journal record and no IN_DOUBT handling — the exact path
// internal/execgw exists to be the only alternative to. This test fails if such a
// field comes back.
//
// TradingService is deliberately not caught by it: it carries the config policy
// and the confirm-token gate, it is what execgw.Gateway wraps, and it does not
// satisfy either mutator interface.
func TestContextExposesNoRawMutator(t *testing.T) {
	brokerIface := reflect.TypeOf((*trading.Broker)(nil)).Elem()
	condIface := reflect.TypeOf((*trading.ConditionalBroker)(nil)).Elem()

	ctxType := reflect.TypeOf(engine.Context{})
	for i := 0; i < ctxType.NumField(); i++ {
		f := ctxType.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Type.Implements(brokerIface) {
			t.Errorf("Context.%s (%s) is an exported trading.Broker: mutations must go through the gateway",
				f.Name, f.Type)
		}
		if f.Type.Implements(condIface) {
			t.Errorf("Context.%s (%s) is an exported conditional-order mutator", f.Name, f.Type)
		}
	}
}
