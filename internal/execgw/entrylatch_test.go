package execgw_test

// entrylatch_test.go covers add-core-domain task 2.4: the four conditions
// risk-management names (401/403 · SLO 위반 · RECONCILE · recovery 미완료) reach
// the Guardian chain through the gate's own verdict, and nothing re-derives them.

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

var latchNow = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

// latchGate has no staleness requirements, so each case raises exactly the one
// condition it is about. TestStalenessReachesTheChainToo covers the other half.
func latchGate() *execgw.EntryGate {
	return execgw.NewEntryGate(clock.NewFake(latchNow), map[execgw.RequiredQuery]time.Duration{})
}

// TestEveryEntryBlockingConditionReachesTheChain is the mapping the task asks
// for, as a table: the condition, how it is raised, and the code the chain's
// AccountState.EntryBlockedReason ends up carrying.
//
// The point of the table is that the right column is *not* invented here — it is
// whatever the gate already reports, which is why the assertion compares against
// CheckEntry rather than against a literal.
func TestEveryEntryBlockingConditionReachesTheChain(t *testing.T) {
	cases := []struct {
		name string
		// spec is the condition as risk-management names it.
		spec string
		// raise puts the gate in that state, the way production does.
		raise func(*testing.T, *execgw.EntryGate)
		want  execgw.ReasonCode
	}{
		{
			name: "401/403",
			spec: "자격증명 실패",
			raise: func(t *testing.T, g *execgw.EntryGate) {
				// Through the Retrier, not through Block: the latch is a side
				// effect of the retry matrix classifying the error, and wiring
				// that is exactly what must not be duplicated elsewhere.
				r := &execgw.Retrier{Clock: clock.NewFake(latchNow), Gate: g}
				err := r.Query(context.Background(), execgw.QueryBuyingPower,
					func(context.Context) error {
						return &official.APIError{Code: http.StatusUnauthorized}
					})
				if err == nil {
					t.Fatal("a 401 must be returned, not swallowed")
				}
			},
			want: execgw.ReasonBrokerAuthRejected,
		},
		{
			name: "SLO",
			spec: "체결 감지 SLO 위반",
			raise: func(_ *testing.T, g *execgw.EntryGate) {
				g.Block(execgw.ReasonFillDetectionSLO, "detection is 90s behind")
			},
			want: execgw.ReasonFillDetectionSLO,
		},
		{
			name: "RECONCILE",
			spec: "계정 대조 불일치",
			raise: func(_ *testing.T, g *execgw.EntryGate) {
				g.RebuildReconcileProjection([]journal.ReconcileState{{
					ID: "rec-1", AccountRef: "acct-1",
					Cause:     journal.ReconcileCauseQuantityMismatch,
					Evidence:  "the engine believes 10, the account says 4",
					EnteredAt: latchNow.Add(-time.Minute),
				}})
			},
			want: execgw.ReasonReconcileMismatch,
		},
		{
			name: "recovery",
			spec: "재시작 복구 미완료",
			raise: func(_ *testing.T, g *execgw.EntryGate) {
				g.Block(execgw.ReasonRecoveryIncomplete, "replaying attempt 3 of 9")
			},
			want: execgw.ReasonRecoveryIncomplete,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gate := latchGate()
			if latch := gate.EntryLatch(); latch.Blocked {
				t.Fatalf("precondition: a fresh gate is not latched, got %+v", latch)
			}

			tc.raise(t, gate)

			latch := gate.EntryLatch()
			if !latch.Blocked {
				t.Fatalf("%s (%s) did not reach the chain's latch field", tc.name, tc.spec)
			}
			if latch.Reason != tc.want {
				t.Fatalf("reason = %s, want %s", latch.Reason, tc.want)
			}
			if latch.Detail == "" {
				t.Fatal("the latch carries no detail; the chain's refusal would name a code with no numbers")
			}

			// The surface is the gate's answer, not a second opinion.
			rejected := gate.CheckEntry()
			if rejected == nil {
				t.Fatal("CheckEntry and EntryLatch disagree about whether the account is blocked")
			}
			if rejected.Reason != latch.Reason || rejected.Detail != latch.Detail {
				t.Fatalf("EntryLatch %+v is not CheckEntry %+v", latch, rejected)
			}
		})
	}
}

// TestStalenessReachesTheChainToo: the chain's latch field is the gate's whole
// verdict, not the four named conditions.
//
// risk-management enumerates four because those are the ones it names; the gate
// also refuses on a required read that has never succeeded, and a Guardian that
// allowed an entry the gateway is about to refuse for staleness would be issuing
// decisions it knows cannot be submitted.
func TestStalenessReachesTheChainToo(t *testing.T) {
	gate := execgw.NewEntryGate(clock.NewFake(latchNow), map[execgw.RequiredQuery]time.Duration{
		execgw.QueryBuyingPower: 45 * time.Second,
	})

	latch := gate.EntryLatch()
	if !latch.Blocked || latch.Reason != execgw.ReasonQueryStale {
		t.Fatalf("a never-observed required query must latch the chain: %+v", latch)
	}

	gate.RecordSuccess(execgw.QueryBuyingPower)
	if latch := gate.EntryLatch(); latch.Blocked {
		t.Fatalf("a successful poll clears staleness by itself: %+v", latch)
	}
}

