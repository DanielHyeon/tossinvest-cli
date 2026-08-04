package execgw

import (
	"context"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

// export_test.go is the one seam the Guardian's tests need and the built binary
// must not have.
//
// The property being tested is negative — "the chain is evaluated exactly once
// per issuance, and the re-collection loop does not run it again" (design D1) —
// and the only honest way to test a negative about a call is to count the calls.
// Counting needs a substitute evaluator; a substitute evaluator on the exported
// API would be a knob that lets a caller replace the risk chain with one that
// allows everything. So the seam lives here, in a _test.go file, which means it
// does not exist outside `go test`.

// SetChainForTest replaces the chain evaluator and returns the restore
// function. TESTS ONLY, and not for parallel tests: the variable is
// package-scoped.
func SetChainForTest(fn func(risk.Input) risk.Decision) func() {
	previous := evaluateChain
	evaluateChain = fn
	return func() { evaluateChain = previous }
}

// SetFXAuthorityForTest supplies the q_final package's private test proof. The
// compiled product has neither this method nor any raw-FXEvidence authority
// setter; production requests must carry opaque officialfx.Evidence.
func (r *QFinalEntryIssuance) SetFXAuthorityForTest(evidence riskbucket.FXEvidence) {
	r.testFXAuthority = evidence
	r.testFXAuthoritySet = true
}

func (r *QFinalEntryIssuance) ClearFXAuthorityForTest() {
	r.testFXAuthority = riskbucket.FXEvidence{}
	r.testFXAuthoritySet = false
}

// SetAccountBaseFXForTest supplies an already sealed risk test capability.
// TESTS ONLY: production callers must bind opaque officialfx.Evidence.
func (r *QFinalEntryIssuance) SetAccountBaseFXForTest(fx risk.AccountBaseFX) {
	r.testAccountBaseFX = fx
	r.testAccountBaseFXSet = true
}

// SetAccountBaseFXForTest supplies a sealed paired-market FX capability to the
// Gateway request. TESTS ONLY; production callers retain officialfx.Evidence.
func (r *StrategyPlaceRequest) SetAccountBaseFXForTest(fx risk.AccountBaseFX) {
	r.testAccountBaseFX = fx
	r.testAccountBaseFXSet = true
}

// SetBeginStrategySubmittingForTest observes the last durable pre-send fence.
// TESTS ONLY; production always calls Journal.BeginStrategyDispatchSubmitting.
func (g *Gateway) SetStrategyDispatchLeaseForTest(
	load func(context.Context, string) (journal.StrategyDispatchLease, error),
	begin func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error),
) {
	g.loadStrategyDispatchLease = load
	g.requireStrategyTransport = func(context.Context, journal.StrategyDispatchLeaseCAS) error { return nil }
	g.skipStrategyEntryABAForTest = true
	g.dispatchStrategyVerified = func(ctx context.Context, attempt *journal.Attempt, cas journal.StrategyDispatchLeaseCAS,
		dispatch journal.DispatchFunc, verify journal.ExistenceCheck,
	) (journal.Result, error) {
		_, err := begin(ctx, cas)
		if err != nil {
			return journal.Result{AttemptID: attempt.ID(), IntentID: attempt.IntentID()}, err
		}
		outcome := dispatch(ctx, attempt)
		result := journal.Result{AttemptID: attempt.ID(), IntentID: attempt.IntentID(), Class: outcome.Class,
			BrokerOrderID: outcome.BrokerOrderID, Detail: outcome.Detail, Err: outcome.Err}
		switch outcome.Class {
		case journal.DispatchAcked:
			if strings.TrimSpace(outcome.BrokerOrderID) == "" {
				result.Final, result.ReasonCode = journal.StateInDoubt, journal.ReasonAckWithoutOrderID
			} else if verify != nil {
				if verifyErr := verify(context.WithoutCancel(ctx), outcome.BrokerOrderID); verifyErr != nil {
					result.Final, result.ReasonCode, result.Err = journal.StateInDoubt, journal.ReasonAckRoundTripUnconfirmed, verifyErr
				} else {
					result.Final, result.ReasonCode = journal.StateConfirmed, journal.ReasonBrokerAcknowledged
				}
			} else {
				result.Final, result.ReasonCode = journal.StateConfirmed, journal.ReasonBrokerAcknowledged
			}
		case journal.DispatchNotSent:
			result.Final, result.ReasonCode = journal.StateNotDispatched, journal.ReasonDispatchNotSent
		case journal.DispatchRejected:
			result.Final, result.ReasonCode = journal.StateFailedConfirmed, journal.ReasonDispatchRejected
		default:
			result.Final, result.ReasonCode = journal.StateInDoubt, journal.ReasonDispatchAmbiguous
		}
		return result, err
	}
}

// SetStrategyFinalTransportProofForTest restores the production final
// EntryGate interlock while replacing only the journal transport proof. Tests
// use it to pause the final refresh and make lease/owner drift deterministic.
func (g *Gateway) SetStrategyFinalTransportProofForTest(
	require func(context.Context, journal.StrategyDispatchLeaseCAS) error,
) {
	g.requireStrategyTransport = require
	g.skipStrategyEntryABAForTest = false
}

// SetStrategyPreTransportRefusalForTest observes the exact terminal refusal
// handoff without exporting a production lease/release writer. TESTS ONLY.
func (g *Gateway) SetStrategyPreTransportRefusalForTest(
	refuse func(context.Context, journal.StrategyDispatchPreTransportRefusalRequest) (journal.StrategyDispatchLease, error),
) {
	g.refuseStrategyPreTransport = refuse
}

// StrategyLineageForFirstLegTest exercises the atomic execgw->journal bridge
// without exposing the legacy DecisionRecord adapter in production builds.
func StrategyLineageForFirstLegTest(record strategyengine.DecisionRecord, quantity, policyVersion, settingsDigest, activationDigest string) (journal.StrategyDecisionLineage, error) {
	return strategyDecisionLineage(record, quantity, policyVersion, settingsDigest, activationDigest)
}

// This package's test binary supplies an opaque, non-scalar harness for legacy
// gateway tests whose subject is not readiness.
//
// Nearly every suite here drives a buy because that mutation carries the most
// machinery. The readiness-specific tests opt into the real adapter instead.
//
// It cannot leak: export_test.go is compiled only into `go test` for this
// package.
func init() {
	defaultProtectionCheckForTest = func(_ context.Context, market string, previous protectionCheckpoint) (protectionCheckpoint, *RejectedError) {
		return protectionCheckpoint{testIdentity: "legacy-test:" + market}, nil
	}
}
