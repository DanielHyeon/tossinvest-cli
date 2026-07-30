package execgw

// protection.go is interlock clause 6, at the place that can enforce it
// (engine-safety "자동화 게이트 기동 인터록"; change interlock-gates-entry-not-exit,
// design D2 and D3).
//
// # What moved, and what did not
//
// The clause used to live in the startup interlock and refuse the whole runtime.
// Its stated purpose was narrower than that — "보호주문 도입 change가 이 표지를
// 배선하기 전에는 자동 진입이 켜질 수 없다" — and the gap between the purpose and
// the mechanism had a cost that showed up in operation: a build with no
// broker-resident protection also got no exit observer, so the account's holdings
// carried no stop at all. The clause names "a position with no protection" as the
// failure it exists to prevent, and refusing the runtime produced exactly that,
// for every holding rather than for new ones.
//
// So the refusal moved down here and narrowed to what the purpose says: a
// mutation that raises exposure. The runtime starts; entries do not.
//
// # Why here and not in the issuer
//
// `RiskGuardian.IssueEntry` was the other candidate. The gateway wins because of
// where its answer comes from: `mutationPlan.raisesExposure` is computed from the
// intent's own shape (`side == "buy"`, and for an amend, a quantity or price that
// moved up) and never from the Safety Class the caller declared — gateway.go says
// so at the point it computes it. Blocking at the issuer would leave "a decision
// that did not come from this issuer" as a shape the block does not see. Blocking
// on the plan leaves no shape at all.
//
// It is also the *only* place. The tracer's own comment already reached this
// conclusion about a second gate: "Adding a second gate here would be a second
// place to get the answer wrong."
//
// # Why the marker is a constant
//
// Unchanged from 5.2, and load-bearing. A config key, an exported Options field
// or a setter would each be a way for a build to claim protective execution it
// does not have, and this clause exists to refuse that claim. The change that
// wires broker-resident protective orders flips the identifier below as the last
// step of its own work — that is the whole interface.
//
// The unexported override beneath it is the same seam the engine's interlock has
// carried since 5.2 (`Options.protectionOverride`, reachable only from
// export_test.go). Without it the tests about *accepted* raising mutations —
// idempotency, IN_DOUBT resolution, reservations, round trips — would all become
// tests about this refusal, and the paths they cover would go untested until the
// protective-order change lands.

// ProtectionReadiness is whether broker-resident protective execution is wired
// into this build.
type ProtectionReadiness string

const (
	// ProtectionUnwired: nothing places or maintains a protective order at the
	// broker. A position is protected only by a local observation loop, which is
	// protection that ends when the process does.
	ProtectionUnwired ProtectionReadiness = "UNWIRED"
	// ProtectionWired: protective execution is wired, so a position survives the
	// engine dying. Nothing in this build produces this value.
	ProtectionWired ProtectionReadiness = "WIRED"
)

// ProfileProtection is this build's readiness.
//
// Exported so the startup interlock can report it without duplicating the fact,
// and constant so that reporting it is the only thing anybody can do with it.
const ProfileProtection = ProtectionUnwired

// WiredProtectionForTest is the value a test assigns to
// Options.ProtectionOverrideForTest.
//
// It exists so that a cross-package test does not have to take the address of a
// local: `opts.ProtectionOverrideForTest = execgw.WiredProtectionForTest` reads
// as what it is, and greps as one string. Non-test code has no reason to name it,
// and protection_test.go proves none does.
var WiredProtectionForTest = &wiredForTest

var wiredForTest = ProtectionWired

// defaultProtection is what a gateway judges by when its Options carry no
// override. It is the build's constant, and no shipped file assigns it —
// protection_test.go proves that over the AST.
//
// A variable rather than a direct read of the constant for one reason: this
// package's own suites drive buys in almost every file, because a buy is the
// mutation with the most machinery behind it (idempotency, reservations,
// IN_DOUBT resolution, the round trip). Making each of them say so at its
// construction site would be nineteen edits that mean nothing, so export_test.go
// flips this default once for the test binary and the two tests that are about
// clause 6 opt back out per gateway. The same trick as evaluateChain next door.
var defaultProtection = ProfileProtection

// protection reports the readiness this gateway judges against.
func (g *Gateway) protection() ProtectionReadiness {
	if g.protectionOverride != nil {
		return *g.protectionOverride
	}
	return defaultProtection
}

// UnwiredProtectionForTest is the value a test assigns to
// Options.ProtectionOverrideForTest to get the shipped behaviour back.
//
// It exists because this package's test binary defaults to WIRED (see
// defaultProtection): a test that wants to observe the refusal has to ask for
// the build's own answer explicitly.
var UnwiredProtectionForTest = &unwiredForTest

var unwiredForTest = ProtectionUnwired

// checkProtection refuses a mutation that raises exposure while no protective
// order can be left at the broker.
//
// Reductions are never refused, and that asymmetry is the change: an entry that
// outlives the process leaves a *new* position with no stop, while an exit that
// does not happen leaves the position that was already there. Refusing the second
// to prevent the first is how the clause used to be read, and it kept the account
// in the state the clause is about.
func (g *Gateway) checkProtection(plan mutationPlan) *RejectedError {
	if !plan.raisesExposure || g.protection() == ProtectionWired {
		return nil
	}
	return reject(ReasonProtectionNotWired,
		"this build has no broker-resident protective order execution, so a stop is a local "+
			"judgement that does not survive the process — %s of %s raises exposure and is refused "+
			"until the protective-order change wires it. Reductions are unaffected",
		plan.kind, plan.symbol)
}