// TestTheLatchClearsThroughTheGatesOwnRules: the chain never releases a latch,
// and the two release behaviours the gate distinguishes stay distinguished when
// seen through this surface.
func TestTheLatchClearsThroughTheGatesOwnRules(t *testing.T) {
	gate := latchGate()
	gate.Block(execgw.ReasonBrokerAuthRejected, "credential refused")

	// A successful poll does not undo an auth latch — that is the whole point of
	// a latch, and it must not become "clear" merely because the chain asked.
	gate.RecordSuccess(execgw.QueryBuyingPower)
	if latch := gate.EntryLatch(); !latch.Blocked {
		t.Fatal("a successful poll cleared an operator-only latch")
	}

	gate.Clear(execgw.ReasonBrokerAuthRejected)
	if latch := gate.EntryLatch(); latch.Blocked {
		t.Fatalf("the operator's Clear did not release the latch: %+v", latch)
	}
}

// TestSymbolScopeSurvivesTheProjection: the narrower question stays narrower.
//
// If the chain asked the account-wide question for a symbol-scoped block it
// would allow the intent, and the gateway — which asks CheckEntryFor — would
// refuse it at submission. The decision would be recorded, the reservation
// taken, and both rolled back for a fact known before either happened.
func TestSymbolScopeSurvivesTheProjection(t *testing.T) {
	gate := latchGate()
	gate.BlockSymbol("kr", "005930", execgw.ReasonBrokerStateUnknown, "a shrinking cumulative fill")

	if latch := gate.EntryLatchFor("kr", "005930"); !latch.Blocked ||
		latch.Reason != execgw.ReasonBrokerStateUnknown {
		t.Fatalf("the blocked symbol is not latched for the chain: %+v", latch)
	}
	if latch := gate.EntryLatchFor("kr", "000660"); latch.Blocked {
		t.Fatalf("one symbol's block stopped an unrelated symbol: %+v", latch)
	}
	// The account-wide question is unchanged, which is why a caller that knows
	// the symbol must ask the narrower one.
	if latch := gate.EntryLatch(); latch.Blocked {
		t.Fatalf("a symbol block became an account block: %+v", latch)
	}
}

// TestTheLatchSurfaceNeverDisagreesWithTheGate is the no-duplicate-judgement
// property, over every combination of the conditions above.
//
// A second implementation of "is the account blocked" would pass every case
// above and still drift here, because the interesting part is precedence: which
// reason surfaces when three conditions hold at once.
func TestTheLatchSurfaceNeverDisagreesWithTheGate(t *testing.T) {
	raisers := []func(*execgw.EntryGate){
		func(g *execgw.EntryGate) { g.Block(execgw.ReasonBrokerAuthRejected, "credential refused") },
		func(g *execgw.EntryGate) { g.Block(execgw.ReasonFillDetectionSLO, "detection is 90s behind") },
		func(g *execgw.EntryGate) { g.Block(execgw.ReasonRecoveryIncomplete, "replaying attempt 3 of 9") },
		func(g *execgw.EntryGate) { g.Block(execgw.ReasonReconcileMismatch, "quantities disagree") },
		func(g *execgw.EntryGate) { g.Block(execgw.ReasonFlattenInProgress, "saga fl-1 is running") },
	}

	for mask := 0; mask < 1<<len(raisers); mask++ {
		gate := latchGate()
		for i, raise := range raisers {
			if mask&(1<<i) != 0 {
				raise(gate)
			}
		}
		latch := gate.EntryLatchFor("kr", "005930")
		rejected := gate.CheckEntryFor("kr", "005930")

		switch {
		case rejected == nil && latch.Blocked:
			t.Fatalf("mask %05b: the chain is blocked and the gateway is not", mask)
		case rejected == nil:
			continue
		case !latch.Blocked:
			t.Fatalf("mask %05b: the gateway refuses (%s) and the chain is not blocked", mask, rejected.Reason)
		case latch.Reason != rejected.Reason:
			t.Fatalf("mask %05b: chain sees %s, gateway sees %s", mask, latch.Reason, rejected.Reason)
		case latch.Detail != rejected.Detail:
			t.Fatalf("mask %05b: details differ (%q vs %q)", mask, latch.Detail, rejected.Detail)
		}
	}
}

// TestAuthClassificationStillLatchesThroughTheSentinel guards the other route a
// 401 arrives by: internal/official's sentinel rather than an APIError.
func TestAuthClassificationStillLatchesThroughTheSentinel(t *testing.T) {
	gate := latchGate()
	r := &execgw.Retrier{Clock: clock.NewFake(latchNow), Gate: gate}
	err := r.Query(context.Background(), execgw.QueryHoldings, func(context.Context) error {
		return official.ErrAuth
	})
	if !errors.Is(err, official.ErrAuth) {
		t.Fatalf("Query returned %v, want the auth sentinel", err)
	}
	if latch := gate.EntryLatch(); !latch.Blocked || latch.Reason != execgw.ReasonBrokerAuthRejected {
		t.Fatalf("the sentinel route did not latch: %+v", latch)
	}
}
