package engine_test

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
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

// officialMutators are the write methods on *official.Client. They are listed
// explicitly rather than derived, so that a new mutator added to the official
// client has to be added here consciously — which is the moment somebody decides
// whether the engine's read-only view should be able to reach it (it should not).
var officialMutators = []string{
	"PlaceOrder",
	"CancelOrder",
	"ModifyOrder",
	"CreateConditionalOrder",
	"CancelConditionalOrder",
	"ModifyConditionalOrder",
}

// TestContextOfficialIsReadOnly is the other half of the seal (task 4.2).
//
// Task 2.5 made the mutators unexported but left Context.Official typed as
// *official.Client — which has PlaceOrder, CancelOrder and ModifyOrder on it. A
// caller holding the engine wiring could therefore still submit an order with no
// journal record, no GuardianDecision and no IN_DOUBT handling: the seal was real
// for one type and cosmetic for the other (issues.md, 2026-07-26).
//
// The field is now an interface that does not declare those methods, so the call
// cannot be spelled. This test fails if the concrete type comes back or if a
// mutator is ever added to OfficialReads.
func TestContextOfficialIsReadOnly(t *testing.T) {
	field, ok := reflect.TypeOf(engine.Context{}).FieldByName("Official")
	if !ok {
		t.Fatal("Context has no Official field")
	}
	if field.Type.Kind() != reflect.Interface {
		t.Fatalf("Context.Official is %s; it must be a read-only interface, not the concrete client",
			field.Type)
	}
	for _, name := range officialMutators {
		if _, has := field.Type.MethodByName(name); has {
			t.Errorf("Context.Official exposes %s: engine mutations must go through internal/execgw.Gateway",
				name)
		}
	}
}

// TestOfficialReadsDeclaresNoMutator asserts the same property of the interface
// itself, so the check survives a future refactor that renames or moves the field.
func TestOfficialReadsDeclaresNoMutator(t *testing.T) {
	reads := reflect.TypeOf((*engine.OfficialReads)(nil)).Elem()
	for _, name := range officialMutators {
		if _, has := reads.MethodByName(name); has {
			t.Errorf("OfficialReads declares %s — it is a read-only surface by definition", name)
		}
	}

	// The list above is only meaningful if these really are methods on the
	// concrete client. A typo would make the test vacuous.
	client := reflect.TypeOf((*official.Client)(nil))
	for _, name := range officialMutators {
		if _, has := client.MethodByName(name); !has {
			t.Errorf("*official.Client has no method %q — the seal test is checking for a name that does not exist",
				name)
		}
	}
}
