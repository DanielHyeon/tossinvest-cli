// Package positioncampaign owns strategy-neutral campaign, leg and immutable
// broker-order lineage rules. It does not submit orders and has no broker or
// runtime-toggle dependency.
package positioncampaign

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransition = errors.New("position campaign: invalid transition")
	ErrExposureBlocked   = errors.New("position campaign: exposure raising command blocked")
)

type CampaignState string

const (
	CampaignPlanned   CampaignState = "PLANNED"
	CampaignActive    CampaignState = "ACTIVE"
	CampaignExiting   CampaignState = "EXITING"
	CampaignClosed    CampaignState = "CLOSED"
	CampaignReconcile CampaignState = "RECONCILE"
)

type LegState string

const (
	LegPlanned   LegState = "PLANNED"
	LegSubmitted LegState = "SUBMITTED"
	LegPartial   LegState = "PARTIAL"
	LegFilled    LegState = "FILLED"
	LegCancelled LegState = "CANCELLED"
	LegReconcile LegState = "RECONCILE"
)

type CampaignEvent string

const (
	CampaignRetry                CampaignEvent = "RETRY"
	CampaignLegPlanned           CampaignEvent = "LEG_PLANNED"
	CampaignOrderLinked          CampaignEvent = "ORDER_LINKED"
	CampaignCancelledBeforeFill  CampaignEvent = "CANCELLED_BEFORE_FILL"
	CampaignStructuralInvalid    CampaignEvent = "STRUCTURAL_INVALID"
	CampaignLineageMismatch      CampaignEvent = "LINEAGE_MISMATCH"
	CampaignGenerationMismatch   CampaignEvent = "GENERATION_MISMATCH"
	CampaignWatermarkMismatch    CampaignEvent = "WATERMARK_MISMATCH"
	CampaignLegProgress          CampaignEvent = "LEG_PROGRESS"
	CampaignExitRequested        CampaignEvent = "EXIT_REQUESTED"
	CampaignPositionClosing      CampaignEvent = "POSITION_CLOSING"
	CampaignPositionClosed       CampaignEvent = "POSITION_CLOSED"
	CampaignRiskReducingProgress CampaignEvent = "RISK_REDUCING_PROGRESS"
	CampaignEvidenceOpen         CampaignEvent = "EVIDENCE_OPEN"
	CampaignEvidenceClosing      CampaignEvent = "EVIDENCE_CLOSING"
	CampaignEvidenceClosed       CampaignEvent = "EVIDENCE_CLOSED"
	CampaignEvidenceIncomplete   CampaignEvent = "EVIDENCE_INCOMPLETE"
	CampaignTerminalRetry        CampaignEvent = "TERMINAL_RETRY"
	CampaignLateFact             CampaignEvent = "LATE_FACT"
)

type CampaignTransition struct {
	State        CampaignState
	EntryBlocked bool
	Quarantined  bool
}

// TransitionCampaign is the executable form of design D4. Events absent from
// the table are rejected; callers may record them as isolated evidence, but may
// not infer a state transition from them.
func TransitionCampaign(from CampaignState, event CampaignEvent) (CampaignTransition, error) {
	next := CampaignTransition{State: from, EntryBlocked: from == CampaignExiting || from == CampaignClosed || from == CampaignReconcile}
	switch from {
	case CampaignPlanned:
		switch event {
		case CampaignRetry, CampaignLegPlanned:
			return next, nil
		case CampaignOrderLinked:
			return CampaignTransition{State: CampaignActive}, nil
		case CampaignCancelledBeforeFill, CampaignStructuralInvalid:
			return CampaignTransition{State: CampaignClosed, EntryBlocked: true}, nil
		case CampaignLineageMismatch, CampaignGenerationMismatch:
			return CampaignTransition{State: CampaignReconcile, EntryBlocked: true, Quarantined: true}, nil
		}
	case CampaignActive:
		switch event {
		case CampaignRetry, CampaignLegPlanned, CampaignOrderLinked, CampaignLegProgress:
			return next, nil
		case CampaignExitRequested, CampaignPositionClosing:
			return CampaignTransition{State: CampaignExiting, EntryBlocked: true}, nil
		case CampaignPositionClosed:
			return CampaignTransition{State: CampaignClosed, EntryBlocked: true}, nil
		case CampaignLineageMismatch, CampaignGenerationMismatch, CampaignWatermarkMismatch:
			return CampaignTransition{State: CampaignReconcile, EntryBlocked: true, Quarantined: true}, nil
		}
	case CampaignExiting:
		switch event {
		case CampaignRiskReducingProgress, CampaignRetry:
			return next, nil
		case CampaignPositionClosed:
			return CampaignTransition{State: CampaignClosed, EntryBlocked: true}, nil
		case CampaignLineageMismatch, CampaignGenerationMismatch, CampaignWatermarkMismatch:
			return CampaignTransition{State: CampaignReconcile, EntryBlocked: true, Quarantined: true}, nil
		}
	case CampaignReconcile:
		switch event {
		case CampaignEvidenceOpen:
			return CampaignTransition{State: CampaignActive}, nil
		case CampaignEvidenceClosing:
			return CampaignTransition{State: CampaignExiting, EntryBlocked: true}, nil
		case CampaignEvidenceClosed:
			return CampaignTransition{State: CampaignClosed, EntryBlocked: true}, nil
		case CampaignEvidenceIncomplete, CampaignRetry, CampaignLateFact:
			return next, nil
		}
	case CampaignClosed:
		switch event {
		case CampaignTerminalRetry:
			return next, nil
		case CampaignLateFact:
			next.Quarantined = true
			return next, nil
		}
	}
	return CampaignTransition{}, fmt.Errorf("%w: campaign %s cannot accept %s", ErrInvalidTransition, from, event)
}

