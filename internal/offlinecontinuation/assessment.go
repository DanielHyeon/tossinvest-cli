// Package offlinecontinuation assesses existing follow-on proposals without
// producing admission, reservation, dispatch, or broker authority.
package offlinecontinuation

import (
	"errors"
	"math/big"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

const (
	familyContinuation = "continuation"
	familyReversal     = "reversal"
	familyWeekly       = "weekly"

	statusAssessed = "ASSESSED"
	statusRefused  = "REFUSED"
	statusNoBudget = "NO_BUDGET"
)

// Assumption binds a caller-supplied account-currency upper bound to one
// native risk snapshot. The bound is not authenticated or execution authority.
type Assumption struct {
	Family                     string
	BindingDigest              string
	Market                     string
	AccountCurrency            string
	PerUnitRiskUpperBoundMinor string
}

// Result contains inspection-only values. Nil quantities or amounts mean that
// the corresponding calculation was not available; zero is never authority.
type Result struct {
	Status           string
	Code             string
	NativeKind       string
	NativeCode       string
	InvalidationCode string

	QCandidate  *uint64
	QOffline    *uint64
	Budget      *string
	Filled      *string
	Held        *string
	Remaining   *string
	PerUnit     *string
	PlannedRisk *string

	InspectionOnly     bool
	LedgerCurrentness  string
	RiskBoundSource    string
	ExecutionAuthority bool
}

type nativeProposal struct {
	kind         string
	code         string
	invalidation string
	quantity     uint64
	accepted     bool
}

func baseResult(n nativeProposal) Result {
	return Result{
		Status: statusRefused, Code: n.code, NativeKind: n.kind, NativeCode: n.code,
		InvalidationCode: n.invalidation, InspectionOnly: true,
		LedgerCurrentness: "unverified", RiskBoundSource: "caller_assumption",
		ExecutionAuthority: false,
	}
}

// AssessContinuationKR preserves the native KR continuation proposal before
// applying a caller-supplied, inspection-only budget bound.
func AssessContinuationKR(request continuationlane.KREvaluationRequest, assumption Assumption) Result {
	n := continuationProposal(continuationlane.ProposeKR(request))
	return assess(n, request.Context.Leg.Ordinal, 2, 3, assumption, familyContinuation, request.Evidence.Envelope.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return continuationlane.CaptureOfflineEvidence(request.Context.Plan, request.Context.Risk, at)
	})
}

// AssessContinuationUS preserves the native US continuation proposal before
// applying a caller-supplied, inspection-only budget bound.
func AssessContinuationUS(request continuationlane.USEvaluationRequest, assumption Assumption) Result {
	n := continuationProposal(continuationlane.ProposeUS(request))
	return assess(n, request.Context.Leg.Ordinal, 2, 3, assumption, familyContinuation, request.Evidence.Envelope.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return continuationlane.CaptureOfflineEvidence(request.Context.Plan, request.Context.Risk, at)
	})
}

// AssessReversalKR preserves the native KR reversal proposal before applying
// a caller-supplied, inspection-only budget bound.
func AssessReversalKR(request reversallane.KREvaluationRequest, assumption Assumption) Result {
	n := reversalProposal(reversallane.ProposeKR(request))
	return assessAt(n, request.Context.Leg.Ordinal, 2, 3, assumption, familyReversal, request.Evidence.CommonEnvelope.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return reversallane.CaptureOfflineEvidence(request.Context.Plan, request.Context.Risk, at)
	})
}

// AssessReversalUS preserves the native US reversal proposal before applying
// a caller-supplied, inspection-only budget bound.
func AssessReversalUS(request reversallane.USEvaluationRequest, assumption Assumption) Result {
	n := reversalProposal(reversallane.ProposeUS(request))
	return assessAt(n, request.Context.Leg.Ordinal, 2, 3, assumption, familyReversal, request.Evidence.CommonEnvelope.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return reversallane.CaptureOfflineEvidence(request.Context.Plan, request.Context.Risk, at)
	})
}

// AssessWeeklyKR preserves the native KR weekly proposal before applying a
// caller-supplied, inspection-only budget bound.
func AssessWeeklyKR(request weeklyvaluelane.EvaluationRequest, assumption Assumption) Result {
	n := weeklyProposal(weeklyvaluelane.ProposeKR(request))
	return assessAt(n, request.Leg.Ordinal, 2, 7, assumption, familyWeekly, request.Evidence.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return weeklyvaluelane.CaptureOfflineEvidence(request.Plan, request.Risk, at)
	})
}

// AssessWeeklyUS preserves the native US weekly proposal before applying a
// caller-supplied, inspection-only budget bound.
func AssessWeeklyUS(request weeklyvaluelane.EvaluationRequest, assumption Assumption) Result {
	n := weeklyProposal(weeklyvaluelane.ProposeUS(request))
	return assessAt(n, request.Leg.Ordinal, 2, 7, assumption, familyWeekly, request.Evidence.EvaluatedAt, func(at time.Time) (evidence.Evidence, error) {
		return weeklyvaluelane.CaptureOfflineEvidence(request.Plan, request.Risk, at)
	})
}

