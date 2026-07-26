package engine_test

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
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
// TradingService used to be exempt, on the grounds that it carries the config
// policy and the confirm-token gate and does not satisfy either mutator
// interface. Task 7.4 inverts that: the confirm token is no seal at all against a
// caller holding the service, because PreviewPlace hands the token back and the
// gate is then one line of local arithmetic away from being satisfied
// (engine-safety: "확인 토큰은 호출자가 로컬에서 계산 가능하므로 봉인이 되지
// 못한다"). The field is unexported now, and
// TestContextExposesNoUnauthorisedMutation below is what would catch it coming
// back under any name.
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
	// Added by verify-execution-capability task 1.5: same call, returns the new
	// identifier the broker issues. The engine's read-only view must not reach it
	// either.
	"ModifyConditionalOrderRef",
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

// --- task 7.4: no mutation without a decision -------------------------------

// mutationVerbs are the method names an order-mutating surface uses anywhere in
// this codebase: trading.Service's, execgw.Gateway's and trading.Broker's.
//
// Listed rather than derived, so that a new verb has to be added here
// consciously — which is the moment somebody decides whether the engine context
// should be able to reach it.
var mutationVerbs = []string{
	"Place", "Cancel", "Amend",
	"ConditionalPlace", "ConditionalCancel", "ConditionalModify",
	"PlacePendingOrder", "CancelPendingOrder", "AmendPendingOrder",
}

// TestContextExposesNoUnauthorisedMutation is the seal in the form the spec
// states it (engine-safety: "엔진 컨텍스트가 노출하는 값들에서 Gateway를 거치지
// 않는 mutation 경로를 찾으면 → 그런 경로가 존재하지 않음이 정적 테스트로
// 증명된다").
//
// Naming the forbidden types would be a weaker test — it fails the day somebody
// wraps one. The property asserted instead is behavioural and type-level: every
// mutation verb reachable from an exported Context field must take a
// GuardianDecision somewhere in its arguments. That is exactly what separates
// execgw.Gateway.Place (PlaceRequest carries the decision) from
// trading.Service.Place (nothing does).
func TestContextExposesNoUnauthorisedMutation(t *testing.T) {
	ctxType := reflect.TypeOf(engine.Context{})
	for i := 0; i < ctxType.NumField(); i++ {
		f := ctxType.Field(i)
		if !f.IsExported() {
			continue
		}
		for _, verb := range mutationVerbs {
			m, has := f.Type.MethodByName(verb)
			if !has {
				continue
			}
			if !requiresGuardianDecision(m.Type) {
				t.Errorf("Context.%s (%s) exposes %s and it does not take a GuardianDecision: "+
					"engine mutations go through the gateway, which cannot be called without one",
					f.Name, f.Type, verb)
			}
		}
	}
}

// TestTheSealTestWouldCatchTheTradingService is the positive control.
//
// The assertion above is vacuous if the predicate accepts everything, and the
// thing it must reject is the exact value that used to be an exported field.
func TestTheSealTestWouldCatchTheTradingService(t *testing.T) {
	svc := reflect.TypeOf((*trading.Service)(nil))
	for _, verb := range []string{"Place", "Cancel", "Amend"} {
		m, has := svc.MethodByName(verb)
		if !has {
			t.Fatalf("*trading.Service has no %s — the control is checking a name that does not exist", verb)
		}
		if requiresGuardianDecision(m.Type) {
			t.Errorf("the predicate thinks *trading.Service.%s requires a decision; it does not, "+
				"and a seal test that accepts it proves nothing", verb)
		}
	}

	gw := reflect.TypeOf((*execgw.Gateway)(nil))
	for _, verb := range []string{"Place", "Cancel", "Amend"} {
		m, has := gw.MethodByName(verb)
		if !has {
			t.Fatalf("*execgw.Gateway has no %s", verb)
		}
		if !requiresGuardianDecision(m.Type) {
			t.Errorf("the predicate rejects *execgw.Gateway.%s, which is the sanctioned surface; "+
				"the seal test would then forbid the only permitted path", verb)
		}
	}
}

// requiresGuardianDecision reports whether any argument of a method carries an
// execgw.GuardianDecision.
//
// One level of struct fields is inspected, plus pointer and interface
// indirection, which covers the request types the gateway takes (PlaceRequest,
// CancelRequest, AmendRequest each have a Decision field). A deeper walk would
// start finding decisions in places that do not make the call authorised.
func requiresGuardianDecision(fn reflect.Type) bool {
	decision := reflect.TypeOf(execgw.GuardianDecision{})
	for i := 0; i < fn.NumIn(); i++ {
		if carriesDecision(fn.In(i), decision) {
			return true
		}
	}
	return false
}

func carriesDecision(t, decision reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == decision {
		return true
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i).Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft == decision {
			return true
		}
	}
	return false
}
