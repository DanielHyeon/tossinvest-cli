package positioncampaign

import (
	"errors"
	"testing"
)

func TestCampaignTransitionTable(t *testing.T) {
	tests := []struct {
		name  string
		from  CampaignState
		event CampaignEvent
		want  CampaignState
		block bool
	}{
		{"planned retry", CampaignPlanned, CampaignRetry, CampaignPlanned, false},
		{"planned leg", CampaignPlanned, CampaignLegPlanned, CampaignPlanned, false},
		{"planned submit", CampaignPlanned, CampaignOrderLinked, CampaignActive, false},
		{"planned cancel", CampaignPlanned, CampaignCancelledBeforeFill, CampaignClosed, true},
		{"planned invalid", CampaignPlanned, CampaignStructuralInvalid, CampaignClosed, true},
		{"planned mismatch", CampaignPlanned, CampaignLineageMismatch, CampaignReconcile, true},
		{"active progress", CampaignActive, CampaignLegProgress, CampaignActive, false},
		{"active exit", CampaignActive, CampaignExitRequested, CampaignExiting, true},
		{"active closing", CampaignActive, CampaignPositionClosing, CampaignExiting, true},
		{"active closed", CampaignActive, CampaignPositionClosed, CampaignClosed, true},
		{"active mismatch", CampaignActive, CampaignWatermarkMismatch, CampaignReconcile, true},
		{"exiting progress", CampaignExiting, CampaignRiskReducingProgress, CampaignExiting, true},
		{"exiting closed", CampaignExiting, CampaignPositionClosed, CampaignClosed, true},
		{"exiting mismatch", CampaignExiting, CampaignLineageMismatch, CampaignReconcile, true},
		{"reconcile open", CampaignReconcile, CampaignEvidenceOpen, CampaignActive, false},
		{"reconcile closing", CampaignReconcile, CampaignEvidenceClosing, CampaignExiting, true},
		{"reconcile closed", CampaignReconcile, CampaignEvidenceClosed, CampaignClosed, true},
		{"reconcile incomplete", CampaignReconcile, CampaignEvidenceIncomplete, CampaignReconcile, true},
		{"closed retry", CampaignClosed, CampaignTerminalRetry, CampaignClosed, true},
		{"closed late fact", CampaignClosed, CampaignLateFact, CampaignClosed, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransitionCampaign(tt.from, tt.event)
			if err != nil {
				t.Fatal(err)
			}
			if got.State != tt.want || got.EntryBlocked != tt.block {
				t.Fatalf("got %+v, want state=%s block=%v", got, tt.want, tt.block)
			}
		})
	}
	if _, err := TransitionCampaign(CampaignClosed, CampaignOrderLinked); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("closed submit err=%v, want invalid transition", err)
	}
}

func TestLegTransitionTable(t *testing.T) {
	tests := []struct {
		name  string
		from  LegState
		event LegEvent
		want  LegState
	}{
		{"planned retry", LegPlanned, LegPlanRetry, LegPlanned},
		{"planned link", LegPlanned, LegOrderLinked, LegSubmitted},
		{"planned cancel", LegPlanned, LegCancelledBeforeSubmit, LegCancelled},
		{"planned orphan fill", LegPlanned, LegPositiveFillWithoutLineage, LegReconcile},
		{"submitted duplicate", LegSubmitted, LegDuplicateObservation, LegSubmitted},
		{"submitted partial", LegSubmitted, LegPartialFill, LegPartial},
		{"submitted full", LegSubmitted, LegFullFill, LegFilled},
		{"submitted cancel", LegSubmitted, LegZeroFillCancelled, LegCancelled},
		{"submitted replacement", LegSubmitted, LegReplacementLinked, LegSubmitted},
		{"partial duplicate", LegPartial, LegDuplicateObservation, LegPartial},
		{"partial more", LegPartial, LegPartialFill, LegPartial},
		{"partial full", LegPartial, LegFullFill, LegFilled},
		{"partial cancel", LegPartial, LegResidualCancelled, LegCancelled},
		{"partial replacement", LegPartial, LegReplacementLinked, LegPartial},
		{"reconcile submitted", LegReconcile, LegEvidenceSubmitted, LegSubmitted},
		{"reconcile partial", LegReconcile, LegEvidencePartial, LegPartial},
		{"reconcile filled", LegReconcile, LegEvidenceFilled, LegFilled},
		{"reconcile cancelled", LegReconcile, LegEvidenceCancelled, LegCancelled},
		{"reconcile incomplete", LegReconcile, LegEvidenceIncomplete, LegReconcile},
		{"filled retry", LegFilled, LegTerminalRetry, LegFilled},
		{"cancelled retry", LegCancelled, LegTerminalRetry, LegCancelled},
		{"filled late", LegFilled, LegLatePositiveFill, LegFilled},
		{"cancelled late", LegCancelled, LegLatePositiveFill, LegCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransitionLeg(tt.from, tt.event)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestExitFirstAdmission(t *testing.T) {
	for _, in := range []Admission{
		{CampaignState: CampaignExiting},
		{CampaignState: CampaignReconcile},
		{CampaignState: CampaignActive, PositionClosing: true},
		{CampaignState: CampaignActive, RiskReducingPending: true},
	} {
		if err := AdmitExposure(in); !errors.Is(err, ErrExposureBlocked) {
			t.Fatalf("AdmitExposure(%+v) err=%v, want blocked", in, err)
		}
	}
	if err := AdmitExposure(Admission{CampaignState: CampaignActive}); err != nil {
		t.Fatalf("active admission: %v", err)
	}
}