func continuationProposal(out continuationlane.Outcome) nativeProposal {
	return nativeProposal{kind: string(out.Kind), code: string(out.Code), invalidation: out.InvalidationCode, quantity: out.Quantity, accepted: out.Kind == continuationlane.OutcomeDecision}
}

func reversalProposal(out reversallane.EvaluationResult) nativeProposal {
	return nativeProposal{kind: string(out.Kind), code: string(out.Code), quantity: out.Quantity, accepted: out.Kind == reversallane.OutcomeDecision}
}

func weeklyProposal(out weeklyvaluelane.Outcome) nativeProposal {
	return nativeProposal{kind: string(out.Kind), code: string(out.Code), invalidation: out.InvalidationCode, quantity: out.Quantity, accepted: out.Kind == weeklyvaluelane.OutcomeDecision}
}

func assess(n nativeProposal, ordinal, first, last int, assumption Assumption, family, evaluatedAt string, capture func(time.Time) (evidence.Evidence, error)) Result {
	if !n.accepted {
		return baseResult(n)
	}
	at, err := time.Parse(time.RFC3339Nano, evaluatedAt)
	if err != nil {
		out := baseResult(n)
		out.Code = "INVALID_EVALUATED_AT"
		return out
	}
	return assessAt(n, ordinal, first, last, assumption, family, at, capture)
}

func assessAt(n nativeProposal, ordinal, first, last int, assumption Assumption, family string, at time.Time, capture func(time.Time) (evidence.Evidence, error)) Result {
	out := baseResult(n)
	if !n.accepted {
		return out
	}
	if ordinal < first || ordinal > last {
		out.Code = "FOLLOW_ON_ORDINAL_REQUIRED"
		return out
	}
	snapshot, err := capture(at)
	if err != nil {
		out.Code = "CAPTURE_FAILED"
		return out
	}
	if err := validateAssumption(assumption, family, snapshot); err != nil {
		out.Code = err.Error()
		return out
	}
	return calculate(out, snapshot, assumption.PerUnitRiskUpperBoundMinor, n.quantity)
}

func validateAssumption(a Assumption, family string, e evidence.Evidence) error {
	if a.Family != family {
		return errors.New("ASSUMPTION_FAMILY_MISMATCH")
	}
	if a.BindingDigest == "" || a.BindingDigest != e.Risk.PlanDigest {
		return errors.New("ASSUMPTION_DIGEST_MISMATCH")
	}
	if a.Market != e.Plan.Scope.Market {
		return errors.New("ASSUMPTION_MARKET_MISMATCH")
	}
	if a.AccountCurrency != e.Plan.AccountCurrency {
		return errors.New("ASSUMPTION_CURRENCY_MISMATCH")
	}
	return nil
}

func calculate(out Result, e evidence.Evidence, rawUnit string, nativeCandidate uint64) Result {
	budget, err := boundedDecimal(e.Plan.Budget, true)
	if err != nil {
		out.Code = "BUDGET_INVALID"
		return out
	}
	filled, err := boundedDecimal(e.Risk.Filled, false)
	if err != nil {
		out.Code = "FILLED_INVALID"
		return out
	}
	held, err := boundedDecimal(e.Risk.Held, false)
	if err != nil {
		out.Code = "HELD_INVALID"
		return out
	}
	unit, err := boundedDecimal(rawUnit, true)
	if err != nil {
		out.Code = "PER_UNIT_BOUND_INVALID"
		return out
	}
	used := new(big.Int).Add(filled, held)
	if used.BitLen() > 256 {
		out.Code = "RISK_SUM_OVERFLOW"
		return out
	}
	if used.Cmp(budget) > 0 {
		out.Code = "RISK_OVERDRAWN"
		return out
	}
	remaining := new(big.Int).Sub(budget, used)
	out.Budget, out.Filled, out.Held, out.Remaining, out.PerUnit = ptr(budget.String()), ptr(filled.String()), ptr(held.String()), ptr(remaining.String()), ptr(unit.String())
	limit := new(big.Int).Quo(remaining, unit)
	candidate := new(big.Int).SetUint64(nativeCandidate)
	if candidate.Cmp(limit) > 0 {
		candidate = limit
	}
	if !candidate.IsUint64() {
		out.Code = "QUANTITY_INVALID"
		return out
	}
	q := candidate.Uint64()
	if q == 0 {
		out.Status, out.Code = statusNoBudget, "NO_BUDGET"
		return out
	}
	planned := new(big.Int).Mul(candidate, unit)
	if planned.BitLen() > 256 {
		out.Code = "PLANNED_RISK_OVERFLOW"
		return out
	}
	out.Status, out.Code, out.QCandidate, out.QOffline, out.PlannedRisk = statusAssessed, "", ptrUint(nativeCandidate), ptrUint(q), ptr(planned.String())
	return out
}

func boundedDecimal(raw string, positive bool) (*big.Int, error) {
	if len(raw) == 0 || len(raw) > 78 || (len(raw) > 1 && raw[0] == '0') {
		return nil, errors.New("invalid")
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return nil, errors.New("invalid")
		}
	}
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok || n.BitLen() > 256 || (positive && n.Sign() == 0) {
		return nil, errors.New("invalid")
	}
	return n, nil
}

func ptr(value string) *string     { return &value }
func ptrUint(value uint64) *uint64 { return &value }
