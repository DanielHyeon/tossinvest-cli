package execgw

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
)

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
// ProfileProtection remains a compatibility/reporting marker and stays
// UNWIRED. It is never execution authority. A mutation is admitted only by an
// immutable, market-scoped snapshot obtained through the sealed adapter below;
// no config key, exported scalar Options field, or setter can claim WIRED.

// ProtectionReadiness is the compatibility status reported for this build. The
// gateway never consumes this scalar as authorization.
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

// defaultProtectionCheckForTest has no production setter and is nil in every
// shipped binary. export_test.go assigns it while building this package's test
// binary so legacy gateway tests can reach behavior unrelated to readiness.
// Production authorization always comes from the sealed market adapter.
var defaultProtectionCheckForTest func(context.Context, string, protectionCheckpoint) (protectionCheckpoint, *RejectedError)

// checkProtection refuses a mutation that raises exposure while no protective
// order can be left at the broker.
//
// Reductions are never refused, and that asymmetry is the change: an entry that
// outlives the process leaves a *new* position with no stop, while an exit that
// does not happen leaves the position that was already there. Refusing the second
// to prevent the first is how the clause used to be read, and it kept the account
// in the state the clause is about.
type protectionCheckpoint struct {
	adapter      protection.Checkpoint
	testIdentity string
}

func (g *Gateway) checkProtection(ctx context.Context, plan mutationPlan, previous protectionCheckpoint) (protectionCheckpoint, *RejectedError) {
	if !plan.raisesExposure {
		return protectionCheckpoint{}, nil
	}
	if g.protectionCheckForTest != nil {
		return g.protectionCheckForTest(ctx, plan.market, previous)
	}
	if g.protectionReadiness == nil {
		return protectionCheckpoint{}, protectionNotWired(plan, "market readiness provider is missing")
	}
	checkpoint, refusal := g.protectionReadiness.Check(ctx, plan.market, g.clk.Now(), previous.adapter)
	if refusal != nil {
		return protectionCheckpoint{}, protectionNotWired(plan, refusal.Error())
	}
	return protectionCheckpoint{adapter: checkpoint}, nil
}