type LegEvent string

const (
	LegPlanRetry                  LegEvent = "PLAN_RETRY"
	LegOrderLinked                LegEvent = "ORDER_LINKED"
	LegCancelledBeforeSubmit      LegEvent = "CANCELLED_BEFORE_SUBMIT"
	LegPositiveFillWithoutLineage LegEvent = "POSITIVE_FILL_WITHOUT_LINEAGE"
	LegDuplicateObservation       LegEvent = "DUPLICATE_OBSERVATION"
	LegPartialFill                LegEvent = "PARTIAL_FILL"
	LegFullFill                   LegEvent = "FULL_FILL"
	LegZeroFillCancelled          LegEvent = "ZERO_FILL_CANCELLED"
	LegResidualCancelled          LegEvent = "RESIDUAL_CANCELLED"
	LegReplacementLinked          LegEvent = "REPLACEMENT_LINKED"
	LegEvidenceSubmitted          LegEvent = "EVIDENCE_SUBMITTED"
	LegEvidencePartial            LegEvent = "EVIDENCE_PARTIAL"
	LegEvidenceFilled             LegEvent = "EVIDENCE_FILLED"
	LegEvidenceCancelled          LegEvent = "EVIDENCE_CANCELLED"
	LegEvidenceIncomplete         LegEvent = "EVIDENCE_INCOMPLETE"
	LegTerminalRetry              LegEvent = "TERMINAL_RETRY"
	LegLatePositiveFill           LegEvent = "LATE_POSITIVE_FILL"
)

func TransitionLeg(from LegState, event LegEvent) (LegState, error) {
	switch from {
	case LegPlanned:
		switch event {
		case LegPlanRetry:
			return from, nil
		case LegOrderLinked:
			return LegSubmitted, nil
		case LegCancelledBeforeSubmit:
			return LegCancelled, nil
		case LegPositiveFillWithoutLineage:
			return LegReconcile, nil
		}
	case LegSubmitted:
		switch event {
		case LegDuplicateObservation:
			return from, nil
		case LegPartialFill:
			return LegPartial, nil
		case LegFullFill:
			return LegFilled, nil
		case LegZeroFillCancelled:
			return LegCancelled, nil
		case LegReplacementLinked:
			return from, nil
		}
	case LegPartial:
		switch event {
		case LegDuplicateObservation, LegPartialFill, LegReplacementLinked:
			return from, nil
		case LegFullFill:
			return LegFilled, nil
		case LegResidualCancelled:
			return LegCancelled, nil
		}
	case LegReconcile:
		switch event {
		case LegEvidenceSubmitted:
			return LegSubmitted, nil
		case LegEvidencePartial:
			return LegPartial, nil
		case LegEvidenceFilled:
			return LegFilled, nil
		case LegEvidenceCancelled:
			return LegCancelled, nil
		case LegEvidenceIncomplete:
			return from, nil
		}
	case LegFilled, LegCancelled:
		switch event {
		case LegTerminalRetry, LegLatePositiveFill:
			return from, nil
		}
	}
	return "", fmt.Errorf("%w: leg %s cannot accept %s", ErrInvalidTransition, from, event)
}

type Admission struct {
	CampaignState       CampaignState
	PositionClosing     bool
	RiskReducingPending bool
}

// AdmitExposure enforces EXIT FIRST at command admission. It says nothing
// about risk-reducing observations, which must continue through every state.
func AdmitExposure(in Admission) error {
	if in.PositionClosing || in.RiskReducingPending || in.CampaignState == CampaignExiting ||
		in.CampaignState == CampaignReconcile || in.CampaignState == CampaignClosed {
		return ErrExposureBlocked
	}
	if in.CampaignState != CampaignPlanned && in.CampaignState != CampaignActive {
		return fmt.Errorf("%w: unknown campaign state %q", ErrExposureBlocked, in.CampaignState)
	}
	return nil
}
